<template>
  <div v-if="show" class="modal-overlay" @click.self="handleOverlayClick">
    <div class="modal-card" :class="variantClass">
      <div v-if="$slots.header" class="modal-header">
        <slot name="header" />
      </div>
      <div v-if="$slots.default" class="modal-body">
        <slot />
      </div>
      <div v-if="$slots.actions" class="modal-actions">
        <slot name="actions" />
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'AppModal',
  props: {
    show: {
      type: Boolean,
      default: false
    },
    variant: {
      type: String,
      default: 'confirm'
    },
    closeOnOverlay: {
      type: Boolean,
      default: true
    }
  },
  emits: ['close'],
  computed: {
    variantClass() {
      return `modal-card--${this.variant}`
    }
  },
  methods: {
    handleOverlayClick() {
      if (!this.closeOnOverlay) return
      this.$emit('close')
    }
  }
}
</script>

<style scoped>
.modal-overlay {
  position: absolute;
  inset: 0;
  background: rgba(255, 255, 255, 0.4);
  backdrop-filter: blur(10px);
  -webkit-backdrop-filter: blur(10px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 30;
}

.modal-card {
  width: min(90vw, 420px);
  padding-top: 10px;
  border-radius: 18px;
  border: 1px solid rgba(255, 255, 255, 0.65);
  background: linear-gradient(135deg, rgba(255, 255, 255, 0.55), rgba(255, 255, 255, 0.25));
  box-shadow: 0 16px 36px rgba(0, 0, 0, 0.14), inset 0 1px 0 rgba(255, 255, 255, 0.45);
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.modal-card--history {
  width: min(90vw, 600px);
  height: 80vh;
  padding: 0;
  border-radius: 24px;
}

.modal-card--choice {
  width: min(92vw, 500px);
  max-height: 80vh;
  padding: 0;
}

.modal-card--confirm {
  width: min(90vw, 420px);
}

.modal-header {
  padding: 16px 20px;
  border-bottom: 1px solid rgba(0, 0, 0, 0.06);
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-shrink: 0;
}

.modal-body {
  padding: 12px 16px;
}

.modal-actions {
  padding: 16px 20px;
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  border-top: 1px solid rgba(0, 0, 0, 0.06);
  background: rgba(255, 255, 255, 0.4);
  flex-shrink: 0;
  border-radius: 18px 0px;
}
</style>
