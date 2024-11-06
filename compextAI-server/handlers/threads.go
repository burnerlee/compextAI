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

func (s *Server) ListThreads(w http.ResponseWriter, r *http.Request) {
	userID, err := utils.GetUserIDFromRequest(r)
	if err != nil {
		responses.Error(w, http.StatusUnauthorized, err.Error())
		return
	}

	// find all the threads from the db
	threads, err := models.GetAllThreads(s.DB, uint(userID))
	if err != nil {
		responses.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	responses.JSON(w, http.StatusOK, threads)
}

func (s *Server) CreateThread(w http.ResponseWriter, r *http.Request) {
	userID, err := utils.GetUserIDFromRequest(r)
	if err != nil {
		responses.Error(w, http.StatusUnauthorized, err.Error())
		return
	}

	var request CreateThreadRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		responses.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := request.Validate(); err != nil {
		responses.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	threadCreated, err := controllers.CreateThread(s.DB, &controllers.CreateThreadRequest{
		UserID:   uint(userID),
		Title:    request.Title,
		Metadata: request.Metadata,
	})
	if err != nil {
		responses.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	responses.JSON(w, http.StatusCreated, threadCreated)
}

func (s *Server) GetThread(w http.ResponseWriter, r *http.Request) {
	userID, err := utils.GetUserIDFromRequest(r)
	if err != nil {
		responses.Error(w, http.StatusUnauthorized, err.Error())
		return
	}

	threadID := mux.Vars(r)["id"]

	if threadID == "" {
		responses.Error(w, http.StatusBadRequest, "id parameter is required")
		return
	}

	thread, err := models.GetThread(s.DB, threadID)
	if err != nil {
		responses.Error(w, http.StatusNotFound, err.Error())
		return
	}

	if thread.UserID != uint(userID) {
		responses.Error(w, http.StatusForbidden, "You are not authorized to access this thread")
		return
	}

	responses.JSON(w, http.StatusOK, thread)
}

func (s *Server) UpdateThread(w http.ResponseWriter, r *http.Request) {
	userID, err := utils.GetUserIDFromRequest(r)
	if err != nil {
		responses.Error(w, http.StatusUnauthorized, err.Error())
		return
	}

	threadID := mux.Vars(r)["id"]

	if threadID == "" {
		responses.Error(w, http.StatusBadRequest, "id parameter is required")
		return
	}

	var request UpdateThreadRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		responses.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	thread, err := models.GetThread(s.DB, threadID)
	if err != nil {
		responses.Error(w, http.StatusNotFound, err.Error())
		return
	}

	if thread.UserID != uint(userID) {
		responses.Error(w, http.StatusForbidden, "You are not authorized to update this thread")
		return
	}

	metadataJsonBlob, err := json.Marshal(request.Metadata)
	if err != nil {
		responses.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	newThread := models.Thread{
		Base: models.Base{
			Identifier: threadID,
		},
		Title:    request.Title,
		Metadata: metadataJsonBlob,
	}

	updatedThread, err := models.UpdateThread(s.DB, &newThread)
	if err != nil {
		responses.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	responses.JSON(w, http.StatusOK, updatedThread)
}

func (s *Server) DeleteThread(w http.ResponseWriter, r *http.Request) {
	threadID := mux.Vars(r)["id"]

	if threadID == "" {
		responses.Error(w, http.StatusBadRequest, "id parameter is required")
		return
	}

	userID, err := utils.GetUserIDFromRequest(r)
	if err != nil {
		responses.Error(w, http.StatusUnauthorized, err.Error())
		return
	}

	thread, err := models.GetThread(s.DB, threadID)
	if err != nil {
		responses.Error(w, http.StatusNotFound, err.Error())
		return
	}

	if thread.UserID != uint(userID) {
		responses.Error(w, http.StatusForbidden, "You are not authorized to delete this thread")
		return
	}

	if err := models.DeleteThread(s.DB, threadID); err != nil {
		responses.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	responses.JSON(w, http.StatusNoContent, "Thread deleted successfully")
}