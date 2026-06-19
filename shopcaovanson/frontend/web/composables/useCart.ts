import type { Cart, CartItem, Product } from '~/types'
import { getProductImage } from '~/composables/useFormat'

export const useCart = () => {
  const cartStore = useCartStore()
  const { apiFetch, isLoggedIn } = useAuth()
  const { formatVND } = useFormat()
  const { getEffectivePrice } = useProducts()

  const toCartItem = (product: Product, quantity: number): CartItem => ({
    product_id: product.id,
    quantity,
    unit_price: getEffectivePrice(product),
    product_name: product.name,
    product_image: getProductImage(product),
  })

  const syncItemToServer = async (item: CartItem) => {
    await apiFetch<Cart>('/api/cart/items', {
      method: 'POST',
      body: { product_id: item.product_id, quantity: item.quantity },
    })
  }

  const fetchCart = async () => {
    cartStore.hydrate()
    if (!isLoggedIn.value) {
      return
    }

    cartStore.setLoading(true)
    try {
      const cart = await apiFetch<Cart>('/api/cart')
      cartStore.setItems(cart.items || [])
    } catch {
      cartStore.hydrate()
    } finally {
      cartStore.setLoading(false)
    }
  }

  const mergeLocalToServer = async () => {
    cartStore.hydrate()
    if (!isLoggedIn.value) {
      return
    }

    const localItems = [...cartStore.items]
    if (!localItems.length) {
      await fetchCart()
      return
    }

    cartStore.setLoading(true)
    try {
      for (const item of localItems) {
        await syncItemToServer(item)
      }
      await fetchCart()
    } catch {
      cartStore.hydrate()
    } finally {
      cartStore.setLoading(false)
    }
  }

  const addItem = async (product: Product, quantity = 1) => {
    const item = toCartItem(product, quantity)
    cartStore.upsertItem(item)

    if (!isLoggedIn.value) {
      return
    }

    cartStore.setLoading(true)
    try {
      const cart = await apiFetch<Cart>('/api/cart/items', {
        method: 'POST',
        body: { product_id: product.id, quantity },
      })
      cartStore.setItems(cart.items || [])
    } catch {
      cartStore.hydrate()
    } finally {
      cartStore.setLoading(false)
    }
  }

  const updateItem = async (productId: string, quantity: number) => {
    cartStore.updateQuantity(productId, quantity)

    if (!isLoggedIn.value) {
      return
    }

    cartStore.setLoading(true)
    try {
      const cart = await apiFetch<Cart>(`/api/cart/items/${productId}`, {
        method: 'PUT',
        body: { quantity },
      })
      cartStore.setItems(cart.items || [])
    } catch {
      cartStore.hydrate()
    } finally {
      cartStore.setLoading(false)
    }
  }

  const removeItem = async (productId: string) => {
    cartStore.removeItem(productId)

    if (!isLoggedIn.value) {
      return
    }

    cartStore.setLoading(true)
    try {
      const cart = await apiFetch<Cart>(`/api/cart/items/${productId}`, {
        method: 'DELETE',
      })
      cartStore.setItems(cart.items || [])
    } catch {
      cartStore.hydrate()
    } finally {
      cartStore.setLoading(false)
    }
  }

  const clearCart = async () => {
    cartStore.clear()

    if (!isLoggedIn.value) {
      return
    }

    cartStore.setLoading(true)
    try {
      await apiFetch('/api/cart', { method: 'DELETE' })
    } catch {
      // keep local cleared state
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
    mergeLocalToServer,
    addItem,
    updateItem,
    removeItem,
    clearCart,
  }
}
