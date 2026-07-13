package services

import (
	"testing"

	"chinese-learning-app/internal/models"
	"chinese-learning-app/internal/repositories"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func newPaymentServiceForTest(t *testing.T) (*PaymentService, *repositories.PaymentRepository) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	if err := db.AutoMigrate(&models.PaymentOrder{}, &models.PaymentSubscription{}); err != nil {
		t.Fatalf("migrate test database: %v", err)
	}

	repo := repositories.NewPaymentRepository(db)
	service := NewPaymentService(repo, nil, "https://example.com", NewPayPalGateway("", "", "https://api-m.sandbox.paypal.com"))
	return service, repo
}

func TestConfirmCheckoutSubscriptionEnablesAutoRenew(t *testing.T) {
	service, repo := newPaymentServiceForTest(t)

	checkout, err := service.CreateCheckout(7, CreateCheckoutRequest{
		Gateway:  "paypal",
		PlanCode: "starter_monthly",
		Currency: "HKD",
	})
	if err != nil {
		t.Fatalf("create checkout: %v", err)
	}

	result, err := service.ConfirmCheckout(7, checkout.Order.ID)
	if err != nil {
		t.Fatalf("confirm checkout: %v", err)
	}
	if result.Order.Status != "paid" {
		t.Fatalf("expected paid order status, got %q", result.Order.Status)
	}

	sub, err := repo.GetLatestSubscriptionByUser(7)
	if err != nil {
		t.Fatalf("load latest subscription: %v", err)
	}
	if sub == nil {
		t.Fatal("expected subscription record")
	}
	if !sub.AutoRenew {
		t.Fatal("expected subscription plan to enable auto renew")
	}
	if sub.CurrentPeriodStart == nil || sub.CurrentPeriodEnd == nil {
		t.Fatal("expected subscription plan to have a billing period")
	}
}

func TestConfirmCheckoutLifetimePlanDisablesAutoRenew(t *testing.T) {
	service, repo := newPaymentServiceForTest(t)

	checkout, err := service.CreateCheckout(11, CreateCheckoutRequest{
		Gateway:  "paypal",
		PlanCode: "lifetime_pack",
		Currency: "USD",
	})
	if err != nil {
		t.Fatalf("create checkout: %v", err)
	}

	result, err := service.ConfirmCheckout(11, checkout.Order.ID)
	if err != nil {
		t.Fatalf("confirm checkout: %v", err)
	}
	if result.Order.Status != "paid" {
		t.Fatalf("expected paid order status, got %q", result.Order.Status)
	}

	sub, err := repo.GetLatestSubscriptionByUser(11)
	if err != nil {
		t.Fatalf("load latest subscription: %v", err)
	}
	if sub == nil {
		t.Fatal("expected lifetime access record")
	}
	if sub.AutoRenew {
		t.Fatal("expected one-time plan to disable auto renew")
	}
	if sub.CurrentPeriodStart == nil {
		t.Fatal("expected lifetime access to keep a start time")
	}
	if sub.CurrentPeriodEnd != nil {
		t.Fatal("expected lifetime access to have no period end")
	}
}
