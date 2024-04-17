package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type ApiServer struct {
	listenAddr string
	store      Storage
}

func (s *ApiServer) Run() {
	e := echo.New()
	// e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.GET("/account", s.handleGetAccount)
	e.POST("/account", s.handleCreateAccount)
	e.GET("/account/:id", withJWTRequired(s.handleGetAccountById))
	e.DELETE("/account/:id", s.handleDeleteAccount)
	e.POST("/transfer/:AccNo", s.handleTransfer)
	e.HideBanner = true
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

	user := c.Get("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	id := claims["id"].(int)

	senderAccount, err := s.store.GetAccountById(id)
	if err != nil {
		return err
	}
	if senderAccount.Balance < transferReq.Amount {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Insufficient balance in sender account"})
	}
	remBalance := senderAccount.Balance - transferReq.Amount

	er := s.store.UpdateBalance(id, remBalance)
	if er != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update sender account"})
	}

	return c.JSON(http.StatusOK, transferReq)
}

func withJWTRequired(handler echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		fmt.Println("middleware called")
		return handler(c)
	}

}
