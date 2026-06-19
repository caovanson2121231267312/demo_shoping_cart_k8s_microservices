export default defineNuxtPlugin(() => {
  if (import.meta.server) return

  const auth = useAuthStore()
  const route = useRoute()
  const analytics = useAnalytics()

  const ping = () => {
    if (!auth.isLoggedIn) return
    analytics.sendPresence(route.path).catch(() => {})
  }

  ping()
  const timer = setInterval(ping, 60_000)

  watch(() => route.path, ping)

  if (import.meta.hot) {
    import.meta.hot.dispose(() => clearInterval(timer))
  }
})
