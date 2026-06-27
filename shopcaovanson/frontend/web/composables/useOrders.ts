import type { Order, OrderListResult, TrackOrderInput } from '~/types'

export interface OrderExportJob {
  job_id: string
  status: 'pending' | 'processing' | 'completed' | 'failed'
  created_at?: string
  completed_at?: string | null
  error?: string | null
  file_name?: string | null
  filters?: Record<string, string>
}

export const useOrders = () => {
  const { apiFetch } = useAuth()
  const authStore = useAuthStore()
  const config = useRuntimeConfig()

  const fetchMyOrders = (page = 1, limit = 20) =>
    apiFetch<OrderListResult>('/api/orders', { query: { page, limit } })

  const fetchOrder = (id: string) => apiFetch<Order>(`/api/orders/${id}`)

  const trackOrder = (input: TrackOrderInput) =>
    $fetch<Order>('/api/orders/track', { method: 'POST', body: input })

  const fetchAdminOrders = (query: Record<string, string | number> = {}) =>
    apiFetch<OrderListResult>('/api/admin/orders', { query })

  const searchAdminOrders = (query: Record<string, string | number> = {}) =>
    apiFetch<OrderListResult>('/api/admin/orders/search', { query })

  const updateOrderStatus = (id: string, status: string) =>
    apiFetch<Order>(`/api/admin/orders/${id}/status`, { method: 'PUT', body: { status } })

  const requestOrdersExport = (filters: Record<string, string> = {}) =>
    apiFetch<OrderExportJob>('/api/admin/orders/export', { method: 'POST', body: filters })

  const getOrdersExportStatus = (jobId: string) =>
    apiFetch<OrderExportJob>(`/api/admin/orders/export/${jobId}`)

  const downloadOrdersExport = async (jobId: string, fileName?: string | null) => {
    const base = (config.public.apiUrl as string) || ''
    const url = `${base}/api/admin/orders/export/${jobId}/file`
    const res = await fetch(url, {
      headers: { Authorization: `Bearer ${authStore.accessToken}` },
    })
    if (!res.ok) {
      const text = await res.text()
      throw new Error(text || 'Tải file Excel thất bại')
    }
    const blob = await res.blob()
    const a = document.createElement('a')
    a.href = URL.createObjectURL(blob)
    a.download = fileName || `don-hang-${jobId.slice(0, 8)}.xlsx`
    a.click()
    URL.revokeObjectURL(a.href)
  }

  const exportOrdersExcel = async (filters: Record<string, string> = {}) => {
    const job = await requestOrdersExport(filters)
    const maxAttempts = 600
    for (let i = 0; i < maxAttempts; i++) {
      await new Promise((r) => setTimeout(r, 1000))
      const status = await getOrdersExportStatus(job.job_id)
      if (status.status === 'completed') {
        await downloadOrdersExport(job.job_id, status.file_name)
        return status
      }
      if (status.status === 'failed') {
        throw new Error(status.error || 'Xuất Excel thất bại')
      }
    }
    throw new Error('Xuất Excel quá thời gian chờ (10 phút) — thử lại sau')
  }

  return {
    fetchMyOrders,
    fetchOrder,
    trackOrder,
    fetchAdminOrders,
    searchAdminOrders,
    updateOrderStatus,
    requestOrdersExport,
    getOrdersExportStatus,
    downloadOrdersExport,
    exportOrdersExcel,
  }
}
