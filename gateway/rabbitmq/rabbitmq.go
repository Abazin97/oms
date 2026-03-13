package rabbitmq

import (
	"context"
	"fmt"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	OrderExchange = "order.exchange"

	OrderCreatedEvent           = "orders.created"
	PaymentCreatedEvent         = "payments.created"
	OrderPaidEvent              = "orders.paid"
	OrderCanceledEvent          = "orders.canceled"
	StockReservedEvent          = "stock.reserved"
	StockStatusChangedEvent     = "stock.status_changed"
	StockReservationFailedEvent = "stock.reservation_failed"

	MaxRetryCount = 3
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

	err = ch.ExchangeDeclare(OrderExchange, "topic", true, false, false, false, nil)
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

	if retryCount >= MaxRetryCount {
		log.Printf("message exceeded retry limit (%d), dropping", retryCount)
		return nil
	}
	retryCount++
	d.Headers["retry_count"] = retryCount

	time.Sleep(time.Second * time.Duration(retryCount))

	return ch.PublishWithContext(
		context.Background(),
		OrderExchange,
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
