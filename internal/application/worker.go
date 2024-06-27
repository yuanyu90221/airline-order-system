package application

import (
	"github.com/yuanyu90221/airline-order-system/internal/service/flight"
	"github.com/yuanyu90221/airline-order-system/internal/service/order"
)

func (app *App) setupOrderWorker() {
	flightCacheStore := flight.NewCacheStore(app.rdb)
	flightStore := flight.NewFlightStore(app.db)
	orderStore := order.NewOrderStore(app.db)
	orderService := order.NewOrderService(app.db, orderStore, flightStore)
	orderWorker := order.NewOrderWorker(orderService, flightCacheStore, app.broker)
	app.orderWorker = orderWorker
}
