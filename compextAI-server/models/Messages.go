package models

import (
	"encoding/json"
	"fmt"

	"github.com/burnerlee/compextAI/constants"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Message struct {
	Base
	ContentMap json.RawMessage `json:"content_map" gorm:"not null;type:jsonb;default:'{}'"`
	Content    string          `json:"content""`
	Role       string          `json:"role" gorm:"not null"`
	ThreadID   string          `json:"thread_id" gorm:"not null;index"`
	Thread     Thread          `json:"thread" gorm:"foreignKey:ThreadID;references:Identifier"`
	Metadata   json.RawMessage `json:"metadata" gorm:"type:jsonb;default:'{}'"`

	// Implement support for tool calls and function calls later on
	// ToolCalls []ToolCall        `json:"tool_calls"`
	// FunctionCall FunctionCall      `json:"function_call"`
}

func GetAllMessages(db *gorm.DB, threadID string) ([]*Message, error) {
	var messages []*Message
	if err := db.Where("thread_id = ?", threadID).Where("role != ?", "execution").Order("created_at ASC").Find(&messages).Error; err != nil {
		return nil, err
	}
	return messages, nil
}

func GetAllMessagesWithExecution(db *gorm.DB, threadID string) ([]*Message, error) {
	var messages []*Message
	if err := db.Where("thread_id = ?", threadID).Order("created_at ASC").Find(&messages).Error; err != nil {
		return nil, err
	}
	return messages, nil
}

func CreateMessage(db *gorm.DB, message *Message) error {
	// create a new message_id
	messageIDUniqueIdentifier := uuid.New().String()
	messageID := fmt.Sprintf("%s%s", constants.MESSAGE_ID_PREFIX, messageIDUniqueIdentifier)
	message.Identifier = messageID
	return db.Create(message).Error
}

func GetMessage(db *gorm.DB, messageID string) (*Message, error) {
	var message Message
	if err := db.Preload("Thread").First(&message, "identifier = ?", messageID).Error; err != nil {
		return nil, err
	}
	return &message, nil
}

func UpdateMessage(db *gorm.DB, message *Message) (*Message, error) {
	updateData := make(map[string]interface{})
	if message.Role != "" {
		updateData["role"] = message.Role
	}
	if message.Metadata != nil {
		updateData["metadata"] = message.Metadata
	}
	if message.ContentMap != nil {
		updateData["content_map"] = message.ContentMap
	}
	if err := db.Model(&Message{}).Where("identifier = ?", message.Identifier).Updates(updateData).Error; err != nil {
		return nil, err
	}
	return message, nil
}

func DeleteMessage(db *gorm.DB, messageID string) error {
	return db.Where("identifier = ?", messageID).Delete(&Message{}).Error
}
