CREATE TABLE "orders" (
    "id" uuid PRIMARY KEY,
    "user_id" uuid NOT NULL,
    "total_amount" BIGINT NOT NULL,
    "status" VARCHAR(50) NOT NULL,
    "shipping_address" JSONB,
    "created_at" TIMESTAMPTZ NOT NULL DEFAULT (now()),
    CONSTRAINT "fk_orders_user" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE RESTRICT
);

CREATE TABLE "order_items" (
    "id" uuid PRIMARY KEY,
    "order_id" uuid NOT NULL,
    "product_id" uuid NOT NULL,
    "quantity" INT NOT NULL,
    "price_at_purchase" BIGINT NOT NULL,
    "created_at" TIMESTAMPTZ NOT NULL DEFAULT (now()),
    CONSTRAINT "fk_order_items_order" FOREIGN KEY ("order_id") REFERENCES "orders" ("id") ON DELETE CASCADE,
    CONSTRAINT "fk_order_items_product" FOREIGN KEY ("product_id") REFERENCES "products" ("id") ON DELETE RESTRICT
);

CREATE INDEX "idx_orders_user_id" ON "orders" ("user_id");

CREATE INDEX "idx_order_items_order_id" ON "order_items" ("order_id");

CREATE INDEX "idx_order_items_product_id" ON "order_items" ("product_id");