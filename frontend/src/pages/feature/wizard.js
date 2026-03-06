import { nextTick, onMounted } from 'vue'
import { driver } from 'driver.js'
import 'driver.js/dist/driver.css'

const WIZARD_STORAGE_KEY = 'paletteflow_wizard_completed_v1'

export function isFeatureWizardCompleted() {
  return localStorage.getItem(WIZARD_STORAGE_KEY) === '1'
}

export function useFeatureWizard() {
  const startWizard = async (force = false) => {
    if (!force && isFeatureWizardCompleted()) {
      return
    }

    await nextTick()
    let firstStepCompleted = false
    let detachFirstStepClick = null
    let sendStepCompleted = false
    let sendStepWaiting = false
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

    const selectFirstColorForAI = () => {
      const firstAiBtn = document.querySelector('[data-tour="color-actions"] .pick-btn')
      if (!firstAiBtn) return false
      firstAiBtn.click()
      return true
    }

    const bindSendStepButton = () => {
      const sendBtn = document.querySelector('[data-tour="send-actions"] .send-btn')
      if (!sendBtn) return

      sendStepCompleted = false
      sendStepWaiting = false
      const handleSendClick = async () => {
        if (sendStepCompleted || sendStepWaiting) return
        sendStepWaiting = true

        const completed = await waitForSendCompletion()
        sendStepWaiting = false
        if (!completed || sendStepCompleted) return

        sendStepCompleted = true
        wizard.moveNext()
      }

      sendBtn.addEventListener('click', handleSendClick)
      detachSendStepClick = () => {
        sendBtn.removeEventListener('click', handleSendClick)
        detachSendStepClick = null
      }
    }

    const waitForSendCompletion = (timeoutMs = 25000) => {
      const startedAt = Date.now()
      let sawLoading = false
      let sawInputConsumed = false
      const inputEl = document.querySelector('[data-tour="chat-input"]')
      const initialValue = inputEl ? (inputEl.value || '') : ''

      return new Promise((resolve) => {
        const check = () => {
          const sendBtn = document.querySelector('[data-tour="send-actions"] .send-btn')
          const currentInputEl = document.querySelector('[data-tour="chat-input"]')
          const currentValue = currentInputEl ? (currentInputEl.value || '') : ''
          if (!sendBtn) {
            if (Date.now() - startedAt > timeoutMs) {
              resolve(false)
              return
            }
            setTimeout(check, 120)
            return
          }

          const isLoading = sendBtn.classList.contains('is-loading') || !!sendBtn.querySelector('.glass-button__spinner')
          if (isLoading) {
            sawLoading = true
          }

          if (!sawInputConsumed && initialValue && currentValue !== initialValue) {
            sawInputConsumed = true
          }

          if (sawLoading && !isLoading) {
            resolve(true)
            return
          }

          // 兼容“秒回”场景：请求很快完成，可能来不及观察到 loading
          if (!sawLoading && sawInputConsumed && !isLoading && Date.now() - startedAt > 300) {
            resolve(true)
            return
          }

          if (Date.now() - startedAt > timeoutMs) {
            resolve(false)
            return
          }

          setTimeout(check, 120)
        }

        check()
      })
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
            description: '这里是主工作区，你可以在此查看主题、输入需求并对配色结果进行持续调整。',
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
            description: '点击发送按钮将使用你的配色需求生成配色方案。如果不清楚具体需求，点击灵感按钮将从古诗词中随机挑选一句发送！',
            side: 'top',
            align: 'start',
            nextBtnText: '发送并继续',
            onNextClick: () => {
              if (sendStepCompleted) {
                wizard.moveNext()
                return
              }

              if (sendStepWaiting) {
                return
              }

              const sendBtn = document.querySelector('[data-tour="send-actions"] .send-btn')
              if (sendBtn) {
                sendBtn.click()
                return
              }

              sendStepWaiting = true
              waitForSendCompletion().then((completed) => {
                sendStepWaiting = false
                if (!completed || sendStepCompleted) return
                sendStepCompleted = true
                wizard.moveNext()
              })
            }
          }
        },
        {
          element: '[data-tour="result-panel"]',
          popover: {
            title: '查看结果',
            description: '右侧显示配色结果与建议，使用建议中的对应颜色点击可高亮显示。',
            side: 'left',
            align: 'start'
          }
        },
        {
          element: '[data-tour="color-actions"]',
          popover: {
            title: '调整与复制',
            description: '此处可使用AI或手动调整颜色，点击复制按钮可复制颜色代码。',
            side: 'bottom',
            align: 'start'
          }
        },
        {
          element: '[data-tour="color-actions"]',
          popover: {
            title: '体验AI单色调整',
            description: '现在自动选中第一种颜色进入AI单色调整模式。',
            side: 'bottom',
            align: 'start',
            nextBtnText: '选中第一个颜色',
            onNextClick: () => {
              selectFirstColorForAI()
              setTimeout(() => {
                wizard.moveNext()
              }, 120)
            }
          }
        },
        {
          element: '[data-tour="chat-input"]',
          popover: {
            title: '输入单色调整需求',
            description: '向输入框自动填入“改成蓝色”，并提交演示请求。',
            side: 'top',
            align: 'start',
            nextBtnText: '填入改成蓝色',
            onNextClick: () => {
              fillPresetPrompt('改成蓝色')
              wizard.moveNext()
            }
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
            title: '提交单色调整',
            description: '点击提交后，AI会根据需求调整配色方案中选中的颜色。',
            side: 'top',
            align: 'start',
            nextBtnText: '发送并继续',
            onNextClick: () => {
              if (sendStepCompleted) {
                wizard.moveNext()
                return
              }

              if (sendStepWaiting) {
                return
              }

              const sendBtn = document.querySelector('[data-tour="send-actions"] .send-btn')
              if (sendBtn) {
                sendBtn.click()
                return
              }

              sendStepWaiting = true
              waitForSendCompletion().then((completed) => {
                sendStepWaiting = false
                if (!completed || sendStepCompleted) return
                sendStepCompleted = true
                wizard.moveNext()
              })
            }
          }
        },
        {
          element: '[data-tour="result-panel"]',
          popover: {
            title: '查看单色调整结果',
            description: '已完成AI单色调整演示，你可以继续手动编辑或再次发起微调。',
            side: 'left',
            align: 'start'
          }
        },
        {
          element: '[data-tour="color-diff"]',
          popover: {
            title: '显示详细颜色差异',
            description: '通过HSL显示颜色差异，可方便地查看颜色前后变化。',
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
