package llm

type StreamDelta struct {
	Content   string
	Reasoning string
	Done      bool
	Err       error
	Retrying  bool
}