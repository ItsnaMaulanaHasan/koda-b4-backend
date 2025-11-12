ALTER TABLE "payment_methods"
DROP CONSTRAINT "fk_payment_methods_updated_by";

ALTER TABLE "payment_methods"
DROP CONSTRAINT "fk_payment_methods_created_by";

ALTER TABLE "order_methods"
DROP CONSTRAINT "fk_order_methods_updated_by";

ALTER TABLE "order_methods"
DROP CONSTRAINT "fk_order_methods_created_by";

ALTER TABLE "variants" DROP CONSTRAINT "fk_variants_updated_by";

ALTER TABLE "variants" DROP CONSTRAINT "fk_variants_created_by";

ALTER TABLE "sizes" DROP CONSTRAINT "fk_sizes_updated_by";

ALTER TABLE "sizes" DROP CONSTRAINT "fk_sizes_created_by";

ALTER TABLE "categories" DROP CONSTRAINT "fk_categories_updated_by";

ALTER TABLE "categories" DROP CONSTRAINT "fk_categories_created_by";

ALTER TABLE "users" DROP CONSTRAINT "fk_users_updated_by";

ALTER TABLE "users" DROP CONSTRAINT "fk_users_created_by";

DROP TABLE "payment_methods";

DROP TABLE "order_methods";

DROP TABLE "variants";

DROP TABLE "sizes";

DROP TABLE "categories";

DROP TABLE "users";

DROP TYPE "status";