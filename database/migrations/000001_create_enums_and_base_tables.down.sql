ALTER TABLE "sizes" DROP CONSTRAINT "fk_sizes_updated_by"

ALTER TABLE "sizes" DROP CONSTRAINT "fk_sizes_created_by";

ALTER TABLE "categories" DROP CONSTRAINT "fk_categories_updated_by";

ALTER TABLE "categories" DROP CONSTRAINT "fk_categories_created_by";

ALTER TABLE "users" DROP CONSTRAINT "fk_users_updated_by";

ALTER TABLE "users" DROP CONSTRAINT "fk_users_created_by";

DROP TABLE "sizes";

DROP TABLE "categories";

DROP TABLE "users";

DROP TYPE "shipping";

DROP TYPE "status";

DROP TYPE "product_sizes";