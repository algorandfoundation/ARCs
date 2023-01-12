import { useAuthStore } from '~~/store/auth'

export default defineNuxtRouteMiddleware(to => {
  if (process.server) {
    const authStore = useAuthStore()
    const { loadSession } = authStore
    loadSession()
    if (!authStore.session && to.path !== '/signin') {
      return '/signin'
    }
    if (authStore.session && to.path === '/signin') {
      return '/'
    }
  }
})
