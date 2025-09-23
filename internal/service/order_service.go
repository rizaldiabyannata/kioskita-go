package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rizaldiabyannata/kioskita-go/internal/core"
	"github.com/rizaldiabyannata/kioskita-go/internal/store"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// OrderService provides business logic for orders.
type OrderService struct {
	db           *gorm.DB // For transactions
	orderStore   *store.OrderStore
	productStore *store.ProductStore
	voucherStore *store.VoucherStore
}

// NewOrderService creates a new OrderService.
func NewOrderService(db *gorm.DB, os *store.OrderStore, ps *store.ProductStore, vs *store.VoucherStore) *OrderService {
	return &OrderService{
		db:           db,
		orderStore:   os,
		productStore: ps,
		voucherStore: vs,
	}
}

// CreateOrderRequest represents the data needed to create a new order.
type CreateOrderRequest struct {
	Items       []OrderItemRequest `json:"items"`
	VoucherCode *string            `json:"voucher_code"`
	// Delivery and Payment info would also be here
}

type OrderItemRequest struct {
	VariantID uuid.UUID `json:"variant_id"`
	Quantity  int       `json:"quantity"`
}

// CreateNewOrder orchestrates the creation of a new order.
func (s *OrderService) CreateNewOrder(req CreateOrderRequest) (*core.Order, error) {
	// 1. Start a transaction
	tx := s.db.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	// Defer rollback in case of error
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 2. Build Order and OrderDetail objects
	order := &core.Order{
		ID:          uuid.New(),
		OrderNumber: fmt.Sprintf("ORD-%d", time.Now().UnixNano()), // Simple unique order number
		OrderDate:   time.Now(),
		Status:      "PENDING", // Initial status
	}
	var orderDetails []core.OrderDetail
	subtotal := decimal.NewFromInt(0)

	for _, item := range req.Items {
		var variant core.ProductVariant
		if err := tx.First(&variant, item.VariantID).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("variant with id %s not found", item.VariantID)
		}

		if variant.Stock < item.Quantity {
			tx.Rollback()
			return nil, fmt.Errorf("not enough stock for variant %s", variant.Name)
		}

		priceAtOrder := variant.Price
		lineTotal := priceAtOrder.Mul(decimal.NewFromInt(int64(item.Quantity)))
		subtotal = subtotal.Add(lineTotal)

		orderDetails = append(orderDetails, core.OrderDetail{
			ID:           uuid.New(),
			OrderID:      order.ID,
			VariantID:    variant.ID,
			Quantity:     item.Quantity,
			PriceAtOrder: priceAtOrder,
		})

		// Decrement stock
		variant.Stock -= item.Quantity
		if err := tx.Save(variant).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
	}
	order.OrderDetails = orderDetails
	order.Subtotal = subtotal

	// 3. Apply voucher if provided
	var totalDiscount decimal.Decimal
	if req.VoucherCode != nil && *req.VoucherCode != "" {
		var voucher core.Voucher
		if err := tx.Where("code = ?", *req.VoucherCode).First(&voucher).Error; err != nil {
			tx.Rollback()
			return nil, errors.New("invalid voucher code")
		}

		if voucher.ValidUntil.Before(time.Now()) {
			tx.Rollback()
			return nil, errors.New("voucher has expired")
		}
		if voucher.Stock <= 0 {
			tx.Rollback()
			return nil, errors.New("voucher is out of stock")
		}

		if voucher.DiscountType == core.PercentageDiscount {
			// Calculate discount as percentage of subtotal
			percentage := voucher.DiscountValue.Div(decimal.NewFromInt(100))
			totalDiscount = subtotal.Mul(percentage)
		} else {
			// Fixed amount discount
			totalDiscount = voucher.DiscountValue
		}

		// Ensure discount is not more than subtotal
		if totalDiscount.GreaterThan(subtotal) {
			totalDiscount = subtotal
		}

		order.VoucherID = &voucher.ID

		// Decrement voucher stock
		voucher.Stock--
		if err := tx.Save(voucher).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
	} else {
		totalDiscount = decimal.NewFromInt(0)
	}

	order.TotalDiscount = totalDiscount
	order.TotalFinal = subtotal.Sub(totalDiscount)

	// 4. Create the order record
	if err := tx.Create(order).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// 5. Commit the transaction
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return order, nil
}
