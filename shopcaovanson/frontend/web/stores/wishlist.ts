import { defineStore } from 'pinia'
import type { Product } from '~/types'

const STORAGE_KEY = 'shop_wishlist'

export const useWishlistStore = defineStore('wishlist', () => {
  const items = ref<Product[]>([])
  const hydrated = ref(false)

  const count = computed(() => items.value.length)
  const ids = computed(() => new Set(items.value.map((p) => p.id)))

  function persist() {
    if (!import.meta.client) return
    localStorage.setItem(STORAGE_KEY, JSON.stringify(items.value))
  }

  function hydrate() {
    if (!import.meta.client || hydrated.value) return
    try {
      const raw = localStorage.getItem(STORAGE_KEY)
      if (raw) items.value = JSON.parse(raw) as Product[]
    } catch {
      items.value = []
    }
    hydrated.value = true
  }

  function isFavorite(productId: string) {
    return ids.value.has(productId)
  }

  function toggle(product: Product) {
    hydrate()
    const idx = items.value.findIndex((p) => p.id === product.id)
    if (idx >= 0) {
      items.value.splice(idx, 1)
    } else {
      items.value.unshift({ ...product })
    }
    persist()
    return idx < 0
  }

  function remove(productId: string) {
    hydrate()
    items.value = items.value.filter((p) => p.id !== productId)
    persist()
  }

  function clear() {
    items.value = []
    persist()
  }

  return { items, count, hydrate, isFavorite, toggle, remove, clear }
})
