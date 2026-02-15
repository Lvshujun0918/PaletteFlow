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
        <!-- 左侧：配色显示面板 -->
        <div class="panel panel-left glass-panel">
          <ColorDisplay :colors="currentColors" :prompt="currentPrompt" :timestamp="currentTimestamp"
            :advice="currentAdvice" @regenerate="handleRegenerate" />
        </div>

        <!-- 右侧：功能面板 -->
        <div class="panel panel-right glass-panel">
          <!-- Tab切换 -->
          <div class="tabs">
            <button v-for="tab in tabs" :key="tab" :class="['tab-btn', { active: activeTab === tab }]"
              @click="activeTab = tab">
              {{ tab === 'generate' ? '生成配色' : tab === 'history' ? '历史记录' : '检查工具' }}
            </button>
          </div>

          <!-- Tab内容 -->
          <div class="tab-content">
            <!-- 生成配色 Tab -->
            <GeneratePanel v-if="activeTab === 'generate'" :loading="loading" @generate="handleGenerate" />

            <!-- 检查工具 Tab -->
            <CheckPanel v-if="activeTab === 'check'" :colors="currentColors" @check-contrast="handleCheckContrast"
              @check-colorblind="handleCheckColorblind" />

            <!-- 历史记录 Tab -->
            <HistoryPanel v-if="activeTab === 'history'" :histories="histories" @select="handleSelectHistory" />
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
import GeneratePanel from '../components/GeneratePanel.vue'
import CheckPanel from '../components/CheckPanel.vue'
import HistoryPanel from '../components/HistoryPanel.vue'
import Notification from '../components/Notification.vue'
import { generatePalette, healthCheck } from '../utils/api'
import { notify } from '../utils/notify'
import logo from '../assets/logo.png'

const STORAGE_KEY = 'ai_color_palette_history'
const MAX_HISTORY = 20

export default {
  name: 'App',
  components: {
    ColorDisplay,
    GeneratePanel,
    CheckPanel,
    HistoryPanel,
    Notification
  },
  data() {
    return {
      logoUrl: logo
    }
  },
  setup() {
    const activeTab = ref('generate')
    const tabs = ['generate', 'check', 'history']
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
      } catch (error) {
        console.error('生成配色失败:', error)
        notify('生成配色失败，请重试', 'error')
      } finally {
        loading.value = false
      }
    }

    const handleSelectHistory = (item) => {
      currentColors.value = item.colors
      currentPrompt.value = item.prompt
      currentTimestamp.value = item.timestamp * 1000
      currentAdvice.value = item.advice || ''
      activeTab.value = 'generate'
      notify('已加载历史配色', 'success')
    }

    const handleRegenerate = () => {
      if (!currentColors.value || currentColors.value.length === 0) {
        notify('请先生成配色方案', 'warning')
        return
      }
      const colorsText = currentColors.value.join('、')
      const newPrompt = `对${colorsText}颜色不满意，请按照${currentPrompt.value}重新生成配色方案`
      handleGenerate(newPrompt)
    }

    const handleCheckContrast = () => {
      activeTab.value = 'check'
      notify('已切换到对比度检查', 'info')
    }

    const handleCheckColorblind = () => {
      activeTab.value = 'check'
      notify('已切换到色盲检查', 'info')
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
    })

    return {
      activeTab,
      tabs,
      loading,
      currentColors,
      currentPrompt,
      currentBackground,
      currentTimestamp,
      currentAdvice,
      histories,
      handleGenerate,
      handleSelectHistory,
      handleRegenerate,
      handleCheckContrast,
      handleCheckColorblind,
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



.tabs {
  display: flex;
  background: rgba(255, 255, 255, 0.4);
  border-bottom: 1px solid rgba(255, 255, 255, 0.35);
  padding: 4px;
  gap: 6px;
  flex-shrink: 0;
  border-radius: 16px;
  margin: 12px 12px 0;
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.45);
}

.tab-btn {
  flex: 1;
  padding: 12px 18px;
  border: none;
  background: transparent;
  cursor: pointer;
  font-size: 0.98rem;
  color: var(--glass-muted);
  transition: all 0.3s;
  border-radius: 12px;
  font-weight: 600;
}

.tab-btn:hover {
  background: rgba(255, 255, 255, 0.45);
}

.tab-btn.active {
  color: #2b6cb0;
  background: rgba(255, 255, 255, 0.65);
  box-shadow: 0 8px 20px rgba(37, 99, 235, 0.16);
}

.tab-content {
  flex: 1;
  overflow-y: auto;
}

.tab-content>div {
  padding: 20px;
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

  .tabs {
    flex-wrap: wrap;
  }

  .tab-btn {
    flex: 1 1 calc(33.333% - 10px);
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

  .tab-content {
    padding: 15px;
  }

  .tabs {
    gap: 0;
  }

  .tab-btn {
    flex: 1;
    padding: 10px 12px;
    font-size: 0.9rem;
  }
}
</style>
