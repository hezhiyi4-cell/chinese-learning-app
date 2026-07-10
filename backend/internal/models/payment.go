package models

import "time"

type PaymentOrder struct {
	ID              uint       `gorm:"primaryKey" json:"id"`
	UserID          uint       `gorm:"not null;index" json:"userId"`
	PlanCode        string     `gorm:"size:50;not null;index" json:"planCode"`
	PlanName        string     `gorm:"size:120;not null" json:"planName"`
	ProductType     string     `gorm:"size:30;not null" json:"productType"`
	BillingCycle    string     `gorm:"size:30" json:"billingCycle"`
	Gateway         string     `gorm:"size:40;not null;index" json:"gateway"`
	Currency        string     `gorm:"size:10;not null" json:"currency"`
	Amount          string     `gorm:"size:32;not null" json:"amount"`
	Status          string     `gorm:"size:30;default:'created';index" json:"status"`
	ExternalOrderID string     `gorm:"size:120;index" json:"externalOrderId"`
	ApprovalURL     string     `gorm:"size:500" json:"approvalUrl"`
	ReceiverHint    string     `gorm:"size:200" json:"receiverHint"`
	GatewayPayload  string     `gorm:"type:text" json:"gatewayPayload"`
	PaidAt          *time.Time `json:"paidAt"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       time.Time  `json:"updatedAt"`
}

type PaymentSubscription struct {
	ID                     uint       `gorm:"primaryKey" json:"id"`
	UserID                 uint       `gorm:"not null;index" json:"userId"`
	PlanCode               string     `gorm:"size:50;not null" json:"planCode"`
	PlanName               string     `gorm:"size:120;not null" json:"planName"`
	Gateway                string     `gorm:"size:40;not null" json:"gateway"`
	Currency               string     `gorm:"size:10;not null" json:"currency"`
	Amount                 string     `gorm:"size:32;not null" json:"amount"`
	Status                 string     `gorm:"size:30;default:'pending';index" json:"status"`
	BillingCycle           string     `gorm:"size:30;not null" json:"billingCycle"`
	ExternalSubscriptionID string     `gorm:"size:120;index" json:"externalSubscriptionId"`
	AutoRenew              bool       `gorm:"default:true" json:"autoRenew"`
	CurrentPeriodStart     *time.Time `json:"currentPeriodStart"`
	CurrentPeriodEnd       *time.Time `json:"currentPeriodEnd"`
	CreatedAt              time.Time  `json:"createdAt"`
	UpdatedAt              time.Time  `json:"updatedAt"`
}
