package main

func main() {

	server := NewApiServer(":8080")
	server.Run()
}
