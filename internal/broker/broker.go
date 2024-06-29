package broker

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Broker struct {
	uri             string
	consumer_conn   *amqp.Connection
	publisher_conn  *amqp.Connection
	consumer_ch     *amqp.Channel
	publisher_ch    *amqp.Channel
	publisher_mutex sync.RWMutex
	consumer_mutex  sync.RWMutex
	broker_mutex    sync.RWMutex
}

func NewBroker(uri string) (*Broker, error) {
	consumer_conn, err := amqp.DialConfig(uri, amqp.Config{
		Properties: map[string]interface{}{"connection_name": "consumer"},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect consumer rabbitmq with %s %w", uri, err)
	}
	consumer_ch, err := consumer_conn.Channel()
	if err != nil {
		defer func() {
			consumer_conn.Close()
		}()
		return nil, fmt.Errorf("failed to create consumer channel with %s %w", uri, err)
	}
	publisher_conn, err := amqp.DialConfig(uri, amqp.Config{
		Properties: map[string]interface{}{"connection_name": "publisher"},
	})
	if err != nil {
		defer func() {
			consumer_ch.Close()
			consumer_conn.Close()
		}()
		return nil, fmt.Errorf("failed to connect publisher rabbitmq with %s %w", uri, err)
	}
	publisher_ch, err := publisher_conn.Channel()
	if err != nil {
		defer func() {
			consumer_ch.Close()
			consumer_conn.Close()
			publisher_conn.Close()
		}()
		return nil, fmt.Errorf("failed to create publisher channel with %s %w", uri, err)
	}
	return &Broker{
		uri:            uri,
		consumer_conn:  consumer_conn,
		consumer_ch:    consumer_ch,
		publisher_conn: publisher_conn,
		publisher_ch:   publisher_ch,
	}, nil
}
func (broker *Broker) HandlePublisherReconnect() error {
	broker.publisher_mutex.Lock()
	defer broker.publisher_mutex.Unlock()
	if broker.publisher_conn != nil {
		err := broker.publisher_conn.Close()
		if err != nil {
			return err
		}
		broker.publisher_conn = nil
	}
	conn, err := amqp.DialConfig(broker.uri, amqp.Config{
		Properties: map[string]interface{}{"connection_name": "publisher"},
	})
	if err != nil {
		defer conn.Close()
		return fmt.Errorf("handle connect failed %w", err)
	}
	broker.publisher_conn = conn
	return nil
}
func (broker *Broker) HandlePublisherConnectCh() error {
	broker.publisher_mutex.Lock()
	defer broker.publisher_mutex.Unlock()
	if broker.publisher_conn == nil || broker.publisher_conn.IsClosed() {
		if err := broker.HandlePublisherReconnect(); err != nil {
			log.Printf("try to reconnect publisher failed %v", err)
			return err
		}
	}
	ch, err := broker.publisher_conn.Channel()
	if err != nil {
		defer ch.Close()
		return fmt.Errorf("handle connect ch failed %w", err)
	}
	broker.publisher_ch = ch
	return nil
}
func (broker *Broker) HandleConsumerReconnect() error {
	broker.publisher_mutex.Lock()
	defer broker.publisher_mutex.Unlock()
	if broker.consumer_conn != nil {
		err := broker.consumer_conn.Close()
		if err != nil {
			return err
		}
		broker.consumer_conn = nil
	}
	conn, err := amqp.DialConfig(broker.uri, amqp.Config{
		Properties: map[string]interface{}{"connection_name": "consumer"},
	})
	if err != nil {
		defer conn.Close()
		return fmt.Errorf("handle connect failed %w", err)
	}
	broker.consumer_conn = conn
	return nil
}
func (broker *Broker) HandleConsumerConnectCh() error {
	broker.consumer_mutex.Lock()
	defer broker.consumer_mutex.Unlock()
	if broker.consumer_conn == nil || broker.consumer_conn.IsClosed() {
		if err := broker.HandleConsumerReconnect(); err != nil {
			return err
		}
	}
	ch, err := broker.consumer_conn.Channel()
	if err != nil {
		defer ch.Close()
		return fmt.Errorf("handle connect ch failed %w", err)
	}
	broker.consumer_ch = ch
	return nil
}
func (broker *Broker) ConsumerClose() error {
	broker.consumer_mutex.Lock()
	defer broker.consumer_mutex.Unlock()
	if broker.consumer_ch != nil && !broker.consumer_ch.IsClosed() {
		if err := broker.consumer_ch.Close(); err != nil {
			return err
		}
	}
	if broker.consumer_conn != nil && !broker.consumer_conn.IsClosed() {
		if err := broker.consumer_conn.Close(); err != nil {
			return err
		}
	}
	broker.consumer_ch = nil
	broker.consumer_conn = nil
	return nil
}
func (broker *Broker) GenerateDeliveryChannel(ctx context.Context, qName string) (<-chan amqp.Delivery, error) {
	broker.publisher_mutex.Lock()
	defer broker.publisher_mutex.Unlock()
	if broker.consumer_ch == nil || broker.consumer_ch.IsClosed() {
		if err := broker.HandleConsumerConnectCh(); err != nil {
			return nil, err
		}
	}
	queue, err := broker.consumer_ch.QueueDeclare(qName, false, false, false, false, nil)
	if err != nil {
		return nil, fmt.Errorf("declare queue failed with %s %w", qName, err)
	}
	msgch, err := broker.consumer_ch.ConsumeWithContext(ctx, queue.Name, "", false, false, false, false, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to consume queue %s %w", queue.Name, err)
	}
	return msgch, err
}
func (broker *Broker) SendMessageToQueue(ctx context.Context, qName string, data []byte) error {
	broker.publisher_mutex.Lock()
	defer broker.publisher_mutex.Unlock()
	if broker.publisher_ch == nil || broker.publisher_ch.IsClosed() {
		if err := broker.HandlePublisherConnectCh(); err != nil {
			return err
		}
	}
	queue, err := broker.publisher_ch.QueueDeclare(qName, false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("declare queue failed with %s %w", qName, err)
	}
	// setup timeout for send queue
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	err = broker.publisher_ch.PublishWithContext(
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
		log.Printf("failed to publish message to queue %v\n", err)
		defer broker.publisher_ch.Close()
		return fmt.Errorf("failed to publish message to queue %w", err)
	}
	return nil
}
func (broker *Broker) PublisherClose() error {
	broker.publisher_mutex.Lock()
	defer broker.publisher_mutex.Unlock()

	if broker.publisher_ch != nil && !broker.publisher_ch.IsClosed() {
		if err := broker.publisher_ch.Close(); err != nil {
			return err
		}
	}
	if broker.publisher_conn != nil && !broker.publisher_conn.IsClosed() {
		if err := broker.publisher_conn.Close(); err != nil {
			return err
		}
	}
	broker.publisher_ch = nil
	broker.publisher_conn = nil
	return nil
}
func (broker *Broker) Close() error {
	broker.broker_mutex.Lock()
	defer broker.broker_mutex.Unlock()
	err := broker.PublisherClose()
	if err != nil {
		return err
	}
	err = broker.ConsumerClose()
	if err != nil {
		return err
	}
	return nil
}
