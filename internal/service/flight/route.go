package flight

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/yuanyu90221/airline-order-system/internal/types"
	"github.com/yuanyu90221/airline-order-system/internal/util"
)

type Handler struct {
	cacheStore  types.OrderCacheStore
	flightStore types.FlightStore
}

func NewHandler(cacheStore types.OrderCacheStore, flightStore types.FlightStore) *Handler {
	return &Handler{
		cacheStore:  cacheStore,
		flightStore: flightStore,
	}
}
func (h *Handler) RegisterRoute(router *gin.RouterGroup) {
	router.POST("/", h.CreateFlight)
}
func (h *Handler) CreateFlight(ctx *gin.Context) {
	var createFlight types.CreateFlightParams
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
	log.Println("createFlight", createFlight)
	flight, err := h.flightStore.CreateFlight(ctx, createFlight)
	if err != nil {
		util.WriteError(ctx.Writer, http.StatusInternalServerError, err)
		return
	}
	util.FailOnError(util.WriteJSON(ctx.Writer, http.StatusCreated, flight), "failed to response json")
}
