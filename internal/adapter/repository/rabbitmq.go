package repository

import (
	"fmt"

	"github.com/streadway/amqp"
)

func NewMQConnection(url string) (*amqp.Connection, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %v", err)
	}

	return conn, nil
}
