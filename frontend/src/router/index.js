import { createRouter, createWebHistory } from 'vue-router'
import App from '../App.vue'
import Feature from '../pages/Feature.vue'
import FontDemo from '../components/FontDemo.vue'

const routes = [
  { path: '/', component: App },
  { path: '/feature', component: Feature }
  ,{ path: '/font-demo', component: FontDemo }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

export default router