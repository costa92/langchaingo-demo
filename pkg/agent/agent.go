package agent

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/tmc/langchaingo/agents"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/tools"

	"github.com/costa92/langchaingo-demo/pkg/translator"
)

// TranslateWithAgent 使用完整的 agent 执行器进行翻译
func TranslateWithAgent(ctx context.Context, llm *openai.LLM, text string, inputLanguage string, outputLanguage string) (string, error) {
	// 添加超时控制，避免长时间阻塞
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	// 输入验证
	if text == "" {
		return "", fmt.Errorf("empty text")
	}
	if inputLanguage == "" {
		return "", fmt.Errorf("empty input language")
	}
	if outputLanguage == "" {
		return "", fmt.Errorf("empty output language")
	}
	if llm == nil {
		return "", fmt.Errorf("LLM client is nil")
	}

	log.Printf("Starting agent-based translation: '%s' from %s to %s", text, inputLanguage, outputLanguage)

	// 优化工具初始化，使用更高效的配置
	translatorTool := translator.NewTranslator(llm)
	calculatorTool := tools.Calculator{}
	agentTools := []tools.Tool{translatorTool, &calculatorTool}

	// 构建简化的输入提示
	inputText := fmt.Sprintf("Translate '%s' from %s to %s.", text, inputLanguage, outputLanguage)

	agent := agents.NewOneShotAgent(llm, agentTools, agents.WithMaxIterations(2))

	executor := agents.NewExecutor(agent)
	// 执行 agent
	result, err := chains.Run(ctx, executor, inputText)
	if err != nil {
		log.Printf("Translation failed: %v", err)
		return "", fmt.Errorf("translation failed: %w", err)
	}
	log.Printf("Translation successful: %s", result)
	return result, nil
}
