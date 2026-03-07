package consumer

import (
	"context"
	"encoding/json"
	"gateway/rabbitmq"
	"log"
	"payments/internal/domain/models"
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
		"", true, false, true, false, nil)
	if err != nil {
		log.Fatal(err)
	}

	err = ch.QueueBind(
		q.Name,
		"",
		rabbitmq.OrderPaidEvent,
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

				var p models.OrderPaidEvent
				if err := json.Unmarshal(d.Body, &p); err != nil {
					log.Printf("Failed to unmarshal payload: %s", err)
					d.Nack(false, false)
					continue
				}

				orderID := p.OrderID
				amount := p.Amount
				currency := p.Currency

				paymentLink, err := c.service.CreatePayment(ctx, orderID, amount, currency)
				if err != nil {
					log.Printf("Error creating payment link: %s", err)

					if err := rabbitmq.HandleRetry(ch, &d); err != nil {
						log.Printf("Error handling retry: %s", err)
					}

					d.Nack(false, false)
					continue
				}

				log.Printf("Payment link created %s", paymentLink)
				d.Ack(false)

			case <-ctx.Done():
				return
			}
		}
	}()

	<-forever
}
