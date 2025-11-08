CREATE TABLE "profiles" (
    "id" serial PRIMARY KEY,
    "user_id" int,
    "image" text,
    "address" varchar(255),
    "phone_number" varchar(20),
    "created_at" timestamp DEFAULT(CURRENT_TIMESTAMP),
    "updated_at" timestamp DEFAULT(CURRENT_TIMESTAMP),
    "created_by" int,
    "updated_by" int
);

CREATE TABLE "sessions" (
    "id" serial PRIMARY KEY,
    "user_id" int,
    "session_token" text UNIQUE NOT NULL,
    "login_time" timestamp DEFAULT(CURRENT_TIMESTAMP),
    "expired_at" timestamp,
    "ip_address" varchar(30),
    "device" varchar(255),
    "is_active" bool DEFAULT true,
    "created_at" timestamp DEFAULT(CURRENT_TIMESTAMP),
    "updated_at" timestamp DEFAULT(CURRENT_TIMESTAMP),
    "created_by" int,
    "updated_by" int
);

CREATE TABLE "password_resets" (
    "id" serial PRIMARY KEY,
    "user_id" int,
    "token_reset" char(6) UNIQUE NOT NULL,
    "expired_at" timestamp NOT NULL DEFAULT(
        CURRENT_TIMESTAMP + INTERVAL '1 hour'
    ),
    "created_at" timestamp DEFAULT(CURRENT_TIMESTAMP),
    "updated_at" timestamp DEFAULT(CURRENT_TIMESTAMP),
    "created_by" int,
    "updated_by" int
);

CREATE TABLE "testimonies" (
    "id" serial PRIMARY KEY,
    "user_id" int,
    "position" varchar(100),
    "rating" numeric(2, 1) DEFAULT 5 CHECK (
        "rating" >= 0
        AND "rating" <= 5
    ),
    "testimonial" text,
    "created_at" timestamp DEFAULT(CURRENT_TIMESTAMP),
    "updated_at" timestamp DEFAULT(CURRENT_TIMESTAMP),
    "created_by" int,
    "updated_by" int
);

CREATE TABLE "carts" (
    "id" serial PRIMARY KEY,
    "user_id" int,
    "product_id" int,
    "amount" int CHECK ("amount" > 0),
    "subtotal" numeric(10, 2),
    "size" product_sizes DEFAULT 'R',
    "created_at" timestamp DEFAULT(CURRENT_TIMESTAMP),
    "updated_at" timestamp DEFAULT(CURRENT_TIMESTAMP),
    "created_by" int,
    "updated_by" int
);

ALTER TABLE "profiles"
ADD CONSTRAINT "fk_profiles_user_id" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE CASCADE;

ALTER TABLE "profiles"
ADD CONSTRAINT "fk_profiles_created_by" FOREIGN KEY ("created_by") REFERENCES "users" ("id");

ALTER TABLE "profiles"
ADD CONSTRAINT "fk_profiles_updated_by" FOREIGN KEY ("updated_by") REFERENCES "users" ("id");

ALTER TABLE "sessions"
ADD CONSTRAINT "fk_sessions_user_id" FOREIGN KEY ("user_id") REFERENCES "users" ("id");

ALTER TABLE "sessions"
ADD CONSTRAINT "fk_sessions_created_by" FOREIGN KEY ("created_by") REFERENCES "users" ("id");

ALTER TABLE "sessions"
ADD CONSTRAINT "fk_sessions_updated_by" FOREIGN KEY ("updated_by") REFERENCES "users" ("id");

ALTER TABLE "password_resets"
ADD CONSTRAINT "fk_password_resets_user_id" FOREIGN KEY ("user_id") REFERENCES "users" ("id");

ALTER TABLE "password_resets"
ADD CONSTRAINT "fk_password_resets_created_by" FOREIGN KEY ("created_by") REFERENCES "users" ("id");

ALTER TABLE "password_resets"
ADD CONSTRAINT "fk_password_resets_updated_by" FOREIGN KEY ("updated_by") REFERENCES "users" ("id");

ALTER TABLE "testimonies"
ADD CONSTRAINT "fk_testimonies_user_id" FOREIGN KEY ("user_id") REFERENCES "users" ("id");

ALTER TABLE "testimonies"
ADD CONSTRAINT "fk_testimonies_created_by" FOREIGN KEY ("created_by") REFERENCES "users" ("id");

ALTER TABLE "testimonies"
ADD CONSTRAINT "fk_testimonies_updated_by" FOREIGN KEY ("updated_by") REFERENCES "users" ("id");

ALTER TABLE "carts"
ADD CONSTRAINT "fk_carts_user_id" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE CASCADE;

ALTER TABLE "carts"
ADD CONSTRAINT "fk_carts_product_id" FOREIGN KEY ("product_id") REFERENCES "products" ("id");

ALTER TABLE "carts"
ADD CONSTRAINT "fk_carts_created_by" FOREIGN KEY ("created_by") REFERENCES "users" ("id");

ALTER TABLE "carts"
ADD CONSTRAINT "fk_carts_updated_by" FOREIGN KEY ("updated_by") REFERENCES "users" ("id");