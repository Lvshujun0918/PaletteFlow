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
          <div class="header-actions">
            <Tooltip text="新建对话" position="bottom">
              <button class="header-action-btn" @click="confirmStartNewConversation">
                <IconPlus size="20" />
              </button>
            </Tooltip>
            <Tooltip text="查看对话历史" position="bottom">
              <button class="header-action-btn" @click="handleShowHistory">
                <IconHistory size="20" />
              </button>
            </Tooltip>
            <Tooltip text="设置" position="bottom">
              <button class="header-action-btn" data-tour="settings-btn" @click="openSettingsModal">
                <IconSettings size="20" />
              </button>
            </Tooltip>
          </div>
        </div>
      </div>

      <div class="main-content">
        <!-- 左侧：对话面板 -->
        <div class="panel panel-left glass-panel">
          <div class="chat-container">
            <div class="chat-header" data-tour="chat-header">
              <div class="chat-header-main">
                配色对话助手
                <p v-if="currentSessionTheme" class="session-theme-title">主题：{{ currentSessionTheme }}</p>
              </div>
            </div>

            <div class="chat-messages">
              <div v-for="(message, index) in chatMessages" :key="message.id" class="chat-message" :class="message.role">
                <div class="chat-bubble" :class="message.role">
                  <div v-if="message.type === 'text'" class="text-message-block">
                    <div>{{ message.content }}</div>
                    <GlassButton
                      v-if="message.role === 'assistant' && (message.payload?.retryContext || message.payload?.retryPrompt)"
                      variant="chip"
                      custom-class="retry-generate-btn"
                      :disabled="loading || loadingSingle"
                      @click="handleRetryFromMessage(message)"
                    >
                      重新生成
                    </GlassButton>
                  </div>

                  <template v-else-if="message.type === 'palette'">
                    <ChatPaletteMessage
                      :payload="message.payload"
                      :isCurrentMessage="isLastPaletteMessage(index)"
                      @pick-color="handlePickColorFromChat"
                      @hover-color="handleAdviceColorHover"
                      @restore="openRestoreConfirm(index)"
                    />
                  </template>

                  <template v-else-if="message.type === 'contrast'">
                    <ChatContrastMessage :payload="message.payload" />
                  </template>

                  <template v-else-if="message.type === 'colorblind'">
                    <ChatColorblindMessage
                      :payload="message.payload"
                      :colorblind-types="colorblindTypes"
                    />
                  </template>
                </div>
              </div>
            </div>

            <div class="chat-input">
              <div v-if="singleColorHex" class="selected-color-tip">
                <div class="selected-color-left">
                  <span class="selected-color-dot" :style="{ backgroundColor: singleColorHex }"></span>
                  <span class="selected-color-text">已选颜色 {{ singleColorHex }} 进行微调，请输入你的调整需求</span>
                </div>
                <button type="button" class="selected-color-close" title="退出单色微调"
                  @click="clearSingleColorMode">✕</button>
              </div>
              <textarea v-model="chatInput" class="input-textarea" data-tour="chat-input" placeholder="输入你的配色需求..."
                @keydown.ctrl.enter="handleSendPrompt"></textarea>
              <div class="input-footer">
                <div class="action-row action-toolbar" data-tour="action-row">
                  <GlassButton variant="chip" custom-class="tool-chip-btn" @click="sendQuickPrompt('不满意，重新生成')">重新生成</GlassButton>
                  <GlassButton variant="chip" custom-class="tool-chip-btn" @click="sendQuickPrompt('对比度检查')">对比度检查</GlassButton>
                  <GlassButton variant="chip" custom-class="tool-chip-btn" @click="sendQuickPrompt('色盲检查')">色盲检查</GlassButton>
                </div>
                <div class="send-actions" data-tour="send-actions">
                  <GlassButton
                    variant="primary"
                    class="send-btn"
                    :loading="loading"
                    :disabled="loading"
                    @click="handleSendPrompt"
                  >
                    <IconSend v-if="!loading" size="18" />发送
                  </GlassButton>
                  <GlassButton class="inspiration-btn" :loading="loadingInspiration"
                    :disabled="loading || loadingInspiration" @click="handleInspirationSend">
                    <IconSparkles v-if="!loadingInspiration" size="18" />灵感
                  </GlassButton>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- 右侧：配色显示面板 -->
        <div class="panel panel-right glass-panel" data-tour="result-panel">
          <ColorDisplay
            :colors="currentColors"
            :previousColors="previousColors"
            :prompt="currentPrompt"
            :timestamp="currentTimestamp"
            :advice="currentAdvice"
            :highlightedColor="hoveredAdviceColor"
            @regenerate="handleRegenerate"
            @pick-color="handlePickColorFromDisplay"
            @select-color="handleSelectColorForAI"
            @hover-color="handleAdviceColorHover"
            @apply-image="openImageApplyModal"
          />
        </div>
      </div>

      <!-- 手动可以调节颜色模态框 -->
      <ColorPickerModal 
        v-if="showColorPicker"
        :visible="showColorPicker" 
        :modelValue="editingColorValue" 
        @update:visible="showColorPicker = $event"
        @update:modelValue="editingColorValue = $event"
        @confirm="handleColorPickerConfirm"
        @close="showColorPicker = false"
      />

      <!-- 历史记录面板 -->
      <AppModal :show="showHistoryPanel" variant="history" :close-on-overlay="false" @close="showHistoryPanel = false">
        <template #header>
          <h3>历史记录</h3>
          <button class="close-btn" @click="showHistoryPanel = false">✕</button>
        </template>
        <div class="history-list-container">
          <div v-if="savedSessions.length === 0" class="empty-history">
            暂无历史会话
          </div>
          <div v-else class="history-session-item" v-for="session in savedSessions" :key="session.id" @click="loadSession(session)">
            <div class="session-info">
              <div class="session-theme">{{ session.theme || '无主题' }}</div>
              <div class="session-time">{{ formatTime(session.timestamp) }}</div>
              <div class="session-preview-colors">
                <span v-for="(c, i) in (session.colors || session.currentColors || [])" :key="i" class="mini-color-dot" :style="{ backgroundColor: c }"></span>
              </div>
            </div>
            <button class="delete-session-btn" @click.stop="deleteSession(session.id)">✕</button>
          </div>
        </div>
      </AppModal>

      <!-- 通知 -->
      <Notification />

      <NewConversationModal
        :show="showNewConversationConfirm"
        :initial-color-count="pendingColorCount"
        @close="cancelStartNewConversation"
        @confirm="handleConfirmNewConversation"
      />

      <AppModal :show="showRestoreConfirm" variant="confirm" @close="cancelRestoreToMessage">
        <template #header>
          <h3 class="modal-title">确认还原？</h3>
        </template>
        <div class="modal-text">还原后将删除该节点之后的聊天记录，此操作不可撤销。</div>
        <template #actions>
          <GlassButton variant="secondary" @click="cancelRestoreToMessage">取消</GlassButton>
          <GlassButton variant="primary" @click="confirmRestoreToMessage">确认还原</GlassButton>
        </template>
      </AppModal>

      <AppModal :show="showSessionChoice" variant="choice" :close-on-overlay="false">
        <template #header>
          <h3 class="modal-title">继续之前的创作</h3>
        </template>
        <div class="session-list-scroll">
           <div v-if="savedSessions.length === 0" class="empty-state">
              暂无历史会话记录
           </div>
           <div v-else class="history-session-item" v-for="session in savedSessions" :key="session.id" @click="loadSession(session)">
            <div class="session-info">
              <div class="session-theme">{{ session.theme || '无主题' }}</div>
              <div class="session-time">{{ formatTime(session.timestamp) }}</div>
              <div class="session-preview-colors">
                <span v-for="(c, i) in (session.colors || session.currentColors || [])" :key="i" class="mini-color-dot" :style="{ backgroundColor: c }"></span>
              </div>
            </div>
            <button class="delete-session-btn" @click.stop="deleteSession(session.id)">✕</button>
          </div>
        </div>
        <template #actions>
          <GlassButton data-tour="chat-start" variant="primary" custom-class="full-width" @click="confirmStartNewConversation">
            <IconEdit2 size="18" style="color: white" />开始新一轮配色
          </GlassButton>
        </template>
      </AppModal>

      <AppModal :show="showSettingsModal" variant="settings" @close="closeSettingsModal">
        <template #header>
          <h3 class="modal-title">设置</h3>
        </template>
        <AppSettings
          :projectInfo="projectInfo"
          :storageSummary="storageSummary"
          :storageItems="storageItems"
          @close="closeSettingsModal"
        />
        <template #actions>
          <GlassButton variant="secondary" @click="closeSettingsModal">完成</GlassButton>
        </template>
      </AppModal>

      <ImageApplyModal
        :show="showImageApplyModal"
        :colors="currentColors"
        @close="closeImageApplyModal"
      />
    </div>
  </div>
</template>

<script>
import { ref } from 'vue'
import { useFeatureLogic } from './featureLogic'
import { useFeatureWizard } from './feature/wizard'
import ColorDisplay from '../components/ColorDisplay.vue'
import Notification from '../components/Notification.vue'
import GlassButton from '../components/GlassButton.vue'
import ChatPaletteMessage from '../components/ChatPaletteMessage.vue'
import ChatContrastMessage from '../components/ChatContrastMessage.vue'
import ChatColorblindMessage from '../components/ChatColorblindMessage.vue'
import ColorPickerModal from '../components/ColorPickerModal.vue'
import AppModal from '../components/AppModal.vue'
import AppSettings from '../components/AppSettings.vue'
import ImageApplyModal from '../components/ImageApplyModal.vue'
import NewConversationModal from '../components/NewConversationModal.vue'
import logo from '../assets/logo.png'
import Tooltip from '../components/Tooltip.vue'
import { STORAGE_KEY, CHAT_STORAGE_KEY, SESSIONS_STORAGE_KEY } from './feature/constants'
import { generateInspirationText } from '../utils/api'

export default {
  name: 'App',
  components: {
    ColorDisplay,
    Notification,
    GlassButton,
    AppModal,
    AppSettings,
    ChatPaletteMessage,
    ChatContrastMessage,
    ChatColorblindMessage,
    ColorPickerModal,
    Tooltip,
    ImageApplyModal,
    NewConversationModal
  },
  data() {
    return {
      logoUrl: logo
    }
  },
  setup() {
    const featureLogic = useFeatureLogic()
    const { autoStartWizard } = useFeatureWizard()
    const hoveredAdviceColor = ref('')
    const loadingInspiration = ref(false)
    const showImageApplyModal = ref(false)
    const pendingColorCount = ref(Number(featureLogic.colorCount.value) || 5)
    const showRestoreConfirm = ref(false)
    const pendingRestoreIndex = ref(-1)
    const showSettingsModal = ref(false)
    const backupFileInput = ref(null)
    const storageItems = ref([])
    const storageSummary = ref({
      keyCount: 0,
      totalSize: '0 B',
      sessionCount: 0,
      chatCount: 0
    })
    const projectInfo = ref({
      name: 'PaletteFlow',
      version: 'v1.0.0',
      mode: import.meta.env.MODE || 'production',
      currentSession: '未命名会话'
    })

    const handleAdviceColorHover = (color) => {
      hoveredAdviceColor.value = color || ''
    }

    const confirmStartNewConversation = () => {
      pendingColorCount.value = Math.max(1, Math.min(10, Number(featureLogic.colorCount.value) || 5))
      featureLogic.confirmStartNewConversation()
    }

    const cancelStartNewConversation = () => {
      featureLogic.cancelStartNewConversation()
    }

    const handleConfirmNewConversation = (count) => {
      pendingColorCount.value = Math.max(1, Math.min(10, Number(count) || 5))
      featureLogic.colorCount.value = pendingColorCount.value
      featureLogic.proceedStartNewConversation()
    }

    const openRestoreConfirm = (index) => {
      pendingRestoreIndex.value = index
      showRestoreConfirm.value = true
    }

    const cancelRestoreToMessage = () => {
      showRestoreConfirm.value = false
      pendingRestoreIndex.value = -1
    }

    const confirmRestoreToMessage = () => {
      if (pendingRestoreIndex.value >= 0) {
        featureLogic.handleRestoreToMessage(pendingRestoreIndex.value)
      }
      cancelRestoreToMessage()
    }

    const handleInspirationSend = async () => {
      if (featureLogic.loading.value || loadingInspiration.value) return

      loadingInspiration.value = true
      try {
        const response = await generateInspirationText()
        const text = (response?.data?.text || '').trim()

        if (!text) {
          featureLogic.notify('灵感生成失败，请重试', 'error')
          return
        }

        featureLogic.chatInput.value = text
        featureLogic.handleSendPrompt()
      } catch (error) {
        console.error('获取灵感文案失败:', error)
        featureLogic.notify('灵感生成失败，请稍后再试', 'error')
      } finally {
        loadingInspiration.value = false
      }
    }

    const handleRetryFromMessage = (message) => {
      featureLogic.retryFailedMessage(message)
    }

    const openImageApplyModal = () => {
      showImageApplyModal.value = true
    }

    const closeImageApplyModal = () => {
      showImageApplyModal.value = false
    }

    autoStartWizard()

    const openSettingsModal = () => {
      projectInfo.value.currentSession = featureLogic.currentSessionTheme?.value || '未命名会话'
      refreshStorageInfo()
      showSettingsModal.value = true
    }

    const closeSettingsModal = () => {
      showSettingsModal.value = false
    }

    const formatBytes = (bytes) => {
      if (bytes <= 0) return '0 B'
      if (bytes < 1024) return `${bytes} B`
      if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(2)} KB`
      return `${(bytes / (1024 * 1024)).toFixed(2)} MB`
    }

    const getStorageLabel = (key) => {
      if (key === STORAGE_KEY) return '历史记录'
      if (key === CHAT_STORAGE_KEY) return '聊天记录'
      if (key === SESSIONS_STORAGE_KEY) return '会话列表'
      if (key.endsWith('_backup')) return '自动备份'
      if (key.includes('emergency')) return '紧急备份'
      return key
    }

    const refreshStorageInfo = () => {
      const items = []
      let totalBytes = 0
      let sessionCount = 0
      let chatCount = 0

      for (let i = 0; i < localStorage.length; i++) {
        const key = localStorage.key(i)
        if (!key || !key.startsWith('ai_color_palette')) continue

        const rawValue = localStorage.getItem(key) || ''
        const size = new Blob([rawValue]).size
        totalBytes += size

        if (key === SESSIONS_STORAGE_KEY) {
          try {
            const parsed = JSON.parse(rawValue)
            sessionCount = Array.isArray(parsed) ? parsed.length : 0
          } catch (error) {
            sessionCount = 0
          }
        }

        if (key === CHAT_STORAGE_KEY) {
          try {
            const parsed = JSON.parse(rawValue)
            chatCount = Array.isArray(parsed) ? parsed.length : 0
          } catch (error) {
            chatCount = 0
          }
        }

        items.push({
          key,
          name: getStorageLabel(key),
          size: formatBytes(size)
        })
      }

      storageItems.value = items.sort((a, b) => a.name.localeCompare(b.name, 'zh-CN'))
      storageSummary.value = {
        keyCount: items.length,
        totalSize: formatBytes(totalBytes),
        sessionCount,
        chatCount
      }
    }
    
    // 判断是否是最后一条palette消息
    const isLastPaletteMessage = (index) => {
      const messages = featureLogic.chatMessages.value
      // 从当前索引往后找，看是否还有palette类型的消息
      for (let i = index + 1; i < messages.length; i++) {
        if (messages[i].type === 'palette') {
          return false
        }
      }
      return true
    }
    
    return {
      ...featureLogic,
      isLastPaletteMessage,
      hoveredAdviceColor,
      loadingInspiration,
      handleAdviceColorHover,
      handleInspirationSend,
      handleRetryFromMessage,
      showImageApplyModal,
      openImageApplyModal,
      closeImageApplyModal,
      pendingColorCount,
      confirmStartNewConversation,
      cancelStartNewConversation,
      handleConfirmNewConversation,
      showRestoreConfirm,
      openRestoreConfirm,
      cancelRestoreToMessage,
      confirmRestoreToMessage,
      showSettingsModal,
      projectInfo,
      storageItems,
      storageSummary,
      openSettingsModal,
      closeSettingsModal,
      refreshStorageInfo
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
  height: 100px;
  padding: 20px;
  text-align: left;
  flex-shrink: 0;
  flex-direction: row;
  align-items: center;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
}

.header-action-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
  border-radius: 999px;
  border: 1px solid rgba(148, 163, 184, 0.35);
  background: rgba(255, 255, 255, 0.8);
  cursor: pointer;
  transition: transform 0.2s ease, box-shadow 0.2s ease;
  color: #333333;
}

.header-action-btn:hover {
  transform: translateY(-1px);
  box-shadow: 0 6px 14px rgba(15, 23, 42, 0.08);
  color: #2563eb;
}

.header-text {
  display: flex;
  align-items: baseline;
  gap: 20px;
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
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  font-weight: 600;
  color: #2d3748;
  font-size: 1.1rem;
}

.chat-header-main {
  display: flex;
  align-items: center;
  gap: 10px;
  min-width: 0;
}

.chat-messages {
  flex: 1;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding: 6px 6px 2px 0;
}

.chat-message.user {
  margin-left: auto;
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
  max-width: 100%;
}

.text-message-block {
  display: flex;
  flex-direction: row;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}

.chat-bubble :deep(.retry-generate-btn) {
  align-self: flex-start;
  min-height: 30px;
  padding: 6px 12px;
  font-size: 0.82rem;
}

.contrast-preview {
  display: flex;
}

.palette-summary {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.palette-title {
  display: flex;
  justify-content: space-between;
}

.palette-title-left {
  font-weight: 600;
  color: #2d3748;
}

.palette-title-right { 
  color: #d9d9d9;
}

.palette-colors {
  display: flex;
  gap: 6px;
  flex-wrap: wrap;
}

.palette-chip {
  display: block;
  margin: 2px;
  width: 24px;
  height: 24px;
  border-radius: 6px;
  border: 1px solid rgba(0, 0, 0, 0.08);
}

.clickable-chip {
  cursor: pointer;
  transition: transform 0.15s ease, box-shadow 0.15s ease;
}

.clickable-chip:hover {
  transform: translateY(-2px);
  box-shadow: 0 6px 14px rgba(0, 0, 0, 0.1);
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
  gap: 0;
  border-radius: 18px;
  background: rgba(255, 255, 255, 0.6);
  border: 1px solid rgba(255, 255, 255, 0.7);
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.6);
  margin: 0 16px 16px;
  overflow: hidden;
}

.action-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 14px;
  font-weight: 600;
  color: #2d3748;
  font-size: 0.95rem;
  background: rgba(255, 255, 255, 0.7);
  border: none;
  cursor: pointer;
}

.toggle-icon {
  font-size: 0.82rem;
  color: #718096;
}

.quick-actions-body {
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding: 12px 14px 14px;
  border-top: 1px solid rgba(148, 163, 184, 0.2);
}

.single-color-panel {
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding: 10px 12px;
  border-radius: 12px;
  background: rgba(255, 255, 255, 0.8);
  border: 1px dashed rgba(148, 163, 184, 0.4);
}

.single-color-header {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
}

.single-color-title {
  font-weight: 600;
  color: #2d3748;
  font-size: 0.95rem;
}

.single-color-preview {
  padding: 6px 10px;
  border-radius: 10px;
  color: #1a202c;
  border: 1px solid rgba(0, 0, 0, 0.08);
  min-width: 100px;
  text-align: center;
  font-family: 'Courier New', monospace;
}

.single-color-placeholder {
  font-size: 0.86rem;
  color: #718096;
}

.single-color-input {
  width: 100%;
  padding: 10px 12px;
  border-radius: 10px;
  border: 1px solid rgba(148, 163, 184, 0.4);
  background: rgba(255, 255, 255, 0.9);
}

.single-color-btn {
  align-self: flex-start;
  min-width: 140px;
}

.action-row {
  display: flex;
  gap: 10px;
  align-items: center;
  grid-template-columns: repeat(auto-fit, minmax(140px, 1fr));
}

.selector-row {
  grid-template-columns: repeat(auto-fit, minmax(160px, 1fr));
}

.chat-action {
  padding: 10px 14px;
  font-size: 0.88rem;
  border-radius: 999px;
  min-height: 46px;
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

.selected-color-tip {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  padding: 10px 12px;
  border-radius: 12px;
  background: rgba(37, 99, 235, 0.08);
  border: 1px solid rgba(37, 99, 235, 0.18);
  color: #1e3a8a;
  font-size: 0.9rem;
}

.selected-color-left {
  display: flex;
  align-items: center;
  gap: 8px;
  flex: 1;
  min-width: 0;
}

.selected-color-dot {
  width: 16px;
  height: 16px;
  border-radius: 50%;
  border: 1px solid rgba(0, 0, 0, 0.08);
}

.selected-color-text {
  flex: 1;
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
}

.selected-color-close {
  background: none;
  border: none;
  cursor: pointer;
  font-size: 0.95rem;
  padding: 4px 6px;
  color: #1e3a8a;
  border-radius: 6px;
  transition: background 0.2s ease, transform 0.2s ease;
}

.selected-color-close:hover {
  background: rgba(37, 99, 235, 0.12);
  transform: scale(1.05);
}

.chat-input .input-textarea {
  flex: 1;
  min-height: 160px;
  padding: 14px 16px;
  border-radius: 16px;
  border: 1px solid rgba(148, 163, 184, 0.45);
  background: rgba(255, 255, 255, 0.72);
  backdrop-filter: blur(8px);
  -webkit-backdrop-filter: blur(8px);
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.7), inset 0 -1px 0 rgba(15, 23, 42, 0.04);
  resize: none;
  font-size: 1rem;
  line-height: 1.6;
  font-family: 'SF Mono', 'JetBrains Mono', 'Consolas', 'Menlo', monospace;
  color: #1f2937;
  transition: border-color 0.2s ease, box-shadow 0.2s ease, background 0.2s ease;
}

.chat-input .input-textarea:focus {
  outline: none;
  border-color: rgba(99, 102, 241, 0.55);
  background: rgba(255, 255, 255, 0.86);
  box-shadow: 0 0 0 3px rgba(99, 102, 241, 0.15), inset 0 1px 0 rgba(255, 255, 255, 0.75);
}

.chat-input .input-textarea::placeholder {
  color: rgba(71, 85, 105, 0.78);
}

.action-toolbar {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 4px;
  border-radius: 12px;
  border: 1px solid rgba(148, 163, 184, 0.35);
  background: rgba(255, 255, 255, 0.5);
  backdrop-filter: blur(8px);
  -webkit-backdrop-filter: blur(8px);
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.7), 0 4px 10px rgba(15, 23, 42, 0.08);
}

.action-toolbar :deep(.tool-chip-btn) {
  min-height: 32px;
  padding: 6px 12px;
  border-radius: 8px;
  border: 1px solid transparent;
  background: transparent;
  box-shadow: none;
  color: #334155;
  font-size: 0.84rem;
  font-weight: 600;
}

.action-toolbar :deep(.tool-chip-btn:hover:not(:disabled)) {
  background: rgba(255, 255, 255, 0.75);
  border-color: rgba(148, 163, 184, 0.35);
  box-shadow: 0 2px 6px rgba(15, 23, 42, 0.1);
  color: #1e293b;
}

.action-toolbar :deep(.tool-chip-btn:active:not(:disabled)) {
  background: rgba(255, 255, 255, 0.86);
}

.send-btn {
  min-width: 156px;
}

.send-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.inspiration-btn {
  min-height: 42px;
  padding: 12px 18px;
  position: relative;
  overflow: hidden;
  border-radius: 999px;
  color: #ffffff !important;
  background: linear-gradient(108deg, 
    rgba(8, 148, 255, 0.6),
    rgba(201, 89, 221, 0.6) 34%,
    rgba(255, 46, 84, 0.6) 68%,
    rgba(255, 144, 4, 0.6)
  );
  backdrop-filter: blur(14px) saturate(175%);
  -webkit-backdrop-filter: blur(14px) saturate(175%);
  box-shadow:
    0 10px 22px rgba(79, 70, 229, 0.34),
    0 0 0 1px rgba(255, 255, 255, 0.2),
    inset 0 1px 0 rgba(255, 255, 255, 0.45) !important;
  transition: transform 0.22s ease, box-shadow 0.22s ease, filter 0.22s ease;
}

.inspiration-btn::before {
  content: '';
  position: absolute;
  inset: 1px;
  border-radius: 999px;
  background: linear-gradient(155deg, rgba(255, 255, 255, 0.48), rgba(255, 255, 255, 0.08));
  opacity: 0.75;
  pointer-events: none;
}

.inspiration-btn::after {
  content: '';
  position: absolute;
  top: -35%;
  left: -30%;
  width: 70%;
  height: 170%;
  transform: rotate(20deg);
  background: linear-gradient(90deg, rgba(255, 255, 255, 0), rgba(255, 255, 255, 0.42), rgba(255, 255, 255, 0));
  pointer-events: none;
  opacity: 0.55;
}

.inspiration-btn:hover:not(:disabled) {
  filter: saturate(1.08) brightness(1.04);
  box-shadow:
    0 14px 28px rgba(79, 70, 229, 0.4),
    0 0 0 1px rgba(255, 255, 255, 0.24),
    inset 0 1px 0 rgba(255, 255, 255, 0.5) !important;
}

.inspiration-btn:active:not(:disabled) {
  transform: translateY(0);
  box-shadow:
    0 8px 18px rgba(79, 70, 229, 0.32),
    0 0 0 1px rgba(255, 255, 255, 0.2),
    inset 0 1px 0 rgba(255, 255, 255, 0.42) !important;
}

.inspiration-btn:disabled {
  opacity: 0.7;
}

.inspiration-btn :deep(.glass-button__content),
.inspiration-btn :deep(.glass-button__spinner) {
  position: relative;
  z-index: 1;
}

.input-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  flex-wrap: wrap;
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
  background: rgba(255, 255, 255, 0.45);
  backdrop-filter: blur(8px);
  -webkit-backdrop-filter: blur(8px);
  border-radius: 15px;
  border: 1px solid rgba(255, 255, 255, 0.15);
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.08);
}

.logo-container:hover {
  background: rgba(255, 255, 255, 0.75);
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.12);
}

.logo {
  width: 64px;
  height: 64px;
  object-fit: contain;
  filter: drop-shadow(0 4px 12px rgba(0, 0, 0, 0.2));
  transition: transform 0.5s ease, filter 0.5s ease;
}

.logo:hover {
  transform: scale(1.1) rotate(5deg);
  filter: drop-shadow(0 6px 20px rgba(0, 0, 0, 0.3));
}

/* 新增样式：会话主题标题 */
.session-theme-title {
  flex-shrink: 9999;
  font-weight: 500;
  color: #4a5568;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

/* 历史记录面板样式 */

.close-btn {
  background: none;
  border: none;
  font-size: 1.5rem;
  color: #a0aec0;
  cursor: pointer;
  padding: 4px 8px;
  border-radius: 8px;
  transition: all 0.2s;
}

.close-btn:hover {
  background: rgba(0, 0, 0, 0.05);
  color: #4a5568;
}

.history-list-container {
  flex: 1;
  overflow-y: auto;
  padding: 20px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.empty-history {
  text-align: center;
  color: #a0aec0;
  margin-top: 40px;
}

.history-session-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px;
  background: rgba(255, 255, 255, 0.6);
  border: 1px solid rgba(255, 255, 255, 0.8);
  border-radius: 16px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.history-session-item:hover {
  background: rgba(255, 255, 255, 0.95);
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.05);
}

.session-info {
  flex: 1;
  min-width: 0;
}

.session-theme {
  font-weight: 600;
  color: #2d3748;
  margin-bottom: 4px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.session-time {
  font-size: 0.8rem;
  color: #718096;
  margin-bottom: 8px;
}

/* 宽屏卡片样式覆盖 */
.session-list-scroll {
  flex: 1;
  overflow-y: auto;
  padding: 12px 16px;
  display: flex;
  flex-direction: column;
  gap: 10px;
  min-height: 100px; /* 最小高度 */
}

.full-width {
  width: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  font-size: 1rem;
  padding: 10px;
  border-radius: 12px;
  transition: all 0.2s;
}

.empty-state {
  text-align: center;
  color: #a0aec0;
  padding: 40px 0;
  font-size: 0.95rem;
}

.session-preview-colors {
  display: flex;
  gap: 4px;
}

.mini-color-dot {
  width: 12px;
  height: 12px;
  border-radius: 50%;
  border: 1px solid rgba(0, 0, 0, 0.1);
}

.delete-session-btn {
  background: none;
  border: none;
  color: #cbd5e0;
  padding: 8px;
  margin-left: 12px;
  cursor: pointer;
  font-size: 1.1rem;
  transition: all 0.2s;
}

.delete-session-btn:hover {
  color: #e53e3e;
  background: rgba(229, 62, 62, 0.1);
  border-radius: 8px;
}

/* 响应式设计 */
@media (max-width: 1024px) {
  .main-content {
    flex-direction: column;
  }

  .panel-left {
    border-right: none;
    border-bottom: 1px solid #e0e0e0;
  }
  .chat-input {
    flex-direction: column;
  }
}

@media (max-width: 768px) {
  .chat-bubble {
    max-width: 100%;
  }

  .chat-header-main{
    flex-direction: column;
    align-items: flex-start;
  }
}
</style>
