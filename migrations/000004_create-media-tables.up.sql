CREATE TABLE "media" (
    "id" uuid PRIMARY KEY,
    "product_id" uuid NOT NULL,
    "url" VARCHAR(255) NOT NULL,
    "type" VARCHAR(50) NOT NULL,
    "created_at" TIMESTAMPTZ NOT NULL DEFAULT (now()),
    CONSTRAINT "fk_media_product" FOREIGN KEY ("product_id") REFERENCES "products" ("id") ON DELETE CASCADE
);

CREATE INDEX "idx_media_product_id" ON "media" ("product_id");