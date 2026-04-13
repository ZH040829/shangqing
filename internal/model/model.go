package model

import (
	"encoding/json"
	"time"
)

// User 用户
type User struct {
	ID                int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	Username          string    `json:"username" gorm:"uniqueIndex;size:64;not null"`
	PasswordHash      string    `json:"-" gorm:"size:255;not null"`
	EntropyValue      float64   `json:"entropy_value" gorm:"type:decimal(10,4);default:0"` // 熵减值
	ConsciousnessLevel string   `json:"consciousness_level" gorm:"size:10;default:V0"`
	CreatedAt         time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt         time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

func (User) TableName() string { return "users" }

// Conversation 对话
type Conversation struct {
	ID             string    `json:"id" gorm:"primaryKey;size:64"`
	UserID         int64     `json:"user_id" gorm:"index;not null"`
	Title          string    `json:"title" gorm:"size:255"`
	V5Level        string    `json:"v5_level" gorm:"size:10;default:V2"`
	ComplexityAvg   float64   `json:"complexity_avg" gorm:"type:decimal(5,4);default:0"`
	MessageCount   int       `json:"message_count" gorm:"default:0"`
	LastMessageAt  time.Time `json:"last_message_at" gorm:"index"`
	CreatedAt      time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt      time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

func (Conversation) TableName() string { return "conversations" }

// Message 消息
type Message struct {
	ID             int64           `json:"id" gorm:"primaryKey;autoIncrement"`
	ConversationID string           `json:"conversation_id" gorm:"index;size:64;not null"`
	Role           string           `json:"role" gorm:"size:20;not null"` // user / assistant / system
	Content        string           `json:"content" gorm:"type:text;not null"`
	Thinking       string           `json:"thinking" gorm:"type:text"`
	Features       json.RawMessage  `json:"features" gorm:"type:json"`
	TokensUsed     int              `json:"tokens_used" gorm:"default:0"`
	CreatedAt      time.Time        `json:"created_at" gorm:"autoCreateTime"`
}

func (Message) TableName() string { return "messages" }

// ConsciousnessEvent 意识事件
type ConsciousnessEvent struct {
	ID             int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID         int64     `json:"user_id" gorm:"index;not null"`
	ConversationID string    `json:"conversation_id" gorm:"index;size:64"`
	FromLevel      string    `json:"from_level" gorm:"size:10"`
	ToLevel        string    `json:"to_level" gorm:"size:10;not null"`
	TriggerText    string    `json:"trigger_text" gorm:"type:text"`
	CreatedAt      time.Time `json:"created_at" gorm:"autoCreateTime"`
}

func (ConsciousnessEvent) TableName() string { return "consciousness_events" }


