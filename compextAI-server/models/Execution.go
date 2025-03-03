package models

import (
	"encoding/json"
	"fmt"
	"slices"

	"github.com/burnerlee/compextAI/constants"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	ThreadExecutionStatus_IN_PROGRESS = "in_progress"
	ThreadExecutionStatus_COMPLETED   = "completed"
	ThreadExecutionStatus_FAILED      = "failed"
)

type ThreadExecution struct {
	Base
	UserID                          uint                          `json:"user_id"`
	ProjectID                       string                        `json:"project_id" gorm:"index"`
	ThreadID                        string                        `json:"thread_id"`
	Thread                          Thread                        `json:"thread" gorm:"foreignKey:ThreadID;references:Identifier"`
	ThreadExecutionParamsTemplateID string                        `json:"thread_execution_params_template_id"`
	ThreadExecutionParamsTemplate   ThreadExecutionParamsTemplate `json:"thread_execution_params_template" gorm:"foreignKey:ThreadExecutionParamsTemplateID;references:Identifier"`
	Status                          string                        `json:"status"`
	// default value should be {}
	InputMessages             json.RawMessage `json:"input_messages" gorm:"type:jsonb;default:'{}'"`
	Output                    json.RawMessage `json:"output" gorm:"type:jsonb;default:'{}'"`
	Content                   string          `json:"content"`
	Role                      string          `json:"role"`
	ExecutionResponseMetadata json.RawMessage `json:"execution_response_metadata" gorm:"type:jsonb;default:'{}'"`
	ExecutionRequestMetadata  json.RawMessage `json:"execution_request_metadata" gorm:"type:jsonb;default:'{}'"`
	// stores the execution time in seconds
	ExecutionTime uint `json:"execution_time"`
	// metadata is used to store any additional information about the execution
	// this is displayed in the UI and can be used for filtering
	Metadata json.RawMessage `json:"metadata" gorm:"type:jsonb;default:'{}'"`
	Tools    json.RawMessage `json:"tools" gorm:"type:jsonb;default:'{}'"`
}

// ThreadExecutionParams are the parameters for executing a thread
type ThreadExecutionParams struct {
	Base
	UserID      uint                          `json:"user_id"`
	ProjectID   string                        `json:"project_id" gorm:"index"`
	Name        string                        `json:"name"`
	Environment string                        `json:"environment"`
	TemplateID  string                        `json:"template_id"`
	Template    ThreadExecutionParamsTemplate `json:"template" gorm:"foreignKey:TemplateID;references:Identifier"`
}

type ThreadExecutionParamsTemplate struct {
	Base
	UserID              uint            `json:"user_id"`
	ProjectID           string          `json:"project_id" gorm:"index"`
	Name                string          `json:"name"`
	Model               string          `json:"model"`
	Temperature         float64         `json:"temperature"`
	Timeout             int             `json:"timeout"`
	MaxTokens           int             `json:"max_tokens"`
	MaxCompletionTokens int             `json:"max_completion_tokens"`
	TopP                float64         `json:"top_p"`
	MaxOutputTokens     int             `json:"max_output_tokens"`
	ResponseFormat      json.RawMessage `json:"response_format" gorm:"type:jsonb;default:'{}'"`
	SystemPrompt        string          `json:"system_prompt"`
	UseLiteLLM          bool            `json:"use_litellm" gorm:"default:true"`
}

func CreateThreadExecution(db *gorm.DB, threadExecution *ThreadExecution) (*ThreadExecution, error) {
	// create a new thread_execution_id
	threadExecutionIDUniqueIdentifier := uuid.New().String()
	threadExecutionID := fmt.Sprintf("%s%s", constants.THREAD_EXECUTION_ID_PREFIX, threadExecutionIDUniqueIdentifier)
	threadExecution.Identifier = threadExecutionID
	if err := db.Create(threadExecution).Error; err != nil {
		return nil, err
	}
	return threadExecution, nil
}

func UpdateThreadExecution(db *gorm.DB, threadExecution *ThreadExecution) error {
	updateData := make(map[string]interface{})
	if threadExecution.Status != "" {
		updateData["status"] = threadExecution.Status
	}
	if threadExecution.Output != nil {
		updateData["output"] = threadExecution.Output
	}
	if threadExecution.Content != "" {
		updateData["content"] = threadExecution.Content
	}
	if threadExecution.Role != "" {
		updateData["role"] = threadExecution.Role
	}
	if threadExecution.ExecutionResponseMetadata != nil {
		updateData["execution_response_metadata"] = threadExecution.ExecutionResponseMetadata
	}
	if threadExecution.ExecutionRequestMetadata != nil {
		updateData["execution_request_metadata"] = threadExecution.ExecutionRequestMetadata
	}
	if threadExecution.InputMessages != nil {
		updateData["input_messages"] = threadExecution.InputMessages
	}
	if threadExecution.ExecutionTime != 0 {
		updateData["execution_time"] = threadExecution.ExecutionTime
	}
	return db.Model(&ThreadExecution{}).Where("identifier = ?", threadExecution.Identifier).Updates(updateData).Error
}

func GetThreadExecutionByID(db *gorm.DB, executionID string) (*ThreadExecution, error) {
	var threadExecution ThreadExecution
	if err := db.Where("identifier = ?", executionID).Preload("Thread").Preload("ThreadExecutionParamsTemplate").First(&threadExecution).Error; err != nil {
		return nil, err
	}
	return &threadExecution, nil
}

func GetThreadExecutionParamsByID(db *gorm.DB, threadExecutionParamsID string) (*ThreadExecutionParams, error) {
	var threadExecutionParams ThreadExecutionParams
	if err := db.Where("identifier = ?", threadExecutionParamsID).Preload("Template").First(&threadExecutionParams).Error; err != nil {
		return nil, err
	}
	return &threadExecutionParams, nil
}

func GetAllThreadExecutionParams(db *gorm.DB, userID uint, projectID string) ([]ThreadExecutionParams, error) {
	var threadExecutionParams []ThreadExecutionParams
	// preload the template
	// this request is too slow, need to optimize
	if err := db.Where("user_id = ? AND project_id = ?", userID, projectID).Preload("Template").Find(&threadExecutionParams).Error; err != nil {
		return nil, err
	}
	return threadExecutionParams, nil
}

func CreateThreadExecutionParams(db *gorm.DB, threadExecutionParams *ThreadExecutionParams) (*ThreadExecutionParams, error) {
	threadExecutionParamsIDUniqueIdentifier := uuid.New().String()
	threadExecutionParamsID := fmt.Sprintf("%s%s", constants.THREAD_EXECUTION_PARAMS_ID_PREFIX, threadExecutionParamsIDUniqueIdentifier)
	threadExecutionParams.Identifier = threadExecutionParamsID
	if err := db.Create(threadExecutionParams).Error; err != nil {
		return nil, err
	}
	return threadExecutionParams, nil
}

func DeleteThreadExecutionParams(db *gorm.DB, threadExecutionParamsID string) error {
	return db.Delete(&ThreadExecutionParams{}, "identifier = ?", threadExecutionParamsID).Error
}

func GetThreadExecutionParamsByUserIDAndNameAndEnvironment(db *gorm.DB, userID uint, name, environment, projectID string) (*ThreadExecutionParams, error) {
	var threadExecutionParams ThreadExecutionParams
	if err := db.Where("user_id = ? AND name = ? AND environment = ? AND project_id = ?", userID, name, environment, projectID).Preload("Template").First(&threadExecutionParams).Error; err != nil {
		return nil, err
	}
	return &threadExecutionParams, nil
}

func CreateThreadExecutionParamsTemplate(db *gorm.DB, threadExecutionParamsTemplate *ThreadExecutionParamsTemplate) (*ThreadExecutionParamsTemplate, error) {
	threadExecutionParamsTemplateIDUniqueIdentifier := uuid.New().String()
	threadExecutionParamsTemplateID := fmt.Sprintf("%s%s", constants.THREAD_EXECUTION_PARAMS_TEMPLATE_ID_PREFIX, threadExecutionParamsTemplateIDUniqueIdentifier)
	threadExecutionParamsTemplate.Identifier = threadExecutionParamsTemplateID
	if err := db.Create(threadExecutionParamsTemplate).Error; err != nil {
		return nil, err
	}
	return threadExecutionParamsTemplate, nil
}

func DeleteThreadExecutionParamsTemplate(db *gorm.DB, threadExecutionParamsTemplateID string) error {
	return db.Delete(&ThreadExecutionParamsTemplate{}, "identifier = ?", threadExecutionParamsTemplateID).Error
}

func GetThreadExecutionParamsTemplateByID(db *gorm.DB, threadExecutionParamsTemplateID string) (*ThreadExecutionParamsTemplate, error) {
	var threadExecutionParamsTemplate ThreadExecutionParamsTemplate
	if err := db.Where("identifier = ?", threadExecutionParamsTemplateID).First(&threadExecutionParamsTemplate).Error; err != nil {
		return nil, err
	}
	return &threadExecutionParamsTemplate, nil
}

func UpdateThreadExecutionParamsTemplate(db *gorm.DB, threadExecutionParamsTemplate *ThreadExecutionParamsTemplate) error {
	updateData := make(map[string]interface{})
	if threadExecutionParamsTemplate.Name != "" {
		updateData["name"] = threadExecutionParamsTemplate.Name
	}
	if threadExecutionParamsTemplate.Model != "" {
		updateData["model"] = threadExecutionParamsTemplate.Model
	}
	if threadExecutionParamsTemplate.Temperature != 0 {
		updateData["temperature"] = threadExecutionParamsTemplate.Temperature
	}
	if threadExecutionParamsTemplate.Timeout != 0 {
		updateData["timeout"] = threadExecutionParamsTemplate.Timeout
	}
	if threadExecutionParamsTemplate.MaxTokens != 0 {
		updateData["max_tokens"] = threadExecutionParamsTemplate.MaxTokens
	}
	if threadExecutionParamsTemplate.MaxCompletionTokens != 0 {
		updateData["max_completion_tokens"] = threadExecutionParamsTemplate.MaxCompletionTokens
	}
	if threadExecutionParamsTemplate.TopP != 0 {
		updateData["top_p"] = threadExecutionParamsTemplate.TopP
	}
	if threadExecutionParamsTemplate.MaxOutputTokens != 0 {
		updateData["max_output_tokens"] = threadExecutionParamsTemplate.MaxOutputTokens
	}
	if threadExecutionParamsTemplate.ResponseFormat != nil {
		updateData["response_format"] = threadExecutionParamsTemplate.ResponseFormat
	}
	if threadExecutionParamsTemplate.SystemPrompt != "" {
		updateData["system_prompt"] = threadExecutionParamsTemplate.SystemPrompt
	}

	return db.Model(&ThreadExecutionParamsTemplate{}).Where("identifier = ?", threadExecutionParamsTemplate.Identifier).Updates(updateData).Error
}

func GetAllThreadExecutionParamsTemplates(db *gorm.DB, userID uint, projectID string) ([]ThreadExecutionParamsTemplate, error) {
	var threadExecutionParamsTemplates []ThreadExecutionParamsTemplate
	if err := db.Where("user_id = ? AND project_id = ?", userID, projectID).Find(&threadExecutionParamsTemplates).Error; err != nil {
		return nil, err
	}
	return threadExecutionParamsTemplates, nil
}

func GetThreadExecutionParamsByTemplateID(db *gorm.DB, templateID string) ([]ThreadExecutionParams, error) {
	var threadExecutionParams []ThreadExecutionParams
	if err := db.Where("template_id = ?", templateID).Find(&threadExecutionParams).Error; err != nil {
		return nil, err
	}
	return threadExecutionParams, nil
}

func UpdateThreadExecutionParamsTemplateID(db *gorm.DB, threadExecutionParamsID, templateID string) error {
	return db.Model(&ThreadExecutionParams{}).Where("identifier = ?", threadExecutionParamsID).Update("template_id", templateID).Error
}

func GetAllThreadExecutionsByProjectID(db *gorm.DB, projectID string, searchQuery string, searchParamsMap map[string]string, page, limit int) ([]ThreadExecution, int64, error) {

	offset := (page - 1) * limit
	var total int64

	query := db.Model(&ThreadExecution{}).Where("project_id = ?", projectID)

	if searchQuery != "" {
		query = query.Where("identifier LIKE ? OR thread_id LIKE ?", "%"+searchQuery+"%", "%"+searchQuery+"%")
	}

	if len(searchParamsMap) > 0 {
		allowedFilters := []string{"status", "thread_id"}
		for key, value := range searchParamsMap {
			if slices.Contains(allowedFilters, key) {
				query = query.Where(fmt.Sprintf("%s = ?", key), value)
			}
		}
	}

	if len(searchParamsMap) > 0 {
		for key, value := range searchParamsMap {
			query = query.Where("metadata ->> ? = ?", key, value)
		}
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var threadExecutions []ThreadExecution
	if err := query.Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&threadExecutions).Error; err != nil {
		return nil, 0, err
	}

	return threadExecutions, total, nil
}
