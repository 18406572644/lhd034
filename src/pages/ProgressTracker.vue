<script setup lang="ts">
import { ref, reactive, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { Message } from '@arco-design/web-vue'
import { sessionApi, playthroughApi, cartridgeApi } from '@/api'
import type { PlayingCartridgeProgress, PlayingSession, Cartridge } from '@/types'

const router = useRouter()
const loading = ref(false)
const playingList = ref<PlayingCartridgeProgress[]>([])
const cartridges = ref<Cartridge[]>([])

const sessionDialog = ref(false)
const selectedCartridge = ref<PlayingCartridgeProgress | null>(null)
const timelineDialog = ref(false)
const timelineCartridge = ref<PlayingCartridgeProgress | null>(null)
const timelineSessions = ref<PlayingSession[]>([])

const sessionForm = reactive({
  sessionDate: new Date().toISOString().split('T')[0],
  durationMinutes: 60,
  progressPercent: 0,
  notes: ''
})

const formatDate = (d: string) => {
  const dt = new Date(d)
  return `${dt.getFullYear()}.${String(dt.getMonth() + 1).padStart(2, '0')}.${String(dt.getDate()).padStart(2, '0')}`
}

const formatDuration = (minutes: number) => {
  if (minutes < 60) return `${minutes}分钟`
  const hours = Math.floor(minutes / 60)
  const mins = minutes % 60
  return mins > 0 ? `${hours}小时${mins}分钟` : `${hours}小时`
}

const formatRemaining = (minutes: number | null) => {
  if (minutes === null || minutes === undefined) return '未知'
  const hours = minutes / 60
  if (hours < 1) return `约${Math.round(minutes)}分钟`
  if (hours < 24) return `约${hours.toFixed(1)}小时`
  const days = hours / 24
  return `约${days.toFixed(1)}天`
}

const loadData = async () => {
  loading.value = true
  try {
    const [playRes, cartRes] = await Promise.all([
      sessionApi.getPlaying(),
      cartridgeApi.getList({ pageSize: 100 })
    ])
    playingList.value = playRes.data
    cartridges.value = cartRes.data.items
  } catch (e) {
    console.error('加载数据失败', e)
  } finally {
    loading.value = false
  }
}

const openSessionDialog = (item: PlayingCartridgeProgress) => {
  selectedCartridge.value = item
  sessionForm.sessionDate = new Date().toISOString().split('T')[0]
  sessionForm.durationMinutes = 60
  sessionForm.progressPercent = item.currentProgress
  sessionForm.notes = ''
  sessionDialog.value = true
}

const saveSession = async () => {
  if (!selectedCartridge.value) return
  try {
    await sessionApi.create({
      cartridgeId: selectedCartridge.value.cartridge.id,
      sessionDate: sessionForm.sessionDate,
      durationMinutes: sessionForm.durationMinutes,
      progressPercent: sessionForm.progressPercent,
      notes: sessionForm.notes
    })
    Message.success('会话记录已保存')
    sessionDialog.value = false
    loadData()
  } catch (e) {
    Message.error('保存失败')
  }
}

const openTimeline = async (item: PlayingCartridgeProgress) => {
  timelineCartridge.value = item
  try {
    const res = await sessionApi.getList({ cartridgeId: item.cartridge.id })
    timelineSessions.value = res.data
    timelineDialog.value = true
  } catch (e) {
    Message.error('加载会话记录失败')
  }
}

const markComplete = async (item: PlayingCartridgeProgress) => {
  try {
    const totalHours = item.totalMinutes / 60
    await playthroughApi.create({
      cartridgeId: item.cartridge.id,
      startDate: item.latestSession?.sessionDate || new Date().toISOString().split('T')[0],
      completionDate: new Date().toISOString().split('T')[0],
      playTimeHours: totalHours > 0 ? Math.round(totalHours * 10) / 10 : 10,
      difficultyRating: 3,
      endingType: '标准结局',
      multipleEndings: false,
      achievedEndings: [],
      notes: ''
    })
    await cartridgeApi.update(item.cartridge.id, { status: 'completed' })
    Message.success('已标记为通关')
    loadData()
  } catch (e) {
    console.error('标记通关失败', e)
    Message.error('标记通关失败')
  }
}

const markShelved = async (item: PlayingCartridgeProgress) => {
  try {
    await cartridgeApi.update(item.cartridge.id, { status: 'shelved' })
    Message.success('已标记为搁置')
    loadData()
  } catch (e) {
    Message.error('操作失败')
  }
}

onMounted(loadData)
</script>

<template>
  <div class="p-6 space-y-6">
    <div class="flex flex-col md:flex-row md:items-center md:justify-between gap-4">
      <div>
        <h1 class="text-bright-yellow glow-yellow">游戏进度追踪</h1>
        <p class="text-text-secondary">正在游玩 {{ playingList.length }} 款游戏</p>
      </div>
      <button class="pixel-btn pixel-btn-primary" @click="router.push('/cartridges')">
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

    <div v-else-if="playingList.length === 0" class="text-center py-16">
      <div class="text-6xl mb-4">🎮</div>
      <h3 class="text-bright-yellow mb-2">暂无正在游玩的游戏</h3>
      <p class="text-text-secondary mb-6">从卡带列表中选择一款游戏开始你的冒险吧！</p>
      <button class="pixel-btn pixel-btn-primary" @click="router.push('/cartridges')">
        前往卡带列表
      </button>
    </div>

    <div v-else class="grid grid-cols-1 md:grid-cols-2 gap-4">
      <div v-for="item in playingList" :key="item.cartridge.id" class="pixel-card p-4">
        <div class="flex justify-between items-start mb-3">
          <div>
            <h4 class="pixel-font text-bright-yellow text-sm mb-1">
              {{ item.cartridge.title }}
            </h4>
            <span class="pixel-badge !text-[10px] pixel-badge-warning">
              进行中
            </span>
          </div>
          <span class="pixel-font text-neon-blue text-lg">
            {{ item.currentProgress }}%
          </span>
        </div>

        <div class="pixel-progress mb-3">
          <div
            class="pixel-progress-bar"
            :style="{ width: `${item.currentProgress}%` }"
          ></div>
        </div>

        <div class="space-y-2 text-sm">
          <div class="flex items-center gap-2">
            <span class="text-text-secondary">📅</span>
            <span>上次游玩: {{ item.latestSession ? formatDate(item.latestSession.sessionDate) : '—' }}</span>
          </div>
          <div class="flex items-center gap-2">
            <span class="text-text-secondary">⏱️</span>
            <span>已游玩: {{ formatDuration(item.totalMinutes) }} ({{ item.totalSessions }}次)</span>
          </div>
          <div class="flex items-center gap-2">
            <span class="text-text-secondary">🔮</span>
            <span>预计剩余: {{ formatRemaining(item.estimatedRemaining) }}</span>
          </div>
          <div class="flex items-center gap-2">
            <span class="text-text-secondary">📝</span>
            <span v-if="item.latestSession?.notes">{{ item.latestSession.notes }}</span>
            <span v-else class="text-text-secondary">暂无备注</span>
          </div>
        </div>

        <div class="flex gap-2 mt-4 flex-wrap">
          <button class="pixel-btn flex-1 !text-xs" @click="openSessionDialog(item)">
            ➕ 记录会话
          </button>
          <button class="pixel-btn !text-xs" @click="openTimeline(item)">
            📜 时间轴
          </button>
          <button class="pixel-btn pixel-btn-success flex-1 !text-xs" @click="markComplete(item)">
            🏆 标记通关
          </button>
          <button class="pixel-btn pixel-btn-danger !text-xs" @click="markShelved(item)">
            😴 搁置
          </button>
        </div>
      </div>
    </div>

    <div
      v-if="sessionDialog"
      class="fixed inset-0 bg-black/70 flex items-center justify-center z-50"
      @click.self="sessionDialog = false"
    >
      <div class="pixel-card p-6 max-w-md w-full mx-4">
        <h3 class="text-bright-yellow mb-4">记录游玩会话</h3>
        <div class="space-y-4">
          <div>
            <label class="block text-sm text-text-secondary mb-1 pixel-font">游玩日期</label>
            <input
              v-model="sessionForm.sessionDate"
              type="date"
              class="pixel-input w-full"
            />
          </div>
          <div>
            <label class="block text-sm text-text-secondary mb-1 pixel-font">游玩时长（分钟）</label>
            <input
              v-model.number="sessionForm.durationMinutes"
              type="number"
              min="0"
              class="pixel-input w-full"
              placeholder="如: 90"
            />
          </div>
          <div>
            <label class="block text-sm text-text-secondary mb-1 pixel-font">当前进度百分比</label>
            <input
              v-model.number="sessionForm.progressPercent"
              type="range"
              min="0"
              max="100"
              class="w-full"
            />
            <div class="text-center pixel-font text-neon-blue mt-1">{{ sessionForm.progressPercent }}%</div>
          </div>
          <div>
            <label class="block text-sm text-text-secondary mb-1 pixel-font">备注</label>
            <textarea
              v-model="sessionForm.notes"
              class="pixel-input w-full"
              rows="3"
              placeholder="记录本次游玩的内容、感受..."
            ></textarea>
          </div>
        </div>
        <div class="flex gap-4 justify-end mt-6">
          <button class="pixel-btn" @click="sessionDialog = false">取消</button>
          <button class="pixel-btn pixel-btn-primary" @click="saveSession">保存</button>
        </div>
      </div>
    </div>

    <div
      v-if="timelineDialog"
      class="fixed inset-0 bg-black/70 flex items-center justify-center z-50"
      @click.self="timelineDialog = false"
    >
      <div class="pixel-card p-6 max-w-2xl w-full mx-4 max-h-[80vh] overflow-hidden flex flex-col">
        <div class="flex items-center justify-between mb-4">
          <h3 class="text-bright-yellow">
            {{ timelineCartridge?.cartridge.title }} - 游玩时间轴
          </h3>
          <button class="pixel-btn !py-2 !px-3 !text-xs" @click="timelineDialog = false">
            ✕
          </button>
        </div>

        <div v-if="timelineSessions.length === 0" class="text-center py-8 text-text-secondary">
          暂无游玩会话记录
        </div>

        <div v-else class="space-y-4 overflow-y-auto scroll-hidden pr-2">
          <div
            v-for="(session, idx) in timelineSessions"
            :key="session.id"
            class="timeline-item pb-4"
          >
            <div class="flex items-center justify-between mb-2">
              <span class="pixel-font text-neon-blue text-sm">
                第 {{ timelineSessions.length - idx }} 次游玩
              </span>
              <span class="text-text-secondary text-sm">{{ formatDate(session.sessionDate) }}</span>
            </div>
            <div class="grid grid-cols-2 gap-2 text-sm mb-2">
              <div>⏱️ 时长: {{ formatDuration(session.durationMinutes) }}</div>
              <div>📊 进度: {{ session.progressPercent }}%</div>
            </div>
            <div class="pixel-progress !h-3 mb-2">
              <div
                class="pixel-progress-bar"
                :style="{ width: `${session.progressPercent}%` }"
              ></div>
            </div>
            <div v-if="session.notes" class="text-text-secondary text-sm">
              📝 {{ session.notes }}
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
