package order

import (
	"context"
	"fmt"

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

/*
*
CreateOrder: create order with flight_id
*/
func (cache *CacheStore) CreateOrder(ctx context.Context, createOrderParam types.OrderCacheCreateRequest,
) (types.OrderCacheResult, error) {
	result := CreateOrderWithFlightID.Run(ctx, cache.rdb, []string{createOrderParam.FlightID},
		createOrderParam.TicketNumbers,
		createOrderParam.CurrentTotal,
		createOrderParam.CurrentWait,
		createOrderParam.CurrentWaitOrder)
	resultList, err := result.Int64Slice()
	if err != nil {
		return types.OrderCacheResult{}, fmt.Errorf("failed to createOrder with flightId: %s, %w", createOrderParam.FlightID, err)
	}
	return types.OrderCacheResult{
		CurrentTotal:     resultList[0],
		CurrentWait:      resultList[1],
		CurrentWaitOrder: resultList[2],
		IsValid:          resultList[3] == 1,
		IsWait:           resultList[4] == 1,
	}, nil
}

/*
*
GetCurrentRemain: get current flight_id remain
*/
func (cache *CacheStore) GetCurrentRemain(ctx context.Context, getRemainParam types.OrderCacheRequest) (types.OrderCacheRemain, error) {
	result := GetCurrentRemainWithFlightID.Run(ctx, cache.rdb, []string{getRemainParam.FlightID},
		getRemainParam.CurrentTotal, getRemainParam.CurrentWait, getRemainParam.CurrentWaitOrder)
	remain, err := result.Int64()
	if err != nil {
		return types.OrderCacheRemain{}, fmt.Errorf("failed to GetRemain with flightId: %s, %w", getRemainParam.FlightID, err)
	}
	return types.OrderCacheRemain{
		CurrentRemain: remain,
	}, nil
}

/*
*
CreateOrderWithFlightID: luascript for execute counter on specific flight_id
input key: flight_id, arguments: request, default_total, default_wait, default_wait_order
return {current_total, current_wait, current_wait_order, is_valid}
*
*/
var CreateOrderWithFlightID = redis.NewScript(`
local total_key = KEYS[1]..":total"
local wait_key = KEYS[1]..":wait"
local wait_order_key = KEYS[1]..":wait_order"
local request = ARGV[1]
local default_total = ARGV[2]
local default_wait = ARGV[3]
local default_wait_order = ARGV[4]
local total = redis.call("GET", total_key)
if not total then
	total = default_total
end
total = tonumber(total)
local wait = redis.call("GET", wait_key)
if not wait then
	wait = default_wait
end
wait = tonumber(wait)
local wait_order = redis.call("GET", wait_order_key)
if not wait_order then
	wait_order = default_wait_order
end
wait_order = tonumber(wait_order)
request = tonumber(request)
local is_valid = 1 
local is_wait = 0
if request < 0 then 
  is_valid = 0
	return {total, wait, wait_order, is_valid, is_wait}
end
if request > total and request > wait then
  is_valid = 0
	return {total, wait, wait_order, is_valid, is_wait}
end
if total + wait >= request and request > 0 then
  if total >= request then
	  total = total - request
	elseif wait >= request then
	  wait = wait - request
		wait_order = wait_order + 1
		is_wait = 1
	end
end
redis.call("SET", total_key, total)
redis.call("SET", wait_key, wait)
redis.call("SET", wait_order_key, wait_order)
return {total, wait, wait_order, is_valid, is_wait}
`)

/*
*
luascript for execute counter on specific flight_id
*
*/
var GetCurrentRemainWithFlightID = redis.NewScript(`
local total_key = KEYS[1]..":total"
local wait_key = KEYS[1]..":wait"
local wait_order_key = KEYS[1]..":wait_order"
local default_total = ARGV[1]
local default_wait = ARGV[2]
local default_wait_order = ARGV[3]
local total = redis.call("GET", total_key)
if not total then
	total = default_total
end
local wait = redis.call("GET", wait_key)
if not wait then
	wait = default_wait
end
local wait_order = redis.call("GET", wait_order_key)
if not wait_order then
	wait_order = default_wait_order
end
total = tonumber(total)
wait = tonumber(wait)
wait_order = tonumber(wait_order)
redis.call("SET", total_key, total)
redis.call("SET", wait_key, wait)
redis.call("SET", wait_order_key, wait_order)
remain = tonumber(total) + tonumber(wait)
return remain
`)
