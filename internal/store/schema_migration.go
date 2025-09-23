package store

import (
	"log"

	"github.com/rizaldiabyannata/kioskita-go/internal/core"
	"gorm.io/gorm"
)

// SetupSchema ensures database schema matches the selected business type.
// It will:
// - AutoMigrate core tables
// - Create the subtype table for the selected type
// - Drop legacy JSON attributes column if present
// - Optionally drop the other subtype table if exists (to keep schema lean)
func SetupSchema(db *gorm.DB, cfg core.StoreConfig) error {
	// Always migrate core tables
	if err := db.AutoMigrate(&core.User{}, &core.Product{}, &core.Order{}, &core.OrderItem{}, &core.Media{}); err != nil {
		return err
	}

	// Remove legacy attributes column from products if exists
	if db.Migrator().HasColumn(&core.Product{}, "attributes") {
		if err := db.Migrator().DropColumn(&core.Product{}, "attributes"); err != nil {
			log.Printf("warning: failed to drop legacy column products.attributes: %v", err)
		}
	}

	// Remove legacy JSON shipping_address column if exists
	if db.Migrator().HasColumn(&core.Order{}, "shipping_address") {
		if err := db.Migrator().DropColumn(&core.Order{}, "shipping_address"); err != nil {
			log.Printf("warning: failed to drop legacy column orders.shipping_address: %v", err)
		}
	}

	// Create table for selected subtype; drop the other one to avoid confusion
	switch cfg.BusinessTypeID {
	case "clothing":
		if err := db.AutoMigrate(&core.ClothingProduct{}); err != nil {
			return err
		}
		if db.Migrator().HasTable(&core.FoodProduct{}) {
			if err := db.Migrator().DropTable(&core.FoodProduct{}); err != nil {
				log.Printf("warning: failed to drop table food_products: %v", err)
			}
		}
	case "cafe", "food":
		if err := db.AutoMigrate(&core.FoodProduct{}); err != nil {
			return err
		}
		if db.Migrator().HasTable(&core.ClothingProduct{}) {
			if err := db.Migrator().DropTable(&core.ClothingProduct{}); err != nil {
				log.Printf("warning: failed to drop table clothing_products: %v", err)
			}
		}
	default:
		// If unknown, do nothing extra; keep core tables only
		log.Printf("info: unknown BusinessTypeID '%s', migrated core tables only", cfg.BusinessTypeID)
	}
	return nil
}
