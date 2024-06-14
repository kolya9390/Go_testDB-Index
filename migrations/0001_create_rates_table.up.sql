-- +migrate Up
CREATE TABLE IF NOT EXISTS rates (
    id SERIAL PRIMARY KEY,
    timestamp INT,
    ask_price VARCHAR NOT NULL,
    bid_price VARCHAR NOT NULL
);
-- +migrate Down
DROP TABLE rates;