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
    const wizard = driver({
      showProgress: true,
      allowClose: true,
      overlayClickBehavior: 'close',
      steps: [
        {
          element: '[data-tour="chat-header"]',
          popover: {
            title: '配色对话助手',
            description: '这里是主工作区，可新建对话与查看历史会话。',
            side: 'bottom',
            align: 'start'
          }
        },
        {
          element: '[data-tour="chat-input"]',
          popover: {
            title: '输入你的需求',
            description: '输入场景、风格或情绪描述，按 Ctrl+Enter 或点击发送。',
            side: 'top',
            align: 'start'
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
          element: '[data-tour="send-actions"]',
          popover: {
            title: '发送与灵感',
            description: '发送会生成/微调配色；灵感按钮会自动生成艺术短句并发送。',
            side: 'top',
            align: 'start'
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
