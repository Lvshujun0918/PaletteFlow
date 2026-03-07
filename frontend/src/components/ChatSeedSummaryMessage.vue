<template>
  <div class="seed-summary" data-tour="seed-summary">
    <div class="seed-title">{{ payload?.title || '新一轮配色配置已生效' }}</div>
    <div class="seed-meta">当前颜色数量：{{ safeColorCount }} 色</div>
    <div class="seed-grid">
      <div
        v-for="n in safeColorCount"
        :key="`seed-preview-${n}`"
        class="seed-box"
        :class="{ active: !!normalizedSeeds[n - 1] }"
        :style="boxStyle(n - 1, n)"
        :title="normalizedSeeds[n - 1] || '未固定'"
      >
        <span class="seed-index">{{ n }}</span>
      </div>
    </div>
    <div class="seed-tip">已固定颜色会在首轮生成时保持不变。</div>
  </div>
</template>

<script>
import { computed } from 'vue'

export default {
  name: 'ChatSeedSummaryMessage',
  props: {
    payload: {
      type: Object,
      default: () => ({})
    }
  },
  setup(props) {
    const safeColorCount = computed(() => {
      const count = Number(props.payload?.colorCount)
      if (Number.isNaN(count)) return 5
      return Math.max(1, Math.min(10, count))
    })

    const normalizedSeeds = computed(() => {
      const source = Array.isArray(props.payload?.seedColors) ? props.payload.seedColors : []
      const next = []
      for (let i = 0; i < safeColorCount.value; i += 1) {
        const raw = source[i]
        if (typeof raw === 'string' && /^#[0-9A-Fa-f]{6}$/.test(raw.trim())) {
          next.push(raw.trim().toUpperCase())
        } else {
          next.push('')
        }
      }
      return next
    })

    const boxStyle = (index, order) => {
      const selected = normalizedSeeds.value[index]
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

    return {
      safeColorCount,
      normalizedSeeds,
      boxStyle
    }
  }
}
</script>

<style scoped>
.seed-summary {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.seed-title {
  font-weight: 600;
  color: #2d3748;
}

.seed-meta {
  font-size: 0.84rem;
  color: #475569;
}

.seed-grid {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.seed-box {
  position: relative;
  overflow: hidden;
  height: 24px;
  width: 24px;
  border-radius: 6px;
  border: 1px solid rgba(148, 163, 184, 0.35);
  background: linear-gradient(135deg, rgba(226, 232, 240, 0.9), rgba(203, 213, 225, 0.8));
}

.seed-box.active {
  box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.12);
}

.seed-index {
  position: absolute;
  right: 3px;
  bottom: 2px;
  font-size: 9px;
  line-height: 1;
  color: rgba(15, 23, 42, 0.72);
  background: rgba(255, 255, 255, 0.72);
  border-radius: 999px;
  padding: 1px 3px;
}

.seed-tip {
  font-size: 0.78rem;
  color: #64748b;
}
</style>