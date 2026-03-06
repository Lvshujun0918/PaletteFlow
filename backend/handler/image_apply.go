package handler

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	_ "image/jpeg"
	_ "image/png"
)

// ApplyImagePaletteHandler 将上传图片映射到指定配色（非 AI）
// form-data:
// - image: 图片文件（必填）
// - colors: 逗号分隔 HEX 颜色（必填），如: #1F2937,#3B82F6,#F59E0B
func ApplyImagePaletteHandler(c *gin.Context) {
	fileHeader, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing image file"})
		return
	}

	paletteInput := strings.TrimSpace(c.PostForm("colors"))
	if paletteInput == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing colors"})
		return
	}

	palette, err := parsePaletteFromForm(paletteInput)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to open image"})
		return
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported or invalid image"})
		return
	}

	processed := mapImageToPalette(img, palette)

	var buf bytes.Buffer
	if err := png.Encode(&buf, processed); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to encode image"})
		return
	}

	base := strings.TrimSuffix(fileHeader.Filename, filepath.Ext(fileHeader.Filename))
	if base == "" {
		base = "image"
	}
	outName := fmt.Sprintf("%s_palette.png", base)

	c.Header("Content-Type", "image/png")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%q", outName))
	c.Data(http.StatusOK, "image/png", buf.Bytes())
}

func parsePaletteFromForm(input string) ([]color.NRGBA, error) {
	parts := strings.FieldsFunc(input, func(r rune) bool {
		return r == ',' || r == ';' || r == '\n' || r == '\r' || r == '\t' || r == ' '
	})

	if len(parts) == 0 {
		return nil, fmt.Errorf("colors is empty")
	}
	if len(parts) > 20 {
		return nil, fmt.Errorf("too many colors, max 20")
	}

	palette := make([]color.NRGBA, 0, len(parts))
	for _, p := range parts {
		hex := strings.TrimSpace(p)
		if hex == "" {
			continue
		}
		c, err := parseHexColor(hex)
		if err != nil {
			return nil, fmt.Errorf("invalid color %q", hex)
		}
		palette = append(palette, c)
	}

	if len(palette) == 0 {
		return nil, fmt.Errorf("no valid colors provided")
	}

	return palette, nil
}

func parseHexColor(hex string) (color.NRGBA, error) {
	h := strings.TrimPrefix(strings.TrimSpace(hex), "#")
	if len(h) != 6 {
		return color.NRGBA{}, fmt.Errorf("hex must be 6 chars")
	}

	v, err := strconv.ParseUint(h, 16, 32)
	if err != nil {
		return color.NRGBA{}, err
	}

	return color.NRGBA{
		R: uint8((v >> 16) & 0xFF),
		G: uint8((v >> 8) & 0xFF),
		B: uint8(v & 0xFF),
		A: 255,
	}, nil
}

func mapImageToPalette(src image.Image, palette []color.NRGBA) *image.NRGBA {
	bounds := src.Bounds()
	dst := image.NewNRGBA(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			px := color.NRGBAModel.Convert(src.At(x, y)).(color.NRGBA)
			nearest := nearestPaletteColor(px, palette)
			nearest.A = px.A
			dst.SetNRGBA(x, y, nearest)
		}
	}

	return dst
}

func nearestPaletteColor(px color.NRGBA, palette []color.NRGBA) color.NRGBA {
	best := palette[0]
	bestDist := colorDistanceSquared(px, best)

	for i := 1; i < len(palette); i++ {
		candidate := palette[i]
		dist := colorDistanceSquared(px, candidate)
		if dist < bestDist {
			bestDist = dist
			best = candidate
		}
	}

	return best
}

func colorDistanceSquared(a, b color.NRGBA) int {
	dr := int(a.R) - int(b.R)
	dg := int(a.G) - int(b.G)
	db := int(a.B) - int(b.B)
	return dr*dr + dg*dg + db*db
}
