package consumer

import (
	"context"
	"encoding/json"
	"gateway/rabbitmq"
	"log"
	"stock/internal/events"
	"stock/internal/services"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Consumer struct {
	service services.StockService
}

func NewConsumer(service services.StockService) *Consumer {
	return &Consumer{
		service: service,
	}
}

func (c *Consumer) Listen(ctx context.Context, ch *amqp.Channel) {
	q, err := ch.QueueDeclare(
		"orders.stock.queue", true, false, false, false, nil)
	if err != nil {
		log.Fatal(err)
	}

	err = ch.QueueBind(
		q.Name,
		rabbitmq.OrderCreatedEvent,
		rabbitmq.OrderExchange,
		false,
		nil)
	err = ch.QueueBind(
		q.Name,
		rabbitmq.OrderPaidEvent,
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

				case rabbitmq.OrderCreatedEvent:

					var event events.OrderCreatedEvent

					err := json.Unmarshal(d.Body, &event)
					if err != nil {
						log.Println(err)
						d.Nack(false, false)
						continue
					}
					log.Printf("Received a message: %s", d.Body)

					_, err = c.service.Reserve(
						ctx,
						event.LotID,
						event.OrderID,
						event.From,
						event.To,
					)

					if err != nil {
						log.Printf("reserve failed: %s", err)
						d.Ack(false)
						continue
					}

					log.Printf("spot reserved for order %s", event.OrderID)
					d.Ack(false)

				case rabbitmq.OrderPaidEvent:

					var event events.OrderPaidEvent

					err := json.Unmarshal(d.Body, &event)
					if err != nil {
						log.Println(err)
					}
					log.Printf("Received a message: %s", d.Body)

					reservationID, err := c.service.GetReservation(ctx, event.OrderID)
					if err != nil {
						log.Println(err)
					}

					err = c.service.ChangeStatus(ctx, reservationID, "paid")
					if err != nil {
						if err := rabbitmq.HandleRetry(ch, &d); err != nil {
							log.Printf("Retry error: %s", err)
						}
						d.Nack(false, false)
						continue
					}
					log.Printf("Reservation status updated for order %s", event.OrderID)
				}

				//d.Ack(false)
			}
		}
	}()

	<-forever
}
