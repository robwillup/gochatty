package mq

import "github.com/rabbitmq/amqp091-go"

type RabbitMQ struct {
	Conn    *amqp091.Connection
	Channel *amqp091.Channel
	Queue   *amqp091.Queue
}

func NewRabbitMQ(url, queueName string) (*RabbitMQ, error) {
	conn, err := amqp091.Dial(url)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	q, err := ch.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}

	return &RabbitMQ{
		Conn:    conn,
		Channel: ch,
		Queue:   &q,
	}, nil
}

func (r *RabbitMQ) Close() {
	r.Channel.Close()
	r.Conn.Close()
}

func (r *RabbitMQ) Publish(body []byte) error {
	return r.Channel.Publish(
		"",
		r.Queue.Name,
		false,
		false,
		amqp091.Publishing{
			ContentType: "application/json",
			Body:        body,
		})
}

func (r *RabbitMQ) Consume(autoAck bool) (<-chan amqp091.Delivery, error) {
	msgs, err := r.Channel.Consume(
		r.Queue.Name,
		"",
		autoAck,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}
	return msgs, nil
}
