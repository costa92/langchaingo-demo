package mock

import (
	"context"
	"testing"
)

func TestMockTranslator_Call(t *testing.T) {
	ctx := context.Background()
	translator := NewMockTranslator()

	tests := []struct {
		name       string
		input      string
		wantOutput string
		wantError  bool
	}{
		{
			name:       "Basic Translation",
			input:      "Translate 'hello world' from English to Chinese",
			wantOutput: "你好，世界",
			wantError:  false,
		},
		{
			name:       "Empty Input",
			input:      "",
			wantOutput: "",
			wantError:  true,
		},
		{
			name:       "Invalid Format",
			input:      "Invalid format",
			wantOutput: "",
			wantError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := translator.Call(ctx, tt.input)
			if (err != nil) != tt.wantError {
				t.Errorf("MockTranslator.Call() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if !tt.wantError && got != tt.wantOutput {
				t.Errorf("MockTranslator.Call() = %v, want %v", got, tt.wantOutput)
			}
		})
	}
}

func TestMockCalculator_Call(t *testing.T) {
	ctx := context.Background()
	calculator := NewMockCalculator()

	tests := []struct {
		name       string
		input      string
		wantOutput string
		wantError  bool
	}{
		{
			name:       "Basic Addition",
			input:      "Calculate 1 + 2",
			wantOutput: "3",
			wantError:  false,
		},
		{
			name:       "Empty Input",
			input:      "",
			wantOutput: "",
			wantError:  true,
		},
		{
			name:       "Invalid Format",
			input:      "Invalid format",
			wantOutput: "",
			wantError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := calculator.Call(ctx, tt.input)
			if (err != nil) != tt.wantError {
				t.Errorf("MockCalculator.Call() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if !tt.wantError && got != tt.wantOutput {
				t.Errorf("MockCalculator.Call() = %v, want %v", got, tt.wantOutput)
			}
		})
	}
}

func TestMockTools_Interface(t *testing.T) {
	// 测试翻译器接口
	translator := NewMockTranslator()
	if name := translator.Name(); name != "mock_translator" {
		t.Errorf("MockTranslator.Name() = %v, want %v", name, "mock_translator")
	}
	if desc := translator.Description(); desc == "" {
		t.Error("MockTranslator.Description() returned empty string")
	}

	// 测试计算器接口
	calculator := NewMockCalculator()
	if name := calculator.Name(); name != "mock_calculator" {
		t.Errorf("MockCalculator.Name() = %v, want %v", name, "mock_calculator")
	}
	if desc := calculator.Description(); desc == "" {
		t.Error("MockCalculator.Description() returned empty string")
	}
}

func TestRunMockTests(t *testing.T) {
	// 这个测试主要是确保 RunMockTests 函数不会 panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("RunMockTests() panicked: %v", r)
		}
	}()

	RunMockTests()
}
