package handler

import (
	"fmt"
	"math/rand"
	"net/http"
	"regexp"
	"strings"
	"time"

	"ai-color-palette/ai"
	"ai-color-palette/logging"

	"github.com/gin-gonic/gin"
)

type ColorPaletteRequest struct {
	Prompt     string   `json:"prompt" binding:"required"`
	ColorCount int      `json:"color_count"`
	SeedColors []string `json:"seed_colors"`
}

type SingleColorRequest struct {
	Prompt      string   `json:"prompt" binding:"required"`
	BaseColors  []string `json:"base_colors" binding:"required"`
	TargetIndex int      `json:"target_index"`
}

type ColorPaletteResponse struct {
	Colors      []string `json:"colors"`
	Advice      string   `json:"advice"`
	Timestamp   int64    `json:"timestamp"`
	Description string   `json:"description"`
}

type RefinePaletteRequest struct {
	CurrentColors []string `json:"current_colors" binding:"required"`
	Prompt        string   `json:"prompt" binding:"required"`
	ColorCount    int      `json:"color_count"`
}

// GeneratePaletteHandler 使用AI生成配色方案，失败时降级到随机生成
func GeneratePaletteHandler(c *gin.Context) {
	requestID := logging.RequestIDFromGin(c)
	logging.Info("palette.generate.start", "generate palette request received", logging.Fields{"request_id": requestID})

	var req ColorPaletteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logging.Warn("palette.generate.invalid_request", "failed to bind generate request", logging.Fields{"request_id": requestID, "error": err.Error()})
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 尝试使用AI生成配色
	logging.Info("palette.generate.input", "generate palette input parsed", logging.Fields{"request_id": requestID, "prompt": req.Prompt, "color_count": req.ColorCount})
	colorCount := req.ColorCount
	if colorCount == 0 {
		colorCount = 5
	}
	if colorCount < 1 || colorCount > 10 {
		logging.Warn("palette.generate.invalid_count", "color_count out of range", logging.Fields{"request_id": requestID, "color_count": colorCount})
		c.JSON(http.StatusBadRequest, gin.H{"error": "color_count must be between 1 and 10"})
		return
	}

	seedColors, err := normalizeSeedColors(req.SeedColors, colorCount)
	if err != nil {
		logging.Warn("palette.generate.invalid_seed_colors", "seed colors invalid", logging.Fields{"request_id": requestID, "error": err.Error()})
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	logging.Info("palette.generate.seed_colors", "seed colors received", logging.Fields{"request_id": requestID, "seed_colors": seedColors})

	normalizedPrompt := strings.TrimSpace(req.Prompt)
	if normalizedPrompt == "森林配色" {
		logging.Info("palette.generate.demo_forest", "forest preset matched", logging.Fields{"request_id": requestID})
		forestColors := []string{
			"#2D5016", "#4A7C59", "#8F9779", "#C8D5B9", "#F8F6F0",
			"#1F3B1A", "#5F8F6F", "#A7B89A", "#DCE6D1", "#EEF3E8",
		}
		response := ColorPaletteResponse{
			Colors:      forestColors[:colorCount],
			Advice:      "主色#2D5016用于标题或重点元素，次色#4A7C59和#8F9779用于正文与背景过渡，点缀色#C8D5B9用于按钮或图标，#F8F6F0作为底色提升可读性。适用于自然主题网站、环保品牌视觉或户外产品包装。",
			Timestamp:   time.Now().Unix(),
			Description: "森林主题演示配色",
		}
		c.JSON(http.StatusOK, response)
		return
	}

	if strings.Contains(req.Prompt, "烧鸡") {
		logging.Info("palette.generate.easter_egg", "easter egg prompt matched", logging.Fields{"request_id": requestID})
		colors := []string{"#000000", "#FFFFFF", "#1E3A5F", "#2D5B8A", "#E5E5E5", "#F59E0B", "#EF4444", "#10B981", "#3B82F6", "#8B5CF6"}
		response := ColorPaletteResponse{
			Colors:      colors[:colorCount],
			Advice:      "你找到了隐藏彩蛋~这是专属于作者烧鸡的配色方案，烧鸡yyds！",
			Timestamp:   time.Now().Unix(),
			Description: "你找到了隐藏彩蛋~这是专属于作者烧鸡的配色方案！",
		}
		c.JSON(http.StatusOK, response)
		return
	}
	result, err := ai.GenerateColorPalette(req.Prompt, colorCount, seedColors)
	if err != nil {
		logging.Error("palette.generate.ai_failed", "ai generation failed, fallback random", logging.Fields{"request_id": requestID, "error": err.Error()})
		// 降级到随机生成
		rand.Seed(time.Now().UnixNano())
		result = &ai.PaletteResult{
			Colors: generateRandomColors(colorCount, req.Prompt),
			Advice: "由于网络原因，AI调用失败。本次为随机生成配色，可作为灵感草案使用。建议在主色与辅色之间调整明度对比以提升层次感。",
		}
	}

	result.Colors = applySeedColorLocks(result.Colors, seedColors)

	logging.Info("palette.generate.success", "generate palette completed", logging.Fields{"request_id": requestID, "color_count": len(result.Colors)})

	response := ColorPaletteResponse{
		Colors:      result.Colors,
		Advice:      result.Advice,
		Timestamp:   time.Now().Unix(),
		Description: fmt.Sprintf("根据提示词 '%s' 生成的配色方案", req.Prompt),
	}

	c.JSON(http.StatusOK, response)
}

// RegenerateSingleColorHandler 仅重新生成指定位置的颜色
func RegenerateSingleColorHandler(c *gin.Context) {
	requestID := logging.RequestIDFromGin(c)
	logging.Info("palette.single.start", "single color regenerate request received", logging.Fields{"request_id": requestID})

	var req SingleColorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logging.Warn("palette.single.invalid_request", "failed to bind single color request", logging.Fields{"request_id": requestID, "error": err.Error()})
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(req.BaseColors) < 1 || len(req.BaseColors) > 10 {
		logging.Warn("palette.single.invalid_base_colors", "base colors length out of range", logging.Fields{"request_id": requestID, "count": len(req.BaseColors)})
		c.JSON(http.StatusBadRequest, gin.H{"error": "base_colors must contain between 1 and 10 colors"})
		return
	}

	if req.TargetIndex < 0 || req.TargetIndex >= len(req.BaseColors) {
		logging.Warn("palette.single.invalid_target_index", "target index out of range", logging.Fields{"request_id": requestID, "target_index": req.TargetIndex, "count": len(req.BaseColors)})
		c.JSON(http.StatusBadRequest, gin.H{"error": "target_index out of range"})
		return
	}

	hexRe := regexp.MustCompile(`^#[0-9A-Fa-f]{6}$`)
	normalized := make([]string, 0, len(req.BaseColors))
	for _, color := range req.BaseColors {
		candidate := strings.ToUpper(strings.TrimSpace(color))
		if !hexRe.MatchString(candidate) {
			logging.Warn("palette.single.invalid_hex", "invalid hex in base colors", logging.Fields{"request_id": requestID, "color": color})
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("invalid color: %s", color)})
			return
		}
		normalized = append(normalized, candidate)
	}

	normalizedPrompt := strings.TrimSpace(req.Prompt)
	if normalizedPrompt == "改成蓝色" {
		logging.Info("palette.single.demo", "single-color demo prompt matched", logging.Fields{"request_id": requestID})
		resultColors := make([]string, len(normalized))
		copy(resultColors, normalized)
		resultColors[req.TargetIndex] = "#2563EB"

		response := ColorPaletteResponse{
			Colors:      resultColors,
			Advice:      "已将目标颜色替换为蓝色 #2563EB，并保持其余颜色不变。建议在标题、链接或按钮处使用该蓝色作为视觉锚点。",
			Timestamp:   time.Now().Unix(),
			Description: fmt.Sprintf("针对第%d个颜色的定向微调（演示直返）", req.TargetIndex+1),
		}
		c.JSON(http.StatusOK, response)
		return
	}

	logging.Info("palette.single.input", "single color regenerate input parsed", logging.Fields{"request_id": requestID, "prompt": req.Prompt, "target_index": req.TargetIndex})
	result, err := ai.GeneratePaletteWithSingleColor(normalized, req.TargetIndex, req.Prompt)
	if err != nil {
		logging.Error("palette.single.ai_failed", "single color ai generation failed, fallback random target", logging.Fields{"request_id": requestID, "error": err.Error()})
		rand.Seed(time.Now().UnixNano())
		replacement := fmt.Sprintf("#%06X", rand.Intn(0xFFFFFF))
		normalized[req.TargetIndex] = replacement
		result = &ai.PaletteResult{
			Colors: normalized,
			Advice: "AI 调用失败，已为指定位置生成备选颜色。建议再尝试一次以获得更佳效果。",
		}
	}

	logging.Info("palette.single.success", "single color regenerate completed", logging.Fields{"request_id": requestID, "target_index": req.TargetIndex})

	// 再次确保只有目标位置被替换
	if len(result.Colors) == len(normalized) {
		keep := make([]string, len(normalized))
		copy(keep, normalized)
		keep[req.TargetIndex] = result.Colors[req.TargetIndex]
		result.Colors = keep
	}

	response := ColorPaletteResponse{
		Colors:      result.Colors,
		Advice:      result.Advice,
		Timestamp:   time.Now().Unix(),
		Description: fmt.Sprintf("针对第%d个颜色的定向微调", req.TargetIndex+1),
	}

	c.JSON(http.StatusOK, response)
}

// RefinePaletteHandler 基于现有配色方案进行微调
func RefinePaletteHandler(c *gin.Context) {
	requestID := logging.RequestIDFromGin(c)
	logging.Info("palette.refine.start", "refine palette request received", logging.Fields{"request_id": requestID})

	var req RefinePaletteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logging.Warn("palette.refine.invalid_request", "failed to bind refine request", logging.Fields{"request_id": requestID, "error": err.Error()})
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	colorCount := req.ColorCount
	if colorCount == 0 {
		colorCount = len(req.CurrentColors)
	}
	if colorCount < 1 || colorCount > 10 {
		logging.Warn("palette.refine.invalid_count", "color_count out of range", logging.Fields{"request_id": requestID, "color_count": colorCount})
		c.JSON(http.StatusBadRequest, gin.H{"error": "color_count must be between 1 and 10"})
		return
	}

	result, err := ai.RefinePalette(req.CurrentColors, req.Prompt, colorCount)
	if err != nil {
		logging.Error("palette.refine.failed", "refine palette failed", logging.Fields{"request_id": requestID, "error": err.Error()})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to refine palette"})
		return
	}
	logging.Info("palette.refine.success", "refine palette completed", logging.Fields{"request_id": requestID, "prompt": req.Prompt, "color_count": len(result.Colors)})

	response := ColorPaletteResponse{
		Colors:      result.Colors,
		Advice:      result.Advice,
		Timestamp:   time.Now().Unix(),
		Description: fmt.Sprintf("基于提示词 '%s' 调整的配色", req.Prompt),
	}

	c.JSON(http.StatusOK, response)
}

// 生成随机配色
func generateRandomColors(count int, seed string) []string {
	colors := []string{}
	rand.Seed(int64(len(seed)))

	for i := 0; i < count; i++ {
		// 生成伪随机颜色
		color := fmt.Sprintf("#%06X", rand.Intn(0xFFFFFF))
		colors = append(colors, color)
	}

	return colors
}

func normalizeSeedColors(seedColors []string, colorCount int) ([]string, error) {
	if colorCount < 1 || colorCount > 10 {
		return nil, fmt.Errorf("color_count must be between 1 and 10")
	}

	normalized := make([]string, colorCount)
	if len(seedColors) == 0 {
		return normalized, nil
	}
	if len(seedColors) > colorCount {
		return nil, fmt.Errorf("seed_colors must not exceed color_count")
	}

	hexRe := regexp.MustCompile(`^#[0-9A-Fa-f]{6}$`)
	for i := 0; i < len(seedColors); i++ {
		candidate := strings.TrimSpace(seedColors[i])
		if candidate == "" {
			continue
		}
		candidate = strings.ToUpper(candidate)
		if !hexRe.MatchString(candidate) {
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
