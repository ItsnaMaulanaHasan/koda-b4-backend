ALTER TABLE "coupon_usage" DROP CONSTRAINT "fk_coupon_usage_user_id";

ALTER TABLE "coupon_usage"
DROP CONSTRAINT "fk_coupon_usage_coupon_id";

ALTER TABLE "coupon_usage"
DROP CONSTRAINT "fk_coupon_usage_transaction_id";

ALTER TABLE "coupon_usage"
DROP CONSTRAINT "fk_coupon_usage_created_by";

ALTER TABLE "coupon_usage"
DROP CONSTRAINT "fk_coupon_usage_updated_by";

ALTER TABLE "coupons" DROP CONSTRAINT "fk_coupons_updated_by";

ALTER TABLE "coupons" DROP CONSTRAINT "fk_coupons_created_by";

ALTER TABLE "transaction_items"
DROP CONSTRAINT "fk_transaction_items_updated_by";

ALTER TABLE "transaction_items"
DROP CONSTRAINT "fk_transaction_items_created_by";

ALTER TABLE "transaction_items"
DROP CONSTRAINT "fk_transaction_items_product_id";

ALTER TABLE "transaction_items"
DROP CONSTRAINT "fk_transaction_items_transaction_id";

ALTER TABLE "transactions"
DROP CONSTRAINT "fk_transactions_updated_by";

ALTER TABLE "transactions"
DROP CONSTRAINT "fk_transactions_created_by";

ALTER TABLE "transactions"
DROP CONSTRAINT "fk_transactions_status_id";

ALTER TABLE "transactions"
DROP CONSTRAINT "fk_transactions_order_method_id";

ALTER TABLE "transactions"
DROP CONSTRAINT "fk_transactions_payment_method_id";

ALTER TABLE "transactions" DROP CONSTRAINT "fk_transactions_user_id";

DROP TABLE "coupon_usage";

DROP TABLE "coupons";

DROP TABLE "transaction_items";

DROP TABLE "transactions";