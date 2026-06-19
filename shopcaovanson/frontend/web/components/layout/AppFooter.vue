<template>
  <footer class="site-footer">
    <!-- Trust highlights -->
    <section class="site-footer__features">
      <v-container class="site-footer__features-inner">
        <div v-for="item in features" :key="item.title" class="site-footer__feature">
          <div class="site-footer__feature-icon" :style="{ background: item.bg }">
            <v-icon :color="item.color" size="26">{{ item.icon }}</v-icon>
          </div>
          <div>
            <div class="site-footer__feature-title">{{ item.title }}</div>
            <div class="site-footer__feature-desc">{{ item.desc }}</div>
          </div>
        </div>
      </v-container>
    </section>

    <!-- Main footer -->
    <section class="site-footer__main">
      <v-container class="py-10 py-md-12">
        <v-row>
          <v-col cols="12" lg="4" class="mb-6 mb-lg-0">
            <NuxtLink to="/" class="site-footer__brand">
              <div class="site-footer__brand-logo">
                <v-icon color="white" size="28">mdi-storefront-outline</v-icon>
              </div>
              <div>
                <div class="site-footer__brand-name">Shop Cao Văn Sơn</div>
                <div class="site-footer__brand-slogan">Mua sắm thông minh — Giao hàng tận tay</div>
              </div>
            </NuxtLink>
            <p class="site-footer__about">
              Nền tảng thương mại điện tử uy tín, cung cấp hàng triệu sản phẩm chính hãng với giá tốt nhất.
              Cam kết giao hàng nhanh, đổi trả dễ dàng và hỗ trợ khách hàng 24/7.
            </p>
            <div class="site-footer__social">
              <a v-for="s in socials" :key="s.label" :href="s.href" :aria-label="s.label" class="site-footer__social-btn">
                <v-icon size="20">{{ s.icon }}</v-icon>
              </a>
            </div>
            <div class="site-footer__app d-none d-md-flex">
              <div class="site-footer__qr">
                <v-icon size="48" color="grey-lighten-1">mdi-qrcode</v-icon>
              </div>
              <div>
                <div class="text-body-2 font-weight-bold mb-1">Tải ứng dụng</div>
                <div class="text-caption text-grey-lighten-1 mb-2">Quét mã QR để mua sắm mọi lúc</div>
                <div class="d-flex ga-2">
                  <v-chip size="small" variant="outlined" color="grey-lighten-1">
                    <v-icon start size="14">mdi-apple</v-icon> App Store
                  </v-chip>
                  <v-chip size="small" variant="outlined" color="grey-lighten-1">
                    <v-icon start size="14">mdi-google-play</v-icon> Google Play
                  </v-chip>
                </div>
              </div>
            </div>
          </v-col>

          <v-col cols="6" sm="4" lg="2">
            <h3 class="site-footer__heading">Về chúng tôi</h3>
            <ul class="site-footer__links">
              <li><NuxtLink to="/">Trang chủ</NuxtLink></li>
              <li><NuxtLink to="/products">Sản phẩm</NuxtLink></li>
              <li><NuxtLink to="/products?sort=newest">Hàng mới về</NuxtLink></li>
              <li><NuxtLink to="/blog">Tin tức & Bài viết</NuxtLink></li>
              <li><NuxtLink to="/products?sort=price_desc">Khuyến mãi</NuxtLink></li>
            </ul>
          </v-col>

          <v-col cols="6" sm="4" lg="2">
            <h3 class="site-footer__heading">Hỗ trợ khách hàng</h3>
            <ul class="site-footer__links">
              <li><NuxtLink to="/orders/track">Tra cứu đơn hàng</NuxtLink></li>
              <li><NuxtLink to="/orders">Đơn hàng của tôi</NuxtLink></li>
              <li><NuxtLink to="/cart">Giỏ hàng</NuxtLink></li>
              <li><NuxtLink to="/checkout">Hướng dẫn thanh toán</NuxtLink></li>
            </ul>
          </v-col>

          <v-col cols="6" sm="4" lg="2">
            <h3 class="site-footer__heading">Chính sách</h3>
            <ul class="site-footer__links">
              <li><a href="#">Chính sách bảo mật</a></li>
              <li><a href="#">Điều khoản sử dụng</a></li>
              <li><a href="#">Chính sách đổi trả</a></li>
              <li><a href="#">Chính sách vận chuyển</a></li>
            </ul>
          </v-col>

          <v-col cols="12" sm="8" lg="2">
            <h3 class="site-footer__heading">Liên hệ</h3>
            <ul class="site-footer__contact">
              <li>
                <v-icon size="18" color="primary">mdi-map-marker</v-icon>
                <span>TP. Hồ Chí Minh, Việt Nam</span>
              </li>
              <li>
                <v-icon size="18" color="primary">mdi-phone</v-icon>
                <a href="tel:19001234">Hotline: <strong>1900 1234</strong></a>
              </li>
              <li>
                <v-icon size="18" color="primary">mdi-email</v-icon>
                <a href="mailto:support@shopcaovanson.xyz">support@shopcaovanson.xyz</a>
              </li>
              <li>
                <v-icon size="18" color="primary">mdi-clock-outline</v-icon>
                <span>8:00 – 22:00 hàng ngày</span>
              </li>
            </ul>
          </v-col>
        </v-row>
      </v-container>
    </section>

    <!-- Payment & shipping -->
    <section class="site-footer__partners">
      <v-container class="site-footer__partners-inner">
        <div class="site-footer__partner-group">
          <span class="site-footer__partner-label">Thanh toán</span>
          <div class="site-footer__badges">
            <span v-for="m in paymentMethods" :key="m" class="site-footer__badge">{{ m }}</span>
          </div>
        </div>
        <div class="site-footer__partner-divider d-none d-md-block" />
        <div class="site-footer__partner-group">
          <span class="site-footer__partner-label">Vận chuyển</span>
          <div class="site-footer__badges">
            <span v-for="b in shippingPartners" :key="b" class="site-footer__badge site-footer__badge--ship">{{ b }}</span>
          </div>
        </div>
        <div class="site-footer__partner-divider d-none d-md-block" />
        <div class="site-footer__cert d-none d-lg-flex">
          <v-icon color="success" size="20" class="mr-2">mdi-shield-check</v-icon>
          <span>Đã đăng ký Bộ Công Thương</span>
        </div>
      </v-container>
    </section>

    <!-- Copyright -->
    <section class="site-footer__bottom">
      <v-container class="site-footer__bottom-inner">
        <span>© {{ year }} Shop Cao Văn Sơn. Bảo lưu mọi quyền.</span>
        <div class="site-footer__bottom-links">
          <span>Thanh toán COD</span>
          <span class="site-footer__dot">•</span>
          <span>Đổi trả 7 ngày</span>
          <span class="site-footer__dot">•</span>
          <span>Bảo mật SSL</span>
        </div>
      </v-container>
    </section>
  </footer>
</template>

<script setup lang="ts">
const year = new Date().getFullYear()

const features = [
  { icon: 'mdi-truck-fast', title: 'Giao hàng toàn quốc', desc: 'Freeship đơn từ 500K', color: 'primary', bg: 'var(--color-primary-alpha-12)' },
  { icon: 'mdi-shield-check', title: 'Hàng chính hãng', desc: 'Cam kết 100% authentic', color: 'success', bg: 'var(--color-success-bg)' },
  { icon: 'mdi-sync', title: 'Đổi trả dễ dàng', desc: 'Miễn phí trong 7 ngày', color: 'info', bg: 'var(--color-info-bg)' },
  { icon: 'mdi-headset', title: 'Hỗ trợ 24/7', desc: 'Chat & hotline luôn sẵn sàng', color: 'secondary', bg: 'var(--color-secondary-50)' },
]

const socials = [
  { icon: 'mdi-facebook', label: 'Facebook', href: '#' },
  { icon: 'mdi-instagram', label: 'Instagram', href: '#' },
  { icon: 'mdi-youtube', label: 'YouTube', href: '#' },
  { icon: 'mdi-twitter', label: 'Twitter', href: '#' },
]

const paymentMethods = ['COD', 'Chuyển khoản', 'Ví MoMo', 'VNPay', 'Thẻ tín dụng']
const shippingPartners = ['GHN', 'GHTK', 'Viettel Post', 'J&T Express']
</script>

<style scoped>
.site-footer {
  margin-top: auto;
}

.site-footer__features {
  background: var(--color-surface);
  border-top: 1px solid var(--color-border);
}

.site-footer__features-inner {
  max-width: 1280px;
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 16px;
  padding-top: 24px;
  padding-bottom: 24px;
}

@media (min-width: 768px) {
  .site-footer__features-inner {
    grid-template-columns: repeat(4, 1fr);
    gap: 24px;
  }
}

.site-footer__feature {
  display: flex;
  align-items: center;
  gap: 12px;
}

.site-footer__feature-icon {
  width: 48px;
  height: 48px;
  border-radius: var(--radius-md);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.site-footer__feature-title {
  font-size: 13px;
  font-weight: 700;
  color: var(--color-text);
  line-height: 1.3;
}

.site-footer__feature-desc {
  font-size: 11px;
  color: var(--color-text-secondary);
  margin-top: 2px;
}

.site-footer__main {
  background: var(--gradient-footer);
  color: var(--color-text-footer);
}

.site-footer__brand {
  display: flex;
  align-items: center;
  gap: 12px;
  text-decoration: none;
  color: inherit;
  margin-bottom: 16px;
}

.site-footer__brand-logo {
  width: 48px;
  height: 48px;
  border-radius: var(--radius-md);
  background: var(--gradient-primary);
  display: flex;
  align-items: center;
  justify-content: center;
}

.site-footer__brand-name {
  font-size: 20px;
  font-weight: 800;
  color: var(--color-text-inverse);
  line-height: 1.2;
}

.site-footer__brand-slogan {
  font-size: 12px;
  color: var(--color-text-footer-muted);
  margin-top: 2px;
}

.site-footer__about {
  font-size: 14px;
  line-height: 1.65;
  color: var(--color-text-footer);
  margin-bottom: 20px;
  max-width: 360px;
}

.site-footer__social {
  display: flex;
  gap: 8px;
  margin-bottom: 24px;
}

.site-footer__social-btn {
  width: 40px;
  height: 40px;
  border-radius: var(--radius-sm);
  background: var(--color-border-white);
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--color-text-footer);
  text-decoration: none;
  transition: background 0.2s, color 0.2s, transform 0.2s;
}

.site-footer__social-btn:hover {
  background: var(--color-primary);
  color: var(--color-text-inverse);
  transform: translateY(-2px);
}

.site-footer__app {
  align-items: flex-start;
  gap: 16px;
  padding: 16px;
  background: rgba(255, 255, 255, 0.04);
  border-radius: var(--radius-md);
  border: 1px solid var(--color-border-white);
}

.site-footer__qr {
  width: 80px;
  height: 80px;
  background: var(--color-surface);
  border-radius: var(--radius-sm);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.site-footer__heading {
  font-size: 15px;
  font-weight: 700;
  color: var(--color-text-inverse);
  margin-bottom: 16px;
  letter-spacing: 0.2px;
}

.site-footer__links {
  list-style: none;
  padding: 0;
  margin: 0;
}

.site-footer__links li {
  margin-bottom: 10px;
}

.site-footer__links a {
  color: var(--color-text-footer);
  text-decoration: none;
  font-size: 14px;
  transition: color 0.2s, padding-left 0.2s;
  display: inline-block;
}

.site-footer__links a:hover {
  color: var(--color-text-inverse);
  padding-left: 4px;
}

.site-footer__contact {
  list-style: none;
  padding: 0;
  margin: 0;
}

.site-footer__contact li {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  margin-bottom: 12px;
  font-size: 14px;
  color: var(--color-text-footer);
}

.site-footer__contact a {
  color: var(--color-text-footer);
  text-decoration: none;
}

.site-footer__contact a:hover {
  color: var(--color-text-inverse);
}

.site-footer__contact strong {
  color: var(--color-text-footer-accent);
}

.site-footer__partners {
  background: var(--color-footer-partners);
  border-top: 1px solid var(--color-border);
}

.site-footer__partners-inner {
  max-width: 1280px;
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 16px;
  padding-top: 20px;
  padding-bottom: 20px;
}

.site-footer__partner-group {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 10px;
}

.site-footer__partner-label {
  font-size: 11px;
  font-weight: 700;
  color: var(--color-text-footer-muted);
  text-transform: uppercase;
  letter-spacing: 0.5px;
  white-space: nowrap;
}

.site-footer__badges {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.site-footer__badge {
  font-size: 11px;
  font-weight: 600;
  padding: 4px 10px;
  border-radius: var(--radius-sm);
  background: var(--color-border-white);
  color: #cfd8dc;
  border: 1px solid rgba(255, 255, 255, 0.1);
}

.site-footer__badge--ship {
  background: var(--color-primary-alpha-12);
  border-color: rgba(21, 101, 192, 0.3);
  color: var(--color-primary-200);
}

.site-footer__partner-divider {
  width: 1px;
  height: 28px;
  background: rgba(255, 255, 255, 0.1);
}

.site-footer__cert {
  align-items: center;
  font-size: 12px;
  color: var(--color-text-footer-muted);
  margin-left: auto;
}

.site-footer__bottom {
  background: var(--color-footer-bottom);
  border-top: 1px solid rgba(255, 255, 255, 0.05);
}

.site-footer__bottom-inner {
  max-width: 1280px;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  padding-top: 16px;
  padding-bottom: 16px;
  font-size: 13px;
  color: var(--color-text-footer-muted);
  text-align: center;
}

@media (min-width: 600px) {
  .site-footer__bottom-inner {
    flex-direction: row;
    justify-content: space-between;
    text-align: left;
  }
}

.site-footer__bottom-links {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px;
}

.site-footer__dot {
  opacity: 0.4;
}
</style>
