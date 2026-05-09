FROM golang:1.25-alpine

WORKDIR /app

COPY . .

RUN go mod download
RUN go build -o payment-service ./cmd

EXPOSE 9091

CMD ["./payment-service"]