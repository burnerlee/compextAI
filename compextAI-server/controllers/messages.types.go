package controllers

type CreateMessageRequest struct {
	ThreadID string           `json:"thread_id"`
	Messages []*CreateMessage `json:"messages"`
}

type CreateMessage struct {
	Content      interface{}            `json:"content"`
	Role         string                 `json:"role"`
	ToolCallID   string                 `json:"tool_call_id"`
	Metadata     map[string]interface{} `json:"metadata"`
	ToolCalls    interface{}            `json:"tool_calls"`
	FunctionCall interface{}            `json:"function_call"`
}
