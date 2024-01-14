build: 
	go build -o bin/fastbank

run: build
	./bin/fastbank

test:
	go test -v ./...


