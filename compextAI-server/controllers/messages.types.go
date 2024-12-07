package controllers

type CreateMessageRequest struct {
	ThreadID string           `json:"thread_id"`
	Messages []*CreateMessage `json:"messages"`
}

type CreateMessage struct {
	Content  interface{}            `json:"content"`
	Role     string                 `json:"role"`
	Metadata map[string]interface{} `json:"metadata"`
}
