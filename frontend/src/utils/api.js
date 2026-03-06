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

// 应用配色到图片（返回处理后的图片 blob）
export const applyImagePalette = (file, colors) => {
  const formData = new FormData()
  formData.append('image', file)
  formData.append('colors', (colors || []).join(','))

  return apiClient.post('/apply-image-palette', formData, {
    responseType: 'blob',
    headers: {
      'Content-Type': 'multipart/form-data'
    }
  })
}

export default apiClient
