package application

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"github.com/yuanyu90221/airline-order-system/internal/config"
	"github.com/yuanyu90221/airline-order-system/internal/db"
	"github.com/yuanyu90221/airline-order-system/internal/util"
)

// define app dependency
type App struct {
	router *gin.Engine
	rdb    *redis.Client
	config *config.Config
	db     *sql.DB
}

func New(config *config.Config) *App {
	dbConn, err := db.Connect(config.DbURL)
	if err != nil {
		util.FailOnError(err, "failed to connect")
	}
	app := &App{
		rdb: redis.NewClient(&redis.Options{
			Addr:     config.RedisAddr,
			Password: config.RedisPassword,
		}),
		config: config,
		db:     dbConn,
	}

	app.loadRoutes()
	app.loadOrderRoutes()
	app.loadFlightRoutes()
	return app
}

func (app *App) Start(ctx context.Context) error {
	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", app.config.Port),
		Handler: app.router,
	}
	err := app.rdb.Ping(ctx).Err()
	if err != nil {
		return fmt.Errorf("failed to connect redis: %w", err)
	}
	// graceful shutdown close redis
	defer func() {
		if err := app.rdb.Close(); err != nil {
			log.Println("failed to close redis", err)
		}
		if err := app.db.Close(); err != nil {
			log.Println("failed to close db connection", err)
		}
	}()
	log.Printf("Starting server on %s", app.config.Port)
	errCh := make(chan error, 1)
	go func() {
		err = server.ListenAndServe()
		if err != nil {
			errCh <- fmt.Errorf("failed to start server: %w", err)
		}
		close(errCh)
	}()
	select {
	case err = <-errCh:
		return err
	case <-ctx.Done():
		timeout, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()
		return server.Shutdown(timeout)
	}
}
