import axios from 'axios'

const API_BASE_URL = '/api'

// 创建axios实例
const apiClient = axios.create({
  baseURL: API_BASE_URL,
  timeout: 10000
})

// 生成配色方案
export const generatePalette = (prompt) => {
  return apiClient.post('/generate-palette', { prompt })
}

// 单色微调：仅替换指定位置的颜色
export const regenerateSingleColor = (payload) => {
  return apiClient.post('/regenerate-color', payload)
}

// 微调配色方案
export const refinePalette = (currentColors, prompt) => {
  return apiClient.post('/refine-palette', {
    current_colors: currentColors,
    prompt: prompt
  })
}

// 健康检查
export const healthCheck = () => {
  return apiClient.get('/health')
}

// 生成灵感短句
export const generateInspirationText = () => {
  return apiClient.get('/generate-inspiration')
}

export default apiClient
