CREATE TABLE "users" (
    "id" serial PRIMARY KEY,
    "first_name" varchar(255) NOT NULL,
    "last_name" varchar(255) NOT NULL,
    "email" varchar(255) UNIQUE NOT NULL,
    "role" varchar(20) NOT NULL DEFAULT 'customer',
    "password" text NOT NULL,
    "created_at" timestamp DEFAULT (CURRENT_TIMESTAMP),
    "updated_at" timestamp DEFAULT (CURRENT_TIMESTAMP),
    "created_by" int,
    "updated_by" int
);

CREATE TABLE "categories" (
    "id" serial PRIMARY KEY,
    "name" varchar(100) UNIQUE NOT NULL,
    "created_at" timestamp DEFAULT (CURRENT_TIMESTAMP),
    "updated_at" timestamp DEFAULT (CURRENT_TIMESTAMP),
    "created_by" int,
    "updated_by" int
);

CREATE TABLE "sizes" (
    "id" serial PRIMARY KEY,
    "name" varchar(10) UNIQUE NOT NULL,
    "size_cost" numeric(10, 2) DEFAULT 0,
    "created_at" timestamp DEFAULT (CURRENT_TIMESTAMP),
    "updated_at" timestamp DEFAULT (CURRENT_TIMESTAMP),
    "created_by" int,
    "updated_by" int
);

CREATE TABLE "variants" (
    "id" serial PRIMARY KEY,
    "name" varchar(50) UNIQUE NOT NULL,
    "variant_cost" numeric(10, 2) DEFAULT 0,
    "created_at" timestamp DEFAULT (CURRENT_TIMESTAMP),
    "updated_at" timestamp DEFAULT (CURRENT_TIMESTAMP),
    "created_by" int,
    "updated_by" int
);

CREATE TABLE "order_methods" (
    "id" serial PRIMARY KEY,
    "name" varchar(30) UNIQUE NOT NULL,
    "delivery_fee" numeric(10, 2) DEFAULT 0,
    "created_at" timestamp DEFAULT (CURRENT_TIMESTAMP),
    "updated_at" timestamp DEFAULT (CURRENT_TIMESTAMP),
    "created_by" int,
    "updated_by" int
);

CREATE TABLE "payment_methods" (
    "id" serial PRIMARY KEY,
    "name" varchar(30) UNIQUE NOT NULL,
    "admin_fee" numeric(10, 2) DEFAULT 0,
    "created_at" timestamp DEFAULT (CURRENT_TIMESTAMP),
    "updated_at" timestamp DEFAULT (CURRENT_TIMESTAMP),
    "created_by" int,
    "updated_by" int
);

CREATE TABLE "status" (
    "id" serial PRIMARY KEY,
    "name" varchar(30) UNIQUE DEFAULT 'On Progress',
    "created_at" timestamp DEFAULT (CURRENT_TIMESTAMP),
    "updated_at" timestamp DEFAULT (CURRENT_TIMESTAMP),
    "created_by" int,
    "updated_by" int
);

ALTER TABLE "users"
ADD CONSTRAINT "fk_users_created_by" FOREIGN KEY ("created_by") REFERENCES "users" ("id");

ALTER TABLE "users"
ADD CONSTRAINT "fk_users_updated_by" FOREIGN KEY ("updated_by") REFERENCES "users" ("id");

ALTER TABLE "categories"
ADD CONSTRAINT "fk_categories_created_by" FOREIGN KEY ("created_by") REFERENCES "users" ("id");

ALTER TABLE "categories"
ADD CONSTRAINT "fk_categories_updated_by" FOREIGN KEY ("updated_by") REFERENCES "users" ("id");

ALTER TABLE "sizes"
ADD CONSTRAINT "fk_sizes_created_by" FOREIGN KEY ("created_by") REFERENCES "users" ("id");

ALTER TABLE "sizes"
ADD CONSTRAINT "fk_sizes_updated_by" FOREIGN KEY ("updated_by") REFERENCES "users" ("id");

ALTER TABLE "variants"
ADD CONSTRAINT "fk_variants_created_by" FOREIGN KEY ("created_by") REFERENCES "users" ("id");

ALTER TABLE "variants"
ADD CONSTRAINT "fk_variants_updated_by" FOREIGN KEY ("updated_by") REFERENCES "users" ("id");

ALTER TABLE "order_methods"
ADD CONSTRAINT "fk_order_methods_created_by" FOREIGN KEY ("created_by") REFERENCES "users" ("id");

ALTER TABLE "order_methods"
ADD CONSTRAINT "fk_order_methods_updated_by" FOREIGN KEY ("updated_by") REFERENCES "users" ("id");

ALTER TABLE "payment_methods"
ADD CONSTRAINT "fk_payment_methods_created_by" FOREIGN KEY ("created_by") REFERENCES "users" ("id");

ALTER TABLE "payment_methods"
ADD CONSTRAINT "fk_payment_methods_updated_by" FOREIGN KEY ("updated_by") REFERENCES "users" ("id");

ALTER TABLE "status"
ADD CONSTRAINT "fk_status_created_by" FOREIGN KEY ("created_by") REFERENCES "users" ("id");

ALTER TABLE "status"
ADD CONSTRAINT "fk_status_updated_by" FOREIGN KEY ("updated_by") REFERENCES "users" ("id");