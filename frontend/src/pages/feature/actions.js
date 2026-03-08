import { regenerateSingleColor, refinePalette, generatePalette } from '../../utils/api'
import { nextTick } from 'vue'
import {
  getContrastRatio,
  getContrastLevel,
  simulateDeuteranopia,
  simulateProtanopia,
  simulateTritanopia,
  simulateAchromatopsia
} from '../../utils/colorUtils'
import { MAX_CHAT_HISTORY, MAX_HISTORY } from './constants'

export function createActionsApi(deps) {
  const {
    router,
    notify,
    loading,
    loadingSingle,
    currentColors,
    previousColors, // 接收 previousColors
    currentPrompt,
    currentTimestamp,
    currentAdvice,
    currentSessionId,
    currentSessionTheme,
    colorCount,
    nextSeedColors,
    histories,
    chatInput,
    chatMessages,
    showHistoryPanel,
    savedSessions,
    selectedColor1,
    selectedColor2,
    singleColorHex,
    singleColorPrompt,
    singleColorIndex,
    singleColorBase,
    singleColorMode,
    isQuickActionsOpen,
    showColorPicker,
    editingColorIndex,
    editingColorValue,
    saveHistoriesToStorage,
    saveChatMessagesToStorage,
    persistSessions,
    cloneMessages,
    clearSingleColorMode
  } = deps

  const resolveErrorMessage = (error, fallback = '操作失败，请稍后再试。') => {
    const backendMessage = error?.response?.data?.error
    if (typeof backendMessage === 'string' && backendMessage.trim()) {
      return backendMessage.trim()
    }

    if (error?.code === 'ECONNABORTED') {
      return '请求超时，已自动重试，请稍后再试。'
    }

    if (!error?.response && error?.message) {
      return '网络连接不稳定，请检查网络后重试。'
    }

    if (typeof error?.message === 'string' && error.message.trim()) {
      return error.message.trim()
    }

    return fallback
  }

  const scrollChatToLatest = async (smooth = true) => {
    await nextTick()
    const chatList = document.querySelector('.chat-messages')
    if (!chatList) return
    chatList.scrollTo({
      top: chatList.scrollHeight,
      behavior: smooth ? 'smooth' : 'auto'
    })
  }

  const addChatMessage = (role, type, content, payload = null) => {
    chatMessages.value.push({
      id: `${Date.now()}-${Math.random().toString(16).slice(2)}`,
      role,
      type,
      content,
      payload
    })

    if (chatMessages.value.length > MAX_CHAT_HISTORY) {
      chatMessages.value.splice(0, chatMessages.value.length - MAX_CHAT_HISTORY)
    }
    saveChatMessagesToStorage()
    scrollChatToLatest(true)
  }

  const handlePickColorFromChat = (palette, index) => {
    if (!palette || palette.length === 0) return
    editingColorIndex.value = index
    editingColorValue.value = palette[index]
    showColorPicker.value = true
  }

  const handleColorPickerConfirm = (newColor) => {
    const newColors = [...currentColors.value]
    if (editingColorIndex.value >= 0 && editingColorIndex.value < newColors.length) {
      newColors[editingColorIndex.value] = newColor
      
      // 保存当前配色为 previousColors（在更新前）
      previousColors.value = [...currentColors.value]
      
      currentColors.value = newColors
      addChatMessage('user', 'text', '手动调整配色方案')
      // Add assistant message
      chatMessages.value.push({
        id: Date.now(),
        role: 'assistant',
        type: 'palette',
        content: null,
        payload: {
          title: '手动调节',
          colors: newColors,
          // advice: `已手动将第 ${editingColorIndex.value + 1} 个颜色调整为 ${newColor}`
          advice: '根据您的手动调节更新了配色。'
        }
      })
      
      saveChatMessagesToStorage()
      persistSessions()
      scrollChatToLatest(true)
    }
    showColorPicker.value = false
  }

  const handlePickColorFromDisplay = (index) => {
    if (!currentColors.value || currentColors.value.length === 0) {
      return
    }
    handlePickColorFromChat([...currentColors.value], index)
  }

  const handleSelectColorForAI = (index) => {
    if (!currentColors.value || currentColors.value.length === 0) {
      notify('当前没有配色可供选择', 'error')
      return
    }
    if (index < 0 || index >= currentColors.value.length) {
      notify('颜色索引无效', 'error')
      return
    }
    singleColorHex.value = currentColors.value[index]
    singleColorIndex.value = index
    singleColorMode.value = true
    singleColorBase.value = [...currentColors.value]
    notify(`已选中颜色 ${currentColors.value[index]}，请输入调整需求`, 'info')
  }

  const handleGenerate = async (prompt, options = {}) => {
    if (loading.value || loadingSingle.value) {
      notify('正在处理中，请稍候', 'info')
      return
    }

    loading.value = true
    let isRefinement = false
    let response
    try {
      const mode = options?.mode || 'auto'
      const canRefine = currentSessionId.value && currentColors.value.length > 0
      isRefinement = mode === 'refine' ? canRefine : (mode === 'generate' ? false : canRefine)

      if (isRefinement) {
        response = await refinePalette(currentColors.value, prompt, colorCount.value)
        currentPrompt.value = prompt
      } else {
        const seedColors = Array.isArray(nextSeedColors.value) ? [...nextSeedColors.value] : []
        response = await generatePalette(prompt, colorCount.value, { seedColors })
        const newId = Date.now()
        currentSessionId.value = newId
        currentSessionTheme.value = prompt
        currentPrompt.value = prompt
        nextSeedColors.value = []

        const newSession = {
          id: newId,
          theme: prompt,
          timestamp: newId,
          colors: response.data.colors,
          prompt,
          advice: response.data.advice || '',
          messages: cloneMessages(chatMessages.value)
        }
        savedSessions.value.unshift(newSession)
        persistSessions()
        router.replace(`/feature/${newId}`)
      }

      const resultColors = Array.isArray(response?.data?.colors) ? response.data.colors : []
      if (resultColors.length === 0) {
        throw new Error('服务返回的配色结果为空')
      }

      // 保存当前配色为 previousColors（在更新前）
      previousColors.value = [...currentColors.value]

      currentColors.value = resultColors
      currentTimestamp.value = Number(response?.data?.timestamp || Date.now()) * 1000
      currentAdvice.value = response.data.advice || ''

      notify('配色生成成功！', 'success')
      addChatMessage('assistant', 'palette', '', {
        title: isRefinement ? '已修改配色' : '已生成配色',
        colors: resultColors,
        prompt,
        advice: response.data.advice || ''
      })

      try {
        const newHistory = {
          id: currentSessionId.value,
          prompt: currentSessionTheme.value,
          currentPrompt: currentPrompt.value,
          colors: resultColors,
          timestamp: response.data.timestamp,
          advice: response.data.advice || ''
        }

        if (isRefinement) {
          const index = histories.value.findIndex((history) => history.id === currentSessionId.value)
          if (index !== -1) {
            histories.value[index] = newHistory
          } else {
            histories.value.unshift(newHistory)
          }
        } else {
          histories.value.unshift(newHistory)
        }

        if (histories.value.length > MAX_HISTORY) {
          histories.value.pop()
        }

        saveHistoriesToStorage()
      } catch (postError) {
        console.error('生成成功，但本地持久化失败:', postError)
        notify('配色已生成，但本地保存失败，可继续使用。', 'warning')
      }
    } catch (error) {
      console.error('生成配色失败:', error)
      const errorMessage = resolveErrorMessage(error, '生成配色失败，请重试')
      notify(errorMessage, 'error')
      addChatMessage('assistant', 'text', `生成失败了：${errorMessage}`, {
        retryPrompt: prompt,
        retryContext: {
          type: isRefinement ? 'refine' : 'generate',
          prompt,
          seedColors: !isRefinement && Array.isArray(nextSeedColors.value) ? [...nextSeedColors.value] : []
        }
      })
    } finally {
      loading.value = false
    }
  }

  const handleRegenerate = () => {
    if (!currentColors.value || currentColors.value.length === 0) {
      notify('请先生成配色方案', 'warning')
      return
    }
    const colorsText = currentColors.value.join('、')
    const newPrompt = `对${colorsText}颜色不满意，请按照“${currentPrompt.value}”重新生成配色方案`
    addChatMessage('user', 'text', newPrompt)
    handleGenerate(newPrompt)
  }

  const handleSingleColorRegenerate = async () => {
    if (loading.value || loadingSingle.value) {
      notify('正在处理中，请稍候', 'info')
      return
    }

    if (!singleColorHex.value) {
      notify('请先从左侧选择需要替换的颜色', 'warning')
      return
    }

    if (!currentSessionId.value) {
      currentSessionId.value = Date.now()
      currentSessionTheme.value = currentSessionTheme.value || currentPrompt.value || '未命名主题'
      router.replace(`/feature/${currentSessionId.value}`)
    }

    const base = singleColorBase.value.length === currentColors.value.length
      ? singleColorBase.value
      : currentColors.value

    if (!base || base.length < 1 || base.length > 10) {
      notify('当前配色数量异常，无法进行单色微调', 'error')
      return
    }

    const targetIdx = typeof singleColorIndex.value === 'number'
      ? singleColorIndex.value
      : base.indexOf(singleColorHex.value)

    if (targetIdx < 0 || targetIdx >= base.length) {
      notify('未能确定需要替换的颜色位置', 'error')
      return
    }

    loadingSingle.value = true
    try {
      const payload = {
        prompt: singleColorPrompt.value || `为颜色 ${singleColorHex.value} 提供一个风格一致的替代色`,
        base_colors: base,
        target_index: targetIdx
      }
      const response = await regenerateSingleColor(payload)
      const resultColors = Array.isArray(response?.data?.colors) ? response.data.colors : []
      if (resultColors.length === 0) {
        throw new Error('服务返回的配色结果为空')
      }

      // 保存当前配色为 previousColors（在更新前）
      previousColors.value = [...currentColors.value]

      currentColors.value = resultColors
      currentPrompt.value = payload.prompt
      currentTimestamp.value = Number(response?.data?.timestamp || Date.now()) * 1000
      currentAdvice.value = response.data.advice || ''

      addChatMessage('assistant', 'palette', '', {
        title: '已修改配色',
        colors: resultColors,
        prompt: currentPrompt.value,
        advice: response.data.advice || ''
      })
      notify('已重生成指定颜色并更新整套配色', 'success')

      try {
        const newHistory = {
          id: currentSessionId.value,
          prompt: currentSessionTheme.value,
          currentPrompt: currentPrompt.value,
          colors: resultColors,
          timestamp: response.data.timestamp,
          advice: response.data.advice || ''
        }

        const index = histories.value.findIndex((history) => history.id === currentSessionId.value)
        if (index !== -1) {
          histories.value[index] = newHistory
        } else {
          histories.value.unshift(newHistory)
        }

        if (histories.value.length > MAX_HISTORY) {
          histories.value.pop()
        }
        saveHistoriesToStorage()
      } catch (postError) {
        console.error('单色修改成功，但本地持久化失败:', postError)
        notify('颜色已更新，但本地保存失败，可继续使用。', 'warning')
      }

      clearSingleColorMode()
    } catch (error) {
      console.error('单色重生成失败:', error)
      const errorMessage = resolveErrorMessage(error, '单色重生成失败，请重试')
      notify(errorMessage, 'error')
      addChatMessage('assistant', 'text', `生成失败了：${errorMessage}`, {
        retryContext: {
          type: 'single'
        }
      })
    } finally {
      loadingSingle.value = false
      loading.value = false
    }
  }

  const retryFailedMessage = (message) => {
    const retryContext = message?.payload?.retryContext
    const fallbackPrompt = message?.payload?.retryPrompt

    if (loading.value || loadingSingle.value) {
      return
    }

    if (retryContext?.type === 'single') {
      handleSingleColorRegenerate()
      return
    }

    const retryPrompt = retryContext?.prompt || fallbackPrompt
    if (!retryPrompt) {
      return
    }

    const mode = retryContext?.type === 'refine' ? 'refine' : 'generate'
    if (mode === 'generate' && Array.isArray(retryContext?.seedColors)) {
      nextSeedColors.value = [...retryContext.seedColors]
    }
    handleGenerate(retryPrompt, { mode })
  }

  const insertQuickInput = (text) => {
    chatInput.value = text
  }

  const toggleQuickActions = () => {
    isQuickActionsOpen.value = !isQuickActionsOpen.value
  }

  const handleShowHistory = () => {
    showHistoryPanel.value = true
  }

  const handleContrastCheck = () => {
    if (!currentColors.value || currentColors.value.length < 2) {
      notify('当前颜色不足，无法进行对比度检查', 'warning')
      return
    }

    const results = []
    for (let i = 0; i < currentColors.value.length; i += 1) {
      for (let j = i + 1; j < currentColors.value.length; j += 1) {
        const color1 = currentColors.value[i]
        const color2 = currentColors.value[j]
        const ratio = getContrastRatio(color1, color2)
        results.push({
          color1,
          color2,
          ratio,
          level: getContrastLevel(ratio),
          score: (ratio / 21) * 100
        })
      }
    }

    if (results.length === 0) {
      notify('未找到可检测的颜色组合', 'warning')
      return
    }

    results.sort((left, right) => left.ratio - right.ratio)
    const minResult = results[0]
    const passCount = results.filter((item) => item.ratio >= 4.5).length

    addChatMessage('assistant', 'contrast', '', {
      results,
      totalPairs: results.length,
      passCount,
      minRatio: minResult.ratio,
      minLevel: minResult.level
    })
  }

  const getMinContrast = (palette) => {
    if (!palette || palette.length < 2) return 0
    let min = Infinity
    for (let i = 0; i < palette.length; i += 1) {
      for (let j = i + 1; j < palette.length; j += 1) {
        const ratio = getContrastRatio(palette[i], palette[j])
        min = Math.min(min, ratio)
      }
    }
    return min === Infinity ? 0 : min
  }

  const buildRecommendations = (minContrast) => {
    const recommendations = []
    if (minContrast < 4.5) {
      recommendations.push('提高明度差或增加饱和度对比')
      recommendations.push('避免相近色相的组合，拉开色相距离')
    }
    if (minContrast < 3) {
      recommendations.push('优先使用高对比度的浅色与深色搭配')
    }
    if (recommendations.length === 0) {
      recommendations.push('当前配色对色盲用户较友好，可继续使用')
    }
    return recommendations
  }

  const handleColorblindCheck = () => {
    if (!currentColors.value || currentColors.value.length === 0) {
      notify('请先生成配色方案', 'warning')
      return
    }
    const deuteranopia = currentColors.value.map(simulateDeuteranopia)
    const protanopia = currentColors.value.map(simulateProtanopia)
    const tritanopia = currentColors.value.map(simulateTritanopia)
    const achromatopsia = currentColors.value.map(simulateAchromatopsia)
    const minContrast = Math.min(
      getMinContrast(deuteranopia),
      getMinContrast(protanopia),
      getMinContrast(tritanopia),
      getMinContrast(achromatopsia)
    )

    addChatMessage('assistant', 'colorblind', '', {
      deuteranopia,
      protanopia,
      tritanopia,
      achromatopsia,
      isAccessible: minContrast >= 4.5,
      recommendations: buildRecommendations(minContrast)
    })
  }

  const handleRestoreToMessage = (messageIndex) => {
    if (!Array.isArray(chatMessages.value) || messageIndex < 0 || messageIndex >= chatMessages.value.length) {
      return
    }

    const targetMessage = chatMessages.value[messageIndex]
    if (!targetMessage || targetMessage.type !== 'palette' || !targetMessage.payload?.colors) {
      return
    }

    // 查找目标消息之前的最后一条palette消息作为previousColors
    let prevPaletteMessage = null
    for (let i = messageIndex - 1; i >= 0; i--) {
      if (chatMessages.value[i].type === 'palette' && chatMessages.value[i].payload?.colors) {
        prevPaletteMessage = chatMessages.value[i]
        break
      }
    }

    // 设置previousColors为前一条palette的颜色，如果没有则清空
    previousColors.value = prevPaletteMessage ? [...prevPaletteMessage.payload.colors] : []

    currentColors.value = [...targetMessage.payload.colors]
    currentPrompt.value = targetMessage.payload.prompt || currentPrompt.value
    currentAdvice.value = targetMessage.payload.advice || ''
    currentTimestamp.value = Date.now()

    chatMessages.value = chatMessages.value.slice(0, messageIndex + 1)
    saveChatMessagesToStorage()
    persistSessions()
    notify('已还原到该配色节点', 'success')
  }

  const processPrompt = (prompt) => {
    if (loading.value || loadingSingle.value) {
      notify('正在处理中，请稍候', 'info')
      return
    }

    if (singleColorMode.value && singleColorHex.value) {
      addChatMessage('user', 'text', prompt)
      singleColorPrompt.value = prompt
      handleSingleColorRegenerate()
      return
    }

    if (prompt.includes('不满意')) {
      handleRegenerate()
      return
    }
    addChatMessage('user', 'text', prompt)
    if (prompt.includes('查看历史')) {
      handleShowHistory()
      return
    }
    if (prompt.includes('对比度检查')) {
      handleContrastCheck()
      return
    }
    if (prompt.includes('色盲检查')) {
      handleColorblindCheck()
      return
    }
    handleGenerate(prompt)
  }

  const handleSendPrompt = () => {
    const prompt = chatInput.value.trim()
    if (!prompt) return
    chatInput.value = ''
    processPrompt(prompt)
  }

  const sendQuickPrompt = (text) => {
    const prompt = typeof text === 'string' ? text.trim() : ''
    if (!prompt) return
    processPrompt(prompt)
  }

  return {
    addChatMessage,
    handlePickColorFromChat,
    handlePickColorFromDisplay,
    handleSelectColorForAI,
    handleColorPickerConfirm,
    handleGenerate,
    handleRegenerate,
    handleSingleColorRegenerate,
    insertQuickInput,
    sendQuickPrompt,
    retryFailedMessage,
    toggleQuickActions,
    handleSendPrompt,
    handleShowHistory,
    handleContrastCheck,
    handleColorblindCheck,
    handleRestoreToMessage
  }
}
