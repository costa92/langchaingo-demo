package agent

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/tmc/langchaingo/agents"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/tools"

	"github.com/costa92/langchaingo-demo/pkg/translator"
)

// TranslateWithAgent 使用完整的 agent 执行器进行翻译
func TranslateWithAgent(ctx context.Context, llm *openai.LLM, text string, inputLanguage string, outputLanguage string) (string, error) {
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

	// 创建翻译工具
	trans := translator.NewTranslator(llm)
	if trans == nil {
		return "", fmt.Errorf("failed to create translator")
	}

	log.Printf("Created translator tool: name=%s", trans.Name())
	log.Printf("Translator description: %s", trans.Description())

	// 创建工具列表
	toolList := []tools.Tool{trans}

	// 创建 agent 执行器，使用更简单的系统消息
	systemMessage := `You are a translation assistant. Your task is to translate text from one language to another.
Please use the translate_text tool to perform translations. Return only the translated text.`

	agentOpts := agents.NewOpenAIOption().WithSystemMessage(systemMessage)
	agent := agents.NewOpenAIFunctionsAgent(llm, toolList, agentOpts)
	if agent == nil {
		return "", fmt.Errorf("failed to create agent: agent is nil")
	}
	executor := agents.NewExecutor(agent, agents.WithMaxIterations(2)) // 减少最大迭代次数
	if executor == nil {
		return "", fmt.Errorf("failed to create executor: executor is nil")
	}

	log.Printf("Created agent executor with max iterations: 2")

	// 构建简化的输入提示
	inputText := fmt.Sprintf("Translate '%s' from %s to %s.", text, inputLanguage, outputLanguage)

	input := map[string]any{
		"input": inputText,
	}

	log.Printf("Preparing to execute agent with input: %s", inputText)

	// 添加重试机制
	maxRetries := 2 // 减少重试次数
	var lastError error

	for retry := 0; retry < maxRetries; retry++ {
		if retry > 0 {
			log.Printf("Retrying translation (attempt %d/%d)...", retry+1, maxRetries)
			time.Sleep(time.Duration(retry) * time.Second)
		}

		// 执行 agent
		result, err := chains.Call(ctx, executor, input)
		if err == nil {
			// 成功执行，处理结果
			if result == nil {
				return "", fmt.Errorf("agent returned nil result")
			}

			log.Printf("Agent execution completed, full result: %+v", result)

			// 提取输出
			output, ok := result["output"].(string)
			if !ok {
				log.Printf("Failed to extract output from result: %+v", result)
				return "", fmt.Errorf("failed to extract output from result")
			}

			// 清理输出
			output = strings.TrimSpace(output)
			if output == "" {
				log.Printf("Agent returned empty output")
				return "", fmt.Errorf("agent returned empty output")
			}

			log.Printf("Successfully extracted output: %s", output)
			return output, nil
		}

		lastError = err
		log.Printf("Translation attempt %d failed: %v", retry+1, err)

		// 处理特定错误
		if strings.Contains(err.Error(), "422") {
			log.Printf("Invalid parameters error detected, attempting direct translation...")
			return translator.Translate(ctx, llm, text, inputLanguage, outputLanguage)
		}

		// 处理其他错误类型
		if strings.Contains(err.Error(), "context deadline exceeded") ||
			strings.Contains(err.Error(), "timeout") {
			log.Printf("Timeout error detected, will retry...")
			continue
		}
	}

	// 如果所有重试都失败，尝试直接翻译
	log.Printf("All agent translation attempts failed (last error: %v), trying direct translation...", lastError)
	return translator.Translate(ctx, llm, text, inputLanguage, outputLanguage)
}
