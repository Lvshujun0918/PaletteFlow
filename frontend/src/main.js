import { createApp } from 'vue'
import App from './App.vue'
import router from './router'
import './style.css'

// Lucide Icons
import {
  Settings,
  Plus,
  Minus,
  History,
  Copy,
  Edit2,
  CheckCircle,
  Send,
  Eye,
  HardDrive,
  Package,
  Download,
  Upload,
	AlertCircle,
  Sparkles
} from 'lucide-vue-next'

const app = createApp(App)

// 注册Lucide icons为全局组件
const icons = {
  Settings,
  Plus,
  Minus,
  History,
  Copy,
  Edit2,
  CheckCircle,
  Send,
  Eye,
  HardDrive,
  Package,
  Download,
  Upload,
	AlertCircle,
  Sparkles
}

Object.entries(icons).forEach(([name, component]) => {
  app.component(`Icon${name}`, component)
})

app.use(router)
app.mount('#app')
