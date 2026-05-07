FROM golang:1.25-alpine

WORKDIR /app

COPY . .

RUN go mod download
RUN go build -o auth-service ./cmd

EXPOSE 9090

CMD ["./auth-service"]