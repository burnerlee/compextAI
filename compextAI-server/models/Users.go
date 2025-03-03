package models

import (
	"encoding/json"

	"gorm.io/gorm"
)

type User struct {
	Base
	Username                  string          `json:"username" gorm:"unique"`
	Email                     string          `json:"email" gorm:"unique"`
	Password                  string          `json:"password" gorm:"not null"`
	APIToken                  string          `json:"api_token" gorm:"unique"`
	OpenAIKey                 string          `json:"openai_key" gorm:"column:openai_key"`
	AnthropicKey              string          `json:"anthropic_key" gorm:"column:anthropic_key"`
	AzureKey                  string          `json:"azure_key" gorm:"column:azure_key"`
	AzureEndpoint             string          `json:"azure_endpoint" gorm:"column:azure_endpoint"`
	GoogleServiceAccountCreds json.RawMessage `json:"google_service_account_creds" gorm:"column:google_service_account_creds; type:jsonb"`
}

func GetUserIDByAPIToken(db *gorm.DB, token string) (uint, error) {
	var user User
	if err := db.Where("api_token = ?", token).First(&user).Error; err != nil {
		return 0, err
	}
	return user.ID, nil
}

func CreateUser(db *gorm.DB, user *User) error {
	return db.Create(user).Error
}

func GetUserByUsername(db *gorm.DB, username string) (*User, error) {
	var user User
	if err := db.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func GetUserByID(db *gorm.DB, id uint) (*User, error) {
	var user User
	if err := db.Where("id = ?", id).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func UpdateUser(db *gorm.DB, user *User) error {
	updateData := make(map[string]interface{})
	if user.AnthropicKey != "" {
		updateData["anthropic_key"] = user.AnthropicKey
	}
	if user.OpenAIKey != "" {
		updateData["openai_key"] = user.OpenAIKey
	}

	return db.Model(user).Updates(updateData).Error
}
