package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/sarthak014/Fast-Bank/internal/adapter/handler"
	"github.com/sarthak014/Fast-Bank/internal/adapter/repository"
	"github.com/sarthak014/Fast-Bank/internal/config"
	"github.com/sarthak014/Fast-Bank/internal/core/service"
	"github.com/sarthak014/Fast-Bank/pkg/utils"
)

func main() {

	cfg := config.LoadConfig()
	store, err := repository.NewPGStore()
	if err != nil {
		log.Fatal(err)
	}
	conn, err := repository.NewMQConnection(cfg.AmqConnectionStr)
	if err != nil {
		log.Fatal(err)
	}
	authService := service.NewAuthService(cfg.JWTSecret)
	accService := service.NewAccountService(store)
	trxService := service.NewTransactionService(store, conn)

	h := handler.NewApiHandler(accService, trxService, authService)

	e := echo.New()
	e.Use(utils.CustomLogger()) // new
	e.Use(middleware.Recover())

	e.GET("/", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"msg": "works", "time": time.Now().UTC().String()})
	})
	e.GET("/account", h.HandleGetAccount)
	e.POST("/account", h.HandleCreateAccount)
	e.POST("/login", h.HandleLogin)

	jwtGroup := e.Group("")
	jwtGroup.Use(h.AuthService.Middleware)
	jwtGroup.GET("/jwt", h.JwtRoute)
	jwtGroup.GET("/account/:id", h.HandleGetAccountById)
	jwtGroup.DELETE("/account/:id", h.HandleDeleteAccount)
	jwtGroup.POST("/transfer/:accno", h.HandleTransfer)
	jwtGroup.GET("/transfer/:id", h.GetTransferStatus)
	e.HideBanner = true

	go h.TransactionService.ProcessTransfers()
	fmt.Println("\033[32m",
		`________  ________  ________   ___  ___      
|\   __  \|\   __  \|\   ___  \|\  \|\  \     
\ \  \|\ /\ \  \|\  \ \  \\ \  \ \  \/  /|_   
 \ \   __  \ \   __  \ \  \\ \  \ \   ___  \  
  \ \  \|\  \ \  \ \  \ \  \\ \  \ \  \\ \  \ 
   \ \_______\ \__\ \__\ \__\\ \__\ \__\\ \__\
    \|_______|\|__|\|__|\|__| \|__|\|__| \|__|
                                              `, "\033[0m")
	log.Fatal(e.Start(cfg.Port))
}
