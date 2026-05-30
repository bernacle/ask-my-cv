package ask

import "github.com/brunosimplicio/ask-my-cv/internal/domain/message"

type Request struct {
	Question       string `json:"question"`
	ConversationID string `json:"conversation_id,omitempty"`
}

type Response struct {
	Answer         string              `json:"answer"`
	Citations      []message.Citation  `json:"citations"`
	ConversationID string              `json:"conversation_id"`
	SourcesUsed    []message.Citation  `json:"sources_used"`
	TokenUsage     *message.TokenUsage `json:"token_usage,omitempty"`
}
