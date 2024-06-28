package types

type CreateOrderEvent struct {
	ID             string `json:"id"`
	FlightID       string `json:"flight_id"`
	WaitOrder      int64  `json:"wait_order"`
	WaitSeats      int64  `json:"wait_seats"`
	AvailableSeats int64  `json:"available_seats"`
	TicketNumbers  int64  `json:"ticket_numbers"`
	IsWait         bool   `json:"is_wait"`
}
