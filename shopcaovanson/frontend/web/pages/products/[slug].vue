<template>
  <v-container class="page-container py-6">
    <LoadingSpinner v-if="loading" />
    <EmptyState
      v-else-if="!product"
      icon="mdi-alert-circle-outline"
      title="Không tìm thấy sản phẩm"
      description="Sản phẩm có thể đã bị xóa hoặc không tồn tại."
    >
      <v-btn color="primary" to="/products" class="mt-4">Quay lại danh sách</v-btn>
    </EmptyState>

    <template v-else>
      <v-breadcrumbs :items="breadcrumbs" class="px-0" />

      <v-row>
        <v-col cols="12" md="6">
          <v-img
            :src="mainImage"
            :alt="product.name"
            aspect-ratio="1"
            cover
            class="rounded-lg"
          />
          <v-row v-if="product.images.length > 1" class="mt-2">
            <v-col v-for="(img, i) in product.images" :key="i" cols="3">
              <v-img
                :src="img"
                aspect-ratio="1"
                cover
                class="rounded cursor-pointer"
                @click="selectedImage = img"
              />
            </v-col>
          </v-row>
        </v-col>

        <v-col cols="12" md="6">
          <div class="d-flex align-start justify-space-between">
            <h1 class="text-h4 font-weight-bold mb-2 flex-grow-1">{{ product.name }}</h1>
            <v-btn
              :icon="isFavorite ? 'mdi-heart' : 'mdi-heart-outline'"
              :color="isFavorite ? 'error' : 'grey'"
              variant="text"
              size="large"
              @click="toggleWishlist"
            />
          </div>

          <div v-if="reviewSummary && reviewSummary.total > 0" class="mb-3">
            <StarRating :value="reviewSummary.average" :count="reviewSummary.total" show-value />
          </div>

          <div v-if="product.category" class="text-body-2 text-grey mb-4">
            Danh mục: {{ product.category.name }}
          </div>

          <div class="d-flex align-center ga-3 mb-4">
            <span class="text-h4 text-primary font-weight-bold">{{ formatVND(effectivePrice) }}</span>
            <span v-if="onSale" class="text-decoration-line-through text-grey text-h6">
              {{ formatVND(product.price) }}
            </span>
            <v-chip v-if="onSale" color="error" size="small">Giảm giá</v-chip>
          </div>

          <p class="text-body-1 mb-4">{{ product.description }}</p>
          <div class="text-body-2 mb-4">Còn lại: <strong>{{ product.stock }}</strong> sản phẩm</div>

          <div class="d-flex align-center ga-3 mb-6">
            <v-text-field
              v-model.number="quantity"
              type="number"
              min="1"
              :max="product.stock"
              label="Số lượng"
              density="compact"
              variant="outlined"
              hide-details
              style="max-width: 120px"
            />
            <v-btn
              color="primary"
              size="large"
              :loading="adding"
              :disabled="product.stock <= 0"
              @click="handleAddToCart"
            >
              <v-icon start>mdi-cart-plus</v-icon>
              Thêm vào giỏ
            </v-btn>
            <v-btn
              variant="outlined"
              size="large"
              :color="isFavorite ? 'error' : 'primary'"
              @click="toggleWishlist"
            >
              <v-icon start>{{ isFavorite ? 'mdi-heart' : 'mdi-heart-outline' }}</v-icon>
              Yêu thích
            </v-btn>
          </div>

          <v-card v-if="product.details?.specifications" variant="outlined" class="mb-4">
            <v-card-title class="text-subtitle-1">Thông số kỹ thuật</v-card-title>
            <v-card-text>
              <v-table density="compact">
                <tbody>
                  <tr v-for="(val, key) in product.details.specifications" :key="key">
                    <td class="font-weight-medium">{{ key }}</td>
                    <td>{{ val }}</td>
                  </tr>
                </tbody>
              </v-table>
            </v-card-text>
          </v-card>

          <div
            v-if="product.details?.rich_description"
            class="text-body-2"
            v-html="product.details.rich_description"
          />
        </v-col>
      </v-row>

      <v-divider class="my-8" />

      <h2 class="text-h5 font-weight-bold mb-4">Đánh giá từ khách hàng</h2>

      <ReviewSummary v-if="reviewSummary && reviewSummary.total > 0" :summary="reviewSummary" />

      <v-card v-if="auth.isLoggedIn.value" variant="outlined" class="mb-6 pa-4">
        <v-form @submit.prevent="submitReview">
          <v-rating v-model="reviewRating" color="secondary" class="mb-2" />
          <v-textarea
            v-model="reviewComment"
            label="Nhận xét của bạn"
            rows="3"
            variant="outlined"
            hide-details
          />
          <v-btn type="submit" color="primary" class="mt-3" :loading="submittingReview">
            Gửi đánh giá
          </v-btn>
        </v-form>
      </v-card>

      <LoadingSpinner v-if="loadingReviews" />
      <EmptyState
        v-else-if="!reviews.length"
        icon="mdi-star-outline"
        title="Chưa có đánh giá"
        description="Hãy là người đầu tiên đánh giá sản phẩm này."
      />
      <div v-else>
        <v-card
          v-for="review in reviews"
          :key="review.id"
          variant="outlined"
          class="mb-3 pa-4 review-card"
        >
          <div class="d-flex align-center mb-2">
            <v-avatar color="primary" size="36" class="mr-3">
              <span class="text-white text-caption">KH</span>
            </v-avatar>
            <div>
              <StarRating :value="review.rating" size="sm" />
              <div class="text-caption text-grey">
                {{ new Date(review.created_at).toLocaleDateString('vi-VN') }}
              </div>
            </div>
          </div>
          <p v-if="review.comment" class="text-body-2 mb-0">{{ review.comment }}</p>
        </v-card>
      </div>
    </template>
  </v-container>
</template>

<script setup lang="ts">
import type { Product, ProductReview } from '~/types'
import type { ReviewSummary } from '~/composables/useReviewSummary'

definePageMeta({ layout: 'default' })

const route = useRoute()
const slug = route.params.slug as string

const auth = useAuth()
const cart = useCart()
const wishlist = useWishlist()
const { formatVND } = useFormat()
const { fetchProductBySlug, fetchReviews, createReview, getEffectivePrice, isOnSale } = useProducts()
const { buildSummary } = useReviewSummary()

const product = ref<Product | null>(null)
const reviews = ref<ProductReview[]>([])
const reviewSummary = ref<ReviewSummary | null>(null)
const loading = ref(true)
const loadingReviews = ref(true)
const adding = ref(false)
const submittingReview = ref(false)
const quantity = ref(1)
const selectedImage = ref('')
const reviewRating = ref(5)
const reviewComment = ref('')

const effectivePrice = computed(() => (product.value ? getEffectivePrice(product.value) : 0))
const onSale = computed(() => (product.value ? isOnSale(product.value) : false))
const isFavorite = computed(() => (product.value ? wishlist.isFavorite(product.value.id) : false))
const mainImage = computed(() => selectedImage.value || (product.value ? getProductImage(product.value) : ''))

const breadcrumbs = computed(() => [
  { title: 'Trang chủ', to: '/' },
  { title: 'Sản phẩm', to: '/products' },
  { title: product.value?.name || '', disabled: true },
])

const loadProduct = async () => {
  loading.value = true
  try {
    product.value = await fetchProductBySlug(slug)
    selectedImage.value = getProductImage(product.value)
  } catch {
    product.value = null
  } finally {
    loading.value = false
  }
}

const loadReviews = async () => {
  if (!product.value) {
    return
  }
  loadingReviews.value = true
  try {
    const result = await fetchReviews(product.value.id, 1, 50)
    reviews.value = result.items
    reviewSummary.value = buildSummary(result.items, result.total)
  } finally {
    loadingReviews.value = false
  }
}

const toggleWishlist = () => {
  if (product.value) wishlist.toggle(product.value)
}

const handleAddToCart = async () => {
  if (!product.value) {
    return
  }
  adding.value = true
  try {
    await cart.addItem(product.value.id, quantity.value)
    useSnackbar().show('Đã thêm vào giỏ hàng', 'success')
  } finally {
    adding.value = false
  }
}

const submitReview = async () => {
  if (!product.value) {
    return
  }
  submittingReview.value = true
  try {
    await createReview(product.value.id, {
      rating: reviewRating.value,
      comment: reviewComment.value || undefined,
    })
    reviewComment.value = ''
    await loadReviews()
  } finally {
    submittingReview.value = false
  }
}

onMounted(async () => {
  await loadProduct()
  await loadReviews()
})
</script>
