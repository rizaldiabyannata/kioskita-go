package worker

import (
	"encoding/json"
	"log"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/rizaldiabyannata/kioskita-go/internal/mq"
	"github.com/rizaldiabyannata/kioskita-go/internal/store"
)

func StartCartUpdateWorker(db *sqlx.DB) {
	ch, err := mq.RabbitMQ.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %s", err)
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"cart_updates", // name
		true,           // durable
		false,          // delete when unused
		false,          // exclusive
		false,          // no-wait
		nil,            // arguments
	)
	if err != nil {
		log.Fatalf("Failed to declare a queue: %s", err)
	}

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		log.Fatalf("Failed to register a consumer: %s", err)
	}

	orderStore := store.NewOrderStore(db)

	var forever chan struct{}

	go func() {
		for d := range msgs {
			log.Printf("Received a message: %s", d.Body)
			var data map[string]string
			if err := json.Unmarshal(d.Body, &data); err != nil {
				log.Printf("Error decoding JSON: %s", err)
				continue
			}

			userID, err := uuid.Parse(data["user_id"])
			if err != nil {
				log.Printf("Error parsing UUID: %s", err)
				continue
			}

			_, err = orderStore.GetCartViewByUserID(userID)
			if err != nil {
				log.Printf("Error updating cart view: %s", err)
			}
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}
