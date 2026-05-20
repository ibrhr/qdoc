package llm

import (
	"encoding/json"
	"strings"
)

type ActionType string

const (
	ActionReadFile ActionType = "read_file"
	ActionAnswer   ActionType = "answer"
)

type ParsedAction struct {
	Action  ActionType
	URL     string
	Content string
}

type jsonAction struct {
	Action ActionType `json:"action"`
	URL    string     `json:"url,omitempty"`
}

func ParseResponses(response string) []ParsedAction {
	lines := strings.Split(response, "\n")
	var actions []ParsedAction

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "{") {
			continue
		}

		// Be lenient: strip trailing garbage after the last '}'
		jsonStr := trimmed
		if lastBrace := strings.LastIndex(trimmed, "}"); lastBrace >= 0 && lastBrace < len(trimmed)-1 {
			jsonStr = trimmed[:lastBrace+1]
		}

		var a jsonAction
		if err := json.Unmarshal([]byte(jsonStr), &a); err != nil {
			continue
		}
		switch a.Action {
		case ActionReadFile:
			if a.URL != "" {
				actions = append(actions, ParsedAction{Action: ActionReadFile, URL: a.URL})
			}
		case ActionAnswer:
			answer := strings.Join(lines[i+1:], "\n")
			actions = append(actions, ParsedAction{Action: ActionAnswer, Content: answer})
			return actions
		}
	}

	if len(actions) > 0 {
		return actions
	}

	return []ParsedAction{{Action: ActionAnswer, Content: response}}
}