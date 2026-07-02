<template>
  <v-avatar :size="size" :color="showImage ? undefined : color" class="user-avatar">
    <v-img v-if="showImage" :src="blobUrl!" cover :alt="name || 'Avatar'" />
    <span v-else class="user-avatar__initials">{{ initials }}</span>
  </v-avatar>
</template>

<script setup lang="ts">
const props = withDefaults(
  defineProps<{
    userId?: string
    name?: string
    hasAvatar?: boolean
    size?: number
    color?: string
    reloadKey?: number
  }>(),
  {
    size: 40,
    color: 'primary',
    hasAvatar: false,
  },
)

const auth = useAuth()
const config = useRuntimeConfig()
const blobUrl = ref<string | null>(null)
const failed = ref(false)

const initials = computed(() => {
  const parts = (props.name || '').trim().split(/\s+/).filter(Boolean)
  if (!parts.length) return '?'
  return parts.map((p) => p[0]).join('').slice(0, 2).toUpperCase()
})

const showImage = computed(() => Boolean(props.hasAvatar && props.userId && blobUrl.value && !failed.value))

async function loadAvatar() {
  if (blobUrl.value) {
    URL.revokeObjectURL(blobUrl.value)
    blobUrl.value = null
  }
  failed.value = false

  if (!props.hasAvatar || !props.userId) return

  const apiBase = (config.public.apiUrl as string) || ''
  try {
    const blob = await $fetch<Blob>(`${apiBase}/api/auth/avatars/${props.userId}`, {
      headers: auth.authHeaders(),
      responseType: 'blob',
    })
    blobUrl.value = URL.createObjectURL(blob)
  } catch {
    failed.value = true
  }
}

watch(
  () => [props.userId, props.hasAvatar, props.reloadKey] as const,
  () => { loadAvatar() },
  { immediate: true },
)

onUnmounted(() => {
  if (blobUrl.value) URL.revokeObjectURL(blobUrl.value)
})

defineExpose({ reload: loadAvatar })
</script>

<style scoped>
.user-avatar__initials {
  font-size: 0.875rem;
  font-weight: 700;
  letter-spacing: 0.02em;
}
</style>
