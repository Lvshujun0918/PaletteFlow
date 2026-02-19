<template>
  <div ref="wrapper" class="tooltip-wrapper" @mouseenter="handleMouseEnter" @mouseleave="handleMouseLeave">
    <slot></slot>
    <Teleport to="body">
      <Transition name="tooltip-fade">
        <div 
          v-if="show && hasContent" 
          class="tooltip-content" 
          :class="[`tooltip-${position}`]"
          :style="tooltipStyle"
        >
          <slot name="tooltip">{{ text }}</slot>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>

<script>
export default {
  name: 'Tooltip',
  props: {
    text: {
      type: String,
      default: ''
    },
    position: {
      type: String,
      default: 'top',
      validator: (value) => ['top', 'bottom', 'left', 'right', 'center'].includes(value)
    },
    delay: {
      type: Number,
      default: 0
    }
  },
  data() {
    return {
      show: false,
      timer: null,
      tooltipStyle: {}
    }
  },
  computed: {
    hasContent() {
      const trimmed = typeof this.text === 'string' ? this.text.trim() : ''
      return Boolean(trimmed) || Boolean(this.$slots.tooltip)
    }
  },
  methods: {
    updatePosition() {
      if (!this.$refs.wrapper) return
      
      const rect = this.$refs.wrapper.getBoundingClientRect()
      const scrollX = window.pageXOffset || document.documentElement.scrollLeft
      const scrollY = window.pageYOffset || document.documentElement.scrollTop
      
      // Base position
      let top = rect.top + scrollY
      let left = rect.left + scrollX
      
      // Adjust based on position prop
      switch (this.position) {
        case 'top':
          top = rect.top + scrollY - 8
          left = rect.left + scrollX + rect.width / 2
          break
        case 'bottom':
          top = rect.bottom + scrollY + 8
          left = rect.left + scrollX + rect.width / 2
          break
        case 'left':
          top = rect.top + scrollY + rect.height / 2
          left = rect.left + scrollX - 8
          break
        case 'right':
          top = rect.top + scrollY + rect.height / 2
          left = rect.right + scrollX + 8
          break
        case 'center':
          top = rect.top + scrollY + rect.height / 2
          left = rect.left + scrollX + rect.width / 2
          break
      }
      
      this.tooltipStyle = {
        top: `${top}px`,
        left: `${left}px`
      }
    },
    handleMouseEnter() {
      if (!this.hasContent) return
      if (this.delay > 0) {
        this.timer = setTimeout(() => {
          this.updatePosition()
          this.show = true
        }, this.delay)
      } else {
        this.updatePosition()
        this.show = true
      }
    },
    handleMouseLeave() {
      if (this.timer) {
        clearTimeout(this.timer)
        this.timer = null
      }
      this.show = false
    }
  },
  beforeUnmount() {
    if (this.timer) {
      clearTimeout(this.timer)
    }
  }
}
</script>

<style scoped>
.tooltip-wrapper {
  position: relative;
  display: inline-block;
}

.tooltip-content {
  position: absolute;
  background: rgba(0, 0, 0, 0.8);
  color: white;
  padding: 6px 10px;
  border-radius: 4px;
  font-size: 0.8rem;
  white-space: nowrap;
  pointer-events: none;
  z-index: 99999;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.2);
}

/* Top position */
.tooltip-top {
  transform: translate(-50%, -100%);
}

/* Bottom position */
.tooltip-bottom {
  transform: translate(-50%, 0);
}

/* Left position */
.tooltip-left {
  transform: translate(-100%, -50%);
}

/* Right position */
.tooltip-right {
  transform: translate(0, -50%);
}

/* Center position (overlay) */
.tooltip-center {
  transform: translate(-50%, -50%);
}

/* Arrow for tooltips */
.tooltip-top::after {
  content: '';
  position: absolute;
  top: 100%;
  left: 50%;
  transform: translateX(-50%);
  border: 5px solid transparent;
  border-top-color: rgba(0, 0, 0, 0.8);
}

.tooltip-bottom::after {
  content: '';
  position: absolute;
  bottom: 100%;
  left: 50%;
  transform: translateX(-50%);
  border: 5px solid transparent;
  border-bottom-color: rgba(0, 0, 0, 0.8);
}

.tooltip-left::after {
  content: '';
  position: absolute;
  left: 100%;
  top: 50%;
  transform: translateY(-50%);
  border: 5px solid transparent;
  border-left-color: rgba(0, 0, 0, 0.8);
}

.tooltip-right::after {
  content: '';
  position: absolute;
  right: 100%;
  top: 50%;
  transform: translateY(-50%);
  border: 5px solid transparent;
  border-right-color: rgba(0, 0, 0, 0.8);
}

/* Fade transition */
.tooltip-fade-enter-active,
.tooltip-fade-leave-active {
  transition: opacity 0.2s ease;
}

.tooltip-fade-enter-from,
.tooltip-fade-leave-to {
  opacity: 0;
}
</style>
