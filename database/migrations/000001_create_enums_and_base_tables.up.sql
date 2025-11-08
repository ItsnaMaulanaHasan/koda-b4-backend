CREATE TYPE "product_sizes" AS ENUM (
  'R',
  'L',
  'XL',
  '250gr'
);

CREATE TYPE "status" AS ENUM (
  'On Progress',
  'Shipping',
  'Done'
);

CREATE TYPE "shipping" AS ENUM (
  'Delivery',
  'Dine In',
  'Pick Up'
);

CREATE TABLE "users" (
    "id" serial PRIMARY KEY,
    "first_name" varchar(255) NOT NULL,
    "last_name" varchar(255) NOT NULL,
    "email" varchar(255) UNIQUE NOT NULL,
    "role" varchar(20) NOT NULL DEFAULT 'customer',
    "password" text NOT NULL,
    "created_at" timestamp DEFAULT(CURRENT_TIMESTAMP),
    "updated_at" timestamp DEFAULT(CURRENT_TIMESTAMP),
    "created_by" int,
    "updated_by" int
);

CREATE TABLE "categories" (
    "id" serial PRIMARY KEY,
    "name" varchar(100) UNIQUE NOT NULL,
    "created_at" timestamp DEFAULT(CURRENT_TIMESTAMP),
    "updated_at" timestamp DEFAULT(CURRENT_TIMESTAMP),
    "created_by" int,
    "updated_by" int
);

CREATE TABLE "sizes" (
    "id" serial PRIMARY KEY,
    "name" varchar(10) UNIQUE NOT NULL,
    "created_at" timestamp DEFAULT(CURRENT_TIMESTAMP),
    "updated_at" timestamp DEFAULT(CURRENT_TIMESTAMP),
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