package broker

import (
	"context"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Broker struct {
	conn *amqp.Connection
	ch   *amqp.Channel
}

func NewBroker(uri string) (*Broker, error) {
	conn, err := amqp.Dial(uri)
	if err != nil {
		return nil, fmt.Errorf("failed to connect rabbitmq with %s %w", uri, err)
	}
	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to create channel with %s %w", uri, err)
	}
	return &Broker{
		conn: conn,
		ch:   ch,
	}, nil
}
func (broker *Broker) HandleConnectCh() error {
	ch, err := broker.conn.Channel()
	if err != nil {
		return fmt.Errorf("handle connect ch failed %w", err)
	}
	broker.ch = ch
	return nil
}
func (broker *Broker) GenerateDeliveryChannel(ctx context.Context, qName string) (<-chan amqp.Delivery, error) {
	if broker.ch.IsClosed() {
		if err := broker.HandleConnectCh(); err != nil {
			return nil, err
		}
	}
	msgch, err := broker.ch.ConsumeWithContext(ctx, qName, "", false, false, false, false, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to consume queue %s %w", qName, err)
	}
	return msgch, err
}
func (broker *Broker) SendMessageToQueue(ctx context.Context, qName string, data []byte) error {
	if broker.ch.IsClosed() {
		if err := broker.HandleConnectCh(); err != nil {
			return err
		}
	}
	queue, err := broker.ch.QueueDeclare(qName, false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("declare queue failed with %s %w", qName, err)
	}
	// setup timeout for send queue
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	err = broker.ch.PublishWithContext(
		ctx,
		"",
		queue.Name,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        []byte(data),
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish message to queue %w", err)
	}
	return nil
}
func (broker *Broker) Close() error {
	if broker.ch != nil && !broker.ch.IsClosed() {
		return broker.ch.Close()
	}
	if broker.conn != nil && !broker.conn.IsClosed() {
		return broker.conn.Close()
	}
	broker.ch = nil
	broker.conn = nil
	return nil
}
