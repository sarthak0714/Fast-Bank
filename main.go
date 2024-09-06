package main

import (
	"fmt"
	"log"
)

func main() {
	store, err := NewPGStore()

	if err != nil {
		log.Fatal("3", err)
	}

	if err := store.Init(); err != nil {
		log.Fatal("2", err)
	}
	accService := NewAccountService(store)

	authService := NewAuthService("SHHHHHH")

	trxService, err := NewTransactionService(store, "amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatal("1", err)
	}

	fmt.Println(colorGreen,
		`________  ________  ________   ___  ___      
|\   __  \|\   __  \|\   ___  \|\  \|\  \     
\ \  \|\ /\ \  \|\  \ \  \\ \  \ \  \/  /|_   
 \ \   __  \ \   __  \ \  \\ \  \ \   ___  \  
  \ \  \|\  \ \  \ \  \ \  \\ \  \ \  \\ \  \ 
   \ \_______\ \__\ \__\ \__\\ \__\ \__\\ \__\
    \|_______|\|__|\|__|\|__| \|__|\|__| \|__|
                                              `, colorReset)
	server := NewApiServer(":8080", store, *accService, *trxService, *authService)
	server.Run()
}
