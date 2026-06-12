<script setup lang="ts">
import { ref, onMounted, watch, nextTick } from 'vue'
import * as echarts from 'echarts'
import { statisticsApi } from '@/api'
import { ConditionLabels } from '@/types'
import type {
  OverviewStats,
  MonthlyData,
  PlatformStat,
  PublisherStat,
  ConditionStat
} from '@/types'

const loading = ref(true)
const selectedYear = ref(new Date().getFullYear())
const years = ref<number[]>([])

const overview = ref<OverviewStats | null>(null)
const monthlyData = ref<MonthlyData[]>([])
const platformData = ref<PlatformStat[]>([])
const publisherData = ref<PublisherStat[]>([])
const conditionData = ref<ConditionStat[]>([])

const annualChartRef = ref<HTMLDivElement | null>(null)
const platformChartRef = ref<HTMLDivElement | null>(null)
const publisherChartRef = ref<HTMLDivElement | null>(null)
const conditionChartRef = ref<HTMLDivElement | null>(null)

let annualChart: echarts.ECharts | null = null
let platformChart: echarts.ECharts | null = null
let publisherChart: echarts.ECharts | null = null
let conditionChart: echarts.ECharts | null = null

const neonColors = ['#00F0FF', '#FFD93D', '#FF6B9D', '#6BCB77', '#FF8C42', '#FF4757', '#A78BFA', '#22D3EE']

const initYears = () => {
  const currentYear = new Date().getFullYear()
  years.value = Array.from({ length: 5 }, (_, i) => currentYear - i)
}

const initCharts = () => {
  if (annualChartRef.value) annualChart = echarts.init(annualChartRef.value)
  if (platformChartRef.value) platformChart = echarts.init(platformChartRef.value)
  if (publisherChartRef.value) publisherChart = echarts.init(publisherChartRef.value)
  if (conditionChartRef.value) conditionChart = echarts.init(conditionChartRef.value)
}

const renderAnnualChart = () => {
  if (!annualChart) return
  const months = ['1月', '2月', '3月', '4月', '5月', '6月', '7月', '8月', '9月', '10月', '11月', '12月']
  const counts = monthlyData.value.map(d => d.count)

  annualChart.setOption({
    backgroundColor: 'transparent',
    grid: { left: '3%', right: '4%', bottom: '3%', top: '10%', containLabel: true },
    xAxis: {
      type: 'category',
      data: months,
      axisLine: { lineStyle: { color: '#00F0FF' } },
      axisLabel: { color: '#E8E8E8', fontFamily: 'Press Start 2P', fontSize: 10 },
      axisTick: { show: false }
    },
    yAxis: {
      type: 'value',
      axisLine: { lineStyle: { color: '#00F0FF' } },
      axisLabel: { color: '#E8E8E8', fontFamily: 'Press Start 2P', fontSize: 10 },
      splitLine: { lineStyle: { color: 'rgba(0, 240, 255, 0.1)' } }
    },
    series: [{
      type: 'bar',
      data: counts,
      itemStyle: {
        color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
          { offset: 0, color: '#00F0FF' },
          { offset: 1, color: '#00B8C4' }
        ]),
        borderRadius: [4, 4, 0, 0],
        borderColor: '#FFD93D',
        borderWidth: 2
      },
      emphasis: {
        itemStyle: {
          shadowBlur: 20,
          shadowColor: '#00F0FF'
        }
      },
      barWidth: '50%',
      animationDuration: 1500,
      animationEasing: 'elasticOut'
    }]
  })
}

const renderPlatformChart = () => {
  if (!platformChart) return
  const data = platformData.value.map(p => ({ name: p.platform, value: p.count }))

  platformChart.setOption({
    backgroundColor: 'transparent',
    tooltip: { trigger: 'item', fontFamily: 'VT323', fontSize: 16 },
    legend: {
      orient: 'vertical',
      right: '5%',
      top: 'center',
      textStyle: { color: '#E8E8E8', fontFamily: 'Press Start 2P', fontSize: 10 },
      itemWidth: 12,
      itemHeight: 12
    },
    series: [{
      type: 'pie',
      radius: ['30%', '60%'],
      center: ['35%', '50%'],
      avoidLabelOverlap: false,
      itemStyle: {
        borderRadius: 4,
        borderColor: '#1A1A2E',
        borderWidth: 3
      },
      label: { show: false },
      emphasis: {
        label: { show: true, fontSize: 14, fontWeight: 'bold', color: '#FFD93D', fontFamily: 'Press Start 2P' },
        itemStyle: { shadowBlur: 20, shadowOffsetX: 0, shadowColor: 'rgba(0, 0, 0, 0.5)' }
      },
      data: data,
      color: neonColors,
      animationDuration: 1500,
      animationEasing: 'bounceOut'
    }]
  })
}

const renderPublisherChart = () => {
  if (!publisherChart) return
  const top10 = [...publisherData.value].sort((a, b) => b.count - a.count).slice(0, 10).reverse()
  const names = top10.map(p => p.publisher)
  const counts = top10.map(p => p.count)

  publisherChart.setOption({
    backgroundColor: 'transparent',
    grid: { left: '3%', right: '10%', bottom: '3%', top: '5%', containLabel: true },
    xAxis: {
      type: 'value',
      axisLine: { lineStyle: { color: '#00F0FF' } },
      axisLabel: { color: '#E8E8E8', fontFamily: 'Press Start 2P', fontSize: 10 },
      splitLine: { lineStyle: { color: 'rgba(0, 240, 255, 0.1)' } }
    },
    yAxis: {
      type: 'category',
      data: names,
      axisLine: { lineStyle: { color: '#00F0FF' } },
      axisLabel: { color: '#E8E8E8', fontFamily: 'Press Start 2P', fontSize: 9 },
      axisTick: { show: false }
    },
    series: [{
      type: 'bar',
      data: counts,
      itemStyle: {
        color: new echarts.graphic.LinearGradient(0, 0, 1, 0, [
          { offset: 0, color: '#FF6B9D' },
          { offset: 1, color: '#FF8C42' }
        ]),
        borderRadius: [0, 4, 4, 0],
        borderColor: '#FFD93D',
        borderWidth: 2
      },
      barWidth: '60%',
      animationDuration: 1500,
      animationEasing: 'quadOut'
    }]
  })
}

const renderConditionChart = () => {
  if (!conditionChart) return
  const data = conditionData.value.map(c => ({
    name: ConditionLabels[c.condition] || c.condition,
    value: c.count
  }))

  conditionChart.setOption({
    backgroundColor: 'transparent',
    tooltip: { trigger: 'item', fontFamily: 'VT323', fontSize: 16 },
    legend: {
      bottom: '0',
      textStyle: { color: '#E8E8E8', fontFamily: 'Press Start 2P', fontSize: 10 },
      itemWidth: 12,
      itemHeight: 12
    },
    series: [{
      type: 'pie',
      radius: ['40%', '70%'],
      center: ['50%', '40%'],
      avoidLabelOverlap: false,
      itemStyle: {
        borderRadius: 6,
        borderColor: '#1A1A2E',
        borderWidth: 3
      },
      label: { show: false },
      emphasis: {
        label: { show: true, fontSize: 14, fontWeight: 'bold', color: '#FFD93D', fontFamily: 'Press Start 2P' },
        itemStyle: { shadowBlur: 20, shadowOffsetX: 0, shadowColor: 'rgba(0, 0, 0, 0.5)' }
      },
      data: data,
      color: ['#6BCB77', '#00F0FF', '#FFD93D', '#FF8C42', '#FF4757'],
      animationDuration: 1500,
      animationEasing: 'bounceOut'
    }]
  })
}

const renderAllCharts = () => {
  nextTick(() => {
    renderAnnualChart()
    renderPlatformChart()
    renderPublisherChart()
    renderConditionChart()
  })
}

const loadData = async () => {
  loading.value = true
  try {
    const [ovRes, annualRes, platRes, pubRes, condRes] = await Promise.all([
      statisticsApi.getOverview(),
      statisticsApi.getAnnual(selectedYear.value),
      statisticsApi.getPlatforms(),
      statisticsApi.getPublishers(),
      statisticsApi.getConditions()
    ])
    if (ovRes.code === 0) overview.value = ovRes.data
    if (annualRes.code === 0) monthlyData.value = annualRes.data
    if (platRes.code === 0) platformData.value = platRes.data
    if (pubRes.code === 0) publisherData.value = pubRes.data
    if (condRes.code === 0) conditionData.value = condRes.data
    renderAllCharts()
  } finally {
    loading.value = false
  }
}

const handleResize = () => {
  annualChart?.resize()
  platformChart?.resize()
  publisherChart?.resize()
  conditionChart?.resize()
}

watch(selectedYear, loadData)

onMounted(() => {
  initYears()
  initCharts()
  loadData()
  window.addEventListener('resize', handleResize)
})
</script>

<template>
  <div class="p-6 space-y-6">
    <div class="text-center py-4">
      <h1 class="text-neon-blue glow-blue mb-2">年度游玩报告</h1>
      <p class="text-text-secondary pixel-font text-xs">ANNUAL PLAY REPORT</p>
    </div>

    <div class="flex justify-center">
      <select v-model="selectedYear" class="pixel-input text-center pixel-font">
        <option v-for="y in years" :key="y" :value="y">{{ y }} 年</option>
      </select>
    </div>

    <div v-if="loading" class="text-center py-16">
      <div class="text-4xl mb-4 animate-pulse">📊</div>
      <p class="text-text-secondary pixel-font">加载中...</p>
    </div>

    <template v-else>
      <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
        <div class="stat-card">
          <div class="text-2xl mb-2">🎮</div>
          <div class="stat-number text-neon-blue">{{ overview?.totalCartridges ?? 0 }}</div>
          <div class="text-text-secondary mt-2 text-sm">总卡带</div>
        </div>
        <div class="stat-card">
          <div class="text-2xl mb-2">🏆</div>
          <div class="stat-number text-pixel-green">{{ overview?.totalPlaythroughs ?? 0 }}</div>
          <div class="text-text-secondary mt-2 text-sm">已通关</div>
        </div>
        <div class="stat-card">
          <div class="text-2xl mb-2">⏱️</div>
          <div class="stat-number text-bright-yellow">{{ overview?.totalPlayTime ?? 0 }}h</div>
          <div class="text-text-secondary mt-2 text-sm">总时长</div>
        </div>
        <div class="stat-card">
          <div class="text-2xl mb-2">💰</div>
          <div class="stat-number text-pixel-pink">¥{{ overview?.totalValue ?? 0 }}</div>
          <div class="text-text-secondary mt-2 text-sm">总价值</div>
        </div>
      </div>

      <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <div class="pixel-card p-4">
          <h3 class="text-neon-blue glow-blue mb-4 text-center">年度通关统计</h3>
          <div ref="annualChartRef" class="w-full h-64"></div>
        </div>

        <div class="pixel-card p-4">
          <h3 class="text-pixel-pink mb-4 text-center">平台占比</h3>
          <div ref="platformChartRef" class="w-full h-64"></div>
        </div>

        <div class="pixel-card p-4">
          <h3 class="text-bright-yellow glow-yellow mb-4 text-center">发行商 Top 10</h3>
          <div ref="publisherChartRef" class="w-full h-64"></div>
        </div>

        <div class="pixel-card p-4">
          <h3 class="text-pixel-green mb-4 text-center">品相分布</h3>
          <div ref="conditionChartRef" class="w-full h-64"></div>
        </div>
      </div>
    </template>
  </div>
</template>
