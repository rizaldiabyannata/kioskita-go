package core

import (
	"time"

	"github.com/google/uuid"
)

type SenderType string

const (
	SenderAdmin    SenderType = "admin"
	SenderCustomer SenderType = "customer"
)

type ChatConversation struct {
	ID                uuid.UUID `db:"id"`
	CustomerSessionID string    `db:"customer_session_id"`
	CreatedAt         time.Time `db:"created_at"`
}

type ChatMessage struct {
	ID             uuid.UUID  `db:"id"`
	ConversationID uuid.UUID  `db:"conversation_id"`
	SenderType     SenderType `db:"sender_type"`
	Text           string     `db:"text"`
	Timestamp      time.Time  `db:"timestamp"`
}

// Payload untuk pesan WebSocket yang masuk
type WebSocketMessage struct {
	Text           string `json:"text"`
	ConversationID string `json:"conversationId"`
}
