package application

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yuanyu90221/airline-order-system/internal/service/flight"
	"github.com/yuanyu90221/airline-order-system/internal/service/order"
)

// define route
func (app *App) loadRoutes() {
	gin.SetMode(app.config.GinMode)
	router := gin.New()
	// recovery middleware
	router.Use(gin.Recovery())

	// default health
	router.GET("/", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, map[string]string{"message": "status ok"})
	})
	app.router = router
}

// setup order route
func (app *App) loadOrderRoutes() {
	orderGroup := app.router.Group("/orders")
	orderCacheStore := order.NewCacheStore(app.rdb)
	flightCacheStore := flight.NewCacheStore(app.rdb)
	orderStore := order.NewOrderStore(app.db)
	orderHandler := order.NewHandler(orderCacheStore, flightCacheStore, app.bFilter, app.broker, orderStore)
	orderHandler.RegisterRoute(orderGroup)
}

// setup flight route
func (app *App) loadFlightRoutes() {
	flightGroup := app.router.Group("/flights")
	orderCacheStore := order.NewCacheStore(app.rdb)
	flightCacheStore := flight.NewCacheStore(app.rdb)
	flightStore := flight.NewFlightStore(app.db)
	flightHandler := flight.NewHandler(orderCacheStore, flightCacheStore, flightStore, app.bFilter)
	flightHandler.RegisterRoute(flightGroup)
}
