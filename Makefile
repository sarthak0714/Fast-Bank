
run: 
	@go run cmd/main.go

test:
	@go test -v ./...


testdbinit:
	@sudo docker run --name test -e POSTGRES_PASSWORD=test -p 5433:5432 -d  postgres

dbinit:
	@sudo docker run --name postgres -e POSTGRES_PASSWORD=jomum -p 5432:5432 -d postgres

mqinit:
	@sudo docker run -d --name rabbitmq -p 5672:5672 -p 15672:15672 rabbitmq:3-management

docker-build:
	docker build -t fastbank-api .

docker-run:
	docker run -p 8080:8080 --name fastbank-container fastbank-api

dcu:
	docker-compose up -d

dcub:
	docker-compose up --build -d

dcd:
	docker-compose down