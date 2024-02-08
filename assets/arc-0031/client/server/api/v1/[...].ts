import { useEnv } from '@/utils/hooks/useEnv'

export default defineEventHandler((event) => {
  const env = useEnv()
  const target = new URL(event.node.req.url as string, env.server.apiUrl)

  return proxyRequest(event, target.toString())
})
