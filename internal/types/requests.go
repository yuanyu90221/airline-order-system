package types

type QueryFlightRequest struct {
	FlightDate  int64  `json:"flight_date" validate:"required"`
	Destination string `json:"destination" validate:"required"`
	Departure   string `json:"depature" validate:"required"`
}
type CreateFlightRequest struct {
	Price          float64 `json:"price" validate:"required"`
	FlightDate     int64   `json:"flight_date" validate:"required"`
	Destination    string  `json:"destination" validate:"required"`
	Departure      string  `json:"departure" validate:"required"`
	AvailableSeats int64   `json:"available_seats" validate:"required"`
	WaitSeats      int64   `json:"wait_seats" validate:"required"`
}

type CreateOrderRequest struct {
	FlightID      string `json:"flight_id" validate:"required"`
	TicketNumbers int64  `json:"ticket_numbers" validate:"required"`
}
