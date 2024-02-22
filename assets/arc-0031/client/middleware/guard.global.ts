import { useSessionCookie } from '@/utils/hooks/useSessionCookie'

export default defineNuxtRouteMiddleware((to) => {
  if (import.meta.server) {
    const { session } = useSessionCookie()
    if (!session.value && to.path !== '/signin') {
      return '/signin'
    }
    if (session.value && to.path === '/signin') {
      return '/'
    }
  }
})
