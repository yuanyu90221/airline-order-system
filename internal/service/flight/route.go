package flight

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	bloomfilter "github.com/alovn/go-bloomfilter"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/yuanyu90221/airline-order-system/internal/types"
	"github.com/yuanyu90221/airline-order-system/internal/util"
)

type Handler struct {
	orderCacheStore  types.OrderCacheStore
	flightCacheStore types.FlightCacheStore
	flightStore      types.FlightStore
	bFilter          bloomfilter.BloomFilter
}

func NewHandler(orderCacheStore types.OrderCacheStore, flightCacheStore types.FlightCacheStore,
	flightStore types.FlightStore, bFilter bloomfilter.BloomFilter) *Handler {
	return &Handler{
		orderCacheStore:  orderCacheStore,
		flightCacheStore: flightCacheStore,
		flightStore:      flightStore,
		bFilter:          bFilter,
	}
}
func (h *Handler) RegisterRoute(router *gin.RouterGroup) {
	router.POST("/", h.CreateFlight)
	router.GET("/", h.GetFlightsByCriteria)
	router.GET("/:id", h.GetFlightById)
}
func (h *Handler) CreateFlight(ctx *gin.Context) {
	var createFlight types.CreateFlightRequest
	if err := util.ParseJSON(ctx.Request, &createFlight); err != nil {
		util.WriteError(ctx.Writer, http.StatusBadRequest, err)
		return
	}
	if err := util.Validdate.Struct(createFlight); err != nil {
		var valErrs validator.ValidationErrors
		if errors.As(err, &valErrs) {
			util.WriteError(ctx.Writer, http.StatusBadRequest, fmt.Errorf("invalid payload:%v", valErrs))
		}
		return
	}
	// log.Println("createFlight", createFlight)
	flight, err := h.flightStore.CreateFlight(ctx, createFlight)
	if err != nil {
		util.WriteError(ctx.Writer, http.StatusInternalServerError, err)
		return
	}
	binaryUUID, err := flight.ID.MarshalBinary()
	if err != nil {
		util.WriteError(ctx.Writer, http.StatusInternalServerError, fmt.Errorf("uuid marshal binnary err %w", err))
		return
	}
	if err := h.bFilter.Put(binaryUUID); err != nil {
		util.WriteError(ctx.Writer, http.StatusInternalServerError, fmt.Errorf("bloom filter put err %w", err))
		return
	}
	_, err = h.flightCacheStore.UpdateFlight(ctx, flight)
	if err != nil {
		util.WriteError(ctx.Writer, http.StatusInternalServerError, fmt.Errorf("failed to update flight err %w", err))
		return
	}
	util.FailOnError(util.WriteJSON(ctx.Writer, http.StatusCreated, flight), "failed to response json")
}

func (h *Handler) GetFlightsByCriteria(ctx *gin.Context) {
	// get params from query
	pagination := types.Pagination{
		Offset: 0,
		Limit:  10,
	}
	query := ctx.Request.URL.Query()
	if query.Has("limit") {
		limit, err := strconv.ParseInt(query.Get("limit"), 10, 64)
		if err != nil {
			util.WriteError(ctx.Writer, http.StatusBadRequest, fmt.Errorf("limit parse err: %w", err))
			return
		}
		pagination.Limit = limit
	}
	if query.Has("offset") {
		offset, err := strconv.ParseInt(query.Get("offset"), 10, 64)
		if err != nil {
			util.WriteError(ctx.Writer, http.StatusBadRequest, fmt.Errorf("offset parse err: %w", err))
			return
		}
		pagination.Offset = offset
	}
	var queryParams types.QueryFlightRequest
	if query.Has("flignt_date") {
		flignt_date, err := strconv.ParseInt(query.Get("flignt_date"), 10, 64)
		if err != nil {
			util.WriteError(ctx.Writer, http.StatusBadRequest, fmt.Errorf("flignt_date parse err: %w", err))
			return
		}
		queryParams.FlightDate = flignt_date
	}
	if query.Has("destination") {
		queryParams.Destination = query.Get("destination")
	}
	if query.Has("departure") {
		queryParams.Departure = query.Get("departure")
	}
	result, err := h.flightStore.GetFlightsByCriteria(ctx, queryParams, pagination)
	if err != nil {
		util.WriteError(ctx.Writer, http.StatusInternalServerError, err)
		return
	}
	util.FailOnError(util.WriteJSON(ctx.Writer, http.StatusOK, result), "failed on response json")
}

func (h *Handler) GetFlightById(ctx *gin.Context) {
	flightID := ctx.Param("id")
	if flightID == "" {
		util.WriteError(ctx.Writer, http.StatusBadRequest, fmt.Errorf("flight id not provided"))
		return
	}
	id, err := uuid.Parse(flightID)
	if err != nil {
		util.WriteError(ctx.Writer, http.StatusBadRequest, fmt.Errorf("failed to parse id %s into uuid %w", flightID, err))
		return
	}
	result, err := h.flightStore.GetFlightById(ctx, id)
	if err != nil {
		util.WriteError(ctx.Writer, http.StatusInternalServerError, fmt.Errorf("failed to get flight by id"))
		return
	}
	util.FailOnError(util.WriteJSON(ctx.Writer, http.StatusOK, result), "failed to response json")
}
