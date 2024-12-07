package handlers

import (
	"fmt"

	"github.com/burnerlee/compextAI/constants"
)

type ExecuteThreadRequest struct {
	ThreadExecutionParamID string `json:"thread_execution_param_id"`
	// messages to execute the thread with - overrides the thread messages
	Messages                    []*createMessage       `json:"messages"`
	ThreadExecutionSystemPrompt string                 `json:"thread_execution_system_prompt"`
	AppendAssistantResponse     bool                   `json:"append_assistant_response"`
	Metadata                    map[string]interface{} `json:"metadata"`
}

func (r *ExecuteThreadRequest) Validate(threadID string) error {
	if r.ThreadExecutionParamID == "" {
		return fmt.Errorf("thread_execution_param_id is required")
	}

	if threadID == constants.THREAD_IDENTIFIER_FOR_NULL_THREAD && len(r.Messages) == 0 {
		return fmt.Errorf("messages are required, when thread_id is not provided")
	}

	return nil
}

type ExecuteThreadResponse struct {
	ThreadExecutionID string `json:"thread_execution_id"`
}

type ThreadExecutionStatusResponse struct {
	Status string `json:"status"`
}

type RerunThreadExecutionRequest struct {
	ThreadExecutionParamTemplateID string `json:"thread_execution_param_template_id"`
	SystemPrompt                   string `json:"system_prompt"`
	AppendAssistantResponse        bool   `json:"append_assistant_response"`
}

func (r *RerunThreadExecutionRequest) Validate() error {
	if r.ThreadExecutionParamTemplateID == "" {
		return fmt.Errorf("thread_execution_param_template_id is required")
	}
	return nil
}
