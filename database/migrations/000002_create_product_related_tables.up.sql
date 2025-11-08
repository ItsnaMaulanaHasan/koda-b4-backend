CREATE TABLE "products" (
    "id" serial PRIMARY KEY,
    "name" varchar(255) UNIQUE NOT NULL,
    "description" text NOT NULL,
    "price" numeric(10, 2) NOT NULL CHECK ("price" > 0),
    "discount_percent" numeric(5, 2),
    "rating" numeric(2, 1) DEFAULT 5 CHECK (
        "rating" >= 0
        AND "rating" <= 5
    ),
    "is_flash_sale" bool DEFAULT false,
    "stock" int CHECK ("stock" >= 0),
    "is_active" bool DEFAULT true,
    "created_at" timestamp DEFAULT(CURRENT_TIMESTAMP),
    "updated_at" timestamp DEFAULT(CURRENT_TIMESTAMP),
    "created_by" int,
    "updated_by" int
);

CREATE TABLE "product_images" (
    "id" serial PRIMARY KEY,
    "image" text NOT NULL,
    "product_id" int,
    "created_at" timestamp DEFAULT(CURRENT_TIMESTAMP),
    "updated_at" timestamp DEFAULT(CURRENT_TIMESTAMP),
    "created_by" int,
    "updated_by" int
);

CREATE TABLE "size_products" (
    "id" serial PRIMARY KEY,
    "product_id" int,
    "size_id" int,
    "created_at" timestamp DEFAULT(CURRENT_TIMESTAMP),
    "updated_at" timestamp DEFAULT(CURRENT_TIMESTAMP),
    "created_by" int,
    "updated_by" int
);

CREATE TABLE "product_categories" (
    "id" serial PRIMARY KEY,
    "product_id" int,
    "category_id" int,
    "created_at" timestamp DEFAULT(CURRENT_TIMESTAMP),
    "updated_at" timestamp DEFAULT(CURRENT_TIMESTAMP),
    "created_by" int,
    "updated_by" int
);

ALTER TABLE "products"
ADD CONSTRAINT "fk_producst_created_by" FOREIGN KEY ("created_by") REFERENCES "users" ("id");

ALTER TABLE "products"
ADD CONSTRAINT "fk_products_updated_by" FOREIGN KEY ("updated_by") REFERENCES "users" ("id");

ALTER TABLE "product_images"
ADD CONSTRAINT "fk_product_images_product_id" FOREIGN KEY ("product_id") REFERENCES "products" ("id");

ALTER TABLE "product_images"
ADD CONSTRAINT "fk_product_images_created_by" FOREIGN KEY ("created_by") REFERENCES "users" ("id");

ALTER TABLE "product_images"
ADD CONSTRAINT "fk_product_images_updated_by" FOREIGN KEY ("updated_by") REFERENCES "users" ("id");

ALTER TABLE "size_products"
ADD CONSTRAINT "fk_size_products_product_id" FOREIGN KEY ("product_id") REFERENCES "products" ("id");

ALTER TABLE "size_products"
ADD CONSTRAINT "fk_size_products_size_id" FOREIGN KEY ("size_id") REFERENCES "sizes" ("id");

ALTER TABLE "size_products"
ADD CONSTRAINT "fk_size_products_created_by" FOREIGN KEY ("created_by") REFERENCES "users" ("id");

ALTER TABLE "size_products"
ADD CONSTRAINT "fk_size_products_updated_by" FOREIGN KEY ("updated_by") REFERENCES "users" ("id");

ALTER TABLE "product_categories"
ADD CONSTRAINT "fk_product_categories_product_id" FOREIGN KEY ("product_id") REFERENCES "products" ("id");

ALTER TABLE "product_categories"
ADD CONSTRAINT "fk_product_categories_category_id" FOREIGN KEY ("category_id") REFERENCES "categories" ("id");

ALTER TABLE "product_categories"
ADD CONSTRAINT "fk_product_categories_created_by" FOREIGN KEY ("created_by") REFERENCES "users" ("id");

ALTER TABLE "product_categories"
ADD CONSTRAINT "fk_product_categories_updated_by" FOREIGN KEY ("updated_by") REFERENCES "users" ("id");