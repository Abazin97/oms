package rabbitmq

import (
	"context"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	OrderCreatedEvent = "order.created"
	OrderPaidEvent    = "order.paid"
)

func Connect(user, pass, host string, port string) (*amqp.Channel, func() error) {
	addr := fmt.Sprintf("amqp://%s:%s@%s:%s/", user, pass, host, port)

	conn, err := amqp.Dial(addr)
	if err != nil {
		log.Fatal(err)
	}

	ch, err := conn.Channel()
	if err != nil {
		log.Fatal(err)
	}

	err = ch.ExchangeDeclare(OrderCreatedEvent, "direct", true, false, false, false, nil)
	if err != nil {
		log.Fatal(err)
	}

	err = ch.ExchangeDeclare(OrderPaidEvent, "fanout", true, false, false, false, nil)
	if err != nil {
		log.Fatal(err)
	}

	return ch, conn.Close
}

func HandleRetry(ch *amqp.Channel, d *amqp.Delivery) error {
	if d.Headers == nil {
		d.Headers = make(amqp.Table)
	}

	retryCount, ok := d.Headers["retry_count"].(int64)
	if !ok {
		retryCount = 0
	}
	d.Headers["retry_count"] = retryCount + 1

	return ch.PublishWithContext(
		context.Background(),
		d.Exchange,
		d.RoutingKey,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			Headers:      d.Headers,
			Body:         d.Body,
			DeliveryMode: amqp.Persistent,
		})
}
