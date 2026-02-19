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
            <div class="chat-header">
              <div class="chat-header-main">
                配色对话助手
                <p v-if="currentSessionTheme" class="session-theme-title">主题：{{ currentSessionTheme }}</p>
              </div>
              <div class="chat-header-actions">
                <button class="chat-header-btn" @click="confirmStartNewConversation"><svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 48 48"><g fill="none" stroke="#333333" stroke-linecap="round" stroke-linejoin="round" stroke-width="4"><path d="M19 10V7a2 2 0 0 1 2-2h20a2 2 0 0 1 2 2v22a2 2 0 0 1-2 2h-4"/><rect width="24" height="24" x="5" y="18" rx="2"/><path d="M17 25v10m-5-5h10"/></g></svg></button>
                <button class="chat-header-btn" @click="handleShowHistory"><svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24"><path fill="#333333" d="M12 21q-3.45 0-6.012-2.287T3.05 13H5.1q.35 2.6 2.313 4.3T12 19q2.925 0 4.963-2.037T19 12t-2.037-4.962T12 5q-1.725 0-3.225.8T6.25 8H9v2H3V4h2v2.35q1.275-1.6 3.113-2.475T12 3q1.875 0 3.513.713t2.85 1.924t1.925 2.85T21 12t-.712 3.513t-1.925 2.85t-2.85 1.925T12 21m2.8-4.8L11 12.4V7h2v4.6l3.2 3.2z"/></svg></button>
              </div>
            </div>

            <div class="chat-messages">
              <div v-for="message in chatMessages" :key="message.id" class="chat-message" :class="message.role">
                <div class="chat-bubble" :class="message.role">
                  <div v-if="message.type === 'text'">{{ message.content }}</div>

                  <template v-else-if="message.type === 'palette'">
                    <div class="palette-summary">
                      <div class="palette-title">
                        <div class="palette-title-left">{{ message.payload.title || '已生成配色' }}</div>
                        <div class="palette-title-right">详细信息请查看右侧面板</div>
                      </div>
                      <div class="palette-colors">
                        <span v-for="(color, index) in message.payload.colors" :key="index" class="palette-chip clickable-chip"
                          :style="{ backgroundColor: color }" :title="color"
                          @click="handlePickColorFromChat(message.payload.colors, index)"></span>
                      </div>
                      <div class="palette-text">{{ message.payload.advice }}</div>
                    </div>
                  </template>

                  <template v-else-if="message.type === 'contrast'">
                    <div class="palette-summary">
                      <div class="palette-title">对比度检查结果</div>
                      <div class="palette-text">共检测 {{ message.payload.totalPairs ?? 1 }} 组颜色组合</div>
                      <div class="palette-text">最低对比度：{{ (message.payload.minRatio ?? message.payload.ratio ?? 0).toFixed(2) }}:1（{{ message.payload.minLevel || message.payload.level || '未知' }}）</div>
                      <div class="palette-text" v-if="message.payload.totalPairs">通过 WCAG AA（4.5:1）组合：{{ message.payload.passCount ?? 0 }}/{{ message.payload.totalPairs }}</div>
                      <div class="colorblind-block" v-for="(item, idx) in (Array.isArray(message.payload.results) && message.payload.results.length > 0 ? message.payload.results : [{ color1: message.payload.color1, color2: message.payload.color2, ratio: message.payload.ratio, level: message.payload.level }])" :key="idx">
                        <div class="contrast-preview">
                          <span class="palette-chip" :style="{ backgroundColor: item.color1 }"></span>
                          <span class="palette-chip" :style="{ backgroundColor: item.color2 }"></span>
                        </div>
                        <div class="palette-text">{{ item.color1 }} vs {{ item.color2 }}：{{ (item.ratio ?? 0).toFixed(2) }}:1（{{ item.level || '未知' }}）</div>
                      </div>
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
              <div v-if="singleColorHex" class="selected-color-tip">
                <div class="selected-color-left">
                  <span class="selected-color-dot" :style="{ backgroundColor: singleColorHex }"></span>
                  <span class="selected-color-text">已选颜色 {{ singleColorHex }} 进行微调，请输入你的调整需求</span>
                </div>
                <button type="button" class="selected-color-close" title="退出单色微调"
                  @click="clearSingleColorMode">✕</button>
              </div>
              <textarea v-model="chatInput" class="input-textarea" placeholder="输入你的配色需求..."
                @keydown.ctrl.enter="handleSendPrompt"></textarea>
              <div class="input-footer">
                <div class="input-tip">示例：温暖秋色调 / 科技感蓝色 / 适合网页仪表盘</div>
                <GlassButton v-if="!loading" class="send-btn" :loading="loading" :disabled="chatInput.trim() === ''"
                  @click="handleSendPrompt">
                  <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24"><g fill="none"><path d="m12.594 23.258l-.012.002l-.071.035l-.02.004l-.014-.004l-.071-.036q-.016-.004-.024.006l-.004.01l-.017.428l.005.02l.01.013l.104.074l.015.004l.012-.004l.104-.074l.012-.016l.004-.017l-.017-.427q-.004-.016-.016-.018m.264-.113l-.014.002l-.184.093l-.01.01l-.003.011l.018.43l.005.012l.008.008l.201.092q.019.005.029-.008l.004-.014l-.034-.614q-.005-.019-.02-.022m-.715.002a.02.02 0 0 0-.027.006l-.006.014l-.034.614q.001.018.017.024l.015-.002l.201-.093l.01-.008l.003-.011l.018-.43l-.003-.012l-.01-.01z"/><path fill="#ffffff" d="M20.235 5.686c.432-1.195-.726-2.353-1.921-1.92L3.709 9.048c-1.199.434-1.344 2.07-.241 2.709l4.662 2.699l4.163-4.163a1 1 0 0 1 1.414 1.414L9.544 15.87l2.7 4.662c.638 1.103 2.274.957 2.708-.241z"/></g></svg>发送
                </GlassButton>
                <GlassButton v-else class="send-btn" :loading="loading" :disabled="chatInput.trim() === ''"
                  @click="handleSendPrompt">
                  生成中...
                </GlassButton>
              </div>
            </div>
          </div>
        </div>

        <!-- 右侧：配色显示面板 -->
        <div class="panel panel-right glass-panel">
          <ColorDisplay :colors="currentColors" :prompt="currentPrompt" :timestamp="currentTimestamp"
            :advice="currentAdvice" @regenerate="handleRegenerate" @pick-color="handlePickColorFromDisplay" />
          <div class="quick-actions-panel" :class="{ collapsed: !isQuickActionsOpen }">
            <button class="action-header" @click="toggleQuickActions">
              <span>快捷指令</span>
              <span class="toggle-icon">{{ isQuickActionsOpen ? '收起' : '展开' }}</span>
            </button>
            <div class="quick-actions-body" v-show="isQuickActionsOpen">
              <div class="action-row">
                <button class="action-chip" @click="insertQuickInput('不满意，重新生成')">重新生成</button>
                <button class="action-chip" @click="insertQuickInput('对比度检查')">对比度检查</button>
                <button class="action-chip" @click="insertQuickInput('色盲检查')">色盲检查</button>
              </div>
              <div class="selector-hint">输入“对比度检查”将自动检测当前全部颜色组合</div>
            </div>
          </div>
        </div>
      </div>

      <!-- 历史记录面板 -->
      <div v-if="showHistoryPanel" class="history-panel-overlay">
        <div class="history-panel-card glass-panel">
          <div class="history-panel-header">
            <h3>历史记录</h3>
            <button class="close-btn" @click="showHistoryPanel = false">✕</button>
          </div>
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
        </div>
      </div>

      <!-- 通知 -->
      <Notification />

      <div v-if="showNewConversationConfirm" class="new-conversation-overlay" @click.self="cancelStartNewConversation">
        <div class="new-conversation-card glass-panel">
          <div class="new-conversation-title">确认新建对话？</div>
          <div class="new-conversation-text">当前未保存的上下文可能会丢失，是否继续？</div>
          <div class="new-conversation-actions">
            <button class="session-btn secondary" @click="cancelStartNewConversation">取消</button>
            <button class="session-btn primary" @click="proceedStartNewConversation">确认新建</button>
          </div>
        </div>
      </div>

      <div v-if="showSessionChoice" class="session-choice-overlay">
        <div class="session-choice-card glass-panel wide-card">
          <div class="session-choice-header">
            <div class="session-choice-title">继续之前的创作</div>
          </div>
          
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

          <div class="session-choice-actions">
            <button class="session-btn primary full-width" @click="confirmStartNewConversation">
              <span><svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 512 512"><path fill="#ffffff" d="M459.94 53.25a16.06 16.06 0 0 0-23.22-.56L424.35 65a8 8 0 0 0 0 11.31l11.34 11.32a8 8 0 0 0 11.34 0l12.06-12c6.1-6.09 6.67-16.01.85-22.38M399.34 90L218.82 270.2a9 9 0 0 0-2.31 3.93L208.16 299a3.91 3.91 0 0 0 4.86 4.86l24.85-8.35a9 9 0 0 0 3.93-2.31L422 112.66a9 9 0 0 0 0-12.66l-9.95-10a9 9 0 0 0-12.71 0"/><path fill="#ffffff" d="M386.34 193.66L264.45 315.79A41.1 41.1 0 0 1 247.58 326l-25.9 8.67a35.92 35.92 0 0 1-44.33-44.33l8.67-25.9a41.1 41.1 0 0 1 10.19-16.87l122.13-121.91a8 8 0 0 0-5.65-13.66H104a56 56 0 0 0-56 56v240a56 56 0 0 0 56 56h240a56 56 0 0 0 56-56V199.31a8 8 0 0 0-13.66-5.65"/></svg></span> 开始新一轮配色
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { useFeatureLogic } from './featureLogic'
import ColorDisplay from '../components/ColorDisplay.vue'
import Notification from '../components/Notification.vue'
import GlassButton from '../components/GlassButton.vue'
import logo from '../assets/logo.png'

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
    return useFeatureLogic()
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

.chat-header-actions {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
}

.chat-header-btn {
  display: inline-flex;
  align-items: center;
  border: 1px solid rgba(148, 163, 184, 0.35);
  background: rgba(255, 255, 255, 0.8);
  color: #2d3748;
  padding: 6px 10px;
  border-radius: 999px;
  font-size: 0.8rem;
  cursor: pointer;
  transition: transform 0.2s ease, box-shadow 0.2s ease;
}

.chat-header-btn:hover {
  transform: translateY(-1px);
  box-shadow: 0 6px 14px rgba(15, 23, 42, 0.08);
}

.session-choice-overlay {
  position: absolute;
  inset: 0;
  background: rgba(255, 255, 255, 0.28);
  backdrop-filter: blur(4px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 15;
}

.new-conversation-overlay {
  position: absolute;
  inset: 0;
  background: rgba(255, 255, 255, 0.28);
  backdrop-filter: blur(8px);
  -webkit-backdrop-filter: blur(8px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 30;
}

.new-conversation-card {
  width: min(90vw, 420px);
  padding: 20px;
  border-radius: 18px;
  border: 1px solid rgba(255, 255, 255, 0.65);
  background: linear-gradient(135deg, rgba(255, 255, 255, 0.55), rgba(255, 255, 255, 0.25));
  box-shadow: 0 16px 36px rgba(0, 0, 0, 0.14), inset 0 1px 0 rgba(255, 255, 255, 0.45);
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.new-conversation-title {
  font-size: 1.05rem;
  font-weight: 600;
  color: #1f2937;
}

.new-conversation-text {
  font-size: 0.92rem;
  color: #475569;
  line-height: 1.5;
}

.new-conversation-actions {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  margin-top: 2px;
}

.session-choice-card {
  width: min(92vw, 460px);
  padding: 22px;
  border-radius: 16px;
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.session-choice-title {
  font-size: 1.08rem;
  font-weight: 600;
  color: #1f2937;
}

.session-choice-text {
  color: #475569;
  font-size: 0.93rem;
  line-height: 1.5;
}

.session-choice-actions {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}

.session-btn {
  border: none;
  border-radius: 10px;
  padding: 9px 14px;
  cursor: pointer;
  font-size: 0.9rem;
}

.session-btn>span {
  display: inline-flex;
  align-items: center;
}

.session-btn.secondary {
  background: rgba(226, 232, 240, 0.9);
  color: #334155;
}

.session-btn.primary {
  background: rgba(37, 99, 235, 0.9);
  color: #fff;
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
  display: grid;
  gap: 10px;
  align-items: center;
  grid-template-columns: repeat(auto-fit, minmax(140px, 1fr));
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

/* 新增样式：会话主题标题 */
.session-theme-title {
  font-weight: 500;
  color: #4a5568;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: 400px;
}

/* 历史记录面板样式 */
.history-panel-overlay {
  position: fixed;
  top: 0;
  left: 0;
  width: 100vw;
  height: 100vh;
  background: rgba(255, 255, 255, 0.28);
  backdrop-filter: blur(8px);
  z-index: 1000;
  display: flex;
  justify-content: center;
  align-items: center;
}

.history-panel-card {
  width: 90%;
  max-width: 600px;
  height: 80vh;
  display: flex;
  flex-direction: column;
  background: rgba(255, 255, 255, 0.85);
  border: 1px solid rgba(255, 255, 255, 0.6);
  border-radius: 24px;
  box-shadow: 0 20px 50px rgba(0, 0, 0, 0.15);
  overflow: hidden;
}

.history-panel-header {
  padding: 20px 24px;
  border-bottom: 1px solid rgba(0, 0, 0, 0.05);
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.history-panel-header h3 {
  margin: 0;
  font-size: 1.25rem;
  color: #2d3748;
}

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
.session-choice-card.wide-card {
  width: min(92vw, 500px);
  max-width: 500px;
  max-height: 80vh;
  padding: 0;
  display: flex;
  flex-direction: column;
  background: rgba(255, 255, 255, 0.9); /* 提高不透明度 */
}

.session-choice-header {
  padding: 16px 20px;
  border-bottom: 1px solid rgba(0, 0, 0, 0.06);
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-shrink: 0;
}

.session-choice-title {
  font-size: 1.15rem;
  font-weight: 600;
  color: #1a202c;
  margin: 0;
}

.session-list-scroll {
  flex: 1;
  overflow-y: auto;
  padding: 12px 16px;
  display: flex;
  flex-direction: column;
  gap: 10px;
  min-height: 100px; /* 最小高度 */
}

.session-choice-actions {
  padding: 16px 20px;
  border-top: 1px solid rgba(0, 0, 0, 0.06);
  background: rgba(255, 255, 255, 0.4);
  flex-shrink: 0;
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
