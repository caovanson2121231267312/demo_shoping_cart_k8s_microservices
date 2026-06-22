import type { Product } from '~/types'
import { storeToRefs } from 'pinia'

export const useWishlist = () => {
  const store = useWishlistStore()
  const { items, count } = storeToRefs(store)
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
    items,
    count,
    isFavorite: (id: string) => store.isFavorite(id),
    toggle,
    remove: store.remove,
    clear: store.clear,
  }
}
