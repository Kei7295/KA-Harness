// Package tools 是工具系统：注册表、envelope、路径白名单。
//
// 遵循 §10.3 envelope 契约 v0.1：
//   - 所有工具返回 {ok, data, error} 三元组
//   - schema_version 冻结为 "0.1"
package tools

import (
	"encoding/json"
	"fmt"
)

const SchemaVersion = "0.1"

// Envelope 是工具返回的统一包装（§10.3）。
type Envelope struct {
	OK           bool    `json:"ok"`
	Data         *Data   `json:"data"`
	Error        *Err    `json:"error"`
	SchemaVersion string `json:"schema_version"`
}

type Data struct {
	Mime string `json:"mime"`
	Body string `json:"body"`
}

type Err struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	Retryable bool   `json:"retryable"`
}

// Success 构造成功 envelope。
func Success(mime, body string) Envelope {
	return Envelope{
		OK:           true,
		Data:         &Data{Mime: mime, Body: body},
		Error:        nil,
		SchemaVersion: SchemaVersion,
	}
}

// Failure 构造失败 envelope。
func Failure(code, message string, retryable bool) Envelope {
	return Envelope{
		OK:           false,
		Data:         nil,
		Error:        &Err{Code: code, Message: message, Retryable: retryable},
		SchemaVersion: SchemaVersion,
	}
}

// Validate 检查 envelope 是否符合 schema。
func (e Envelope) Validate() error {
	if e.SchemaVersion != SchemaVersion {
		return fmt.Errorf("schema_version mismatch: got %q, want %q", e.SchemaVersion, SchemaVersion)
	}
	if e.OK {
		if e.Data == nil {
			return fmt.Errorf("ok=true but data is nil")
		}
		if e.Error != nil {
			return fmt.Errorf("ok=true but error is not nil")
		}
	} else {
		if e.Error == nil {
			return fmt.Errorf("ok=false but error is nil")
		}
		if e.Data != nil {
			return fmt.Errorf("ok=false but data is not nil")
		}
	}
	return nil
}

// ToJSON 序列化为 JSON 字符串。
func (e Envelope) ToJSON() (string, error) {
	b, err := json.Marshal(e)
	return string(b), err
}
