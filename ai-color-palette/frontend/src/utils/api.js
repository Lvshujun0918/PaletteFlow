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

// 健康检查
export const healthCheck = () => {
  return apiClient.get('/health')
}

export default apiClient
