package tui

import (
	"github.com/ibrhr/qdoc/internal/docsource"
	"github.com/ibrhr/qdoc/internal/sessionlog"
)

type docIndexMsg struct {
	source  string
	entries []docsource.Entry
	log     *sessionlog.Logger
}

type docContentMsg struct {
	url     string
	content string
}

type docErrorMsg struct {
	err error
	log *sessionlog.Logger
}

type progressTickMsg struct {
	tick int
}

type queryCompleteMsg struct {
	answer string
	source string
	steps  []queryStep
}

type setupCompleteMsg struct{}