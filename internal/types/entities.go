package types

import (
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
	ID            uuid.UUID    `json:"id" db:"id"`
	FlightID      uuid.UUID    `json:"flight_id" db:"flight_id"`
	CreatedAt     time.Time    `json:"created_at" db:"created_at"`
	CanceledAt    sql.NullTime `json:"canceled_at,omitempty" db:"canceled_at"`
	PaidAt        sql.NullTime `json:"paid_at,omitempty" db:"paid_at"`
	WaitOrder     int32        `json:"wait_order,omitempty" db:"wait_order"`
	TicketNumbers int32        `json:"ticket_numbers" db:"ticket_numbers"`
}
