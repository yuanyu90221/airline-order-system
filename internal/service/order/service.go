package order

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/yuanyu90221/airline-order-system/internal/types"
)

// handle create order
type OrderService struct {
	db          *sql.DB
	orderStore  types.OrderStore
	flightStore types.FlightStore
}

func NewOrderService(db *sql.DB, orderStore types.OrderStore, flightStore types.FlightStore) *OrderService {
	return &OrderService{
		db:          db,
		orderStore:  orderStore,
		flightStore: flightStore,
	}
}

func (orderService *OrderService) CreateOrderHandler(ctx context.Context,
	createOrderParams types.CreateOrderEntityRequest,
	updateFlightParams types.UpdateFlightEntityRequest,
) (types.Flight, types.Order, error) {
	// create db transaction
	tx, err := orderService.db.BeginTx(ctx, nil)
	if err != nil {
		return types.Flight{}, types.Order{}, fmt.Errorf("create db tx failed %w", err)
	}
	log.Println(createOrderParams, updateFlightParams)
	order, err := orderService.orderStore.CreateOrder(tx, ctx, createOrderParams)
	if err != nil {
		log.Printf("failed to create order %v", err)
		err = tx.Rollback()
		if err != nil {
			return types.Flight{}, types.Order{}, fmt.Errorf("tx roolback failed %w", err)
		}
		return types.Flight{}, types.Order{}, err
	}
	flight, err := orderService.flightStore.UpdateFlight(tx, ctx, updateFlightParams)
	if err != nil {
		log.Printf("failed to create order %v", err)
		err = tx.Rollback()
		if err != nil {
			return types.Flight{}, types.Order{}, fmt.Errorf("tx roolback failed %w", err)
		}
		return types.Flight{}, types.Order{}, err
	}
	if err := tx.Commit(); err != nil {
		return types.Flight{}, types.Order{}, err
	}
	return flight, order, nil
}
