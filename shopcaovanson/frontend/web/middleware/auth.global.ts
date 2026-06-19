export default defineNuxtRouteMiddleware(async (to) => {
  const auth = useAuth()
  await auth.ensureAuth()

  const publicPaths = ['/auth/login', '/auth/register', '/auth/forgot-password', '/auth/reset-password']
  if (publicPaths.some((p) => to.path.startsWith(p))) {
    if (auth.isLoggedIn.value) {
      return navigateTo('/')
    }
    return
  }

  const protectedPrefixes = ['/profile', '/checkout', '/orders', '/admin']
  const needsAuth = protectedPrefixes.some((p) => to.path.startsWith(p))

  if (needsAuth && !auth.isLoggedIn.value) {
    return navigateTo('/auth/login')
  }

  if (to.path.startsWith('/admin') && !auth.isStaff.value) {
    return navigateTo('/')
  }
})
