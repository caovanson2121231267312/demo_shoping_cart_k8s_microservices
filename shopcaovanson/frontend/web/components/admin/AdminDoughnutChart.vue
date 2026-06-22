<template>
  <div class="admin-chart-wrap">
    <div v-if="!hasData" class="admin-chart-wrap__empty">
      <v-icon size="40" color="grey-lighten-1">mdi-chart-donut</v-icon>
      <span>Chưa có dữ liệu</span>
    </div>
    <Doughnut v-else :data="chartData" :options="chartOptions" />
  </div>
</template>

<script setup lang="ts">
import { Doughnut } from 'vue-chartjs'
import type { ChartData, ChartOptions } from 'chart.js'
import { ensureChartJs, orderStatusColors, orderStatusLabels } from '~/utils/adminChart'

const props = defineProps<{
  items: Record<string, number>
  labelMap?: Record<string, string>
  colorMap?: Record<string, string>
}>()

ensureChartJs()

const entries = computed(() =>
  Object.entries(props.items).filter(([, v]) => v > 0),
)

const hasData = computed(() => entries.value.length > 0)

const chartData = computed<ChartData<'doughnut'>>(() => {
  const labels = entries.value.map(([k]) => (props.labelMap ?? orderStatusLabels)[k] ?? k)
  const colors = entries.value.map(([k]) => (props.colorMap ?? orderStatusColors)[k] ?? '#94a3b8')
  return {
    labels,
    datasets: [
      {
        data: entries.value.map(([, v]) => v),
        backgroundColor: colors,
        hoverBackgroundColor: colors.map((c) => c),
        borderWidth: 2,
        borderColor: '#fff',
        hoverOffset: 6,
      },
    ],
  }
})

const chartOptions = computed<ChartOptions<'doughnut'>>(() => ({
  responsive: true,
  maintainAspectRatio: false,
  cutout: '68%',
  plugins: {
    legend: {
      position: 'bottom',
      labels: { padding: 14, font: { size: 12 } },
    },
    tooltip: {
      backgroundColor: 'rgba(15, 23, 42, 0.92)',
      padding: 12,
      cornerRadius: 10,
      callbacks: {
        label: (ctx) => {
          const total = (ctx.dataset.data as number[]).reduce((a, b) => a + b, 0)
          const pct = total ? ((ctx.parsed / total) * 100).toFixed(1) : '0'
          return ` ${ctx.label}: ${ctx.parsed} (${pct}%)`
        },
      },
    },
  },
}))
</script>

<style scoped>
.admin-chart-wrap {
  position: relative;
  height: 280px;
  width: 100%;
}

.admin-chart-wrap__empty {
  height: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 8px;
  color: var(--admin-text-muted, #94a3b8);
  font-size: 13px;
  background: rgba(148, 163, 184, 0.06);
  border-radius: var(--admin-radius-sm);
}
</style>
