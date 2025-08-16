
CREATE TABLE orders (
    order_uid VARCHAR(32) PRIMARY KEY,
    track_number VARCHAR(32),
    entry VARCHAR(16),
    locale VARCHAR(8),
    internal_signature VARCHAR(128),
    customer_id VARCHAR(32),
    delivery_service VARCHAR(32),
    shardkey VARCHAR(8),
    sm_id INTEGER,
    date_created TIMESTAMP,
    oof_shard VARCHAR(8)
);


CREATE TABLE delivery (
    order_uid VARCHAR(32) REFERENCES orders(order_uid) ON DELETE CASCADE,
    name VARCHAR(128),
    phone VARCHAR(32),
    zip VARCHAR(16),
    city VARCHAR(64),
    address VARCHAR(128),
    region VARCHAR(64),
    email VARCHAR(64),
    PRIMARY KEY (order_uid)
);


CREATE TABLE payment (
    order_uid VARCHAR(32) REFERENCES orders(order_uid) ON DELETE CASCADE,
    transaction VARCHAR(32),
    request_id VARCHAR(32),
    currency VARCHAR(8),
    provider VARCHAR(32),
    amount INTEGER,
    payment_dt BIGINT,
    bank VARCHAR(32),
    delivery_cost INTEGER,
    goods_total INTEGER,
    custom_fee INTEGER,
    PRIMARY KEY (order_uid)
);

CREATE TABLE items (
    id SERIAL PRIMARY KEY,
    order_uid VARCHAR(32) REFERENCES orders(order_uid) ON DELETE CASCADE,
    chrt_id BIGINT,
    track_number VARCHAR(32),
    price INTEGER,
    rid VARCHAR(32),
    name VARCHAR(128),
    sale INTEGER,
    size VARCHAR(16),
    total_price INTEGER,
    nm_id BIGINT,
    brand VARCHAR(64),
    status INTEGER
);
