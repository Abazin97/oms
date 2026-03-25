package consumer

import (
	"context"
	"encoding/json"
	"gateway/rabbitmq"
	"log"
	"payments/internal/events"
	"payments/internal/services"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Consumer struct {
	service services.PaymentService
}

func NewConsumer(service services.PaymentService) *Consumer {
	return &Consumer{
		service: service,
	}
}

func (c *Consumer) Listen(ctx context.Context, ch *amqp.Channel) {
	q, err := ch.QueueDeclare(
		"payments.order-created.queue", true, false, false, false, nil)
	if err != nil {
		log.Fatal(err)
	}

	err = ch.QueueBind(
		q.Name,
		rabbitmq.StockReservedEvent,
		rabbitmq.OrderExchange,
		false,
		nil)
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

				case rabbitmq.StockReservedEvent:

					var p events.StockReservedEvent
					if err := json.Unmarshal(d.Body, &p); err != nil {
						log.Printf("Failed to unmarshal payload: %s", err)
						d.Nack(false, false)
						continue
					}
					log.Println("event body:", string(d.Body))

					payment, err := c.service.CreatePayment(ctx, p.OrderID, p.Amount, p.Currency)
					if err != nil {
						log.Printf("Error creating payment link: %s", err)

						if err := rabbitmq.HandleRetry(ch, &d); err != nil {
							log.Printf("Error handling retry: %s", err)
						}

						d.Nack(false, true)
					}
					log.Printf("Payment link created for order %s: %s", p.OrderID, payment.Confirmation.ConfirmationURL)

				case rabbitmq.StockReservationFailedEvent:
					log.Println("stock reservation failed:", string(d.Body))
					d.Ack(false)
				}

			case <-ctx.Done():
				return
			}
		}
	}()

	<-forever
}
