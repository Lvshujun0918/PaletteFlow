import { regenerateSingleColor, refinePalette, generatePalette } from '../../utils/api'
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
    currentPrompt,
    currentTimestamp,
    currentAdvice,
    currentSessionId,
    currentSessionTheme,
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
      currentColors.value = newColors

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

  const handleGenerate = async (prompt) => {
    loading.value = true
    try {
      let response
      let isRefinement = false

      if (currentSessionId.value && currentColors.value.length === 5) {
        isRefinement = true
        response = await refinePalette(currentColors.value, prompt)
        currentPrompt.value = prompt
      } else {
        response = await generatePalette(prompt)
        const newId = Date.now()
        currentSessionId.value = newId
        currentSessionTheme.value = prompt
        currentPrompt.value = prompt

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

      currentColors.value = response.data.colors
      currentTimestamp.value = response.data.timestamp * 1000
      currentAdvice.value = response.data.advice || ''

      const newHistory = {
        id: currentSessionId.value,
        prompt: currentSessionTheme.value,
        currentPrompt: currentPrompt.value,
        colors: response.data.colors,
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

      notify('配色生成成功！', 'success')
      addChatMessage('assistant', 'palette', '', {
        title: isRefinement ? '已修改配色' : '已生成配色',
        colors: response.data.colors,
        prompt,
        advice: response.data.advice || ''
      })
    } catch (error) {
      console.error('生成配色失败:', error)
      notify('生成配色失败，请重试', 'error')
      addChatMessage('assistant', 'text', '生成失败了，请稍后再试。')
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
    const newPrompt = `对${colorsText}颜色不满意，请按照${currentPrompt.value}重新生成配色方案`
    addChatMessage('user', 'text', newPrompt)
    handleGenerate(newPrompt)
  }

  const handleSingleColorRegenerate = async () => {
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

    if (!base || base.length !== 5) {
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
      currentColors.value = response.data.colors
      currentPrompt.value = payload.prompt
      currentTimestamp.value = response.data.timestamp * 1000
      currentAdvice.value = response.data.advice || ''

      const newHistory = {
        id: currentSessionId.value,
        prompt: currentSessionTheme.value,
        currentPrompt: currentPrompt.value,
        colors: response.data.colors,
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

      addChatMessage('assistant', 'palette', '', {
        title: '已修改配色',
        colors: response.data.colors,
        prompt: currentPrompt.value,
        advice: response.data.advice || ''
      })
      notify('已重生成指定颜色并更新整套配色', 'success')
      clearSingleColorMode()
    } catch (error) {
      console.error('单色重生成失败:', error)
      notify('单色重生成失败，请重试', 'error')
    } finally {
      loadingSingle.value = false
    }
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

  const handleSendPrompt = () => {
    const prompt = chatInput.value.trim()
    if (!prompt) return

    if (singleColorMode.value && singleColorHex.value) {
      addChatMessage('user', 'text', prompt)
      singleColorPrompt.value = prompt
      chatInput.value = ''
      handleSingleColorRegenerate()
      return
    }

    addChatMessage('user', 'text', prompt)
    chatInput.value = ''

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
    if (prompt.includes('不满意')) {
      handleRegenerate()
      return
    }
    handleGenerate(prompt)
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
    toggleQuickActions,
    handleSendPrompt,
    handleShowHistory,
    handleContrastCheck,
    handleColorblindCheck
  }
}
