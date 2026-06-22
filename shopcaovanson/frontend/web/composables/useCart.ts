import type { Cart, CartItem, Product } from '~/types'
import { storeToRefs } from 'pinia'
import { getProductImage } from '~/composables/useFormat'

const CART_MERGE_PENDING_KEY = 'shop_cart_merge_pending'

export function parseCartError(err: unknown): string {
  const data = (err as { data?: { error?: string } })?.data
  const msg = data?.error || (err as Error)?.message || ''
  if (msg.includes('insufficient stock')) {
    return 'Sản phẩm không đủ số lượng trong kho'
  }
  if (msg.includes('product is not active')) {
    return 'Sản phẩm không còn bán'
  }
  return msg || 'Không thể cập nhật giỏ hàng'
}

export const useCart = () => {
  const cartStore = useCartStore()
  const { items, count, total, loading } = storeToRefs(cartStore)
  const { apiFetch, isLoggedIn } = useAuth()
  const { formatVND } = useFormat()
  const { getEffectivePrice } = useProducts()

  const formattedTotal = computed(() => formatVND(total.value))

  const markGuestCartPending = () => {
    if (import.meta.client) {
      sessionStorage.setItem(CART_MERGE_PENDING_KEY, '1')
    }
  }

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

    const shouldMerge =
      import.meta.client &&
      sessionStorage.getItem(CART_MERGE_PENDING_KEY) === '1' &&
      cartStore.items.length > 0

    cartStore.setLoading(true)
    try {
      if (shouldMerge) {
        const localItems = [...cartStore.items]
        for (const item of localItems) {
          await syncItemToServer(item)
        }
        sessionStorage.removeItem(CART_MERGE_PENDING_KEY)
      }
      const cart = await apiFetch<Cart>('/api/cart')
      cartStore.setItems(cart.items || [])
    } catch {
      try {
        const cart = await apiFetch<Cart>('/api/cart')
        cartStore.setItems(cart.items || [])
      } catch {
        cartStore.hydrate()
      }
    } finally {
      cartStore.setLoading(false)
    }
  }

  const addItem = async (product: Product, quantity = 1) => {
    const item = toCartItem(product, quantity)
    cartStore.upsertItem(item)

    if (!isLoggedIn.value) {
      markGuestCartPending()
      return
    }

    cartStore.setLoading(true)
    try {
      const cart = await apiFetch<Cart>('/api/cart/items', {
        method: 'POST',
        body: { product_id: product.id, quantity },
      })
      cartStore.setItems(cart.items || [])
    } catch (err) {
      await fetchCart()
      throw err
    } finally {
      cartStore.setLoading(false)
    }
  }

  const updateItem = async (productId: string, quantity: number) => {
    if (!isLoggedIn.value) {
      cartStore.updateQuantity(productId, quantity)
      markGuestCartPending()
      return
    }

    cartStore.setLoading(true)
    try {
      const cart = await apiFetch<Cart>(`/api/cart/items/${productId}`, {
        method: 'PUT',
        body: { quantity },
      })
      cartStore.setItems(cart.items || [])
    } catch (err) {
      await fetchCart()
      throw err
    } finally {
      cartStore.setLoading(false)
    }
  }

  const removeItem = async (productId: string) => {
    if (!isLoggedIn.value) {
      cartStore.removeItem(productId)
      markGuestCartPending()
      return
    }

    cartStore.setLoading(true)
    try {
      const cart = await apiFetch<Cart>(`/api/cart/items/${productId}`, {
        method: 'DELETE',
      })
      cartStore.setItems(cart.items || [])
    } catch (err) {
      await fetchCart()
      throw err
    } finally {
      cartStore.setLoading(false)
    }
  }

  const clearCart = async () => {
    if (!isLoggedIn.value) {
      cartStore.clear()
      if (import.meta.client) {
        sessionStorage.removeItem(CART_MERGE_PENDING_KEY)
      }
      return
    }

    cartStore.setLoading(true)
    try {
      await apiFetch('/api/cart', { method: 'DELETE' })
      cartStore.clear()
    } catch {
      // keep local cleared state
    } finally {
      cartStore.setLoading(false)
    }
  }

  return {
    items,
    count,
    total,
    loading,
    formattedTotal,
    fetchCart,
    mergeLocalToServer,
    addItem,
    updateItem,
    removeItem,
    clearCart,
  }
}
