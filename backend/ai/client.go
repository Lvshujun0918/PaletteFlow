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

// GenerateColorPalette 使用AI生成配色方案
func GenerateColorPalette(prompt string) ([]string, error) {
	cfg := config.AppConfig
	if cfg.AIAPIKey == "" {
		return nil, fmt.Errorf("AI API key not configured")
	}

	systemPrompt := `你是一个专业的配色设计师。用户会给你一个配色需求描述，你需要返回5个精确的HEX颜色代码。

要求：
1. 返回格式必须严格为：#RRGGBB（例如：#FF5733）
2. 恰好返回5个颜色
3. 颜色之间用逗号或空格分隔
4. 不要包含任何解释文字，只返回颜色代码
5. 颜色应该协调、美观、符合用户需求
6. 你必须通过调用（Function Call）名为 return_palette 的函数工具返回结果，不要直接输出文本。`

	userPrompt := fmt.Sprintf("请你帮我生成这样的配色：%s", prompt)

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
			colors, err := parseToolCallColors(call)
			if err != nil {
				return nil, err
			}
			log.Println("[INFO] AI Tool Call Generated Successfully")
			return colors, nil
		}
		return nil, fmt.Errorf("tool call returned without expected palette data")
	}

	if message.Content != "" {
		colors, ok := parseColorsFromContent(message.Content)
		if ok {
			log.Println("[INFO] AI returned colors in content, using parsed result")
			return colors, nil
		}
	}

	return nil, fmt.Errorf("[ERROR] AI Tool Call Failed: no tool_calls and no parsable colors in content")
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

func parseColorsFromContent(content string) ([]string, bool) {
	var payload struct {
		Colors []string `json:"colors"`
	}
	if err := json.Unmarshal([]byte(content), &payload); err == nil {
		if len(payload.Colors) == 5 {
			return normalizeColors(payload.Colors)
		}
	}

	colors := extractColors(content)
	if len(colors) >= 5 {
		return colors[:5], true
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
			Name:        paletteToolName,
			Description: "Return exactly five HEX colors that satisfy the user's palette request.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"colors": map[string]interface{}{
						"type":        "array",
						"description": "List of five HEX colors in #RRGGBB format.",
						"items": map[string]interface{}{
							"type":    "string",
							"pattern": "^#[0-9A-Fa-f]{6}$",
						},
						"minItems": 5,
						"maxItems": 5,
					},
				},
				"required": []string{"colors"},
			},
		},
	}
}

func parseToolCallColors(call ToolCall) ([]string, error) {
	if strings.ToLower(call.Type) != "function" {
		return nil, fmt.Errorf("unexpected tool call type: %s", call.Type)
	}
	if call.Function.Name != paletteToolName {
		return nil, fmt.Errorf("unexpected tool call function: %s", call.Function.Name)
	}

	var payload struct {
		Colors []string `json:"colors"`
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

	return normalized, nil
}
