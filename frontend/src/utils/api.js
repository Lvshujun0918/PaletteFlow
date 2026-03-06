import axios from 'axios'

const API_BASE_URL = '/api'

// 创建axios实例
const apiClient = axios.create({
  baseURL: API_BASE_URL,
  timeout: 10000
})

// 生成配色方案
export const generatePalette = (prompt, colorCount = 5) => {
  return apiClient.post('/generate-palette', { prompt, color_count: colorCount })
}

// 单色微调：仅替换指定位置的颜色
export const regenerateSingleColor = (payload) => {
  return apiClient.post('/regenerate-color', payload)
}

// 微调配色方案
export const refinePalette = (currentColors, prompt, colorCount = currentColors.length) => {
  return apiClient.post('/refine-palette', {
    current_colors: currentColors,
    prompt: prompt,
    color_count: colorCount
  })
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
  return apiClient.get(`/apply-image-palette/task/${taskId}`)
}

// 下载图片套色任务结果
export const downloadImagePaletteTaskResult = (taskId) => {
  return apiClient.get(`/apply-image-palette/task/${taskId}/result`, {
    responseType: 'blob'
  })
}

export default apiClient
