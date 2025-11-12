ALTER TABLE "product_variants"
DROP CONSTRAINT "fk_product_variants_updated_by";

ALTER TABLE "product_variants"
DROP CONSTRAINT "fk_product_variants_created_by";

ALTER TABLE "product_variants"
DROP CONSTRAINT "fk_product_variant_variant_id";

ALTER TABLE "product_variants"
DROP CONSTRAINT "fk_product_variants_product_id";

ALTER TABLE "product_categories"
DROP CONSTRAINT "fk_product_categories_updated_by";

ALTER TABLE "product_categories"
DROP CONSTRAINT "fk_product_categories_created_by";

ALTER TABLE "product_categories"
DROP CONSTRAINT "fk_product_categories_category_id";

ALTER TABLE "product_categories"
DROP CONSTRAINT "fk_product_categories_product_id";

ALTER TABLE "product_sizes"
DROP CONSTRAINT "fk_product_sizes_updated_by";

ALTER TABLE "product_sizes"
DROP CONSTRAINT "fk_product_sizes_created_by";

ALTER TABLE "product_sizes"
DROP CONSTRAINT "fk_product_sizes_size_id";

ALTER TABLE "product_sizes"
DROP CONSTRAINT "fk_product_sizes_product_id";

ALTER TABLE "product_images"
DROP CONSTRAINT "fk_product_images_updated_by";

ALTER TABLE "product_images"
DROP CONSTRAINT "fk_product_images_created_by";

ALTER TABLE "product_images"
DROP CONSTRAINT "fk_product_images_product_id";

ALTER TABLE "products" DROP CONSTRAINT "fk_products_updated_by";

ALTER TABLE "products" DROP CONSTRAINT "fk_producst_created_by";

DROP TABLE "product_variants";

DROP TABLE "product_categories";

DROP TABLE "product_sizes";

DROP TABLE "product_images";

DROP TABLE "products"