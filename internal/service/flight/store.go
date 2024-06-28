package flight

import (
	"context"
	"database/sql"
	"errors"
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

func (flightStore *FlightStore) CreateFlight(ctx context.Context, createParams types.CreateFlightRequest) (types.Flight, error) {
	// generate uuid
	flightID := uuid.New()
	queryBuilder, err := flightStore.db.Prepare("INSERT INTO flights(id,price,destination,departure,available_seats, wait_seats,flight_date) VALUES($1,$2,$3,$4,$5,$6,$7) RETURNING *;")
	if err != nil {
		return types.Flight{}, fmt.Errorf("prepare statement flights: %w", err)
	}
	defer queryBuilder.Close()
	var result types.Flight

	err = queryBuilder.QueryRowContext(ctx, flightID, createParams.Price, createParams.Destination, createParams.Departure,
		createParams.AvailableSeats, createParams.WaitSeats, time.Unix(createParams.FlightDate, 0).UTC()).Scan(&result.ID, &result.Departure, &result.Destination, &result.Price, &result.FlightDate, &result.AvailableSeats, &result.WaitSeats,
		&result.NextWaitOrder, &result.CreatedAt, &result.UpdatedAt)
	if err != nil {
		return types.Flight{}, fmt.Errorf("could not insert flights: %w", err)
	}
	return result, nil
}

func (flightStore *FlightStore) GetFlightsByCriteria(ctx context.Context,
	queryParams types.QueryFlightRequest,
	pageInfo types.Pagination) (types.FlightsFetchResponse, error) {
	// original sql
	queryBuilder := sq.Select("id", "price", "departure", "destination", "flight_date", "available_seats", "wait_seats", "next_wait_order", "created_at", "updated_at").From("flights").PlaceholderFormat(sq.Dollar)
	// fligt_date >= time.Now()
	whereCondition := []sq.Sqlizer{sq.GtOrEq{"flight_date": time.Now().UTC()},
		sq.Or{sq.NotEq{"available_seats": 0}, sq.NotEq{"wait_seats": 0}}}
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
		return types.FlightsFetchResponse{}, err
	}
	rows, err := flightStore.db.QueryContext(ctx, query, args...)
	if err != nil {
		return types.FlightsFetchResponse{}, err
	}
	result := types.FlightsFetchResponse{}
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
			&flight.UpdatedAt,
		)
		if err != nil {
			return types.FlightsFetchResponse{}, err
		}
		result.Flights = append(result.Flights, types.ConvertFlightToRespone(flight))
	}
	result.Limit = pageInfo.Limit
	result.Offset = pageInfo.Offset
	if len(result.Flights) > 0 {
		result.NextOffset = int64(offset + limit)
	}
	return result, nil
}

func (flightStore *FlightStore) GetFlightById(ctx context.Context, flightID uuid.UUID) (types.FlightResponse, error) {
	queryBuilder := sq.Select("id", "price", "departure", "destination", "flight_date", "available_seats", "wait_seats", "next_wait_order", "created_at", "updated_at").From("flights").PlaceholderFormat(sq.Dollar)
	queryBuilder = queryBuilder.Where(sq.Eq{"id": flightID})
	query, args, err := queryBuilder.ToSql()
	if err != nil {
		return types.FlightResponse{}, fmt.Errorf("failed to use query builder: %w", err)
	}
	rows, err := flightStore.db.QueryContext(ctx, query, args...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return types.FlightResponse{}, nil
		}
		return types.FlightResponse{}, fmt.Errorf("failed to executed %w", err)
	}
	var result types.FlightResponse
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
			&flight.UpdatedAt,
		)
		if err != nil {
			return types.FlightResponse{}, err
		}
		result = types.ConvertFlightToRespone(flight)
	}
	return result, nil
}

func (flightStore *FlightStore) UpdateFlight(tx *sql.Tx, ctx context.Context,
	updateFlightParams types.UpdateFlightEntityParam) (types.Flight, error) {
	updatedAt := time.Now().UTC()
	queryBuilder := sq.Update("flights").Set("available_seats", updateFlightParams.AvailableSeats)
	queryBuilder = queryBuilder.Set("wait_seats", updateFlightParams.WaitSeats)
	queryBuilder = queryBuilder.Set("next_wait_order", updateFlightParams.NextWaitOrder)
	queryBuilder = queryBuilder.Set("updated_at", updatedAt)
	queryBuilder = queryBuilder.Where(sq.Eq{"id": updateFlightParams.ID}).Suffix("RETURNING *;")
	queryBuilder = queryBuilder.PlaceholderFormat(sq.Dollar)
	query, args, err := queryBuilder.ToSql()
	if err != nil {
		return types.Flight{}, fmt.Errorf("failed to use query builder: %w", err)
	}
	rows, err := tx.QueryContext(ctx, query, args...)
	if err != nil {
		return types.Flight{}, fmt.Errorf("failed to executed %w", err)
	}
	var flight types.Flight
	if rows.Next() {
		err = rows.Scan(&flight.ID,
			&flight.Departure,
			&flight.Destination,
			&flight.Price,
			&flight.FlightDate,
			&flight.AvailableSeats,
			&flight.WaitSeats,
			&flight.NextWaitOrder,
			&flight.CreatedAt,
			&flight.UpdatedAt,
		)
		if err != nil {
			return types.Flight{}, err
		}
	}
	return flight, nil
}
