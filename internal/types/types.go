package types

import (
	"github.com/google/uuid"
)

type Pagination struct {
	NextOffset int64 `json:"next_offset"`
	Offset     int64 `json:"offset"`
	Limit      int64 `json:"limit"`
}

type UpdateFlightEntityParam struct {
	ID             uuid.UUID `json:"id" db:"id"`
	AvailableSeats int32     `json:"available_seats" db:"available_seats"`
	WaitSeats      int32     `json:"wait_seats" db:"wait_seats"`
	NextWaitOrder  int32     `json:"next_wait_order" db:"next_wait_order"`
}

type CreateOrderEntityParam struct {
	ID            uuid.UUID `json:"id" db:"id"`
	FlightID      uuid.UUID `json:"flight_id" db:"flight_id"`
	WaitOrder     int32     `json:"wait_order,omitempty" db:"wait_order"`
	TicketNumbers int32     `json:"ticket_numbers" db:"ticket_numbers"`
}
type OrderCacheParam struct {
	FlightID         string `json:"flight_id" validate:"required"`
	CurrentTotal     int64  `json:"current_total" validate:"required"`
	CurrentWait      int64  `json:"current_wait" validate:"required"`
	CurrentWaitOrder int64  `json:"current_wait_order" validate:"required"`
}

type OrderCacheCreateParam struct {
	OrderCacheParam
	TicketNumbers int64 `json:"ticket_numbers" validate:"required"`
}
type OrderCacheResult struct {
	CurrentTotal     int64 `json:"current_total" validate:"required"`
	CurrentWait      int64 `json:"current_wait" validate:"required"`
	CurrentWaitOrder int64 `json:"current_wait_order" validate:"required"`
	IsValid          bool  `json:"is_valid"`
	IsWait           bool  `json:"is_wait"`
}
type OrderCacheRemain struct {
	CurrentRemain int64 `json:"current_remain" validate:"required"`
}
