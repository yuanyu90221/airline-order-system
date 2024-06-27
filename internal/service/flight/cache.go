package flight

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/yuanyu90221/airline-order-system/internal/types"
)

type CacheStore struct {
	rdb *redis.Client
}

func NewCacheStore(rdb *redis.Client) *CacheStore {
	return &CacheStore{
		rdb: rdb,
	}
}

func (cacheStore *CacheStore) UpdateFlight(ctx context.Context, flightInfo types.Flight) (types.Flight, error) {
	// start redis multi(transaction)
	tx := cacheStore.rdb.TxPipeline()
	// use flight id as key
	flightID := flightInfo.ID.String()
	// clear flight previous record
	tx.Del(ctx, flightID)
	tx.Del(ctx, fmt.Sprintf("%s:total", flightID))
	tx.Del(ctx, fmt.Sprintf("%s:wait", flightID))
	tx.Del(ctx, fmt.Sprintf("%s:wait_order", flightID))
	// serialize flightInfo
	jsonData, err := json.Marshal(flightInfo)
	if err != nil {
		return types.Flight{}, fmt.Errorf("marshal flightInfo err %w", err)
	}
	// setup duration = flightInfo.FlightDate - now
	expiredDuration := flightInfo.FlightDate.Sub(time.Now().UTC())
	// update flight record
	tx.Set(ctx, flightID, jsonData, expiredDuration)
	_, err = tx.Exec(ctx)
	if err != nil {
		return types.Flight{}, fmt.Errorf("failed to exec update : %w", err)
	}
	return flightInfo, nil
}

func (cacheStore *CacheStore) GetFlightCacheInfo(ctx context.Context, flightID string) (types.Flight, error) {
	result := cacheStore.rdb.Get(ctx, flightID)
	var flightInfo types.Flight
	resultBody, err := result.Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return types.Flight{}, fmt.Errorf("flightID %s is not in cache store %w", flightID, err)
		}
		return types.Flight{}, fmt.Errorf("get flight result error %w", err)
	}
	err = json.Unmarshal(resultBody, &flightInfo)
	if err != nil {
		return types.Flight{}, fmt.Errorf("unmarshal flight result error %w", err)
	}
	return flightInfo, nil
}
