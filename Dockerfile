FROM golang:1.22.2-alpine

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go mod tidy

RUN go build -o main ./cmd/main.go

EXPOSE 8080

CMD ["./main"]