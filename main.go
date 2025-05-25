// Package main is the main package for the application.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/tmc/langchaingo/agents"
	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/prompts"
	"github.com/tmc/langchaingo/tools"
)

func main() {
	ctx := context.Background()

	// 检查环境变量
	apiURL := os.Getenv("SILICONFLOW_API_URL")
	apiKey := os.Getenv("SILICONFLOW_API_KEY")

	if apiURL == "" {
		apiURL = "https://api.siliconflow.cn/v1"
		log.Printf("SILICONFLOW_API_URL not set, using default: %s", apiURL)
	}

	if apiKey == "" {
		log.Printf("SILICONFLOW_API_KEY not set, using mock translation for testing")
		// 使用模拟翻译进行测试
		testMockTranslation()
		testMockAgent()
		return
	}

	// 使用支持 function calling 的模型
	// 优先级：Qwen2.5 > GPT-4 > GPT-3.5-turbo
	supportedModels := []string{
		"Qwen/Qwen2.5-72B-Instruct",
		"Qwen/Qwen2.5-32B-Instruct",
		"gpt-4o-mini",
		"gpt-3.5-turbo",
	}

	model := supportedModels[0] // 默认使用第一个模型
	log.Printf("Using model: %s", model)

	llm, err := openai.New(
		openai.WithModel(model),
		openai.WithBaseURL(apiURL),
		openai.WithToken(apiKey),
	)
	if err != nil {
		log.Fatalf("Failed to initialize LLM: %v", err)
	}

	// Test the translator
	text := "Hello world"
	// translated, err := translateWithTool(ctx, llm, text, "English", "Chinese")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	translated, err := translateWithAgent(ctx, llm, text, "English", "Chinese")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Original: %s\nTranslated: %s\n", text, translated)
}

// testMockTranslation 测试模拟翻译功能
func testMockTranslation() {
	ctx := context.Background()

	fmt.Printf("=== Mock Translation Test ===\n")

	// 创建模拟翻译器
	translator := MockTranslator{}

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
}

// testMockAgent 测试模拟 agent 功能
func testMockAgent() {
	ctx := context.Background()

	fmt.Printf("\n=== Mock Agent Test ===\n")

	// 创建模拟工具列表
	tools := []tools.Tool{
		MockTranslator{},
		MockCalculator{},
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

// MockTranslator 实现模拟翻译器用于测试
type MockTranslator struct {
	CallbacksHandler callbacks.Handler
}

func (m MockTranslator) Call(ctx context.Context, input string) (string, error) {
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

func (m MockTranslator) Description() string {
	return "A mock tool that translates text between different languages for testing purposes. Input should be the text to translate."
}

func (m MockTranslator) Name() string {
	return "mock_translator"
}

// 确保 MockTranslator 实现了 tools.Tool 接口
var _ tools.Tool = MockTranslator{}

// MockCalculator 实现模拟计算器用于测试
type MockCalculator struct {
	CallbacksHandler callbacks.Handler
}

func (m MockCalculator) Call(ctx context.Context, input string) (string, error) {
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

func (m MockCalculator) Description() string {
	return "A mock calculator tool that performs basic arithmetic operations for testing purposes. Input should be a mathematical expression."
}

func (m MockCalculator) Name() string {
	return "mock_calculator"
}

// 确保 MockCalculator 实现了 tools.Tool 接口
var _ tools.Tool = MockCalculator{}

// getProductName is a function that gets a product name from the LLM
func getProductName(ctx context.Context, llm *openai.LLM, product string) (string, error) {
	prompt := prompts.NewPromptTemplate(
		"What is a good name for a company that makes {{.product}}?",
		[]string{"product"},
	)

	llmChain := chains.NewLLMChain(llm, prompt)

	out, err := chains.Run(ctx, llmChain, "socks")
	if err != nil {
		return "", err
	}
	return out, nil
}

// translate is a function that translates text from one language to another
func translate(ctx context.Context, llm *openai.LLM, text string, inputLanguage string, outputLanguage string) (string, error) {
	prompt := prompts.NewPromptTemplate(
		"Translate the following text from {{.inputLanguage}} to {{.outputLanguage}}. {{.text}}",
		[]string{"inputLanguage", "outputLanguage", "text"},
	)

	llmChain := chains.NewLLMChain(llm, prompt)

	outputValues, err := chains.Call(ctx, llmChain, map[string]any{
		"inputLanguage":  inputLanguage,
		"outputLanguage": outputLanguage,
		"text":           text,
	})
	if err != nil {
		return "", fmt.Errorf("translation failed: %w", err)
	}

	out, ok := outputValues[llmChain.OutputKey].(string)
	if !ok {
		return "", fmt.Errorf("invalid chain return")
	}
	return out, nil
}

// Translator implements the tools.Tool interface for translation tasks
type Translator struct {
	LLM              *openai.LLM
	CallbacksHandler callbacks.Handler
}

// Call implements the actual translation using the LLM
func (t Translator) Call(ctx context.Context, input string) (string, error) {
	log.Printf("Translator tool called with input: %s", input)

	if t.CallbacksHandler != nil {
		t.CallbacksHandler.HandleToolStart(ctx, input)
	}

	// 解析输入以提取翻译信息
	// 支持格式：
	// 1. "Translate 'text' from source to target"
	// 2. "text" (默认从英语翻译到中文)
	var text, sourceLang, targetLang string

	// 简单的输入解析
	if strings.Contains(strings.ToLower(input), "translate") {
		// 尝试解析完整的翻译请求
		parts := strings.Split(input, "'")
		if len(parts) >= 3 {
			text = parts[1]
			// 尝试提取语言信息
			if strings.Contains(strings.ToLower(input), "english") && strings.Contains(strings.ToLower(input), "chinese") {
				sourceLang = "English"
				targetLang = "Chinese"
			} else {
				sourceLang = "English"
				targetLang = "Chinese"
			}
		} else {
			text = input
			sourceLang = "English"
			targetLang = "Chinese"
		}
	} else {
		// 直接翻译输入文本
		text = input
		sourceLang = "English"
		targetLang = "Chinese"
	}

	log.Printf("Translating '%s' from %s to %s", text, sourceLang, targetLang)

	// 使用内置的 translate 函数进行实际翻译
	result, err := translate(ctx, t.LLM, text, sourceLang, targetLang)
	if err != nil {
		log.Printf("Translation error: %v", err)
		return "", fmt.Errorf("translation failed: %w", err)
	}

	log.Printf("Translation result: %s", result)

	if t.CallbacksHandler != nil {
		t.CallbacksHandler.HandleToolEnd(ctx, result)
	}

	return result, nil
}

func (t Translator) Description() string {
	return "A tool that translates text between different languages while preserving meaning and context. Input should be the text to translate."
}

func (t Translator) Name() string {
	return "translator"
}

// 确保 Translator 实现了 tools.Tool 接口
var _ tools.Tool = Translator{}

// translateWithTool uses LangChain tools to perform translation
func translateWithTool(ctx context.Context, llm *openai.LLM, text string, inputLanguage string, outputLanguage string) (string, error) {
	log.Printf("Starting translation with tool: '%s' from %s to %s", text, inputLanguage, outputLanguage)

	// 方法1：直接使用翻译工具（推荐）
	translator := Translator{
		LLM: llm,
	}

	// 直接调用翻译工具
	inputText := fmt.Sprintf("Translate '%s' from %s to %s", text, inputLanguage, outputLanguage)
	result, err := translator.Call(ctx, inputText)
	if err != nil {
		log.Printf("Direct tool call failed: %v", err)
		// 如果工具调用失败，尝试使用简单的翻译函数
		return translate(ctx, llm, text, inputLanguage, outputLanguage)
	}

	log.Printf("Tool translation successful: %s", result)
	return result, nil
}

// translateWithAgent 使用完整的 agent 执行器进行翻译（高级用法）
func translateWithAgent(ctx context.Context, llm *openai.LLM, text string, inputLanguage string, outputLanguage string) (string, error) {
	log.Printf("Starting agent-based translation: '%s' from %s to %s", text, inputLanguage, outputLanguage)

	// 创建翻译工具
	translator := Translator{
		LLM: llm,
	}

	// 创建工具列表
	toolList := []tools.Tool{translator}

	// 创建 agent
	agent := agents.NewOpenAIFunctionsAgent(llm, toolList)

	// 创建执行器
	executor := agents.NewExecutor(agent)

	// 构建输入
	inputText := fmt.Sprintf("Please use the translator tool to translate '%s' from %s to %s", text, inputLanguage, outputLanguage)
	input := map[string]any{
		"input": inputText,
	}

	log.Printf("Agent input: %s", inputText)

	// 执行 agent
	result, err := chains.Call(ctx, executor, input)
	if err != nil {
		log.Printf("Agent execution failed: %v", err)
		// 回退到直接工具调用
		return translateWithTool(ctx, llm, text, inputLanguage, outputLanguage)
	}

	// 提取输出
	if output, ok := result["output"].(string); ok {
		log.Printf("Agent translation successful: %s", output)
		return output, nil
	}

	// 如果没有找到 output 键，尝试其他可能的键
	for key, value := range result {
		if str, ok := value.(string); ok && str != "" {
			log.Printf("Found result in key '%s': %s", key, str)
			return str, nil
		}
	}

	log.Printf("Agent result format unexpected: %+v", result)
	// 回退到直接工具调用
	return translateWithTool(ctx, llm, text, inputLanguage, outputLanguage)
}
