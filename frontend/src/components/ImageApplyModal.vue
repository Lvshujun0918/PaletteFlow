<template>
  <AppModal :show="show" variant="settings" @close="handleClose">
    <template #header>
      <h3 class="modal-title">应用配色到图片</h3>
    </template>

    <div class="image-apply-modal">
      <input
        ref="imageInput"
        type="file"
        class="hidden-file-input"
        accept="image/png,image/jpeg"
        @change="handleFileChange"
      />

      <div class="image-apply-main">
        <section class="image-apply-config glass-card">
          <div class="section-title">图片与方案</div>

          <div class="palette-row">
            <span class="label">当前配色</span>
            <div class="palette-dots">
              <span
                v-for="(color, index) in colors"
                :key="`apply-color-${index}`"
                class="dot"
                :style="{ backgroundColor: color }"
                :title="color"
              ></span>
            </div>
          </div>

          <div class="mode-row">
            <div class="mode-header">
              <span class="label">替换方案</span>
              <span class="mode-name">{{ modeName }}</span>
            </div>
            <div class="slider-row">
              <span class="edge-label">艺术</span>
              <input
                v-model.number="modeLevel"
                type="range"
                min="0"
                max="2"
                step="1"
                class="mode-slider"
                @input="handleModeChange"
              />
              <span class="edge-label">现实</span>
            </div>
          </div>

          <div v-if="taskId" class="progress-row">
            <div class="progress-meta">
              <span>处理进度：{{ progress }}%</span>
              <span>{{ statusText }}</span>
            </div>
            <div class="progress-track">
              <div class="progress-fill" :style="{ width: `${progress}%` }"></div>
            </div>
          </div>
        </section>

        <section class="image-apply-preview glass-card">
          <div class="section-title">预览</div>
          <div class="preview-grid" :class="{ 'single-panel': !resultUrl }">
            <div class="preview-panel upload-panel" @click="pickFile">
              <div class="preview-title">原图（点击上传）</div>
              <div v-if="sourceUrl" class="preview-image-wrap">
                <img :src="sourceUrl" alt="原图预览" loading="lazy" decoding="async" />
              </div>
              <div v-else class="upload-placeholder">点击此处上传图片</div>
              <div class="upload-filename">{{ fileName || '支持 PNG / JPG' }}</div>
            </div>
            <div class="preview-panel" v-if="resultUrl">
              <div class="preview-title">应用结果</div>
              <img :src="resultUrl" alt="套色结果" loading="lazy" decoding="async" />
            </div>
          </div>
        </section>
      </div>
    </div>

    <template #actions>
      <GlassButton variant="secondary" @click="handleClose">关闭</GlassButton>
      <GlassButton variant="primary" :loading="loading" :disabled="!file" @click="applyPaletteToImage">开始处理</GlassButton>
      <GlassButton variant="primary" :disabled="!resultUrl || loading" @click="downloadResult">下载结果</GlassButton>
    </template>
  </AppModal>
</template>

<script>
import { computed, onBeforeUnmount, ref, watch } from 'vue'
import AppModal from './AppModal.vue'
import GlassButton from './GlassButton.vue'
import {
  createImagePaletteTask,
  downloadImagePaletteTaskResult,
  getImagePaletteTask
} from '../utils/api'
import { notify } from '../utils/notify'

export default {
  name: 'ImageApplyModal',
  components: {
    AppModal,
    GlassButton
  },
  props: {
    show: {
      type: Boolean,
      default: false
    },
    colors: {
      type: Array,
      default: () => []
    }
  },
  emits: ['close'],
  setup(props, { emit }) {
    const imageInput = ref(null)
    const file = ref(null)
    const fileName = ref('')
    const sourceUrl = ref('')
    const resultUrl = ref('')
    const loading = ref(false)
    const taskId = ref('')
    const progress = ref(0)
    const statusText = ref('等待开始')
    const pollingCancelled = ref(false)
    const modeLevel = ref(2)
    const mode = ref('preserve_luma')
    const modeMap = ['lab', 'soft_blend', 'preserve_luma']
    const modeNameMap = ['LAB最近邻', '软混合', '保留亮度']

    const modeName = computed(() => {
      const index = Math.max(0, Math.min(2, Number(modeLevel.value || 0)))
      return modeNameMap[index]
    })

    const revokeUrls = () => {
      if (sourceUrl.value) {
        URL.revokeObjectURL(sourceUrl.value)
        sourceUrl.value = ''
      }
      if (resultUrl.value) {
        URL.revokeObjectURL(resultUrl.value)
        resultUrl.value = ''
      }
    }

    const resetState = () => {
      pollingCancelled.value = true
      loading.value = false
      file.value = null
      fileName.value = ''
      taskId.value = ''
      progress.value = 0
      statusText.value = '等待开始'
      modeLevel.value = 2
      mode.value = 'preserve_luma'
      revokeUrls()
      if (imageInput.value) {
        imageInput.value.value = ''
      }
    }

    const handleClose = () => {
      resetState()
      emit('close')
    }

    const handleModeChange = () => {
      const level = Math.max(0, Math.min(2, Number(modeLevel.value || 0)))
      modeLevel.value = level
      mode.value = modeMap[level]
    }

    const pickFile = () => {
      if (imageInput.value) {
        imageInput.value.click()
      }
    }

    const handleFileChange = (event) => {
      const selectedFile = event?.target?.files?.[0]
      if (!selectedFile) return

      if (!selectedFile.type.startsWith('image/')) {
        notify('请选择 PNG 或 JPG 图片', 'warning')
        return
      }

      file.value = selectedFile
      fileName.value = selectedFile.name

      if (sourceUrl.value) {
        URL.revokeObjectURL(sourceUrl.value)
      }
      sourceUrl.value = URL.createObjectURL(selectedFile)

      if (resultUrl.value) {
        URL.revokeObjectURL(resultUrl.value)
        resultUrl.value = ''
      }
    }

    const sleep = (ms) => new Promise((resolve) => setTimeout(resolve, ms))

    const pollTask = async (id) => {
      for (let i = 0; i < 240; i++) {
        if (pollingCancelled.value) {
          return null
        }

        const statusResp = await getImagePaletteTask(id)
        const task = statusResp?.data || {}
        const taskStatus = task.status || 'processing'
        progress.value = Math.max(0, Math.min(100, Number(task.progress || 0)))

        if (taskStatus === 'queued') {
          statusText.value = '排队中'
        } else if (taskStatus === 'processing') {
          statusText.value = '处理中'
        } else if (taskStatus === 'completed') {
          statusText.value = '处理完成'
          progress.value = 100
          return task
        } else if (taskStatus === 'failed') {
          throw new Error(task.error || '图片处理失败')
        } else {
          statusText.value = '处理中'
        }

        await sleep(500)
      }

      throw new Error('图片处理超时，请稍后重试')
    }

    const applyPaletteToImage = async () => {
      if (!file.value) {
        notify('请先选择图片', 'warning')
        return
      }
      if (!props.colors || props.colors.length === 0) {
        notify('当前没有可用配色', 'warning')
        return
      }

      loading.value = true
      pollingCancelled.value = false
      progress.value = 0
      statusText.value = '正在创建任务'
      try {
        const createResp = await createImagePaletteTask(file.value, props.colors, mode.value)
        const id = createResp?.data?.task_id
        if (!id) {
          throw new Error('任务创建失败')
        }

        taskId.value = id
        const finishedTask = await pollTask(id)
        if (!finishedTask) {
          return
        }

        const resultResp = await downloadImagePaletteTaskResult(id)
        const blob = resultResp?.data
        if (!blob) {
          throw new Error('empty image response')
        }

        if (resultUrl.value) {
          URL.revokeObjectURL(resultUrl.value)
        }
        resultUrl.value = URL.createObjectURL(blob)
        notify('图片套色完成', 'success')
      } catch (error) {
        statusText.value = '处理失败'
        notify(error?.message || '图片套色失败，请稍后重试', 'error')
      } finally {
        loading.value = false
      }
    }

    const downloadResult = () => {
      if (!resultUrl.value) return
      const filename = (fileName.value || 'image').replace(/\.[^.]+$/, '')
      const link = document.createElement('a')
      link.href = resultUrl.value
      link.download = `${filename}_palette.png`
      link.click()
    }

    watch(
      () => props.show,
      (visible) => {
        if (!visible) {
          resetState()
        }
      }
    )

    onBeforeUnmount(() => {
      resetState()
    })

    return {
      imageInput,
      file,
      fileName,
      sourceUrl,
      resultUrl,
      loading,
      taskId,
      progress,
      statusText,
      modeLevel,
      modeName,
      pickFile,
      handleClose,
      handleFileChange,
      applyPaletteToImage,
      downloadResult,
      handleModeChange
    }
  }
}
</script>

<style scoped>
.image-apply-modal {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.image-apply-main {
  display: grid;
  grid-template-columns: minmax(300px, 380px) minmax(360px, 1fr);
  gap: 14px;
}

.glass-card {
  border: 1px solid rgba(255, 255, 255, 0.75);
  background: rgba(255, 255, 255, 0.62);
  border-radius: 14px;
  padding: 14px;
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.75);
}

.section-title {
  font-size: 0.95rem;
  font-weight: 600;
  color: #334155;
  margin-bottom: 12px;
}

.image-apply-config {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.label {
  font-size: 0.84rem;
  color: #475569;
}

.palette-row {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.palette-dots {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.dot {
  width: 16px;
  height: 16px;
  border-radius: 50%;
  border: 1px solid rgba(0, 0, 0, 0.1);
}

.mode-row {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.mode-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.mode-name {
  font-size: 0.8rem;
  color: #64748b;
}

.slider-row {
  display: flex;
  align-items: center;
  gap: 10px;
}

.edge-label {
  min-width: 28px;
  font-size: 0.8rem;
  color: #64748b;
  text-align: center;
}

.mode-slider {
  flex: 1;
}

.progress-row {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.progress-meta {
  display: flex;
  justify-content: space-between;
  font-size: 0.82rem;
  color: #475569;
}

.progress-track {
  width: 100%;
  height: 8px;
  border-radius: 999px;
  background: rgba(148, 163, 184, 0.22);
  overflow: hidden;
}

.progress-fill {
  height: 100%;
  width: 0;
  border-radius: 999px;
  background: rgba(99, 102, 241, 0.75);
  transition: width 0.25s ease;
}

.image-apply-preview {
  display: flex;
  flex-direction: column;
  min-height: 300px;
}

.preview-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 10px;
}

.preview-grid.single-panel {
  grid-template-columns: 1fr;
}

.preview-panel {
  border: 1px solid rgba(148, 163, 184, 0.3);
  border-radius: 12px;
  background: rgba(255, 255, 255, 0.55);
  padding: 8px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.upload-panel {
  cursor: pointer;
  transition: border-color 0.2s ease, box-shadow 0.2s ease, background 0.2s ease;
}

.upload-panel:hover {
  border-color: rgba(99, 102, 241, 0.45);
  background: rgba(255, 255, 255, 0.72);
  box-shadow: 0 0 0 2px rgba(99, 102, 241, 0.12);
}

.preview-title {
  font-size: 0.84rem;
  color: #334155;
}

.preview-image-wrap {
  flex: 1;
}

.upload-placeholder {
  min-height: 220px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #94a3b8;
  font-size: 0.88rem;
  border: 1px dashed rgba(148, 163, 184, 0.5);
  border-radius: 10px;
  background: rgba(248, 250, 252, 0.45);
}

.upload-filename {
  font-size: 0.82rem;
  color: #64748b;
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
}

.preview-panel img {
  width: 100%;
  max-height: 300px;
  object-fit: contain;
  border-radius: 8px;
  background: rgba(248, 250, 252, 0.8);
}

.hidden-file-input {
  display: none;
}

@media (max-width: 1024px) {
  .image-apply-main {
    grid-template-columns: 1fr;
  }

  .preview-grid {
    grid-template-columns: 1fr;
  }
}
</style>
