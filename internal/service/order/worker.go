package order

import (
	"context"
	"encoding/json"
	"log"

	"github.com/google/uuid"
	"github.com/yuanyu90221/airline-order-system/internal/broker"
	"github.com/yuanyu90221/airline-order-system/internal/config"
	"github.com/yuanyu90221/airline-order-system/internal/types"
)

type OrderWorker struct {
	orderService     types.OrderServcie
	flightCacheStore types.FlightCacheStore
	mq               *broker.Broker
}

func NewOrderWorker(orderService types.OrderServcie, flightCacheStore types.FlightCacheStore,
	mq *broker.Broker,
) *OrderWorker {
	return &OrderWorker{
		orderService:     orderService,
		flightCacheStore: flightCacheStore,
		mq:               mq,
	}
}

func (orderWorker *OrderWorker) Run(ctx context.Context) error {
	msgch, err := orderWorker.mq.GenerateDeliveryChannel(ctx, config.AppConfig.OrderQueueName)
	log.Println("worker start")
	if err != nil {
		return err
	}
	for msg := range msgch {
		data := msg.Body
		var createOrderEvent types.CreateOrderEvent
		err := json.Unmarshal(data, &createOrderEvent)
		if err != nil {
			log.Println("unmarchal event failed", err)
			continue
		}
		log.Println(createOrderEvent)
		flightID, err := uuid.Parse(createOrderEvent.FlightID)
		if err != nil {
			log.Println("parse flightID failed:", err)
			continue
		}
		ID, err := uuid.Parse(createOrderEvent.ID)
		if err != nil {
			log.Println("parse flightID failed:", err)
			continue
		}
		createOrderParam := types.CreateOrderEntityParam{
			ID:            ID,
			FlightID:      flightID,
			WaitOrder:     int32(createOrderEvent.WaitOrder),
			TicketNumbers: int32(createOrderEvent.TicketNumbers),
		}

		updateFlightParams := types.UpdateFlightEntityParam{
			ID:             flightID,
			AvailableSeats: int32(createOrderEvent.AvailableSeats),
			WaitSeats:      int32(createOrderEvent.WaitSeats),
			NextWaitOrder:  int32(createOrderEvent.WaitOrder),
		}
		flight, order, err := orderWorker.orderService.CreateOrderHandler(ctx, createOrderParam, updateFlightParams)
		if err != nil {
			log.Printf("failed to create order %v", err)
			continue
		}
		_, err = orderWorker.flightCacheStore.UpdateFlight(ctx, flight)
		if err != nil {
			log.Printf("faield to update flight cache %v", err)
			continue
		}
		log.Printf("finish update flight: %v\n order: %v\n", flight, order)
	}
	log.Println("worker end")
	<-ctx.Done()
	return nil
}
