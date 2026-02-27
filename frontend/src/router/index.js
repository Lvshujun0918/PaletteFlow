import { createRouter, createWebHistory } from 'vue-router'
import App from '../App.vue'
import Feature from '../pages/Feature.vue'
import { isFeatureWizardCompleted } from '../pages/feature/wizard'

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
