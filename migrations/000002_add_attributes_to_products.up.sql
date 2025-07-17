-- Deskripsi: Menambahkan kolom 'attributes' (JSONB) untuk kustomisasi produk.

ALTER TABLE products
ADD COLUMN attributes JSONB NOT NULL DEFAULT '{}'::jsonb;

-- Membuat GIN index pada kolom attributes untuk mempercepat query.
CREATE INDEX idx_products_attributes ON products USING GIN (attributes);