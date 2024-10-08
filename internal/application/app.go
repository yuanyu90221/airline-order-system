package application

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	bloomfilter "github.com/alovn/go-bloomfilter"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"github.com/yuanyu90221/airline-order-system/internal/broker"
	"github.com/yuanyu90221/airline-order-system/internal/config"
	"github.com/yuanyu90221/airline-order-system/internal/db"
	"github.com/yuanyu90221/airline-order-system/internal/types"
	"github.com/yuanyu90221/airline-order-system/internal/util"
)

// define app dependency
type App struct {
	router      *gin.Engine
	rdb         *redis.Client
	config      *config.Config
	db          *sql.DB
	bFilter     bloomfilter.BloomFilter
	broker      *broker.Broker
	orderWorker types.Worker
}

func New(config *config.Config) *App {
	dbConn, err := db.Connect(config.DbURL)
	if err != nil {
		util.FailOnError(err, "failed to connect")
	}
	opts, err := redis.ParseURL(config.RedisUrl)
	if err != nil {
		util.FailOnError(err, "failed to parse redis url")
	}
	rdb := redis.NewClient(opts)
	broker, err := broker.NewBroker(config.RabbitMQURL)
	if err != nil {
		util.FailOnError(err, "failed to connect rabbitMq")
	}
	app := &App{
		rdb:     rdb,
		config:  config,
		db:      dbConn,
		bFilter: bloomfilter.NewRedisBloomFilter(rdb, "redis-bloom-filter", 100000),
		broker:  broker,
	}

	app.loadRoutes()
	app.loadOrderRoutes()
	app.loadFlightRoutes()
	app.setupOrderWorker()
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
		if err := app.broker.Close(); err != nil {
			log.Println("failed to close rabbitmq", err)
		}
	}()
	log.Printf("Starting server on %s", app.config.Port)
	errCh := make(chan error, 1)
	go func() {
		err = server.ListenAndServe()
		if err != nil {
			errCh <- fmt.Errorf("failed to start server: %w", err)
		}
		util.CloseChannel(errCh)
	}()
	go func() {
		err = app.orderWorker.Run(ctx)
		if err != nil {
			errCh <- fmt.Errorf("failed to run worker: %w", err)
		}
		util.CloseChannel(errCh)
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
