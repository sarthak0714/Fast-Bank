package service

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/sarthak014/Fast-Bank/internal/core/domain"
	"github.com/sarthak014/Fast-Bank/internal/core/port"
	"github.com/streadway/amqp"

	"github.com/sarthak014/Fast-Bank/pkg/utils"
)

type transactionService struct {
	store     port.StorageService
	rabbitMQ  *amqp.Connection
	queueName string
}

func NewTransactionService(store port.StorageService, conn *amqp.Connection) port.TransactionService {

	return &transactionService{
		store:     store,
		rabbitMQ:  conn,
		queueName: "transfers",
	}
}

func (s *transactionService) PublishTransferMessage(msg domain.TransferMessage) error {

	// defer s.rabbitMQ.Close() // Removed
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

func (s *transactionService) AddTransferRecord(msg *domain.TransferMessage) error {
	return s.store.AddTransfer(msg)
}

func (s *transactionService) ProcessTransfers() {
	// defer s.rabbitMQ.Close() // Removed to keep the connection open

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
			var transferMsg domain.TransferMessage
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
				utils.TransferLogger(transferMsg.SenderId, transferMsg.ToAccount, transferMsg.Amount)
			}
		}
	}()

	<-forever
}

func (s *transactionService) ExecuteTransfer(msg domain.TransferMessage) error {
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
	// proper trx
	if err := s.store.Transcation(senderAccount, recipientAccount, &msg); err != nil {
		return err
	}

	// update trx log
	err = s.store.UpdateTransferStatus(msg.TransferId, "completed")
	if err != nil {
		return err
	}
	return nil
}

func (s *transactionService) GetTransferStatus(trxid string) (string, error) {
	return s.store.GetTransferStatus(trxid)
}

// Close closes the RabbitMQ connection
func (s *transactionService) Close() error {
	return s.rabbitMQ.Close()
}
