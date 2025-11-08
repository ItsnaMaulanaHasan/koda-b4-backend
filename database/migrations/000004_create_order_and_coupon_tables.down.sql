ALTER TABLE "coupon_usage" DROP CONSTRAINT "fk_coupon_usage_user_id";

ALTER TABLE "coupon_usage"
DROP CONSTRAINT "fk_coupon_usage_coupon_id";

ALTER TABLE "coupon_usage"
DROP CONSTRAINT "fk_coupon_usage_order_id";

ALTER TABLE "coupon_usage"
DROP CONSTRAINT "fk_coupon_usage_created_by";

ALTER TABLE "coupon_usage"
DROP CONSTRAINT "fk_coupon_usage_updated_by";

ALTER TABLE "coupon" DROP CONSTRAINT "fk_coupons_updated_by";

ALTER TABLE "coupon" DROP CONSTRAINT "fk_coupons_created_by";

ALTER TABLE "ordered_products"
DROP CONSTRAINT "fk_ordered_products_updated_by";

ALTER TABLE "ordered_products"
DROP CONSTRAINT "fk_ordered_products_created_by";

ALTER TABLE "ordered_products"
DROP CONSTRAINT "fk_ordered_products_product_id";

ALTER TABLE "ordered_products"
DROP CONSTRAINT "fk_ordered_products_order_id";

ALTER TABLE "orders" DROP CONSTRAINT "fk_orders_updated_by";

ALTER TABLE "orders" DROP CONSTRAINT "fk_orders_created_by";

ALTER TABLE "orders" DROP CONSTRAINT "fk_orders_user_id";

DROP TABLE "coupon_usage";

DROP TABLE "coupons";

DROP TABLE "ordered_products";

DROP TABLE "orders";