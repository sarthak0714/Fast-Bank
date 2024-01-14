package main

import (
	"log"
)

func main() {
	store, err := NewPGStore()
	
	if err != nil {
		log.Fatal(err)
	}

	if err:=store.Init(); err!=nil{
		log.Fatal(err)
	}

	server := NewApiServer(":8080", store)
	server.Run()
}
