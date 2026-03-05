package consumer

import (
	"context"
	"encoding/json"
	"gateway/rabbitmq"
	"log"
	"orders/internal/domain/models"
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

				var order models.Order
				err := json.Unmarshal(d.Body, &order)
				if err != nil {
					log.Println(err)
					d.Nack(false, false)
					continue
				}

				err = c.service.UpdateOrder(ctx, order.Id, order.Status)
				if err != nil {
					log.Printf("Error creating payment link: %s", err)

					if err := rabbitmq.HandleRetry(ch, &d); err != nil {
						log.Printf("Error handling retry: %s", err)
					}

					d.Nack(false, false)
					continue
				}

				log.Printf("Order updated")
				d.Ack(false)
			}

		}
	}()

	<-forever
}
