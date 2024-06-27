package types

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type Flight struct {
	ID             uuid.UUID `json:"id" db:"id"`
	Departure      string    `json:"departure" db:"departure"`
	Destination    string    `json:"destination" db:"destination"`
	FlightDate     time.Time `json:"flight_date" db:"flight_date"`
	Price          float64   `json:"price" db:"price"`
	AvailableSeats int32     `json:"available_seats" db:"available_seats"`
	WaitSeats      int32     `json:"wait_seats" db:"wait_seats"`
	NextWaitOrder  int32     `json:"next_wait_order" db:"next_wait_order"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

type Order struct {
	ID            uuid.UUID `json:"id" db:"id"`
	FlightID      uuid.UUID `json:"flight_id" db:"flight_id"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	CanceledAt    time.Time `json:"canceled_at,omitempty" db:"canceled_at"`
	PaidAt        time.Time `json:"paid_at,omitempty" db:"paid_at"`
	WaitOrder     int32     `json:"wait_order,omitempty" db:"wait_order"`
	TicketNumbers int32     `json:"ticket_numbers" db:"ticket_numbers"`
}

type QueryFlightParams struct {
	FlightDate  int64  `json:"flight_date"`
	Destination string `json:"destination"`
	Departure   string `json:"depature"`
}
type CreateFlightParams struct {
	Price          float64 `json:"price" validate:"required"`
	FlightDate     int64   `json:"flight_date" validate:"required"`
	Destination    string  `json:"destination" validate:"required"`
	Departure      string  `json:"departure" validate:"required"`
	AvailableSeats int64   `json:"available_seats" validate:"required"`
	WaitSeats      int64   `json:"wait_seats" validate:"required"`
}
type Pagination struct {
	NextOffset int64 `json:"next_offset"`
	Offset     int64 `json:"offset"`
	Limit      int64 `json:"limit"`
}
type FlightResponse struct {
	Flight
	Remain int `json:"remain"`
}
type FlightFetchResult struct {
	Flights []FlightResponse `json:"flights"`
	Pagination
}
type FlightStore interface {
	GetFlightsByCriteria(ctx context.Context, queryParams QueryFlightParams, pagination Pagination) (FlightFetchResult, error)
	CreateFlight(ctx context.Context, createParams CreateFlightParams) (Flight, error)
	GetFlightById(ctx context.Context, flightID uuid.UUID) (FlightResponse, error)
}

type OrderStore interface {
	CreateOrder(tx *sql.Tx, ctx context.Context) (Order, error)
	GetOrderById(ctx context.Context) (Order, error)
}
type OrderCacheRequest struct {
	FlightID         string `json:"flight_id" validate:"required"`
	CurrentTotal     int64  `json:"current_total" validate:"required"`
	CurrentWait      int64  `json:"current_wait" validate:"required"`
	CurrentWaitOrder int64  `json:"current_wait_order" validate:"required"`
}
type CreateOrderRequest struct {
	FlightID      string `json:"flight_id" validate:"required"`
	TicketNumbers int64  `json:"ticket_numbers" validate:"required"`
}
type OrderCacheCreateRequest struct {
	OrderCacheRequest
	TicketNumbers int64 `json:"ticket_numbers" validate:"required"`
}
type OrderCacheResult struct {
	CurrentTotal     int64 `json:"current_total" validate:"required"`
	CurrentWait      int64 `json:"current_wait" validate:"required"`
	CurrentWaitOrder int64 `json:"current_wait_order" validate:"required"`
	IsValid          bool  `json:"is_valid"`
}
type OrderCacheRemain struct {
	CurrentRemain int64 `json:"current_remain" validate:"required"`
}
type OrderCacheStore interface {
	CreateOrder(ctx context.Context, createOrderParam OrderCacheCreateRequest) (OrderCacheResult, error)
	GetCurrentRemain(ctx context.Context, getOrderRemain OrderCacheRequest) (OrderCacheRemain, error)
}

type FlightCacheStore interface {
	UpdateFlight(ctx context.Context, flightInfo Flight) (Flight, error)
	GetFlightCacheInfo(ctx context.Context, fligtID string) (Flight, error)
}
