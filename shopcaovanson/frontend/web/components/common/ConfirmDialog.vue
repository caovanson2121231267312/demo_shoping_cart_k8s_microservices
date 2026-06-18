<template>
  <v-dialog :model-value="modelValue" max-width="420" @update:model-value="$emit('update:modelValue', $event)">
    <v-card>
      <v-card-title>{{ title }}</v-card-title>
      <v-card-text>{{ message }}</v-card-text>
      <v-card-actions>
        <v-spacer />
        <v-btn variant="text" @click="$emit('update:modelValue', false)">{{ cancelText }}</v-btn>
        <v-btn color="primary" variant="flat" :loading="loading" @click="$emit('confirm')">
          {{ confirmText }}
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script setup lang="ts">
withDefaults(
  defineProps<{
    modelValue: boolean
    title?: string
    message?: string
    confirmText?: string
    cancelText?: string
    loading?: boolean
  }>(),
  {
    title: 'Xác nhận',
    message: 'Bạn có chắc chắn muốn thực hiện hành động này?',
    confirmText: 'Xác nhận',
    cancelText: 'Hủy',
    loading: false,
  },
)

defineEmits<{
  'update:modelValue': [value: boolean]
  confirm: []
}>()
</script>
