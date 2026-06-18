<template>
  <v-card variant="outlined" class="mb-3">
    <v-card-text class="d-flex align-center ga-3">
      <v-img
        :src="imageUrl"
        width="72"
        height="72"
        cover
        class="rounded flex-shrink-0"
      />
      <div class="flex-grow-1 min-width-0">
        <div class="text-body-1 font-weight-medium text-truncate">{{ item.product_name }}</div>
        <div class="text-primary font-weight-bold">{{ formatVND(item.unit_price) }}</div>
        <div class="d-flex align-center ga-2 mt-2">
          <v-btn
            icon
            size="x-small"
            variant="outlined"
            :disabled="item.quantity <= 1 || updating"
            @click="changeQty(item.quantity - 1)"
          >
            <v-icon>mdi-minus</v-icon>
          </v-btn>
          <span class="text-body-2">{{ item.quantity }}</span>
          <v-btn
            icon
            size="x-small"
            variant="outlined"
            :disabled="updating"
            @click="changeQty(item.quantity + 1)"
          >
            <v-icon>mdi-plus</v-icon>
          </v-btn>
        </div>
      </div>
      <div class="text-right">
        <div class="font-weight-bold mb-2">{{ formatVND(lineTotal) }}</div>
        <v-btn
          icon
          size="small"
          variant="text"
          color="error"
          :loading="removing"
          @click="handleRemove"
        >
          <v-icon>mdi-delete-outline</v-icon>
        </v-btn>
      </div>
    </v-card-text>
  </v-card>
</template>

<script setup lang="ts">
import type { CartItem } from '~/types'

const props = defineProps<{ item: CartItem }>()

const { formatVND } = useFormat()
const cart = useCart()

const updating = ref(false)
const removing = ref(false)

const lineTotal = computed(() => props.item.unit_price * props.item.quantity)
const imageUrl = computed(() => {
  if (props.item.product_image) {
    return props.item.product_image
  }
  return productImageUrl(props.item.product_id)
})

const changeQty = async (qty: number) => {
  if (qty < 1) {
    return
  }
  updating.value = true
  try {
    await cart.updateItem(props.item.product_id, qty)
  } finally {
    updating.value = false
  }
}

const handleRemove = async () => {
  removing.value = true
  try {
    await cart.removeItem(props.item.product_id)
  } finally {
    removing.value = false
  }
}
</script>
