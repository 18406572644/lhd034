<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useCartridgeStore } from '@/stores/cartridge'
import { cartridgeApi, playthroughApi } from '@/api'
import type { Cartridge, Playthrough, OverviewStats } from '@/types'

const router = useRouter()
const store = useCartridgeStore()

const recentCartridges = ref<Cartridge[]>([])
const recentPlaythroughs = ref<Playthrough[]>([])
const loading = ref(true)

const stats = computed<OverviewStats | null>(() => store.stats)

const statCards = computed(() => [
  { label: '卡带总数', value: stats.value?.totalCartridges ?? 0, color: 'text-neon-blue', icon: '🎮' },
  { label: '已通关数', value: stats.value?.totalPlaythroughs ?? 0, color: 'text-pixel-green', icon: '🏆' },
  { label: '待玩数', value: stats.value?.wishlistCount ?? 0, color: 'text-bright-yellow', icon: '📋' },
  { label: '借出中', value: stats.value?.borrowedCount ?? 0, color: 'text-pixel-orange', icon: '📤' },
  { label: '总价值', value: `¥${stats.value?.totalValue ?? 0}`, color: 'text-pixel-pink', icon: '💰' },
  { label: '游玩时长', value: `${stats.value?.totalPlayTime ?? 0}h`, color: 'text-neon-blue', icon: '⏱️' },
  { label: '本月新增', value: stats.value?.newThisMonth ?? 0, color: 'text-pixel-green', icon: '✨' },
  { label: '本月通关', value: stats.value?.completedThisMonth ?? 0, color: 'text-bright-yellow', icon: '🎯' }
])

const quickActions = [
  { label: '添加卡带', route: '/cartridges/new', color: 'pixel-btn-primary', icon: '➕' },
  { label: '添加通关', route: '/playthroughs/new', color: 'pixel-btn-success', icon: '✅' },
  { label: '待玩清单', route: '/wishlist', color: '', icon: '📋' },
  { label: '虚拟展柜', route: '/showcase', color: '', icon: '🖼️' }
]

const timelineItems = computed(() => {
  const items: Array<{
    type: 'cartridge' | 'playthrough'
    title: string
    subtitle: string
    date: string
    color: string
  }> = []

  recentCartridges.value.forEach(c => {
    items.push({
      type: 'cartridge',
      title: c.title,
      subtitle: `新卡带入库 · ${c.platform}`,
      date: c.createdAt,
      color: 'var(--neon-blue)'
    })
  })

  recentPlaythroughs.value.forEach(p => {
    items.push({
      type: 'playthrough',
      title: p.cartridge?.title ?? '未知游戏',
      subtitle: `通关记录 · ${p.playTimeHours}小时`,
      date: p.completionDate || p.createdAt,
      color: 'var(--pixel-green)'
    })
  })

  return items.sort((a, b) => new Date(b.date).getTime() - new Date(a.date).getTime()).slice(0, 8)
})

const formatDate = (dateStr: string) => {
  const date = new Date(dateStr)
  return `${date.getMonth() + 1}/${date.getDate()}`
}

const loadData = async () => {
  loading.value = true
  try {
    await store.fetchStats()
    const [cartRes, playRes] = await Promise.all([
      cartridgeApi.getList({ pageSize: 5 }),
      playthroughApi.getList({ pageSize: 5 })
    ])
    if (cartRes?.code === 0) recentCartridges.value = Array.isArray(cartRes?.data?.items) ? cartRes.data.items : []
    if (playRes?.code === 0) recentPlaythroughs.value = Array.isArray(playRes?.data?.items) ? playRes.data.items : []
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadData()
})
</script>

<template>
  <div class="p-6 space-y-8">
    <div class="text-center py-8 pixel-fade-in" style="animation-delay: 0.1s">
      <h1 class="text-3xl text-neon-blue glow-blue mb-2">PRESS START</h1>
      <p class="text-text-secondary text-lg">欢迎回到卡带档案管理系统</p>
    </div>

    <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
      <div
        v-for="(stat, index) in statCards"
        :key="stat.label"
        class="stat-card pixel-fade-in"
        :style="{ animationDelay: `${0.2 + index * 0.05}s` }"
      >
        <div class="text-2xl mb-2">{{ stat.icon }}</div>
        <div class="stat-number" :class="stat.color">
          {{ stat.value }}
        </div>
        <div class="text-text-secondary mt-2 text-sm">{{ stat.label }}</div>
      </div>
    </div>

    <div class="grid grid-cols-2 md:grid-cols-4 gap-4 pixel-fade-in" style="animation-delay: 0.6s">
      <button
        v-for="action in quickActions"
        :key="action.label"
        class="pixel-btn py-8 text-lg flex flex-col items-center gap-3"
        :class="action.color"
        @click="router.push(action.route)"
      >
        <span class="text-3xl">{{ action.icon }}</span>
        <span>{{ action.label }}</span>
      </button>
    </div>

    <div class="pixel-card p-6 pixel-fade-in" style="animation-delay: 0.7s">
      <h2 class="text-bright-yellow glow-yellow mb-6">最近动态</h2>
      <div v-if="loading" class="text-center py-8 text-text-secondary">
        加载中...
      </div>
      <div v-else-if="timelineItems.length === 0" class="text-center py-8 text-text-secondary">
        暂无动态记录
      </div>
      <div v-else class="space-y-4">
        <div
          v-for="(item, index) in timelineItems"
          :key="index"
          class="timeline-item py-2"
          :style="{ borderLeftColor: item.color }"
        >
          <div class="flex justify-between items-start">
            <div>
              <div class="font-bold text-text-primary">{{ item.title }}</div>
              <div class="text-text-secondary text-sm">{{ item.subtitle }}</div>
            </div>
            <div class="text-text-secondary text-sm pixel-font">
              {{ formatDate(item.date) }}
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
