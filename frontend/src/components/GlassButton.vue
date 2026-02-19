<template>
  <button
    :class="['glass-button', `glass-button--${variant}`, { 'is-loading': loading }, customClass]"
    :type="type"
    :disabled="disabled || loading"
    @click="$emit('click', $event)"
  >
    <span v-if="loading" class="glass-button__spinner"></span>
    <span class="glass-button__content">
      <slot />
    </span>
  </button>
</template>

<script>
export default {
  name: 'GlassButton',
  props: {
    variant: {
      type: String,
      default: 'cta'
    },
    type: {
      type: String,
      default: 'button'
    },
    disabled: {
      type: Boolean,
      default: false
    },
    loading: {
      type: Boolean,
      default: false
    },
    customClass: {
      type: String,
      default: ''
    }
  },
  emits: ['click']
}
</script>

<style scoped>
.glass-button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 12px 22px;
  border-radius: 999px;
  border: 1px solid rgba(255, 255, 255, 0.4);
  cursor: pointer;
  font-size: 0.95rem;
  font-weight: 600;
  transition: transform 0.2s ease, box-shadow 0.2s ease, background 0.2s ease, color 0.2s ease;
  white-space: nowrap;
}

.glass-button--cta {
  background: var(--glass-cta);
  color: #fff;
  box-shadow: var(--glass-cta-shadow);
}

.glass-button--cta:hover:not(:disabled) {
  background: var(--glass-cta-hover);
  transform: translateY(-1px);
}

.glass-button--primary {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: #fff;
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.4);
  border: 1px solid rgba(255, 255, 255, 0.2);
}

.glass-button--primary:hover:not(:disabled) {
  background: linear-gradient(135deg, #5568d3 0%, #6a3f91 100%);
  transform: translateY(-2px);
  box-shadow: 0 6px 16px rgba(102, 126, 234, 0.5);
}

.glass-button--secondary {
  background: #e2e8f0;
  color: #475569;
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.08);
  border: 1px solid rgba(0, 0, 0, 0.08);
}

.glass-button--secondary:hover:not(:disabled) {
  background: #cbd5e1;
  color: #334155;
  transform: translateY(-1px);
  box-shadow: 0 4px 10px rgba(0, 0, 0, 0.12);
}

.glass-button--ghost {
  background: rgba(255, 255, 255, 0.28);
  color: #1f2937;
  backdrop-filter: blur(12px) saturate(150%);
  -webkit-backdrop-filter: blur(12px) saturate(150%);
  box-shadow: 0 8px 16px rgba(15, 23, 42, 0.12);
}

.glass-button--ghost:hover:not(:disabled) {
  background: rgba(255, 255, 255, 0.4);
  transform: translateY(-1px);
}

.glass-button:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.glass-button__spinner {
  width: 14px;
  height: 14px;
  border-radius: 50%;
  border: 2px solid rgba(255, 255, 255, 0.6);
  border-top-color: rgba(255, 255, 255, 0.2);
  animation: spin 0.8s linear infinite;
}

.glass-button__content {
  display: inline-flex;
  align-items: center;
  gap: 8px;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}
</style>
