-- +goose Up
CREATE TABLE IF NOT EXISTS orders (
  id uuid PRIMARY KEY NOT NULL,
  flight_id UUID NOT NULL REFERENCES flights(id) ON DELETE CASCADE,
  paid_at TIMESTAMP DEFAULT NULL,
  canceled_at TIMESTAMP DEFAULT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT now(),
  wait_order INTEGER DEFAULT NULL,
  ticket_numbers INTEGER NOT NULL
);
-- +goose Down

DROP TABLE IF EXISTS orders;