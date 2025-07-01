# Stage 1: Build binary
FROM golang:1.24-alpine

WORKDIR /app

COPY go.* ./
RUN go mod download

COPY . .

# Build statically linked binary
RUN go build -o main main.go

COPY .env .env

EXPOSE 8080

CMD ["./main"]
