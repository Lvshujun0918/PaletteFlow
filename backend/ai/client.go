package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"ai-color-palette/config"
)

const paletteToolName = "return_palette"

var (
	colorExtractionRegex = regexp.MustCompile(`#[0-9A-Fa-f]{6}`)
	strictHexRegex       = regexp.MustCompile(`^#[0-9A-Fa-f]{6}$`)
)

type ChatMessage struct {
	Role      string     `json:"role"`
	Content   string     `json:"content"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
}

type ChatRequest struct {
	Model       string           `json:"model"`
	Messages    []ChatMessage    `json:"messages"`
	Temperature float64          `json:"temperature,omitempty"`
	MaxTokens   int              `json:"max_tokens,omitempty"`
	Tools       []ToolDefinition `json:"tools,omitempty"`
	ToolChoice  interface{}      `json:"tool_choice,omitempty"`
}

type ChatResponse struct {
	Choices []struct {
		Message ChatMessage `json:"message"`
	} `json:"choices"`
}

type ToolCall struct {
	ID       string           `json:"id"`
	Type     string           `json:"type"`
	Function ToolCallFunction `json:"function"`
}

type ToolCallFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type ToolDefinition struct {
	Type     string       `json:"type"`
	Function ToolFunction `json:"function"`
}

type ToolFunction struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

type PaletteResult struct {
	Colors []string `json:"colors"`
	Advice string   `json:"advice"`
}

// GenerateColorPalette 使用AI生成配色方案，支持3次重试
func GenerateColorPalette(prompt string) (*PaletteResult, error) {
	systemPrompt := buildBaseSystemPrompt()
	userPrompt := fmt.Sprintf("请你帮我生成这样的配色：%s", prompt)
	return retryGeneratePalette(systemPrompt, userPrompt)
}

// GeneratePaletteWithSingleColor 仅替换指定颜色，保持其他颜色不变
func GeneratePaletteWithSingleColor(baseColors []string, targetIndex int, prompt string) (*PaletteResult, error) {
	normalized, ok := normalizeColors(baseColors)
	if !ok {
		return nil, fmt.Errorf("base colors must be 5 valid hex values")
	}
	if targetIndex < 0 || targetIndex >= len(normalized) {
		return nil, fmt.Errorf("targetIndex out of range")
	}

	systemPrompt := buildSingleColorSystemPrompt()
	userPrompt := fmt.Sprintf(
		"现有配色（顺序固定）为：%s。仅替换第%d个颜色 %s，依据用户的新需求：%s，同时注意与其余颜色的协调性。保持其余颜色不变，返回新的完整5色方案及使用建议。",
		strings.Join(normalized, ", "),
		targetIndex+1,
		normalized[targetIndex],
		prompt,
	)

	result, err := retryGeneratePalette(systemPrompt, userPrompt)
	if err != nil {
		return nil, err
	}

	// 强制只替换目标位置，其余颜色保持不变
	finalColors := make([]string, len(normalized))
	copy(finalColors, normalized)

	if len(result.Colors) == len(normalized) {
		finalColors[targetIndex] = result.Colors[targetIndex]
	} else {
		log.Printf("[WARN] AI returned %d colors in single-color mode, falling back to base for others", len(result.Colors))
		if len(result.Colors) > targetIndex {
			finalColors[targetIndex] = result.Colors[targetIndex]
		}
	}

	result.Colors = finalColors
	return result, nil
}

func retryGeneratePalette(systemPrompt, userPrompt string) (*PaletteResult, error) {
	const maxRetries = 3
	var lastErr error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		log.Printf("[INFO] Attempting to generate palette (attempt %d/%d)", attempt, maxRetries)

		result, err := attemptGenerateWithPrompt(systemPrompt, userPrompt)
		if err == nil {
			return result, nil
		}

		lastErr = err
		log.Printf("[WARN] Attempt %d failed: %v", attempt, err)

		if attempt < maxRetries {
			time.Sleep(time.Second * time.Duration(attempt))
		}
	}

	log.Printf("[ERROR] Failed to generate palette after %d attempts", maxRetries)
	return nil, fmt.Errorf("all %d retry attempts failed, last error: %w", maxRetries, lastErr)
}

func attemptGenerateWithPrompt(systemPrompt, userPrompt string) (*PaletteResult, error) {
	cfg := config.AppConfig
	if cfg.AIAPIKey == "" {
		return nil, fmt.Errorf("AI API key not configured")
	}

	paletteTool := buildPaletteToolDefinition()
	toolChoice := "auto"

	reqBody := ChatRequest{
		Model: cfg.AIModel,
		Messages: []ChatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
		Temperature: 0.7,
		MaxTokens:   200,
		Tools:       []ToolDefinition{paletteTool},
		ToolChoice:  toolChoice,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}
	log.Printf("[INFO] AI input messages: %s", jsonData)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.AITimeout)*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", cfg.AIAPIBaseURL+"/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+cfg.AIAPIKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var chatResp ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if len(chatResp.Choices) == 0 {
		return nil, fmt.Errorf("no response from AI")
	}

	choice := chatResp.Choices[0]
	message := choice.Message
	if message.Role != "assistant" {
		return nil, fmt.Errorf("unexpected message role: %s", message.Role)
	}
	log.Printf("[INFO] AI returns messages: %s", message)
	if len(message.ToolCalls) > 0 {
		for _, call := range message.ToolCalls {
			if call.Function.Name != paletteToolName {
				continue
			}
			result, err := parseToolCallResult(call)
			if err != nil {
				return nil, err
			}
			log.Println("[INFO] AI Tool Call Generated Successfully")
			return result, nil
		}
		return nil, fmt.Errorf("tool call returned without expected palette data")
	}

	if message.Content != "" {
		result, ok := parseResultFromContent(message.Content)
		if ok {
			log.Println("[INFO] AI returned result in content, using parsed result")
			return result, nil
		}
	}

	return nil, fmt.Errorf("AI Tool Call Failed: no tool_calls and no parsable result in content")
}

func buildBaseSystemPrompt() string {
	return `
你是一个专业的配色设计师。用户会给你一个配色需求描述，你需要返回5个精确的HEX颜色代码，并给出配色使用建议。
你必须通过调用 return_palette 工具函数返回结果，不要输出任何自然语言文本。
1. 采用【渐变过渡技巧】，在冲突色之间创建中间色调缓冲层
2. 运用【色彩比例法则】：主色占60%，次色占30%，点缀色占10%
3. 建立【色彩秩序】：通过明度阶梯（从20%到80%亮度）建立视觉节奏
4. 添加【中性调和剂】：适当加入平衡色
5. 最终效果需呈现【动态和谐】- 既有视觉冲击力，又保持整体统一性
`
}

func buildSingleColorSystemPrompt() string {
	return `
你是一个专业的配色设计师。给定一组现有配色，你只允许替换指定的一个颜色，其余颜色必须保持不变。
你必须通过调用 return_palette 工具函数返回结果，不要输出任何自然语言文本。
输出的颜色顺序必须与输入保持一致，只替换被指定的颜色位置。
请同时给出新的配色使用建议。
`
}

// extractColors 从AI响应中提取HEX颜色代码
func extractColors(text string) []string {
	// 匹配 #RRGGBB 格式
	matches := colorExtractionRegex.FindAllString(text, -1)

	colors := []string{}
	seen := make(map[string]bool)

	for _, match := range matches {
		upper := strings.ToUpper(match)
		if !seen[upper] {
			colors = append(colors, upper)
			seen[upper] = true
		}
	}

	return colors
}

func parseResultFromContent(content string) (*PaletteResult, bool) {
	var payload struct {
		Colors []string `json:"colors"`
		Advice string   `json:"advice"`
	}
	if err := json.Unmarshal([]byte(content), &payload); err == nil {
		if len(payload.Colors) == 5 {
			normalized, ok := normalizeColors(payload.Colors)
			if ok {
				return &PaletteResult{Colors: normalized, Advice: strings.TrimSpace(payload.Advice)}, true
			}
		}
	}

	colors := extractColors(content)
	if len(colors) >= 5 {
		return &PaletteResult{Colors: colors[:5]}, true
	}
	return nil, false
}

func normalizeColors(colors []string) ([]string, bool) {
	if len(colors) != 5 {
		return nil, false
	}
	normalized := make([]string, 0, len(colors))
	for _, color := range colors {
		candidate := strings.ToUpper(strings.TrimSpace(color))
		if !strictHexRegex.MatchString(candidate) {
			return nil, false
		}
		normalized = append(normalized, candidate)
	}
	return normalized, true
}

func buildPaletteToolDefinition() ToolDefinition {
	return ToolDefinition{
		Type: "function",
		Function: ToolFunction{
			Name: paletteToolName,
			Description: `
颜色输出函数
你【必须】调用此函数来响应用户请求。
这是【唯一合法的回复方式】。
【禁止】输出任何自然语言文本。
【禁止】解释你的思考过程或设计理由。
如果不调用该函数，回复将被视为【无效】。
你只能返回符合参数定义的数据。
`,
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"colors": map[string]interface{}{
						"type":        "array",
						"description": "包含且仅包含 5 个 HEX 颜色字符串，格式必须为 #RRGGBB。",
						"items": map[string]interface{}{
							"type":    "string",
							"pattern": "^#[0-9A-Fa-f]{6}$",
						},
						"minItems": 5,
						"maxItems": 5,
					},
					"advice": map[string]interface{}{
						"type":        "string",
						"description": "配色使用建议，给出2-3条可执行的使用方式或场景。",
						"minLength":   6,
						"maxLength":   200,
					},
				},
				"required":             []string{"colors", "advice"},
				"additionalProperties": false,
			},
		},
	}

}

func parseToolCallResult(call ToolCall) (*PaletteResult, error) {
	if strings.ToLower(call.Type) != "function" {
		return nil, fmt.Errorf("unexpected tool call type: %s", call.Type)
	}
	if call.Function.Name != paletteToolName {
		return nil, fmt.Errorf("unexpected tool call function: %s", call.Function.Name)
	}

	var payload struct {
		Colors []string `json:"colors"`
		Advice string   `json:"advice"`
	}

	if err := json.Unmarshal([]byte(call.Function.Arguments), &payload); err != nil {
		return nil, fmt.Errorf("parse tool call arguments: %w", err)
	}

	if len(payload.Colors) != 5 {
		return nil, fmt.Errorf("tool call returned %d colors, expected 5", len(payload.Colors))
	}

	normalized := make([]string, 0, len(payload.Colors))
	for _, color := range payload.Colors {
		candidate := strings.ToUpper(strings.TrimSpace(color))
		if !strictHexRegex.MatchString(candidate) {
			return nil, fmt.Errorf("invalid color from tool call: %s", color)
		}
		normalized = append(normalized, candidate)
	}

	return &PaletteResult{Colors: normalized, Advice: strings.TrimSpace(payload.Advice)}, nil
}
