package consumer

import (
	"context"
	"encoding/json"
	"gateway/rabbitmq"
	"log"
	"orders/internal/events"
	"orders/internal/services"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Consumer struct {
	service services.OrdersService
}

func NewConsumer(service services.OrdersService) *Consumer {
	return &Consumer{service: service}
}

func (c *Consumer) Listen(ctx context.Context, ch *amqp.Channel) {
	q, err := ch.QueueDeclare(
		"orders.order.queue", true, false, true, false, nil)
	if err != nil {
		log.Fatal(err)
	}

	err = ch.QueueBind(
		q.Name,
		rabbitmq.PaymentCreatedEvent,
		rabbitmq.OrderExchange,
		false,
		nil,
	)

	err = ch.QueueBind(
		q.Name,
		rabbitmq.OrderPaidEvent,
		rabbitmq.OrderExchange,
		false,
		nil,
	)
	if err != nil {
		log.Fatal(err)
	}

	msgs, err := ch.Consume(q.Name, "", false, false, false, false, nil)
	if err != nil {
		log.Fatal(err)
	}

	var forever = make(chan struct{})

	go func() {
		for {
			select {
			case d, ok := <-msgs:
				if !ok {
					return
				}

				switch d.RoutingKey {

				case rabbitmq.PaymentCreatedEvent:

					var event events.PaymentCreatedEvent
					if err := json.Unmarshal(d.Body, &event); err != nil {
						log.Println(err)
						d.Nack(false, false)
						continue
					}

					err := c.service.UpdatePaymentLink(ctx, event.OrderID, event.PaymentURL)
					if err != nil {
						log.Printf("UpdatePaymentLink error: %s", err)

						if err := rabbitmq.HandleRetry(ch, &d); err != nil {
							log.Printf("Retry error: %s", err)
						}

						d.Nack(false, false)
						continue
					}
					log.Printf("Payment link updated for order %s", event.OrderID)

				case rabbitmq.OrderPaidEvent:

					var event events.OrderPaidEvent
					if err := json.Unmarshal(d.Body, &event); err != nil {
						log.Println(err)
						d.Nack(false, false)
						continue
					}

					err := c.service.UpdateOrder(ctx, event.OrderID, event.Status)
					if err != nil {
						log.Printf("UpdateStatus error: %s", err)

						if err := rabbitmq.HandleRetry(ch, &d); err != nil {
							log.Printf("Retry error: %s", err)
						}

						d.Nack(false, false)
						continue
					}

					log.Printf("Order status updated for order %s", event.OrderID)
				}
				d.Ack(false)
			}

		}
	}()

	<-forever
}
