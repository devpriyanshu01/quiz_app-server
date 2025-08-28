FROM golang:1.24.6-alpine

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod tidy

COPY . .

RUN go build -o main ./cmd/api/main.go

RUN chmod +x main

EXPOSE 3001

CMD ["./main"]

