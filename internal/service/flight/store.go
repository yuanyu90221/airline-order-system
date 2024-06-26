package flight

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/yuanyu90221/airline-order-system/internal/types"
)

type FlightStore struct {
	db *sql.DB
}

func NewFlightStore(db *sql.DB) *FlightStore {
	return &FlightStore{
		db: db,
	}
}

func (flightStore *FlightStore) CreateFlight(ctx context.Context, createParams types.CreateFlightParams) (types.Flight, error) {
	// generate uuid
	flightID := uuid.New()
	stmt, err := flightStore.db.Prepare("INSERT INTO flights(id,price,destination,departure,available_seats, wait_seats, flight_date) VALUES($1,$2,$3,$4,$5,$6,$7) RETURNING *;")
	if err != nil {
		return types.Flight{}, fmt.Errorf("prepare statement flights: %w", err)
	}
	defer stmt.Close()
	var result types.Flight

	err = stmt.QueryRowContext(ctx, flightID, createParams.Price, createParams.Destination, createParams.Departure,
		createParams.AvailableSeats, createParams.WaitSeats, time.Unix(createParams.FlightDate, 0)).Scan(&result.ID, &result.Departure, &result.Destination, &result.Price, &result.FlightDate, &result.AvailableSeats, &result.WaitSeats, &result.NextWaitOrder, &result.CreatedAt)
	if err != nil {
		return types.Flight{}, fmt.Errorf("could not insert flights: %w", err)
	}
	return result, nil
}

func (flightSore *FlightStore) GetFlightsByCriteria(ctx context.Context,
	queryParams types.QueryFlightParams,
	pageInfo types.Pagination) (types.FlightFetchResult, error) {
	// TODO: Parse parameters
	// TODO: get pagenation info
	return types.FlightFetchResult{}, nil
}
