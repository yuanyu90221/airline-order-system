package types

import "context"

type OrderServcie interface {
	CreateOrderHandler(ctx context.Context,
		createOrderParams CreateOrderEntityRequest,
		updateFlightParams UpdateFlightEntityRequest,
	) (Flight, Order, error)
}
