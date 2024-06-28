package order

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/yuanyu90221/airline-order-system/internal/types"
)

type OrderStore struct {
	db *sql.DB
}

func NewOrderStore(db *sql.DB) *OrderStore {
	return &OrderStore{db: db}
}
func (orderStore *OrderStore) CreateOrder(tx *sql.Tx, ctx context.Context, createOrderParam types.CreateOrderEntityParam) (types.Order, error) {
	queryBuilder := sq.Insert("orders").Columns("id", "flight_id", "wait_order", "ticket_numbers").Values(createOrderParam.ID,
		createOrderParam.FlightID, createOrderParam.WaitOrder, createOrderParam.TicketNumbers).Suffix("RETURNING *;").PlaceholderFormat(sq.Dollar)
	query, args, err := queryBuilder.ToSql()
	if err != nil {
		return types.Order{}, fmt.Errorf("create order query builder failed %w", err)
	}
	rows, err := tx.QueryContext(ctx, query, args...)
	if err != nil {
		return types.Order{}, fmt.Errorf("insert order failed %w", err)
	}
	var resultOrder types.Order
	for rows.Next() {
		err := rows.Scan(
			&resultOrder.ID,
			&resultOrder.FlightID,
			&resultOrder.PaidAt,
			&resultOrder.CanceledAt,
			&resultOrder.CreatedAt,
			&resultOrder.WaitOrder,
			&resultOrder.TicketNumbers,
		)
		if err != nil {
			return types.Order{}, fmt.Errorf("scan order failed %w", err)
		}
	}
	return resultOrder, nil
}

func (orderStore *OrderStore) GetOrderById(ctx context.Context, orderID uuid.UUID) (types.Order, error) {
	queryBuilder := sq.Select("id",
		"flight_id", "paid_at", "canceled_at",
		"created_at", "wait_order", "ticket_numbers").From("orders").Where(sq.Eq{"id": orderID}).PlaceholderFormat(sq.Dollar)
	query, args, err := queryBuilder.ToSql()
	if err != nil {
		return types.Order{}, fmt.Errorf("failed to create query string %w", err)
	}
	rows, err := orderStore.db.QueryContext(ctx, query, args...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return types.Order{}, fmt.Errorf("no order with id %s %w", orderID.String(), err)
		}
		return types.Order{}, fmt.Errorf("failed to query order %w", err)
	}
	var resultOrder types.Order
	for rows.Next() {
		err := rows.Scan(
			&resultOrder.ID,
			&resultOrder.FlightID,
			&resultOrder.PaidAt,
			&resultOrder.CanceledAt,
			&resultOrder.CreatedAt,
			&resultOrder.WaitOrder,
			&resultOrder.TicketNumbers,
		)
		if err != nil {
			return types.Order{}, fmt.Errorf("scan order failed %w", err)
		}
	}
	return resultOrder, nil
}
