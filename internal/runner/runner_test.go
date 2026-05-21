package runner

import (
	"testing"

	"github.com/ibrhr/qdoc/internal/config"
)

func TestRun_UnknownSource(t *testing.T) {
	cfg := config.Config{Provider: "openai", Keys: map[string]string{}}
	result := Run("nonexistent-source", "query", cfg)
	if result.Err == nil {
		t.Fatal("expected error for unknown source")
	}
	_, ok := result.Err.(*ErrUnknownSource)
	if !ok {
		t.Fatalf("expected ErrUnknownSource, got %T: %v", result.Err, result.Err)
	}
}

func TestErrUnknownSource_Error(t *testing.T) {
	e := &ErrUnknownSource{Name: "bogus"}
	msg := e.Error()
	if len(msg) == 0 {
		t.Error("error message should not be empty")
	}
}

func TestStep(t *testing.T) {
	s := Step{Phase: "Testing", Detail: "some detail"}
	if s.Phase != "Testing" {
		t.Errorf("Phase = %q", s.Phase)
	}
	if s.Detail != "some detail" {
		t.Errorf("Detail = %q", s.Detail)
	}
}

func TestResult(t *testing.T) {
	r := &Result{
		Answer: "Hello",
		Source: "go",
		Steps:  []Step{{Phase: "Init", Detail: "start"}},
	}
	if r.Answer != "Hello" {
		t.Errorf("Answer = %q", r.Answer)
	}
	if r.Source != "go" {
		t.Errorf("Source = %q", r.Source)
	}
}

func TestContains(t *testing.T) {
	slice := []string{"a", "b", "c"}
	if !contains(slice, "a") {
		t.Error("should find 'a'")
	}
	if !contains(slice, "c") {
		t.Error("should find 'c'")
	}
	if contains(slice, "d") {
		t.Error("should not find 'd'")
	}
	if contains(nil, "x") {
		t.Error("nil slice should not find anything")
	}
}
