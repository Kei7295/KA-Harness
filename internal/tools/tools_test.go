package tools

import (
	"os"
	"testing"
)

func TestEnvelopeSuccess(t *testing.T) {
	e := Success("text/plain", "hello")
	if err := e.Validate(); err != nil {
		t.Errorf("success envelope invalid: %v", err)
	}
	if e.SchemaVersion != "0.1" {
		t.Errorf("schema_version = %q, want 0.1", e.SchemaVersion)
	}
}

func TestEnvelopeFailure(t *testing.T) {
	e := Failure("FS_NOT_FOUND", "file missing", false)
	if err := e.Validate(); err != nil {
		t.Errorf("failure envelope invalid: %v", err)
	}
}

func TestEnvelopeInvalid(t *testing.T) {
	// ok=true but error present
	e := Envelope{OK: true, Error: &Err{Code: "X"}}
	if err := e.Validate(); err == nil {
		t.Error("expected validation error for ok=true with error")
	}

	// ok=false but data present
	e2 := Envelope{OK: false, Data: &Data{Mime: "text/plain"}}
	if err := e2.Validate(); err == nil {
		t.Error("expected validation error for ok=false with data")
	}
}

func TestReadFileSuccess(t *testing.T) {
	// 在当前测试目录创建临时文件
	tmpFile := "test_tmp.md"
	if err := writeTestFile(tmpFile, "hello harness"); err != nil {
		t.Fatal(err)
	}
	defer removeTestFile(tmpFile)

	e := DefaultRegistry.Execute("read_file", map[string]any{"path": tmpFile})
	if err := e.Validate(); err != nil {
		t.Fatalf("envelope invalid: %v", err)
	}
	if !e.OK {
		t.Errorf("expected ok, got error: %v", e.Error)
	}
}

func writeTestFile(name, content string) error {
	return os.WriteFile(name, []byte(content), 0644)
}

func removeTestFile(name string) {
	os.Remove(name)
}

func TestReadFileNotFound(t *testing.T) {
	e := DefaultRegistry.Execute("read_file", map[string]any{"path": "nonexistent.md"})
	if !e.OK && e.Error.Code != "FS_NOT_FOUND" {
		t.Errorf("expected FS_NOT_FOUND, got %s", e.Error.Code)
	}
}

func TestReadFileOutOfScope(t *testing.T) {
	e := DefaultRegistry.Execute("read_file", map[string]any{"path": "../../../etc/passwd"})
	if !e.OK && e.Error.Code != "FS_PATH_OUT_OF_SCOPE" {
		t.Errorf("expected FS_PATH_OUT_OF_SCOPE, got %s", e.Error.Code)
	}
}

func TestListDir(t *testing.T) {
	e := DefaultRegistry.Execute("list_dir", map[string]any{"path": "."})
	if err := e.Validate(); err != nil {
		t.Fatalf("envelope invalid: %v", err)
	}
	if !e.OK {
		t.Errorf("list_dir failed: %v", e.Error)
	}
}

func TestToolNotFound(t *testing.T) {
	e := DefaultRegistry.Execute("nonexistent", nil)
	if !e.OK && e.Error.Code != "TOOL_NOT_FOUND" {
		t.Errorf("expected TOOL_NOT_FOUND, got %s", e.Error.Code)
	}
}

func TestInvalidArgs(t *testing.T) {
	e := DefaultRegistry.Execute("read_file", map[string]any{"wrong_key": "x"})
	if !e.OK && e.Error.Code != "TOOL_INVALID_ARGS" {
		t.Errorf("expected TOOL_INVALID_ARGS, got %s", e.Error.Code)
	}
}
