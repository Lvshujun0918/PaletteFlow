<template>
  <div class="color-display">
    <!-- 配色卡片网格 -->
    <div class="color-cards">
      <div
        v-for="(color, index) in colors"
        :key="index"
        class="color-card glass-card"
        :class="{ 'is-highlighted': isHighlightedColor(color) }"
      >
        <div class="color-preview" :style="{ backgroundColor: color }"></div>
        <div class="color-info">
          <div class="color-code">{{ color }}</div>
          <div class="color-actions">
            <Tooltip text="复制颜色值" position="top">
              <button class="copy-btn" @click.stop="copyToClipboard(color)"><svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24"><g fill="none"><path fill="#333333" d="m12.593 23.258l-.011.002l-.071.035l-.02.004l-.014-.004l-.071-.035q-.016-.005-.024.005l-.004.01l-.017.428l.005.02l.01.013l.104.074l.015.004l.012-.004l.104-.074l.012-.016l.004-.017l-.017-.427q-.004-.016-.017-.018m.265-.113l-.013.002l-.185.093l-.01.01l-.003.011l.018.43l.005.012l.008.007l.201.093q.019.005.029-.008l.004-.014l-.034-.614q-.005-.018-.02-.022m-.715.002a.02.02 0 0 0-.027.006l-.006.014l-.034.614q.001.018.017.024l.015-.002l.201-.093l.01-.008l.004-.011l.017-.43l-.003-.012l-.01-.01z"/><path fill="#333333" d="M19 2a2 2 0 0 1 2 2v12a2 2 0 0 1-2 2h-2v2a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h2V4a2 2 0 0 1 2-2zm-9 13H8a1 1 0 0 0-.117 1.993L8 17h2a1 1 0 0 0 .117-1.993zm9-11H9v2h6a2 2 0 0 1 2 2v8h2zm-7 7H8a1 1 0 1 0 0 2h4a1 1 0 1 0 0-2"/></g></svg></button>
            </Tooltip>
            <Tooltip text="手动调节该颜色" position="top">
              <button class="edit-btn" @click.stop="emitPickColor(index)"><svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24"><path fill="#333333" d="M5 19h1.4l8.625-8.625l-1.4-1.4L5 17.6zm-1 2q-.425 0-.712-.288T3 20v-2.825q0-.2.075-.388t.225-.337l10.3-10.3q.3-.3.675-.45t.75-.15q.4 0 .763.15t.662.45L17.925 8.6q.275.3.425.663T18.5 10q0 .375-.137.738t-.438.662l-10.3 10.3q-.15.15-.337.225T6.825 22zM14.325 9.675l-.7-.7l1.4 1.4z"/></svg></button>
            </Tooltip>
            <Tooltip text="AI调节该颜色" position="top">
              <button class="pick-btn" @click.stop="emitSelectColor(index)"><svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24"><path fill="#333333" d="M20 2H8a2 2 0 0 0-2 2v12a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V4a2 2 0 0 0-2-2m-6.933 12.481l-3.274-3.274l1.414-1.414l1.726 1.726l4.299-5.159l1.537 1.281z"/><path fill="#333333" d="M4 22h11v-2H4V8H2v12c0 1.103.897 2 2 2"/></svg></button>
            </Tooltip>
          </div>
        </div>
        <!-- 对比差值显示 -->
        <div v-if="showComparison && previousColors.length > index" class="color-diff">
          <span class="diff-label">相对上组:</span>
          <span class="diff-value" :class="getDiffClass(getHSLDiff(color, previousColors[index]).dH)">
            H{{ formatDiff(getHSLDiff(color, previousColors[index]).dH) }}°
          </span>
          <span class="diff-value" :class="getDiffClass(getHSLDiff(color, previousColors[index]).dS)">
            S{{ formatDiff(getHSLDiff(color, previousColors[index]).dS) }}%
          </span>
          <span class="diff-value" :class="getDiffClass(getHSLDiff(color, previousColors[index]).dL)">
            L{{ formatDiff(getHSLDiff(color, previousColors[index]).dL) }}%
          </span>
        </div>
      </div>
    </div>

    <!-- 配色信息 -->
    <div class="palette-info glass-card">
      <div class="info-item">
        <span class="label">提示词:</span>
        <span class="value">{{ prompt || "还没有生成~"}}</span>
      </div>
      <div class="info-item">
        <span class="label">生成时间:</span>
        <span class="value">{{ formatTime(timestamp) }}</span>
      </div>
      <div class="info-item">
        <span class="label">颜色数量:</span>
        <span class="value">{{ colors.length }} 个</span>
      </div>
      <div class="info-item advice-item">
        <span class="label">使用建议:</span>
        <AdviceText
          class="value advice-text"
          :text="advice || '暂无建议'"
          @hover-color="emitHoverColor"
        />
      </div>
    </div>

    <!-- 快速操作 -->
    <div class="quick-actions">
      <GlassButton class="action-btn" @click="exportAsCSS">导出CSS</GlassButton>
      <GlassButton class="action-btn" @click="exportAsJSON">导出JSON</GlassButton>
      <GlassButton class="action-btn" @click="exportAsImage">导出图片</GlassButton>
    </div>
  </div>
</template>

<script>
import { notify } from '../utils/notify'
import { getHSLDifference } from '../utils/colorUtils'
import GlassButton from './GlassButton.vue'
import Tooltip from './Tooltip.vue'
import AdviceText from './AdviceText.vue'
import { computed } from 'vue'

export default {
  name: 'ColorDisplay',
  components: {
    GlassButton,
    Tooltip,
    AdviceText
  },
  emits: ['regenerate', 'pick-color', 'select-color', 'hover-color'],
  props: {
    colors: {
      type: Array,
      default: () => []
    },
    previousColors: {
      type: Array,
      default: () => []
    },
    prompt: {
      type: String,
      default: '默认配色'
    },
    timestamp: {
      type: Number,
      default: 0
    },
    advice: {
      type: String,
      default: ''
    },
    highlightedColor: {
      type: String,
      default: ''
    }
  },
  setup(props) {
    const showComparison = computed(() => {
      return props.previousColors && props.previousColors.length > 0
    })

    const getHSLDiff = (color1, color2) => {
      return getHSLDifference(color2, color1)
    }

    const formatDiff = (value) => {
      return value > 0 ? `+${value}` : `${value}`
    }

    const getDiffClass = (value) => {
      if (value > 0) return 'diff-positive'
      if (value < 0) return 'diff-negative'
      return 'diff-zero'
    }

    return {
      showComparison,
      getHSLDiff,
      formatDiff,
      getDiffClass
    }
  },
  methods: {
        emitPickColor(index) {
          this.$emit('pick-color', index)
        },
        emitSelectColor(index) {
          this.$emit('select-color', index)
        },
        emitHoverColor(color) {
          this.$emit('hover-color', color)
        },
        normalizeColor(color) {
          return typeof color === 'string' ? color.trim().toLowerCase() : ''
        },
        isHighlightedColor(color) {
          if (!this.highlightedColor) return false
          return this.normalizeColor(color) === this.normalizeColor(this.highlightedColor)
        },
    copyToClipboard(color) {
      navigator.clipboard.writeText(color).then(() => {
        notify(`已复制: ${color}`, 'success')
      }).catch(() => {
        notify('复制失败', 'error')
      })
    },
    formatTime(timestamp) {
      if (!timestamp) return '未知'
      const date = new Date(timestamp)
      return date.toLocaleString('zh-CN')
    },
    exportAsCSS() {
      let css = ':root {\n'
      this.colors.forEach((color, index) => {
        css += `  --color-${index + 1}: ${color};\n`
      })
      css += '}'

      this.downloadFile(css, 'colors.css', 'text/css')
    },
    exportAsJSON() {
      const data = {
        prompt: this.prompt,
        colors: this.colors,
        timestamp: this.timestamp,
        advice: this.advice
      }

      this.downloadFile(
        JSON.stringify(data, null, 2),
        'colors.json',
        'application/json'
      )
    },
    exportAsImage() {
      const canvas = document.createElement('canvas')
      const boxWidth = 100
      const boxHeight = 80
      const padding = 10

      canvas.width = boxWidth * this.colors.length + padding * (this.colors.length + 1)
      canvas.height = boxHeight + padding * 2

      const ctx = canvas.getContext('2d')
      ctx.fillStyle = '#fff'
      ctx.fillRect(0, 0, canvas.width, canvas.height)

      this.colors.forEach((color, index) => {
        const x = padding + index * (boxWidth + padding)
        const y = padding

        ctx.fillStyle = color
        ctx.fillRect(x, y, boxWidth, boxHeight)

        ctx.fillStyle = '#000'
        ctx.font = '12px Arial'
        ctx.textAlign = 'center'
        ctx.fillText(color, x + boxWidth / 2, y + boxHeight + 20)
      })

      canvas.toBlob((blob) => {
        const url = URL.createObjectURL(blob)
        const a = document.createElement('a')
        a.href = url
        a.download = 'colors.png'
        a.click()
        URL.revokeObjectURL(url)
      })
    },
    downloadFile(content, filename, type) {
      const blob = new Blob([content], { type })
      const url = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = filename
      a.click()
      URL.revokeObjectURL(url)
    },
  }
}
</script>

<style scoped>
.color-display {
  display: flex;
  flex-direction: column;
  height: 100%;
  overflow-y: auto;
  padding: 20px;
  gap: 20px;
  min-height: 0;
}

.color-cards {
  display: grid;
  gap: 8px;
  grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
}

.color-card {
  border-radius: 8px;
  overflow: hidden;
  cursor: pointer;
  transition: transform 0.2s, box-shadow 0.2s;
  display: flex;
  height: auto;
  min-height: 36px;
  flex-direction: column;
  justify-content: flex-start;
  align-items: stretch;
  position: relative;
}

.color-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 6px 16px rgba(0, 0, 0, 0.2);
}

.color-card.is-highlighted {
  box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.6), 0 8px 18px rgba(59, 130, 246, 0.25);
  transform: translateY(-2px);
}

.color-preview {
  width: 100%;
  height: 36px;
  border-radius: 5px 5px 0 0;
}

.color-info {
  padding: 10px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 0.85rem;
}

.color-code {
  font-family: 'Courier New', monospace;
  font-weight: 500;
  color: #333;
}

.color-diff {
  padding: 5px 10px 8px;
  display: flex;
  gap: 8px;
  align-items: center;
  font-size: 0.75rem;
  background: rgba(240, 240, 240, 0.5);
  border-top: 1px solid rgba(0, 0, 0, 0.05);
}

.diff-label {
  font-weight: 600;
  color: #666;
  font-size: 0.7rem;
}

.diff-value {
  font-family: 'Courier New', monospace;
  font-weight: 600;
  padding: 2px 5px;
  border-radius: 3px;
  font-size: 0.7rem;
}

.diff-positive {
  color: #16a34a;
  background: rgba(22, 163, 74, 0.1);
}

.diff-negative {
  color: #dc2626;
  background: rgba(220, 38, 38, 0.1);
}

.diff-zero {
  color: #6b7280;
  background: rgba(107, 114, 128, 0.1);
}

.copy-btn {
  display: inline-flex;
  align-items: center;
  background: none;
  border: none;
  cursor: pointer;
  font-size: 1.2rem;
  padding: 0 5px;
  transition: transform 0.2s;
}

.copy-btn:hover {
  transform: scale(1.2);
}

.color-actions {
  display: flex;
  align-items: center;
}

.pick-btn {
  display: inline-flex;
  align-items: center;
  background: none;
  border: none;
  cursor: pointer;
  font-size: 1.1rem;
  padding: 0 5px;
  transition: transform 0.2s;
}

.pick-btn:hover {
  transform: scale(1.15);
}

.edit-btn {
  display: inline-flex;
  align-items: center;
  background: none;
  border: none;
  cursor: pointer;
  font-size: 1.1rem;
  padding: 0 5px;
  transition: transform 0.2s;
}

.edit-btn:hover {
  transform: scale(1.15);
}

.palette-info {
  border-radius: 8px;
  padding: 15px;
  flex-shrink: 0;
}

.info-item {
  display: flex;
  justify-content: space-between;
  padding: 8px 0;
  border-bottom: 1px solid #f0f0f0;
  font-size: 0.95rem;
}

.info-item>.label {
  flex-shrink: 0;
}

.info-item:last-child {
  border-bottom: none;
}

.label {
  color: #666;
  font-weight: 500;
}

.value {
  color: #333;
  word-break: break-all;
}

.advice-item {
  align-items: flex-start;
}

.advice-text {
  line-height: 1.5;
  text-align: right;
}

.quick-actions {
  display: flex;
  gap: 10px;
  flex-shrink: 0;
  flex-wrap: wrap;
}

.action-btn {
  flex: 1;
  min-width: 100px;
  padding: 10px 15px;
  font-size: 0.9rem;
}

.action-btn:active {
  transform: scale(0.98);
}

@media (max-width: 768px) {
  .color-display {
    padding: 15px;
    gap: 15px;
  }

  .color-cards {
    grid-template-columns: repeat(auto-fit, minmax(100px, 1fr));
    gap: 10px;
  }

  .action-btn {
    flex: 1 1 calc(33.333% - 7px);
    padding: 8px 10px;
    font-size: 0.85rem;
  }
}
</style>
