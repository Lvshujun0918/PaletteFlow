<template>
  <div class="generate-panel">
    <div class="input-section">
      <label for="prompt-input" class="label">描述你想要的配色：</label>
      <textarea
        id="prompt-input"
        v-model="prompt"
        class="input-textarea"
        placeholder="输入你的想法...支持中英文"
        @keydown.ctrl.enter="handleGenerate"
      ></textarea>
      <div class="char-count">{{ prompt.length }} / 500</div>
    </div>

    <GlassButton
      class="generate-btn"
      :loading="loading"
      :disabled="prompt.trim() === ''"
      @click="handleGenerate"
    >
      <span v-if="!loading">生成配色</span>
      <span v-else>正在生成中...</span>
    </GlassButton>

    <!-- 快速模板 -->
    <div class="templates">
      <h3>💡 快速模板</h3>
      <div class="template-list">
        <button
          v-for="template in templates"
          :key="template"
          class="template-btn glass-pill"
          @click="selectTemplate(template)"
        >
          {{ template }}
        </button>
      </div>
    </div>
  </div>
</template>

<script>
import { ref } from 'vue'
import GlassButton from './GlassButton.vue'

export default {
  name: 'GeneratePanel',
  components: {
    GlassButton
  },
  props: {
    loading: {
      type: Boolean,
      default: false
    }
  },
  emits: ['generate'],
  setup(props, { emit }) {
    const prompt = ref('')
    const templates = [
      '科技感蓝色',
      '温暖秋色调',
      '优雅紫色系',
      '清爽绿色',
      '活力彩虹色',
      '深邃黑金',
      '柔和粉色',
      '商务灰色'
    ]

    const handleGenerate = () => {
      if (prompt.value.trim()) {
        emit('generate', prompt.value.trim())
      }
    }

    const selectTemplate = (template) => {
      prompt.value = template
    }

    return {
      prompt,
      templates,
      handleGenerate,
      selectTemplate
    }
  }
}
</script>

<style scoped>
.generate-panel {
  display: flex;
  flex-direction: column;
  position: relative;
  z-index: 2;
  max-width: 800px;
  gap: 20px;
}

.input-section {
  display: flex;
  flex-direction: column;
  flex: 1;
}

.label {
  font-weight: 600;
  color: #333;
  font-size: 1rem;
  margin-bottom: 5px;
}

.input-textarea {
  flex: 1;
  padding: 15px;
  border: 1px solid rgba(148, 163, 184, 0.35);
  border-radius: 8px;
  font-family: inherit;
  font-size: 1rem;
  resize: none;
  transition: border-color 0.3s;
  min-height: 120px;
  background: rgba(255, 255, 255, 0.6);
}

.input-textarea:focus {
  outline: none;
  border-color: #667eea;
  box-shadow: 0 0 0 3px rgba(102, 126, 234, 0.1);
}

.char-count {
  font-size: 0.85rem;
  color: #999;
  text-align: right;
}

.generate-btn {
  font-size: 1.1rem;
  padding: 14px 24px;
}

.templates {
  flex-shrink: 0;
}

.templates h3 {
  margin-bottom: 10px;
  color: #333;
  font-size: 1rem;
}

.template-list {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 10px;
}

.template-btn {
  padding: 10px 15px;
  background: rgba(255, 255, 255, 0.6);
  border: 1px solid rgba(148, 163, 184, 0.3);
  border-radius: 6px;
  cursor: pointer;
  font-size: 0.9rem;
  transition: all 0.2s;
  text-align: center;
}

.template-btn:hover {
  background: rgba(255, 255, 255, 0.8);
  border-color: rgba(99, 102, 241, 0.4);
  color: #3b82f6;
}

.template-btn:active {
  transform: scale(0.98);
}

@media (max-width: 768px) {
  .generate-panel {
    gap: 15px;
  }

  .input-textarea {
    min-height: 100px;
  }

  .template-list {
    grid-template-columns: repeat(2, 1fr);
  }
}
</style>
