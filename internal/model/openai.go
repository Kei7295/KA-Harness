package model

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// OpenAIModel 是 OpenAI 兼容 API 的客户端（支持 DeepSeek / 通义 / OpenAI 等）。
type OpenAIModel struct {
	BaseURL    string
	APIKey     string
	ModelName  string
	HTTPClient *http.Client
}

type openAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIToolCall struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
}

type openAIRequest struct {
	Model    string           `json:"model"`
	Messages []openAIMessage  `json:"messages"`
	Tools    []map[string]any `json:"tools,omitempty"`
}

type openAIResponse struct {
	Choices []struct {
		Message struct {
			Content   string          `json:"content"`
			ToolCalls []openAIToolCall `json:"tool_calls"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
}

// NewOpenAIModel 从环境变量创建模型客户端。
// 环境变量：KA_API_KEY, KA_BASE_URL, KA_MODEL
func NewOpenAIModel() (*OpenAIModel, error) {
	apiKey := os.Getenv("KA_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("KA_API_KEY 环境变量未设置")
	}

	baseURL := os.Getenv("KA_BASE_URL")
	if baseURL == "" {
		baseURL = "https://api.deepseek.com/v1"
	}

	modelName := os.Getenv("KA_MODEL")
	if modelName == "" {
		modelName = "deepseek-chat"
	}

	return &OpenAIModel{
		BaseURL:   baseURL,
		APIKey:    apiKey,
		ModelName: modelName,
		HTTPClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}, nil
}

func (m *OpenAIModel) Complete(task string, history []string) (*Response, error) {
	messages := []openAIMessage{
		{Role: "system", Content: "你是一个终端 Agent。你可以调用工具来完成用户任务。如果需要调用工具，返回 tool_calls；如果任务完成，直接返回最终答复。"},
		{Role: "user", Content: task},
	}

	for _, h := range history {
		messages = append(messages, openAIMessage{Role: "assistant", Content: h})
	}

	reqBody := openAIRequest{
		Model:    m.ModelName,
		Messages: messages,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	req, err := http.NewRequest("POST", m.BaseURL+"/chat/completions", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+m.APIKey)

	resp, err := m.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("API 请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API 返回错误 %d: %s", resp.StatusCode, string(body))
	}

	var openAIResp openAIResponse
	if err := json.Unmarshal(body, &openAIResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	if len(openAIResp.Choices) == 0 {
		return nil, fmt.Errorf("API 返回空 choices")
	}

	choice := openAIResp.Choices[0]

	// 有 tool_call
	if len(choice.Message.ToolCalls) > 0 {
		tc := choice.Message.ToolCalls[0]
		var args map[string]any
		if err := json.Unmarshal([]byte(tc.Function.Arguments), &args); err != nil {
			args = map[string]any{"raw": tc.Function.Arguments}
		}
		return &Response{
			Text: "",
			ToolCall: &ToolCall{
				Name:      tc.Function.Name,
				Arguments: args,
			},
			Done: false,
		}, nil
	}

	// 最终答复
	return &Response{
		Text:     choice.Message.Content,
		ToolCall: nil,
		Done:     true,
	}, nil
}
