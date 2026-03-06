package handler

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"math"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	_ "image/jpeg"
	_ "image/png"
)

const (
	imageModeLAB          = "lab"
	imageModePreserveLuma = "preserve_luma"
	imageModeSoftBlend    = "soft_blend"
)

type paletteEntry struct {
	Color color.NRGBA
	L     float64
	A     float64
	B     float64
}

type imageApplyTask struct {
	ID          string
	Status      string
	Progress    int
	Mode        string
	Filename    string
	Error       string
	ResultBytes []byte
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type imageApplyTaskResponse struct {
	TaskID    string `json:"task_id"`
	Status    string `json:"status"`
	Progress  int    `json:"progress"`
	Mode      string `json:"mode"`
	Filename  string `json:"filename,omitempty"`
	Error     string `json:"error,omitempty"`
	CreatedAt int64  `json:"created_at"`
	UpdatedAt int64  `json:"updated_at"`
}

var (
	imageTaskStoreMu sync.RWMutex
	imageTaskStore   = map[string]*imageApplyTask{}
)

// ApplyImagePaletteHandler 创建图片套色任务（异步）
// form-data:
// - image: 图片文件（必填）
// - colors: 逗号分隔 HEX 颜色（必填），如: #1F2937,#3B82F6,#F59E0B
// - mode: 映射模式（可选）: lab | preserve_luma | soft_blend，默认 preserve_luma
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

	mode := normalizeImageMode(c.PostForm("mode"))

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

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read image"})
		return
	}

	task := createImageTask(fileHeader.Filename, mode)
	go processImageTask(task.ID, fileBytes, palette, mode)

	c.JSON(http.StatusAccepted, toTaskResponse(task))
}

// GetImagePaletteTaskHandler 查询图片套色任务进度
func GetImagePaletteTaskHandler(c *gin.Context) {
	taskID := strings.TrimSpace(c.Param("taskId"))
	task, ok := getImageTask(taskID)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
		return
	}

	c.JSON(http.StatusOK, toTaskResponse(task))
}

// DownloadImagePaletteTaskResultHandler 下载图片套色任务结果
func DownloadImagePaletteTaskResultHandler(c *gin.Context) {
	taskID := strings.TrimSpace(c.Param("taskId"))
	task, ok := getImageTask(taskID)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
		return
	}

	if task.Status == "failed" {
		c.JSON(http.StatusBadRequest, gin.H{"error": task.Error})
		return
	}

	if task.Status != "completed" || len(task.ResultBytes) == 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "task not completed"})
		return
	}

	filename := task.Filename
	if filename == "" {
		filename = "image_palette.png"
	}

	c.Header("Content-Type", "image/png")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
	c.Data(http.StatusOK, "image/png", task.ResultBytes)
}

func processImageTask(taskID string, fileBytes []byte, palette []color.NRGBA, mode string) {
	setTaskProgress(taskID, 5, "processing", "正在解析图片")

	img, _, err := image.Decode(bytes.NewReader(fileBytes))
	if err != nil {
		setTaskFailed(taskID, "unsupported or invalid image")
		return
	}

	entries := buildPaletteEntries(palette)
	setTaskProgress(taskID, 15, "processing", "正在应用配色")
	processed := mapImageToPalette(img, entries, mode, func(doneRows, totalRows int) {
		if totalRows <= 0 {
			return
		}
		ratio := float64(doneRows) / float64(totalRows)
		progress := 15 + int(math.Round(ratio*75))
		if progress > 90 {
			progress = 90
		}
		setTaskProgress(taskID, progress, "processing", "正在应用配色")
	})

	setTaskProgress(taskID, 95, "processing", "正在生成结果")
	var out bytes.Buffer
	if err := png.Encode(&out, processed); err != nil {
		setTaskFailed(taskID, "failed to encode image")
		return
	}

	setTaskCompleted(taskID, out.Bytes())
}

func normalizeImageMode(mode string) string {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case imageModeLAB:
		return imageModeLAB
	case imageModeSoftBlend:
		return imageModeSoftBlend
	case imageModePreserveLuma, "":
		return imageModePreserveLuma
	default:
		return imageModePreserveLuma
	}
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

	return color.NRGBA{R: uint8((v >> 16) & 0xFF), G: uint8((v >> 8) & 0xFF), B: uint8(v & 0xFF), A: 255}, nil
}

func buildPaletteEntries(palette []color.NRGBA) []paletteEntry {
	entries := make([]paletteEntry, 0, len(palette))
	for _, c := range palette {
		l, a, b := rgbToLab(c)
		entries = append(entries, paletteEntry{Color: c, L: l, A: a, B: b})
	}
	return entries
}

func mapImageToPalette(src image.Image, entries []paletteEntry, mode string, onProgress func(doneRows, totalRows int)) *image.NRGBA {
	bounds := src.Bounds()
	dst := image.NewNRGBA(bounds)
	totalRows := bounds.Max.Y - bounds.Min.Y

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			px := color.NRGBAModel.Convert(src.At(x, y)).(color.NRGBA)

			var mapped color.NRGBA
			switch mode {
			case imageModeLAB:
				mapped = nearestPaletteColorLAB(px, entries)
			case imageModeSoftBlend:
				mapped = softBlendPaletteColor(px, entries)
			case imageModePreserveLuma:
				fallthrough
			default:
				mapped = preserveLumaPaletteColor(px, entries)
			}

			mapped.A = px.A
			dst.SetNRGBA(x, y, mapped)
		}

		if onProgress != nil {
			doneRows := (y - bounds.Min.Y) + 1
			if doneRows == totalRows || doneRows%8 == 0 {
				onProgress(doneRows, totalRows)
			}
		}
	}

	return dst
}

func nearestPaletteColorLAB(px color.NRGBA, entries []paletteEntry) color.NRGBA {
	l, a, b := rgbToLab(px)
	best := entries[0]
	bestDist := labDistanceSquared(l, a, b, best.L, best.A, best.B)

	for i := 1; i < len(entries); i++ {
		candidate := entries[i]
		dist := labDistanceSquared(l, a, b, candidate.L, candidate.A, candidate.B)
		if dist < bestDist {
			bestDist = dist
			best = candidate
		}
	}
	return best.Color
}

func preserveLumaPaletteColor(px color.NRGBA, entries []paletteEntry) color.NRGBA {
	_, _, srcLum := rgbToHSL(px)
	_, srcS, _ := rgbToHSL(px)
	nearest := nearestPaletteColorLAB(px, entries)
	nearestH, nearestS, _ := rgbToHSL(nearest)
	mixedS := clamp01(nearestS*0.75 + srcS*0.25)
	return hslToRGB(nearestH, mixedS, srcLum)
}

func softBlendPaletteColor(px color.NRGBA, entries []paletteEntry) color.NRGBA {
	l, a, b := rgbToLab(px)
	bestIdx := 0
	secondIdx := 0
	bestDist := math.MaxFloat64
	secondDist := math.MaxFloat64

	for i := 0; i < len(entries); i++ {
		d := labDistanceSquared(l, a, b, entries[i].L, entries[i].A, entries[i].B)
		if d < bestDist {
			secondDist = bestDist
			secondIdx = bestIdx
			bestDist = d
			bestIdx = i
		} else if d < secondDist {
			secondDist = d
			secondIdx = i
		}
	}

	first := entries[bestIdx].Color
	second := entries[secondIdx].Color
	if secondDist == math.MaxFloat64 || bestIdx == secondIdx {
		return first
	}

	w1 := 1.0 / (bestDist + 1e-6)
	w2 := 1.0 / (secondDist + 1e-6)
	sum := w1 + w2
	if sum <= 0 {
		return first
	}

	r := (float64(first.R)*w1 + float64(second.R)*w2) / sum
	g := (float64(first.G)*w1 + float64(second.G)*w2) / sum
	bv := (float64(first.B)*w1 + float64(second.B)*w2) / sum
	return color.NRGBA{R: uint8(clamp255(r)), G: uint8(clamp255(g)), B: uint8(clamp255(bv)), A: 255}
}

func rgbToLab(c color.NRGBA) (float64, float64, float64) {
	r := srgbToLinear(float64(c.R) / 255.0)
	g := srgbToLinear(float64(c.G) / 255.0)
	b := srgbToLinear(float64(c.B) / 255.0)
	x := r*0.4124564 + g*0.3575761 + b*0.1804375
	y := r*0.2126729 + g*0.7151522 + b*0.0721750
	z := r*0.0193339 + g*0.1191920 + b*0.9503041
	fx := xyzToLabF(x / 0.95047)
	fy := xyzToLabF(y / 1.00000)
	fz := xyzToLabF(z / 1.08883)
	l := 116.0*fy - 16.0
	a := 500.0 * (fx - fy)
	bv := 200.0 * (fy - fz)
	return l, a, bv
}

func srgbToLinear(v float64) float64 {
	if v <= 0.04045 {
		return v / 12.92
	}
	return math.Pow((v+0.055)/1.055, 2.4)
}

func xyzToLabF(t float64) float64 {
	if t > 0.008856 {
		return math.Cbrt(t)
	}
	return 7.787*t + 16.0/116.0
}

func rgbToHSL(c color.NRGBA) (float64, float64, float64) {
	r := float64(c.R) / 255.0
	g := float64(c.G) / 255.0
	b := float64(c.B) / 255.0
	maxV := math.Max(r, math.Max(g, b))
	minV := math.Min(r, math.Min(g, b))
	delta := maxV - minV
	l := (maxV + minV) / 2.0
	if delta == 0 {
		return 0, 0, l
	}
	var h float64
	s := delta / (1 - math.Abs(2*l-1))
	switch maxV {
	case r:
		h = math.Mod((g-b)/delta, 6)
	case g:
		h = (b-r)/delta + 2
	default:
		h = (r-g)/delta + 4
	}
	h *= 60
	if h < 0 {
		h += 360
	}
	return h, clamp01(s), clamp01(l)
}

func hslToRGB(h, s, l float64) color.NRGBA {
	h = math.Mod(h, 360)
	if h < 0 {
		h += 360
	}
	c := (1 - math.Abs(2*l-1)) * s
	x := c * (1 - math.Abs(math.Mod(h/60.0, 2)-1))
	m := l - c/2
	var r1, g1, b1 float64
	switch {
	case h < 60:
		r1, g1, b1 = c, x, 0
	case h < 120:
		r1, g1, b1 = x, c, 0
	case h < 180:
		r1, g1, b1 = 0, c, x
	case h < 240:
		r1, g1, b1 = 0, x, c
	case h < 300:
		r1, g1, b1 = x, 0, c
	default:
		r1, g1, b1 = c, 0, x
	}
	r := (r1 + m) * 255.0
	g := (g1 + m) * 255.0
	b := (b1 + m) * 255.0
	return color.NRGBA{R: uint8(clamp255(r)), G: uint8(clamp255(g)), B: uint8(clamp255(b)), A: 255}
}

func labDistanceSquared(l1, a1, b1, l2, a2, b2 float64) float64 {
	dl := l1 - l2
	da := a1 - a2
	db := b1 - b2
	return dl*dl + da*da + db*db
}

func clamp255(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return math.Round(v)
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

func createImageTask(filename, mode string) *imageApplyTask {
	base := strings.TrimSuffix(filename, filepath.Ext(filename))
	if base == "" {
		base = "image"
	}
	now := time.Now()
	task := &imageApplyTask{
		ID:        newTaskID(),
		Status:    "queued",
		Progress:  0,
		Mode:      mode,
		Filename:  fmt.Sprintf("%s_palette.png", base),
		CreatedAt: now,
		UpdatedAt: now,
	}
	imageTaskStoreMu.Lock()
	imageTaskStore[task.ID] = task
	imageTaskStoreMu.Unlock()
	return cloneTask(task)
}

func getImageTask(taskID string) (*imageApplyTask, bool) {
	imageTaskStoreMu.RLock()
	task, ok := imageTaskStore[taskID]
	imageTaskStoreMu.RUnlock()
	if !ok {
		return nil, false
	}
	return cloneTask(task), true
}

func setTaskProgress(taskID string, progress int, status, _ string) {
	if progress < 0 {
		progress = 0
	}
	if progress > 100 {
		progress = 100
	}
	imageTaskStoreMu.Lock()
	if task, ok := imageTaskStore[taskID]; ok {
		if progress > task.Progress {
			task.Progress = progress
		}
		task.Status = status
		task.UpdatedAt = time.Now()
	}
	imageTaskStoreMu.Unlock()
}

func setTaskFailed(taskID, errMsg string) {
	imageTaskStoreMu.Lock()
	if task, ok := imageTaskStore[taskID]; ok {
		task.Status = "failed"
		task.Progress = 100
		task.Error = errMsg
		task.UpdatedAt = time.Now()
	}
	imageTaskStoreMu.Unlock()
}

func setTaskCompleted(taskID string, result []byte) {
	imageTaskStoreMu.Lock()
	if task, ok := imageTaskStore[taskID]; ok {
		task.Status = "completed"
		task.Progress = 100
		task.ResultBytes = result
		task.UpdatedAt = time.Now()
	}
	imageTaskStoreMu.Unlock()
}

func toTaskResponse(task *imageApplyTask) imageApplyTaskResponse {
	return imageApplyTaskResponse{
		TaskID:    task.ID,
		Status:    task.Status,
		Progress:  task.Progress,
		Mode:      task.Mode,
		Filename:  task.Filename,
		Error:     task.Error,
		CreatedAt: task.CreatedAt.Unix(),
		UpdatedAt: task.UpdatedAt.Unix(),
	}
}

func cloneTask(task *imageApplyTask) *imageApplyTask {
	if task == nil {
		return nil
	}
	copied := *task
	if task.ResultBytes != nil {
		copied.ResultBytes = make([]byte, len(task.ResultBytes))
		copy(copied.ResultBytes, task.ResultBytes)
	}
	return &copied
}

func newTaskID() string {
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		return fmt.Sprintf("task_%d", time.Now().UnixNano())
	}
	return fmt.Sprintf("task_%d_%s", time.Now().UnixNano(), hex.EncodeToString(buf))
}
