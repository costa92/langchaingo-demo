package mock

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/tools"
)

// MockTranslator 实现模拟翻译器用于测试
type MockTranslator struct {
	CallbacksHandler callbacks.Handler
}

// NewMockTranslator 创建一个新的模拟翻译器
func NewMockTranslator() *MockTranslator {
	return &MockTranslator{}
}

func (m *MockTranslator) Call(ctx context.Context, input string) (string, error) {
	log.Printf("MockTranslator tool called with input: %s", input)

	if m.CallbacksHandler != nil {
		m.CallbacksHandler.HandleToolStart(ctx, input)
	}

	// 模拟翻译结果
	var result string
	switch strings.ToLower(input) {
	case "hello world":
		result = "你好，世界"
	case "good morning":
		result = "早上好"
	case "thank you":
		result = "谢谢"
	default:
		result = fmt.Sprintf("翻译：%s", input)
	}

	log.Printf("Mock translation result: %s", result)

	if m.CallbacksHandler != nil {
		m.CallbacksHandler.HandleToolEnd(ctx, result)
	}

	return result, nil
}

func (m *MockTranslator) Description() string {
	return "A mock tool that translates text between different languages for testing purposes. Input should be the text to translate."
}

func (m *MockTranslator) Name() string {
	return "mock_translator"
}

// 确保 MockTranslator 实现了 tools.Tool 接口
var _ tools.Tool = (*MockTranslator)(nil)

// MockCalculator 实现模拟计算器用于测试
type MockCalculator struct {
	CallbacksHandler callbacks.Handler
}

// NewMockCalculator 创建一个新的模拟计算器
func NewMockCalculator() *MockCalculator {
	return &MockCalculator{}
}

func (m *MockCalculator) Call(ctx context.Context, input string) (string, error) {
	log.Printf("MockCalculator tool called with input: %s", input)

	if m.CallbacksHandler != nil {
		m.CallbacksHandler.HandleToolStart(ctx, input)
	}

	// 模拟计算结果
	var result string
	switch strings.TrimSpace(input) {
	case "2 + 3":
		result = "5"
	case "10 - 4":
		result = "6"
	case "3 * 7":
		result = "21"
	case "15 / 3":
		result = "5"
	default:
		result = fmt.Sprintf("计算结果：%s = ?", input)
	}

	log.Printf("Mock calculation result: %s", result)

	if m.CallbacksHandler != nil {
		m.CallbacksHandler.HandleToolEnd(ctx, result)
	}

	return result, nil
}

func (m *MockCalculator) Description() string {
	return "A mock calculator tool that performs basic arithmetic operations for testing purposes. Input should be a mathematical expression."
}

func (m *MockCalculator) Name() string {
	return "mock_calculator"
}

// 确保 MockCalculator 实现了 tools.Tool 接口
var _ tools.Tool = (*MockCalculator)(nil)

// RunMockTests 运行所有模拟测试
func RunMockTests() {
	ctx := context.Background()

	fmt.Printf("=== Mock Translation Test ===\n")

	// 创建模拟翻译器
	translator := NewMockTranslator()

	// 测试工具调用
	result, err := translator.Call(ctx, "Hello world")
	if err != nil {
		log.Printf("Mock translation failed: %v", err)
		return
	}

	fmt.Printf("Original: Hello world\n")
	fmt.Printf("Translated: %s\n", result)
	fmt.Printf("Tool Name: %s\n", translator.Name())
	fmt.Printf("Tool Description: %s\n", translator.Description())

	log.Printf("Mock translation test completed successfully!\n")

	fmt.Printf("\n=== Mock Agent Test ===\n")

	// 创建模拟工具列表
	tools := []tools.Tool{
		NewMockTranslator(),
		NewMockCalculator(),
	}

	// 测试每个工具
	for _, tool := range tools {
		fmt.Printf("\nTesting tool: %s\n", tool.Name())
		fmt.Printf("Description: %s\n", tool.Description())

		var testInput string
		switch tool.Name() {
		case "mock_translator":
			testInput = "Hello world"
		case "mock_calculator":
			testInput = "2 + 3"
		}

		result, err := tool.Call(ctx, testInput)
		if err != nil {
			log.Printf("Tool %s failed: %v", tool.Name(), err)
			continue
		}

		fmt.Printf("Input: %s\n", testInput)
		fmt.Printf("Output: %s\n", result)
	}

	log.Printf("Mock agent test completed successfully!\n")
}
