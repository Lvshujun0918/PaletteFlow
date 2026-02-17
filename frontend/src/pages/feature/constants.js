export const STORAGE_KEY = 'ai_color_palette_ng_history'
export const CHAT_STORAGE_KEY = 'ai_color_palette_chat_history'
export const SESSIONS_STORAGE_KEY = 'ai_color_palette_sessions'

export const MAX_HISTORY = 20
export const MAX_CHAT_HISTORY = 200
export const MAX_SESSIONS = 50

export const DEFAULT_COLORS = [
  '#ffc2c2',
  '#ffe0c2',
  '#feffd6',
  '#d9ffcc',
  '#b9f9ff'
]

export const COLORBLIND_TYPES = [
  { key: 'deuteranopia', name: '红绿色盲 (Deuteranopia)' },
  { key: 'protanopia', name: '红绿色弱 (Protanopia)' },
  { key: 'tritanopia', name: '蓝黄色盲 (Tritanopia)' },
  { key: 'achromatopsia', name: '完全色盲 (Achromatopsia)' }
]

export const createWelcomeMessage = () => ({
  id: Date.now(),
  role: 'assistant',
  type: 'text',
  content: '你好！我是“PaletteFlow”智能体。描述你的配色需求，我会生成配色并提供使用建议。'
})
