package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"chinese-learning-app/internal/models"
	"chinese-learning-app/internal/repositories"
)

type PaymentGatewayInfo struct {
	Code              string   `json:"code"`
	Name              string   `json:"name"`
	Status            string   `json:"status"`
	Sandbox           bool     `json:"sandbox"`
	SupportsCheckout  bool     `json:"supportsCheckout"`
	SettlementTargets []string `json:"settlementTargets"`
	Note              string   `json:"note"`
}

type PaymentPlanPrice struct {
	Currency string `json:"currency"`
	Amount   string `json:"amount"`
	Label    string `json:"label"`
}

type PaymentPlan struct {
	Code         string             `json:"code"`
	Name         string             `json:"name"`
	Description  string             `json:"description"`
	ProductType  string             `json:"productType"`
	BillingCycle string             `json:"billingCycle"`
	Features     []string           `json:"features"`
	Prices       []PaymentPlanPrice `json:"prices"`
}

type PaymentCatalogResponse struct {
	Gateways            []PaymentGatewayInfo        `json:"gateways"`
	Plans               []PaymentPlan               `json:"plans"`
	SupportedCurrencies []string                    `json:"supportedCurrencies"`
	LatestSubscription  *models.PaymentSubscription `json:"latestSubscription"`
	Orders              []models.PaymentOrder       `json:"orders"`
}

type CreateCheckoutRequest struct {
	Gateway    string `json:"gateway"`
	PlanCode   string `json:"planCode"`
	Currency   string `json:"currency"`
	SuccessURL string `json:"successUrl"`
	CancelURL  string `json:"cancelUrl"`
}

type CheckoutResponse struct {
	Order       *models.PaymentOrder `json:"order"`
	ApprovalURL string               `json:"approvalUrl"`
	Message     string               `json:"message"`
}

type PaymentService struct {
	paymentRepo     *repositories.PaymentRepository
	frontendBaseURL string
	paypalGateway   *PayPalGateway
}

func NewPaymentService(paymentRepo *repositories.PaymentRepository, userRepo *repositories.UserRepository, frontendBaseURL string, paypalGateway *PayPalGateway) *PaymentService {
	return &PaymentService{
		paymentRepo:     paymentRepo,
		frontendBaseURL: strings.TrimRight(frontendBaseURL, "/"),
		paypalGateway:   paypalGateway,
	}
}

func (s *PaymentService) GetCatalog(userID uint) (*PaymentCatalogResponse, error) {
	orders, err := s.paymentRepo.ListOrdersByUser(userID, 8)
	if err != nil {
		return nil, err
	}
	sub, err := s.paymentRepo.GetLatestSubscriptionByUser(userID)
	if err != nil {
		return nil, err
	}

	return &PaymentCatalogResponse{
		Gateways:            s.gatewayInfos(),
		Plans:               defaultPaymentPlans(),
		SupportedCurrencies: []string{"HKD", "USD", "CNY"},
		LatestSubscription:  sub,
		Orders:              orders,
	}, nil
}

func (s *PaymentService) CreateCheckout(userID uint, req CreateCheckoutRequest) (*CheckoutResponse, error) {
	plan, ok := findPaymentPlan(req.PlanCode)
	if !ok {
		return nil, fmt.Errorf("payment plan not found")
	}
	price, ok := findPlanPrice(plan, req.Currency)
	if !ok {
		return nil, fmt.Errorf("currency is not supported for this plan")
	}

	gatewayCode := normalizeGateway(req.Gateway)
	if gatewayCode != "paypal" {
		return nil, fmt.Errorf("selected gateway is reserved but not enabled yet")
	}

	order := &models.PaymentOrder{
		UserID:       userID,
		PlanCode:     plan.Code,
		PlanName:     plan.Name,
		ProductType:  plan.ProductType,
		BillingCycle: plan.BillingCycle,
		Gateway:      gatewayCode,
		Currency:     price.Currency,
		Amount:       price.Amount,
		Status:       "created",
		ReceiverHint: "支持后续切换至香港收款账户或深圳对公主体",
	}
	if err := s.paymentRepo.CreateOrder(order); err != nil {
		return nil, err
	}

	successURL := s.defaultSuccessURL(order.ID)
	cancelURL := s.defaultCancelURL(order.ID)
	if req.SuccessURL != "" {
		successURL = req.SuccessURL
	}
	if req.CancelURL != "" {
		cancelURL = req.CancelURL
	}

	checkout, err := s.paypalGateway.CreateCheckout(context.Background(), GatewayCheckoutInput{
		Order:      order,
		SuccessURL: successURL,
		CancelURL:  cancelURL,
	})
	if err != nil {
		order.Status = "failed"
		order.GatewayPayload = err.Error()
		_ = s.paymentRepo.UpdateOrder(order)
		return nil, err
	}

	order.ExternalOrderID = checkout.ExternalOrderID
	order.ApprovalURL = checkout.ApprovalURL
	order.Status = checkout.Status
	order.GatewayPayload = checkout.Payload
	if err := s.paymentRepo.UpdateOrder(order); err != nil {
		return nil, err
	}

	message := "订单已创建，可前往支付"
	if order.ApprovalURL == "" {
		message = "PayPal Sandbox 凭据未配置，当前使用本地确认模式演示支付流程"
	}

	return &CheckoutResponse{
		Order:       order,
		ApprovalURL: order.ApprovalURL,
		Message:     message,
	}, nil
}

func (s *PaymentService) ConfirmCheckout(userID, orderID uint) (*CheckoutResponse, error) {
	order, err := s.paymentRepo.GetOrderByIDAndUser(orderID, userID)
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, fmt.Errorf("payment order not found")
	}
	if order.Status == "paid" {
		return &CheckoutResponse{
			Order:       order,
			ApprovalURL: order.ApprovalURL,
			Message:     "订单已经支付成功",
		}, nil
	}

	confirm, err := s.paypalGateway.ConfirmCheckout(context.Background(), order)
	if err != nil {
		return nil, err
	}

	order.Status = confirm.Status
	order.GatewayPayload = confirm.Payload
	if confirm.ExternalOrderID != "" {
		order.ExternalOrderID = confirm.ExternalOrderID
	}

	if confirm.Paid {
		now := time.Now()
		order.Status = "paid"
		order.PaidAt = &now
		subscription := &models.PaymentSubscription{
			UserID:                 userID,
			PlanCode:               order.PlanCode,
			PlanName:               order.PlanName,
			Gateway:                order.Gateway,
			Currency:               order.Currency,
			Amount:                 order.Amount,
			Status:                 "active",
			BillingCycle:           order.BillingCycle,
			ExternalSubscriptionID: order.ExternalOrderID,
			AutoRenew:              true,
			CurrentPeriodStart:     &now,
			CurrentPeriodEnd:       ptrTime(nextBillingTime(now, order.BillingCycle)),
		}
		if err := s.paymentRepo.ReplaceUserSubscription(userID, subscription); err != nil {
			return nil, err
		}
	}

	if err := s.paymentRepo.UpdateOrder(order); err != nil {
		return nil, err
	}

	return &CheckoutResponse{
		Order:       order,
		ApprovalURL: order.ApprovalURL,
		Message:     "支付状态已更新",
	}, nil
}

func (s *PaymentService) gatewayInfos() []PaymentGatewayInfo {
	paypalConfigured := s.paypalGateway != nil && s.paypalGateway.clientID != "" && s.paypalGateway.secret != ""
	paypalNote := "优先使用 PayPal 企业账户 Sandbox 做联调。"
	if !paypalConfigured {
		paypalNote = "已预留 PayPal Sandbox 接口；待填入 PAYPAL_CLIENT_ID / PAYPAL_SECRET 后即可切换到真实沙盒。"
	}

	return []PaymentGatewayInfo{
		{
			Code:              "paypal",
			Name:              "PayPal Sandbox",
			Status:            ternaryString(paypalConfigured, "sandbox_ready", "sandbox_mock"),
			Sandbox:           true,
			SupportsCheckout:  true,
			SettlementTargets: []string{"香港收款账户", "深圳对公账户"},
			Note:              paypalNote,
		},
		{
			Code:              "lianlian",
			Name:              "连连国际",
			Status:            "reserved",
			Sandbox:           false,
			SupportsCheckout:  false,
			SettlementTargets: []string{"香港收款账户", "深圳对公账户"},
			Note:              "已预留接口位，待商户开通后接入。",
		},
		{
			Code:              "airwallex",
			Name:              "Airwallex",
			Status:            "reserved",
			Sandbox:           false,
			SupportsCheckout:  false,
			SettlementTargets: []string{"香港收款账户", "深圳对公账户"},
			Note:              "已预留接口位，待申请通过后接入。",
		},
		{
			Code:              "photonpay",
			Name:              "PhotonPay",
			Status:            "reserved",
			Sandbox:           false,
			SupportsCheckout:  false,
			SettlementTargets: []string{"香港收款账户", "深圳对公账户"},
			Note:              "已预留接口位，待商户资料齐备后接入。",
		},
		{
			Code:              "kpay",
			Name:              "KPay",
			Status:            "reserved",
			Sandbox:           false,
			SupportsCheckout:  false,
			SettlementTargets: []string{"香港收款账户"},
			Note:              "已预留接口位，待香港主体资料齐备后接入。",
		},
	}
}

func defaultPaymentPlans() []PaymentPlan {
	return []PaymentPlan{
		{
			Code:         "starter_monthly",
			Name:         "入门会员",
			Description:  "适合刚开始系统学中文的用户。",
			ProductType:  "subscription",
			BillingCycle: "monthly",
			Features:     []string{"每月解锁全部课程", "AI 练习增强", "学习数据云端保存"},
			Prices: []PaymentPlanPrice{
				{Currency: "HKD", Amount: "68.00", Label: "HK$68 / 月"},
				{Currency: "USD", Amount: "8.90", Label: "$8.90 / 月"},
				{Currency: "CNY", Amount: "49.00", Label: "¥49 / 月"},
			},
		},
		{
			Code:         "pro_quarterly",
			Name:         "进阶会员",
			Description:  "适合高频学习和持续打卡用户。",
			ProductType:  "subscription",
			BillingCycle: "quarterly",
			Features:     []string{"季度订阅更优惠", "优先使用新课程", "多支付通道兼容"},
			Prices: []PaymentPlanPrice{
				{Currency: "HKD", Amount: "188.00", Label: "HK$188 / 季"},
				{Currency: "USD", Amount: "24.90", Label: "$24.90 / 季"},
				{Currency: "CNY", Amount: "138.00", Label: "¥138 / 季"},
			},
		},
		{
			Code:         "lifetime_pack",
			Name:         "终身包",
			Description:  "一次付费，长期锁定内容权益。",
			ProductType:  "one_time",
			BillingCycle: "one_time",
			Features:     []string{"一次买断", "长期使用", "后续可切换正式支付通道"},
			Prices: []PaymentPlanPrice{
				{Currency: "HKD", Amount: "888.00", Label: "HK$888 / 一次性"},
				{Currency: "USD", Amount: "119.00", Label: "$119 / 一次性"},
				{Currency: "CNY", Amount: "699.00", Label: "¥699 / 一次性"},
			},
		},
	}
}

func findPaymentPlan(code string) (PaymentPlan, bool) {
	for _, plan := range defaultPaymentPlans() {
		if plan.Code == code {
			return plan, true
		}
	}
	return PaymentPlan{}, false
}

func findPlanPrice(plan PaymentPlan, currency string) (PaymentPlanPrice, bool) {
	for _, price := range plan.Prices {
		if strings.EqualFold(price.Currency, currency) {
			return price, true
		}
	}
	return PaymentPlanPrice{}, false
}

func normalizeGateway(gateway string) string {
	gateway = strings.ToLower(strings.TrimSpace(gateway))
	if gateway == "" {
		return "paypal"
	}
	return gateway
}

func (s *PaymentService) defaultSuccessURL(orderID uint) string {
	return fmt.Sprintf("%s/web.html?payment=success&orderId=%d", s.frontendBaseURL, orderID)
}

func (s *PaymentService) defaultCancelURL(orderID uint) string {
	return fmt.Sprintf("%s/web.html?payment=cancel&orderId=%d", s.frontendBaseURL, orderID)
}

func ptrTime(t time.Time) *time.Time {
	return &t
}

func nextBillingTime(start time.Time, cycle string) time.Time {
	switch cycle {
	case "monthly":
		return start.AddDate(0, 1, 0)
	case "quarterly":
		return start.AddDate(0, 3, 0)
	case "yearly":
		return start.AddDate(1, 0, 0)
	default:
		return start.AddDate(100, 0, 0)
	}
}

func ternaryString(condition bool, whenTrue, whenFalse string) string {
	if condition {
		return whenTrue
	}
	return whenFalse
}
