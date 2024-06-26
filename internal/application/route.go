package application

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yuanyu90221/airline-order-system/internal/cache"
	"github.com/yuanyu90221/airline-order-system/internal/service/flight"
	"github.com/yuanyu90221/airline-order-system/internal/service/order"
)

// define route
func (app *App) loadRoutes() {
	gin.SetMode(app.config.GinMode)
	router := gin.Default()
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
	orderCacheStore := cache.NewCacheStore(app.rdb)
	orderHandler := order.NewHandler(orderCacheStore, app.bFilter)
	orderHandler.RegisterRoute(orderGroup)
}

// setup flight route
func (app *App) loadFlightRoutes() {
	flightGroup := app.router.Group("/flights")
	orderCacheStore := cache.NewCacheStore(app.rdb)
	flightStore := flight.NewFlightStore(app.db)
	flightHandler := flight.NewHandler(orderCacheStore, flightStore, app.bFilter)
	flightHandler.RegisterRoute(flightGroup)
}
