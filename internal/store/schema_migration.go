package store

import (
	"github.com/rizaldiabyannata/kioskita-go/internal/core"
	"gorm.io/gorm"
)

// AutoMigrateDB runs the GORM auto-migration for all the application's models.
// This will create or update tables, columns, and foreign keys as needed.
func AutoMigrateDB(db *gorm.DB) error {
	return db.AutoMigrate(
		&core.Admin{},
		&core.Category{},
		&core.Product{},
		&core.ProductImage{},
		&core.ProductVariant{},
		&core.Voucher{},
		&core.Order{},
		&core.OrderDetail{},
		&core.Payment{},
		&core.Delivery{},
	)
}
