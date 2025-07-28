package store

import (
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/rizaldiabyannata/kioskita-go/internal/core"
)

type ChatStore struct {
	db *sqlx.DB
}

func NewChatStore(db *sqlx.DB) *ChatStore {
	return &ChatStore{db: db}
}

func (s *ChatStore) FindOrCreateConversation(sessionID string) (*core.ChatConversation, error) {
	var conversation core.ChatConversation
	query := `SELECT * FROM chat_conversations WHERE customer_session_id = $1`
	err := s.db.Get(&conversation, query, sessionID)
	if err == nil {
		return &conversation, nil
	}

	newConversation := &core.ChatConversation{
		ID:                uuid.New(),
		CustomerSessionID: sessionID,
	}
	insertQuery := `INSERT INTO chat_conversations (id, customer_session_id) VALUES ($1, $2) RETURNING created_at`
	err = s.db.QueryRowx(insertQuery, newConversation.ID, newConversation.CustomerSessionID).Scan(&newConversation.CreatedAt)
	if err != nil {
		return nil, err
	}
	return newConversation, nil
}

func (s *ChatStore) CreateMessage(message *core.ChatMessage) error {
	query := `INSERT INTO chat_messages (id, conversation_id, sender_type, text) VALUES ($1, $2, $3, $4)`
	_, err := s.db.Exec(query, message.ID, message.ConversationID, message.SenderType, message.Text)
	return err
}

func (s *ChatStore) GetMessagesByConversationID(convID uuid.UUID) ([]core.ChatMessage, error) {
	var messages []core.ChatMessage
	query := `SELECT * FROM chat_messages WHERE conversation_id = $1 ORDER BY timestamp ASC`
	err := s.db.Select(&messages, query, convID)
	return messages, err
}
