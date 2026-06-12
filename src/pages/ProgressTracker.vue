<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { wishlistApi, cartridgeApi, playthroughApi } from '@/api'
import type { WishlistItem, Cartridge } from '@/types'

const router = useRouter()
const loading = ref(false)
const items = ref<WishlistItem[]>([])
const cartridges = ref<Cartridge[]>([])
const progressMap = reactive<Record<number, { progress: number; chapter: string; remaining: string }>>({})

const updateDialog = ref(false)
const currentItem = ref<WishlistItem | null>(null)
const progressForm = reactive({ progress: 0, chapter: '', remaining: '' })

const loadData = async () => {
  loading.value = true
  try {
    const [wishRes, cartRes] = await Promise.all([
      wishlistApi.getList(),
      cartridgeApi.getList({ pageSize: 100 })
    ])
    items.value = wishRes.data
    cartridges.value = cartRes.data.items
    items.value.forEach(item => {
      if (!progressMap[item.id]) {
        progressMap[item.id] = { progress: 0, chapter: '第1章', remaining: '未知' }
      }
    })
  } finally {
    loading.value = false
  }
}

const formatDate = (d: string) => {
  const dt = new Date(d)
  return `${dt.getFullYear()}.${String(dt.getMonth() + 1).padStart(2, '0')}.${String(dt.getDate()).padStart(2, '0')}`
}

const openUpdateDialog = (item: WishlistItem) => {
  currentItem.value = item
  const prog = progressMap[item.id]
  progressForm.progress = prog.progress
  progressForm.chapter = prog.chapter
  progressForm.remaining = prog.remaining
  updateDialog.value = true
}

const saveProgress = () => {
  if (currentItem.value) {
    progressMap[currentItem.value.id] = { ...progressForm }
    updateDialog.value = false
  }
}

const markComplete = async (item: WishlistItem) => {
  try {
    await playthroughApi.create({
      cartridgeId: item.cartridgeId,
      startDate: item.plannedStartDate || new Date().toISOString().split('T')[0],
      completionDate: new Date().toISOString().split('T')[0],
      playTimeHours: progressMap[item.id]?.progress || 10,
      difficultyRating: 3,
      endingType: '标准结局',
      multipleEndings: false,
      achievedEndings: [],
      notes: ''
    })
    await wishlistApi.delete(item.id)
    delete progressMap[item.id]
    loadData()
  } catch (e) {
    console.error('标记通关失败', e)
  }
}

const getCartridge = (id: number) => cartridges.value.find(c => c.id === id)

onMounted(loadData)
</script>

<template>
  <div class="p-6 space-y-6">
    <div class="flex flex-col md:flex-row md:items-center md:justify-between gap-4">
      <div>
        <h1 class="text-bright-yellow glow-yellow">游戏进度追踪</h1>
        <p class="text-text-secondary">正在游玩 {{ items.length }} 款游戏</p>
      </div>
      <button class="pixel-btn pixel-btn-primary" @click="router.push('/wishlist')">
        🎮 选择卡带开始游玩
      </button>
    </div>

    <div v-if="loading" class="grid grid-cols-1 md:grid-cols-2 gap-4">
      <div v-for="i in 4" :key="i" class="pixel-card p-4 opacity-50">
        <div class="h-4 bg-dark-bg-3 w-3/4 animate-pulse mb-4"></div>
        <div class="h-6 bg-dark-bg-3 w-full animate-pulse mb-2"></div>
        <div class="h-3 bg-dark-bg-3 w-1/2 animate-pulse"></div>
      </div>
    </div>

    <div v-else-if="items.length === 0" class="text-center py-16">
      <div class="text-6xl mb-4">🎮</div>
      <h3 class="text-bright-yellow mb-2">暂无正在游玩的游戏</h3>
      <p class="text-text-secondary mb-6">从待玩清单中选择一款游戏开始你的冒险吧！</p>
      <button class="pixel-btn pixel-btn-primary" @click="router.push('/wishlist')">
        前往待玩清单
      </button>
    </div>

    <div v-else class="grid grid-cols-1 md:grid-cols-2 gap-4">
      <div v-for="item in items" :key="item.id" class="pixel-card p-4">
        <div class="flex justify-between items-start mb-3">
          <div>
            <h4 class="pixel-font text-bright-yellow text-sm mb-1">
              {{ getCartridge(item.cartridgeId)?.title || '未知游戏' }}
            </h4>
            <span class="pixel-badge !text-[10px]">
              {{ getCartridge(item.cartridgeId)?.platform || '未知平台' }}
            </span>
          </div>
          <span class="pixel-font text-neon-blue text-lg">
            {{ progressMap[item.id]?.progress || 0 }}%
          </span>
        </div>

        <div class="pixel-progress mb-3">
          <div
            class="pixel-progress-bar"
            :style="{ width: `${progressMap[item.id]?.progress || 0}%` }"
          ></div>
        </div>

        <div class="space-y-2 text-sm">
          <div class="flex items-center gap-2">
            <span class="text-text-secondary">📍</span>
            <span>{{ progressMap[item.id]?.chapter || '第1章' }}</span>
          </div>
          <div class="flex items-center gap-2">
            <span class="text-text-secondary">⏱️</span>
            <span>预计剩余: {{ progressMap[item.id]?.remaining || '未知' }}</span>
          </div>
          <div class="flex items-center gap-2">
            <span class="text-text-secondary">📅</span>
            <span>开始于: {{ formatDate(item.plannedStartDate || item.addedAt) }}</span>
          </div>
        </div>

        <div class="flex gap-2 mt-4">
          <button class="pixel-btn flex-1 !text-xs" @click="openUpdateDialog(item)">
            ✏️ 更新进度
          </button>
          <button class="pixel-btn pixel-btn-success flex-1 !text-xs" @click="markComplete(item)">
            🏆 标记通关
          </button>
        </div>
      </div>
    </div>

    <div
      v-if="updateDialog"
      class="fixed inset-0 bg-black/70 flex items-center justify-center z-50"
      @click.self="updateDialog = false"
    >
      <div class="pixel-card p-6 max-w-md w-full mx-4">
        <h3 class="text-bright-yellow mb-4">更新进度</h3>
        <div class="space-y-4">
          <div>
            <label class="block text-sm text-text-secondary mb-1 pixel-font">进度百分比</label>
            <input
              v-model.number="progressForm.progress"
              type="range"
              min="0"
              max="100"
              class="w-full"
            />
            <div class="text-center pixel-font text-neon-blue mt-1">{{ progressForm.progress }}%</div>
          </div>
          <div>
            <label class="block text-sm text-text-secondary mb-1 pixel-font">当前章节/关卡</label>
            <input
              v-model="progressForm.chapter"
              type="text"
              class="pixel-input w-full"
              placeholder="如: 第5章 - 龙之城"
            />
          </div>
          <div>
            <label class="block text-sm text-text-secondary mb-1 pixel-font">预计剩余时长</label>
            <input
              v-model="progressForm.remaining"
              type="text"
              class="pixel-input w-full"
              placeholder="如: 约5小时"
            />
          </div>
        </div>
        <div class="flex gap-4 justify-end mt-6">
          <button class="pixel-btn" @click="updateDialog = false">取消</button>
          <button class="pixel-btn pixel-btn-primary" @click="saveProgress">保存</button>
        </div>
      </div>
    </div>
  </div>
</template>
