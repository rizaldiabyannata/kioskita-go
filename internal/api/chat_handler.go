package api

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/rabbitmq/amqp091-go"
	"github.com/rizaldiabyannata/kioskita-go/internal/core"
	"github.com/rizaldiabyannata/kioskita-go/internal/store"
)

type Hub struct {
	clients    map[*websocket.Conn]bool
	broadcast  chan []byte
	register   chan *websocket.Conn
	unregister chan *websocket.Conn
	mutex      sync.Mutex
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*websocket.Conn]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *websocket.Conn),
		unregister: make(chan *websocket.Conn),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mutex.Lock()
			h.clients[client] = true
			h.mutex.Unlock()
		case client := <-h.unregister:
			h.mutex.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				client.Close()
			}
			h.mutex.Unlock()
		case message := <-h.broadcast:
			h.mutex.Lock()
			for client := range h.clients {
				if err := client.WriteMessage(websocket.TextMessage, message); err != nil {
					log.Printf("error: %v", err)
					client.Close()
					delete(h.clients, client)
				}
			}
			h.mutex.Unlock()
		}
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {

		return true
	},
}

type ChatHandler struct {
	chatStore *store.ChatStore
	hub       *Hub
	amqpChan  *amqp091.Channel
}

func NewChatHandler(chatStore *store.ChatStore, amqpChan *amqp091.Channel, hub *Hub) *ChatHandler {
	return &ChatHandler{
		chatStore: chatStore,
		hub:       hub,
		amqpChan:  amqpChan,
	}
}

func (h *ChatHandler) RegisterRoutes(router *gin.RouterGroup) {
	router.GET("/chat", h.handleWebSocket)
	go h.ConsumeMessages()
}

func (h *ChatHandler) ConsumeMessages() {
	err := h.amqpChan.ExchangeDeclare(
		"chat_exchange", // name
		"fanout",        // type
		true,            // durable
		false,           // auto-deleted
		false,           // internal
		false,           // no-wait
		nil,             // arguments
	)
	if err != nil {
		log.Fatalf("Failed to declare an exchange: %s", err)
	}

	q, err := h.amqpChan.QueueDeclare(
		"",    // name
		false, // durable
		false, // delete when unused
		true,  // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		log.Fatalf("Failed to declare a queue: %s", err)
	}

	err = h.amqpChan.QueueBind(
		q.Name,          // queue name
		"",              // routing key
		"chat_exchange", // exchange
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("Failed to bind a queue: %s", err)
	}

	msgs, err := h.amqpChan.Consume(
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

	go func() {
		for d := range msgs {
			var msg core.ChatMessage
			if err := json.Unmarshal(d.Body, &msg); err != nil {
				log.Printf("Error decoding message: %s", err)
				continue
			}
			h.hub.broadcast <- []byte(msg.Text)
		}
	}()
}

func (h *ChatHandler) handleWebSocket(c *gin.Context) {
	sessionID := c.Query("sessionId")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "sessionId is required"})
		return
	}

	conversation, err := h.chatStore.FindOrCreateConversation(sessionID)
	if err != nil {
		log.Printf("Failed to find or create conversation: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to handle conversation"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()
	h.hub.register <- conn

	messages, err := h.chatStore.GetMessagesByConversationID(conversation.ID)
	if err != nil {
		log.Printf("Failed to get messages: %v", err)
	} else {
		for _, msg := range messages {
			if err := conn.WriteJSON(msg); err != nil {
				log.Printf("Failed to send message history: %v", err)
				break
			}
		}
	}

	for {

		var msg core.WebSocketMessage
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Printf("error reading json: %v", err)
			h.hub.unregister <- conn
			break
		}

		dbMessage := &core.ChatMessage{
			ID:             uuid.New(),
			ConversationID: conversation.ID,
			Text:           msg.Text,
			SenderType:     core.SenderCustomer,
		}
		if err := h.chatStore.CreateMessage(dbMessage); err != nil {
			log.Printf("Failed to save message: %v", err)
		}

		// Publish message to RabbitMQ
		body, err := json.Marshal(dbMessage)
		if err != nil {
			log.Printf("Failed to marshal message: %v", err)
		} else {
			err = h.amqpChan.Publish(
				"chat_exchange", // exchange
				"",              // routing key
				false,           // mandatory
				false,           // immediate
				amqp091.Publishing{
					ContentType: "application/json",
					Body:        body,
				})
			if err != nil {
				log.Printf("Failed to publish message: %v", err)
			}
		}
	}
}
