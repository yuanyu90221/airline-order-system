package flight

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
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
	queryBuilder, err := flightStore.db.Prepare("INSERT INTO flights(id,price,destination,departure,available_seats, wait_seats, flight_date) VALUES($1,$2,$3,$4,$5,$6,$7) RETURNING *;")
	if err != nil {
		return types.Flight{}, fmt.Errorf("prepare statement flights: %w", err)
	}
	defer queryBuilder.Close()
	var result types.Flight

	err = queryBuilder.QueryRowContext(ctx, flightID, createParams.Price, createParams.Destination, createParams.Departure,
		createParams.AvailableSeats, createParams.WaitSeats, time.Unix(createParams.FlightDate, 0)).Scan(&result.ID, &result.Departure, &result.Destination, &result.Price, &result.FlightDate, &result.AvailableSeats, &result.WaitSeats, &result.NextWaitOrder, &result.CreatedAt)
	if err != nil {
		return types.Flight{}, fmt.Errorf("could not insert flights: %w", err)
	}
	return result, nil
}

func (flightSore *FlightStore) GetFlightsByCriteria(ctx context.Context,
	queryParams types.QueryFlightParams,
	pageInfo types.Pagination) (types.FlightFetchResult, error) {
	// original sql
	queryBuilder := sq.Select("id", "price", "departure", "destination", "flight_date", "available_seats", "wait_seats", "next_wait_order", "created_at").From("flights").PlaceholderFormat(sq.Dollar)
	// fligt_date >= time.Now()
	whereCondition := []sq.Sqlizer{sq.GtOrEq{"flight_date": time.Now().UTC()},
		sq.NotEq{"available_seats": 0}, sq.NotEq{"wait_seats": 0}}
	// Parse parameters
	if queryParams.FlightDate > 0 {
		whereCondition = append(whereCondition, sq.GtOrEq{"flight_date": time.Unix(queryParams.FlightDate, 0)})
	}
	if queryParams.Departure != "" {
		whereCondition = append(whereCondition, sq.Eq{"departure": queryParams.Departure})
	}
	if queryParams.Destination != "" {
		whereCondition = append(whereCondition, sq.Eq{"destination": queryParams.Destination})
	}
	queryBuilder = queryBuilder.Where(sq.And(whereCondition))
	// get pagenation info
	offset := uint64(pageInfo.Offset)
	limit := uint64(pageInfo.Limit)
	if pageInfo.Offset > 0 {
		queryBuilder = queryBuilder.Offset(offset)
	}
	if limit > 0 {
		queryBuilder = queryBuilder.Limit(limit)
	}
	queryBuilder = queryBuilder.OrderBy("flight_date ASC")

	query, args, err := queryBuilder.ToSql()
	if err != nil {
		return types.FlightFetchResult{}, err
	}
	rows, err := flightSore.db.QueryContext(ctx, query, args...)
	if err != nil {
		return types.FlightFetchResult{}, err
	}
	result := types.FlightFetchResult{}
	for rows.Next() {
		var flight types.Flight
		err = rows.Scan(&flight.ID,
			&flight.Price,
			&flight.Departure,
			&flight.Destination,
			&flight.FlightDate,
			&flight.AvailableSeats,
			&flight.WaitSeats,
			&flight.NextWaitOrder,
			&flight.CreatedAt,
		)
		if err != nil {
			return types.FlightFetchResult{}, err
		}

		result.Flights = append(result.Flights, types.FlightWithRemain{
			Flight: flight,
			Remain: int(flight.AvailableSeats) + int(flight.WaitSeats),
		})
	}
	result.Limit = pageInfo.Limit
	result.Offset = pageInfo.Offset
	if len(result.Flights) > 0 {
		result.NextOffset = int64(offset + limit)
	}
	return result, nil
}
