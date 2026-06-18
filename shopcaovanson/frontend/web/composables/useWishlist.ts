import type { Product } from '~/types'

export const useWishlist = () => {
  const store = useWishlistStore()
  const snackbar = useSnackbar()

  onMounted(() => store.hydrate())

  const toggle = (product: Product) => {
    const added = store.toggle(product)
    snackbar.show(
      added ? `Đã thêm "${product.name}" vào yêu thích` : 'Đã xóa khỏi danh sách yêu thích',
      added ? 'success' : 'info',
    )
    return added
  }

  return {
    items: computed(() => store.items),
    count: computed(() => store.count),
    isFavorite: (id: string) => store.isFavorite(id),
    toggle,
    remove: store.remove,
    clear: store.clear,
  }
}
