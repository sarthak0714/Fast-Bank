package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/streadway/amqp"
)

type TransactionService struct {
	store     Storage
	rabbitMQ  *amqp.Connection
	queueName string
}

func NewTransactionService(store Storage, rabbitMQURL string) (*TransactionService, error) {
	conn, err := amqp.Dial(rabbitMQURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %v", err)
	}

	return &TransactionService{
		store:     store,
		rabbitMQ:  conn,
		queueName: "transfers",
	}, nil
}

func (s *TransactionService) PublishTransferMessage(msg TransferMessage) error {

	defer s.rabbitMQ.Close()

	ch, err := s.rabbitMQ.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"transfers", // name
		true,        // durable
		false,       // delete when unused
		false,       // exclusive
		false,       // no-wait
		nil,         // arguments
	)
	if err != nil {
		return err
	}

	body, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %v", err)
	}

	err = ch.Publish(
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		})

	return err
}

func (s *TransactionService) ProcessTransfers() {

	defer s.rabbitMQ.Close()

	ch, err := s.rabbitMQ.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %v", err)
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"transfers", // name
		true,        // durable
		false,       // delete when unused
		false,       // exclusive
		false,       // no-wait
		nil,         // arguments
	)
	if err != nil {
		log.Fatalf("Failed to declare a queue: %v", err)
	}

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		log.Fatalf("Failed to register a consumer: %v", err)
	}

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			var transferMsg TransferMessage
			err := json.Unmarshal(d.Body, &transferMsg)
			if err != nil {
				log.Printf("Error decoding message: %v", err)
				continue
			}

			err = s.ExecuteTransfer(transferMsg)
			if err != nil {
				log.Printf("Error processing transfer: %v", err)
			}
			if err == nil {
				transferLogger(transferMsg.SenderId, transferMsg.ToAccount, transferMsg.Amount)
			}
		}
	}()

	<-forever
}

func (s *TransactionService) ExecuteTransfer(msg TransferMessage) error {
	senderAccount, err := s.store.GetAccountById(msg.SenderId)
	if err != nil {
		er := s.store.UpdateTransferStatus(msg.TransferId, "failed")
		if er != nil {
			return er
		}
		return fmt.Errorf("failed to retrieve sender account: %v", err)
	}

	if senderAccount.Balance < msg.Amount {
		er := s.store.UpdateTransferStatus(msg.TransferId, "failed")
		if er != nil {
			return er
		}
		return fmt.Errorf("insufficient balance in sender account")
	}

	recipientAccount, err := s.store.GetAccountById(msg.ToAccount)
	if err != nil {
		er := s.store.UpdateTransferStatus(msg.TransferId, "failed")
		if er != nil {
			return er
		}
		return fmt.Errorf("failed to retrieve recipient account: %v", err)
	}

	senderNewBalance := senderAccount.Balance - msg.Amount
	err = s.store.UpdateBalance(msg.SenderId, senderNewBalance)
	if err != nil {
		er := s.store.UpdateTransferStatus(msg.TransferId, "failed")
		if er != nil {
			return er
		}
		return fmt.Errorf("failed to update sender account: %v", err)
	}

	recipientNewBalance := recipientAccount.Balance + msg.Amount
	err = s.store.UpdateBalance(msg.ToAccount, recipientNewBalance)
	if err != nil {
		// Rollback
		s.store.UpdateBalance(msg.SenderId, senderAccount.Balance)
		return fmt.Errorf("failed to update recipient account: %v", err)
	}
	// update trx log
	err = s.store.UpdateTransferStatus(msg.TransferId, "completed")
	if err != nil {
		return err
	}
	return nil
}

func (s *TransactionService) GetTransferStatus(trxid string) (string, error) {
	return s.store.GetTransferStatus(trxid)
}

// Close closes the RabbitMQ connection
func (s *TransactionService) Close() error {
	return s.rabbitMQ.Close()
}
