<template>
  <div class="check-panel">
    <!-- 对比度检查 -->
    <div class="check-section">
      <h3>对比度检查</h3>
      <div class="color-pair-selector">
        <div class="selector-group">
          <label>颜色1:</label>
          <select v-model="selectedColor1">
            <option v-for="(color, index) in colors" :key="index" :value="color">
              {{ color }}
            </option>
          </select>
        </div>
        <div class="selector-group">
          <label>颜色2:</label>
          <select v-model="selectedColor2">
            <option v-for="(color, index) in colors" :key="index" :value="color">
              {{ color }}
            </option>
          </select>
        </div>
        <button class="check-btn" @click="performContrastCheck">检查对比度</button>
      </div>

      <!-- 对比度结果 -->
      <div v-if="contrastResult" class="result-card glass-card">
        <div class="preview-row">
          <div
            class="color-preview"
            :style="{ backgroundColor: contrastResult.color1 }"
          ></div>
          <div
            class="color-preview"
            :style="{ backgroundColor: contrastResult.color2 }"
          ></div>
        </div>
        <div class="result-details">
          <div class="result-item">
            <span class="label">对比度比率:</span>
            <span class="value">{{ contrastResult.ratio.toFixed(2) }}:1</span>
          </div>
          <div class="result-item">
            <span class="label">WCAG等级:</span>
            <span :class="['value', 'level-' + contrastResult.level.toLowerCase()]">
              {{ contrastResult.level }}
            </span>
          </div>
          <div class="result-item">
            <span class="label">可访问性评分:</span>
            <span class="value">{{ contrastResult.score.toFixed(1) }}/100</span>
          </div>
          <div class="recommendation">
            <span v-if="contrastResult.level === 'AAA'" class="success">
              ✅ 优秀！满足所有WCAG 2.0对比度要求
            </span>
            <span v-else-if="contrastResult.level === 'AA'" class="warning">
              ⚠️ 良好 满足WCAG 2.0 AA等级要求
            </span>
            <span v-else class="error">
              ❌ 不足 对比度过低，建议调整
            </span>
          </div>
        </div>
      </div>
    </div>

    <div class="divider"></div>

    <!-- 色盲检查 -->
    <div class="check-section">
      <h3>🎨 色盲检查</h3>
      <button class="check-btn" @click="performColorblindCheck">检查色盲友好性</button>

      <!-- 色盲结果 -->
      <div v-if="colorblindResult" class="colorblind-result">
        <div class="colorblind-types">
          <div v-for="type in colorblindTypes" :key="type.key" class="colorblind-type">
            <h4>{{ type.name }}</h4>
            <div class="colorblind-colors">
              <div
                v-for="(color, index) in colorblindResult[type.key]"
                :key="index"
                class="colorblind-color"
                :style="{ backgroundColor: color }"
                :title="color"
              ></div>
            </div>
          </div>
        </div>

        <div class="accessibility-info">
          <div v-if="colorblindResult.isAccessible" class="success">
            ✅ 此配色方案对色盲友好
          </div>
          <div v-else class="error">
            ❌ 建议调整配色以改善色盲可访问性
          </div>

          <div class="recommendations">
            <h4>改进建议:</h4>
            <ul>
              <li v-for="(rec, index) in colorblindResult.recommendations" :key="index">
                {{ rec }}
              </li>
            </ul>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { ref, watch } from 'vue'
import {
  getContrastRatio,
  getContrastLevel,
  simulateDeuteranopia,
  simulateProtanopia,
  simulateTritanopia,
  simulateAchromatopsia
} from '../utils/colorUtils'

export default {
  name: 'CheckPanel',
  props: {
    colors: {
      type: Array,
      default: () => []
    }
  },
  emits: ['check-contrast', 'check-colorblind'],
  setup(props) {
    const selectedColor1 = ref('')
    const selectedColor2 = ref('')
    const contrastResult = ref(null)
    const colorblindResult = ref(null)

    const colorblindTypes = [
      { key: 'deuteranopia', name: '红绿色盲 (Deuteranopia)' },
      { key: 'protanopia', name: '红绿色弱 (Protanopia)' },
      { key: 'tritanopia', name: '蓝黄色盲 (Tritanopia)' },
      { key: 'achromatopsia', name: '完全色盲 (Achromatopsia)' }
    ]

    const getMinContrast = (palette) => {
      if (!palette || palette.length < 2) return 0
      let min = Infinity
      for (let i = 0; i < palette.length; i += 1) {
        for (let j = i + 1; j < palette.length; j += 1) {
          const ratio = getContrastRatio(palette[i], palette[j])
          min = Math.min(min, ratio)
        }
      }
      return min === Infinity ? 0 : min
    }

    const buildRecommendations = (minContrast) => {
      const recommendations = []
      if (minContrast < 4.5) {
        recommendations.push('提高明度差或增加饱和度对比')
        recommendations.push('避免相近色相的组合，拉开色相距离')
      }
      if (minContrast < 3) {
        recommendations.push('优先使用高对比度的浅色与深色搭配')
      }
      if (recommendations.length === 0) {
        recommendations.push('当前配色对色盲用户较友好，可继续使用')
      }
      return recommendations
    }

    // 初始化颜色选择
    watch(
      () => props.colors,
      (newColors) => {
        if (newColors.length > 0) {
          selectedColor1.value = newColors[0]
          selectedColor2.value = newColors[1] || newColors[0]
        }
      },
      { immediate: true }
    )

    const performContrastCheck = async () => {
      if (!selectedColor1.value || !selectedColor2.value) {
        alert('请选择两个颜色进行对比度检查')
        return
      }

      try {
        const ratio = getContrastRatio(selectedColor1.value, selectedColor2.value)
        const level = getContrastLevel(ratio)
        contrastResult.value = {
          color1: selectedColor1.value,
          color2: selectedColor2.value,
          ratio,
          level,
          score: (ratio / 21) * 100
        }
      } catch (error) {
        console.error('对比度检查失败:', error)
        alert('检查失败，请重试')
      }
    }

    const performColorblindCheck = async () => {
      if (props.colors.length === 0) {
        alert('请先生成配色方案')
        return
      }

      try {
        const deuteranopia = props.colors.map(simulateDeuteranopia)
        const protanopia = props.colors.map(simulateProtanopia)
        const tritanopia = props.colors.map(simulateTritanopia)
        const achromatopsia = props.colors.map(simulateAchromatopsia)

        const minContrast = Math.min(
          getMinContrast(deuteranopia),
          getMinContrast(protanopia),
          getMinContrast(tritanopia),
          getMinContrast(achromatopsia)
        )

        colorblindResult.value = {
          deuteranopia,
          protanopia,
          tritanopia,
          achromatopsia,
          isAccessible: minContrast >= 4.5,
          recommendations: buildRecommendations(minContrast)
        }
      } catch (error) {
        console.error('色盲检查失败:', error)
        alert('检查失败，请重试')
      }
    }

    return {
      selectedColor1,
      selectedColor2,
      contrastResult,
      colorblindResult,
      colorblindTypes,
      performContrastCheck,
      performColorblindCheck
    }
  }
}
</script>

<style scoped>
.check-panel {
  display: flex;
  flex-direction: column;
  gap: 20px;
  height: 100%;
  overflow-y: auto;
}

.check-section {
  display: flex;
  flex-direction: column;
  gap: 15px;
}

.check-section h3 {
  color: #333;
  font-size: 1.1rem;
  margin-bottom: 10px;
}

.color-pair-selector {
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
  align-items: flex-end;
}

.selector-group {
  display: flex;
  flex-direction: column;
  gap: 5px;
  flex: 1;
  min-width: 150px;
}

.selector-group label {
  font-size: 0.9rem;
  color: #666;
  font-weight: 500;
}

.selector-group select {
  padding: 8px 12px;
  border: 1px solid rgba(148, 163, 184, 0.35);
  border-radius: 6px;
  font-size: 0.9rem;
  cursor: pointer;
  background: rgba(255, 255, 255, 0.6);
}

.selector-group select:focus {
  outline: none;
  border-color: #667eea;
  box-shadow: 0 0 0 2px rgba(102, 126, 234, 0.1);
}

.check-btn {
  padding: 8px 16px;
  background: #667eea;
  color: white;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-size: 0.9rem;
  font-weight: 500;
  transition: background 0.3s;
  white-space: nowrap;
}

.check-btn:hover {
  background: #764ba2;
}

.divider {
  height: 1px;
  background: #e0e0e0;
  margin: 10px 0;
}

.result-card {
  border-radius: 8px;
  padding: 15px;
  margin-top: 10px;
}

.preview-row {
  display: flex;
  gap: 10px;
  margin-bottom: 15px;
}

.color-preview {
  flex: 1;
  height: 80px;
  border-radius: 6px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.result-details {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.result-item {
  display: flex;
  justify-content: space-between;
  font-size: 0.95rem;
  padding: 8px;
  background: rgba(255, 255, 255, 0.7);
  border-radius: 4px;
}

.result-item .label {
  color: #666;
  font-weight: 500;
}

.result-item .value {
  color: #333;
  font-weight: 600;
}

.level-aaa {
  color: #22c55e;
}

.level-aa {
  color: #f59e0b;
}

.level-fail {
  color: #ef4444;
}

.recommendation {
  margin-top: 10px;
  padding: 12px;
  border-radius: 6px;
  font-size: 0.9rem;
}

.success {
  background: #d1fae5;
  color: #065f46;
}

.warning {
  background: #fef3c7;
  color: #92400e;
}

.error {
  background: #fee2e2;
  color: #991b1b;
}

.colorblind-types {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
  gap: 15px;
  margin-bottom: 20px;
}

.colorblind-type {
  background: white;
  border: 1px solid #e0e0e0;
  border-radius: 8px;
  padding: 12px;
}

.colorblind-type h4 {
  font-size: 0.85rem;
  color: #333;
  margin-bottom: 10px;
  font-weight: 600;
}

.colorblind-colors {
  display: flex;
  gap: 5px;
  flex-wrap: wrap;
}

.colorblind-color {
  flex: 1;
  min-width: 30px;
  height: 40px;
  border-radius: 4px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.accessibility-info {
  background: white;
  border: 1px solid #e0e0e0;
  border-radius: 8px;
  padding: 15px;
}

.accessibility-info > div:first-child {
  margin-bottom: 15px;
  padding: 10px;
  border-radius: 6px;
  font-weight: 500;
}

.recommendations {
  margin-top: 15px;
}

.recommendations h4 {
  font-size: 0.9rem;
  color: #333;
  margin-bottom: 10px;
  font-weight: 600;
}

.recommendations ul {
  list-style: none;
  padding-left: 0;
}

.recommendations li {
  padding: 8px;
  color: #666;
  font-size: 0.9rem;
  line-height: 1.5;
}

.recommendations li:before {
  content: '▸ ';
  color: #667eea;
  margin-right: 8px;
}

@media (max-width: 768px) {
  .color-pair-selector {
    flex-direction: column;
  }

  .selector-group {
    min-width: 100%;
  }

  .colorblind-types {
    grid-template-columns: repeat(2, 1fr);
  }
}
</style>
