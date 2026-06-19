import { createVuetify } from 'vuetify'
import * as components from 'vuetify/components'
import * as directives from 'vuetify/directives'
import { theme } from '~/utils/theme'

export default defineNuxtPlugin((nuxtApp) => {
  const vuetify = createVuetify({
    components,
    directives,
    theme: {
      defaultTheme: 'light',
      themes: {
        light: {
          colors: {
            primary: theme.primary,
            secondary: theme.secondary,
            accent: theme.info,
            error: theme.sale,
            success: theme.success,
            warning: theme.warning,
            info: theme.info,
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
