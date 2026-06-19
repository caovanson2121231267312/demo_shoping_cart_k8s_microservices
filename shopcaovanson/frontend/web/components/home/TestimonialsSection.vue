<template>
  <section class="testimonials py-10">
    <v-container class="page-container">
      <div class="text-center mb-8">
        <h2 class="text-h5 font-weight-bold">Khách hàng nói gì về chúng tôi</h2>
        <p class="text-body-2 text-grey mt-1">Hàng nghìn đánh giá 5 sao từ cộng đồng mua sắm</p>
      </div>
      <v-row>
        <v-col v-for="item in items" :key="item.name" cols="12" md="4">
          <v-card class="testimonial-card h-100 pa-5" elevation="0" rounded="xl">
            <div class="d-flex align-center mb-3">
              <v-avatar :color="item.color" size="44" class="mr-3">
                <span class="text-white font-weight-bold">{{ item.initials }}</span>
              </v-avatar>
              <div>
                <div class="font-weight-bold">{{ item.name }}</div>
                <StarRating :value="5" size="sm" />
              </div>
            </div>
            <p class="text-body-2 text-grey-darken-1 mb-0">"{{ item.comment }}"</p>
          </v-card>
        </v-col>
      </v-row>
    </v-container>
  </section>
</template>

<script setup lang="ts">
const items = [
  { name: 'Nguyễn Văn A', initials: 'NA', color: 'primary', comment: 'Giao hàng nhanh, đóng gói cẩn thận. Sẽ ủng hộ shop lâu dài!' },
  { name: 'Trần Thị B', initials: 'TB', color: 'secondary', comment: 'Giá tốt, nhiều mã giảm giá. Flash sale săn được deal ngon.' },
  { name: 'Lê Minh C', initials: 'LC', color: 'success', comment: 'Hỗ trợ chat nhiệt tình, đổi size dễ dàng. Rất hài lòng.' },
]
</script>

<style scoped>
.testimonials {
  background: var(--color-surface-muted);
}

.testimonial-card {
  border: 1px solid var(--color-border);
  background: var(--color-surface);
  transition: box-shadow 0.25s ease, transform 0.25s ease;
}

.testimonial-card:hover {
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.08);
  transform: translateY(-2px);
}
</style>
