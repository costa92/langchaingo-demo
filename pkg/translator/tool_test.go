package translator

import (
	"context"
	"testing"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/schema"
)

// mockCallbackHandler 用于测试的回调处理器
type mockCallbackHandler struct {
	startCalled bool
	endCalled   bool
}

func (m *mockCallbackHandler) HandleText(ctx context.Context, text string)                 {}
func (m *mockCallbackHandler) HandleLLMStart(ctx context.Context, prompts []string)        {}
func (m *mockCallbackHandler) HandleLLMEnd(ctx context.Context, output string)             {}
func (m *mockCallbackHandler) HandleChainStart(ctx context.Context, inputs map[string]any) {}
func (m *mockCallbackHandler) HandleChainEnd(ctx context.Context, outputs map[string]any)  {}
func (m *mockCallbackHandler) HandleToolStart(ctx context.Context, input string) {
	m.startCalled = true
}
func (m *mockCallbackHandler) HandleToolEnd(ctx context.Context, output string) {
	m.endCalled = true
}
func (m *mockCallbackHandler) HandleToolError(ctx context.Context, err error)                   {}
func (m *mockCallbackHandler) HandleAgentAction(ctx context.Context, action schema.AgentAction) {}
func (m *mockCallbackHandler) HandleAgentEnd(ctx context.Context, action schema.AgentFinish)    {}
func (m *mockCallbackHandler) HandleAgentFinish(ctx context.Context, finish schema.AgentFinish) {}
func (m *mockCallbackHandler) HandleChainError(ctx context.Context, err error)                  {}
func (m *mockCallbackHandler) HandleLLMError(ctx context.Context, err error)                    {}
func (m *mockCallbackHandler) HandleLLMGenerateContentStart(ctx context.Context, ms []llms.MessageContent) {
}
func (m *mockCallbackHandler) HandleLLMGenerateContentEnd(ctx context.Context, res *llms.ContentResponse) {
}
func (m *mockCallbackHandler) HandleRetrieverStart(ctx context.Context, query string) {}
func (m *mockCallbackHandler) HandleRetrieverEnd(ctx context.Context, query string, documents []schema.Document) {
}
func (m *mockCallbackHandler) HandleStreamingFunc(ctx context.Context, chunk []byte) {}

func TestTranslator_Interface(t *testing.T) {
	translator := NewTranslator(nil)

	// 测试 Name 方法
	if name := translator.Name(); name != "translator" {
		t.Errorf("Name() = %v, want %v", name, "translator")
	}

	// 测试 Description 方法
	if desc := translator.Description(); desc == "" {
		t.Error("Description() returned empty string")
	}
}

func TestNewTranslator(t *testing.T) {
	// 创建 LLM 客户端
	llm := setupLLM(t)

	translator := NewTranslator(llm)

	// 验证翻译器是否正确初始化
	if translator == nil {
		t.Error("NewTranslator() returned nil")
	}

	if translator.LLM != llm {
		t.Error("NewTranslator() did not set LLM correctly")
	}
}
