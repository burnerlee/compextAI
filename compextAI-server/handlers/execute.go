package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/burnerlee/compextAI/constants"
	"github.com/burnerlee/compextAI/controllers"
	"github.com/burnerlee/compextAI/internal/logger"
	"github.com/burnerlee/compextAI/models"
	"github.com/burnerlee/compextAI/utils"
	"github.com/burnerlee/compextAI/utils/responses"
	"github.com/gorilla/mux"
)

func (s *Server) GetThreadExecution(w http.ResponseWriter, r *http.Request) {
	executionID := mux.Vars(r)["id"]

	if executionID == "" {
		responses.Error(w, http.StatusBadRequest, "id parameter is required")
		return
	}

	userID, err := utils.GetUserIDFromRequest(r)
	if err != nil {
		responses.Error(w, http.StatusUnauthorized, err.Error())
		return
	}

	hasAccess, err := utils.CheckThreadExecutionAccess(s.DB, executionID, uint(userID))
	if err != nil {
		responses.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !hasAccess {
		responses.Error(w, http.StatusForbidden, "You are not authorized to access this thread execution")
		return
	}

	threadExecution, err := models.GetThreadExecutionByID(s.DB, executionID)
	if err != nil {
		responses.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	responses.JSON(w, http.StatusOK, threadExecution)
}

func (s *Server) ListThreadExecutions(w http.ResponseWriter, r *http.Request) {
	projectName := mux.Vars(r)["projectname"]
	if projectName == "" {
		responses.Error(w, http.StatusBadRequest, "projectname parameter is required")
		return
	}

	// find the search query and params from the request
	// the following is the type definition of the params
	// export interface ListExecutionsParams {
	// 	page: number;
	// 	limit: number;
	// 	search?: string;
	// 	filters?: Record<string, string>;
	//   }
	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil {
		page = 1
	}
	limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil {
		limit = 10
	}
	searchQuery := r.URL.Query().Get("search")
	searchFilters := r.URL.Query().Get("filters")
	var searchFiltersMap map[string]string

	if searchFilters != "" {
		// parse the searchParams into a map[string]string
		// first url decode the searchParams
		searchFiltersDecoded, err := url.QueryUnescape(searchFilters)
		if err != nil {
			responses.Error(w, http.StatusBadRequest, err.Error())
			return
		}
		if searchFiltersDecoded != "" {
			if err := json.Unmarshal([]byte(searchFiltersDecoded), &searchFiltersMap); err != nil {
				responses.Error(w, http.StatusBadRequest, err.Error())
				return
			}
		}
	}

	logger.GetLogger().Infof("searchQuery: %s, searchFiltersMap: %v, page: %d, limit: %d", searchQuery, searchFiltersMap, page, limit)

	userID, err := utils.GetUserIDFromRequest(r)
	if err != nil {
		responses.Error(w, http.StatusUnauthorized, err.Error())
		return
	}

	projectID, err := utils.GetProjectIDFromName(s.DB, projectName, uint(userID))
	if err != nil {
		responses.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	threadExecutions, total, err := models.GetAllThreadExecutionsByProjectID(s.DB, projectID, searchQuery, searchFiltersMap, page, limit)
	if err != nil {
		responses.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	responses.JSON(w, http.StatusOK, struct {
		Executions []models.ThreadExecution `json:"executions"`
		Total      int                      `json:"total"`
	}{
		Executions: threadExecutions,
		Total:      int(total),
	})
}

func (s *Server) ExecuteThread(w http.ResponseWriter, r *http.Request) {
	threadID := mux.Vars(r)["id"]

	if threadID == "" {
		responses.Error(w, http.StatusBadRequest, "id parameter is required")
		return
	}

	var request ExecuteThreadRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		responses.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := request.Validate(threadID); err != nil {
		responses.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	userID, err := utils.GetUserIDFromRequest(r)
	if err != nil {
		responses.Error(w, http.StatusUnauthorized, err.Error())
		return
	}

	if threadID != constants.THREAD_IDENTIFIER_FOR_NULL_THREAD {
		hasAccess, err := utils.CheckThreadAccess(s.DB, threadID, uint(userID))
		if err != nil {
			responses.Error(w, http.StatusInternalServerError, err.Error())
			return
		}
		if !hasAccess {
			responses.Error(w, http.StatusForbidden, "You are not authorized to execute this thread")
			return
		}
	}

	threadExecutionParam, err := models.GetThreadExecutionParamsByID(s.DB, request.ThreadExecutionParamID)
	if err != nil {
		responses.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	threadExecution, err := controllers.ExecuteThread(s.DB, &controllers.ExecuteThreadRequest{
		UserID:                         uint(userID),
		ThreadID:                       threadID,
		ThreadExecutionParamTemplateID: threadExecutionParam.TemplateID,
		AppendAssistantResponse:        request.AppendAssistantResponse,
		ThreadExecutionSystemPrompt:    request.ThreadExecutionSystemPrompt,
		Messages:                       request.Messages,
		FetchMessagesFromThread:        true,
		ProjectID:                      threadExecutionParam.ProjectID,
	})
	if err != nil {
		responses.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	responses.JSON(w, http.StatusOK, threadExecution)
}

func (s *Server) GetThreadExecutionStatus(w http.ResponseWriter, r *http.Request) {
	executionID := mux.Vars(r)["id"]

	if executionID == "" {
		responses.Error(w, http.StatusBadRequest, "id parameter is required")
		return
	}

	userID, err := utils.GetUserIDFromRequest(r)
	if err != nil {
		responses.Error(w, http.StatusUnauthorized, err.Error())
		return
	}

	hasAccess, err := utils.CheckThreadExecutionAccess(s.DB, executionID, uint(userID))
	if err != nil {
		responses.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !hasAccess {
		responses.Error(w, http.StatusForbidden, "You are not authorized to access this thread execution")
		return
	}

	threadExecution, err := models.GetThreadExecutionByID(s.DB, executionID)
	if err != nil {
		responses.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	responses.JSON(w, http.StatusOK, ThreadExecutionStatusResponse{Status: threadExecution.Status})
}

func (s *Server) GetThreadExecutionResponse(w http.ResponseWriter, r *http.Request) {
	executionID := mux.Vars(r)["id"]

	if executionID == "" {
		responses.Error(w, http.StatusBadRequest, "id parameter is required")
		return
	}

	userID, err := utils.GetUserIDFromRequest(r)
	if err != nil {
		responses.Error(w, http.StatusUnauthorized, err.Error())
		return
	}

	hasAccess, err := utils.CheckThreadExecutionAccess(s.DB, executionID, uint(userID))
	if err != nil {
		responses.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !hasAccess {
		responses.Error(w, http.StatusForbidden, "You are not authorized to access this thread execution")
		return
	}

	threadExecution, err := models.GetThreadExecutionByID(s.DB, executionID)
	if err != nil {
		responses.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	if threadExecution.Status != models.ThreadExecutionStatus_COMPLETED {
		responses.Error(w, http.StatusBadRequest, fmt.Sprintf("Thread execution is: %s", threadExecution.Status))
		return
	}

	var responseContent interface{}
	if err := json.Unmarshal(threadExecution.Output, &responseContent); err != nil {
		responses.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	responses.JSON(w, http.StatusOK, map[string]interface{}{
		"response": responseContent,
		"content":  threadExecution.Content,
		"role":     threadExecution.Role,
	})
}

func (s *Server) RerunThreadExecution(w http.ResponseWriter, r *http.Request) {
	executionID := mux.Vars(r)["id"]

	if executionID == "" {
		responses.Error(w, http.StatusBadRequest, "id parameter is required")
		return
	}

	userID, err := utils.GetUserIDFromRequest(r)
	if err != nil {
		responses.Error(w, http.StatusUnauthorized, err.Error())
		return
	}

	hasAccess, err := utils.CheckThreadExecutionAccess(s.DB, executionID, uint(userID))
	if err != nil {
		responses.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !hasAccess {
		responses.Error(w, http.StatusForbidden, "You are not authorized to rerun this thread execution")
		return
	}

	var request RerunThreadExecutionRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		responses.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := request.Validate(); err != nil {
		responses.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	threadExecution, err := controllers.RerunThreadExecution(s.DB, &controllers.RerunThreadExecutionRequest{
		UserID:                         uint(userID),
		ExecutionID:                    executionID,
		ThreadExecutionParamTemplateID: request.ThreadExecutionParamTemplateID,
		SystemPrompt:                   request.SystemPrompt,
		AppendAssistantResponse:        request.AppendAssistantResponse,
	})
	if err != nil {
		responses.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	responses.JSON(w, http.StatusOK, threadExecution)
}
