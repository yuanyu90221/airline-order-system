package flight

import (
	"context"

	"github.com/yuanyu90221/airline-order-system/internal/types"
)

// sync data change while order event happend
type FlightService struct {
	flightCacheStore types.FlightCacheStore
}

func NewFlightService(flightCacheStore types.FlightCacheStore) *FlightService {
	return &FlightService{
		flightCacheStore: flightCacheStore,
	}
}

func (flightService *FlightService) UpdateFlightCache(ctx context.Context, flightInfo types.Flight) (types.Flight, error) {
	return flightService.flightCacheStore.UpdateFlight(ctx, flightInfo)
}
