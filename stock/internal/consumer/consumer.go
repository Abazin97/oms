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

				var order events.OrderCreatedEvent

				err := json.Unmarshal(d.Body, &order)
				if err != nil {
					log.Println(err)
					//d.Nack(false, false)
					//continue
				}
				log.Printf("Received a message: %s", d.Body)

				_, err = c.service.Reserve(
					ctx,
					order.LotID,
					order.OrderID,
					order.From,
					order.To,
				)

				if err != nil {
					log.Printf("reserve failed: %s", err)
					d.Nack(false, true)
					continue
				}

				log.Printf("spot reserved for order %s", order.OrderID)

				d.Ack(false)
			}
		}
	}()

	<-forever
}
