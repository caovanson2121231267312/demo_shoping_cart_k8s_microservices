import type { Cart, CartItem } from '~/types'

export const useCart = () => {
  const cartStore = useCartStore()
  const { apiFetch, isLoggedIn } = useAuth()
  const { formatVND } = useFormat()

  const fetchCart = async () => {
    if (!isLoggedIn.value) {
      cartStore.clear()
      return
    }
    cartStore.setLoading(true)
    try {
      const cart = await apiFetch<Cart>('/api/cart')
      cartStore.setItems(cart.items || [])
    } finally {
      cartStore.setLoading(false)
    }
  }

  const addItem = async (productId: string, quantity = 1) => {
    if (!isLoggedIn.value) {
      await navigateTo('/auth/login')
      return
    }
    cartStore.setLoading(true)
    try {
      const cart = await apiFetch<Cart>('/api/cart/items', {
        method: 'POST',
        body: { product_id: productId, quantity },
      })
      cartStore.setItems(cart.items || [])
    } finally {
      cartStore.setLoading(false)
    }
  }

  const updateItem = async (productId: string, quantity: number) => {
    cartStore.setLoading(true)
    try {
      const cart = await apiFetch<Cart>(`/api/cart/items/${productId}`, {
        method: 'PUT',
        body: { quantity },
      })
      cartStore.setItems(cart.items || [])
    } finally {
      cartStore.setLoading(false)
    }
  }

  const removeItem = async (productId: string) => {
    cartStore.setLoading(true)
    try {
      const cart = await apiFetch<Cart>(`/api/cart/items/${productId}`, {
        method: 'DELETE',
      })
      cartStore.setItems(cart.items || [])
    } finally {
      cartStore.setLoading(false)
    }
  }

  const clearCart = async () => {
    cartStore.setLoading(true)
    try {
      await apiFetch('/api/cart', { method: 'DELETE' })
      cartStore.clear()
    } finally {
      cartStore.setLoading(false)
    }
  }

  return {
    items: computed(() => cartStore.items),
    count: computed(() => cartStore.count),
    total: computed(() => cartStore.total),
    loading: computed(() => cartStore.loading),
    formattedTotal: computed(() => formatVND(cartStore.total)),
    fetchCart,
    addItem,
    updateItem,
    removeItem,
    clearCart,
  }
}
