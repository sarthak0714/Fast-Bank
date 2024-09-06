package main

import (
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/crypto/bcrypt"
)

type ApiServer struct {
	listenAddr         string
	store              Storage
	accountService     *AccountService
	transactionService *TransactionService
}

func NewApiServer(addr string, store Storage, accService AccountService, trxService TransactionService) *ApiServer {
	return &ApiServer{
		listenAddr:         addr,
		store:              store,
		accountService:     &accService,
		transactionService: &trxService,
	}
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

	go s.transactionService.ProcessTransfers()

	log.Fatal(e.Start(s.listenAddr))
}

func (s *ApiServer) handleGetAccount(c echo.Context) error {
	accounts, err := s.accountService.GetAccounts()
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, accounts)
}

func (s *ApiServer) handleGetAccountById(c echo.Context) error {
	id := c.Param("id")
	acc, err := s.accountService.GetAccountById(id)
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
	acc, err := s.accountService.CreateAccount(accReq)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, acc)
}

func (s *ApiServer) handleDeleteAccount(c echo.Context) error {
	id := c.Param("id")
	err := s.accountService.DeleteAccount(id)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "Account deleted successfully"})
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
	err := s.transactionService.PublishTransferMessage(transferMsg)
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

func (s *ApiServer) handleLogin(c echo.Context) error {
	payload := new(struct {
		Id       int    `json:"id"`
		Password string `json:"password"`
	})
	if err := c.Bind(payload); err != nil {
		return err
	}

	user, err := s.accountService.store.GetAccountById(payload.Id)
	if err != nil {
		return err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.EPassword), []byte(payload.Password)); err != nil {
		return echo.ErrUnauthorized
	}

	token, err := generateJWT(user.Id)
	if err != nil {
		return err
	}

	return c.JSON(200, map[string]string{
		"token": token,
	})
}

func (s *ApiServer) getTransferStatus(c echo.Context) error {
	_, ok := c.Get("user").(*JWTClaims)
	if !ok {
		return errors.New("failed to get user claims")
	}

	trxid := c.Param("id")

	status, err := s.transactionService.GetTransferStatus(trxid)
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
