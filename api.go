package main

import (
	"log"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

type ApiServer struct {
	listenAddr string
	store      Storage
}

func (s *ApiServer) Run() {
	e := echo.New()
	e.GET("/account", s.handleGetAccount)
	e.POST("/account", s.handleCreateAccount)
	e.GET("/account/:id", s.handleGetAccountById)
	e.DELETE("/account/:id", s.handleDeleteAccount)
	e.POST("/transfer/:AccNo",s.handleTransfer)
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
	acc := NewAccount(accReq.Fname, accReq.Lname)
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

	return c.JSON(http.StatusOK, transferReq)
}
