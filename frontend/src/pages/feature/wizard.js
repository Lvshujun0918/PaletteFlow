import { nextTick, onMounted } from 'vue'
import { driver } from 'driver.js'
import 'driver.js/dist/driver.css'

const WIZARD_STORAGE_KEY = 'paletteflow_wizard_completed_v3'

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
      const sendBtn = document.querySelector('[data-tour="send-actions"]')
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
          const sendBtn = document.querySelector('[data-tour="send-actions"]')
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

    const waitForElement = (selector, timeoutMs = 5000) => {
      const startedAt = Date.now()
      return new Promise((resolve) => {
        const check = () => {
          const el = document.querySelector(selector)
          if (el) {
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

    const firstStepReady = await waitForElement('[data-tour="chat-start"]', 15000)
    if (!firstStepReady) {
      console.warn('[wizard] first step element not ready: [data-tour="chat-start"]')
      return
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
            title: '从新建开始',
            description: '点击按钮，让我们开启便捷配色之旅吧！',
            side: 'top',
            align: 'start',
            nextBtnText: '打开新建流程',
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

              waitForElement('[data-tour="chat-start"]', 5000).then((ready) => {
                if (!ready || firstStepCompleted) return
                const delayedButton = document.querySelector('[data-tour="chat-start"]')
                if (delayedButton) {
                  delayedButton.click()
                }
              })
            }
          }
        },
        {
          element: '[data-tour="new-chat-count"]',
          popover: {
            title: '新建配色',
            description: '在新建配色中，你可以设置配色数量，并预览固定色。',
            side: 'bottom',
            align: 'start'
          }
        },
        {
          element: '[data-tour="new-chat-preview"]',
          popover: {
            title: '固定主色',
            description: '点击预览格可设置固定色。已固定的颜色会在首轮生成中保持不变，并作为重点参考。',
            side: 'bottom',
            align: 'start'
          }
        },
        {
          element: '[data-tour="new-chat-confirm"]',
          popover: {
            title: '确认并创建新一轮',
            description: '确认后会创建新会话，并在聊天区自动生成“颜色数量 + 固定色预览”的助手摘要。',
            side: 'top',
            align: 'end',
            nextBtnText: '确认并继续',
            onNextClick: () => {
              const confirmBtn = document.querySelector('[data-tour="new-chat-confirm"]')
              if (confirmBtn) {
                confirmBtn.click()
              }

              waitForElement('[data-tour="seed-summary"]', 6000).then(() => {
                wizard.moveNext()
              })
            }
          }
        },
        {
          element: '[data-tour="new-chat"]',
          popover: {
            title: '新建配色对话',
            description: '单击此处可随时新建配色对话，开启新的创作灵感！',
            side: 'top',
            align: 'start'
          }
        },
        {
          element: '[data-tour="old-chat"]',
          popover: {
            title: '查看历史对话',
            description: '单击此处可查看历史对话，查看更多历史配色。',
            side: 'top',
            align: 'start'
          }
        },
        {
          element: '[data-tour="seed-summary"]',
          popover: {
            title: '查看摘要',
            description: '此处显示本轮配色配置，方便你随时回看当前颜色数量和固定色。',
            side: 'top',
            align: 'start'
          }
        },
        {
          element: '[data-tour="chat-input"]',
          popover: {
            title: '输入本轮需求',
            description: '描述场景、风格或情绪后发送，将会自动生成你想要的配色方案！',
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
          element: '[data-tour="inspiration-action"]',
          popover: {
            title: '灵感模式',
            description: '如果一时想不到具体需求，可以点击“灵感”按钮获取一句古诗词，激发你的创作灵感！',
            side: 'right',
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
            title: '发送生成首轮配色',
            description: '点击发送后将基于你的需求生成配色结果，并保持已固定颜色不变。',
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

              const sendBtn = document.querySelector('[data-tour="send-actions"]')
              if (sendBtn) {
                sendBtn.click()
                sendStepWaiting = true
                waitForSendCompletion().then((completed) => {
                  sendStepWaiting = false
                  if (!completed || sendStepCompleted) return
                  sendStepCompleted = true
                  wizard.moveNext()
                })
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
            title: '查看生成结果',
            description: '右侧面板展示当前配色、建议与时间信息，是后续迭代与导出的核心区域。',
            side: 'left',
            align: 'start'
          }
        },
        {
          element: '[data-tour="apply-image-btn"]',
          popover: {
            title: '一键套色到图片',
            description: '上传图片后，当前配色方案将套用到图片上，并且可以有现实到艺术三挡风格可以选择，帮你快速预览配色效果！',
            side: 'top',
            align: 'end'
          }
        },
        {
          element: '[data-tour="color-actions"]',
          popover: {
            title: '体验单色修改',
            description: '现在自动选中第一种颜色，进入 AI 单色修改模式。',
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
            title: '输入单色需求',
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
            title: '提交单色修改',
            description: '点击发送后，AI 将仅针对选中颜色进行替换并更新整套配色。',
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

              const sendBtn = document.querySelector('[data-tour="send-actions"]')
              if (sendBtn) {
                sendBtn.click()
                sendStepWaiting = true
                waitForSendCompletion().then((completed) => {
                  sendStepWaiting = false
                  if (!completed || sendStepCompleted) return
                  sendStepCompleted = true
                  wizard.moveNext()
                })
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
          element: '[data-tour="color-diff"]',
          popover: {
            title: '显示颜色差异',
            description: '颜色修改将以HSL对比形式呈现，方便你对比修改前后的差异。',
            side: 'bottom',
            align: 'start'
          }
        },
        {
          element: '[data-tour="settings-btn"]',
          popover: {
            title: '设置与数据管理',
            description: '可在设置里查看存储、备份与项目信息，保障创作数据安全。',
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
