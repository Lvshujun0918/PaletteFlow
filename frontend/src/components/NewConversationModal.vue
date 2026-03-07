<template>
  <AppModal :show="show" variant="confirm" @close="handleClose">
    <template #header>
      <h3 class="modal-title">新建配色对话</h3>
    </template>

    <div class="new-chat-modal">
      <p class="modal-desc">你确定要新建一个配色对话吗？</p>

      <div class="count-card">
        <div class="count-header">
          <span class="count-label">配色数量</span>
        </div>
        <div class="count-row">
          <button
            type="button"
            class="count-btn"
            :disabled="localColorCount <= 1"
            @click="changeCount(-1)"
          >
            <IconMinus size="16" />
          </button>
          <div class="count-value">{{ localColorCount }} 色</div>
          <button
            type="button"
            class="count-btn"
            :disabled="localColorCount >= 10"
            @click="changeCount(1)"
          >
            <IconPlus size="16" />
          </button>
        </div>
      </div>

      <div class="preview-card">
        <div class="preview-title">待配色预览（点击方框设置初始色）</div>
        <div class="preview-grid">
          <button
            v-for="n in previewBoxes"
            :key="`preview-color-${n}`"
            type="button"
            class="preview-box"
            :class="{ active: !!seedColors[n - 1] }"
            :style="previewBoxStyle(n - 1, n)"
            @click="focusColorInput(n - 1)"
            :title="seedColors[n - 1] || '点击设置初始色'"
          >
            <input
              :ref="setColorInputRef(n - 1)"
              class="preview-color-input"
              type="color"
              :value="seedColors[n - 1] || '#94a3b8'"
              @input="setSeedColor(n - 1, $event)"
            />
            <span class="preview-index">{{ n }}</span>
          </button>
        </div>
        <p class="preview-hint">已设置的初始色会在首轮生成时保持不变，并被重点参考。</p>
      </div>
    </div>

    <template #actions>
      <GlassButton variant="secondary" @click="handleClose">取消</GlassButton>
      <GlassButton variant="primary" @click="handleConfirm">确认新建</GlassButton>
    </template>
  </AppModal>
</template>

<script>
import { computed, ref, watch } from 'vue'
import AppModal from './AppModal.vue'
import GlassButton from './GlassButton.vue'

const clampColorCount = (value) => {
  const count = Number(value)
  if (Number.isNaN(count)) return 5
  return Math.max(1, Math.min(10, count))
}

export default {
  name: 'NewConversationModal',
  components: {
    AppModal,
    GlassButton
  },
  props: {
    show: {
      type: Boolean,
      default: false
    },
    initialColorCount: {
      type: Number,
      default: 5
    },
    initialSeedColors: {
      type: Array,
      default: () => []
    }
  },
  emits: ['close', 'confirm'],
  setup(props, { emit }) {
    const localColorCount = ref(clampColorCount(props.initialColorCount))
    const seedColors = ref([])
    const colorInputRefs = ref([])

    const normalizeSeedColors = (count, source = []) => {
      const next = []
      for (let i = 0; i < count; i += 1) {
        const raw = source[i]
        if (typeof raw === 'string' && /^#[0-9A-Fa-f]{6}$/.test(raw.trim())) {
          next.push(raw.trim().toUpperCase())
        } else {
          next.push('')
        }
      }
      return next
    }

    watch(
      () => props.show,
      (visible) => {
        if (visible) {
          localColorCount.value = clampColorCount(props.initialColorCount)
          seedColors.value = normalizeSeedColors(localColorCount.value, props.initialSeedColors)
        }
      }
    )

    watch(
      () => props.initialColorCount,
      (value) => {
        if (!props.show) {
          localColorCount.value = clampColorCount(value)
          seedColors.value = normalizeSeedColors(localColorCount.value, props.initialSeedColors)
        }
      }
    )

    watch(
      () => props.initialSeedColors,
      (value) => {
        if (!props.show) {
          seedColors.value = normalizeSeedColors(localColorCount.value, value)
        }
      },
      { deep: true }
    )

    watch(localColorCount, (nextCount, prevCount) => {
      if (nextCount === prevCount) return
      seedColors.value = normalizeSeedColors(nextCount, seedColors.value)
      colorInputRefs.value = colorInputRefs.value.slice(0, nextCount)
    })

    const previewBoxes = computed(() => Array.from({ length: localColorCount.value }, (_, i) => i + 1))

    const changeCount = (delta) => {
      localColorCount.value = clampColorCount(localColorCount.value + delta)
    }

    const setColorInputRef = (index) => (el) => {
      colorInputRefs.value[index] = el || null
    }

    const focusColorInput = (index) => {
      const input = colorInputRefs.value[index]
      if (!input) return
      input.click()
    }

    const setSeedColor = (index, event) => {
      const value = String(event?.target?.value || '').trim().toUpperCase()
      if (!/^#[0-9A-F]{6}$/.test(value)) return
      seedColors.value[index] = value
    }

    const previewBoxStyle = (index, order) => {
      const selected = seedColors.value[index]
      if (selected) {
        return {
          background: selected,
          borderColor: 'rgba(15, 23, 42, 0.28)',
          opacity: 1
        }
      }
      return {
        opacity: Math.max(0.42, 1 - order * 0.05)
      }
    }

    const handleClose = () => {
      emit('close')
    }

    const handleConfirm = () => {
      emit('confirm', {
        colorCount: localColorCount.value,
        seedColors: [...seedColors.value]
      })
    }

    return {
      localColorCount,
      previewBoxes,
      seedColors,
      changeCount,
      setColorInputRef,
      focusColorInput,
      setSeedColor,
      previewBoxStyle,
      handleClose,
      handleConfirm
    }
  }
}
</script>

<style scoped>
.new-chat-modal {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.modal-title {
  font-size: 1.02rem;
  font-weight: 700;
  color: #1f2937;
}

.modal-desc {
  margin: 0;
  font-size: 0.9rem;
  color: #475569;
}

.count-card,
.preview-card {
  border: 1px solid rgba(148, 163, 184, 0.28);
  border-radius: 14px;
  background: rgba(255, 255, 255, 0.58);
  padding: 12px;
}

.count-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 10px;
}

.count-label {
  font-size: 0.84rem;
  color: #475569;
  font-weight: 600;
}

.count-badge {
  font-size: 0.78rem;
  color: #334155;
  padding: 4px 8px;
  border-radius: 999px;
  background: rgba(241, 245, 249, 0.92);
  border: 1px solid rgba(148, 163, 184, 0.25);
}

.count-row {
  display: grid;
  grid-template-columns: 34px 1fr 34px;
  align-items: center;
  gap: 10px;
}

.count-btn {
  width: 34px;
  height: 34px;
  border: 1px solid rgba(148, 163, 184, 0.38);
  border-radius: 10px;
  background: rgba(255, 255, 255, 0.78);
  color: #334155;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
}

.count-btn:disabled {
  opacity: 0.42;
  cursor: not-allowed;
}

.count-value {
  height: 34px;
  border-radius: 10px;
  border: 1px dashed rgba(148, 163, 184, 0.45);
  background: rgba(248, 250, 252, 0.72);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1rem;
  font-weight: 700;
  color: #1f2937;
}

.preview-title {
  font-size: 0.82rem;
  color: #64748b;
  margin-bottom: 8px;
}

.preview-grid {
  display: grid;
  grid-template-columns: repeat(8, minmax(0, 1fr));
  gap: 8px;
}

.preview-box {
  position: relative;
  overflow: hidden;
  height: 36px;
  width: 36px;
  border-radius: 7px;
  border: 1px solid rgba(148, 163, 184, 0.35);
  background: linear-gradient(135deg, rgba(226, 232, 240, 0.9), rgba(203, 213, 225, 0.8));
  cursor: pointer;
  transition: transform 0.15s ease, border-color 0.2s ease;
}

.preview-box:hover {
  transform: translateY(-1px);
}

.preview-box.active {
  box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.18);
}

.preview-color-input {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  opacity: 0;
  cursor: pointer;
}

.preview-index {
  position: absolute;
  right: 4px;
  bottom: 2px;
  font-size: 10px;
  line-height: 1;
  color: rgba(15, 23, 42, 0.72);
  background: rgba(255, 255, 255, 0.72);
  border-radius: 999px;
  padding: 1px 4px;
}

.preview-hint {
  margin: 8px 0 0;
  font-size: 0.76rem;
  color: #64748b;
}
</style>
