CREATE TABLE "products" (
    "id" serial PRIMARY KEY,
    "name" varchar(255) UNIQUE NOT NULL,
    "description" text NOT NULL,
    "price" numeric(10, 2) NOT NULL CHECK ("price" > 0),
    "discount_percent" numeric(5, 2) DEFAULT 0,
    "rating" numeric(2, 1) DEFAULT 5 CHECK (
        "rating" >= 0
        AND "rating" <= 5
    ),
    "is_flash_sale" bool DEFAULT false,
    "stock" int CHECK ("stock" >= 0),
    "is_active" bool DEFAULT true,
    "is_favourite" bool DEFAULT false,
    "created_at" timestamp DEFAULT (CURRENT_TIMESTAMP),
    "updated_at" timestamp DEFAULT (CURRENT_TIMESTAMP),
    "created_by" int,
    "updated_by" int
);

CREATE TABLE "product_images" (
    "id" serial PRIMARY KEY,
    "product_id" int,
    "image" text NOT NULL,
    "created_at" timestamp DEFAULT (CURRENT_TIMESTAMP),
    "updated_at" timestamp DEFAULT (CURRENT_TIMESTAMP),
    "created_by" int,
    "updated_by" int
);

CREATE TABLE "product_sizes" (
    "id" serial PRIMARY KEY,
    "product_id" int,
    "size_id" int,
    "created_at" timestamp DEFAULT (CURRENT_TIMESTAMP),
    "updated_at" timestamp DEFAULT (CURRENT_TIMESTAMP),
    "created_by" int,
    "updated_by" int
);

CREATE TABLE "product_categories" (
    "id" serial PRIMARY KEY,
    "product_id" int,
    "category_id" int,
    "created_at" timestamp DEFAULT (CURRENT_TIMESTAMP),
    "updated_at" timestamp DEFAULT (CURRENT_TIMESTAMP),
    "created_by" int,
    "updated_by" int
);

CREATE TABLE "product_variants" (
    "id" serial PRIMARY KEY,
    "product_id" int,
    "variant_id" int,
    "created_at" timestamp DEFAULT (CURRENT_TIMESTAMP),
    "updated_at" timestamp DEFAULT (CURRENT_TIMESTAMP),
    "created_by" int,
    "updated_by" int
);

ALTER TABLE "products"
ADD CONSTRAINT "fk_producst_created_by" FOREIGN KEY ("created_by") REFERENCES "users" ("id");

ALTER TABLE "products"
ADD CONSTRAINT "fk_products_updated_by" FOREIGN KEY ("updated_by") REFERENCES "users" ("id");

ALTER TABLE "product_images"
ADD CONSTRAINT "fk_product_images_product_id" FOREIGN KEY ("product_id") REFERENCES "products" ("id") ON DELETE CASCADE;

ALTER TABLE "product_images"
ADD CONSTRAINT "fk_product_images_created_by" FOREIGN KEY ("created_by") REFERENCES "users" ("id");

ALTER TABLE "product_images"
ADD CONSTRAINT "fk_product_images_updated_by" FOREIGN KEY ("updated_by") REFERENCES "users" ("id");

ALTER TABLE "product_sizes"
ADD CONSTRAINT "fk_product_sizes_product_id" FOREIGN KEY ("product_id") REFERENCES "products" ("id") ON DELETE CASCADE;

ALTER TABLE "product_sizes"
ADD CONSTRAINT "fk_product_sizes_size_id" FOREIGN KEY ("size_id") REFERENCES "sizes" ("id");

ALTER TABLE "product_sizes"
ADD CONSTRAINT "fk_product_sizes_created_by" FOREIGN KEY ("created_by") REFERENCES "users" ("id");

ALTER TABLE "product_sizes"
ADD CONSTRAINT "fk_product_sizes_updated_by" FOREIGN KEY ("updated_by") REFERENCES "users" ("id");

ALTER TABLE "product_categories"
ADD CONSTRAINT "fk_product_categories_product_id" FOREIGN KEY ("product_id") REFERENCES "products" ("id") ON DELETE CASCADE;

ALTER TABLE "product_categories"
ADD CONSTRAINT "fk_product_categories_category_id" FOREIGN KEY ("category_id") REFERENCES "categories" ("id");

ALTER TABLE "product_categories"
ADD CONSTRAINT "fk_product_categories_created_by" FOREIGN KEY ("created_by") REFERENCES "users" ("id");

ALTER TABLE "product_categories"
ADD CONSTRAINT "fk_product_categories_updated_by" FOREIGN KEY ("updated_by") REFERENCES "users" ("id");

ALTER TABLE "product_variants"
ADD CONSTRAINT "fk_product_variants_product_id" FOREIGN KEY ("product_id") REFERENCES "products" ("id") ON DELETE CASCADE;

ALTER TABLE "product_variants"
ADD CONSTRAINT "fk_product_variant_variant_id" FOREIGN KEY ("variant_id") REFERENCES "variants" ("id");

ALTER TABLE "product_variants"
ADD CONSTRAINT "fk_product_variants_created_by" FOREIGN KEY ("created_by") REFERENCES "users" ("id");

ALTER TABLE "product_variants"
ADD CONSTRAINT "fk_product_variants_updated_by" FOREIGN KEY ("updated_by") REFERENCES "users" ("id");