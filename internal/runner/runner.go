package runner

import (
	"fmt"
	"strings"

	"github.com/ibrhr/qdoc/internal/config"
	"github.com/ibrhr/qdoc/internal/docsource"
	"github.com/ibrhr/qdoc/internal/llm"
	"github.com/ibrhr/qdoc/internal/provider"
)

const maxIters = 5

type Step struct {
	Phase  string
	Detail string
}

type Result struct {
	Answer string
	Source string
	Steps  []Step
	Err    error
}

func Run(sourceName, query string, cfg config.Config) *Result {
	result := &Result{Source: sourceName}

	source, found := docsource.Find(sourceName)
	if !found {
		result.Err = &ErrUnknownSource{Name: sourceName}
		return result
	}

	entries, err := source.FetchIndex()
	if err != nil {
		result.Err = err
		return result
	}
	result.Steps = append(result.Steps, Step{"Fetched index", fmt.Sprintf("%d pages", len(entries))})

	client, err := provider.ResolveClient(cfg)
	if err != nil {
		result.Err = err
		return result
	}
	result.Steps = append(result.Steps, Step{"Calling", fmt.Sprintf("%s via %s", client.ModelName(), client.ProviderName())})

	systemPrompt := llm.BuildSystemPrompt(source, entries, query)
	messages := []llm.ChatMessage{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: query},
	}

	readFiles := make(map[string]string)
	gotAnswer := false

	for iteration := 0; iteration <= maxIters; iteration++ {
		full, err := streamToCompletion(client, messages)
		if err != nil {
			result.Err = err
			return result
		}

		actions := llm.ParseResponses(full)
		var urls []string

		for _, a := range actions {
			if a.Action == llm.ActionReadFile {
				urls = append(urls, a.URL)
			} else if a.Action == llm.ActionAnswer {
				result.Answer = a.Content
				gotAnswer = true
			}
		}

		if gotAnswer {
			fetchFiles(source, urls, readFiles)
			return result
		}

		if len(urls) == 0 {
			result.Answer = "(The model did not return a READ_FILE or ANSWER command.)"
			return result
		}

		newURLs := fetchFiles(source, urls, readFiles)
		for _, url := range newURLs {
			result.Steps = append(result.Steps, Step{"Reading", url})
		}

		for url, content := range readFiles {
			if !contains(urls, url) {
				continue
			}
			contentMsg := fmt.Sprintf("FILE: %s\n\n%s", url, content)
			messages = append(messages, llm.ChatMessage{
				Role:    "user",
				Content: contentMsg,
			})
		}
	}

	if !gotAnswer {
		result.Answer = "(Answer not found within iteration limit.)"
	}
	return result
}

func streamToCompletion(client llm.Streamer, messages []llm.ChatMessage) (string, error) {
	ch := make(chan llm.StreamDelta, 256)
	go client.Stream(messages, ch)

	var sb strings.Builder
	for d := range ch {
		if d.Err != nil {
			return "", d.Err
		}
		if d.Done {
			break
		}
		sb.WriteString(d.Content)
	}
	return sb.String(), nil
}

func fetchFiles(source docsource.Source, urls []string, cache map[string]string) (newURLs []string) {
	for _, url := range urls {
		if _, ok := cache[url]; ok {
			continue
		}
		content, err := source.FetchContent(url)
		if err != nil {
			continue
		}
		cache[url] = content
		newURLs = append(newURLs, url)
	}
	return newURLs
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

type ErrUnknownSource struct {
	Name string
}

func (e *ErrUnknownSource) Error() string {
	return "qdoc: unknown source \"" + e.Name + "\""
}
