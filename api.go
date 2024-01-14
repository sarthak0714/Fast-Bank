package main

import (
	"log"
	"net/http"

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
	e.GET("/account/:id", s.handleGetAccount)
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
	vars := c.Param("id")
	return c.JSON(http.StatusOK, vars)
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
	return nil
}

func (s *ApiServer) handleTransfer(c echo.Context) error {
	return nil
}
