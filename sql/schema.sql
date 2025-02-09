CREATE TYPE USER_ROLE AS ENUM ('user','support','admin');

CREATE OR REPLACE FUNCTION update_updated_at_column()
    RETURNS TRIGGER
    LANGUAGE plpgsql
    set search_path = ''
AS
$$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$;

CREATE TABLE users
(
    id         UUID PRIMARY KEY         DEFAULT gen_random_uuid(),
    email      VARCHAR(120) UNIQUE NOT NULL,
    fname      VARCHAR(100)        NOT NULL,
    lname      VARCHAR(100)        NOT NULL,
    password   VARCHAR(144)        NOT NULL,
    role       USER_ROLE           NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TRIGGER update_users_updated_at
    BEFORE UPDATE
    ON users
    FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

CREATE TYPE CHAT_STATUS AS ENUM ('open','closed');
CREATE TABLE chats
(
    id         UUID PRIMARY KEY         DEFAULT gen_random_uuid(),
    status     CHAT_STATUS NOT NULL,
    created_by UUID        NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TRIGGER update_chats_updated_at
    BEFORE UPDATE
    ON chats
    FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

CREATE TABLE messages
(
    id         UUID PRIMARY KEY         DEFAULT gen_random_uuid(),
    chat_id    UUID NOT NULL REFERENCES chats (id) ON DELETE CASCADE,
    user_id    UUID NOT NULL REFERENCES users (id) ON DELETE SET NULL,
    content    TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TYPE ORDER_TYPE AS ENUM ('pending','completed','returned');

CREATE TABLE orders
(
    id         UUID PRIMARY KEY         DEFAULT gen_random_uuid(),
    user_id    UUID       REFERENCES users (id) ON DELETE SET NULL,
    status     ORDER_TYPE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TRIGGER update_orders_updated_at
    BEFORE UPDATE
    ON orders
    FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

CREATE TABLE order_details
(
    id               UUID PRIMARY KEY      DEFAULT gen_random_uuid(),
    order_id         UUID         NOT NULL REFERENCES orders (id) ON DELETE CASCADE,
    address          VARCHAR(255) NOT NULL,
    phone_number     VARCHAR(24),
    return_statement TEXT,
    created_at       TIMESTAMP    NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMP    NOT NULL DEFAULT NOW()
);

CREATE TRIGGER update_orders_details_updated_at
    BEFORE UPDATE
    ON order_details
    FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

CREATE TABLE order_items
(
    id                UUID PRIMARY KEY         DEFAULT gen_random_uuid(),
    order_id          UUID           NOT NULL REFERENCES orders (id) ON DELETE CASCADE,
    product_id        UUID           NOT NULL REFERENCES products (id) ON DELETE RESTRICT,
    quantity          INT            NOT NULL CHECK (quantity > 0),
    price_at_purchase DECIMAL(10, 2) NOT NULL,
    created_at        TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TYPE DELIVERY_STATUS AS ENUM ('shipped','in transit','delivered','returned');

-- CREATE TABLE deliveries
-- (
--     id                 UUID PRIMARY KEY         DEFAULT gen_random_uuid(),
--     order_id           UUID            NOT NULL REFERENCES orders (id) ON DELETE CASCADE,
--     status             DELIVERY_STATUS NOT NULL,
--     tracking_number    VARCHAR(100),
--     estimated_delivery TIMESTAMP WITH TIME ZONE,
--     delivered_at       TIMESTAMP WITH TIME ZONE,
--     created_at         TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
--     updated_at         TIMESTAMP WITH TIME ZONE DEFAULT NOW()
-- );
--
-- CREATE TRIGGER update_deliveries_updated_at
--     BEFORE UPDATE
--     ON deliveries
--     FOR EACH ROW
-- EXECUTE FUNCTION update_updated_at_column();

CREATE TABLE products
(
    id          UUID PRIMARY KEY         DEFAULT gen_random_uuid(),
    img         VARCHAR(255)   NOT NULL,
    name        VARCHAR(255)   NOT NULL,
    price       DECIMAL(10, 2) NOT NULL,
    discount    DECIMAL(5, 2) CHECK (discount >= 0 AND discount <= 100),
    description TEXT,
    type        UUID           NOT NULL REFERENCES tags (id),
    category    UUID           NOT NULL REFERENCES tags (id),
    created_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TRIGGER update_products_updated_at
    BEFORE UPDATE
    ON products
    FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

CREATE TYPE CATEGORY_TYPE AS ENUM ('plant', 'tool', 'seed','soil');

CREATE TABLE tags
(
    id         UUID PRIMARY KEY         DEFAULT gen_random_uuid(),
    name       VARCHAR(50) UNIQUE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TRIGGER update_tags_updated_at
    BEFORE UPDATE
    ON tags
    FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

CREATE TYPE PROD_INTERACTION_TYPE AS ENUM ('review','question');

-- CREATE TABLE product_interactions
-- (
--     id          UUID PRIMARY KEY         DEFAULT gen_random_uuid(),
--     product_id  UUID                  NOT NULL REFERENCES products (id) ON DELETE CASCADE,
--     user_id     UUID                  REFERENCES users (id) ON DELETE SET NULL,
--     type        PROD_INTERACTION_TYPE NOT NULL,
--     content     TEXT                  NOT NULL,
--     is_answered BOOLEAN                  DEFAULT FALSE,
--     response    TEXT,
--     created_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
--     updated_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW()
-- );

-- CREATE TRIGGER update_prod_interactions_updated_at
--     BEFORE UPDATE
--     ON product_interactions
--     FOR EACH ROW
-- EXECUTE FUNCTION update_updated_at_column();

-- CREATE TABLE favorites
-- (
--     user_id    UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
--     product_id UUID NOT NULL REFERENCES products (id) ON DELETE CASCADE,
--     created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
--     PRIMARY KEY (user_id, product_id)
-- );


CREATE INDEX idx_products_description_en ON products USING gin (to_tsvector('english', description));
CREATE INDEX idx_products_description_ru ON products USING gin (to_tsvector('russian', description));
-- CREATE INDEX idx_reviews_content_en ON product_interactions USING gin (to_tsvector('english', content));
-- CREATE INDEX idx_reviews_content_ru ON product_interactions USING gin (to_tsvector('russian', content));
-- CREATE INDEX idx_product_interactions_compound ON product_interactions (product_id, user_id);
CREATE INDEX idx_chat_status ON chats (status);
CREATE INDEX idx_messages_chat_id ON messages (chat_id);
CREATE INDEX idx_orders_user_id ON orders (user_id);
-- CREATE INDEX idx_product_interactions_product_id ON product_interactions (product_id);