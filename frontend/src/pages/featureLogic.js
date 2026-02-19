import { ref, onMounted, onBeforeUnmount, computed, watch } from 'vue'
import { useRoute, useRouter, onBeforeRouteLeave } from 'vue-router'
import { healthCheck } from '../utils/api'
import { notify } from '../utils/notify'
import {
  DEFAULT_COLORS,
  COLORBLIND_TYPES,
  createWelcomeMessage
} from './feature/constants'
import { createStorageApi } from './feature/storage'
import { createSessionApi } from './feature/session'
import { createActionsApi } from './feature/actions'

export function useFeatureLogic() {
  const router = useRouter()
  const route = useRoute()

  const loading = ref(false)
  const currentColors = ref([...DEFAULT_COLORS])
  const currentPrompt = ref('默认配色方案')
  const currentTimestamp = ref(Date.now())
  const currentAdvice = ref('')
  const currentSessionId = ref(null)
  const currentSessionTheme = ref('')
  const histories = ref([])
  const chatInput = ref('')
  const chatMessages = ref([createWelcomeMessage()])
  const showSessionChoice = ref(false)
  const showNewConversationConfirm = ref(false)
  const showHistoryPanel = ref(false)
  const savedSessions = ref([])
  const selectedColor1 = ref('')
  const selectedColor2 = ref('')
  const singleColorHex = ref('')
  const singleColorPrompt = ref('')
  const singleColorIndex = ref(0)
  const singleColorBase = ref([])
  const loadingSingle = ref(false)
  const singleColorMode = ref(false)
  const isQuickActionsOpen = ref(true)
  const showColorPicker = ref(false)
  const editingColorIndex = ref(0)
  const editingColorValue = ref('#000000')
  const colorblindTypes = COLORBLIND_TYPES

  const currentBackground = computed(() => {
    if (!currentColors.value || currentColors.value.length === 0) {
      return 'linear-gradient(135deg, rgb(255, 194, 194), rgb(255, 224, 194), rgb(254, 255, 214), rgb(217, 255, 204), rgb(185, 249, 255))'
    }
    return `linear-gradient(135deg, ${currentColors.value.join(', ')})`
  })

  const cloneMessages = (messages) => {
    if (!Array.isArray(messages)) return []
    try {
      return JSON.parse(JSON.stringify(messages))
    } catch (error) {
      console.error('复制对话记录失败:', error)
      return [...messages]
    }
  }

  const storageApi = createStorageApi({
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
  })

  const sessionApi = createSessionApi({
    router,
    notify,
    currentColors,
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
    loadSessionsFromStorage: storageApi.loadSessionsFromStorage,
    loadChatMessagesFromStorage: storageApi.loadChatMessagesFromStorage,
    saveChatMessagesToStorage: storageApi.saveChatMessagesToStorage,
    saveCurrentSession: storageApi.saveCurrentSession,
    persistSessions: storageApi.persistSessions
  })

  const actionsApi = createActionsApi({
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
    saveHistoriesToStorage: storageApi.saveHistoriesToStorage,
    saveChatMessagesToStorage: storageApi.saveChatMessagesToStorage,
    persistSessions: storageApi.persistSessions,
    cloneMessages,
    clearSingleColorMode: sessionApi.clearSingleColorMode
  })

  watch(
    () => route.params.sessionId,
    (sessionId) => {
      if (!sessionId) return
      const applied = sessionApi.loadSessionById(sessionId, { updateRoute: false, notifyUser: false })
      if (applied) {
        showHistoryPanel.value = false
      }
    }
  )

  const persistBeforeLeave = () => {
    try {
      storageApi.saveCurrentSession()
      storageApi.saveHistoriesToStorage()
    } catch (error) {
      console.error('离开页面前保存失败:', error)
    }
  }

  const handleBeforeUnload = () => {
    persistBeforeLeave()
  }

  const handlePageHide = () => {
    persistBeforeLeave()
  }

  onBeforeRouteLeave(() => {
    persistBeforeLeave()
  })

  onMounted(async () => {
    window.addEventListener('beforeunload', handleBeforeUnload)
    window.addEventListener('pagehide', handlePageHide)

    try {
      await healthCheck()
      notify('连接到服务器成功', 'success')
    } catch (error) {
      console.error('服务器连接失败:', error)
      notify('无法连接到服务器，请确保后端已启动', 'error')
    }

    storageApi.loadSessionsFromStorage()

    const routeSessionId = route.params.sessionId
    if (routeSessionId) {
      const applied = sessionApi.loadSessionById(routeSessionId, { updateRoute: false, notifyUser: true })
      if (!applied) {
        sessionApi.startNewConversation()
        router.replace('/feature')
      }
    } else {
      storageApi.loadHistoriesFromStorage()
      const storedChat = storageApi.getStoredChatMessages()
      if (storedChat.length > 1) {
        showSessionChoice.value = true
      } else {
        sessionApi.startNewConversation()
      }
    }

    if (currentColors.value && currentColors.value.length > 0) {
      selectedColor1.value = currentColors.value[0]
      selectedColor2.value = currentColors.value[1] || currentColors.value[0]
    }
  })

  onBeforeUnmount(() => {
    window.removeEventListener('beforeunload', handleBeforeUnload)
    window.removeEventListener('pagehide', handlePageHide)
    persistBeforeLeave()
  })

  const handleLogoError = (event) => {
    if (event?.target) {
      event.target.style.display = 'none'
    }
  }

  const confirmStartNewConversation = () => {
    showNewConversationConfirm.value = true
  }

  const cancelStartNewConversation = () => {
    showNewConversationConfirm.value = false
  }

  const proceedStartNewConversation = () => {
    showNewConversationConfirm.value = false
    sessionApi.startNewConversation()
  }

  return {
    loading,
    showSessionChoice,
    showNewConversationConfirm,
    showHistoryPanel,
    savedSessions,
    clearSingleColorMode: sessionApi.clearSingleColorMode,
    startNewConversation: sessionApi.startNewConversation,
    confirmStartNewConversation,
    cancelStartNewConversation,
    proceedStartNewConversation,
    restoreConversation: sessionApi.restoreConversation,
    loadSession: sessionApi.loadSession,
    deleteSession: sessionApi.deleteSession,
    currentColors,
    currentPrompt,
    currentBackground,
    currentTimestamp,
    currentAdvice,
    currentSessionTheme,
    histories,
    chatInput,
    chatMessages,
    selectedColor1,
    selectedColor2,
    singleColorHex,
    singleColorPrompt,
    loadingSingle,
    isQuickActionsOpen,
    colorblindTypes,
    showColorPicker,
    editingColorValue,
    handleGenerate: actionsApi.handleGenerate,
    handleRegenerate: actionsApi.handleRegenerate,
    handleSingleColorRegenerate: actionsApi.handleSingleColorRegenerate,
    handleSendPrompt: actionsApi.handleSendPrompt,
    handlePickColorFromChat: actionsApi.handlePickColorFromChat,
    handlePickColorFromDisplay: actionsApi.handlePickColorFromDisplay,
    handleSelectColorForAI: actionsApi.handleSelectColorForAI,
    insertQuickInput: actionsApi.insertQuickInput,
    toggleQuickActions: actionsApi.toggleQuickActions,
    handleShowHistory: actionsApi.handleShowHistory,
    handleContrastCheck: actionsApi.handleContrastCheck,
    handleColorblindCheck: actionsApi.handleColorblindCheck,
    handleColorPickerConfirm: actionsApi.handleColorPickerConfirm,
    formatTime: sessionApi.formatTime,
    handleLogoError,
    notify
  }
}
