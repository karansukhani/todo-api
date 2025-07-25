package rabbitmq

import (
    "log"
    "os"

    "github.com/streadway/amqp"
)

func ConnectRabbitMQ() *amqp.Connection {
    url := os.Getenv("RABBITMQ_URL")
    conn, err := amqp.Dial(url)
    if err != nil {
        log.Fatalf("Failed to connect to RabbitMQ: %v", err)
    }
    log.Println("Connected to RabbitMQ")
    return conn
}