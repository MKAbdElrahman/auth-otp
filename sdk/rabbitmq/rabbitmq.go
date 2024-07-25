package rabbitmq

import (
	"context"
	"log"

	"github.com/streadway/amqp"
)

type RabbitMQBroker struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

func NewRabbitMQBroker(dsn string) (*RabbitMQBroker, error) {
	conn, err := amqp.Dial(dsn)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	return &RabbitMQBroker{
		conn:    conn,
		channel: ch,
	}, nil
}

func (r *RabbitMQBroker) Publish(ctx context.Context, topic string, message []byte) error {
	_, err := r.channel.QueueDeclare(
		topic, // queue name
		true,  // durable
		false, // auto-delete
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return err
	}

	return r.channel.Publish(
		"",    // exchange
		topic, // routing key
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        message,
		})
}

func (r *RabbitMQBroker) Subscribe(ctx context.Context, topic string, handler func(ctx context.Context, message []byte)) error {
	_, err := r.channel.QueueDeclare(
		topic, // queue name
		true,  // durable
		false, // auto-delete
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return err
	}

	msgs, err := r.channel.Consume(
		topic, // queue
		"",    // consumer
		false, // auto-ack (set to false to manually ack messages)
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case d, ok := <-msgs:
				if !ok {
					return
				}

				// Pass context to the handler
				handler(ctx, d.Body)

				// Acknowledge the message after handling
				if err := d.Ack(false); err != nil {
					// Log the error and handle it appropriately
					log.Printf("Failed to acknowledge message: %v", err)
				}
			case <-ctx.Done():
				// Handle context cancellation (e.g., log, clean up)
				log.Println("Context cancelled, stopping message consumption")
				return
			}
		}
	}()

	return nil
}

func (r *RabbitMQBroker) Close() error {
	if err := r.channel.Close(); err != nil {
		return err
	}
	return r.conn.Close()
}
