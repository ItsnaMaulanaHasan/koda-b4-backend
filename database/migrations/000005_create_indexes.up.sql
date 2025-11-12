CREATE INDEX idx_products_name ON products (name);

CREATE INDEX idx_products_is_flash_sale ON products (is_flash_sale)
WHERE
    is_flash_sale = true;

CREATE INDEX idx_transactions_user_date ON transactions (user_id, date_transaction);

CREATE INDEX idx_carts_user ON carts (user_id);

CREATE INDEX idx_transactions_status ON transactions (status);

CREATE INDEX idx_products_active ON products (is_active)
WHERE
    is_active = true;

CREATE INDEX idx_product_images_product ON product_images (product_id);

CREATE INDEX idx_transaction_items_transactions ON transaction_items (transaction_id);

CREATE INDEX idx_sessions_token ON sessions (session_token);

CREATE INDEX idx_sessions_user_expired ON sessions (user_id, expired_at);

CREATE INDEX idx_sessions_active ON sessions (user_id, is_active)
WHERE
    is_active = true;

CREATE INDEX idx_password_resets_expired ON password_resets (expired_at);