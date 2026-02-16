<template>
  <div class="background-container" :style="{ background: currentBackground }">
    <div class="container glass-surface">
      <div class="top-content">
        <div class="header glass-panel">
          <div class="logo-container">
            <img :src="logoUrl" alt="Logo" class="logo" @error="handleLogoError">
          </div>
          <div class="header-text">
            <h1>PaletteFlow</h1>
            <p>配色，易如反掌</p>
          </div>
        </div>
      </div>

      <div class="main-content">
        <!-- 左侧：对话面板 -->
        <div class="panel panel-left glass-panel">
          <div class="chat-container">
            <div class="chat-header">配色对话助手</div>

            <div class="chat-messages">
              <div v-for="message in chatMessages" :key="message.id" class="chat-message" :class="message.role">
                <div class="chat-bubble" :class="message.role">
                  <div v-if="message.type === 'text'">{{ message.content }}</div>

                  <template v-else-if="message.type === 'palette'">
                    <div class="palette-summary">
                      <div class="palette-title">已生成配色</div>
                      <div class="palette-colors">
                        <span v-for="(color, index) in message.payload.colors" :key="index" class="palette-chip"
                          :style="{ backgroundColor: color }" :title="color"></span>
                      </div>
                      <div class="palette-text">详细信息请查看右侧配色面板</div>
                    </div>
                  </template>

                  <template v-else-if="message.type === 'history'">
                    <div class="history-list">
                      <div class="palette-title">历史记录</div>
                      <button v-for="item in message.payload" :key="item.id" class="history-item"
                        @click="handleSelectHistory(item)">
                        <span class="history-prompt">{{ item.prompt }}</span>
                        <span class="history-time">{{ formatTime(item.timestamp * 1000) }}</span>
                      </button>
                    </div>
                  </template>

                  <template v-else-if="message.type === 'contrast'">
                    <div class="palette-summary">
                      <div class="palette-title">对比度检查结果</div>
                      <div class="contrast-preview">
                        <span class="palette-chip" :style="{ backgroundColor: message.payload.color1 }"></span>
                        <span class="palette-chip" :style="{ backgroundColor: message.payload.color2 }"></span>
                      </div>
                      <div class="palette-text">对比度：{{ message.payload.ratio.toFixed(2) }}:1</div>
                      <div class="palette-text">等级：{{ message.payload.level }}</div>
                      <div class="palette-text">评分：{{ message.payload.score.toFixed(1) }}/100</div>
                    </div>
                  </template>

                  <template v-else-if="message.type === 'colorblind'">
                    <div class="palette-summary">
                      <div class="palette-title">色盲检查结果</div>
                      <div class="colorblind-block" v-for="type in colorblindTypes" :key="type.key">
                        <div class="palette-text">{{ type.name }}</div>
                        <div class="palette-colors">
                          <span v-for="(color, index) in message.payload[type.key]" :key="index"
                            class="palette-chip" :style="{ backgroundColor: color }"></span>
                        </div>
                      </div>
                      <div class="palette-text">
                        {{ message.payload.isAccessible ? '✅ 配色对色盲友好' : '❌ 建议调整以改善色盲可访问性' }}
                      </div>
                      <div class="palette-text">改进建议：{{ message.payload.recommendations.join('；') }}</div>
                    </div>
                  </template>
                </div>
              </div>
            </div>

            <div class="chat-input">
              <textarea v-model="chatInput" class="input-textarea" placeholder="输入你的配色需求..."
                @keydown.ctrl.enter="handleSendPrompt"></textarea>
              <div class="input-footer">
                <div class="input-tip">示例：温暖秋色调 / 科技感蓝色 / 适合网页仪表盘</div>
                <GlassButton class="send-btn" :loading="loading" :disabled="chatInput.trim() === ''"
                  @click="handleSendPrompt">
                  <span v-if="!loading">发送</span>
                  <span v-else>生成中...</span>
                </GlassButton>
              </div>
            </div>
          </div>
        </div>

        <!-- 右侧：配色显示面板 -->
        <div class="panel panel-right glass-panel">
          <ColorDisplay :colors="currentColors" :prompt="currentPrompt" :timestamp="currentTimestamp"
            :advice="currentAdvice" @regenerate="handleRegenerate" />
          <div class="quick-actions-panel">
            <div class="action-header">快捷指令</div>
            <div class="action-row">
              <button class="action-chip" @click="insertQuickInput('查看历史记录')">查看历史</button>
              <button class="action-chip" @click="insertQuickInput('不满意，重新生成')">重新生成</button>
              <button class="action-chip" @click="insertQuickInput('对比度检查')">对比度检查</button>
              <button class="action-chip" @click="insertQuickInput('色盲检查')">色盲检查</button>
            </div>
            <div class="action-row selector-row">
              <div class="selector-group">
                <label>颜色1</label>
                <select v-model="selectedColor1">
                  <option v-for="(color, index) in currentColors" :key="index" :value="color">{{ color }}</option>
                </select>
              </div>
              <div class="selector-group">
                <label>颜色2</label>
                <select v-model="selectedColor2">
                  <option v-for="(color, index) in currentColors" :key="index" :value="color">{{ color }}</option>
                </select>
              </div>
              <div class="selector-hint">选择颜色后输入“对比度检查”</div>
            </div>
          </div>
        </div>
      </div>

      <!-- 通知 -->
      <Notification />
    </div>
  </div>
</template>

<script>
import { ref, onMounted, computed } from 'vue'
import ColorDisplay from '../components/ColorDisplay.vue'
import Notification from '../components/Notification.vue'
import GlassButton from '../components/GlassButton.vue'
import { generatePalette, healthCheck } from '../utils/api'
import { notify } from '../utils/notify'
import {
  getContrastRatio,
  getContrastLevel,
  simulateDeuteranopia,
  simulateProtanopia,
  simulateTritanopia,
  simulateAchromatopsia
} from '../utils/colorUtils'
import logo from '../assets/logo.png'

const STORAGE_KEY = 'ai_color_palette_history'
const MAX_HISTORY = 20

export default {
  name: 'App',
  components: {
    ColorDisplay,
    Notification,
    GlassButton
  },
  data() {
    return {
      logoUrl: logo
    }
  },
  setup() {
    const loading = ref(false)
    const currentColors = ref([
      '#ffc2c2',
      '#ffe0c2',
      '#feffd6',
      '#d9ffcc',
      '#b9f9ff'
    ])
    const currentPrompt = ref('默认配色方案')
    const currentTimestamp = ref(Date.now())
    const currentAdvice = ref('')
    const histories = ref([])
    const chatInput = ref('')
    const chatMessages = ref([
      {
        id: Date.now(),
        role: 'assistant',
        type: 'text',
        content: '你好！描述你的配色需求，我会生成配色并提供使用建议。'
      }
    ])
    const selectedColor1 = ref('')
    const selectedColor2 = ref('')
    const colorblindTypes = [
      { key: 'deuteranopia', name: '红绿色盲 (Deuteranopia)' },
      { key: 'protanopia', name: '红绿色弱 (Protanopia)' },
      { key: 'tritanopia', name: '蓝黄色盲 (Tritanopia)' },
      { key: 'achromatopsia', name: '完全色盲 (Achromatopsia)' }
    ]

    // 计算属性：动态背景
    const currentBackground = computed(() => {
      // 如果还没有生成颜色，使用默认渐变
      if (!currentColors.value || currentColors.value.length === 0) {
        return 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)'
      }

      // 直接用最新生成的颜色创建渐变
      return `linear-gradient(135deg, ${currentColors.value.join(', ')})`
    })

    // localStorage相关函数
    const loadHistoriesFromStorage = () => {
      try {
        const stored = localStorage.getItem(STORAGE_KEY)
        if (stored) {
          histories.value = JSON.parse(stored)
          if (histories.value.length > 0) {
            const latest = histories.value[0]
            currentColors.value = latest.colors || []
            currentPrompt.value = latest.prompt || '默认配色方案'
            currentTimestamp.value = (latest.timestamp || Date.now()) * 1000
            currentAdvice.value = latest.advice || ''
          }
        }
      } catch (error) {
        console.error('加载历史记录失败:', error)
      }
    }

    const saveHistoriesToStorage = () => {
      try {
        localStorage.setItem(STORAGE_KEY, JSON.stringify(histories.value))
      } catch (error) {
        console.error('保存历史记录失败:', error)
      }
    }

    const addChatMessage = (role, type, content, payload = null) => {
      chatMessages.value.push({
        id: `${Date.now()}-${Math.random().toString(16).slice(2)}`,
        role,
        type,
        content,
        payload
      })
    }

    const formatTime = (timestamp) => {
      if (!timestamp) return '未知'
      const date = new Date(timestamp)
      return date.toLocaleString('zh-CN')
    }

    const handleGenerate = async (prompt) => {
      loading.value = true
      try {
        const response = await generatePalette(prompt)
        currentColors.value = response.data.colors
        currentPrompt.value = prompt
        currentTimestamp.value = response.data.timestamp * 1000
        currentAdvice.value = response.data.advice || ''

        // 保存到历史记录
        const newHistory = {
          id: Date.now(),
          prompt: prompt,
          colors: response.data.colors,
          timestamp: response.data.timestamp,
          advice: response.data.advice || ''
        }

        histories.value.unshift(newHistory)

        // 最多保存20条记录
        if (histories.value.length > MAX_HISTORY) {
          histories.value.pop()
        }

        // 保存到localStorage
        saveHistoriesToStorage()

        notify('配色生成成功！', 'success')
        addChatMessage('assistant', 'palette', '', {
          colors: response.data.colors,
          prompt,
          advice: response.data.advice || ''
        })
      } catch (error) {
        console.error('生成配色失败:', error)
        notify('生成配色失败，请重试', 'error')
        addChatMessage('assistant', 'text', '生成失败了，请稍后再试。')
      } finally {
        loading.value = false
      }
    }

    const handleSelectHistory = (item) => {
      currentColors.value = item.colors
      currentPrompt.value = item.prompt
      currentTimestamp.value = item.timestamp * 1000
      currentAdvice.value = item.advice || ''
      notify('已加载历史配色', 'success')
      addChatMessage('assistant', 'palette', '', {
        colors: item.colors,
        prompt: item.prompt,
        advice: item.advice || ''
      })
    }

    const handleRegenerate = () => {
      if (!currentColors.value || currentColors.value.length === 0) {
        notify('请先生成配色方案', 'warning')
        return
      }
      const colorsText = currentColors.value.join('、')
      const newPrompt = `对${colorsText}颜色不满意，请按照${currentPrompt.value}重新生成配色方案`
      addChatMessage('user', 'text', newPrompt)
      handleGenerate(newPrompt)
    }

    const insertQuickInput = (text) => {
      chatInput.value = text
    }

    const handleSendPrompt = () => {
      const prompt = chatInput.value.trim()
      if (!prompt) return
      addChatMessage('user', 'text', prompt)
      chatInput.value = ''
      if (prompt.includes('查看历史')) {
        handleShowHistory()
        return
      }
      if (prompt.includes('对比度检查')) {
        handleContrastCheck()
        return
      }
      if (prompt.includes('色盲检查')) {
        handleColorblindCheck()
        return
      }
      if (prompt.includes('不满意')) {
        handleRegenerate()
        return
      }
      handleGenerate(prompt)
    }

    const handleShowHistory = () => {
      if (histories.value.length === 0) {
        addChatMessage('assistant', 'text', '暂无历史记录，先生成一次配色吧。')
        return
      }
      addChatMessage('assistant', 'history', '', histories.value)
    }

    const handleContrastCheck = () => {
      if (!selectedColor1.value || !selectedColor2.value) {
        notify('请选择两个颜色进行对比度检查', 'warning')
        return
      }
      const ratio = getContrastRatio(selectedColor1.value, selectedColor2.value)
      const level = getContrastLevel(ratio)
      addChatMessage('assistant', 'contrast', '', {
        color1: selectedColor1.value,
        color2: selectedColor2.value,
        ratio,
        level,
        score: (ratio / 21) * 100
      })
    }

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

    const handleColorblindCheck = () => {
      if (!currentColors.value || currentColors.value.length === 0) {
        notify('请先生成配色方案', 'warning')
        return
      }
      const deuteranopia = currentColors.value.map(simulateDeuteranopia)
      const protanopia = currentColors.value.map(simulateProtanopia)
      const tritanopia = currentColors.value.map(simulateTritanopia)
      const achromatopsia = currentColors.value.map(simulateAchromatopsia)
      const minContrast = Math.min(
        getMinContrast(deuteranopia),
        getMinContrast(protanopia),
        getMinContrast(tritanopia),
        getMinContrast(achromatopsia)
      )
      addChatMessage('assistant', 'colorblind', '', {
        deuteranopia,
        protanopia,
        tritanopia,
        achromatopsia,
        isAccessible: minContrast >= 4.5,
        recommendations: buildRecommendations(minContrast)
      })
    }

    onMounted(async () => {
      // 健康检查
      try {
        await healthCheck()
        notify('连接到服务器成功', 'success')
      } catch (error) {
        console.error('服务器连接失败:', error)
        notify('无法连接到服务器，请确保后端已启动', 'error')
      }

      // 从localStorage加载历史记录
      loadHistoriesFromStorage()
      if (currentColors.value && currentColors.value.length > 0) {
        selectedColor1.value = currentColors.value[0]
        selectedColor2.value = currentColors.value[1] || currentColors.value[0]
      }
    })

    return {
      loading,
      currentColors,
      currentPrompt,
      currentBackground,
      currentTimestamp,
      currentAdvice,
      histories,
      chatInput,
      chatMessages,
      selectedColor1,
      selectedColor2,
      colorblindTypes,
      handleGenerate,
      handleSelectHistory,
      handleRegenerate,
      handleSendPrompt,
      insertQuickInput,
      handleShowHistory,
      handleContrastCheck,
      handleColorblindCheck,
      formatTime,
      notify
    }
  }
}
</script>

<style scoped>
.background-container {
  min-height: 100vh;
  background-attachment: fixed;
  /* 固定背景 */
  background-size: cover;
  position: relative;
  z-index: 0;
}

/* 修改 .container 样式，移除背景相关属性 */
.container {
  display: flex;
  flex-direction: column;
  min-height: 100vh;
  width: 100%;
  max-width: 1400px;
  margin: 0 auto;
  overflow: hidden;
  transform: scale(0.95);
  transform-origin: center center;
  position: relative;
  z-index: 1;
}

.header {
  display: flex;
  color: rgb(80, 76, 76);
  width: 100%;
  height: 160px;
  padding: 20px;
  text-align: left;
  flex-shrink: 0;
  flex-direction: row;
  align-items: center;
}

.header-text {
  margin-left: 20px;
  text-align: left;
  flex: 1;
}

.header h1 {
  font-size: 3.5rem;
  font-weight: 600;
  margin: 0 0 0.5rem 0;
  color: #333333;
  font-family: 'Playfair Display', Georgia, 'Times New Roman', serif;
  text-shadow: 0 2px 4px rgba(0, 0, 0, 0.1), 0 -1px 2px rgba(255, 255, 255, 0.3);
  letter-spacing: -0.5px;
  text-align: left;
  line-height: 1.2;
}

.header p {
  font-size: 1rem;
  opacity: 0.9;
  margin: 0;
  /* 移除默认边距 */
  text-align: left;
  /* 确保左对齐 */
  line-height: 1.5;
  /* 调整行高 */
}

.top-content {
  display: flex;
  gap: 20px;
  padding: 20px 20px 0px 20px;
}

.main-content {
  display: flex;
  gap: 20px;
  padding: 20px;
  flex: 1;
  overflow: hidden;
}

.panel {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  min-height: 0;
}

.chat-container {
  display: flex;
  flex-direction: column;
  height: 100%;
  gap: 14px;
  padding: 18px;
}

.chat-header {
  font-weight: 600;
  color: #2d3748;
  font-size: 1.1rem;
}

.chat-messages {
  flex: 1;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding: 6px 6px 2px 0;
}

.chat-message {
  display: flex;
}

.chat-message.user {
  justify-content: flex-end;
}

.chat-message.assistant {
  justify-content: flex-start;
}

.chat-bubble {
  max-width: 80%;
  padding: 12px 14px;
  border-radius: 14px;
  background: rgba(255, 255, 255, 0.72);
  color: #2d3748;
  box-shadow: 0 8px 20px rgba(0, 0, 0, 0.08);
}

.chat-bubble.user {
  background: rgba(37, 99, 235, 0.12);
  border: 1px solid rgba(37, 99, 235, 0.2);
}

.palette-summary {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.palette-title {
  font-weight: 600;
  color: #2d3748;
}

.palette-colors {
  display: flex;
  gap: 6px;
  flex-wrap: wrap;
}

.palette-chip {
  width: 24px;
  height: 24px;
  border-radius: 6px;
  border: 1px solid rgba(0, 0, 0, 0.08);
}

.palette-text {
  font-size: 0.88rem;
  color: #4a5568;
}

.history-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.history-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 10px;
  border-radius: 10px;
  border: 1px solid rgba(148, 163, 184, 0.2);
  background: rgba(255, 255, 255, 0.6);
  cursor: pointer;
}

.history-item:hover {
  background: rgba(255, 255, 255, 0.8);
}

.history-prompt {
  font-size: 0.9rem;
  color: #2d3748;
}

.history-time {
  font-size: 0.8rem;
  color: #718096;
}

.quick-actions-panel {
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding: 14px;
  border-radius: 18px;
  background: rgba(255, 255, 255, 0.6);
  border: 1px solid rgba(255, 255, 255, 0.7);
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.6);
  margin: 0 16px 16px;
}

.action-row {
  display: grid;
  gap: 10px;
  align-items: center;
  grid-template-columns: repeat(auto-fit, minmax(140px, 1fr));
}

.action-header {
  font-weight: 600;
  color: #2d3748;
  font-size: 0.95rem;
}

.action-chip {
  border: 1px solid rgba(148, 163, 184, 0.35);
  background: rgba(255, 255, 255, 0.8);
  color: #2d3748;
  padding: 8px 14px;
  border-radius: 999px;
  font-size: 0.88rem;
  cursor: pointer;
  transition: transform 0.2s ease, box-shadow 0.2s ease;
}

.action-chip:hover {
  transform: translateY(-1px);
  box-shadow: 0 8px 16px rgba(15, 23, 42, 0.1);
}

.selector-row {
  grid-template-columns: repeat(auto-fit, minmax(160px, 1fr));
}

.chat-action {
  padding: 10px 14px;
  font-size: 0.88rem;
  border-radius: 999px;
  min-height: 38px;
}

.selector-group {
  display: flex;
  flex-direction: column;
  gap: 4px;
  font-size: 0.85rem;
  color: #4a5568;
  min-width: 120px;
}

.selector-group select {
  padding: 8px 10px;
  border-radius: 8px;
  border: 1px solid rgba(148, 163, 184, 0.3);
  background: rgba(255, 255, 255, 0.7);
}

.selector-hint {
  font-size: 0.82rem;
  color: #718096;
}

.chat-input {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.chat-input .input-textarea {
  flex: 1;
  min-height: 160px;
  padding: 14px 16px;
  border-radius: 16px;
  border: 1px solid rgba(148, 163, 184, 0.3);
  background: rgba(255, 255, 255, 0.8);
  resize: none;
  font-size: 1rem;
  line-height: 1.6;
}

.send-btn {
  padding: 12px 22px;
  font-size: 0.95rem;
  min-height: 42px;
}

.input-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.input-tip {
  font-size: 0.85rem;
  color: #718096;
}

.logo-container {
  margin-bottom: 0;
  flex-shrink: 0;
  aspect-ratio: 1 / 1;
  display: flex;
  justify-content: center;
  align-items: center;
  margin-left: 0;
  margin-right: 0;
  margin-left: 0;
  background: rgba(255, 255, 255, 0.15);
  backdrop-filter: blur(8px);
  -webkit-backdrop-filter: blur(8px);
  border-radius: 30px;
  border: 1px solid rgba(255, 255, 255, 0.15);
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.08);
}

.logo-container:hover {
  background: rgba(255, 255, 255, 0.25);
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.12);
}

.logo {
  width: 96px;
  height: 96px;
  object-fit: contain;
  filter: drop-shadow(0 4px 12px rgba(0, 0, 0, 0.2));
  transition: transform 0.5s ease, filter 0.5s ease;
}

.logo:hover {
  transform: scale(1.1) rotate(5deg);
  filter: drop-shadow(0 6px 20px rgba(0, 0, 0, 0.3));
}

/* 响应式设计 */
@media (max-width: 1024px) {
  .header {
    height: 125px;
  }

  .main-content {
    flex-direction: column;
  }

  .panel-left {
    border-right: none;
    border-bottom: 1px solid #e0e0e0;
  }

  .header h1 {
    font-size: 2rem;
  }

  .chat-input {
    flex-direction: column;
  }
}

@media (max-width: 768px) {
  .header {
    padding: 20px 15px;
    height: 100px;
  }

  .logo-container {
    transform: scale(0.8);
  }

  .chat-bubble {
    max-width: 100%;
  }
}
</style>
