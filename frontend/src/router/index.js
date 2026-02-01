import { createRouter, createWebHistory } from 'vue-router'
import App from '../App.vue'
import Feature from '../pages/Feature.vue'

const routes = [
  { path: '/', component: App },
  { path: '/feature', component: Feature }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

export default router
