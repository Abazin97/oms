package consumer

import (
	"gateway/rabbitmq"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Consumer struct {
}

func NewConsumer() *Consumer {
	return &Consumer{}
}

func (c *Consumer) Listen(ch *amqp.Channel) {
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
		for d := range msgs {
			log.Printf("Received a message: %s", d.Body)
			d.Ack(false)
		}
	}()

	<-forever
}
