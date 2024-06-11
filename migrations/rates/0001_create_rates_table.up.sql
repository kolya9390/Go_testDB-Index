-- +migrate Up
CREATE TABLE IF NOT EXISTS rates (
    id SERIAL PRIMARY KEY,
    timestamp BIGINT NOT NULL,
    asks_price VARCHAR NOT NULL,
    bids_price VARCHAR NOT NULL,
);
-- +migrate Down
DROP TABLE rates;
