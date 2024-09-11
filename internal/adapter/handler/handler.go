package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/sarthak014/Fast-Bank/internal/core/domain"
	"github.com/sarthak014/Fast-Bank/internal/core/port"
	"golang.org/x/crypto/bcrypt"
)

type ApiHandler struct {
	AccountService     port.AccountService
	TransactionService port.TransactionService
	AuthService        port.AuthService
}

func NewApiHandler(accountService port.AccountService, transactionService port.TransactionService, authService port.AuthService) *ApiHandler {
	return &ApiHandler{
		AuthService:        authService,
		TransactionService: transactionService,
		AccountService:     accountService,
	}
}

func (s *ApiHandler) HandleGetAccount(c echo.Context) error {
	accounts, err := s.AccountService.GetAll()
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, accounts)
}

func (s *ApiHandler) HandleGetAccountById(c echo.Context) error {
	id := c.Param("id")
	claims, _ := c.Get("user").(*domain.JWTClaims)
	if id != strconv.Itoa(claims.Id) {
		return echo.ErrUnauthorized
	}

	acc, err := s.AccountService.GetByAccNo(claims.Id)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, acc)
}

func (s *ApiHandler) HandleCreateAccount(c echo.Context) error {
	accReq := new(domain.CreateAccountReq)
	if err := c.Bind(&accReq); err != nil {
		return err
	}
	acc, err := s.AccountService.Create(accReq)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, acc)
}

func (s *ApiHandler) HandleDeleteAccount(c echo.Context) error {
	id := c.Param("id")
	err := s.AccountService.Delete(id)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "Account deleted successfully"})
}

func (s *ApiHandler) HandleTransfer(c echo.Context) error {
	transferReq := new(domain.TransferReq)
	if err := c.Bind(transferReq); err != nil {
		return err
	}
	toId, er := strconv.Atoi(c.Param("accno"))
	if er != nil {
		return er
	}
	claims, ok := c.Get("user").(*domain.JWTClaims)
	if !ok {
		return echo.ErrUnauthorized
	}

	senderId := claims.Id //claims.Id

	// Create transfer message
	transferMsg := domain.TransferMessage{
		TransferId: uuid.NewString(),
		SenderId:   senderId,
		ToAccount:  toId,
		Amount:     transferReq.Amount,
		Status:     "pending",
		CreatedAt:  time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
	}

	// Publish
	err := s.TransactionService.PublishTransferMessage(transferMsg)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprint("Failed to initiate transfer:", err))
	}

	err = s.TransactionService.AddTransferRecord(&transferMsg)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Error adding transfer")

	}
	return c.JSON(http.StatusAccepted, map[string]string{
		"message":     "Transfer initiated",
		"transfer_id": transferMsg.TransferId,
	})
}

func (s *ApiHandler) HandleLogin(c echo.Context) error {
	payload := new(struct {
		Id       int    `json:"id"`
		Password string `json:"password"`
	})
	if err := c.Bind(payload); err != nil {
		return err
	}
	user, err := s.AccountService.GetByAccNo(payload.Id)
	if err != nil {
		return err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.EPassword), []byte(payload.Password)); err != nil {
		return echo.ErrUnauthorized
	}

	token, err := s.AuthService.Generate(int(user.AcNumber))
	if err != nil {
		return err
	}

	return c.JSON(200, map[string]string{
		"token": token,
	})
}

func (s *ApiHandler) GetTransferStatus(c echo.Context) error {
	_, ok := c.Get("user").(*domain.JWTClaims)
	if !ok {
		return errors.New("failed to get user claims")
	}

	trxid := c.Param("id")

	status, err := s.TransactionService.GetTransferStatus(trxid)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]string{"status": status})
}

func (s *ApiHandler) JwtRoute(c echo.Context) error {
	claims, ok := c.Get("user").(*domain.JWTClaims)
	if !ok {
		return errors.New("failed to get user claims")
	}
	return c.JSON(http.StatusOK, claims)
}

func (s *ApiHandler) GetTrxByAcc(c echo.Context) error {
	claims, ok := c.Get("user").(*domain.JWTClaims)
	if !ok {
		return errors.New("failed to get user claims")
	}
	trxs, err := s.TransactionService.GetByAccNo(claims.Id)
	if err != nil {
		return errors.New("failed to get user transactions")
	}

	return c.JSON(http.StatusOK, trxs)
}
