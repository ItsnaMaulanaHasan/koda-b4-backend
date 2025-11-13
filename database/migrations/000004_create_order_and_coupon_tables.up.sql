CREATE TABLE "transactions" (
    "id" serial PRIMARY KEY,
    "user_id" int,
    "no_invoice" varchar(255) NOT NULL,
    "date_transaction" timestamp DEFAULT (CURRENT_TIMESTAMP),
    "full_name" varchar(255) NOT NULL,
    "email" varchar(255) NOT NULL,
    "address" varchar(255) NOT NULL,
    "phone" varchar(20) NOT NULL,
    "payment_method_id" int,
    "order_method_id" int,
    "status_id" int DEFAULT 1,
    "delivery_fee" numeric(10, 2) DEFAULT 0,
    "admin_fee" numeric(10, 2) DEFAULT 0,
    "tax" numeric(10, 2) DEFAULT 0,
    "total_transaction" numeric(10, 2) NOT NULL DEFAULT 0,
    "created_at" timestamp DEFAULT (CURRENT_TIMESTAMP),
    "updated_at" timestamp DEFAULT (CURRENT_TIMESTAMP),
    "created_by" int,
    "updated_by" int
);

CREATE TABLE "transaction_items" (
    "id" serial PRIMARY KEY,
    "transaction_id" int,
    "product_id" int,
    "product_name" varchar(255) NOT NULL,
    "product_price" numeric(10, 2) NOT NULL CHECK ("product_price" > 0),
    "discount_percent" numeric(5, 2) DEFAULT 0,
    "discount_price" numeric(10, 2) DEFAULT 0,
    "size" varchar(10),
    "size_cost" numeric(10, 2) DEFAULT 0,
    "variant" varchar(50),
    "variant_cost" numeric(10, 2) DEFAULT 0,
    "amount" int NOT NULL CHECK ("amount" > 0),
    "subtotal" numeric(10, 2) NOT NULL DEFAULT 0,
    "created_at" timestamp DEFAULT (CURRENT_TIMESTAMP),
    "updated_at" timestamp DEFAULT (CURRENT_TIMESTAMP),
    "created_by" int,
    "updated_by" int
);

CREATE TABLE "coupons" (
    "id" serial PRIMARY KEY,
    "title" varchar(255) UNIQUE NOT NULL,
    "description" text NOT NULL,
    "discount_percent" numeric(5, 2) NOT NULL DEFAULT 0,
    "min_purchase" numeric(10, 2) DEFAULT 0,
    "image" text NOT NULL,
    "bg_color" varchar(20) NOT NULL,
    "valid_until" timestamp,
    "is_active" bool DEFAULT true,
    "created_at" timestamp DEFAULT (CURRENT_TIMESTAMP),
    "updated_at" timestamp DEFAULT (CURRENT_TIMESTAMP),
    "created_by" int,
    "updated_by" int
);

CREATE TABLE "coupon_usage" (
    "id" serial PRIMARY KEY,
    "user_id" int,
    "coupon_id" int,
    "transaction_id" int,
    "discount_amount" numeric(10, 2) DEFAULT 0,
    "used_at" timestamp DEFAULT (CURRENT_TIMESTAMP),
    "updated_at" timestamp DEFAULT (CURRENT_TIMESTAMP),
    "created_by" int,
    "updated_by" int
);

ALTER TABLE "transactions"
ADD CONSTRAINT "fk_transactions_user_id" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE RESTRICT;

ALTER TABLE "transactions"
ADD CONSTRAINT "fk_transactions_payment_method_id" FOREIGN KEY ("payment_method_id") REFERENCES "payment_methods" ("id");

ALTER TABLE "transactions"
ADD CONSTRAINT "fk_transactions_order_method_id" FOREIGN KEY ("order_method_id") REFERENCES "order_methods" ("id");

ALTER TABLE "transactions"
ADD CONSTRAINT "fk_transactions_status_id" FOREIGN KEY ("status_id") REFERENCES "status" ("id");

ALTER TABLE "transactions"
ADD CONSTRAINT "fk_transactions_created_by" FOREIGN KEY ("created_by") REFERENCES "users" ("id");

ALTER TABLE "transactions"
ADD CONSTRAINT "fk_transactions_updated_by" FOREIGN KEY ("updated_by") REFERENCES "users" ("id");

ALTER TABLE "transaction_items"
ADD CONSTRAINT "fk_transaction_items_transaction_id" FOREIGN KEY ("transaction_id") REFERENCES "transactions" ("id");

ALTER TABLE "transaction_items"
ADD CONSTRAINT "fk_transaction_items_product_id" FOREIGN KEY ("product_id") REFERENCES "products" ("id");

ALTER TABLE "transaction_items"
ADD CONSTRAINT "fk_transaction_items_created_by" FOREIGN KEY ("created_by") REFERENCES "users" ("id");

ALTER TABLE "transaction_items"
ADD CONSTRAINT "fk_transaction_items_updated_by" FOREIGN KEY ("updated_by") REFERENCES "users" ("id");

ALTER TABLE "coupons"
ADD CONSTRAINT "fk_coupons_created_by" FOREIGN KEY ("created_by") REFERENCES "users" ("id");

ALTER TABLE "coupons"
ADD CONSTRAINT "fk_coupons_updated_by" FOREIGN KEY ("updated_by") REFERENCES "users" ("id");

ALTER TABLE "coupon_usage"
ADD CONSTRAINT "fk_coupon_usage_user_id" FOREIGN KEY ("user_id") REFERENCES "users" ("id");

ALTER TABLE "coupon_usage"
ADD CONSTRAINT "fk_coupon_usage_coupon_id" FOREIGN KEY ("coupon_id") REFERENCES "coupons" ("id");

ALTER TABLE "coupon_usage"
ADD CONSTRAINT "fk_coupon_usage_transaction_id" FOREIGN KEY ("transaction_id") REFERENCES "transactions" ("id");

ALTER TABLE "coupon_usage"
ADD CONSTRAINT "fk_coupon_usage_created_by" FOREIGN KEY ("created_by") REFERENCES "users" ("id");

ALTER TABLE "coupon_usage"
ADD CONSTRAINT "fk_coupon_usage_updated_by" FOREIGN KEY ("updated_by") REFERENCES "users" ("id");