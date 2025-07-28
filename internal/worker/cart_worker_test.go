package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/rabbitmq/amqp091-go"
	"github.com/rizaldiabyannata/kioskita-go/internal/core"
	"github.com/rizaldiabyannata/kioskita-go/internal/mq"
	"github.com/rizaldiabyannata/kioskita-go/internal/store"
	"github.com/stretchr/testify/assert"
)

var db *sqlx.DB

func TestMain(m *testing.M) {
	// Manually set env vars for testing
	os.Setenv("DB_HOST", "127.0.0.1")
	os.Setenv("DB_PORT", "5432")
	os.Setenv("DB_USER", "postgres")
	os.Setenv("DB_PASSWORD", "password")
	os.Setenv("DB_NAME", "kioskita")
	os.Setenv("DB_SSL_MODE", "disable")
	os.Setenv("RABBITMQ_USER", "guest")
	os.Setenv("RABBITMQ_PASS", "guest")
	os.Setenv("RABBITMQ_HOST", "localhost")
	os.Setenv("RABBITMQ_PORT", "5672")

	log.Printf("DB_HOST: %s", os.Getenv("DB_HOST"))
	log.Printf("DB_PORT: %s", os.Getenv("DB_PORT"))

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_SSL_MODE"),
	)
	var err error
	db, err = sqlx.Connect("pgx", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	mq.InitRabbitMQ()
	defer mq.CloseRabbitMQ()

	os.Exit(m.Run())
}

func TestCartUpdateWorker(t *testing.T) {
	// 1. Setup: Create a user, product, and cart
	userStore := store.NewUserStore(db)
	productStore := store.NewProductStore(db)
	orderStore := store.NewOrderStore(db)

	user := &core.User{
		ID:           uuid.New(),
		Email:        "test@example.com",
		PasswordHash: "password", // This will be hashed in the real implementation
	}
	err := userStore.Create(user)
	assert.NoError(t, err)

	product := &core.Product{
		ID:    uuid.New(),
		Name:  "Test Product",
		Price: 100,
		Stock: 10,
	}
	err = productStore.Create(product)
	assert.NoError(t, err)

	order := &core.Order{
		ID:     uuid.New(),
		UserID: user.ID,
		Status: core.StatusCart,
	}
	err = orderStore.CreateOrder(order)
	assert.NoError(t, err)

	orderItem := &core.OrderItem{
		ID:        uuid.New(),
		OrderID:   order.ID,
		ProductID: product.ID,
		Quantity:  1,
	}
	err = orderStore.AddOrderItem(orderItem)
	assert.NoError(t, err)

	// 2. Start the worker in a goroutine
	go StartCartUpdateWorker(db)

	// 3. Publish a message to the queue
	ch, err := mq.RabbitMQ.Channel()
	assert.NoError(t, err)
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"cart_updates", // name
		true,           // durable
		false,          // delete when unused
		false,          // exclusive
		false,          // no-wait
		nil,            // arguments
	)
	assert.NoError(t, err)

	body, err := json.Marshal(map[string]string{"user_id": user.ID.String()})
	assert.NoError(t, err)

	err = ch.PublishWithContext(context.Background(),
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp091.Publishing{
			ContentType: "application/json",
			Body:        body,
		})
	assert.NoError(t, err)

	// 4. Give the worker time to process the message
	time.Sleep(2 * time.Second)

	// 5. Assert: Check if the cart total has been updated
	cartView, err := orderStore.GetCartViewByUserID(user.ID)
	assert.NoError(t, err)
	assert.NotNil(t, cartView)
	assert.Equal(t, int64(100), cartView.Order.TotalAmount)
}
