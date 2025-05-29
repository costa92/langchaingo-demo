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

// TranslateWithAgent 使用完整的 agent 执行器进行翻译（性能优化版本）
func TranslateWithAgentOptimized(ctx context.Context, llm *openai.LLM, text string, inputLanguage string, outputLanguage string) (string, error) {
	// 添加超时控制
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
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

	log.Printf("Starting optimized agent-based translation: '%s' from %s to %s", text, inputLanguage, outputLanguage)

	// 创建翻译工具（只创建一次）
	trans := translator.NewTranslator(llm)
	if trans == nil {
		return "", fmt.Errorf("failed to create translator")
	}

	log.Printf("Created translator tool: name=%s", trans.Name())
	log.Printf("Translator description: %s", trans.Description())

	// 创建工具列表
	toolList := []tools.Tool{trans}

	// 构建简化的输入提示
	inputText := fmt.Sprintf("Translate '%s' from %s to %s.", text, inputLanguage, outputLanguage)

	// 初始化 agent 执行器（只初始化一次）
	executor, err := agents.Initialize(
		llm,
		toolList,
		agents.ZeroShotReactDescription,
		agents.WithMaxIterations(3),
	)
	if err != nil {
		return "", fmt.Errorf("failed to initialize agent: %w", err)
	}

	// 添加优化的重试机制
	maxRetries := 2
	var lastError error

	for retry := 0; retry < maxRetries; retry++ {
		// 检查上下文是否已取消
		if ctx.Err() != nil {
			return "", ctx.Err()
		}

		if retry > 0 {
			log.Printf("Retrying translation (attempt %d/%d)...", retry+1, maxRetries)
			// 使用指数退避策略
			backoff := time.Duration(retry*retry) * 100 * time.Millisecond
			time.Sleep(backoff)
		}

		// 执行 agent
		result, err := chains.Run(ctx, executor, inputText)
		if err != nil {
			log.Printf("Translation attempt %d failed: %v", retry+1, err)
			lastError = err
			continue
		}

		log.Printf("Translation successful: %s", result)
		return result, nil
	}

	return "", fmt.Errorf("translation failed after %d retries, last error: %w", maxRetries, lastError)
}
