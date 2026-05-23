package llm

import "github.com/ibrhr/qdoc/internal/auth"

type StreamDelta struct {
	Content   string
	Reasoning string
	Done      bool
	Err       error
	Retrying  bool
}

type Streamer interface {
	Stream(messages []ChatMessage, ch chan<- StreamDelta)
	ModelName() string
	ProviderName() string
}

type Client interface {
	Streamer
	Send(messages []ChatMessage) (string, error)
}

type Config struct {
	Auth     auth.AuthMethod
	Token    auth.Token
	BaseURL  string
	Model    string
	Provider string
	Headers  map[string]string
}
