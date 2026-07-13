package repositories

import (
	"chinese-learning-app/internal/models"

	"gorm.io/gorm"
)

type PaymentRepository struct {
	db *gorm.DB
}

func NewPaymentRepository(db *gorm.DB) *PaymentRepository {
	return &PaymentRepository{db: db}
}

func (r *PaymentRepository) CreateOrder(order *models.PaymentOrder) error {
	return r.db.Create(order).Error
}

func (r *PaymentRepository) UpdateOrder(order *models.PaymentOrder) error {
	return r.db.Save(order).Error
}

func (r *PaymentRepository) GetOrderByIDAndUser(orderID, userID uint) (*models.PaymentOrder, error) {
	var order models.PaymentOrder
	err := r.db.Where("id = ? AND user_id = ?", orderID, userID).First(&order).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &order, nil
}

func (r *PaymentRepository) ListOrdersByUser(userID uint, limit int) ([]models.PaymentOrder, error) {
	var orders []models.PaymentOrder
	query := r.db.Where("user_id = ?", userID).Order("created_at DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	if err := query.Find(&orders).Error; err != nil {
		return nil, err
	}
	return orders, nil
}

func (r *PaymentRepository) GetLatestSubscriptionByUser(userID uint) (*models.PaymentSubscription, error) {
	var sub models.PaymentSubscription
	err := r.db.Where("user_id = ?", userID).Order("updated_at DESC").First(&sub).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &sub, nil
}

func (r *PaymentRepository) ReplaceUserSubscription(userID uint, sub *models.PaymentSubscription) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.PaymentSubscription{}).
			Where("user_id = ? AND status = ?", userID, "active").
			Updates(map[string]any{"status": "replaced"}).Error; err != nil {
			return err
		}
		return tx.Select("*").Create(sub).Error
	})
}
