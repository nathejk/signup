import { createRouter, createWebHistory } from 'vue-router'
import HomeView from '../views/HomeView.vue'
import IndskrivningView from '../views/IndskrivningView.vue'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'home',
      component: HomeView
    },
    { path: '/indskrivning/:memberId', component: IndskrivningView, props: true },
    { path: '/tak', name: 'thankyou', component: () => import('../views/ThankyouView.vue') },
  ]
})

export default router
