<template>
  <div class="history-panel">
    <div v-if="histories.length === 0" class="empty-state">
      <div class="empty-icon">📋</div>
      <p>还没有历史记录</p>
      <p class="hint">生成配色后，历史记录将显示在这里</p>
    </div>

    <div v-else class="history-list">
      <div
        v-for="item in histories"
        :key="item.id"
        class="history-item glass-card"
        @click="selectHistory(item)"
      >
        <div class="history-colors">
          <div
            v-for="(color, index) in item.colors.slice(0, 5)"
            :key="index"
            class="history-color"
            :style="{ backgroundColor: color }"
          ></div>
        </div>

        <div class="history-info">
          <div class="history-prompt">{{ item.prompt }}</div>
          <div class="history-time">{{ formatTime(item.timestamp) }}</div>
        </div>

        <button class="select-btn glass-pill">选择</button>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'HistoryPanel',
  props: {
    histories: {
      type: Array,
      default: () => []
    }
  },
  emits: ['select'],
  methods: {
    selectHistory(item) {
      this.$emit('select', item)
    },
    formatTime(timestamp) {
      if (!timestamp) return '未知'
      const date = new Date(timestamp * 1000)
      const now = new Date()
      const diff = now - date

      if (diff < 60000) return '刚刚'
      if (diff < 3600000) return `${Math.floor(diff / 60000)}分钟前`
      if (diff < 86400000) return `${Math.floor(diff / 3600000)}小时前`

      return date.toLocaleDateString('zh-CN')
    }
  }
}
</script>

<style scoped>
.history-panel {
  height: 100%;
}

.empty-state {
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  height: 100%;
  color: #999;
}

.empty-icon {
  font-size: 3rem;
  margin-bottom: 20px;
  opacity: 0.5;
}

.empty-state p {
  margin: 5px 0;
  font-size: 0.95rem;
}

.hint {
  font-size: 0.85rem;
  opacity: 0.7;
}

.history-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.history-item {
  display: flex;
  gap: 12px;
  padding: 12px;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s;
  align-items: center;
}

.history-item:hover {
  border-color: #667eea;
  box-shadow: 0 2px 8px rgba(102, 126, 234, 0.1);
  transform: translateX(4px);
}

.history-colors {
  display: flex;
  gap: 4px;
  flex-shrink: 0;
}

.history-color {
  width: 32px;
  height: 32px;
  border-radius: 4px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.history-info {
  flex: 1;
  min-width: 0;
}

.history-prompt {
  font-weight: 500;
  color: #333;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 0.95rem;
}

.history-time {
  font-size: 0.85rem;
  color: #999;
  margin-top: 4px;
}

.select-btn {
  padding: 6px 12px;
  background: rgba(255, 255, 255, 0.6);
  color: #1f2937;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 0.85rem;
  font-weight: 500;
  white-space: nowrap;
  transition: background 0.3s;
  flex-shrink: 0;
}

.select-btn:hover {
  background: rgba(255, 255, 255, 0.8);
}

@media (max-width: 768px) {
  .history-item {
    flex-direction: column;
    gap: 10px;
  }

  .history-colors {
    width: 100%;
  }

  .select-btn {
    width: 100%;
    padding: 8px;
  }
}
</style>
