package order

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	bloomfilter "github.com/alovn/go-bloomfilter"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/yuanyu90221/airline-order-system/internal/types"
	"github.com/yuanyu90221/airline-order-system/internal/util"
)

type Handler struct {
	cacheStore types.OrderCacheStore
	bFilter    bloomfilter.BloomFilter
}

func NewHandler(cacheStore types.OrderCacheStore, bFilter bloomfilter.BloomFilter) *Handler {
	return &Handler{
		cacheStore: cacheStore,
		bFilter:    bFilter,
	}
}

func (h *Handler) RegisterRoute(router *gin.RouterGroup) {
	router.POST("/", h.CreateOrder)
	router.GET("/:id", h.GetOrderById)
}

func (h *Handler) CreateOrder(ctx *gin.Context) {
	var requestOrder types.OrderCacheCreateRequest
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
	// create order from cache store
	result, err := h.cacheStore.CreateOrder(ctx, requestOrder)
	if err != nil {
		util.WriteError(ctx.Writer, http.StatusInternalServerError, fmt.Errorf("could not create order in cachestore: %w", err))
		return
	}
	if !result.IsValid {
		util.WriteError(ctx.Writer, http.StatusBadRequest, fmt.Errorf(`seats insufficient, could not create order with request ticket numbers: %d , with available seats %d, wait seats %d `, requestOrder.TicketNumbers, result.CurrentTotal, result.CurrentWait))
		return
	}
	// TODO: update result to order and flight
	util.FailOnError(util.WriteJSON(ctx.Writer, http.StatusCreated, result), "failed to write result")
}

func (h *Handler) GetOrderById(ctx *gin.Context) {
	// TODO: add get order by id
}
