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

    <button
      class="generate-btn"
      @click="handleGenerate"
      :disabled="loading || prompt.trim() === ''"
    >
      <span v-if="!loading">生成配色</span>
      <span v-else>正在生成中...</span>
    </button>

    <!-- 快速模板 -->
    <div class="templates">
      <h3>💡 快速模板</h3>
      <div class="template-list">
        <button
          v-for="template in templates"
          :key="template"
          class="template-btn"
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

export default {
  name: 'GeneratePanel',
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
  width: 100%;
  padding: 3.5rem 3.5rem;
  background: rgba(255, 255, 255, 0.5);
  backdrop-filter: blur(10px) saturate(180%);
  -webkit-backdrop-filter: blur(10px) saturate(180%);
  border-radius: 24px;
  box-shadow: 
    0 20px 60px rgba(0, 0, 0, 0.3),
    inset 0 1px 0 rgba(255, 255, 255, 0.3);
  border: 1px solid rgba(255, 255, 255, 0.2);
}

.input-section {
  display: flex;
  flex-direction: column;
  gap: 10px;
  flex: 1;
}

.label {
  font-weight: 600;
  color: #333;
  font-size: 1rem;
}

.input-textarea {
  flex: 1;
  padding: 15px;
  border: 2px solid #e0e0e0;
  border-radius: 8px;
  font-family: inherit;
  font-size: 1rem;
  resize: none;
  transition: border-color 0.3s;
  min-height: 120px;
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
  padding: 14px 24px;
  background: linear-gradient(135deg, #ff9900 0%, #ffe100 100%);
  color: white;
  border: none;
  border-radius: 8px;
  font-size: 1.1rem;
  font-weight: 600;
  cursor: pointer;
  transition: transform 0.2s, box-shadow 0.2s;
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.4);
}

.generate-btn:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 6px 16px rgba(102, 126, 234, 0.5);
}

.generate-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.generate-btn:active:not(:disabled) {
  transform: translateY(0);
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
  background: #f5f5f5;
  border: 1px solid #e0e0e0;
  border-radius: 6px;
  cursor: pointer;
  font-size: 0.9rem;
  transition: all 0.2s;
  text-align: center;
}

.template-btn:hover {
  background: #f0f0f0;
  border-color: #667eea;
  color: #667eea;
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
