import type {
  AnalyticsOverview,
  AnalyticsPeriod,
  AnalyticsSeriesPoint,
  OnlineUser,
  TopProductRow,
} from '~/types'

export const useAnalytics = () => {
  const { apiFetch } = useAuth()
  const authStore = useAuthStore()
  const config = useRuntimeConfig()

  const fetchOverview = (period: AnalyticsPeriod) =>
    apiFetch<AnalyticsOverview>(`/api/admin/analytics/overview`, { query: { period } })

  const fetchRevenue = (period: AnalyticsPeriod) =>
    apiFetch<{ period: string; series: AnalyticsSeriesPoint[] }>(`/api/admin/analytics/revenue`, { query: { period } })

  const fetchUsers = (period: AnalyticsPeriod) =>
    apiFetch<{ period: string; series: AnalyticsSeriesPoint[] }>(`/api/admin/analytics/users`, { query: { period } })

  const fetchReviews = (period: AnalyticsPeriod) =>
    apiFetch<{ period: string; series: AnalyticsSeriesPoint[] }>(`/api/admin/analytics/reviews`, { query: { period } })

  const fetchTopProducts = (period: AnalyticsPeriod, limit = 10) =>
    apiFetch<{ period: string; items: TopProductRow[] }>(`/api/admin/analytics/products/top`, { query: { period, limit } })

  const fetchOnlineUsers = () =>
    apiFetch<{ count: number; users: OnlineUser[] }>(`/api/admin/analytics/online`)

  const sendPresence = (page: string) =>
    apiFetch('/api/analytics/presence', { method: 'POST', body: { page } })

  const exportExcel = async (period: AnalyticsPeriod) => {
    const base = (config.public.apiUrl as string) || ''
    const url = `${base}/api/admin/analytics/export/excel?period=${period}`
    const res = await fetch(url, {
      headers: { Authorization: `Bearer ${authStore.accessToken}` },
    })
    if (!res.ok) throw new Error('Xuất Excel thất bại')
    const blob = await res.blob()
    const a = document.createElement('a')
    a.href = URL.createObjectURL(blob)
    a.download = `bao-cao-${period}.xlsx`
    a.click()
    URL.revokeObjectURL(a.href)
  }

  return {
    fetchOverview,
    fetchRevenue,
    fetchUsers,
    fetchReviews,
    fetchTopProducts,
    fetchOnlineUsers,
    sendPresence,
    exportExcel,
  }
}
