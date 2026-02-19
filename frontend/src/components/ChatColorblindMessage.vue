<template>
  <div class="colorblind-summary">
    <div class="colorblind-title">色盲检查结果</div>
    <div class="colorblind-wrap">
        <div class="colorblind-block" v-for="type in colorblindTypes" :key="type.key">
        <div class="colorblind-name">{{ type.name }}</div>
        <div class="palette-colors">
            <span v-for="(color, index) in payload?.[type.key] || []" :key="index" class="palette-chip" :style="{ backgroundColor: color }"></span>
        </div>
        </div>
    </div>
    <div class="colorblind-status">
      {{ payload?.isAccessible ? '✅ 配色对色盲友好' : '❌ 建议调整以改善色盲可访问性' }}
    </div>
    <div class="colorblind-advice">改进建议：{{ (payload?.recommendations || []).join('；') }}</div>
  </div>
</template>

<script>
export default {
  name: 'ChatColorblindMessage',
  props: {
    payload: {
      type: Object,
      required: true
    },
    colorblindTypes: {
      type: Array,
      required: true
    }
  }
}
</script>

<style scoped>
.colorblind-summary {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.colorblind-title {
  font-weight: 600;
  color: #2d3748;
}

.colorblind-wrap {
  display: grid;
  gap: 8px;
  grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
}

.colorblind-block {
  display: flex;
  flex-direction: column;
  gap: 6px;
  padding: 8px 10px;
  border-radius: 12px;
  background: rgba(255, 255, 255, 0.7);
  border: 1px solid rgba(148, 163, 184, 0.2);
}

.colorblind-name {
  font-size: 0.85rem;
  color: #475569;
}

.palette-colors {
  display: flex;
  gap: 6px;
  flex-wrap: wrap;
}

.palette-chip {
  width: 20px;
  height: 20px;
  border-radius: 6px;
  border: 1px solid rgba(0, 0, 0, 0.08);
}

.colorblind-status {
  font-size: 0.88rem;
  color: #334155;
}

.colorblind-advice {
  font-size: 0.82rem;
  color: #64748b;
}
</style>
