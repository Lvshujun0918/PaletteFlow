import axios from 'axios'

const API_BASE_URL = '/api'

// 创建axios实例
const apiClient = axios.create({
  baseURL: API_BASE_URL,
  timeout: 30000
})

const sleep = (ms) => new Promise((resolve) => setTimeout(resolve, ms))

const isRetryableError = (error) => {
  if (!error) return false
  if (error.code === 'ECONNABORTED') return true
  const status = error?.response?.status
  if (!status) return true
  return status === 429 || status >= 500
}

const requestWithRetry = async (requestFactory, options = {}) => {
  const retries = Number.isInteger(options.retries) ? options.retries : 2
  const baseDelay = Number.isInteger(options.baseDelay) ? options.baseDelay : 400

  let lastError
  for (let attempt = 0; attempt <= retries; attempt += 1) {
    try {
      return await requestFactory()
    } catch (error) {
      lastError = error
      if (attempt >= retries || !isRetryableError(error)) {
        throw error
      }
      const backoff = baseDelay * (attempt + 1)
      await sleep(backoff)
    }
  }

  throw lastError
}

// 生成配色方案
export const generatePalette = (prompt, colorCount = 5, options = {}) => {
  const payload = { prompt, color_count: colorCount }
  if (Array.isArray(options.seedColors)) {
    payload.seed_colors = options.seedColors
  }
  return requestWithRetry(
    () => apiClient.post('/generate-palette', payload, { timeout: 45000 }),
    { retries: 2, baseDelay: 500 }
  )
}

// 单色微调：仅替换指定位置的颜色
export const regenerateSingleColor = (payload) => {
  return requestWithRetry(
    () => apiClient.post('/regenerate-color', payload, { timeout: 45000 }),
    { retries: 2, baseDelay: 500 }
  )
}

// 微调配色方案
export const refinePalette = (currentColors, prompt, colorCount = currentColors.length) => {
  return requestWithRetry(
    () => apiClient.post('/refine-palette', {
      current_colors: currentColors,
      prompt: prompt,
      color_count: colorCount
    }, { timeout: 45000 }),
    { retries: 2, baseDelay: 500 }
  )
}

// 健康检查
export const healthCheck = () => {
  return apiClient.get('/health')
}

// 生成灵感短句
export const generateInspirationText = async () => {
  const response = await axios.get('https://v1.hitokoto.cn', {
    timeout: 10000,
    params: {
      c: 'i'
    }
  })

  const hitokoto = (response?.data?.hitokoto || '').trim()
  const from = (response?.data?.from || '').trim()

  if (!hitokoto) {
    throw new Error('empty hitokoto text')
  }

  const text = from ? `${hitokoto} ——《${from}》` : hitokoto
  return { data: { text } }
}

// 创建图片套色任务
export const createImagePaletteTask = (file, colors, mode = 'preserve_luma') => {
  const formData = new FormData()
  formData.append('image', file)
  formData.append('colors', (colors || []).join(','))
  formData.append('mode', mode)

  return apiClient.post('/apply-image-palette', formData, {
    headers: {
      'Content-Type': 'multipart/form-data'
    }
  })
}

// 查询图片套色任务状态
export const getImagePaletteTask = (taskId) => {
  return apiClient.get(`/apply-image-palette/task/${taskId}`, {
    timeout: 60000
  })
}

// 下载图片套色任务结果
export const downloadImagePaletteTaskResult = (taskId) => {
  return requestWithRetry(
    () => apiClient.get(`/apply-image-palette/task/${taskId}/result`, {
      responseType: 'blob',
      timeout: 0
    }),
    { retries: 2, baseDelay: 700 }
  )
}

export default apiClient
