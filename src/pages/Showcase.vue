<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { cartridgeApi } from '@/api'
import { ConditionLabels, ConditionOptions } from '@/types'
import type { Cartridge } from '@/types'

const router = useRouter()
const cartridges = ref<Cartridge[]>([])
const platforms = ref<string[]>([])
const publishers = ref<string[]>([])
const loading = ref(true)

const filters = reactive({
  platform: '',
  publisher: '',
  condition: ''
})

const filteredCartridges = computed(() => {
  return cartridges.value.filter(c => {
    if (filters.platform && c.platform !== filters.platform) return false
    if (filters.publisher && c.publisher !== filters.publisher) return false
    if (filters.condition && c.condition !== filters.condition) return false
    return true
  })
})

const groupedByPlatform = computed(() => {
  const groups: Record<string, Cartridge[]> = {}
  filteredCartridges.value.forEach(c => {
    if (!groups[c.platform]) groups[c.platform] = []
    groups[c.platform].push(c)
  })
  return groups
})

const platformOrder = computed(() => {
  return platforms.value.filter(p => groupedByPlatform.value[p])
})

const conditionBadgeClass = (condition: string) => {
  const map: Record<string, string> = {
    mint: 'pixel-badge-success',
    excellent: 'pixel-badge-success',
    good: '',
    fair: 'pixel-badge-warning',
    poor: 'pixel-badge-danger'
  }
  return map[condition] || ''
}

const handleFilterChange = () => {}

const loadData = async () => {
  loading.value = true
  try {
    const [cartRes, platRes, pubRes] = await Promise.all([
      cartridgeApi.getList({ pageSize: 100 }),
      cartridgeApi.getPlatforms(),
      cartridgeApi.getPublishers()
    ])
    if (cartRes.code === 0) cartridges.value = cartRes.data.items
    if (platRes.code === 0) platforms.value = platRes.data
    if (pubRes.code === 0) publishers.value = pubRes.data
  } finally {
    loading.value = false
  }
}

onMounted(loadData)
</script>

<template>
  <div class="p-6 space-y-6 scanline-overlay">
    <div class="text-center py-4 crt-effect">
      <h1 class="text-neon-blue glow-blue mb-2">虚拟展柜</h1>
      <p class="text-text-secondary pixel-font text-xs">VIRTUAL SHOWCASE</p>
    </div>

    <div class="pixel-card p-4">
      <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
        <div>
          <label class="block text-sm text-text-secondary mb-1 pixel-font">平台</label>
          <select v-model="filters.platform" class="pixel-input w-full" @change="handleFilterChange">
            <option value="">全部平台</option>
            <option v-for="p in platforms" :key="p" :value="p">{{ p }}</option>
          </select>
        </div>
        <div>
          <label class="block text-sm text-text-secondary mb-1 pixel-font">发行商</label>
          <select v-model="filters.publisher" class="pixel-input w-full" @change="handleFilterChange">
            <option value="">全部发行商</option>
            <option v-for="p in publishers" :key="p" :value="p">{{ p }}</option>
          </select>
        </div>
        <div>
          <label class="block text-sm text-text-secondary mb-1 pixel-font">品相</label>
          <select v-model="filters.condition" class="pixel-input w-full" @change="handleFilterChange">
            <option value="">全部品相</option>
            <option v-for="c in ConditionOptions" :key="c.value" :value="c.value">{{ c.label }}</option>
          </select>
        </div>
      </div>
    </div>

    <div v-if="loading" class="text-center py-16">
      <div class="text-4xl mb-4 animate-pulse">🎮</div>
      <p class="text-text-secondary pixel-font">加载中...</p>
    </div>

    <div v-else-if="platformOrder.length === 0" class="text-center py-16">
      <div class="text-6xl mb-4">📦</div>
      <h3 class="text-bright-yellow mb-2">展柜空空如也</h3>
      <p class="text-text-secondary">添加一些卡带点亮你的展柜吧！</p>
    </div>

    <div v-else class="space-y-8">
      <div v-for="platform in platformOrder" :key="platform" class="showcase-shelf p-6 pt-10">
        <div class="flex items-center justify-between mb-6">
          <h2 class="text-bright-yellow glow-yellow pixel-font text-sm">
            ▣ {{ platform }}
          </h2>
          <span class="pixel-badge !text-[10px]">
            {{ groupedByPlatform[platform].length }} 张
          </span>
        </div>

        <div class="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-6 xl:grid-cols-8 gap-4">
          <div
            v-for="cartridge in groupedByPlatform[platform]"
            :key="cartridge.id"
            class="pixel-jitter"
          >
            <div
              class="cartridge-case cursor-pointer"
              @click="router.push(`/cartridges/${cartridge.id}`)"
            >
              <div class="cartridge-label">
                <img v-if="cartridge.coverImage" :src="cartridge.coverImage" :alt="cartridge.title" />
                <div v-else class="flex flex-col items-center justify-center h-full">
                  <div class="text-[10px] pixel-font text-bright-yellow leading-tight px-1">
                    {{ cartridge.title }}
                  </div>
                </div>
              </div>
            </div>
            <div class="mt-2 text-center">
              <div class="text-xs text-text-secondary truncate px-1">{{ cartridge.title }}</div>
              <span
                class="pixel-badge !text-[8px] mt-1"
                :class="conditionBadgeClass(cartridge.condition)"
              >
                {{ ConditionLabels[cartridge.condition] }}
              </span>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.showcase-shelf::before {
  animation: neon-pulse 2s ease-in-out infinite;
}

@keyframes neon-pulse {
  0%, 100% {
    opacity: 1;
    box-shadow: 0 0 20px var(--neon-blue);
  }
  50% {
    opacity: 0.8;
    box-shadow: 0 0 30px var(--neon-blue), 0 0 40px var(--pixel-pink);
  }
}
</style>
