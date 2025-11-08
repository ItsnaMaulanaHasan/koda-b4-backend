ALTER TABLE "product_categories"
DROP CONSTRAINT "fk_product_categories_updated_by";

ALTER TABLE "product_categories"
DROP CONSTRAINT "fk_product_categories_created_by";

ALTER TABLE "product_categories"
DROP CONSTRAINT "fk_product_categories_category_id";

ALTER TABLE "product_categories"
DROP CONSTRAINT "fk_product_categories_product_id";

ALTER TABLE "size_products"
DROP CONSTRAINT "fk_size_products_updated_by";

ALTER TABLE "size_products"
DROP CONSTRAINT "fk_size_products_created_by";

ALTER TABLE "size_products"
DROP CONSTRAINT "fk_size_products_size_id";

ALTER TABLE "size_products"
DROP CONSTRAINT "fk_size_products_product_id";

ALTER TABLE "product_images"
DROP CONSTRAINT "fk_product_images_updated_by";

ALTER TABLE "product_images"
DROP CONSTRAINT "fk_product_images_created_by";

ALTER TABLE "product_images"
DROP CONSTRAINT "fk_product_images_product_id";

ALTER TABLE "products" DROP CONSTRAINT "fk_products_updated_by";

ALTER TABLE "products" DROP CONSTRAINT "fk_producst_created_by";

DROP TABLE "product_categories";

DROP TABLE "size_products";

DROP TABLE "product_images";

DROP TABLE "products"