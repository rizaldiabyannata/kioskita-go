-- Deskripsi: Membuat tabel 'users' dan 'products' sebagai fondasi awal.

CREATE TABLE "users" (
    "id" uuid PRIMARY KEY,
    "email" VARCHAR NOT NULL UNIQUE,
    "password_hash" VARCHAR NOT NULL,
    "created_at" TIMESTAMPTZ NOT NULL DEFAULT (now())
);

CREATE TABLE "products" (
    "id" uuid PRIMARY KEY,
    "name" VARCHAR NOT NULL,
    "description" TEXT,
    "price" BIGINT NOT NULL,
    "stock" INT NOT NULL DEFAULT 0,
    "created_at" TIMESTAMPTZ NOT NULL DEFAULT (now())
);

CREATE INDEX ON "users" ("email");