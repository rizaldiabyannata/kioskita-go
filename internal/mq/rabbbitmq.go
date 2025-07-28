package mq

import (
	"fmt"
	"log"
	"os"

	"github.com/rabbitmq/amqp091-go"
)

var RabbitMQ *amqp091.Connection

func InitRabbitMQ() {
	amqpURL := fmt.Sprintf("amqp://%s:%s@%s:%s/",
		os.Getenv("RABBITMQ_USER"),
		os.Getenv("RABBITMQ_PASS"),
		os.Getenv("RABBITMQ_HOST"),
		os.Getenv("RABBITMQ_PORT"),
	)

	var err error
	RabbitMQ, err = amqp091.Dial(amqpURL)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %s", err)
	}
	log.Println("Successfully connected to RabbitMQ!")
}

func CloseRabbitMQ() {
	if RabbitMQ != nil {
		RabbitMQ.Close()
		log.Println("RabbitMQ connection closed.")
	}
}
