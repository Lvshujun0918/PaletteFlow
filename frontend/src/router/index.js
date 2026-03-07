import { createRouter, createWebHistory } from 'vue-router'
import App from '../App.vue'
import Feature from '../pages/Feature.vue'

const WIZARD_STORAGE_KEY = 'paletteflow_wizard_completed_v3'
const isFeatureWizardCompleted = () => localStorage.getItem(WIZARD_STORAGE_KEY) === '1'

const routes = [
  { path: '/', component: App },
  { path: '/feature', component: Feature },
  { path: '/feature/:sessionId', component: Feature, props: true }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

router.beforeEach((to) => {
  if (to.path.startsWith('/feature/') && !isFeatureWizardCompleted()) {
    return { path: '/feature' }
  }

  return true
})

export default router
