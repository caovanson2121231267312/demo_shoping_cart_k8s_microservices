<template>
  <div class="chat-composer">
    <input
      ref="fileInput"
      type="file"
      accept="image/jpeg,image/png,image/gif,image/webp"
      class="d-none"
      @change="onFileChange"
    >

    <div v-if="imagePreview" class="chat-composer__preview mb-2">
      <img :src="imagePreview" alt="Xem trước" class="chat-composer__preview-img">
      <v-btn icon size="x-small" variant="flat" color="error" class="chat-composer__preview-remove" @click="clearPreview">
        <v-icon size="16">mdi-close</v-icon>
      </v-btn>
    </div>

    <div class="d-flex align-end ga-1">
      <v-btn
        icon
        size="small"
        variant="text"
        color="primary"
        :disabled="disabled || uploading"
        title="Gửi ảnh"
        @click="fileInput?.click()"
      >
        <v-icon>mdi-image-outline</v-icon>
      </v-btn>

      <v-menu location="top" :close-on-content-click="false">
        <template #activator="{ props: menuProps }">
          <v-btn
            v-bind="menuProps"
            icon
            size="small"
            variant="text"
            color="primary"
            :disabled="disabled"
            title="Emoji"
          >
            <v-icon>mdi-emoticon-outline</v-icon>
          </v-btn>
        </template>
        <v-card class="pa-2 chat-composer__emoji-panel" min-width="220">
          <div class="chat-composer__emoji-grid">
            <button
              v-for="emoji in CHAT_INPUT_EMOJIS"
              :key="emoji"
              type="button"
              class="chat-composer__emoji-btn"
              @click="appendEmoji(emoji)"
            >
              {{ emoji }}
            </button>
          </div>
        </v-card>
      </v-menu>

      <v-text-field
        :model-value="modelValue"
        :placeholder="placeholder"
        density="compact"
        variant="outlined"
        hide-details
        class="flex-grow-1"
        :disabled="disabled || uploading"
        @update:model-value="emit('update:modelValue', $event)"
        @keyup.enter="submit"
        @input="emit('typing')"
      />

      <v-btn
        icon
        size="small"
        color="primary"
        variant="flat"
        :disabled="!canSend"
        :loading="uploading"
        @click="submit"
      >
        <v-icon>mdi-send</v-icon>
      </v-btn>
    </div>

    <div v-if="hint" class="text-caption text-medium-emphasis mt-2">
      {{ hint }}
    </div>
  </div>
</template>

<script setup lang="ts">
import { CHAT_INPUT_EMOJIS } from '~/utils/chat'

const props = defineProps<{
  modelValue: string
  roomId?: string | null
  disabled?: boolean
  uploading?: boolean
  placeholder?: string
  hint?: string
}>()

const emit = defineEmits<{
  'update:modelValue': [value: string]
  send: []
  typing: []
  'send-image': [file: File]
}>()

const fileInput = ref<HTMLInputElement | null>(null)
const pendingFile = ref<File | null>(null)
const imagePreview = ref<string | null>(null)

const canSend = computed(() => {
  if (props.disabled || props.uploading) {
    return false
  }
  return !!props.modelValue.trim() || !!pendingFile.value
})

function appendEmoji(emoji: string) {
  emit('update:modelValue', props.modelValue + emoji)
}

function clearPreview() {
  pendingFile.value = null
  imagePreview.value = null
  if (fileInput.value) {
    fileInput.value.value = ''
  }
}

function onFileChange(event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) {
    return
  }
  if (file.size > 5 * 1024 * 1024) {
    alert('Ảnh tối đa 5MB')
    input.value = ''
    return
  }
  pendingFile.value = file
  imagePreview.value = URL.createObjectURL(file)
}

function submit() {
  if (pendingFile.value && props.roomId) {
    emit('send-image', pendingFile.value)
    clearPreview()
    return
  }
  if (props.modelValue.trim()) {
    emit('send')
  }
}

onUnmounted(() => {
  if (imagePreview.value) {
    URL.revokeObjectURL(imagePreview.value)
  }
})
</script>

<style scoped>
.chat-composer__preview {
  position: relative;
  display: inline-block;
}

.chat-composer__preview-img {
  max-width: 120px;
  max-height: 80px;
  border-radius: 8px;
  object-fit: cover;
  border: 1px solid var(--color-border, rgba(0, 0, 0, 0.1));
}

.chat-composer__preview-remove {
  position: absolute;
  top: -6px;
  right: -6px;
}

.chat-composer__emoji-grid {
  display: grid;
  grid-template-columns: repeat(6, 1fr);
  gap: 4px;
}

.chat-composer__emoji-btn {
  border: none;
  background: transparent;
  font-size: 1.25rem;
  padding: 4px;
  border-radius: 6px;
  cursor: pointer;
}

.chat-composer__emoji-btn:hover {
  background: rgba(var(--v-theme-primary), 0.08);
}
</style>
