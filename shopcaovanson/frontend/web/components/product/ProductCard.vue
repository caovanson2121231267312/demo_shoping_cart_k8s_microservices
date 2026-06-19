<template>
  <v-card class="product-card h-100 d-flex flex-column" elevation="0" rounded="lg">
    <div class="product-card__media position-relative">
      <NuxtLink :to="`/products/${product.slug}`" class="product-card__link">
        <v-img
          :src="imageUrl"
          :alt="product.name"
          aspect-ratio="1"
          cover
          class="product-card__image"
        />
      </NuxtLink>

      <div class="product-card__badges">
        <v-chip v-if="onSale" color="error" size="x-small" variant="flat" class="font-weight-bold">
          -{{ discountPercent }}%
        </v-chip>
        <v-chip v-if="product.stock <= 5 && product.stock > 0" color="warning" size="x-small" variant="flat">
          Sắp hết
        </v-chip>
        <v-chip v-if="product.stock <= 0" color="grey" size="x-small" variant="flat">Hết hàng</v-chip>
      </div>

      <v-btn
        class="product-card__wishlist"
        :icon="isFav ? 'mdi-heart' : 'mdi-heart-outline'"
        :color="isFav ? 'error' : 'white'"
        size="small"
        variant="flat"
        @click.stop="onToggleWishlist"
      />

      <div class="product-card__overlay">
        <v-btn
          color="white"
          variant="flat"
          size="small"
          class="text-primary"
          :to="`/products/${product.slug}`"
        >
          Xem chi tiết
        </v-btn>
        <v-btn
          color="secondary"
          variant="flat"
          size="small"
          :loading="adding"
          :disabled="product.stock <= 0"
          @click.stop="handleAddToCart"
        >
          Thêm giỏ
        </v-btn>
      </div>
    </div>

    <v-card-text class="product-card__body flex-grow-1 pa-3">
      <div v-if="product.category" class="text-caption text-primary mb-1 text-truncate">
        {{ product.category.name }}
      </div>
      <NuxtLink :to="`/products/${product.slug}`" class="product-card__title text-body-2 font-weight-medium">
        {{ product.name }}
      </NuxtLink>
      <div v-if="summary && summary.total > 0" class="mt-1">
        <StarRating :value="summary.average" :count="summary.total" size="sm" />
      </div>
      <div class="d-flex align-center flex-wrap ga-2 mt-2">
        <span class="text-subtitle-1 text-primary font-weight-bold">{{ formatVND(effectivePrice) }}</span>
        <span v-if="onSale" class="text-decoration-line-through text-grey text-caption">
          {{ formatVND(product.price) }}
        </span>
      </div>
    </v-card-text>

    <v-card-actions class="pa-3 pt-0">
      <v-btn
        block
        color="primary"
        variant="tonal"
        rounded="lg"
        :loading="adding"
        :disabled="product.stock <= 0"
        @click="handleAddToCart"
      >
        <v-icon start>mdi-cart-plus</v-icon>
        {{ product.stock > 0 ? 'Mua ngay' : 'Hết hàng' }}
      </v-btn>
    </v-card-actions>
  </v-card>
</template>

<script setup lang="ts">
import type { Product } from '~/types'
import type { ReviewSummary } from '~/composables/useReviewSummary'

const props = defineProps<{
  product: Product
  summary?: ReviewSummary | null
}>()

const { formatVND } = useFormat()
const { getEffectivePrice, isOnSale } = useProducts()
const cart = useCart()
const wishlist = useWishlist()

const adding = ref(false)
const imageUrl = computed(() => getProductImage(props.product))
const effectivePrice = computed(() => getEffectivePrice(props.product))
const onSale = computed(() => isOnSale(props.product))
const isFav = computed(() => wishlist.isFavorite(props.product.id))

const discountPercent = computed(() => {
  if (!onSale.value || !props.product.sale_price) return 0
  return Math.round((1 - props.product.sale_price / props.product.price) * 100)
})

const handleAddToCart = async () => {
  adding.value = true
  try {
    await cart.addItem(props.product)
    useSnackbar().show('Đã thêm vào giỏ hàng', 'success')
  } finally {
    adding.value = false
  }
}

const onToggleWishlist = () => wishlist.toggle(props.product)
</script>

<style scoped>
.product-card {
  border: 1px solid var(--color-border);
  transition: transform 0.25s ease, box-shadow 0.25s ease;
  overflow: hidden;
  background: var(--color-surface);
}

.product-card:hover {
  transform: translateY(-4px);
  box-shadow: var(--shadow-lg);
}

.product-card__image {
  transition: transform 0.4s ease;
}

.product-card:hover .product-card__image {
  transform: scale(1.05);
}

.product-card__media {
  overflow: hidden;
}

.product-card__link {
  display: block;
}

.product-card__badges {
  position: absolute;
  top: 8px;
  left: 8px;
  display: flex;
  flex-direction: column;
  gap: 4px;
  z-index: 2;
}

.product-card__wishlist {
  position: absolute;
  top: 8px;
  right: 8px;
  z-index: 3;
  opacity: 0;
  transform: scale(0.8);
  transition: opacity 0.2s, transform 0.2s;
}

.product-card:hover .product-card__wishlist,
.product-card__wishlist:focus-visible {
  opacity: 1;
  transform: scale(1);
}

.product-card__overlay {
  position: absolute;
  inset: 0;
  background: rgba(0, 0, 0, 0.35);
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 8px;
  opacity: 0;
  transition: opacity 0.25s ease;
  z-index: 1;
}

.product-card:hover .product-card__overlay {
  opacity: 1;
}

.product-card__title {
  color: inherit;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  line-height: 1.35;
  min-height: 2.7em;
}

.product-card__title:hover {
  color: rgb(var(--v-theme-primary));
}
</style>
