package main

import (
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
)

type ApiServer struct {
	listenAddr string
}

func (s *ApiServer) Run() {
	e := echo.New()
	e.GET("/account", s.handleGetAccount)
	e.GET("/account/:id", s.handleGetAccount)
	e.HideBanner = true
	log.Fatal(e.Start(s.listenAddr))
}

func NewApiServer(addr string) *ApiServer {
	return &ApiServer{
		listenAddr: addr,
	}
}

func (s *ApiServer) handleGetAccount(c echo.Context) error {
	vars := c.Param("id")
	return c.JSON(http.StatusOK, vars)
}

func (s *ApiServer) handleCreateAccount(c echo.Context) error {
	return nil
}

func (s *ApiServer) handleDeleteAccount(c echo.Context) error {
	return nil
}

func (s *ApiServer) handleTransfer(c echo.Context) error {
	return nil
}
