import type { Session } from '@/utils/hooks/useAuth'

export default defineNuxtRouteMiddleware((to) => {
  if (import.meta.server) {
    const session = useCookie<Session>('session')
    if (!session.value && to.path !== '/signin') {
      return '/signin'
    }
    if (session.value && to.path === '/signin') {
      return '/'
    }
  }
})
