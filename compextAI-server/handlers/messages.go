package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/burnerlee/compextAI/controllers"
	"github.com/burnerlee/compextAI/models"
	"github.com/burnerlee/compextAI/utils"
	"github.com/burnerlee/compextAI/utils/responses"
	"github.com/gorilla/mux"
)

func (s *Server) ListMessages(w http.ResponseWriter, r *http.Request) {
	threadID := mux.Vars(r)["thread_id"]

	if threadID == "" {
		responses.Error(w, http.StatusBadRequest, "thread_id parameter is required")
		return
	}

	userID, err := utils.GetUserIDFromRequest(r)
	if err != nil {
		responses.Error(w, http.StatusUnauthorized, err.Error())
		return
	}

	hasAccess, err := utils.CheckThreadAccess(s.DB, threadID, uint(userID))
	if err != nil {
		responses.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !hasAccess {
		responses.Error(w, http.StatusForbidden, "You are not authorized to access this thread")
		return
	}

	messages, err := models.GetAllMessages(s.DB, threadID)
	if err != nil {
		responses.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	responses.JSON(w, http.StatusOK, messages)
}

func (s *Server) CreateMessage(w http.ResponseWriter, r *http.Request) {
	threadID := mux.Vars(r)["thread_id"]

	if threadID == "" {
		responses.Error(w, http.StatusBadRequest, "thread_id parameter is required")
		return
	}

	userID, err := utils.GetUserIDFromRequest(r)
	if err != nil {
		responses.Error(w, http.StatusUnauthorized, err.Error())
		return
	}

	hasAccess, err := utils.CheckThreadAccess(s.DB, threadID, uint(userID))
	if err != nil {
		responses.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !hasAccess {
		responses.Error(w, http.StatusForbidden, "You are not authorized to access this thread")
		return
	}

	var message CreateMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&message); err != nil {
		responses.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := message.Validate(); err != nil {
		responses.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	createdMessage, err := controllers.CreateMessage(s.DB, &controllers.CreateMessageRequest{
		ThreadID: threadID,
		Content:  message.Content,
		Role:     message.Role,
		Metadata: message.Metadata,
	})
	if err != nil {
		responses.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	responses.JSON(w, http.StatusCreated, createdMessage)
}

func (s *Server) GetMessage(w http.ResponseWriter, r *http.Request) {
	messageID := mux.Vars(r)["id"]

	if messageID == "" {
		responses.Error(w, http.StatusBadRequest, "id parameter is required")
		return
	}

	userID, err := utils.GetUserIDFromRequest(r)
	if err != nil {
		responses.Error(w, http.StatusUnauthorized, err.Error())
		return
	}

	hasAccess, err := utils.CheckMessageAccess(s.DB, messageID, uint(userID))
	if err != nil {
		responses.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !hasAccess {
		responses.Error(w, http.StatusForbidden, "You are not authorized to access this message")
		return
	}

	message, err := models.GetMessage(s.DB, messageID)
	if err != nil {
		responses.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	responses.JSON(w, http.StatusOK, message)
}

func (s *Server) UpdateMessage(w http.ResponseWriter, r *http.Request) {
	messageID := mux.Vars(r)["id"]

	if messageID == "" {
		responses.Error(w, http.StatusBadRequest, "id parameter is required")
		return
	}

	var message UpdateMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&message); err != nil {
		responses.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := message.Validate(); err != nil {
		responses.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	userID, err := utils.GetUserIDFromRequest(r)
	if err != nil {
		responses.Error(w, http.StatusUnauthorized, err.Error())
		return
	}

	hasAccess, err := utils.CheckMessageAccess(s.DB, messageID, uint(userID))
	if err != nil {
		responses.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !hasAccess {
		responses.Error(w, http.StatusForbidden, "You are not authorized to update this message")
		return
	}

	metadataJsonBlob, err := json.Marshal(message.Metadata)
	if err != nil {
		responses.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	updatedMessage, err := models.UpdateMessage(s.DB, &models.Message{
		Base: models.Base{
			Identifier: messageID,
		},
		Content:  message.Content,
		Role:     message.Role,
		Metadata: metadataJsonBlob,
	})
	if err != nil {
		responses.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	responses.JSON(w, http.StatusOK, updatedMessage)
}

func (s *Server) DeleteMessage(w http.ResponseWriter, r *http.Request) {
	messageID := mux.Vars(r)["id"]

	if messageID == "" {
		responses.Error(w, http.StatusBadRequest, "id parameter is required")
		return
	}

	userID, err := utils.GetUserIDFromRequest(r)
	if err != nil {
		responses.Error(w, http.StatusUnauthorized, err.Error())
		return
	}

	hasAccess, err := utils.CheckMessageAccess(s.DB, messageID, uint(userID))
	if err != nil {
		responses.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !hasAccess {
		responses.Error(w, http.StatusForbidden, "You are not authorized to delete this message")
		return
	}

	if err := models.DeleteMessage(s.DB, messageID); err != nil {
		responses.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	responses.JSON(w, http.StatusNoContent, "message deleted successfully")
}