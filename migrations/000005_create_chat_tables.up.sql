CREATE TABLE "chat_conversations" (
    "id" uuid PRIMARY KEY,
    "customer_session_id" VARCHAR(255) NOT NULL, -- Bisa juga ID pelanggan jika login
    "created_at" TIMESTAMPTZ NOT NULL DEFAULT (now())
);

CREATE TABLE "chat_messages" (
    "id" uuid PRIMARY KEY,
    "conversation_id" uuid NOT NULL,
    "sender_type" VARCHAR(50) NOT NULL CHECK (
        sender_type IN ('admin', 'customer')
    ),
    "text" TEXT NOT NULL,
    "timestamp" TIMESTAMPTZ NOT NULL DEFAULT (now()),
    CONSTRAINT "fk_messages_conversation" FOREIGN KEY ("conversation_id") REFERENCES "chat_conversations" ("id") ON DELETE CASCADE
);

CREATE INDEX "idx_chat_conversations_customer_session" ON "chat_conversations" ("customer_session_id");

CREATE INDEX "idx_chat_messages_conversation_id" ON "chat_messages" ("conversation_id");