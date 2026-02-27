import { nextTick, onMounted } from 'vue'
import { driver } from 'driver.js'
import 'driver.js/dist/driver.css'

const WIZARD_STORAGE_KEY = 'paletteflow_wizard_completed_v1'

export function useFeatureWizard() {
  const startWizard = async (force = false) => {
    if (!force && localStorage.getItem(WIZARD_STORAGE_KEY) === '1') {
      return
    }

    await nextTick()
    let firstStepCompleted = false
    let detachFirstStepClick = null
    let sendStepCompleted = false
    let detachSendStepClick = null
    let wizard

    const bindFirstStepButton = () => {
      const startButton = document.querySelector('[data-tour="chat-start"]')
      if (!startButton) return

      const handleStartClick = () => {
        if (firstStepCompleted) return
        firstStepCompleted = true
        wizard.moveNext()
      }

      startButton.addEventListener('click', handleStartClick)
      detachFirstStepClick = () => {
        startButton.removeEventListener('click', handleStartClick)
        detachFirstStepClick = null
      }
    }

    const fillPresetPrompt = (text) => {
      const inputEl = document.querySelector('[data-tour="chat-input"]')
      if (!inputEl) return
      inputEl.focus()
      inputEl.value = text
      inputEl.dispatchEvent(new Event('input', { bubbles: true }))
      inputEl.dispatchEvent(new Event('change', { bubbles: true }))
    }

    const bindSendStepButton = () => {
      const sendBtn = document.querySelector('[data-tour="send-actions"] .send-btn')
      if (!sendBtn) return

      sendStepCompleted = false
      const handleSendClick = () => {
        if (sendStepCompleted) return
        sendStepCompleted = true
        wizard.moveNext()
      }

      sendBtn.addEventListener('click', handleSendClick)
      detachSendStepClick = () => {
        sendBtn.removeEventListener('click', handleSendClick)
        detachSendStepClick = null
      }
    }

    wizard = driver({
      showProgress: true,
      allowClose: true,
      overlayClickBehavior: 'close',
      steps: [
        {
          element: '[data-tour="chat-start"]',
          onHighlighted: () => {
            bindFirstStepButton()
          },
          onDeselected: () => {
            if (detachFirstStepClick) detachFirstStepClick()
          },
          popover: {
            title: '让我们开始',
            description: '这是新建按钮，让我们创建一个新对话来开始配色之旅吧！',
            side: 'top',
            align: 'start',
            nextBtnText: '新建对话',
            onNextClick: () => {
              if (firstStepCompleted) {
                wizard.moveNext()
                return
              }

              const startButton = document.querySelector('[data-tour="chat-start"]')
              if (startButton) {
                startButton.click()
                return
              }

              firstStepCompleted = true
              wizard.moveNext()
            }
          }
        },
        {
          element: '[data-tour="chat-header"]',
          popover: {
            title: '配色对话助手',
            description: '这里是主工作区，点击右侧按钮可以新建对话与查看历史会话。',
            side: 'bottom',
            align: 'start'
          }
        },
        {
          element: '[data-tour="chat-input"]',
          popover: {
            title: '输入你的需求',
            description: '输入场景、风格或情绪描述，按 Ctrl+Enter 或点击发送。第一次发送会生成配色方案，之后可以微调颜色或数量。',
            side: 'top',
            align: 'start',
            nextBtnText: '填入示例词',
            onNextClick: () => {
              fillPresetPrompt('森林配色')
              wizard.moveNext()
            }
          }
        },
        {
          element: '[data-tour="color-count"]',
          popover: {
            title: '选择颜色数量',
            description: '支持 1 到 10 色，可先生成再精细调整。',
            side: 'top',
            align: 'start'
          }
        },
        {
          element: '[data-tour="action-row"]',
          popover: {
            title: '快捷短语',
            description: '支持查看颜色的色盲适配情况和对比度，支持重新生成配色方案。',
            side: 'top',
            align: 'start'
          }
        },
        {
          element: '[data-tour="send-actions"]',
          onHighlighted: () => {
            bindSendStepButton()
          },
          onDeselected: () => {
            if (detachSendStepClick) detachSendStepClick()
          },
          popover: {
            title: '发送与灵感',
            description: '发送会生成/微调配色；灵感按钮会自动生成艺术短句并发送。',
            side: 'top',
            align: 'start',
            nextBtnText: '发送并继续',
            onNextClick: () => {
              if (sendStepCompleted) {
                wizard.moveNext()
                return
              }

              const sendBtn = document.querySelector('[data-tour="send-actions"] .send-btn')
              if (sendBtn) {
                sendBtn.click()
                return
              }

              sendStepCompleted = true
              wizard.moveNext()
            }
          }
        },
        {
          element: '[data-tour="result-panel"]',
          popover: {
            title: '查看结果',
            description: '右侧显示配色结果与建议，可复制、微调、检查可访问性。',
            side: 'left',
            align: 'start'
          }
        },
        {
          element: '[data-tour="settings-btn"]',
          popover: {
            title: '设置与数据管理',
            description: '可在设置里查看存储与备份，保障创作数据安全。',
            side: 'bottom',
            align: 'end'
          }
        }
      ],
      onDestroyed: () => {
        if (detachFirstStepClick) detachFirstStepClick()
        if (detachSendStepClick) detachSendStepClick()
        localStorage.setItem(WIZARD_STORAGE_KEY, '1')
      }
    })

    wizard.drive()
  }

  const autoStartWizard = () => {
    onMounted(() => {
      setTimeout(() => {
        startWizard(false)
      }, 400)
    })
  }

  return {
    startWizard,
    autoStartWizard
  }
}
