package orchestrator

import (
	"errors"
	"testing"

	"aiharness/internal/model"
)

func TestOrchestratorHappyPath(t *testing.T) {
	m := model.NewStubModel()
	o := New(m)

	code := o.Run("test task")
	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}
}

func TestOrchestratorMaxIterations(t *testing.T) {
	m := &infiniteToolModel{}
	o := New(m)

	code := o.Run("infinite task")
	if code != 70 {
		t.Errorf("expected exit code 70 (EX_SOFTWARE) for max iterations, got %d", code)
	}
}

func TestOrchestratorModelError(t *testing.T) {
	m := &errorModel{}
	o := New(m)

	code := o.Run("error task")
	if code != 69 {
		t.Errorf("expected exit code 69 (EX_UNAVAILABLE) for model error, got %d", code)
	}
}

func TestOrchestratorEmptyTask(t *testing.T) {
	m := model.NewStubModel()
	o := New(m)

	code := o.Run("")
	if code == 0 {
		t.Error("empty task should not return exit code 0")
	}
}

func TestToolCallDisplayEmptyArgs(t *testing.T) {
	tc := &model.ToolCall{Name: "read_file", Arguments: map[string]any{}}
	got := tc.Display()
	want := "read_file()"
	if got != want {
		t.Errorf("Display() with empty args = %q, want %q", got, want)
	}
}

type infiniteToolModel struct{}

func (m *infiniteToolModel) Complete(task string, history []string) (*model.Response, error) {
	return &model.Response{
		ToolCall: &model.ToolCall{Name: "read_file", Arguments: map[string]any{"path": "x"}},
		Done:     false,
	}, nil
}

type errorModel struct{}

func (m *errorModel) Complete(task string, history []string) (*model.Response, error) {
	return nil, errors.New("model service down")
}

