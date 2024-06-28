package types

import "context"

type OrderServcie interface {
	CreateOrderHandler(ctx context.Context,
		createOrderParam CreateOrderEntityParam,
		updateFlightParam UpdateFlightEntityParam,
	) (Flight, Order, error)
}
