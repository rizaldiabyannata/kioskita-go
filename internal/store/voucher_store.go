package store

import (
	"github.com/rizaldiabyannata/kioskita-go/internal/core"
	"gorm.io/gorm"
)

// VoucherStore handles database operations for Vouchers.
type VoucherStore struct {
	db *gorm.DB
}

// NewVoucherStore creates a new VoucherStore.
func NewVoucherStore(db *gorm.DB) *VoucherStore {
	return &VoucherStore{db: db}
}

// Create creates a new voucher.
func (s *VoucherStore) Create(voucher *core.Voucher) error {
	return s.db.Create(voucher).Error
}

// FindByCode retrieves a voucher by its code.
func (s *VoucherStore) FindByCode(code string) (*core.Voucher, error) {
	var voucher core.Voucher
	if err := s.db.Where("code = ?", code).First(&voucher).Error; err != nil {
		return nil, err
	}
	return &voucher, nil
}
