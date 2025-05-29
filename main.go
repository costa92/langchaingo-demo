// Package main is the main package for the application.
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/tmc/langchaingo/llms/openai"

	"github.com/costa92/langchaingo-demo/pkg/agent"
	"github.com/costa92/langchaingo-demo/pkg/mock"
	"github.com/costa92/langchaingo-demo/pkg/translator"
)

func main() {
	ctx := context.Background()

	// 检查环境变量
	apiURL := "https://api.deepseek.com/v1"
	apiKey := os.Getenv("DEEPSEEK_API_KEY")

	if apiURL == "" {
		apiURL = "https://api.siliconflow.cn/v1"
		log.Printf("SILICONFLOW_API_URL not set, using default: %s", apiURL)
	}

	if apiKey == "" {
		log.Printf("SILICONFLOW_API_KEY not set, using mock translation for testing")
		// 使用模拟翻译进行测试
		mock.RunMockTests()
		return
	}

	// 打印配置信息（注意不要打印完整的 API Key）
	log.Printf("Configuration:")
	log.Printf("API URL: %s", apiURL)
	log.Printf("API Key: %s...%s", apiKey[:4], apiKey[len(apiKey)-4:])

	// 使用支持 function calling 的模型
	model := "deepseek-chat" // 使用更稳定的模型
	log.Printf("Using model: %s", model)

	llm, err := openai.New(
		openai.WithModel(model),
		openai.WithBaseURL(apiURL),
		openai.WithToken(apiKey),
	)
	if err != nil {
		log.Fatalf("Failed to initialize LLM: %v", err)
	}

	// Test the translator using different methods
	text := "I like you"

	// str, err := basicTranslation(ctx, llm, text, "English", "Chinese")
	// if err != nil {
	// 	log.Printf("Basic translation failed: %v", err)
	// }
	// fmt.Printf("Basic translation: %s\n", str)
	// str, err := toolTranslation(ctx, llm, text, "English", "Chinese")
	// if err != nil {
	// 	log.Printf("Tool translation failed: %v", err)
	// }
	// fmt.Printf("Tool translation: %s\n", str)
	str, err := agentTranslation(ctx, llm, text, "English", "Chinese")
	if err != nil {
		log.Printf("Agent translation failed: %v", err)
	}
	fmt.Printf("Agent translation: %s\n", str)

}

func basicTranslation(ctx context.Context, llm *openai.LLM, text string, inputLanguage string, outputLanguage string) (string, error) {
	log.Printf("\nTrying basic translation...")
	translated, err := translator.Translate(ctx, llm, text, "English", "Chinese")
	if err != nil {
		log.Printf("Basic translation failed: %v", err)
		log.Println("Falling back to mock translation...")
		mock.RunMockTests()
		return "", err
	} else {
		fmt.Printf("\n=== Basic Translation ===\n")
		fmt.Printf("Original: %s\nTranslated: %s\n", text, translated)
	}
	return translated, nil
}

func toolTranslation(ctx context.Context, llm *openai.LLM, text string, inputLanguage string, outputLanguage string) (string, error) {
	log.Printf("\nTrying tool-based translation...")
	translated, err := translator.TranslateWithTool(ctx, llm, text, "English", "Chinese")
	if err != nil {
		log.Printf("Tool-based translation failed: %v", err)
		return "", err
	} else {
		fmt.Printf("\n=== Tool-based Translation ===\n")
		fmt.Printf("Original: %s\nTranslated: %s\n", text, translated)
	}
	return translated, nil
}

func agentTranslation(ctx context.Context, llm *openai.LLM, text string, inputLanguage string, outputLanguage string) (string, error) {
	log.Printf("\nTrying agent-based translation...")
	translated, err := agent.TranslateWithAgent(ctx, llm, text, "English", "Chinese")
	if err != nil {
		log.Printf("Agent-based translation failed: %v", err)
		return "", err
	} else {
		fmt.Printf("\n=== Agent-based Translation ===\n")
		fmt.Printf("Original: %s\nTranslated: %s\n", text, translated)
	}
	return translated, nil
}
