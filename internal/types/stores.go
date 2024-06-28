package types

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
)

type OrderStore interface {
	CreateOrder(tx *sql.Tx, ctx context.Context, createOrderInfo CreateOrderEntityParam) (Order, error)
	GetOrderById(ctx context.Context, orderID uuid.UUID) (Order, error)
}

type OrderCacheStore interface {
	CreateOrder(ctx context.Context, createOrderParam OrderCacheCreateParam) (OrderCacheResult, error)
	GetCurrentRemain(ctx context.Context, getOrderRemain OrderCacheParam) (OrderCacheRemain, error)
}

type FlightCacheStore interface {
	UpdateFlight(ctx context.Context, flightInfo Flight) (Flight, error)
	GetFlightCacheInfo(ctx context.Context, fligtID string) (Flight, error)
}

type FlightStore interface {
	GetFlightsByCriteria(ctx context.Context, queryParams QueryFlightRequest, pagination Pagination) (FlightsFetchResponse, error)
	CreateFlight(ctx context.Context, createParams CreateFlightRequest) (Flight, error)
	GetFlightById(ctx context.Context, flightID uuid.UUID) (FlightResponse, error)
	UpdateFlight(tx *sql.Tx, ctx context.Context, updateFlightParams UpdateFlightEntityParam) (Flight, error)
}
