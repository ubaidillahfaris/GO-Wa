import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/login',
      name: 'login',
      component: () => import('@/views/auth/LoginView.vue'),
      meta: { requiresGuest: true },
    },
    {
      path: '/register',
      name: 'register',
      component: () => import('@/views/auth/RegisterView.vue'),
      meta: { requiresGuest: true },
    },
    {
      path: '/',
      component: () => import('@/views/layout/DashboardLayout.vue'),
      meta: { requiresAuth: true },
      children: [
        {
          path: '',
          name: 'dashboard',
          component: () => import('@/views/dashboard/DashboardView.vue'),
        },
        {
          path: 'devices',
          name: 'devices',
          component: () => import('@/views/devices/DeviceListView.vue'),
        },
        {
          path: 'messages',
          name: 'messages',
          component: () => import('@/views/messages/SendMessageView.vue'),
        },
        {
          path: 'api-keys',
          name: 'api-keys',
          component: () => import('@/views/api-keys/ApiKeyListView.vue'),
        },
      ],
    },
  ],
})

// Navigation guards
router.beforeEach((to, _, next) => {
  const authStore = useAuthStore()

  console.log('Router guard:', {
    to: to.name,
    requiresAuth: to.meta.requiresAuth,
    requiresGuest: to.meta.requiresGuest,
    isAuthenticated: authStore.isAuthenticated,
    hasToken: !!authStore.token,
    hasUser: !!authStore.user,
  })

  if (to.meta.requiresAuth && !authStore.isAuthenticated) {
    console.log('Redirecting to login (not authenticated)')
    next({ name: 'login' })
  } else if (to.meta.requiresGuest && authStore.isAuthenticated) {
    console.log('Redirecting to dashboard (already authenticated)')
    next({ name: 'dashboard' })
  } else {
    console.log('Allowing navigation')
    next()
  }
})

export default router
