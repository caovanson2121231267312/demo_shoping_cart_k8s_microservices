export default defineEventHandler(async (event) => {
  const config = useRuntimeConfig()
  const target = config.apiProxyTarget as string
  const path = event.path.replace(/^\/api/, '')
  const url = `${target}/api${path}`

  return proxyRequest(event, url)
})
