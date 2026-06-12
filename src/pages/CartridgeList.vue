<script setup lang="ts">
import { ref, reactive, onMounted, watch, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useCartridgeStore } from '@/stores/cartridge'
import { ConditionLabels, ConditionOptions } from '@/types'
import type { Cartridge } from '@/types'

const router = useRouter()
const store = useCartridgeStore()

const page = ref(1)
const pageSize = ref(12)
const deleteConfirmId = ref<number | null>(null)

const filters = reactive({
  search: '',
  platform: '',
  publisher: '',
  condition: '',
  year: ''
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

const totalPages = computed(() => Math.ceil(store.total / pageSize.value))

const hasPlaythroughs = (cartridge: Cartridge) => {
  return cartridge.playthroughs && cartridge.playthroughs.length > 0
}

const loadList = () => {
  const params = {
    page: page.value,
    pageSize: pageSize.value,
    ...filters
  }
  Object.keys(params).forEach(key => {
    if (!params[key as keyof typeof params]) delete params[key as keyof typeof params]
  })
  store.fetchList(params)
}

const handleSearch = () => {
  page.value = 1
  loadList()
}

const handleFilterChange = () => {
  page.value = 1
  loadList()
}

const handleDelete = async (id: number) => {
  if (await store.deleteCartridge(id)) {
    deleteConfirmId.value = null
    if (store.cartridges.length === 0 && page.value > 1) {
      page.value--
    }
    loadList()
  }
}

const goToPage = (p: number) => {
  if (p >= 1 && p <= totalPages.value) {
    page.value = p
    loadList()
  }
}

const visiblePages = computed(() => {
  const pages: number[] = []
  const total = totalPages.value
  const current = page.value

  if (total <= 7) {
    for (let i = 1; i <= total; i++) pages.push(i)
  } else {
    if (current <= 4) {
      for (let i = 1; i <= 5; i++) pages.push(i)
      pages.push(-1, total)
    } else if (current >= total - 3) {
      pages.push(1, -1)
      for (let i = total - 4; i <= total; i++) pages.push(i)
    } else {
      pages.push(1, -1, current - 1, current, current + 1, -1, total)
    }
  }
  return pages
})

watch(page, loadList)

onMounted(async () => {
  await Promise.all([store.fetchPlatforms(), store.fetchPublishers()])
  loadList()
})
</script>

<template>
  <div class="p-6 space-y-6">
    <div class="flex flex-col md:flex-row md:items-center md:justify-between gap-4">
      <div>
        <h1 class="text-neon-blue glow-blue">卡带档案</h1>
        <p class="text-text-secondary">共 {{ store.total }} 张卡带</p>
      </div>
      <button class="pixel-btn pixel-btn-primary" @click="router.push('/cartridges/new')">
        ➕ 添加卡带
      </button>
    </div>

    <div class="pixel-card p-4">
      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-5 gap-4">
        <div>
          <label class="block text-sm text-text-secondary mb-1 pixel-font">搜索</label>
          <input
            v-model="filters.search"
            type="text"
            class="pixel-input w-full"
            placeholder="游戏标题..."
            @keyup.enter="handleSearch"
          />
        </div>
        <div>
          <label class="block text-sm text-text-secondary mb-1 pixel-font">平台</label>
          <select v-model="filters.platform" class="pixel-input w-full" @change="handleFilterChange">
            <option value="">全部平台</option>
            <option v-for="p in store.platforms" :key="p" :value="p">{{ p }}</option>
          </select>
        </div>
        <div>
          <label class="block text-sm text-text-secondary mb-1 pixel-font">发行商</label>
          <select v-model="filters.publisher" class="pixel-input w-full" @change="handleFilterChange">
            <option value="">全部发行商</option>
            <option v-for="p in store.publishers" :key="p" :value="p">{{ p }}</option>
          </select>
        </div>
        <div>
          <label class="block text-sm text-text-secondary mb-1 pixel-font">品相</label>
          <select v-model="filters.condition" class="pixel-input w-full" @change="handleFilterChange">
            <option value="">全部品相</option>
            <option v-for="c in ConditionOptions" :key="c.value" :value="c.value">{{ c.label }}</option>
          </select>
        </div>
        <div>
          <label class="block text-sm text-text-secondary mb-1 pixel-font">年份</label>
          <input
            v-model="filters.year"
            type="text"
            class="pixel-input w-full"
            placeholder="如: 1998"
            @keyup.enter="handleFilterChange"
            @blur="handleFilterChange"
          />
        </div>
      </div>
    </div>

    <div v-if="store.loading" class="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-6">
      <div v-for="i in 8" :key="i" class="cartridge-case opacity-50">
        <div class="cartridge-label">
          <div class="h-4 bg-dark-bg-3 w-3/4 animate-pulse"></div>
        </div>
      </div>
    </div>

    <div v-else-if="store.cartridges.length === 0" class="text-center py-16">
      <div class="text-6xl mb-4">🎮</div>
      <h3 class="text-bright-yellow mb-2">暂无卡带</h3>
      <p class="text-text-secondary mb-6">点击右上角按钮添加你的第一张卡带</p>
      <button class="pixel-btn pixel-btn-primary" @click="router.push('/cartridges/new')">
        添加第一张卡带
      </button>
    </div>

    <div v-else class="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-6">
      <div
        v-for="cartridge in store.cartridges"
        :key="cartridge.id"
        class="relative group"
      >
        <div class="cartridge-case cursor-pointer" @click="router.push(`/cartridges/${cartridge.id}`)">
          <div class="cartridge-label">
            <img v-if="cartridge.coverImage" :src="cartridge.coverImage" :alt="cartridge.title" />
            <div v-else class="flex flex-col items-center justify-center h-full">
              <div class="text-xs pixel-font text-bright-yellow leading-relaxed">{{ cartridge.title }}</div>
              <div class="text-xs text-neon-blue mt-1">{{ cartridge.platform }}</div>
            </div>
          </div>
        </div>

        <div class="absolute top-2 right-2 flex flex-col gap-1 opacity-0 group-hover:opacity-100 transition-opacity z-10">
          <button
            class="pixel-btn !p-2 !text-xs"
            title="编辑"
            @click.stop="router.push(`/cartridges/${cartridge.id}/edit`)"
          >
            ✏️
          </button>
          <button
            class="pixel-btn pixel-btn-danger !p-2 !text-xs"
            title="删除"
            @click.stop="deleteConfirmId = cartridge.id"
          >
            🗑️
          </button>
        </div>

        <div class="mt-3 space-y-2">
          <div class="flex items-start justify-between gap-2">
            <h4 class="pixel-font text-bright-yellow text-xs leading-tight truncate">{{ cartridge.title }}</h4>
            <span v-if="hasPlaythroughs(cartridge)" class="pixel-badge pixel-badge-success !text-[8px] shrink-0">
              ✓ 通关
            </span>
          </div>
          <div class="flex items-center gap-2 flex-wrap">
            <span class="pixel-badge !text-[8px]">{{ cartridge.platform }}</span>
            <span class="pixel-badge !text-[8px]" :class="conditionBadgeClass(cartridge.condition)">
              {{ ConditionLabels[cartridge.condition] }}
            </span>
          </div>
          <div class="flex items-center justify-between text-sm text-text-secondary">
            <span>📅 {{ cartridge.releaseYear }}</span>
            <span>💰 ¥{{ cartridge.purchasePrice }}</span>
          </div>
        </div>
      </div>
    </div>

    <div v-if="totalPages > 1 && !store.loading" class="flex justify-center items-center gap-2 pt-4">
      <button
        class="pixel-btn !py-2 !px-3"
        :disabled="page === 1"
        @click="goToPage(page - 1)"
      >
        ◀
      </button>
      <template v-for="p in visiblePages" :key="p">
        <span v-if="p === -1" class="text-text-secondary px-2">...</span>
        <button
          v-else
          class="pixel-btn !py-2 !px-4"
          :class="{ 'pixel-btn-primary': p === page }"
          @click="goToPage(p)"
        >
          {{ p }}
        </button>
      </template>
      <button
        class="pixel-btn !py-2 !px-3"
        :disabled="page === totalPages"
        @click="goToPage(page + 1)"
      >
        ▶
      </button>
    </div>

    <div
      v-if="deleteConfirmId !== null"
      class="fixed inset-0 bg-black/70 flex items-center justify-center z-50"
      @click.self="deleteConfirmId = null"
    >
      <div class="pixel-card p-6 max-w-sm w-full mx-4">
        <h3 class="text-bright-yellow mb-4">确认删除</h3>
        <p class="text-text-secondary mb-6">确定要删除这张卡带吗？此操作无法撤销。</p>
        <div class="flex gap-4 justify-end">
          <button class="pixel-btn" @click="deleteConfirmId = null">取消</button>
          <button class="pixel-btn pixel-btn-danger" @click="handleDelete(deleteConfirmId)">
            确认删除
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
