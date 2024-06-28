package types

import (
	"time"

	"github.com/google/uuid"
)

type FlightResponse struct {
	ID             uuid.UUID `json:"id"`
	Departure      string    `json:"departure"`
	Destination    string    `json:"destination"`
	FlightDate     time.Time `json:"flight_date"`
	Price          float64   `json:"price"`
	AvailableSeats int32     `json:"available_seats"`
	WaitSeats      int32     `json:"wait_seats"`
	NextWaitOrder  int32     `json:"next_wait_order"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	Remain         int       `json:"remain"`
}
type FlightsFetchResponse struct {
	Flights []FlightResponse `json:"flights"`
	Pagination
}

func ConvertFlightToRespone(flight Flight) FlightResponse {
	return FlightResponse{
		ID:             flight.ID,
		Departure:      flight.Departure,
		Destination:    flight.Destination,
		FlightDate:     flight.FlightDate,
		Price:          flight.Price,
		AvailableSeats: flight.AvailableSeats,
		WaitSeats:      flight.WaitSeats,
		NextWaitOrder:  flight.NextWaitOrder,
		CreatedAt:      flight.CreatedAt,
		UpdatedAt:      flight.UpdatedAt,
		Remain:         int(flight.AvailableSeats) + int(flight.WaitSeats),
	}
}

type CreateOrderResponse struct {
	ID            string `json:"id"`
	FlightID      string `json:"flight_id"`
	WaitOrder     int64  `json:"wait_order"`
	TicketNumbers int64  `json:"ticket_numbers"`
	IsWait        bool   `json:"is_wait"`
}

func ConvertCreateOrderEventToResponse(event CreateOrderEvent) CreateOrderResponse {
	var response CreateOrderResponse
	response.ID = event.ID
	response.WaitOrder = event.WaitOrder
	if !event.IsWait {
		response.WaitOrder = -1
	}
	response.FlightID = event.ID
	response.FlightID = event.FlightID
	response.TicketNumbers = event.TicketNumbers
	response.IsWait = event.IsWait
	return response
}

type QueryOrderResponse struct {
	ID            string    `json:"id"`
	FlightID      string    `json:"flight_id"`
	CreatedAt     time.Time `json:"created_at"`
	CanceledAt    string    `json:"canceled_at,omitempty"`
	PaidAt        string    `json:"paid_at,omitempty"`
	WaitOrder     int32     `json:"wait_order"`
	TicketNumbers int32     `json:"ticket_numbers"`
}

func ConvertOrderEntityToResponse(order Order) QueryOrderResponse {
	var response QueryOrderResponse
	response.ID = order.ID.String()
	response.FlightID = order.FlightID.String()
	response.CreatedAt = order.CreatedAt
	if order.CanceledAt.Valid {
		response.CanceledAt = order.CanceledAt.Time.UTC().String()
	}
	if order.PaidAt.Valid {
		response.PaidAt = order.PaidAt.Time.UTC().String()
	}
	response.TicketNumbers = order.TicketNumbers
	response.WaitOrder = order.WaitOrder
	return response
}
