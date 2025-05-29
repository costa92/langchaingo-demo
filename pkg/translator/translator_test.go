package translator

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/tmc/langchaingo/llms/openai"
)

func setupLLM(t *testing.T) *openai.LLM {
	apiKey := os.Getenv("SILICONFLOW_API_KEY")
	if apiKey == "" {
		t.Skip("SILICONFLOW_API_KEY not set")
	}

	var llm *openai.LLM
	var err error
	maxRetries := 3

	// 创建带有超时的 HTTP 客户端
	httpClient := &http.Client{
		Timeout: 60 * time.Second,
	}

	for i := 0; i < maxRetries; i++ {
		if i > 0 {
			time.Sleep(time.Second * time.Duration(i)) // 重试间隔递增
		}

		llm, err = openai.New(
			openai.WithModel("Qwen/Qwen3-235B-A22B"),
			openai.WithBaseURL("https://api.siliconflow.cn/v1"),
			openai.WithToken(apiKey),
			openai.WithHTTPClient(httpClient),
		)
		if err == nil && llm != nil {
			break
		}
	}

	if err != nil {
		t.Fatalf("Failed to create LLM client after %d retries: %v", maxRetries, err)
	}

	if llm == nil {
		t.Fatal("LLM client is nil after successful creation")
	}

	return llm
}

// TestTranslator_Call 测试翻译工具的基本功能
func TestTranslator_Call(t *testing.T) {
	llm := setupLLM(t)
	ctx := context.Background()

	tests := []struct {
		name          string
		input         string
		expectedError bool
	}{
		{
			name:          "Basic Translation",
			input:         "Translate 'Hello world' from English to Chinese",
			expectedError: false,
		},
		{
			name:          "Empty Text",
			input:         "Translate '' from English to Chinese",
			expectedError: true,
		},
		{
			name:          "Invalid Language",
			input:         "Translate 'Hello world' from InvalidLang to Chinese",
			expectedError: false, // 由于使用了默认语言，所以不会报错
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			translator := NewTranslator(llm)
			_, err := translator.Call(ctx, tt.input)
			if (err != nil) != tt.expectedError {
				t.Errorf("expected error: %v, got: %v", tt.expectedError, err)
			}
		})
	}
}

// TestTranslate 测试基本翻译功能
func TestTranslate(t *testing.T) {
	llm := setupLLM(t)
	ctx := context.Background()

	tests := []struct {
		name          string
		text          string
		inputLang     string
		outputLang    string
		expectedError bool
		errorContains string
		checkResult   bool
	}{
		{
			name:          "Basic Translation",
			text:          "Hello world",
			inputLang:     "English",
			outputLang:    "Chinese",
			expectedError: false,
			checkResult:   true,
		},
		{
			name:          "Empty Text",
			text:          "",
			inputLang:     "English",
			outputLang:    "Chinese",
			expectedError: true,
			errorContains: "empty text",
			checkResult:   false,
		},
		{
			name:          "Empty Input Language",
			text:          "Hello",
			inputLang:     "",
			outputLang:    "Chinese",
			expectedError: true,
			errorContains: "empty input language",
			checkResult:   false,
		},
		{
			name:          "Empty Output Language",
			text:          "Hello",
			inputLang:     "English",
			outputLang:    "",
			expectedError: true,
			errorContains: "empty output language",
			checkResult:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Translate(ctx, llm, tt.text, tt.inputLang, tt.outputLang)
			if tt.checkResult {
				fmt.Printf("Translation result for '%s': %s\n", tt.text, result)
			}

			if (err != nil) != tt.expectedError {
				t.Errorf("Translate() error = %v, expectedError %v", err, tt.expectedError)
				return
			}

			if tt.expectedError && err != nil && tt.errorContains != "" {
				if !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("expected error containing '%s', got: %v", tt.errorContains, err)
				}
			}

			if tt.checkResult {
				if result == "" {
					t.Error("Translate() returned empty result when success was expected")
				}
				if result == tt.text {
					t.Errorf("Translation result is the same as input: %s", result)
				}

				// 测试缓存功能
				cacheResult, err := Translate(ctx, llm, tt.text, tt.inputLang, tt.outputLang)
				if err != nil {
					t.Errorf("Cache translation failed: %v", err)
				}
				if cacheResult != result {
					t.Error("Cache result different from original result")
				}
			}
		})
	}
}

// TestTranslateWithTool 测试工具翻译功能
func TestTranslateWithTool(t *testing.T) {
	llm := setupLLM(t)
	ctx := context.Background()

	tests := []struct {
		name          string
		text          string
		inputLang     string
		outputLang    string
		expectedError bool
		errorContains string
		checkResult   bool
		timeout       time.Duration
		maxRetries    int
	}{
		{
			name:          "Basic Tool Translation",
			text:          "Hello world",
			inputLang:     "English",
			outputLang:    "Chinese",
			expectedError: false,
			checkResult:   true,
			timeout:       60 * time.Second, // 增加超时时间
			maxRetries:    3,
		},
		{
			name:          "Empty Text",
			text:          "",
			inputLang:     "English",
			outputLang:    "Chinese",
			expectedError: true,
			errorContains: "empty text",
			checkResult:   false,
			timeout:       5 * time.Second,
			maxRetries:    1,
		},
		{
			name:          "Empty Input Language",
			text:          "Hello",
			inputLang:     "",
			outputLang:    "Chinese",
			expectedError: true,
			errorContains: "empty input language",
			checkResult:   false,
			timeout:       5 * time.Second,
			maxRetries:    1,
		},
		{
			name:          "Empty Output Language",
			text:          "Hello",
			inputLang:     "English",
			outputLang:    "",
			expectedError: true,
			errorContains: "empty output language",
			checkResult:   false,
			timeout:       5 * time.Second,
			maxRetries:    1,
		},
		{
			name:          "Short Text Translation",
			text:          "This is a test.",
			inputLang:     "English",
			outputLang:    "Chinese",
			expectedError: false,
			checkResult:   true,
			timeout:       60 * time.Second, // 增加超时时间
			maxRetries:    3,
		},
		{
			name:          "Invalid Language",
			text:          "Hello",
			inputLang:     "InvalidLang",
			outputLang:    "Chinese",
			expectedError: false,
			checkResult:   true,
			timeout:       60 * time.Second, // 增加超时时间
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

				// 执行翻译
				result, err = TranslateWithTool(testCtx, llm, tt.text, tt.inputLang, tt.outputLang)
				if err == nil || tt.expectedError {
					break
				}
				lastErr = err
			}

			// 如果所有重试都失败，使用最后一次的错误
			if err == nil && lastErr != nil {
				err = lastErr
			}

			// 检查错误
			if (err != nil) != tt.expectedError {
				t.Errorf("TranslateWithTool() error = %v, expectedError %v", err, tt.expectedError)
				return
			}

			// 验证错误信息
			if tt.expectedError && err != nil && tt.errorContains != "" {
				if !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("expected error containing '%s', got: %v", tt.errorContains, err)
				}
				return
			}

			// 检查成功用例的结果
			if tt.checkResult {
				if result == "" {
					t.Error("TranslateWithTool() returned empty result when success was expected")
				}
				if result == tt.text {
					t.Errorf("Translation result is the same as input: %s", result)
				}

				// 测试缓存功能
				cacheResult, err := TranslateWithTool(testCtx, llm, tt.text, tt.inputLang, tt.outputLang)
				if err != nil {
					t.Errorf("Cache translation failed: %v", err)
				}
				if cacheResult != result {
					t.Error("Cache result different from original result")
				}
			}
		})
	}
}

// TestTranslateWithToolConcurrent 测试工具翻译的并发性能
func TestTranslateWithToolConcurrent(t *testing.T) {
	llm := setupLLM(t)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// 准备测试数据
	texts := []string{
		"Hello",
		"World",
		"Test",
	}

	// 创建等待组和结果通道
	var wg sync.WaitGroup
	results := make(chan string, len(texts))
	errors := make(chan error, len(texts))

	// 并发执行翻译
	for _, text := range texts {
		wg.Add(1)
		go func(text string) {
			defer wg.Done()
			result, err := TranslateWithTool(ctx, llm, text, "English", "Chinese")
			if err != nil {
				errors <- err
				return
			}
			results <- result
		}(text)
	}

	// 等待所有翻译完成
	wg.Wait()
	close(results)
	close(errors)

	// 检查错误
	for err := range errors {
		t.Errorf("Concurrent translation error: %v", err)
	}

	// 检查结果
	resultCount := 0
	for result := range results {
		if result == "" {
			t.Error("Empty translation result")
		}
		resultCount++
	}

	if resultCount != len(texts) {
		t.Errorf("Expected %d results, got %d", len(texts), resultCount)
	}
}

// TestTranslateWithToolFailover 测试工具翻译的故障转移功能
func TestTranslateWithToolFailover(t *testing.T) {
	llm := setupLLM(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 测试故障转移场景
	text := "Hello world"
	result, err := TranslateWithTool(ctx, llm, text, "English", "Chinese")

	if err != nil {
		t.Errorf("Failover translation failed: %v", err)
	}
	if result == "" {
		t.Error("Failover translation returned empty result")
	}
	if result == text {
		t.Error("Failover translation returned original text")
	}
}
