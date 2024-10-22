CREATE TYPE USER_ROLE AS ENUM ('user','support','admin');

CREATE TABLE users
(
    id         UUID PRIMARY KEY         DEFAULT gen_random_uuid(),
    email      VARCHAR(255) UNIQUE NOT NULL,
    fname      VARCHAR(100)        NOT NULL,
    lname      VARCHAR(100)        NOT NULL,
    password   VARCHAR(255)        NOT NULL,
    role       USER_ROLE           NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TYPE CHAT_STATUS AS ENUM ('open','closed','pending');

CREATE TABLE chat
(
    id          UUID PRIMARY KEY         DEFAULT gen_random_uuid(),
    status      CHAT_STATUS NOT NULL,
    created_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    is_resolved BOOLEAN                  DEFAULT FALSE
);

CREATE TABLE messages
(
    id         UUID PRIMARY KEY         DEFAULT gen_random_uuid(),
    chat_id    UUID NOT NULL REFERENCES chat (id) ON DELETE CASCADE,
    user_id    UUID NOT NULL REFERENCES users (id) ON DELETE SET NULL,
    content    TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TYPE ORDER_TYPE AS ENUM ('pending','completed','returned');

CREATE TABLE orders
(
    id           UUID PRIMARY KEY         DEFAULT gen_random_uuid(),
    user_id      UUID           NOT NULL REFERENCES users (id) ON DELETE SET NULL,
    total_amount DECIMAL(10, 2) NOT NULL,
    status       ORDER_TYPE     NOT NULL,
    address      VARCHAR(255)   NOT NULL,
    created_at   TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at   TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE order_items
(
    id                UUID PRIMARY KEY         DEFAULT gen_random_uuid(),
    order_id          UUID           NOT NULL REFERENCES orders (id) ON DELETE CASCADE,
    product_id        UUID           NOT NULL REFERENCES products (id) ON DELETE RESTRICT,
    quantity          INT            NOT NULL CHECK (quantity > 0),
    price_at_purchase DECIMAL(10, 2) NOT NULL,
    created_at        TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE products
(
    id          UUID PRIMARY KEY         DEFAULT gen_random_uuid(),
    name        VARCHAR(255)   NOT NULL,
    price       DECIMAL(10, 2) NOT NULL,
    discount    DECIMAL(5, 2) CHECK (discount >= 0 AND discount <= 100),
    description TEXT,
    created_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TYPE CATEGORY_TYPE AS ENUM ('plant', 'tool', 'seed','soil');

CREATE TABLE tags
(
    id         UUID PRIMARY KEY         DEFAULT gen_random_uuid(),
    name       VARCHAR(50) UNIQUE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE product_tags
(
    product_id UUID NOT NULL REFERENCES products (id) ON DELETE CASCADE,
    tag_id     UUID NOT NULL REFERENCES tags (id) ON DELETE CASCADE,
    PRIMARY KEY (product_id, tag_id)
);

CREATE TYPE PROD_INTERACTION_TYPE AS ENUM ('review','question');

CREATE TABLE product_interactions
(
    id             UUID PRIMARY KEY         DEFAULT gen_random_uuid(),
    product_id     UUID                  NOT NULL REFERENCES products (id) ON DELETE CASCADE,
    user_id        UUID                  REFERENCES users (id) ON DELETE SET NULL,
    type           PROD_INTERACTION_TYPE NOT NULL,
    content        TEXT                  NOT NULL,
    is_answered    BOOLEAN                  DEFAULT FALSE,
    admin_response TEXT,
    created_at     TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at     TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE favorites
(
    user_id    UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES products (id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    PRIMARY KEY (user_id, product_id)
);

CREATE TYPE DELIVERY_STATUS AS ENUM ('shipped','in transit','delivered','returned');

CREATE TABLE deliveries
(
    id                 UUID PRIMARY KEY         DEFAULT gen_random_uuid(),
    order_id           UUID        NOT NULL REFERENCES orders (id) ON DELETE CASCADE,
    status             DELIVERY_STATUS NOT NULL,
    tracking_number    VARCHAR(100),
    estimated_delivery TIMESTAMP WITH TIME ZONE,
    delivered_at       TIMESTAMP WITH TIME ZONE,
    created_at         TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at         TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);


CREATE INDEX idx_products_description_en ON products USING gin (to_tsvector('english', description));
CREATE INDEX idx_products_description_ru ON products USING gin (to_tsvector('russian', description));
CREATE INDEX idx_reviews_content_en ON product_interactions USING gin (to_tsvector('english', content));
CREATE INDEX idx_reviews_content_ru ON product_interactions USING gin (to_tsvector('russian', content));
CREATE INDEX idx_product_interactions_compound ON product_interactions (product_id, user_id);
CREATE INDEX idx_chat_status ON chat (status);
CREATE INDEX idx_messages_chat_id ON messages (chat_id);
CREATE INDEX idx_orders_user_id ON orders (user_id);
CREATE INDEX idx_product_interactions_product_id ON product_interactions (product_id);
