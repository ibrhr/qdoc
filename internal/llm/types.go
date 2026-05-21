package llm

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
