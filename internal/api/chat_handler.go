package api

import (
	"log"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/rizaldiabyannata/kioskita-go/internal/core"
	"github.com/rizaldiabyannata/kioskita-go/internal/store"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {

		return true
	},
}

type Hub struct {
	clients    map[*websocket.Conn]bool
	broadcast  chan []byte
	register   chan *websocket.Conn
	unregister chan *websocket.Conn
	mutex      sync.Mutex
}

func newHub() *Hub {
	return &Hub{
		clients:    make(map[*websocket.Conn]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *websocket.Conn),
		unregister: make(chan *websocket.Conn),
	}
}

func (h *Hub) run() {
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

type ChatHandler struct {
	chatStore *store.ChatStore
	hub       *Hub
}

func NewChatHandler(chatStore *store.ChatStore) *ChatHandler {
	hub := newHub()
	go hub.run()
	return &ChatHandler{
		chatStore: chatStore,
		hub:       hub,
	}
}

func (h *ChatHandler) RegisterRoutes(router *gin.RouterGroup) {
	router.GET("/chat", h.handleWebSocket)
}

func (h *ChatHandler) handleWebSocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()
	h.hub.register <- conn

	for {

		var msg core.WebSocketMessage
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Printf("error reading json: %v", err)
			h.hub.unregister <- conn
			break
		}

		convID, _ := uuid.Parse(msg.ConversationID)
		dbMessage := &core.ChatMessage{
			ID:             uuid.New(),
			ConversationID: convID,
			Text:           msg.Text,
			SenderType:     core.SenderCustomer,
		}
		if err := h.chatStore.CreateMessage(dbMessage); err != nil {
			log.Printf("Failed to save message: %v", err)
		}

		h.hub.broadcast <- []byte(msg.Text)
	}
}
