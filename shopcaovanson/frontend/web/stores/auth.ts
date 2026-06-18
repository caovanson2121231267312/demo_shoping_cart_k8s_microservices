import { defineStore } from 'pinia'
import type { User, UserRole } from '~/types'
import { hasMinRole, STAFF_ROLES } from '~/utils/roles'

const REFRESH_TOKEN_KEY = 'shop_rt'

export const useAuthStore = defineStore('auth', () => {
  const user = ref<User | null>(null)
  const accessToken = ref<string | null>(null)
  const initialized = ref(false)
  const refreshTimer = ref<ReturnType<typeof setTimeout> | null>(null)

  const isLoggedIn = computed(() => !!accessToken.value)
  const isAdmin = computed(() => hasMinRole(user.value?.role ?? 'customer', 'admin'))
  const isStaff = computed(() => STAFF_ROLES.includes((user.value?.role ?? 'customer') as UserRole))
  const userRole = computed(() => user.value?.role ?? 'customer')

  function can(minRole: UserRole) {
    return hasMinRole(userRole.value, minRole)
  }

  function setTokens(access: string, refresh?: string) {
    accessToken.value = access
    if (refresh) {
      if (import.meta.client) {
        localStorage.setItem(REFRESH_TOKEN_KEY, refresh)
      }
    }
    scheduleTokenRefresh(access)
  }

  function clearTokens() {
    accessToken.value = null
    if (import.meta.client) {
      localStorage.removeItem(REFRESH_TOKEN_KEY)
    }
    clearRefreshTimer()
  }

  function setUser(value: User | null) {
    user.value = value
  }

  function getRefreshToken(): string | null {
    if (!import.meta.client) {
      return null
    }
    return localStorage.getItem(REFRESH_TOKEN_KEY)
  }

  function clearRefreshTimer() {
    if (refreshTimer.value) {
      clearTimeout(refreshTimer.value)
      refreshTimer.value = null
    }
  }

  function parseJwtExpiry(token: string): number | null {
    try {
      const payload = JSON.parse(atob(token.split('.')[1]))
      return typeof payload.exp === 'number' ? payload.exp * 1000 : null
    } catch {
      return null
    }
  }

  function scheduleTokenRefresh(token: string) {
    clearRefreshTimer()
    const expiry = parseJwtExpiry(token)
    if (!expiry) {
      return
    }
    const refreshAt = expiry - 60_000
    const delay = Math.max(refreshAt - Date.now(), 0)
    refreshTimer.value = setTimeout(async () => {
      const refreshToken = getRefreshToken()
      if (!refreshToken) {
        return
      }
      const config = useRuntimeConfig()
      try {
        const data = await $fetch<{ access_token: string }>(
          `${(config.public.apiUrl as string) || ''}/api/auth/refresh`,
          {
            method: 'POST',
            body: { refresh_token: refreshToken },
          },
        )
        accessToken.value = data.access_token
        scheduleTokenRefresh(data.access_token)
      } catch {
        clearTokens()
        user.value = null
      }
    }, delay)
  }

  function markInitialized() {
    initialized.value = true
  }

  return {
    user,
    accessToken,
    initialized,
    isLoggedIn,
    isAdmin,
    isStaff,
    userRole,
    can,
    setTokens,
    clearTokens,
    setUser,
    getRefreshToken,
    scheduleTokenRefresh,
    clearRefreshTimer,
    markInitialized,
  }
})
