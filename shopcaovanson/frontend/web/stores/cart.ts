import { defineStore } from 'pinia'
import type { CartItem } from '~/types'

const STORAGE_KEY = 'shop_cart'

export const useCartStore = defineStore('cart', () => {
  const items = ref<CartItem[]>([])
  const loading = ref(false)
  const hydrated = ref(false)

  const count = computed(() =>
    items.value.reduce((sum, item) => sum + item.quantity, 0),
  )

  const total = computed(() =>
    items.value.reduce((sum, item) => sum + item.unit_price * item.quantity, 0),
  )

  function persist() {
    if (!import.meta.client) {
      return
    }
    localStorage.setItem(STORAGE_KEY, JSON.stringify(items.value))
  }

  function hydrate() {
    if (!import.meta.client || hydrated.value) {
      return
    }
    try {
      const raw = localStorage.getItem(STORAGE_KEY)
      if (raw) {
        items.value = JSON.parse(raw) as CartItem[]
      }
    } catch {
      items.value = []
    }
    hydrated.value = true
  }

  function setItems(value: CartItem[]) {
    items.value = value
    persist()
  }

  function setLoading(value: boolean) {
    loading.value = value
  }

  function upsertItem(item: CartItem) {
    hydrate()
    const idx = items.value.findIndex((i) => i.product_id === item.product_id)
    if (idx >= 0) {
      items.value[idx] = {
        ...items.value[idx],
        quantity: items.value[idx].quantity + item.quantity,
        unit_price: item.unit_price,
        product_name: item.product_name,
        product_image: item.product_image,
      }
    } else {
      items.value.push({ ...item })
    }
    persist()
  }

  function updateQuantity(productId: string, quantity: number) {
    hydrate()
    if (quantity <= 0) {
      removeItem(productId)
      return
    }
    const idx = items.value.findIndex((i) => i.product_id === productId)
    if (idx >= 0) {
      items.value[idx].quantity = quantity
      persist()
    }
  }

  function removeItem(productId: string) {
    hydrate()
    items.value = items.value.filter((i) => i.product_id !== productId)
    persist()
  }

  function clear() {
    items.value = []
    persist()
  }

  return {
    items,
    loading,
    count,
    total,
    hydrate,
    setItems,
    setLoading,
    upsertItem,
    updateQuantity,
    removeItem,
    clear,
  }
})
