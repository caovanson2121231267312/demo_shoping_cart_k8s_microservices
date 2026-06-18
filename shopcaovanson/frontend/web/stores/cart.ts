import { defineStore } from 'pinia'
import type { CartItem } from '~/types'

export const useCartStore = defineStore('cart', () => {
  const items = ref<CartItem[]>([])
  const loading = ref(false)

  const count = computed(() =>
    items.value.reduce((sum, item) => sum + item.quantity, 0),
  )

  const total = computed(() =>
    items.value.reduce((sum, item) => sum + item.unit_price * item.quantity, 0),
  )

  function setItems(value: CartItem[]) {
    items.value = value
  }

  function setLoading(value: boolean) {
    loading.value = value
  }

  function clear() {
    items.value = []
  }

  return {
    items,
    loading,
    count,
    total,
    setItems,
    setLoading,
    clear,
  }
})
