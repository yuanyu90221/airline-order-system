package order

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	bloomfilter "github.com/alovn/go-bloomfilter"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/yuanyu90221/airline-order-system/internal/broker"
	"github.com/yuanyu90221/airline-order-system/internal/config"
	"github.com/yuanyu90221/airline-order-system/internal/types"
	"github.com/yuanyu90221/airline-order-system/internal/util"
)

type Handler struct {
	orderCacheStore  types.OrderCacheStore
	flightCacheStore types.FlightCacheStore
	bFilter          bloomfilter.BloomFilter
	mq               *broker.Broker
	orderStore       types.OrderStore
}

func NewHandler(orderCacheStore types.OrderCacheStore, flightCacheStore types.FlightCacheStore,
	bFilter bloomfilter.BloomFilter, mq *broker.Broker, orderStore types.OrderStore) *Handler {
	return &Handler{
		orderCacheStore:  orderCacheStore,
		flightCacheStore: flightCacheStore,
		bFilter:          bFilter,
		mq:               mq,
		orderStore:       orderStore,
	}
}

func (h *Handler) RegisterRoute(router *gin.RouterGroup) {
	router.POST("/", h.CreateOrder)
	router.GET("/:id", h.GetOrderById)
}

func (h *Handler) CreateOrder(ctx *gin.Context) {
	var requestOrder types.CreateOrderRequest
	// load input
	if err := util.ParseJSON(ctx.Request, &requestOrder); err != nil {
		util.WriteError(ctx.Writer, http.StatusBadRequest, err)
		return
	}
	// validate input
	if err := util.Validdate.Struct(requestOrder); err != nil {
		var valErrs validator.ValidationErrors
		if errors.As(err, &valErrs) {
			util.WriteError(ctx.Writer, http.StatusBadRequest, fmt.Errorf("invalid payload:%v", valErrs))
		}
		return
	}
	log.Println("requestOrder", requestOrder)
	// use bloomfilter to check flightID exists
	binaryFlightID, status, err := util.ParseFlightIDIntoBinary(requestOrder.FlightID)
	if err != nil {
		util.WriteError(ctx.Writer, status, err)
		return
	}
	isExist, err := h.bFilter.MightContain(binaryFlightID)
	if err != nil {
		util.WriteError(ctx.Writer, http.StatusInternalServerError, fmt.Errorf("failed to check FlightID in bloomfilter %w", err))
		return
	}
	if !isExist {
		util.WriteError(ctx.Writer, http.StatusBadRequest, fmt.Errorf("FlightID %s not in bloomfilter", requestOrder.FlightID))
		return
	}
	flightInfo, err := h.flightCacheStore.GetFlightCacheInfo(ctx, requestOrder.FlightID)
	if err != nil {
		util.WriteError(ctx.Writer, http.StatusBadRequest, fmt.Errorf("FlightID %s not in flight cache %w", requestOrder.FlightID, err))
		return
	}
	cacheRequest := types.OrderCacheRequest{
		FlightID:         requestOrder.FlightID,
		CurrentTotal:     int64(flightInfo.AvailableSeats),
		CurrentWait:      int64(flightInfo.WaitSeats),
		CurrentWaitOrder: int64(flightInfo.NextWaitOrder),
	}
	// create order from cache store
	result, err := h.orderCacheStore.CreateOrder(ctx, types.OrderCacheCreateRequest{
		OrderCacheRequest: cacheRequest,
		TicketNumbers:     requestOrder.TicketNumbers,
	})
	if err != nil {
		util.WriteError(ctx.Writer, http.StatusInternalServerError, fmt.Errorf("could not create order in cachestore: %w", err))
		return
	}
	if !result.IsValid {
		util.WriteError(ctx.Writer, http.StatusBadRequest, fmt.Errorf(`seats insufficient, could not create order with request ticket numbers: %d , with available seats %d, wait seats %d `, requestOrder.TicketNumbers, result.CurrentTotal, result.CurrentWait))
		return
	}
	// update result to rabbitmq
	requestEvent := types.CreateOrderEvent{
		FlightID:       requestOrder.FlightID,
		TicketNumbers:  requestOrder.TicketNumbers,
		AvailableSeats: result.CurrentTotal,
		WaitOrder:      result.CurrentWaitOrder,
		WaitSeats:      result.CurrentWait,
		IsWait:         result.IsWait,
	}

	data, err := json.Marshal(requestEvent)
	if err != nil {
		util.WriteError(ctx.Writer, http.StatusInternalServerError, fmt.Errorf("marshal data error %w", err))
		return
	}
	err = h.mq.SendMessageToQueue(ctx, config.AppConfig.OrderQueueName, data)
	if err != nil {
		util.WriteError(ctx.Writer, http.StatusInternalServerError, fmt.Errorf("send rabbitmq error %w", err))
		return
	}
	if !result.IsWait {
		requestEvent.WaitOrder = -1
	}
	util.FailOnError(util.WriteJSON(ctx.Writer, http.StatusCreated, requestEvent), "failed to write result")
}

func (h *Handler) GetOrderById(ctx *gin.Context) {
	orderID := ctx.Param("id")
	if orderID == "" {
		util.WriteError(ctx.Writer, http.StatusBadRequest, fmt.Errorf("flight id not provided"))
		return
	}
	id, err := uuid.Parse(orderID)
	if err != nil {
		util.WriteError(ctx.Writer, http.StatusBadRequest, fmt.Errorf("failed to parse id %s into uuid %w", orderID, err))
		return
	}
	result, err := h.orderStore.GetOrderById(ctx, id)
	if err != nil {
		util.WriteError(ctx.Writer, http.StatusInternalServerError, fmt.Errorf("failed to get order %w", err))
		return
	}
	util.FailOnError(util.WriteJSON(ctx.Writer, http.StatusOK, result), "failed to response json")
}
