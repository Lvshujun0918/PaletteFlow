export function createSessionApi(deps) {
  const {
    router,
    notify,
    currentColors,
    previousColors,
    currentPrompt,
    currentTimestamp,
    currentAdvice,
    currentSessionId,
    currentSessionTheme,
    chatInput,
    chatMessages,
    showSessionChoice,
    showHistoryPanel,
    savedSessions,
    selectedColor1,
    selectedColor2,
    singleColorHex,
    singleColorPrompt,
    singleColorMode,
    singleColorIndex,
    singleColorBase,
    createWelcomeMessage,
    cloneMessages,
    loadSessionsFromStorage,
    loadChatMessagesFromStorage,
    saveChatMessagesToStorage,
    saveCurrentSession,
    persistSessions
  } = deps

  const clearSingleColorMode = () => {
    singleColorHex.value = ''
    singleColorPrompt.value = ''
    singleColorMode.value = false
    singleColorIndex.value = 0
    singleColorBase.value = []
    chatInput.value = ''
  }

  const startNewConversation = () => {
    clearSingleColorMode()
    currentColors.value = []
    previousColors.value = [] // 清空上一组配色
    currentPrompt.value = ''
    currentTimestamp.value = Date.now()
    currentAdvice.value = ''
    currentSessionId.value = null
    currentSessionTheme.value = ''
    showSessionChoice.value = false
    router.replace('/feature')
    chatMessages.value = [createWelcomeMessage()]
    saveChatMessagesToStorage(true)
  }

  const findSessionById = (sessionId) => {
    if (!sessionId) return null
    const targetId = String(sessionId)
    return savedSessions.value.find((session) => String(session.id) === targetId) || null
  }

  const applySession = (session) => {
    if (!session) return false

    clearSingleColorMode()

    currentSessionId.value = session.id
    currentSessionTheme.value = session.theme || ''
    const restoredColors = session.colors || session.currentColors || []
    currentColors.value = restoredColors
    currentPrompt.value = session.prompt || ''
    currentAdvice.value = session.advice || ''
    const rawTimestamp = session.timestamp || Date.now()
    currentTimestamp.value = rawTimestamp > 1_000_000_000_000 ? rawTimestamp : rawTimestamp * 1000
    chatMessages.value = Array.isArray(session.messages) && session.messages.length > 0
      ? cloneMessages(session.messages)
      : [createWelcomeMessage()]

    // 根据聊天记录设置previousColors：查找倒数第二条palette消息
    let paletteCount = 0
    let previousPaletteColors = []
    for (let i = chatMessages.value.length - 1; i >= 0; i--) {
      if (chatMessages.value[i].type === 'palette' && chatMessages.value[i].payload?.colors) {
        paletteCount++
        if (paletteCount === 2) {
          previousPaletteColors = chatMessages.value[i].payload.colors
          break
        }
      }
    }
    previousColors.value = paletteCount >= 2 ? [...previousPaletteColors] : []

    if (restoredColors.length > 0) {
      selectedColor1.value = restoredColors[0]
      selectedColor2.value = restoredColors[1] || restoredColors[0]
    }

    saveChatMessagesToStorage()
    return true
  }

  const loadSessionById = (sessionId, options = {}) => {
    const { updateRoute = false, notifyUser = true } = options
    if (savedSessions.value.length === 0) {
      loadSessionsFromStorage()
    }
    const session = findSessionById(sessionId)
    if (!session) {
      if (notifyUser) notify('未找到该会话', 'warning')
      return false
    }

    if (currentSessionId.value && currentSessionId.value !== session.id) {
      saveCurrentSession()
    }

    const applied = applySession(session)
    if (applied && updateRoute) {
      router.push(`/feature/${session.id}`)
    }
    return applied
  }

  const loadSession = (session) => {
    if (currentSessionId.value === session.id) {
      showHistoryPanel.value = false
      showSessionChoice.value = false
      return
    }

    const applied = loadSessionById(session.id, { updateRoute: true, notifyUser: false })
    if (applied) {
      showHistoryPanel.value = false
      showSessionChoice.value = false
      notify(`已切换至会话: ${session.theme || '未命名主题'}`, 'success')
    }
  }

  const deleteSession = (sessionId) => {
    savedSessions.value = savedSessions.value.filter((session) => session.id !== sessionId)
    persistSessions()
    if (currentSessionId.value === sessionId) {
      startNewConversation()
    }
  }

  const restoreConversation = () => {
    if (savedSessions.value.length === 0) {
      loadSessionsFromStorage()
    }

    if (savedSessions.value.length > 0) {
      savedSessions.value.sort((left, right) => right.timestamp - left.timestamp)
      const latest = savedSessions.value[0]
      loadSessionById(latest.id, { updateRoute: true, notifyUser: false })
    } else {
      loadChatMessagesFromStorage()
      if (chatMessages.value.length > 1) {
        const newId = Date.now()
        currentSessionId.value = newId
        currentSessionTheme.value = '恢复的会话'
        saveCurrentSession()
        router.replace(`/feature/${newId}`)
      }
    }
    showSessionChoice.value = false
  }

  const formatTime = (timestamp) => {
    if (!timestamp) return '未知'
    return new Date(timestamp).toLocaleString('zh-CN')
  }

  return {
    clearSingleColorMode,
    startNewConversation,
    findSessionById,
    applySession,
    loadSessionById,
    loadSession,
    deleteSession,
    restoreConversation,
    formatTime
  }
}
