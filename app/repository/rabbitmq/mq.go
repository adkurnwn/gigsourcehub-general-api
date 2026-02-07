package rabbitmqrepo

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/adkurnwn/gigsourcehub-general-api/domain"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
)

type rabbitMQRepo struct {
	conn *amqp.Connection
	ch   *amqp.Channel
}

func NewRabbitMQRepo(url string) (domain.MessageBroker, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open a channel: %w", err)
	}

	return &rabbitMQRepo{
		conn: conn,
		ch:   ch,
	}, nil
}

func (r *rabbitMQRepo) Publish(ctx context.Context, queueName string, message interface{}) error {
	// 1. Declare Queue (idempotent)
	q, err := r.ch.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare a queue: %w", err)
	}

	// 2. Marshal Message
	body, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// 3. Publish
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err = r.ch.PublishWithContext(ctx,
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
			Timestamp:   time.Now(),
		})
	if err != nil {
		return fmt.Errorf("failed to publish a message: %w", err)
	}

	logrus.Infof("Sent message to queue %s", queueName)
	return nil
}

func (r *rabbitMQRepo) Close() error {
	if r.ch != nil {
		r.ch.Close()
	}
	if r.conn != nil {
		return r.conn.Close()
	}
	return nil
}
