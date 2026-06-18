export const formatVND = (amount: number) =>
  new Intl.NumberFormat('vi-VN', { style: 'currency', currency: 'VND' }).format(amount)

export const useFormat = () => ({
  formatVND,
})

export const productImageUrl = (slug: string, width = 400, height = 400) =>
  `https://picsum.photos/seed/${encodeURIComponent(slug)}/${width}/${height}`

export const getProductImage = (product: { slug: string; images?: string[] }) => {
  if (product.images && product.images.length > 0) {
    return product.images[0]
  }
  return productImageUrl(product.slug)
}
