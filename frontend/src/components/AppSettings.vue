<template>
  <div class="app-settings">
    <div class="settings-sidebar">
      <button
        v-for="tab in tabs"
        :key="tab.id"
        class="settings-menu-item"
        :class="{ active: activeTab === tab.id }"
        @click="activeTab = tab.id"
      >
        <span class="menu-icon">{{ tab.icon }}</span>
        <span class="menu-label">{{ tab.label }}</span>
      </button>
    </div>

    <div class="settings-content">
      <!-- 概览tab -->
      <template v-if="activeTab === 'overview'">
        <div class="settings-panel">
          <h2 class="settings-panel-title">概览</h2>
          
          <div class="settings-section">
            <h3 class="section-title">项目信息</h3>
            <div class="settings-row">
              <span class="row-label">应用名称</span>
              <span class="row-value">{{ projectInfo.name }}</span>
            </div>
            <div class="settings-row">
              <span class="row-label">版本</span>
              <span class="row-value">{{ projectInfo.version }}</span>
            </div>
            <div class="settings-row">
              <span class="row-label">运行环境</span>
              <span class="row-value">{{ projectInfo.mode }}</span>
            </div>
            <div class="settings-row">
              <span class="row-label">当前会话</span>
              <span class="row-value">{{ projectInfo.currentSession }}</span>
            </div>
          </div>

          <div class="settings-section">
            <h3 class="section-title">本地存储</h3>
            <div class="settings-row">
              <span class="row-label">存储键数量</span>
              <span class="row-value">{{ storageSummary.keyCount }}</span>
            </div>
            <div class="settings-row">
              <span class="row-label">数据总大小</span>
              <span class="row-value">{{ storageSummary.totalSize }}</span>
            </div>
            <div class="settings-row">
              <span class="row-label">会话条数</span>
              <span class="row-value">{{ storageSummary.sessionCount }}</span>
            </div>
            <div class="settings-row">
              <span class="row-label">消息条数</span>
              <span class="row-value">{{ storageSummary.chatCount }}</span>
            </div>
          </div>
        </div>
      </template>

      <!-- 存储详情tab -->
      <template v-if="activeTab === 'storage'">
        <div class="settings-panel">
          <h2 class="settings-panel-title">存储详情</h2>
          
          <div class="settings-section">
            <div class="storage-table">
              <div class="storage-table-header">
                <span class="storage-col-name">数据项</span>
                <span class="storage-col-size">大小</span>
              </div>
              <div v-if="storageItems.length === 0" class="storage-empty">
                暂无可展示的本地存储数据
              </div>
              <div v-else v-for="item in storageItems" :key="item.key" class="storage-table-row">
                <span class="storage-col-name">{{ item.name }}</span>
                <span class="storage-col-size">{{ item.size }}</span>
              </div>
            </div>
          </div>
        </div>
      </template>

      <!-- 备份tab -->
      <template v-if="activeTab === 'backup'">
        <div class="settings-panel">
          <h2 class="settings-panel-title">备份与恢复</h2>
          
          <div class="settings-section">
            <h3 class="section-title">数据备份</h3>
            <p class="section-desc">导出当前浏览器中的全部应用数据（会话、聊天记录、历史记录等）为 JSON 文件。</p>
            <GlassButton variant="secondary" @click="exportAllDataBackup">导出备份</GlassButton>
          </div>

          <div class="settings-section">
            <h3 class="section-title">数据恢复</h3>
            <p class="section-desc">从备份文件恢复全部数据。恢复后会覆盖当前本地数据并自动刷新页面。</p>
            <input ref="backupFileInput" type="file" accept="application/json" class="hidden-file-input" @change="handleBackupFileChange" />
            <GlassButton variant="primary" @click="openBackupFilePicker">选择备份并恢复</GlassButton>
          </div>
        </div>
      </template>
    </div>
  </div>
</template>

<script>
import { ref } from 'vue'
import GlassButton from './GlassButton.vue'
import { STORAGE_KEY, CHAT_STORAGE_KEY, SESSIONS_STORAGE_KEY } from '../pages/feature/constants'

export default {
  name: 'AppSettings',
  components: {
    GlassButton
  },
  props: {
    projectInfo: {
      type: Object,
      required: true
    },
    storageSummary: {
      type: Object,
      required: true
    },
    storageItems: {
      type: Array,
      required: true
    }
  },
  emits: ['close'],
  setup(props, { emit }) {
    const activeTab = ref('overview')
    const backupFileInput = ref(null)

    const tabs = [
      { id: 'overview', label: '概览', icon: '👁️' },
      { id: 'storage', label: '存储详情', icon: '💾' },
      { id: 'backup', label: '备份', icon: '📦' }
    ]

    const formatBytes = (bytes) => {
      if (bytes <= 0) return '0 B'
      if (bytes < 1024) return `${bytes} B`
      if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(2)} KB`
      return `${(bytes / (1024 * 1024)).toFixed(2)} MB`
    }

    const exportAllDataBackup = () => {
      try {
        const storage = {}
        for (let i = 0; i < localStorage.length; i++) {
          const key = localStorage.key(i)
          if (key && key.startsWith('ai_color_palette')) {
            storage[key] = localStorage.getItem(key)
          }
        }

        const payload = {
          app: 'PaletteFlow',
          version: 1,
          exportedAt: Date.now(),
          storage
        }

        const blob = new Blob([JSON.stringify(payload, null, 2)], { type: 'application/json' })
        const url = URL.createObjectURL(blob)
        const a = document.createElement('a')
        const stamp = new Date().toISOString().slice(0, 19).replace(/[T:]/g, '-')
        a.href = url
        a.download = `paletteflow-backup-${stamp}.json`
        a.click()
        URL.revokeObjectURL(url)
      } catch (error) {
        console.error('导出备份失败:', error)
      }
    }

    const openBackupFilePicker = () => {
      if (!backupFileInput.value) return
      backupFileInput.value.value = ''
      backupFileInput.value.click()
    }

    const handleBackupFileChange = (event) => {
      const file = event?.target?.files?.[0]
      if (!file) return

      const reader = new FileReader()
      reader.onload = () => {
        try {
          const raw = String(reader.result || '')
          const parsed = JSON.parse(raw)

          if (!parsed || typeof parsed !== 'object' || !parsed.storage || typeof parsed.storage !== 'object') {
            return
          }

          Object.entries(parsed.storage).forEach(([key, value]) => {
            if (key && key.startsWith('ai_color_palette')) {
              localStorage.setItem(key, value == null ? '' : String(value))
            }
          })

          setTimeout(() => {
            window.location.reload()
          }, 500)
        } catch (error) {
          console.error('恢复备份失败:', error)
        }
      }
      reader.onerror = () => {
        console.error('读取备份文件失败')
      }
      reader.readAsText(file)
    }

    return {
      activeTab,
      tabs,
      backupFileInput,
      exportAllDataBackup,
      openBackupFilePicker,
      handleBackupFileChange
    }
  }
}
</script>

<style scoped>
.app-settings {
  display: flex;
  height: 100%;
  width: 100%;
  background: rgba(255, 255, 255, 0.5);
  gap: 0;
  min-height: 500px;
}

.settings-sidebar {
  width: 200px;
  border-right: 1px solid rgba(148, 163, 184, 0.2);
  overflow-y: auto;
  background: rgba(255, 255, 255, 0.3);
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  padding: 8px 0;
}

.settings-menu-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
  border: none;
  background: transparent;
  cursor: pointer;
  color: #64748b;
  font-size: 0.95rem;
  transition: all 0.2s ease;
  text-align: left;
  white-space: nowrap;
  border-left: 3px solid transparent;
}

.settings-menu-item:hover {
  background: rgba(255, 255, 255, 0.4);
  color: #475569;
}

.settings-menu-item.active {
  background: rgba(37, 99, 235, 0.1);
  color: #2563eb;
  border-left-color: #2563eb;
  font-weight: 500;
}

.menu-icon {
  font-size: 1.2rem;
  width: 20px;
  text-align: center;
}

.menu-label {
  flex: 1;
}

.settings-content {
  flex: 1;
  overflow-y: auto;
  padding: 24px;
  min-width: 0;
}

.settings-panel {
  max-width: 600px;
}

.settings-panel-title {
  font-size: 1.8rem;
  font-weight: 600;
  color: #0f172a;
  margin: 0 0 24px 0;
}

.settings-section {
  margin-bottom: 32px;
}

.section-title {
  font-size: 0.95rem;
  font-weight: 600;
  color: #64748b;
  margin: 0 0 12px 0;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.section-desc {
  font-size: 0.9rem;
  color: #64748b;
  line-height: 1.6;
  margin: 0 0 12px 0;
}

.settings-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 0;
  border-bottom: 1px solid rgba(148, 163, 184, 0.15);
}

.settings-row:last-child {
  border-bottom: none;
}

.row-label {
  font-size: 0.95rem;
  color: #475569;
  font-weight: 500;
}

.row-value {
  font-size: 0.95rem;
  color: #94a3b8;
  font-family: 'Monaco', 'Courier New', monospace;
  word-break: break-all;
  text-align: right;
  max-width: 50%;
}

.storage-table {
  border: 1px solid rgba(148, 163, 184, 0.2);
  border-radius: 12px;
  overflow: hidden;
  background: rgba(255, 255, 255, 0.6);
}

.storage-table-header,
.storage-table-row {
  display: grid;
  grid-template-columns: 1fr auto;
  gap: 16px;
  padding: 12px 16px;
  align-items: center;
}

.storage-table-header {
  background: rgba(255, 255, 255, 0.7);
  border-bottom: 1px solid rgba(148, 163, 184, 0.15);
  font-weight: 600;
  font-size: 0.85rem;
  color: #64748b;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.storage-table-row {
  border-top: 1px solid rgba(148, 163, 184, 0.1);
  font-size: 0.9rem;
  color: #475569;
}

.storage-table-row:first-of-type {
  border-top: none;
}

.storage-col-name {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.storage-col-size {
  font-family: 'Monaco', 'Courier New', monospace;
  color: #94a3b8;
  text-align: right;
  flex-shrink: 0;
}

.storage-empty {
  padding: 32px 16px;
  text-align: center;
  color: #cbd5e0;
  font-size: 0.9rem;
}

.hidden-file-input {
  display: none;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .app-settings {
    flex-direction: column;
  }

  .settings-sidebar {
    width: 100%;
    border-right: none;
    border-bottom: 1px solid rgba(148, 163, 184, 0.2);
    flex-direction: row;
    overflow-x: auto;
    overflow-y: hidden;
    padding: 0 8px;
  }

  .settings-menu-item {
    padding: 10px 12px;
    border-left: none;
    border-bottom: 2px solid transparent;
    font-size: 0.9rem;
  }

  .settings-menu-item.active {
    border-left: none;
    border-bottom-color: #2563eb;
  }

  .settings-content {
    padding: 16px;
  }

  .settings-panel-title {
    font-size: 1.4rem;
  }
}
</style>
