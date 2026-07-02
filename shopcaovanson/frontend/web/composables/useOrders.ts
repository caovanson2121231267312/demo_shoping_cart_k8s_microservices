import type { FetchOptions } from 'ofetch'
import type { Order, OrderListResult, TrackOrderInput } from '~/types'
import { downloadBlob } from '~/utils/downloadFile'

export interface OrderExportJob {
  job_id: string
  status: 'pending' | 'processing' | 'completed' | 'failed'
  progress?: number
  progress_message?: string | null
  created_at?: string
  completed_at?: string | null
  error?: string | null
  file_name?: string | null
  filters?: Record<string, string>
}

export interface OrderExportProgress {
  progress: number
  message: string
  status: OrderExportJob['status']
}

export const useOrders = () => {
  const { apiFetch } = useAuth()

  const fetchMyOrders = (page = 1, limit = 20) =>
    apiFetch<OrderListResult>('/api/orders', { query: { page, limit } })

  const fetchOrder = (id: string) => apiFetch<Order>(`/api/orders/${id}`)

  const trackOrder = (input: TrackOrderInput) =>
    $fetch<Order>('/api/orders/track', { method: 'POST', body: input })

  const fetchAdminOrders = (query: Record<string, string | number> = {}) =>
    apiFetch<OrderListResult>('/api/admin/orders', { query })

  const searchAdminOrders = (query: Record<string, string | number> = {}) =>
    apiFetch<OrderListResult>('/api/admin/orders/search', { query })

  const fetchAdminOrder = (id: string) => apiFetch<Order>(`/api/admin/orders/${id}`)

  const updateOrderStatus = (id: string, status: string) =>
    apiFetch<Order>(`/api/admin/orders/${id}/status`, { method: 'PUT', body: { status } })

  const requestOrdersExport = (filters: Record<string, string> = {}) =>
    apiFetch<OrderExportJob>('/api/admin/orders/export', { method: 'POST', body: filters })

  const getOrdersExportStatus = (jobId: string) =>
    apiFetch<OrderExportJob>(`/api/admin/orders/export/${jobId}`)

  const downloadOrdersExport = async (jobId: string, fileName?: string | null) => {
    const blob = await apiFetch<Blob>(`/api/admin/orders/export/${jobId}/file`, {
      responseType: 'blob',
    } as FetchOptions<'json'>)

    if (!blob?.size) {
      throw new Error('File Excel trống hoặc chưa sẵn sàng — thử lại sau vài giây')
    }

    const safeName = (fileName || `don-hang-${jobId.slice(0, 8)}.xlsx`).replace(/[\\/:*?"<>|]/g, '-')
    downloadBlob(blob, safeName)
  }

  const exportOrdersExcel = async (
    filters: Record<string, string> = {},
    onProgress?: (progress: OrderExportProgress) => void,
    onBeforeDownload?: () => void | Promise<void>,
  ) => {
    const emit = (status: OrderExportJob) => {
      onProgress?.({
        progress: status.progress ?? (status.status === 'pending' ? 5 : 50),
        message: status.progress_message || 'Đang xử lý...',
        status: status.status,
      })
    }

    const job = await requestOrdersExport(filters)
    emit({ ...job, status: job.status, progress: 0, progress_message: 'Đã tạo yêu cầu xuất...' })

    const maxAttempts = 600
    for (let i = 0; i < maxAttempts; i++) {
      await new Promise((r) => setTimeout(r, 1000))
      const status = await getOrdersExportStatus(job.job_id)
      emit(status)
      if (status.status === 'completed') {
        onProgress?.({ progress: 99, message: 'Đang tải file về...', status: 'completed' })
        await onBeforeDownload?.()
        await downloadOrdersExport(job.job_id, status.file_name)
        onProgress?.({ progress: 100, message: 'Hoàn tất', status: 'completed' })
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
    fetchAdminOrder,
    updateOrderStatus,
    requestOrdersExport,
    getOrdersExportStatus,
    downloadOrdersExport,
    exportOrdersExcel,
  }
}
