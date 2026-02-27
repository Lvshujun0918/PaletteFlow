package handler

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"regexp"
	"strings"
	"time"

	"ai-color-palette/ai"

	"github.com/gin-gonic/gin"
)

type ColorPaletteRequest struct {
	Prompt     string `json:"prompt" binding:"required"`
	ColorCount int    `json:"color_count"`
}

type SingleColorRequest struct {
	Prompt      string   `json:"prompt" binding:"required"`
	BaseColors  []string `json:"base_colors" binding:"required"`
	TargetIndex int      `json:"target_index" binding:"required"`
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

type InspirationResponse struct {
	Text string `json:"text"`
}

// GeneratePaletteHandler 使用AI生成配色方案，失败时降级到随机生成
func GeneratePaletteHandler(c *gin.Context) {
	var req ColorPaletteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 尝试使用AI生成配色
	log.Printf("[INFO] Using %s to create colors:\n", req.Prompt)
	colorCount := req.ColorCount
	if colorCount == 0 {
		colorCount = 5
	}
	if colorCount < 1 || colorCount > 10 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "color_count must be between 1 and 10"})
		return
	}

	normalizedPrompt := strings.TrimSpace(req.Prompt)
	if normalizedPrompt == "森林配色" {
		log.Printf("[INFO] Demo prompt hit, return preset palette directly: %s\n", req.Prompt)
		forestColors := []string{
			"#2D5016", "#4A7C59", "#8F9779", "#C8D5B9", "#F8F6F0",
			"#1F3B1A", "#5F8F6F", "#A7B89A", "#DCE6D1", "#EEF3E8",
		}
		response := ColorPaletteResponse{
			Colors:      forestColors[:colorCount],
			Advice:      "主色#2D5016用于标题或重点元素，次色#4A7C59和#8F9779用于正文与背景过渡，点缀色#C8D5B9用于按钮或图标，#F8F6F0作为底色提升可读性。适用于自然主题网站、环保品牌视觉或户外产品包装。",
			Timestamp:   time.Now().Unix(),
			Description: "森林主题演示配色（直返）",
		}
		c.JSON(http.StatusOK, response)
		return
	}

	if strings.Contains(req.Prompt, "烧鸡") {
		log.Printf("[INFO] Bingo~ %s\n", req.Prompt)
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
	result, err := ai.GenerateColorPalette(req.Prompt, colorCount)
	if err != nil {
		log.Printf("[ERROR] AI generation failed: %v, falling back to random generation", err)
		// 降级到随机生成
		rand.Seed(time.Now().UnixNano())
		result = &ai.PaletteResult{
			Colors: generateRandomColors(colorCount, req.Prompt),
			Advice: "由于网络原因，AI调用失败。本次为随机生成配色，可作为灵感草案使用。建议在主色与辅色之间调整明度对比以提升层次感。",
		}
	}

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
	var req SingleColorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(req.BaseColors) < 1 || len(req.BaseColors) > 10 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "base_colors must contain between 1 and 10 colors"})
		return
	}

	if req.TargetIndex < 0 || req.TargetIndex >= len(req.BaseColors) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "target_index out of range"})
		return
	}

	hexRe := regexp.MustCompile(`^#[0-9A-Fa-f]{6}$`)
	normalized := make([]string, 0, len(req.BaseColors))
	for _, color := range req.BaseColors {
		candidate := strings.ToUpper(strings.TrimSpace(color))
		if !hexRe.MatchString(candidate) {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("invalid color: %s", color)})
			return
		}
		normalized = append(normalized, candidate)
	}
	log.Printf("[INFO] Using %s to replace single color:\n", req.Prompt)
	result, err := ai.GeneratePaletteWithSingleColor(normalized, req.TargetIndex, req.Prompt)
	if err != nil {
		log.Printf("[ERROR] AI single color generation failed: %v, fallback to replace target only", err)
		rand.Seed(time.Now().UnixNano())
		replacement := fmt.Sprintf("#%06X", rand.Intn(0xFFFFFF))
		normalized[req.TargetIndex] = replacement
		result = &ai.PaletteResult{
			Colors: normalized,
			Advice: "AI 调用失败，已为指定位置生成备选颜色。建议再尝试一次以获得更佳效果。",
		}
	}

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
	var req RefinePaletteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	colorCount := req.ColorCount
	if colorCount == 0 {
		colorCount = len(req.CurrentColors)
	}
	if colorCount < 1 || colorCount > 10 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "color_count must be between 1 and 10"})
		return
	}

	result, err := ai.RefinePalette(req.CurrentColors, req.Prompt, colorCount)
	if err != nil {
		log.Printf("[ERROR] Refine palette failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to refine palette"})
		return
	}
	log.Printf("[INFO] Using %s to refine colors:\n", req.Prompt)

	response := ColorPaletteResponse{
		Colors:      result.Colors,
		Advice:      result.Advice,
		Timestamp:   time.Now().Unix(),
		Description: fmt.Sprintf("基于提示词 '%s' 调整的配色", req.Prompt),
	}

	c.JSON(http.StatusOK, response)
}

// GenerateInspirationHandler 生成灵感文案（约20字艺术短句）
func GenerateInspirationHandler(c *gin.Context) {
	text, err := ai.GenerateInspirationText()
	if err != nil {
		log.Printf("[ERROR] Inspiration generation failed: %v", err)
		fallback := []string{
			"暮色流金映窗影留白自成诗",
			"雾蓝轻覆旧梦光影在指尖醒",
			"晨风穿过画布暖灰悄然生花",
			"晚霞落入湖心静谧缓缓铺陈",
			"月白漫过砖墙余温仍在回响",
		}
		rand.Seed(time.Now().UnixNano())
		c.JSON(http.StatusOK, InspirationResponse{Text: fallback[rand.Intn(len(fallback))]})
		return
	}

	clean := strings.TrimSpace(text)
	runes := []rune(clean)
	if len(runes) > 20 {
		clean = string(runes[:20])
	}
	if clean == "" {
		clean = "暮色流金映窗影留白自成诗"
	}

	c.JSON(http.StatusOK, InspirationResponse{Text: clean})
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
