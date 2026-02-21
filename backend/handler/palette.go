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
	Prompt string `json:"prompt" binding:"required"`
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
	if strings.Contains(req.Prompt, "烧鸡") {
		log.Printf("[INFO] Bingo~ %s\n", req.Prompt)
		colors := []string{"#000000", "#FFFFFF", "#1E3A5F", "#2D5B8A", "#E5E5E5"}
		response := ColorPaletteResponse{
			Colors:      colors,
			Advice:      "你找到了隐藏彩蛋~这是专属于作者烧鸡的配色方案，烧鸡yyds！",
			Timestamp:   time.Now().Unix(),
			Description: "你找到了隐藏彩蛋~这是专属于作者烧鸡的配色方案！",
		}
		c.JSON(http.StatusOK, response)
		return
	}
	result, err := ai.GenerateColorPalette(req.Prompt)
	if err != nil {
		log.Printf("[ERROR] AI generation failed: %v, falling back to random generation", err)
		// 降级到随机生成
		rand.Seed(time.Now().UnixNano())
		result = &ai.PaletteResult{
			Colors: generateRandomColors(5, req.Prompt),
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

	if len(req.BaseColors) != 5 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "base_colors must contain 5 colors"})
		return
	}

	if req.TargetIndex < 0 || req.TargetIndex >= len(req.BaseColors) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "target_index out of range"})
		return
	}

	hexRe := regexp.MustCompile(`^#[0-9A-Fa-f]{6}$`)
	normalized := make([]string, 0, 5)
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

	result, err := ai.RefinePalette(req.CurrentColors, req.Prompt)
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
