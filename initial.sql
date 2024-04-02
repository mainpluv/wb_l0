CREATE TABLE IF NOT EXISTS "delivery" (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255),
    phone VARCHAR(20),
    zip VARCHAR(20),
    city VARCHAR(255),
    address TEXT,
    region VARCHAR(255),
    email VARCHAR(255)
);

CREATE TABLE IF NOT EXISTS "payment" (
    id SERIAL PRIMARY KEY,
    transaction UUID DEFAULT gen_random_uuid(),
    request_id VARCHAR(255),
    currency VARCHAR(10),
    provider VARCHAR(50),
    amount INT,
    payment_dt TIMESTAMP,
    bank VARCHAR(50),
    delivery_cost INT,
    goods_total INT,
    custom_fee INT
);

CREATE TABLE IF NOT EXISTS "items" (
    id SERIAL PRIMARY KEY,
    chrt_id INT,
    track_number VARCHAR(255),
    price INT,
    rid VARCHAR(255),
    name VARCHAR(255),
    sale INT,
    size VARCHAR(50),
    total_price INT,
    nm_id INT,
    brand VARCHAR(255),
    status INT
);
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS "orders" (
    order_uuid UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    track_number VARCHAR(255),
    entry VARCHAR(50),
    delivery_id INT,
    payment_id INT,
    locale VARCHAR(10),
    internal_signature VARCHAR(255),
    customer_id VARCHAR(255),
    delivery_service VARCHAR(50),
    shard_key VARCHAR(10),
    sm_id INT,
    date_created TIMESTAMP,
    oof_shard VARCHAR(10),
    FOREIGN KEY (delivery_id) REFERENCES delivery (id),
    FOREIGN KEY (payment_id) REFERENCES payment (id)
);

CREATE TABLE IF NOT EXISTS "orders_items" (
    "order_uuid" UUID PRIMARY KEY REFERENCES "orders" ("order_uuid"),
    "item_id" SERIAL REFERENCES "items" ("id")
)
