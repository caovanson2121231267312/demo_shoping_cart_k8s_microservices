import type { FetchOptions } from 'ofetch'
import type { RefreshResponse, TokenPair, User } from '~/types'

let initPromise: Promise<void> | null = null

export const useAuth = () => {
  const authStore = useAuthStore()
  const config = useRuntimeConfig()
  const router = useRouter()

  const apiBase = computed(() => (config.public.apiUrl as string) || '')

  const authHeaders = (): Record<string, string> => {
    const headers: Record<string, string> = {}
    if (authStore.accessToken) {
      headers.Authorization = `Bearer ${authStore.accessToken}`
    }
    return headers
  }

  const apiFetch = async <T>(path: string, options: FetchOptions<'json'> = {}): Promise<T> => {
    const url = path.startsWith('http') ? path : `${apiBase.value}${path}`

    try {
      return await $fetch<T>(url, {
        ...options,
        headers: {
          ...authHeaders(),
          ...(options.headers as Record<string, string> | undefined),
        },
      })
    } catch (error: unknown) {
      const status = (error as { statusCode?: number })?.statusCode
      if (status === 401 && !path.includes('/api/auth/refresh') && !path.includes('/api/auth/login')) {
        const refreshed = await refreshAccessToken()
        if (refreshed) {
          return await $fetch<T>(url, {
            ...options,
            headers: {
              ...authHeaders(),
              ...(options.headers as Record<string, string> | undefined),
            },
          })
        }
        await logout(false)
        await router.push('/auth/login')
      }
      throw error
    }
  }

  const refreshAccessToken = async (): Promise<boolean> => {
    const refreshToken = authStore.getRefreshToken()
    if (!refreshToken) {
      return false
    }

    try {
      const data = await $fetch<RefreshResponse>(`${apiBase.value}/api/auth/refresh`, {
        method: 'POST',
        body: { refresh_token: refreshToken },
      })
      authStore.setTokens(data.access_token)
      return true
    } catch {
      return false
    }
  }

  const login = async (email: string, password: string) => {
    const data = await $fetch<TokenPair>(`${apiBase.value}/api/auth/login`, {
      method: 'POST',
      body: { email, password },
    })
    authStore.setTokens(data.access_token, data.refresh_token)
    await getCurrentUser()
  }

  const register = async (fullName: string, email: string, password: string) => {
    return await $fetch<import('~/types').RegisterResponse>(`${apiBase.value}/api/auth/register`, {
      method: 'POST',
      body: { full_name: fullName, email, password },
    })
  }

  const verifyEmail = async (token: string) => {
    const data = await $fetch<TokenPair>(`${apiBase.value}/api/auth/verify-email`, {
      method: 'POST',
      body: { token },
    })
    authStore.setTokens(data.access_token, data.refresh_token)
    await getCurrentUser()
  }

  const resendVerification = async (email: string) => {
    await $fetch(`${apiBase.value}/api/auth/resend-verification`, {
      method: 'POST',
      body: { email },
    })
  }

  const forgotPassword = async (email: string) => {
    return await $fetch<{ message: string; email?: string }>(`${apiBase.value}/api/auth/forgot-password`, {
      method: 'POST',
      body: { email },
    })
  }

  const resetPassword = async (email: string, otp: string, newPassword: string) => {
    return await $fetch<{ message: string }>(`${apiBase.value}/api/auth/reset-password`, {
      method: 'POST',
      body: { email, otp, new_password: newPassword },
    })
  }

  const logout = async (callApi = true) => {
    if (callApi && authStore.accessToken) {
      try {
        await apiFetch('/api/auth/logout', { method: 'POST' })
      } catch {
        // ignore logout errors
      }
    }
    authStore.clearTokens()
    authStore.setUser(null)
    useCartStore().hydrate()
  }

  const getCurrentUser = async () => {
    const user = await apiFetch<User>('/api/auth/me')
    authStore.setUser(user)
    return user
  }

  const updateProfile = async (fullName: string) => {
    const user = await apiFetch<User>('/api/auth/me', {
      method: 'PUT',
      body: { full_name: fullName },
    })
    authStore.setUser(user)
    return user
  }

  const ensureAuth = async () => {
    if (authStore.initialized) {
      return
    }
    if (!initPromise) {
      initPromise = (async () => {
        const refreshToken = authStore.getRefreshToken()
        if (refreshToken) {
          const ok = await refreshAccessToken()
          if (ok) {
            try {
              await getCurrentUser()
            } catch {
              authStore.clearTokens()
              authStore.setUser(null)
            }
          }
        }
        authStore.markInitialized()
      })()
    }
    await initPromise
  }

  return {
    user: computed(() => authStore.user),
    isLoggedIn: computed(() => authStore.isLoggedIn),
    isAdmin: computed(() => authStore.isAdmin),
    isStaff: computed(() => authStore.isStaff),
    initialized: computed(() => authStore.initialized),
    accessToken: computed(() => authStore.accessToken),
    apiFetch,
    login,
    logout,
    register,
    verifyEmail,
    resendVerification,
    forgotPassword,
    resetPassword,
    getCurrentUser,
    updateProfile,
    refreshAccessToken,
    ensureAuth,
  }
}
