package sessionlog

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ibrhr/qdoc/internal/config"
)

type Logger struct {
	f *os.File
}

func New(source, query string) (*Logger, error) {
	dir, err := config.LogDir()
	if err != nil {
		return nil, err
	}

	ts := time.Now().Format("2006-01-02_150405")
	safe := sanitize(query)
	if len(safe) > 40 {
		safe = safe[:40]
	}
	name := fmt.Sprintf("%s_%s.log", ts, safe)
	path := filepath.Join(dir, name)

	f, err := os.Create(path)
	if err != nil {
		return nil, err
	}

	l := &Logger{f: f}
	l.Log("=== QDOC SESSION LOG ===")
	l.Log("Time: %s", time.Now().Format(time.RFC3339))
	l.Log("Source: %s", source)
	l.Log("Query: %s", query)
	l.Log("")
	return l, nil
}

func (l *Logger) Log(format string, args ...interface{}) {
	ts := time.Now().Format("15:04:05.000")
	line := fmt.Sprintf(format, args...)
	fmt.Fprintf(l.f, "[%s] %s\n", ts, line)
}

func (l *Logger) Section(title string) {
	l.Log("")
	l.Log("─── %s ───", title)
	l.Log("")
}

func (l *Logger) Raw(title, content string) {
	l.Section(title)
	l.f.WriteString(content)
	if !strings.HasSuffix(content, "\n") {
		l.f.WriteString("\n")
	}
}

func (l *Logger) Close() error {
	l.Log("")
	l.Log("=== END ===")
	return l.f.Close()
}

func sanitize(s string) string {
	var b strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			b.WriteRune(r)
		} else {
			b.WriteByte('_')
		}
	}
	result := b.String()
	if result == "" {
		return "query"
	}
	return result
}
