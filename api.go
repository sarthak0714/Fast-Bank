package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"encoding/json"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/streadway/amqp"
)

type ApiServer struct {
	listenAddr string
	store      Storage
}

func (s *ApiServer) Run() {
	e := echo.New()
	e.Use(CustomLogger()) // new
	e.Use(middleware.Recover())
	e.GET("/", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"msg": "works", "time": time.Now().UTC().String()})
	})
	e.GET("/account", s.handleGetAccount)
	e.POST("/account", s.handleCreateAccount)
	e.POST("/login", s.handleLogin)

	jwtGroup := e.Group("")
	jwtGroup.Use(JWTMiddleware)
	jwtGroup.GET("/jwt", s.JwtRoute)
	jwtGroup.GET("/account/:id", s.handleGetAccountById)
	jwtGroup.DELETE("/account/:id", s.handleDeleteAccount)
	jwtGroup.POST("/transfer/:accno", s.handleTransfer)
	jwtGroup.GET("/transfer/:id", s.getTransferStatus)
	e.HideBanner = true

	go s.processTransfers()

	log.Fatal(e.Start(s.listenAddr))
}

func NewApiServer(addr string, store Storage) *ApiServer {

	return &ApiServer{
		listenAddr: addr,
		store:      store,
	}
}

func (s *ApiServer) handleGetAccount(c echo.Context) error {
	accoutns, err := s.store.GetAccounts()
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, accoutns)
}

func (s *ApiServer) handleGetAccountById(c echo.Context) error {
	id, er := strconv.Atoi(c.Param("id"))
	if er != nil {
		return er
	}
	acc, err := s.store.GetAccountById(id)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, acc)
}

func (s *ApiServer) handleCreateAccount(c echo.Context) error {
	accReq := new(CreateAccountReq)
	if err := c.Bind(&accReq); err != nil {
		return err
	}
	acc, er := NewAccount(accReq.Fname, accReq.Lname, accReq.Password)
	if er != nil {
		return er
	}
	if err := s.store.CreateAccount(acc); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, acc)
}

func (s *ApiServer) handleDeleteAccount(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]int{"deleted": id})
}

func (s *ApiServer) handleTransfer(c echo.Context) error {
	transferReq := new(TransferReq)
	if err := c.Bind(transferReq); err != nil {
		return err
	}
	toId, er := strconv.Atoi(c.Param("accno"))
	if er != nil {
		return er
	}
	claims, ok := c.Get("user").(*JWTClaims)
	if !ok {
		return echo.ErrUnauthorized
	}

	senderId := claims.Id

	// Create transfer message
	transferMsg := TransferMessage{
		TransferId: uuid.NewString(),
		SenderId:   senderId,
		ToAccount:  toId,
		Amount:     transferReq.Amount,
		Status:     "pending",
	}

	// Publish
	err := s.publishTransferMessage(transferMsg)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to initiate transfer")
	}

	err = s.store.AddTransfer(&transferMsg)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Error adding transfer")

	}
	return c.JSON(http.StatusAccepted, map[string]string{
		"message":     "Transfer initiated",
		"transfer_id": transferMsg.TransferId,
	})
}

func (s *ApiServer) publishTransferMessage(msg TransferMessage) error {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		return err
	}
	defer conn.Close()

	ch, err := conn.Channel()
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
		return err
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

func (s *ApiServer) processTransfers() {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
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

			err = s.executeTransfer(transferMsg)
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

func (s *ApiServer) executeTransfer(msg TransferMessage) error {
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

func (s *ApiServer) getTransferStatus(c echo.Context) error {
	_, ok := c.Get("user").(*JWTClaims)
	if !ok {
		return errors.New("failed to get user claims")
	}

	trxid := c.Param("id")

	status, err := s.store.GetTransferStatus(trxid)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]string{"status": status})

}

func (s *ApiServer) JwtRoute(c echo.Context) error {
	claims, ok := c.Get("user").(*JWTClaims)
	if !ok {
		return errors.New("failed to get user claims")
	}
	return c.JSON(http.StatusOK, claims)
}
