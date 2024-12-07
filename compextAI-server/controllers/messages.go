package controllers

import (
	"encoding/json"
	"fmt"

	"github.com/burnerlee/compextAI/models"
	"gorm.io/gorm"
)

func CreateMessages(db *gorm.DB, req *CreateMessageRequest) ([]*models.Message, error) {
	// validate if the thread exists
	if _, err := models.GetThread(db, req.ThreadID); err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("thread not found")
		}
		return nil, fmt.Errorf("failed to get thread: %w", err)
	}

	tx := db.Begin()
	if tx.Error != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}

	var messages []*models.Message
	for _, message := range req.Messages {
		metadataJsonBlob, err := json.Marshal(message.Metadata)
		if err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to marshal metadata: %w", err)
		}

		contentMap := map[string]interface{}{
			"content": message.Content,
		}
		contentJsonBlob, err := json.Marshal(contentMap)
		if err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to marshal content: %w", err)
		}

		message := &models.Message{
			ThreadID:   req.ThreadID,
			ContentMap: contentJsonBlob,
			Role:       message.Role,
			Metadata:   metadataJsonBlob,
		}

		if err := models.CreateMessage(tx, message); err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to create message: %w", err)
		}
		messages = append(messages, message)
	}

	tx.Commit()
	return messages, nil
}
