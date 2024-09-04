package main

import (
	"fmt"
	"log"
)

func main() {
	store, err := NewPGStore()

	if err != nil {
		log.Fatal(err)
	}

	if err := store.Init(); err != nil {
		log.Fatal(err)
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
	server := NewApiServer(":8080", store)
	server.Run()
}
