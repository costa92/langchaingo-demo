# LangChain Go Agent 翻译工具示例

这个项目展示了如何使用 [LangChain Go](https://github.com/tmc/langchaingo) 创建一个带有自定义工具的 AI Agent，专门用于文本翻译任务。

## 🚀 功能特性

- ✅ **自定义工具实现**：实现了符合 `tools.Tool` 接口的翻译工具
- ✅ **Agent 集成**：使用 OpenAI Functions Agent 来调用工具
- ✅ **错误处理**：完善的错误处理和日志记录
- ✅ **模拟测试**：提供模拟工具用于测试和开发
- ✅ **多工具支持**：支持多个工具（翻译器、计算器等）

## 🛠️ 安装和设置

### 1. 克隆项目

```bash
git clone <your-repo-url>
cd ai-damo
```

### 2. 安装依赖

```bash
go mod tidy
```

### 3. 设置环境变量（可选）

如果你有真实的 API 密钥，可以设置以下环境变量：

```bash
export SILICONFLOW_API_KEY="your_api_key_here"
export SILICONFLOW_API_URL="https://api.siliconflow.cn/v1"  # 可选，有默认值
```

如果不设置环境变量，程序会自动使用模拟工具进行测试。

## 🎯 使用方法

### 运行示例

#### 使用模拟工具测试（无需 API 密钥）

```bash
go run main.go
```

#### 使用真实 API 测试

```bash
# 方法 1：直接设置环境变量
export SILICONFLOW_API_KEY="your_api_key_here"
go run main.go

# 方法 2：使用测试脚本
./test_with_api.sh your_api_key_here
```

### 输出示例

#### 模拟工具测试（无 API 密钥）
```
=== Mock Translation Test ===
Original: Hello world
Translated: 你好，世界
Tool Name: mock_translator
Tool Description: A mock tool that translates text between different languages for testing purposes.

=== Mock Agent Test ===

Testing tool: mock_translator
Input: Hello world
Output: 你好，世界

Testing tool: mock_calculator
Input: 2 + 3
Output: 5
```

#### 真实 API 测试（有 API 密钥）
```
🚀 Testing LangChain Go Agent with real API...
Using API Key: sk-vighydx...

📝 Running translation test...
2025/05/25 21:43:06 Using model: Qwen/Qwen2.5-72B-Instruct
2025/05/25 21:43:06 Starting translation with tool: 'Hello world' from English to Chinese
2025/05/25 21:43:06 Translator tool called with input: Translate 'Hello world' from English to Chinese
2025/05/25 21:43:06 Translating 'Hello world' from English to Chinese
2025/05/25 21:43:06 Translation result: 你好，世界
2025/05/25 21:43:06 Tool translation successful: 你好，世界
Original: Hello world
Translated: 你好，世界

✅ Test completed!
```

## 📚 代码结构

### 核心组件

1. **Translator 工具**：实现真实的翻译功能
2. **MockTranslator 工具**：用于测试的模拟翻译器
3. **MockCalculator 工具**：用于测试的模拟计算器
4. **Agent 执行器**：管理工具调用的 AI Agent

### 翻译方法对比

项目提供了三种不同的翻译实现方式：

1. **`translate()`** - 直接翻译函数
   - ✅ 简单直接，性能最好
   - ✅ 错误处理简单
   - ❌ 不支持工具链和复杂推理

2. **`translateWithTool()`** - 工具调用方式（推荐）
   - ✅ 支持工具接口，便于扩展
   - ✅ 可以与其他工具组合
   - ✅ 错误处理完善，有回退机制
   - ✅ 性能良好

3. **`translateWithAgent()`** - Agent 执行器方式
   - ✅ 支持复杂的推理和工具链
   - ✅ 可以处理多步骤任务
   - ❌ 复杂度高，可能出现格式问题
   - ❌ 性能相对较低

### 工具接口实现

每个工具都必须实现 `tools.Tool` 接口：

```go
type Tool interface {
    Call(ctx context.Context, input string) (string, error)
    Description() string
    Name() string
}
```

### 示例工具实现

```go
type MockTranslator struct {
    CallbacksHandler callbacks.Handler
}

func (m MockTranslator) Call(ctx context.Context, input string) (string, error) {
    // 实现工具逻辑
    return "翻译结果", nil
}

func (m MockTranslator) Description() string {
    return "工具描述"
}

func (m MockTranslator) Name() string {
    return "tool_name"
}
```

## 🔧 问题解决

### 常见问题

1. **API 返回 400 错误**
   - 检查输入格式是否正确
   - 确保 Agent 输入包含 "input" 键
   - 验证模型是否支持 function calling

2. **API 返回 401 错误**
   - 检查 API 密钥是否正确设置
   - 验证 API URL 是否正确

3. **工具未被调用**
   - 确保工具正确实现了 `tools.Tool` 接口
   - 检查工具描述是否清晰
   - 验证 Agent 输入格式
   - **重要**：确保使用支持 function calling 的模型

4. **模型不支持 function calling**
   - ❌ 不支持：`deepseek-ai/DeepSeek-R1`
   - ✅ 支持：`Qwen/Qwen2.5-72B-Instruct`, `gpt-4o-mini`, `gpt-3.5-turbo`

### 调试技巧

1. **启用详细日志**：代码中已包含详细的日志输出
2. **使用模拟工具**：先用模拟工具测试逻辑
3. **检查返回值**：确保工具返回正确的数据类型

## 🎨 扩展功能

### 添加新工具

1. 创建新的结构体实现 `tools.Tool` 接口
2. 在工具列表中添加新工具
3. 更新 Agent 配置

```go
type MyCustomTool struct {
    CallbacksHandler callbacks.Handler
}

func (t MyCustomTool) Call(ctx context.Context, input string) (string, error) {
    // 实现自定义逻辑
    return "结果", nil
}

func (t MyCustomTool) Description() string {
    return "自定义工具描述"
}

func (t MyCustomTool) Name() string {
    return "my_custom_tool"
}
```

### 集成其他 LLM

修改 LLM 初始化代码：

```go
// 使用 OpenAI
llm, err := openai.New(
    openai.WithModel("gpt-3.5-turbo"),
    openai.WithToken(apiKey),
)

// 使用其他提供商
llm, err := openai.New(
    openai.WithModel("your-model"),
    openai.WithBaseURL("your-api-url"),
    openai.WithToken(apiKey),
)
```

## 📖 参考资料

- [LangChain Go 官方文档](https://tmc.github.io/langchaingo/)
- [LangChain Go GitHub](https://github.com/tmc/langchaingo)
- [LangChain Go 示例](https://github.com/tmc/langchaingo/tree/main/examples)
- [OpenAI API 文档](https://platform.openai.com/docs)

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

## �� 许可证

MIT License 