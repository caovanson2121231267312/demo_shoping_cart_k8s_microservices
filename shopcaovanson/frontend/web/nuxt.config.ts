import vuetify from 'vite-plugin-vuetify'

export default defineNuxtConfig({
  compatibilityDate: '2024-08-01',
  devtools: { enabled: true },
  ssr: false,

  modules: ['@pinia/nuxt'],

  components: [
    {
      path: '~/components',
      pathPrefix: false,
    },
  ],

  css: ['~/assets/css/main.css', 'vuetify/styles', '@mdi/font/css/materialdesignicons.css'],

  build: {
    transpile: ['vuetify'],
  },

  vite: {
    plugins: [
      vuetify({
        autoImport: true,
      }),
    ],
  },

  runtimeConfig: {
    apiProxyTarget: process.env.NUXT_API_PROXY_TARGET || 'https://shopapicaovanson.xyz',
    public: {
      apiUrl: process.env.NUXT_PUBLIC_API_URL ?? 'https://shopapicaovanson.xyz',
      wsUrl: process.env.NUXT_PUBLIC_WS_URL || 'wss://shopapicaovanson.xyz',
    },
  },

  app: {
    head: {
      title: 'Shop Cao Van Son',
      meta: [
        { charset: 'utf-8' },
        { name: 'viewport', content: 'width=device-width, initial-scale=1' },
        { name: 'description', content: 'Cửa hàng trực tuyến Shop Cao Van Son' },
      ],
      link: [
        { rel: 'icon', type: 'image/x-icon', href: '/favicon.ico' },
      ],
    },
  },
})
