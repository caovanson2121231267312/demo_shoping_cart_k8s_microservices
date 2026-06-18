import type { Order, OrderListResult, TrackOrderInput } from '~/types'

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

  const updateOrderStatus = (id: string, status: string) =>
    apiFetch<Order>(`/api/admin/orders/${id}/status`, { method: 'PUT', body: { status } })

  return {
    fetchMyOrders,
    fetchOrder,
    trackOrder,
    fetchAdminOrders,
    searchAdminOrders,
    updateOrderStatus,
  }
}
