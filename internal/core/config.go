package core

// AttributeSchema mendefinisikan satu field kustom untuk produk.
type AttributeSchema struct {
	Key      string   `json:"key"`
	Label    string   `json:"label"`
	Type     string   `json:"type"` // "text", "number", "boolean", "select"
	Options  []string `json:"options,omitempty"`
	Required bool     `json:"required"`
}

// StoreConfig adalah struktur utama untuk file store_config.json.
type StoreConfig struct {
	StoreName     string            `json:"storeName"`
	BusinessType  string            `json:"businessType"`
	SetupComplete bool              `json:"setupComplete"`
	ProductSchema []AttributeSchema `json:"productSchema"`
}

// AdminSetupPayload adalah data yang dibutuhkan untuk membuat admin pertama.
type AdminSetupPayload struct {
	AdminEmail    string `json:"adminEmail" binding:"required,email"`
	AdminPassword string `json:"adminPassword" binding:"required,min=6"`
}

// StoreSetupPayload sekarang hanya butuh ID tipe bisnis.
type StoreSetupPayload struct {
	StoreName      string `json:"storeName" binding:"required"`
	BusinessTypeID string `json:"businessTypeID" binding:"required"`
}
