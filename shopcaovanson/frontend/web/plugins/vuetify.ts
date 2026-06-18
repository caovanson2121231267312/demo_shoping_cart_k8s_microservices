import { createVuetify } from 'vuetify'
import * as components from 'vuetify/components'
import * as directives from 'vuetify/directives'

export default defineNuxtPlugin((nuxtApp) => {
  const vuetify = createVuetify({
    components,
    directives,
    theme: {
      defaultTheme: 'light',
      themes: {
        light: {
          colors: {
            primary: '#1565C0',
            secondary: '#F57C00',
            accent: '#00897B',
            error: '#D32F2F',
            success: '#388E3C',
          },
        },
      },
    },
    defaults: {
      VBtn: {
        rounded: 'lg',
      },
      VCard: {
        rounded: 'lg',
      },
    },
  })

  nuxtApp.vueApp.use(vuetify)
})
