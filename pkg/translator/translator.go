package translator

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/prompts"
)

// 配置常量
const (
	defaultTimeout = 60 * time.Second // 默认超时时间
	cacheDuration  = 24 * time.Hour   // 缓存有效期
	maxConcurrency = 2                // 最大并发数
	batchSize      = 3                // 批处理大小
)

// TranslationCache 用于缓存翻译结果
type TranslationCache struct {
	cache map[string]cacheEntry
	mu    sync.RWMutex
}

type cacheEntry struct {
	result    string
	timestamp time.Time
}

var (
	defaultCache = &TranslationCache{
		cache: make(map[string]cacheEntry),
	}
)

// getCacheKey 生成缓存键
func getCacheKey(text, inputLang, outputLang string) string {
	return fmt.Sprintf("%s:%s:%s", text, inputLang, outputLang)
}

// Get 从缓存获取翻译结果
func (c *TranslationCache) Get(text, inputLang, outputLang string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	key := getCacheKey(text, inputLang, outputLang)
	if entry, ok := c.cache[key]; ok {
		if time.Since(entry.timestamp) < cacheDuration {
			return entry.result, true
		}
		// 清理过期缓存
		delete(c.cache, key)
	}
	return "", false
}

// Set 设置缓存
func (c *TranslationCache) Set(text, inputLang, outputLang, result string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := getCacheKey(text, inputLang, outputLang)
	c.cache[key] = cacheEntry{
		result:    result,
		timestamp: time.Now(),
	}
}

// Translate 是一个基本的翻译函数
func Translate(ctx context.Context, llm *openai.LLM, text string, inputLanguage string, outputLanguage string) (string, error) {
	// 验证输入
	if text == "" {
		return "", fmt.Errorf("empty text input")
	}
	if inputLanguage == "" {
		return "", fmt.Errorf("empty input language")
	}
	if outputLanguage == "" {
		return "", fmt.Errorf("empty output language")
	}

	// 检查缓存
	if result, ok := defaultCache.Get(text, inputLanguage, outputLanguage); ok {
		log.Printf("Cache hit for text: %s", text)
		return result, nil
	}

	// 优化的 prompt 模板
	prompt := prompts.NewPromptTemplate(
		`Translate "{{.text}}" from {{.inputLanguage}} to {{.outputLanguage}}. Output the translation only, no explanations.`,
		[]string{"inputLanguage", "outputLanguage", "text"},
	)

	llmChain := chains.NewLLMChain(llm, prompt)

	// 设置超时
	timeoutCtx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	outputValues, err := chains.Call(timeoutCtx, llmChain, map[string]any{
		"inputLanguage":  inputLanguage,
		"outputLanguage": outputLanguage,
		"text":           text,
	})
	if err != nil {
		// 记录详细错误信息，帮助定位 OpenAI API 返回 400 错误的原因
		log.Printf("OpenAI API 调用失败，详细错误信息: %v", err)
		return "", fmt.Errorf("translation failed: %w", err)
	}

	out, ok := outputValues[llmChain.OutputKey].(string)
	if !ok {
		return "", fmt.Errorf("invalid chain return")
	}

	// 缓存结果
	defaultCache.Set(text, inputLanguage, outputLanguage, out)
	return out, nil
}

// TranslateBatch 批量翻译文本
func TranslateBatch(ctx context.Context, llm *openai.LLM, texts []string, inputLanguage string, outputLanguage string) ([]string, error) {
	if len(texts) == 0 {
		return nil, fmt.Errorf("empty texts input")
	}

	results := make([]string, len(texts))
	errChan := make(chan error, len(texts))
	var wg sync.WaitGroup

	// 限制并发数
	semaphore := make(chan struct{}, maxConcurrency)

	// 分批处理
	for i := 0; i < len(texts); i += batchSize {
		end := i + batchSize
		if end > len(texts) {
			end = len(texts)
		}

		batch := texts[i:end]
		for j, text := range batch {
			wg.Add(1)
			go func(index, batchIndex int, text string) {
				defer wg.Done()

				// 获取信号量
				semaphore <- struct{}{}
				defer func() { <-semaphore }()

				// 检查缓存
				if result, ok := defaultCache.Get(text, inputLanguage, outputLanguage); ok {
					results[index] = result
					return
				}

				// 为每个翻译任务设置独立的超时
				taskCtx, cancel := context.WithTimeout(ctx, defaultTimeout)
				defer cancel()

				result, err := Translate(taskCtx, llm, text, inputLanguage, outputLanguage)
				if err != nil {
					errChan <- fmt.Errorf("failed to translate text at index %d: %w", index, err)
					return
				}
				results[index] = result

				// 添加延迟以避免 API 限制
				time.Sleep(500 * time.Millisecond)
			}(i+j, i/batchSize, text)
		}

		// 等待当前批次完成
		wg.Wait()

		// 检查错误
		select {
		case err := <-errChan:
			close(errChan)
			return nil, fmt.Errorf("batch translation error: %v", err)
		default:
			// 没有错误，继续处理
		}

		// 批次间添加延迟以避免 API 限制
		if end < len(texts) {
			time.Sleep(1 * time.Second)
		}
	}

	return results, nil
}

// TranslateWithTool 使用 LangChain 工具进行翻译
func TranslateWithTool(ctx context.Context, llm *openai.LLM, text string, inputLanguage string, outputLanguage string) (string, error) {
	// 验证输入
	if text == "" {
		return "", fmt.Errorf("empty text input")
	}
	if inputLanguage == "" {
		return "", fmt.Errorf("empty input language")
	}
	if outputLanguage == "" {
		return "", fmt.Errorf("empty output language")
	}

	// 检查缓存
	if result, ok := defaultCache.Get(text, inputLanguage, outputLanguage); ok {
		log.Printf("Cache hit for text: %s", text)
		return result, nil
	}

	log.Printf("Starting translation with tool: '%s' from %s to %s", text, inputLanguage, outputLanguage)

	// 设置超时
	timeoutCtx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	translator := NewTranslator(llm)
	inputText := fmt.Sprintf("Translate '%s' from %s to %s. Output the translation only.", text, inputLanguage, outputLanguage)
	result, err := translator.Call(timeoutCtx, inputText)
	if err != nil {
		log.Printf("Direct tool call failed: %v", err)
		return "", err
		// return Translate(ctx, llm, text, inputLanguage, outputLanguage)
	}

	// 缓存结果
	defaultCache.Set(text, inputLanguage, outputLanguage, result)
	log.Printf("Tool translation successful: %s", result)
	return result, nil
}
