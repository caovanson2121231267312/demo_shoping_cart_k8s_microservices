<template>
  <v-footer color="grey-darken-4" class="site-footer text-white pa-0">
    <div class="site-footer__top w-100">
      <v-container class="py-10">
        <v-row>
          <v-col cols="12" md="4">
            <div class="d-flex align-center mb-4">
              <v-icon color="primary" size="32" class="mr-2">mdi-storefront</v-icon>
              <span class="text-h6 font-weight-bold">Shop Cao Văn Sơn</span>
            </div>
            <p class="text-body-2 text-grey-lighten-1 mb-4">
              Sàn thương mại điện tử chuyên cung cấp sản phẩm chất lượng, giao hàng nhanh toàn quốc,
              hỗ trợ khách hàng 24/7.
            </p>
            <div class="d-flex ga-2">
              <v-btn icon variant="text" color="grey-lighten-1" size="small"><v-icon>mdi-facebook</v-icon></v-btn>
              <v-btn icon variant="text" color="grey-lighten-1" size="small"><v-icon>mdi-instagram</v-icon></v-btn>
              <v-btn icon variant="text" color="grey-lighten-1" size="small"><v-icon>mdi-youtube</v-icon></v-btn>
            </div>
          </v-col>
          <v-col cols="6" md="2">
            <div class="text-subtitle-1 font-weight-bold mb-3">Mua sắm</div>
            <div class="d-flex flex-column ga-2 footer-link">
              <NuxtLink to="/products">Tất cả sản phẩm</NuxtLink>
              <NuxtLink to="/wishlist">Yêu thích</NuxtLink>
              <NuxtLink to="/cart">Giỏ hàng</NuxtLink>
              <NuxtLink to="/checkout">Thanh toán</NuxtLink>
            </div>
          </v-col>
          <v-col cols="6" md="2">
            <div class="text-subtitle-1 font-weight-bold mb-3">Hỗ trợ</div>
            <div class="d-flex flex-column ga-2 footer-link">
              <NuxtLink to="/orders/track">Tra cứu đơn hàng</NuxtLink>
              <NuxtLink to="/orders">Đơn hàng của tôi</NuxtLink>
              <NuxtLink to="/auth/login">Đăng nhập</NuxtLink>
              <NuxtLink to="/auth/register">Đăng ký</NuxtLink>
            </div>
          </v-col>
          <v-col cols="12" md="4">
            <div class="text-subtitle-1 font-weight-bold mb-3">Liên hệ</div>
            <div class="text-body-2 text-grey-lighten-1 d-flex flex-column ga-2">
              <span><v-icon size="16" class="mr-1">mdi-map-marker</v-icon> TP. Hồ Chí Minh, Việt Nam</span>
              <span><v-icon size="16" class="mr-1">mdi-phone</v-icon> Hotline: 1900 1234</span>
              <span><v-icon size="16" class="mr-1">mdi-email</v-icon> support@shopcaovanson.xyz</span>
              <span><v-icon size="16" class="mr-1">mdi-clock-outline</v-icon> 8:00 – 22:00 hàng ngày</span>
            </div>
          </v-col>
        </v-row>
      </v-container>
    </div>
    <v-divider class="border-opacity-25" />
    <v-container class="py-4">
      <div class="d-flex flex-column flex-sm-row justify-space-between align-center text-body-2 text-grey-lighten-1">
        <span>© {{ year }} Shop Cao Văn Sơn. All rights reserved.</span>
        <div class="d-flex ga-4 mt-2 mt-sm-0">
          <span>Thanh toán COD</span>
          <span>Đổi trả 7 ngày</span>
          <span>Bảo mật SSL</span>
        </div>
      </div>
    </v-container>
  </v-footer>
</template>

<script setup lang="ts">
const year = new Date().getFullYear()
</script>

<style scoped>
.site-footer__top {
  background: linear-gradient(180deg, #212121 0%, #1a1a1a 100%);
}

.footer-link a {
  color: #bdbdbd;
  text-decoration: none;
  font-size: 14px;
}

.footer-link a:hover {
  color: #fff;
}
</style>
