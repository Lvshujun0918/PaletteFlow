import {
  STORAGE_KEY,
  CHAT_STORAGE_KEY,
  SESSIONS_STORAGE_KEY,
  MAX_SESSIONS
} from './constants'

export function createStorageApi(deps) {
  const {
    histories,
    savedSessions,
    chatMessages,
    currentColors,
    currentSessionId,
    currentSessionTheme,
    currentPrompt,
    currentTimestamp,
    currentAdvice,
    cloneMessages
  } = deps

  const loadHistoriesFromStorage = () => {
    try {
      const stored = localStorage.getItem(STORAGE_KEY)
      if (stored) {
        histories.value = JSON.parse(stored)
        if (histories.value.length > 0) {
          const latest = histories.value[0]
          currentColors.value = latest.colors || []
          currentSessionId.value = latest.id
          currentSessionTheme.value = latest.prompt
          currentPrompt.value = latest.currentPrompt || latest.prompt
          currentTimestamp.value = (latest.timestamp || Date.now()) * 1000
          currentAdvice.value = latest.advice || ''
        }
      }
    } catch (error) {
      console.error('加载历史记录失败:', error)
    }
  }

  const saveHistoriesToStorage = () => {
    try {
      localStorage.setItem(STORAGE_KEY, JSON.stringify(histories.value))
    } catch (error) {
      console.error('保存历史记录失败:', error)
    }
  }

  const loadChatMessagesFromStorage = () => {
    try {
      const stored = localStorage.getItem(CHAT_STORAGE_KEY)
      if (!stored) return
      const parsed = JSON.parse(stored)
      if (!Array.isArray(parsed) || parsed.length === 0) return

      const safeMessages = parsed.filter((item) => item && typeof item === 'object' && item.role && item.type)
      if (safeMessages.length > 0) {
        chatMessages.value = safeMessages
      }
    } catch (error) {
      console.error('加载对话记录失败:', error)
    }
  }

  const getStoredChatMessages = () => {
    try {
      const stored = localStorage.getItem(CHAT_STORAGE_KEY)
      if (!stored) return []
      const parsed = JSON.parse(stored)
      if (!Array.isArray(parsed)) return []
      return parsed.filter((item) => item && typeof item === 'object' && item.role && item.type)
    } catch (error) {
      console.error('读取对话记录失败:', error)
      return []
    }
  }

  const loadSessionsFromStorage = () => {
    try {
      const stored = localStorage.getItem(SESSIONS_STORAGE_KEY)
      if (stored) {
        const parsed = JSON.parse(stored)
        if (Array.isArray(parsed)) {
          savedSessions.value = parsed.map((session) => ({
            ...session,
            messages: Array.isArray(session.messages) ? session.messages : []
          }))
        }
      }
    } catch (error) {
      console.error('加载历史会话失败', error)
      savedSessions.value = []
    }
  }

  const persistSessions = () => {
    try {
      // 【数据保护】验证数据有效性
      if (!Array.isArray(savedSessions.value)) {
        console.error('会话列表格式错误')
        return
      }

      // 【数据保护】备份现有数据
      const backupKey = SESSIONS_STORAGE_KEY + '_backup'
      const existingData = localStorage.getItem(SESSIONS_STORAGE_KEY)
      if (existingData) {
        try {
          localStorage.setItem(backupKey, existingData)
        } catch (backupError) {
          console.warn('会话列表备份失败:', backupError)
        }
      }

      localStorage.setItem(SESSIONS_STORAGE_KEY, JSON.stringify(savedSessions.value))
    } catch (error) {
      console.error('保存会话列表失败', error)
      // 【数据保护】尝试从备份恢复
      try {
        const backupKey = SESSIONS_STORAGE_KEY + '_backup'
        const backup = localStorage.getItem(backupKey)
        if (backup) {
          localStorage.setItem(SESSIONS_STORAGE_KEY, backup)
          console.log('已从备份恢复会话列表')
        }
      } catch (restoreError) {
        console.error('从备份恢复会话列表失败:', restoreError)
      }
    }
  }

  const saveCurrentSession = () => {
    if (!currentSessionId.value) return

    // 【数据保护】验证聊天记录有效性
    if (!Array.isArray(chatMessages.value) || chatMessages.value.length === 0) {
      console.warn('会话没有有效的聊天记录，跳过保存')
      return
    }

    try {
      const existing = savedSessions.value.find((session) => String(session.id) === String(currentSessionId.value))
      const payload = {
        id: existing ? existing.id : currentSessionId.value,
        theme: currentSessionTheme.value || '未命名主题',
        timestamp: Date.now(),
        colors: currentColors.value ? [...currentColors.value] : [],
        prompt: currentPrompt.value || '',
        advice: currentAdvice.value || '',
        messages: cloneMessages(chatMessages.value)
      }

      // 【数据保护】验证 payload 的有效性
      if (!payload.messages || payload.messages.length === 0) {
        console.warn('会话消息为空，跳过保存')
        return
      }

      if (existing) {
        Object.assign(existing, payload)
      } else {
        savedSessions.value.unshift(payload)
      }

      savedSessions.value.sort((left, right) => right.timestamp - left.timestamp)
      if (savedSessions.value.length > MAX_SESSIONS) {
        savedSessions.value = savedSessions.value.slice(0, MAX_SESSIONS)
      }
      persistSessions()
    } catch (error) {
      console.error('保存当前会话失败:', error)
      throw error // 向上抛出以便调用者处理
    }
  }

  const saveChatMessagesToStorage = (startNewSession = false) => {
    try {
      // 【数据保护】验证数据有效性
      if (!Array.isArray(chatMessages.value)) {
        console.error('聊天记录格式错误，跳过保存')
        return
      }

      const validMessages = chatMessages.value.filter(
        (msg) => msg && typeof msg === 'object' && msg.role && msg.type
      )
      
      if (validMessages.length === 0) {
        console.warn('没有有效的聊天记录，跳过保存')
        return
      }

      // 【数据保护】保存前先备份到临时key（防止写入失败导致数据丢失）
      const backupKey = CHAT_STORAGE_KEY + '_backup'
      const existingData = localStorage.getItem(CHAT_STORAGE_KEY)
      if (existingData) {
        try {
          localStorage.setItem(backupKey, existingData)
        } catch (backupError) {
          console.warn('备份失败，但继续保存新数据:', backupError)
        }
      }

      localStorage.setItem(CHAT_STORAGE_KEY, JSON.stringify(validMessages))
      
      if (currentSessionId.value && !startNewSession) {
        saveCurrentSession()
      }
    } catch (error) {
      console.error('保存对话记录失败:', error)
      // 【数据保护】如果保存失败，尝试从备份恢复
      try {
        const backupKey = CHAT_STORAGE_KEY + '_backup'
        const backup = localStorage.getItem(backupKey)
        if (backup) {
          localStorage.setItem(CHAT_STORAGE_KEY, backup)
          console.log('已从备份恢复对话记录')
        }
      } catch (restoreError) {
        console.error('从备份恢复失败:', restoreError)
      }
    }
  }

  return {
    loadHistoriesFromStorage,
    saveHistoriesToStorage,
    loadChatMessagesFromStorage,
    getStoredChatMessages,
    loadSessionsFromStorage,
    persistSessions,
    saveCurrentSession,
    saveChatMessagesToStorage
  }
}
