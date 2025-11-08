ALTER TABLE "carts" DROP CONSTRAINT "fk_carts_updated_by";

ALTER TABLE "carts" DROP CONSTRAINT "fk_carts_created_by";

ALTER TABLE "carts" DROP CONSTRAINT "fk_carts_product_id";

ALTER TABLE "carts" DROP CONSTRAINT "fk_carts_user_id";

ALTER TABLE "testimonies"
DROP CONSTRAINT "fk_testimonies_updated_by";

ALTER TABLE "testimonies"
DROP CONSTRAINT "fk_testimonies_created_by";

ALTER TABLE "testimonies" DROP CONSTRAINT "fk_testimonies_user_id";

ALTER TABLE "password_resets"
DROP CONSTRAINT "fk_password_resets_updated_by";

ALTER TABLE "password_resets"
DROP CONSTRAINT "fk_password_resets_created_by";

ALTER TABLE "password_resets"
DROP CONSTRAINT "fk_password_resets_user_id";

ALTER TABLE "sessions" DROP CONSTRAINT "fk_sessions_updated_by";

ALTER TABLE "sessions" DROP CONSTRAINT "fk_sessions_created_by";

ALTER TABLE "sessions" DROP CONSTRAINT "fk_sessions_user_id";

ALTER TABLE "profiles" DROP CONSTRAINT "fk_profiles_updated_by";

ALTER TABLE "profiles" DROP CONSTRAINT "fk_profiles_created_by";

ALTER TABLE "profiles" DROP CONSTRAINT "fk_profiles_user_id";

DROP TABLE "carts";

DROP TABLE "testimonies";

DROP TABLE "password_resets";

DROP TABLE "sessions";

DROP TABLE "profiles";