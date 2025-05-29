package translator

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/tools"
)

// Translator 实现了 tools.Tool 接口用于翻译任务
type Translator struct {
	LLM              *openai.LLM
	CallbacksHandler callbacks.Handler
}

// NewTranslator 创建一个新的翻译器实例
func NewTranslator(llm *openai.LLM) *Translator {
	return &Translator{
		LLM: llm,
	}
}

// Call 实现实际的翻译功能
func (t *Translator) Call(ctx context.Context, input string) (string, error) {
	log.Printf("Translator tool called with input: %s", input)

	if t.CallbacksHandler != nil {
		t.CallbacksHandler.HandleToolStart(ctx, input)
	}

	// 尝试解析 JSON 输入
	var text, sourceLang, targetLang string
	if strings.HasPrefix(strings.TrimSpace(input), "{") {
		// 尝试解析 JSON
		var params struct {
			Text           string `json:"text"`
			SourceLanguage string `json:"source_language"`
			TargetLanguage string `json:"target_language"`
		}
		if err := json.Unmarshal([]byte(input), &params); err == nil {
			text = params.Text
			sourceLang = params.SourceLanguage
			targetLang = params.TargetLanguage
		}
	}

	// 如果 JSON 解析失败，使用默认处理
	if text == "" {
		text = strings.Trim(input, "'\"")
		text = strings.TrimSpace(text)
		sourceLang = "English"
		targetLang = "Chinese"
	}

	// 设置默认值
	if sourceLang == "" {
		sourceLang = "English"
	}
	if targetLang == "" {
		targetLang = "Chinese"
	}

	log.Printf("Translating '%s' from %s to %s", text, sourceLang, targetLang)

	// 使用内置的 translate 函数进行实际翻译
	result, err := Translate(ctx, t.LLM, text, sourceLang, targetLang)
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

func (t *Translator) Description() string {
	return `A translation tool that converts text between languages.
Parameters:
- text: The text to translate (required)
- source_language: The source language (default: English)
- target_language: The target language (default: Chinese)

Example: "Hello world" -> "你好，世界"`
}

func (t *Translator) Name() string {
	return "translate_text"
}

// 确保 Translator 实现了 tools.Tool 接口
var _ tools.Tool = (*Translator)(nil)
