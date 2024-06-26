-- +goose Up
CREATE TABLE IF NOT EXISTS flights (
  id UUID PRIMARY KEY,
  departure VARCHAR(100) NOT NULL,
  destination VARCHAR(100) NOT NULL,
  price DECIMAL(10,2) NOT NULL,
  flight_date TIMESTAMP NOT NULL,
  available_seats INTEGER NOT NUlL,
  wait_seats INTEGER NOT NULL,
  next_wait_order INTEGER NOT NULL DEFAULT -1,
  created_at TIMESTAMP NOT NULL DEFAULT now(),
  updated_at TIMESTAMP NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS flight_date ON flights (flight_date);
CREATE INDEX IF NOT EXISTS departure ON flights (departure);
CREATE INDEX IF NOT EXISTS destination ON flights (destination);
CREATE INDEX IF NOT EXISTS search_criteria ON flights (departure, destination, flight_date);

-- +goose Down
DROP INDEX IF EXISTS search_criteria CASCADE;
DROP INDEX IF EXISTS destination CASCADE;
DROP INDEX IF EXISTS departure CASCADE;
DROP INDEX IF EXISTS flight_date CASCADE;
DROP TABLE IF EXISTS flights;