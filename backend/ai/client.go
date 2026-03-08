package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"ai-color-palette/config"
	"ai-color-palette/logging"
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
func GenerateColorPalette(prompt string, colorCount int, seedColors []string) (*PaletteResult, error) {
	logging.Info("ai.generate.start", "generate palette called", logging.Fields{"color_count": colorCount, "prompt": prompt, "seed_colors": seedColors})
	if colorCount < 1 || colorCount > 10 {
		logging.Warn("ai.generate.invalid_count", "colorCount out of range", logging.Fields{"color_count": colorCount})
		return nil, fmt.Errorf("colorCount must be between 1 and 10")
	}

	normalizedSeeds, err := normalizeSeedColors(seedColors, colorCount)
	if err != nil {
		logging.Warn("ai.generate.invalid_seeds", "seed colors invalid", logging.Fields{"error": err.Error(), "seed_colors": seedColors})
		return nil, err
	}

	systemPrompt := buildBaseSystemPrompt(colorCount)
	userPrompt := buildGenerateUserPrompt(prompt, normalizedSeeds)
	result, genErr := retryGeneratePalette(systemPrompt, userPrompt, colorCount)
	if genErr != nil {
		return nil, genErr
	}

	result.Colors = applySeedColorLocks(result.Colors, normalizedSeeds)
	return result, nil
}

// GeneratePaletteWithSingleColor 仅替换指定颜色，保持其他颜色不变
func GeneratePaletteWithSingleColor(baseColors []string, targetIndex int, prompt string) (*PaletteResult, error) {
	logging.Info("ai.single.start", "single color generation called", logging.Fields{"base_count": len(baseColors), "target_index": targetIndex, "prompt": prompt})
	normalized, ok := normalizeColorsInRange(baseColors, 1, 10)
	if !ok {
		logging.Warn("ai.single.invalid_base", "invalid base colors in single color generation", logging.Fields{"base_count": len(baseColors)})
		return nil, fmt.Errorf("base colors must contain 1 to 10 valid hex values")
	}
	if targetIndex < 0 || targetIndex >= len(normalized) {
		logging.Warn("ai.single.invalid_index", "target index out of range", logging.Fields{"target_index": targetIndex, "base_count": len(normalized)})
		return nil, fmt.Errorf("targetIndex out of range")
	}

	systemPrompt := buildSingleColorSystemPrompt()
	userPrompt := fmt.Sprintf(
		"现有配色（顺序固定）为：%s。仅替换第%d个颜色 %s，依据用户的新需求：%s，同时注意与其余颜色的协调性。保持其余颜色不变，返回新的完整%d色方案及使用建议。",
		strings.Join(normalized, ", "),
		targetIndex+1,
		normalized[targetIndex],
		prompt,
		len(normalized),
	)

	result, err := retryGeneratePalette(systemPrompt, userPrompt, len(normalized))
	if err != nil {
		return nil, err
	}

	// 强制只替换目标位置，其余颜色保持不变
	finalColors := make([]string, len(normalized))
	copy(finalColors, normalized)

	if len(result.Colors) == len(normalized) {
		finalColors[targetIndex] = result.Colors[targetIndex]
	} else {
		logging.Warn("ai.single.unexpected_count", "single color mode returned unexpected count", logging.Fields{"result_count": len(result.Colors), "expected_count": len(normalized)})
		if len(result.Colors) > targetIndex {
			finalColors[targetIndex] = result.Colors[targetIndex]
		}
	}

	result.Colors = finalColors
	logging.Info("ai.single.success", "single color generation completed", logging.Fields{"result_count": len(result.Colors), "target_index": targetIndex})
	return result, nil
}

// RefinePalette 基于现有配色方案进行微调，可指定目标数量
func RefinePalette(currentColors []string, prompt string, targetColorCount int) (*PaletteResult, error) {
	logging.Info("ai.refine.start", "refine palette called", logging.Fields{"current_count": len(currentColors), "target_count": targetColorCount, "prompt": prompt})
	normalized, ok := normalizeColorsInRange(currentColors, 1, 10)
	if !ok {
		logging.Warn("ai.refine.invalid_base", "invalid current colors in refine", logging.Fields{"current_count": len(currentColors)})
		return nil, fmt.Errorf("current colors must contain 1 to 10 valid hex values")
	}

	if targetColorCount < 1 || targetColorCount > 10 {
		logging.Warn("ai.refine.invalid_target_count", "targetColorCount out of range", logging.Fields{"target_count": targetColorCount})
		return nil, fmt.Errorf("targetColorCount must be between 1 and 10")
	}

	systemPrompt := buildBaseSystemPrompt(targetColorCount)
	originalCount := len(normalized)
	countInstruction := "请保留原有配色基调并进行微调。"
	if targetColorCount > originalCount {
		countInstruction = fmt.Sprintf("目标数量从%d增加到%d：请在保留原方案风格的前提下补充%d个新颜色，新颜色需与原有颜色和谐过渡并拉开层次。", originalCount, targetColorCount, targetColorCount-originalCount)
	} else if targetColorCount < originalCount {
		countInstruction = fmt.Sprintf("目标数量从%d减少到%d：请合并或去除冗余颜色，但保留核心风格与视觉层次。", originalCount, targetColorCount)
	}

	userPrompt := fmt.Sprintf(
		"现有配色为：%s。用户希望在此基础上进行调整：%s。%s 请根据用户的修改意见，生成一个新的%d色方案，必须要与原方案具有**较大的相似性**。如果不涉及具体颜色修改，请保持原有风格。返回新的完整%d色方案及使用建议。",
		strings.Join(normalized, ", "),
		prompt,
		countInstruction,
		targetColorCount,
		targetColorCount,
	)

	return retryGeneratePalette(systemPrompt, userPrompt, targetColorCount)
}

func buildGenerateUserPrompt(prompt string, seedColors []string) string {
	seedHints := make([]string, 0)
	for i, c := range seedColors {
		if strings.TrimSpace(c) == "" {
			continue
		}
		seedHints = append(seedHints, fmt.Sprintf("第%d位固定为%s", i+1, c))
	}

	if len(seedHints) == 0 {
		return fmt.Sprintf("请你帮我生成这样的配色：%s", prompt)
	}

	return fmt.Sprintf(
		"请你帮我生成这样的配色：%s。以下初始色为用户手动指定，必须在对应位置保持不变：%s。其余位置请围绕这些初始色生成协调方案，并重点参考这些初始色进行延展。",
		prompt,
		strings.Join(seedHints, "，"),
	)
}

func normalizeSeedColors(seedColors []string, colorCount int) ([]string, error) {
	normalized := make([]string, colorCount)
	if len(seedColors) == 0 {
		return normalized, nil
	}
	if len(seedColors) > colorCount {
		return nil, fmt.Errorf("seedColors length exceeds color count")
	}

	for i := 0; i < len(seedColors); i++ {
		candidate := strings.ToUpper(strings.TrimSpace(seedColors[i]))
		if candidate == "" {
			continue
		}
		if !strictHexRegex.MatchString(candidate) {
			return nil, fmt.Errorf("invalid seed color at index %d", i)
		}
		normalized[i] = candidate
	}

	return normalized, nil
}

func applySeedColorLocks(colors []string, seedColors []string) []string {
	if len(colors) == 0 || len(seedColors) == 0 {
		return colors
	}
	locked := make([]string, len(colors))
	copy(locked, colors)
	for i := 0; i < len(locked) && i < len(seedColors); i++ {
		if strings.TrimSpace(seedColors[i]) == "" {
			continue
		}
		locked[i] = strings.ToUpper(strings.TrimSpace(seedColors[i]))
	}
	return locked
}

func retryGeneratePalette(systemPrompt, userPrompt string, colorCount int) (*PaletteResult, error) {
	const maxRetries = 3
	var lastErr error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		logging.Info("ai.retry.attempt", "attempting to generate palette", logging.Fields{"attempt": attempt, "max_retries": maxRetries, "color_count": colorCount})

		result, err := attemptGenerateWithPrompt(systemPrompt, userPrompt, colorCount)
		if err == nil {
			logging.Info("ai.retry.success", "palette generation attempt succeeded", logging.Fields{"attempt": attempt, "max_retries": maxRetries, "color_count": len(result.Colors)})
			return result, nil
		}

		lastErr = err
		logging.Warn("ai.retry.failed", "palette generation attempt failed", logging.Fields{"attempt": attempt, "max_retries": maxRetries, "error": err.Error()})

		if attempt < maxRetries {
			time.Sleep(time.Second * time.Duration(attempt))
		}
	}

	logging.Error("ai.retry.exhausted", "failed to generate palette after all retries", logging.Fields{"max_retries": maxRetries, "error": lastErr.Error()})
	return nil, fmt.Errorf("all %d retry attempts failed, last error: %w", maxRetries, lastErr)
}

func attemptGenerateWithPrompt(systemPrompt, userPrompt string, colorCount int) (*PaletteResult, error) {
	cfg := config.AppConfig
	if cfg.AIAPIKey == "" {
		logging.Error("ai.request.config_missing", "AI API key not configured", nil)
		return nil, fmt.Errorf("AI API key not configured")
	}

	paletteTool := buildPaletteToolDefinition(colorCount)
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
		logging.Error("ai.request.marshal_failed", "failed to marshal AI request", logging.Fields{"error": err.Error()})
		return nil, fmt.Errorf("marshal request: %w", err)
	}
	logging.Info("ai.request.payload", "AI request payload prepared", logging.Fields{"payload_size": len(jsonData), "model": cfg.AIModel, "color_count": colorCount})
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.AITimeout)*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", cfg.AIAPIBaseURL+"/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		logging.Error("ai.request.build_failed", "failed to build AI request", logging.Fields{"error": err.Error()})
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+cfg.AIAPIKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logging.Error("ai.request.send_failed", "failed to send AI request", logging.Fields{"error": err.Error()})
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()
	logging.Info("ai.response.received", "AI response received", logging.Fields{"status": resp.StatusCode})

	bodyBytes, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		logging.Error("ai.response.read_failed", "failed to read AI response body", logging.Fields{"status": resp.StatusCode, "error": readErr.Error()})
		return nil, fmt.Errorf("read response body: %w", readErr)
	}
	bodyText := string(bodyBytes)
	logging.Info("ai.response.raw", "AI raw response body", logging.Fields{"status": resp.StatusCode, "body": bodyText})

	if resp.StatusCode != http.StatusOK {
		logging.Error("ai.response.status_error", "AI API returned non-200", logging.Fields{"status": resp.StatusCode, "body": bodyText})
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, bodyText)
	}

	var chatResp ChatResponse
	if err := json.Unmarshal(bodyBytes, &chatResp); err != nil {
		logging.Error("ai.response.decode_failed", "failed to decode AI response", logging.Fields{"error": err.Error(), "body": bodyText})
		return nil, fmt.Errorf("decode response: %w", err)
	}

	logging.Info("ai.response.parsed", "AI parsed response object", logging.Fields{"response": chatResp})

	if len(chatResp.Choices) == 0 {
		logging.Error("ai.response.empty_choices", "no choices in AI response", nil)
		return nil, fmt.Errorf("no response from AI")
	}

	choice := chatResp.Choices[0]
	message := choice.Message
	if message.Role != "assistant" {
		logging.Error("ai.response.invalid_role", "unexpected role in AI response", logging.Fields{"role": message.Role})
		return nil, fmt.Errorf("unexpected message role: %s", message.Role)
	}
	logging.Info("ai.response.message", "assistant message received", logging.Fields{"message": message, "tool_call_count": len(message.ToolCalls), "has_content": strings.TrimSpace(message.Content) != ""})
	if len(message.ToolCalls) > 0 {
		for _, call := range message.ToolCalls {
			logging.Info("ai.tool_call.raw", "received tool call", logging.Fields{"tool_call": call})
			if call.Function.Name != paletteToolName {
				continue
			}
			result, err := parseToolCallResult(call, colorCount)
			if err != nil {
				logging.Error("ai.tool_call.parse_failed", "failed to parse palette tool call", logging.Fields{"error": err.Error()})
				return nil, err
			}
			logging.Info("ai.tool_call.success", "palette tool call parsed", logging.Fields{"color_count": len(result.Colors)})
			return result, nil
		}
		logging.Error("ai.tool_call.missing_palette", "tool call exists but no palette tool call found", nil)
		return nil, fmt.Errorf("tool call returned without expected palette data")
	}

	if message.Content != "" {
		logging.Info("ai.content.raw", "assistant content returned", logging.Fields{"content": message.Content})
		result, ok := parseResultFromContent(message.Content, colorCount)
		if ok {
			logging.Info("ai.content.parse_success", "parsed palette from assistant content", logging.Fields{"color_count": len(result.Colors)})
			return result, nil
		}
		logging.Warn("ai.content.parse_failed", "assistant content exists but could not parse palette", nil)
	}

	logging.Error("ai.response.invalid_payload", "AI response missing usable palette data", nil)
	return nil, fmt.Errorf("AI Tool Call Failed: no tool_calls and no parsable result in content")
}

func buildBaseSystemPrompt(colorCount int) string {
	return fmt.Sprintf(`
你是一个专业的配色设计师。用户会给你一个配色需求描述，你需要返回%d个精确的HEX颜色代码，并给出配色使用建议。
你必须通过调用 return_palette 工具函数返回结果，不要输出任何自然语言文本。
1. 采用【渐变过渡技巧】，在冲突色之间创建中间色调缓冲层
2. 运用【色彩比例法则】：主色占60%%，次色占30%%，点缀色占10%%
3. 建立【色彩秩序】：通过明度阶梯（从20%%到80%%亮度）建立视觉节奏
4. 添加【中性调和剂】：适当加入平衡色
5. 最终效果需呈现【动态和谐】- 既有视觉冲击力，又保持整体统一性
`, colorCount)
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

func parseResultFromContent(content string, colorCount int) (*PaletteResult, bool) {
	var payload struct {
		Colors []string `json:"colors"`
		Advice string   `json:"advice"`
	}
	if err := json.Unmarshal([]byte(content), &payload); err == nil {
		if len(payload.Colors) == colorCount {
			normalized, ok := normalizeColorsInRange(payload.Colors, colorCount, colorCount)
			if ok {
				return &PaletteResult{Colors: normalized, Advice: strings.TrimSpace(payload.Advice)}, true
			}
		}
	}

	colors := extractColors(content)
	if len(colors) >= colorCount {
		return &PaletteResult{Colors: colors[:colorCount]}, true
	}
	return nil, false
}

func normalizeColorsInRange(colors []string, minCount, maxCount int) ([]string, bool) {
	if len(colors) < minCount || len(colors) > maxCount {
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

func buildPaletteToolDefinition(colorCount int) ToolDefinition {
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
						"description": fmt.Sprintf("包含且仅包含 %d 个 HEX 颜色字符串，格式必须为 #RRGGBB。", colorCount),
						"items": map[string]interface{}{
							"type":    "string",
							"pattern": "^#[0-9A-Fa-f]{6}$",
						},
						"minItems": colorCount,
						"maxItems": colorCount,
					},
					"advice": map[string]interface{}{
						"type":        "string",
						"description": "配色使用建议，给出2-3条可执行的使用方式或场景，不要分点，要包含colors中的HEX颜色字符串",
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

func parseToolCallResult(call ToolCall, colorCount int) (*PaletteResult, error) {
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

	if len(payload.Colors) != colorCount {
		return nil, fmt.Errorf("tool call returned %d colors, expected %d", len(payload.Colors), colorCount)
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
