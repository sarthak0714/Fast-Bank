build: 
	@go build -o bin/fastbank

run: build
	@./bin/fastbank

test:
	@go test -v ./...

dbinit:
	@sudo docker run --name some-postgres -e POSTGRES_PASSWORD=jomum -p 5432:5432 -d postgres

