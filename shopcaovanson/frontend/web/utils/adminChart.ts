import {
  ArcElement,
  BarController,
  BarElement,
  CategoryScale,
  Chart,
  DoughnutController,
  Filler,
  Legend,
  LineController,
  LineElement,
  LinearScale,
  PointElement,
  Tooltip,
  type ChartOptions,
} from 'chart.js'

let registered = false

export function ensureChartJs() {
  if (registered || !import.meta.client) {
    return
  }
  Chart.register(
    CategoryScale,
    LinearScale,
    BarController,
    BarElement,
    LineController,
    LineElement,
    PointElement,
    DoughnutController,
    ArcElement,
    Filler,
    Tooltip,
    Legend,
  )
  Chart.defaults.font.family = "'Segoe UI', system-ui, -apple-system, sans-serif"
  Chart.defaults.color = '#64748b'
  Chart.defaults.plugins.legend.labels.usePointStyle = true
  Chart.defaults.plugins.legend.labels.boxWidth = 8
  registered = true
}

export function readAdminChartColors() {
  if (!import.meta.client) {
    return {
      primary: '#1565c0',
      accent: '#0288d1',
      success: '#059669',
      warning: '#f57c00',
      muted: '#94a3b8',
    }
  }
  const el = document.querySelector('.admin-shell') ?? document.documentElement
  const style = getComputedStyle(el)
  return {
    primary: style.getPropertyValue('--admin-primary').trim() || '#1565c0',
    accent: style.getPropertyValue('--admin-accent').trim() || '#0288d1',
    success: '#059669',
    warning: '#f57c00',
    muted: '#94a3b8',
  }
}

export function hexToRgba(hex: string, alpha: number): string {
  const h = hex.replace('#', '')
  if (h.length !== 6) {
    return `rgba(21, 101, 192, ${alpha})`
  }
  const r = parseInt(h.slice(0, 2), 16)
  const g = parseInt(h.slice(2, 4), 16)
  const b = parseInt(h.slice(4, 6), 16)
  return `rgba(${r}, ${g}, ${b}, ${alpha})`
}

export function baseCartesianOptions(): ChartOptions<'bar' | 'line'> {
  return {
    responsive: true,
    maintainAspectRatio: false,
    interaction: { mode: 'index', intersect: false },
    plugins: {
      legend: {
        position: 'top',
        align: 'end',
        labels: { padding: 16, font: { size: 12, weight: 500 } },
      },
      tooltip: {
        backgroundColor: 'rgba(15, 23, 42, 0.92)',
        titleFont: { size: 13, weight: '600' },
        bodyFont: { size: 12 },
        padding: 12,
        cornerRadius: 10,
        displayColors: true,
      },
    },
    scales: {
      x: {
        grid: { display: false },
        border: { display: false },
        ticks: { maxRotation: 0, autoSkip: true, maxTicksLimit: 12, font: { size: 11 } },
      },
      y: {
        beginAtZero: true,
        grid: { color: 'rgba(148, 163, 184, 0.15)' },
        border: { display: false },
        ticks: { font: { size: 11 } },
      },
    },
  }
}

export const orderStatusColors: Record<string, string> = {
  pending: '#f59e0b',
  confirmed: '#3b82f6',
  shipping: '#8b5cf6',
  delivered: '#10b981',
  cancelled: '#ef4444',
}

export const orderStatusLabels: Record<string, string> = {
  pending: 'Chờ xử lý',
  confirmed: 'Đã xác nhận',
  shipping: 'Đang giao',
  delivered: 'Đã giao',
  cancelled: 'Đã hủy',
}
