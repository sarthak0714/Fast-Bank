package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
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
	e.GET("/jwt", s.JwtRoute, JWTMiddleware)
	e.GET("/account/:id", s.handleGetAccountById, JWTMiddleware)
	e.DELETE("/account/:id", s.handleDeleteAccount, JWTMiddleware)
	e.POST("/transfer/:AccNo", s.handleTransfer, JWTMiddleware)
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

	claims, ok := c.Get("user").(*JWTClaims)
	if !ok {
		return echo.ErrUnauthorized
	}

	senderId := claims.Id

	senderAccount, err := s.store.GetAccountById(senderId)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve sender account")
	}

	if senderAccount.Balance < transferReq.Amount {
		return echo.NewHTTPError(http.StatusBadRequest, "Insufficient balance in sender account")
	}

	recipientAccount, err := s.store.GetAccountById(transferReq.ToAccount)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve recipient account")
	}

	senderNewBalance := senderAccount.Balance - transferReq.Amount
	err = s.store.UpdateBalance(senderId, senderNewBalance)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update sender account")
	}

	recipientNewBalance := recipientAccount.Balance + transferReq.Amount
	err = s.store.UpdateBalance(transferReq.ToAccount, recipientNewBalance)
	if err != nil {
		s.store.UpdateBalance(senderId, senderAccount.Balance)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update recipient account")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message":                "Transfer successful",
		"transferDetails":        transferReq,
		"senderRemainingBalance": senderNewBalance,
		"recipientNewBalance":    recipientNewBalance,
	})
}

func (s *ApiServer) JwtRoute(c echo.Context) error {
	claims, ok := c.Get("user").(*JWTClaims)
	if !ok {
		return errors.New("failed to get user claims")
	}
	return c.JSON(http.StatusOK, claims)
}

const (
	colorRed       = "\033[31m"
	colorGreen     = "\033[32m"
	colorYellow    = "\033[33m"
	colorBlue      = "\033[34m"
	colorPurple    = "\033[35m"
	colorCyan      = "\033[36m"
	colorGray      = "\033[37m"
	colorReset     = "\033[0m"
	colorLightCyan = "\033[96m"
)

func statusColor(code int) string {
	switch {
	case code >= 100 && code < 200:
		return colorYellow
	case code >= 200 && code < 300:
		return colorGreen
	case code >= 300 && code < 400:
		return colorRed
	case code >= 400 && code < 500:
		return colorBlue
	case code >= 500:
		return colorPurple
	default:
		return colorReset
	}
}

func CustomLogger() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()

			err := next(c)
			if err != nil {
				c.Error(err)
			}

			req := c.Request()
			res := c.Response()

			id := req.Header.Get(echo.HeaderXRequestID)
			if id == "" {
				id = res.Header().Get(echo.HeaderXRequestID)
			}

			logMessage := fmt.Sprintf("%s[%s]%s %s%s%s %s%s%s %s%d%s %s%v%s %s",
				colorLightCyan, time.Now().Format("2006-01-02 15:04:05"), colorReset,
				colorGray, req.Method, colorReset,
				colorCyan, req.URL.Path, colorReset,
				statusColor(res.Status), res.Status, colorReset,
				colorGray, time.Since(start), colorReset,
				id,
			)

			fmt.Println(logMessage)

			return nil
		}
	}
}
