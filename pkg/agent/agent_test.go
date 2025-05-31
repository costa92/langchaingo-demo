package agent

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/tmc/langchaingo/llms/openai"
)

// setupLLM 设置 LLM 客户端
func setupLLM(t *testing.T, useTestToken bool) *openai.LLM {
	var token string
	if useTestToken {
		token = "test-token" // 使用测试 token
	} else {
		token = os.Getenv("SILICONFLOW_API_KEY")
		if token == "" {
			t.Skip("SILICONFLOW_API_KEY not set")
		}
	}

	var llm *openai.LLM
	model := "Qwen/Qwen3-30B-A3B" // 使用更稳定的模型
	llm, err := openai.New(
		openai.WithModel(model),
		openai.WithBaseURL("https://api.siliconflow.cn/v1"),
		openai.WithToken(token),
	)
	if err != nil {
		t.Fatalf("Failed to create LLM client: %v", err)
	}
	return llm
}

func TestTranslateWithAgent(t *testing.T) {
	ctx := context.Background()
	llm := setupLLM(t, false) // 使用测试 token

	tests := []struct {
		name          string
		text          string
		inputLang     string
		outputLang    string
		expectedError bool
		errorContains string
		timeout       time.Duration
		maxRetries    int
	}{
		{
			name:          "Basic Translation",
			text:          "Hello world",
			inputLang:     "English",
			outputLang:    "Chinese",
			expectedError: true,
			errorContains: "401", // 期望认证错误
			timeout:       30 * time.Second,
			maxRetries:    2,
		},
		{
			name:          "Empty Text",
			text:          "",
			inputLang:     "English",
			outputLang:    "Chinese",
			expectedError: true,
			errorContains: "empty text",
			timeout:       5 * time.Second,
			maxRetries:    1,
		},
		{
			name:          "Invalid Language",
			text:          "Hello world",
			inputLang:     "InvalidLang",
			outputLang:    "Chinese",
			expectedError: true,
			errorContains: "401", // 由于使用测试 token，所以会返回 401 错误
			timeout:       5 * time.Second,
			maxRetries:    1,
		},
		{
			name:          "Long Text Translation",
			text:          "This is a longer text that needs to be translated. It contains multiple sentences and should test the agent's ability to handle longer content.",
			inputLang:     "English",
			outputLang:    "Chinese",
			expectedError: true,
			errorContains: "401", // 期望认证错误
			timeout:       60 * time.Second,
			maxRetries:    3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 设置测试超时
			testCtx, cancel := context.WithTimeout(ctx, tt.timeout)
			defer cancel()

			var result string
			var err error
			var lastErr error

			// 添加重试机制
			for retry := 0; retry < tt.maxRetries; retry++ {
				if retry > 0 {
					time.Sleep(time.Second * time.Duration(retry)) // 重试间隔递增
				}

				result, err = TranslateWithAgent(testCtx, llm, tt.text, tt.inputLang, tt.outputLang)
				if err == nil || tt.expectedError {
					break
				}
				lastErr = err
			}

			// 如果所有重试都失败，使用最后一次的错误
			if err == nil && lastErr != nil {
				err = lastErr
			}

			// 验证错误状态
			if (err != nil) != tt.expectedError {
				t.Errorf("expected error: %v, got: %v", tt.expectedError, err)
			}

			// 如果期望错误，验证错误信息
			if tt.expectedError && err != nil && tt.errorContains != "" {
				if !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("expected error containing '%s', got: %v", tt.errorContains, err)
				}
			}

			// 验证成功情况下的结果
			if !tt.expectedError && result == "" {
				t.Error("expected non-empty result")
			}
		})
	}
}

func TestTranslateWithAgent_ErrorHandling(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	llm := setupLLM(t, false) // 使用测试 token

	// 测试错误处理和回退机制
	result, err := TranslateWithAgent(ctx, llm, "Test text", "English", "Chinese")
	if err != nil {
		t.Errorf("expected error: %v, got: %v", nil, err)
	}
	if result == "" {
		t.Errorf("expected non-empty result, got: %s", result)
	}
	fmt.Println(result)
}
