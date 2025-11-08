CREATE TABLE "orders" (
    "id" serial PRIMARY KEY,
    "user_id" int,
    "date_order" timestamp DEFAULT(CURRENT_TIMESTAMP),
    "full_name" varchar(255) NOT NULL,
    "email" varchar(255) NOT NULL,
    "address" varchar(255) NOT NULL,
    "phone" varchar(20) NOT NULL,
    "payment_method" varchar(50),
    "shipping" shipping DEFAULT 'Dine In',
    "status" status DEFAULT 'On Progress',
    "total_transaction" numeric(10, 2) NOT NULL CHECK ("total_transaction" > 0),
    "delivery_fee" numeric(10, 2) DEFAULT 0,
    "tax" numeric(10, 2) DEFAULT 0,
    "created_at" timestamp DEFAULT(CURRENT_TIMESTAMP),
    "updated_at" timestamp DEFAULT(CURRENT_TIMESTAMP),
    "created_by" int,
    "updated_by" int
);

CREATE TABLE "ordered_products" (
    "id" serial PRIMARY KEY,
    "order_id" int,
    "product_id" int,
    "product_name" varchar(255) NOT NULL,
    "product_price" numeric(10, 2) NOT NULL CHECK ("product_price" > 0),
    "discount_percent" numeric(5, 2),
    "amount" int NOT NULL CHECK ("amount" > 0),
    "subtotal" numeric(10, 2) NOT NULL,
    "size" product_sizes DEFAULT 'L',
    "created_at" timestamp DEFAULT(CURRENT_TIMESTAMP),
    "updated_at" timestamp DEFAULT(CURRENT_TIMESTAMP),
    "created_by" int,
    "updated_by" int
);

CREATE TABLE "coupons" (
    "id" serial PRIMARY KEY,
    "title" varchar(255) UNIQUE NOT NULL,
    "description" text NOT NULL,
    "discount_percent" numeric(5, 2) NOT NULL,
    "min_purchase" numeric(10, 2),
    "image" text NOT NULL,
    "bg_color" varchar(20) NOT NULL,
    "valid_until" timestamp,
    "is_active" bool DEFAULT true,
    "created_at" timestamp DEFAULT(CURRENT_TIMESTAMP),
    "updated_at" timestamp DEFAULT(CURRENT_TIMESTAMP),
    "created_by" int,
    "updated_by" int
);

CREATE TABLE "coupon_usage" (
    "id" serial PRIMARY KEY,
    "user_id" int,
    "coupon_id" int,
    "order_id" int,
    "discount_amount" numeric(10, 2),
    "used_at" timestamp DEFAULT(CURRENT_TIMESTAMP),
    "updated_at" timestamp DEFAULT(CURRENT_TIMESTAMP),
    "created_by" int,
    "updated_by" int
);

ALTER TABLE "orders"
ADD CONSTRAINT "fk_orders_user_id" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE RESTRICT;

ALTER TABLE "orders"
ADD CONSTRAINT "fk_orders_created_by" FOREIGN KEY ("created_by") REFERENCES "users" ("id");

ALTER TABLE "orders"
ADD CONSTRAINT "fk_orders_updated_by" FOREIGN KEY ("updated_by") REFERENCES "users" ("id");

ALTER TABLE "ordered_products"
ADD CONSTRAINT "fk_ordered_products_order_id" FOREIGN KEY ("order_id") REFERENCES "orders" ("id");

ALTER TABLE "ordered_products"
ADD CONSTRAINT "fk_ordered_products_product_id" FOREIGN KEY ("product_id") REFERENCES "products" ("id");

ALTER TABLE "ordered_products"
ADD CONSTRAINT "fk_ordered_products_created_by" FOREIGN KEY ("created_by") REFERENCES "users" ("id");

ALTER TABLE "ordered_products"
ADD CONSTRAINT "fk_ordered_products_updated_by" FOREIGN KEY ("updated_by") REFERENCES "users" ("id");

ALTER TABLE "coupons"
ADD CONSTRAINT "fk_coupons_created_by" FOREIGN KEY ("created_by") REFERENCES "users" ("id");

ALTER TABLE "coupons"
ADD CONSTRAINT "fk_coupons_updated_by" FOREIGN KEY ("updated_by") REFERENCES "users" ("id");

ALTER TABLE "coupon_usage"
ADD CONSTRAINT "fk_coupon_usage_user_id" FOREIGN KEY ("user_id") REFERENCES "users" ("id");

ALTER TABLE "coupon_usage"
ADD CONSTRAINT "fk_coupon_usage_coupon_id" FOREIGN KEY ("coupon_id") REFERENCES "coupons" ("id");

ALTER TABLE "coupon_usage"
ADD CONSTRAINT "fk_coupon_usage_order_id" FOREIGN KEY ("order_id") REFERENCES "orders" ("id");

ALTER TABLE "coupon_usage"
ADD CONSTRAINT "fk_coupon_usage_created_by" FOREIGN KEY ("created_by") REFERENCES "users" ("id");

ALTER TABLE "coupon_usage"
ADD CONSTRAINT "fk_coupon_usage_updated_by" FOREIGN KEY ("updated_by") REFERENCES "users" ("id");