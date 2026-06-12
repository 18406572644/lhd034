<script setup lang="ts">
import { ref, reactive, onMounted, watch, computed } from 'vue'
import { useRouter } from 'vue-router'
import { Message } from '@arco-design/web-vue'
import { useCartridgeStore } from '@/stores/cartridge'
import { ConditionLabels, ConditionOptions, PlayStatusLabels, PlayStatusBadgeClass } from '@/types'
import type { Cartridge } from '@/types'
import { sessionApi, cartridgeApi } from '@/api'

const router = useRouter()
const store = useCartridgeStore()

const page = ref(1)
const pageSize = ref(12)
const deleteConfirmId = ref<number | null>(null)
const statusDialog = ref(false)
const statusTarget = ref<Cartridge | null>(null)
const selectedStatus = ref<'unstarted' | 'playing' | 'completed' | 'shelved'>('unstarted')
const newSessionDialog = ref(false)
const newSessionTarget = ref<Cartridge | null>(null)

const newSessionForm = reactive({
  sessionDate: new Date().toISOString().split('T')[0],
  durationMinutes: 60,
  progressPercent: 0,
  notes: ''
})

const filters = reactive({
  search: '',
  platform: '',
  publisher: '',
  condition: '',
  year: '',
  status: ''
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

const getCurrentProgress = (cartridge: Cartridge | null | undefined) => {
  if (!cartridge || !cartridge.sessions || cartridge.sessions.length === 0) return 0
  const percents = cartridge.sessions.map(s => s?.progressPercent ?? 0)
  if (percents.length === 0) return 0
  return Math.max(...percents)
}

const openStatusDialog = (cartridge: Cartridge) => {
  if (!cartridge) return
  statusTarget.value = cartridge
  selectedStatus.value = (cartridge.status || 'unstarted') as typeof selectedStatus.value
  statusDialog.value = true
}

const saveStatus = async () => {
  if (!statusTarget.value) return
  try {
    await cartridgeApi.update(statusTarget.value.id, { status: selectedStatus.value })
    Message.success('状态已更新')
    statusDialog.value = false
    loadList()
  } catch (e) {
    Message.error('更新失败')
  }
}

const setStatus = (key: string) => {
  selectedStatus.value = key as typeof selectedStatus.value
}

const openNewSessionDialog = async (cartridge: Cartridge) => {
  if (!cartridge) return
  newSessionTarget.value = cartridge
  newSessionForm.sessionDate = new Date().toISOString().split('T')[0]
  newSessionForm.durationMinutes = 60
  try {
    const progress = await cartridgeApi.getProgress(cartridge.id)
    newSessionForm.progressPercent = progress?.data?.currentProgress ?? getCurrentProgress(cartridge)
  } catch {
    newSessionForm.progressPercent = getCurrentProgress(cartridge)
  }
  newSessionForm.notes = ''
  newSessionDialog.value = true
}

const saveNewSession = async () => {
  if (!newSessionTarget.value) return
  try {
    await sessionApi.create({
      cartridgeId: newSessionTarget.value.id,
      sessionDate: newSessionForm.sessionDate,
      durationMinutes: newSessionForm.durationMinutes,
      progressPercent: newSessionForm.progressPercent,
      notes: newSessionForm.notes
    })
    Message.success('会话记录已保存')
    newSessionDialog.value = false
    loadList()
  } catch (e) {
    Message.error('保存失败')
  }
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
      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-6 gap-4">
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
          <label class="block text-sm text-text-secondary mb-1 pixel-font">游玩状态</label>
          <select v-model="filters.status" class="pixel-input w-full" @change="handleFilterChange">
            <option value="">全部状态</option>
            <option value="unstarted">未开始</option>
            <option value="playing">进行中</option>
            <option value="completed">已通关</option>
            <option value="shelved">搁置</option>
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
        :key="cartridge?.id ?? Math.random()"
        class="relative group"
      >
        <div class="cartridge-case cursor-pointer" @click="router.push(`/cartridges/${cartridge?.id}`)">
          <div class="cartridge-label">
            <img v-if="cartridge?.coverImage" :src="cartridge.coverImage" :alt="cartridge?.title || '卡带封面'" />
            <div v-else class="flex flex-col items-center justify-center h-full">
              <div class="text-xs pixel-font text-bright-yellow leading-relaxed">{{ cartridge?.title || '未知游戏' }}</div>
              <div class="text-xs text-neon-blue mt-1">{{ cartridge?.platform || '—' }}</div>
            </div>
          </div>
        </div>

        <div class="absolute top-2 right-2 flex flex-col gap-1 opacity-0 group-hover:opacity-100 transition-opacity z-10">
          <button
            class="pixel-btn !p-2 !text-xs"
            title="记录游玩会话"
            @click.stop="openNewSessionDialog(cartridge)"
          >
            ➕
          </button>
          <button
            class="pixel-btn !p-2 !text-xs"
            title="修改状态"
            @click.stop="openStatusDialog(cartridge)"
          >
            🏷️
          </button>
          <button
            class="pixel-btn !p-2 !text-xs"
            title="编辑"
            @click.stop="router.push(`/cartridges/${cartridge?.id}/edit`)"
          >
            ✏️
          </button>
          <button
            class="pixel-btn pixel-btn-danger !p-2 !text-xs"
            title="删除"
            @click.stop="deleteConfirmId = cartridge?.id ?? null"
          >
            🗑️
          </button>
        </div>

        <div class="mt-3 space-y-2">
          <div class="flex items-start justify-between gap-2">
            <h4 class="pixel-font text-bright-yellow text-xs leading-tight truncate">{{ cartridge?.title || '未知游戏' }}</h4>
            <span
              class="pixel-badge !text-[8px] shrink-0"
              :class="PlayStatusBadgeClass[cartridge?.status || 'unstarted']"
            >
              {{ PlayStatusLabels[cartridge?.status || 'unstarted'] }}
            </span>
          </div>
          <div
            v-if="cartridge?.status === 'playing' && getCurrentProgress(cartridge) > 0"
            class="pixel-progress !h-2"
          >
            <div
              class="pixel-progress-bar"
              :style="{ width: `${getCurrentProgress(cartridge)}%` }"
            ></div>
          </div>
          <div class="flex items-center gap-2 flex-wrap">
            <span class="pixel-badge !text-[8px]">{{ cartridge?.platform || '—' }}</span>
            <span class="pixel-badge !text-[8px]" :class="conditionBadgeClass(cartridge?.condition || 'good')">
              {{ ConditionLabels[cartridge?.condition || 'good'] }}
            </span>
          </div>
          <div class="flex items-center justify-between text-sm text-text-secondary">
            <span>📅 {{ cartridge?.releaseYear ?? '—' }}</span>
            <span>💰 ¥{{ cartridge?.purchasePrice ?? 0 }}</span>
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

    <div
      v-if="statusDialog"
      class="fixed inset-0 bg-black/70 flex items-center justify-center z-50"
      @click.self="statusDialog = false"
    >
      <div class="pixel-card p-6 max-w-md w-full mx-4">
        <h3 class="text-bright-yellow mb-4">修改游玩状态</h3>
        <p class="text-text-secondary mb-4">
          当前卡带: <span class="text-neon-blue">{{ statusTarget?.title }}</span>
        </p>
        <div class="space-y-2 mb-6">
          <label
            v-for="(label, key) in PlayStatusLabels"
            :key="key"
            class="flex items-center gap-3 cursor-pointer p-3 pixel-border"
            :class="{ '!border-bright-yellow': selectedStatus === key }"
            @click="setStatus(key)"
          >
            <div
              class="w-4 h-4 border-2"
              :class="selectedStatus === key ? 'bg-bright-yellow border-bright-yellow' : 'border-neon-blue'"
            ></div>
            <span class="pixel-badge" :class="PlayStatusBadgeClass[key]">{{ label }}</span>
          </label>
        </div>
        <div class="flex gap-4 justify-end">
          <button class="pixel-btn" @click="statusDialog = false">取消</button>
          <button class="pixel-btn pixel-btn-primary" @click="saveStatus">保存</button>
        </div>
      </div>
    </div>

    <div
      v-if="newSessionDialog"
      class="fixed inset-0 bg-black/70 flex items-center justify-center z-50"
      @click.self="newSessionDialog = false"
    >
      <div class="pixel-card p-6 max-w-md w-full mx-4">
        <h3 class="text-bright-yellow mb-4">记录游玩会话</h3>
        <p class="text-text-secondary mb-4">
          当前卡带: <span class="text-neon-blue">{{ newSessionTarget?.title }}</span>
        </p>
        <div class="space-y-4">
          <div>
            <label class="block text-sm text-text-secondary mb-1 pixel-font">游玩日期</label>
            <input
              v-model="newSessionForm.sessionDate"
              type="date"
              class="pixel-input w-full"
            />
          </div>
          <div>
            <label class="block text-sm text-text-secondary mb-1 pixel-font">游玩时长（分钟）</label>
            <input
              v-model.number="newSessionForm.durationMinutes"
              type="number"
              min="0"
              class="pixel-input w-full"
              placeholder="如: 90"
            />
          </div>
          <div>
            <label class="block text-sm text-text-secondary mb-1 pixel-font">当前进度百分比</label>
            <input
              v-model.number="newSessionForm.progressPercent"
              type="range"
              min="0"
              max="100"
              class="w-full"
            />
            <div class="text-center pixel-font text-neon-blue mt-1">
              {{ newSessionForm.progressPercent }}%
            </div>
          </div>
          <div>
            <label class="block text-sm text-text-secondary mb-1 pixel-font">备注</label>
            <textarea
              v-model="newSessionForm.notes"
              class="pixel-input w-full"
              rows="3"
              placeholder="记录本次游玩的内容、感受..."
            ></textarea>
          </div>
        </div>
        <div class="flex gap-4 justify-end mt-6">
          <button class="pixel-btn" @click="newSessionDialog = false">取消</button>
          <button class="pixel-btn pixel-btn-primary" @click="saveNewSession">保存</button>
        </div>
      </div>
    </div>
  </div>
</template>
