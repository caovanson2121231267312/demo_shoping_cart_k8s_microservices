<template>
  <div class="admin-chart-wrap">
    <div v-if="!items.length" class="admin-chart-wrap__empty">
      <v-icon size="40" color="grey-lighten-1">mdi-chart-bar</v-icon>
      <span>Chưa có dữ liệu</span>
    </div>
    <Bar v-else :data="chartData" :options="chartOptions" />
  </div>
</template>

<script setup lang="ts">
import { Bar } from 'vue-chartjs'
import type { ChartData, ChartOptions } from 'chart.js'
import { baseCartesianOptions, ensureChartJs, hexToRgba, readAdminChartColors } from '~/utils/adminChart'

export interface HorizontalBarItem {
  label: string
  value: number
  subLabel?: string
}

const props = defineProps<{
  items: HorizontalBarItem[]
  color?: string
  formatValue?: (v: number) => string
}>()

ensureChartJs()

const chartColor = computed(() => props.color || readAdminChartColors().primary)

const chartData = computed<ChartData<'bar'>>(() => ({
  labels: props.items.map((i) => i.label),
  datasets: [
    {
      label: 'Doanh thu',
      data: props.items.map((i) => i.value),
      backgroundColor: props.items.map((_, idx) =>
        hexToRgba(chartColor.value, 0.92 - idx * 0.06),
      ),
      hoverBackgroundColor: chartColor.value,
      borderRadius: 6,
      borderSkipped: false,
      barThickness: 22,
    },
  ],
}))

const chartOptions = computed<ChartOptions<'bar'>>(() => {
  const base = baseCartesianOptions()
  const fmt = props.formatValue ?? ((v: number) => String(v))
  return {
    ...base,
    indexAxis: 'y',
    plugins: {
      ...base.plugins,
      legend: { display: false },
      tooltip: {
        ...base.plugins?.tooltip,
        callbacks: {
          label: (ctx) => fmt(Number(ctx.parsed.x)),
        },
      },
    },
    scales: {
      x: {
        beginAtZero: true,
        grid: { color: 'rgba(148, 163, 184, 0.15)' },
        border: { display: false },
        ticks: {
          font: { size: 11 },
          callback: (v) => fmt(Number(v)),
        },
      },
      y: {
        grid: { display: false },
        border: { display: false },
        ticks: { font: { size: 11, weight: 500 } },
      },
    },
  }
})
</script>

<style scoped>
.admin-chart-wrap {
  position: relative;
  height: 320px;
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
