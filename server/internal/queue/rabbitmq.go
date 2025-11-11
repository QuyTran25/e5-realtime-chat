package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// RabbitMQ client wrapper
type RabbitMQ struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	queues  map[string]amqp.Queue
}

// Queue names
const (
	QueueMessageNotification = "message.notification"
	QueueEmailNotification   = "email.notification"
	QueueMessageProcessing   = "message.processing"
	QueueEventBroadcast      = "event.broadcast"
)

// NewRabbitMQ creates a new RabbitMQ client
func NewRabbitMQ(host, port, user, pass string) (*RabbitMQ, error) {
	url := fmt.Sprintf("amqp://%s:%s@%s:%s/", user, pass, host, port)

	// Connect with retry
	var conn *amqp.Connection
	var err error

	for i := 0; i < 5; i++ {
		conn, err = amqp.Dial(url)
		if err == nil {
			break
		}
		log.Printf("⚠️ RabbitMQ connection attempt %d failed: %v", i+1, err)
		time.Sleep(time.Second * 2)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ after retries: %w", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	// Set QoS
	err = channel.Qos(
		10,    // prefetch count
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to set QoS: %w", err)
	}

	mq := &RabbitMQ{
		conn:    conn,
		channel: channel,
		queues:  make(map[string]amqp.Queue),
	}

	// Declare all queues
	if err := mq.declareQueues(); err != nil {
		mq.Close()
		return nil, err
	}

	log.Println("✅ Connected to RabbitMQ successfully")
	return mq, nil
}

// declareQueues declares all required queues
func (mq *RabbitMQ) declareQueues() error {
	queueNames := []string{
		QueueMessageNotification,
		QueueEmailNotification,
		QueueMessageProcessing,
		QueueEventBroadcast,
	}

	for _, name := range queueNames {
		q, err := mq.channel.QueueDeclare(
			name,  // name
			true,  // durable
			false, // delete when unused
			false, // exclusive
			false, // no-wait
			nil,   // arguments
		)
		if err != nil {
			return fmt.Errorf("failed to declare queue %s: %w", name, err)
		}
		mq.queues[name] = q
		log.Printf("📮 Queue declared: %s", name)
	}

	return nil
}

// Publish publishes a message to a queue
func (mq *RabbitMQ) Publish(queueName string, message interface{}) error {
	body, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = mq.channel.PublishWithContext(
		ctx,
		"",        // exchange
		queueName, // routing key
		false,     // mandatory
		false,     // immediate
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         body,
			Timestamp:    time.Now(),
		},
	)

	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	return nil
}

// Consume consumes messages from a queue
func (mq *RabbitMQ) Consume(queueName string, handler func([]byte) error) error {
	msgs, err := mq.channel.Consume(
		queueName, // queue
		"",        // consumer
		false,     // auto-ack
		false,     // exclusive
		false,     // no-local
		false,     // no-wait
		nil,       // args
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	log.Printf("🔄 Started consuming from queue: %s", queueName)

	go func() {
		for msg := range msgs {
			err := handler(msg.Body)
			if err != nil {
				log.Printf("❌ Error processing message: %v", err)
				msg.Nack(false, true) // requeue
			} else {
				msg.Ack(false)
			}
		}
	}()

	return nil
}

// Close closes the connection
func (mq *RabbitMQ) Close() error {
	if mq.channel != nil {
		mq.channel.Close()
	}
	if mq.conn != nil {
		return mq.conn.Close()
	}
	return nil
}

// GetQueueStats returns queue statistics
func (mq *RabbitMQ) GetQueueStats(queueName string) (messages, consumers int, err error) {
	q, err := mq.channel.QueueInspect(queueName)
	if err != nil {
		return 0, 0, err
	}
	return q.Messages, q.Consumers, nil
}
