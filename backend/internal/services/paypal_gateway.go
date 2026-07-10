package services

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"chinese-learning-app/internal/models"
)

type PaymentGateway interface {
	Code() string
	CreateCheckout(ctx context.Context, input GatewayCheckoutInput) (*GatewayCheckoutResult, error)
	ConfirmCheckout(ctx context.Context, order *models.PaymentOrder) (*GatewayConfirmResult, error)
}

type GatewayCheckoutInput struct {
	Order      *models.PaymentOrder
	SuccessURL string
	CancelURL  string
}

type GatewayCheckoutResult struct {
	ExternalOrderID string
	ApprovalURL     string
	Status          string
	Payload         string
}

type GatewayConfirmResult struct {
	Paid            bool
	ExternalOrderID string
	Status          string
	Payload         string
}

type PayPalGateway struct {
	clientID   string
	secret     string
	baseURL    string
	httpClient *http.Client
}

func NewPayPalGateway(clientID, secret, baseURL string) *PayPalGateway {
	return &PayPalGateway{
		clientID:   clientID,
		secret:     secret,
		baseURL:    strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{Timeout: 20 * time.Second},
	}
}

func (g *PayPalGateway) Code() string {
	return "paypal"
}

func (g *PayPalGateway) CreateCheckout(ctx context.Context, input GatewayCheckoutInput) (*GatewayCheckoutResult, error) {
	if input.Order == nil {
		return nil, fmt.Errorf("order is required")
	}
	if g.clientID == "" || g.secret == "" {
		mockID := fmt.Sprintf("PAYPAL-MOCK-%d", time.Now().UnixNano())
		payload, _ := json.Marshal(map[string]any{
			"mode":    "mock",
			"message": "PayPal Sandbox credentials not configured yet",
		})
		return &GatewayCheckoutResult{
			ExternalOrderID: mockID,
			ApprovalURL:     "",
			Status:          "pending_confirmation",
			Payload:         string(payload),
		}, nil
	}

	token, err := g.fetchAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	requestBody := map[string]any{
		"intent": "CAPTURE",
		"purchase_units": []map[string]any{
			{
				"reference_id": input.Order.PlanCode,
				"description":  input.Order.PlanName,
				"amount": map[string]string{
					"currency_code": input.Order.Currency,
					"value":         input.Order.Amount,
				},
			},
		},
		"application_context": map[string]string{
			"brand_name":  "Chinese Learning App",
			"user_action": "PAY_NOW",
			"return_url":  input.SuccessURL,
			"cancel_url":  input.CancelURL,
		},
	}

	bodyBytes, _ := json.Marshal(requestBody)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, g.baseURL+"/v2/checkout/orders", bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	payloadBytes, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("paypal create order failed: %s", string(payloadBytes))
	}

	var parsed struct {
		ID     string `json:"id"`
		Status string `json:"status"`
		Links  []struct {
			Rel  string `json:"rel"`
			Href string `json:"href"`
		} `json:"links"`
	}
	if err := json.Unmarshal(payloadBytes, &parsed); err != nil {
		return nil, err
	}

	approvalURL := ""
	for _, link := range parsed.Links {
		if link.Rel == "approve" {
			approvalURL = link.Href
			break
		}
	}

	return &GatewayCheckoutResult{
		ExternalOrderID: parsed.ID,
		ApprovalURL:     approvalURL,
		Status:          "pending_approval",
		Payload:         string(payloadBytes),
	}, nil
}

func (g *PayPalGateway) ConfirmCheckout(ctx context.Context, order *models.PaymentOrder) (*GatewayConfirmResult, error) {
	if order == nil {
		return nil, fmt.Errorf("order is required")
	}

	if strings.HasPrefix(order.ExternalOrderID, "PAYPAL-MOCK-") {
		payload, _ := json.Marshal(map[string]any{
			"mode":    "mock",
			"message": "Mock PayPal sandbox payment confirmed",
		})
		return &GatewayConfirmResult{
			Paid:            true,
			ExternalOrderID: order.ExternalOrderID,
			Status:          "paid",
			Payload:         string(payload),
		}, nil
	}

	token, err := g.fetchAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, g.baseURL+"/v2/checkout/orders/"+order.ExternalOrderID+"/capture", http.NoBody)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	payloadBytes, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("paypal capture failed: %s", string(payloadBytes))
	}

	var parsed struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	}
	if err := json.Unmarshal(payloadBytes, &parsed); err != nil {
		return nil, err
	}

	return &GatewayConfirmResult{
		Paid:            strings.EqualFold(parsed.Status, "COMPLETED"),
		ExternalOrderID: parsed.ID,
		Status:          strings.ToLower(parsed.Status),
		Payload:         string(payloadBytes),
	}, nil
}

func (g *PayPalGateway) fetchAccessToken(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, g.baseURL+"/v1/oauth2/token", strings.NewReader("grant_type=client_credentials"))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	basic := base64.StdEncoding.EncodeToString([]byte(g.clientID + ":" + g.secret))
	req.Header.Set("Authorization", "Basic "+basic)

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	payloadBytes, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("paypal auth failed: %s", string(payloadBytes))
	}

	var parsed struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal(payloadBytes, &parsed); err != nil {
		return "", err
	}
	if parsed.AccessToken == "" {
		return "", fmt.Errorf("paypal auth returned empty access token")
	}
	return parsed.AccessToken, nil
}
