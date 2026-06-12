<script setup lang="ts">
import { ref, reactive, onMounted, computed } from 'vue'
import { borrowApi, cartridgeApi } from '@/api'
import { StatusLabels } from '@/types'
import type { BorrowRecord, Cartridge } from '@/types'

const activeTab = ref<'borrowed' | 'history'>('borrowed')
const records = ref<BorrowRecord[]>([])
const cartridges = ref<Cartridge[]>([])
const loading = ref(true)
const searchQuery = ref('')
const statusFilter = ref('')

const showBorrowDialog = ref(false)
const showExtendDialog = ref(false)
const showReturnDialog = ref(false)
const selectedRecord = ref<BorrowRecord | null>(null)

const borrowForm = reactive({
  cartridgeId: 0,
  borrowerName: '',
  borrowerContact: '',
  borrowDate: new Date().toISOString().split('T')[0],
  expectedReturnDate: '',
  notes: ''
})

const extendForm = reactive({
  expectedReturnDate: ''
})

const returnForm = reactive({
  actualReturnDate: new Date().toISOString().split('T')[0]
})

const calcDaysRemaining = (expected: string) => {
  const now = new Date()
  now.setHours(0, 0, 0, 0)
  const exp = new Date(expected)
  exp.setHours(0, 0, 0, 0)
  return Math.ceil((exp.getTime() - now.getTime()) / (1000 * 60 * 60 * 24))
}

const calcBorrowDuration = (borrowDate: string, returnDate: string) => {
  const start = new Date(borrowDate)
  const end = new Date(returnDate)
  return Math.ceil((end.getTime() - start.getTime()) / (1000 * 60 * 60 * 24))
}

const getStatusInfo = (record: BorrowRecord) => {
  if (record.status === 'returned') {
    return { text: '已归还', class: 'pixel-badge-success', daysClass: 'text-pixel-green' }
  }
  if (record.status === 'overdue') {
    return { text: '已逾期', class: 'pixel-badge-danger', daysClass: 'text-pixel-red pixel-blink' }
  }
  const days = calcDaysRemaining(record.expectedReturnDate)
  if (days > 7) return { text: '正常', class: 'pixel-badge-success', daysClass: 'text-pixel-green' }
  if (days >= 3) return { text: '即将到期', class: 'pixel-badge-warning', daysClass: 'text-bright-yellow' }
  if (days >= 0) return { text: '即将到期', class: 'pixel-badge-warning', daysClass: 'text-pixel-orange' }
  return { text: '已逾期', class: 'pixel-badge-danger', daysClass: 'text-pixel-red pixel-blink' }
}

const borrowedRecords = computed(() => {
  return records.value
    .filter(r => r.status !== 'returned')
    .filter(r => !searchQuery.value || r.borrowerName.includes(searchQuery.value))
    .filter(r => !statusFilter.value || r.status === statusFilter.value)
})

const historyRecords = computed(() => {
  return records.value
    .filter(r => r.status === 'returned')
    .filter(r => !searchQuery.value || r.borrowerName.includes(searchQuery.value) || (r.cartridge?.title || '').includes(searchQuery.value))
})

const loadData = async () => {
  loading.value = true
  try {
    const [res1, res2] = await Promise.all([
      borrowApi.getList(),
      cartridgeApi.getList({ pageSize: 1000 })
    ])
    if (res1.code === 0) records.value = res1.data
    if (res2.code === 0) cartridges.value = res2.data.items
  } finally {
    loading.value = false
  }
}

const openBorrowDialog = () => {
  Object.assign(borrowForm, {
    cartridgeId: 0,
    borrowerName: '',
    borrowerContact: '',
    borrowDate: new Date().toISOString().split('T')[0],
    expectedReturnDate: '',
    notes: ''
  })
  showBorrowDialog.value = true
}

const openExtendDialog = (record: BorrowRecord) => {
  selectedRecord.value = record
  extendForm.expectedReturnDate = record.expectedReturnDate
  showExtendDialog.value = true
}

const openReturnDialog = (record: BorrowRecord) => {
  selectedRecord.value = record
  returnForm.actualReturnDate = new Date().toISOString().split('T')[0]
  showReturnDialog.value = true
}

const handleBorrow = async () => {
  if (!borrowForm.cartridgeId || !borrowForm.borrowerName || !borrowForm.expectedReturnDate) return
  const res = await borrowApi.create(borrowForm)
  if (res.code === 0) {
    showBorrowDialog.value = false
    loadData()
  }
}

const handleExtend = async () => {
  if (!selectedRecord.value || !extendForm.expectedReturnDate) return
  const res = await borrowApi.update(selectedRecord.value.id, { expectedReturnDate: extendForm.expectedReturnDate })
  if (res.code === 0) {
    showExtendDialog.value = false
    loadData()
  }
}

const handleReturn = async () => {
  if (!selectedRecord.value || !returnForm.actualReturnDate) return
  const res = await borrowApi.update(selectedRecord.value.id, { actualReturnDate: returnForm.actualReturnDate })
  if (res.code === 0) {
    showReturnDialog.value = false
    loadData()
  }
}

const availableCartridges = computed(() => {
  const borrowedIds = records.value.filter(r => r.status !== 'returned').map(r => r.cartridgeId)
  return cartridges.value.filter(c => !borrowedIds.includes(c.id))
})

onMounted(loadData)
</script>

<template>
  <div class="p-6 space-y-6">
    <div class="flex flex-col md:flex-row md:items-center md:justify-between gap-4">
      <div>
        <h1 class="text-neon-blue glow-blue">卡带借还中心</h1>
        <p class="text-text-secondary">共 {{ records.length }} 条借还记录</p>
      </div>
      <button class="pixel-btn pixel-btn-primary" @click="openBorrowDialog">
        ➕ 登记借出
      </button>
    </div>

    <div class="pixel-card p-4">
      <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
        <div>
          <label class="block text-sm text-text-secondary mb-1 pixel-font">搜索借用人</label>
          <input
            v-model="searchQuery"
            type="text"
            class="pixel-input w-full"
            placeholder="输入借用人姓名..."
          />
        </div>
        <div>
          <label class="block text-sm text-text-secondary mb-1 pixel-font">状态筛选</label>
          <select v-model="statusFilter" class="pixel-input w-full">
            <option value="">全部状态</option>
            <option value="borrowed">借出中</option>
            <option value="overdue">已逾期</option>
            <option value="returned">已归还</option>
          </select>
        </div>
      </div>
    </div>

    <div class="flex gap-2">
      <button
        class="pixel-btn"
        :class="{ 'pixel-btn-primary': activeTab === 'borrowed' }"
        @click="activeTab = 'borrowed'"
      >
        📤 借出中 ({{ borrowedRecords.length }})
      </button>
      <button
        class="pixel-btn"
        :class="{ 'pixel-btn-primary': activeTab === 'history' }"
        @click="activeTab = 'history'"
      >
        📚 历史记录 ({{ historyRecords.length }})
      </button>
    </div>

    <div v-if="loading" class="text-center py-16">
      <div class="text-4xl mb-4 animate-pulse">📤</div>
      <p class="text-text-secondary pixel-font">加载中...</p>
    </div>

    <div v-else-if="activeTab === 'borrowed' && borrowedRecords.length === 0" class="text-center py-16">
      <div class="text-6xl mb-4">📦</div>
      <h3 class="text-bright-yellow mb-2">暂无借出记录</h3>
      <p class="text-text-secondary">所有卡带都在展柜中</p>
    </div>

    <div v-else-if="activeTab === 'history' && historyRecords.length === 0" class="text-center py-16">
      <div class="text-6xl mb-4">📚</div>
      <h3 class="text-bright-yellow mb-2">暂无历史记录</h3>
      <p class="text-text-secondary">还没有完成的借还记录</p>
    </div>

    <div v-else class="space-y-4">
      <div
        v-for="record in (activeTab === 'borrowed' ? borrowedRecords : historyRecords)"
        :key="record.id"
        class="pixel-card p-4"
      >
        <div class="flex flex-col md:flex-row gap-4">
          <div class="w-24 h-32 shrink-0">
            <div class="cartridge-case">
              <div class="cartridge-label">
                <img v-if="record.cartridge?.coverImage" :src="record.cartridge.coverImage" :alt="record.cartridge.title" />
                <div v-else class="flex items-center justify-center h-full">
                  <span class="text-xs pixel-font text-bright-yellow">{{ record.cartridge?.title || '未知' }}</span>
                </div>
              </div>
            </div>
          </div>
          <div class="flex-1 space-y-2">
            <div class="flex flex-wrap items-center justify-between gap-2">
              <h4 class="pixel-font text-bright-yellow text-sm">
                {{ record.cartridge?.title || '未知卡带' }}
              </h4>
              <span class="pixel-badge" :class="getStatusInfo(record).class">
                {{ getStatusInfo(record).text }}
              </span>
            </div>
            <div class="grid grid-cols-2 md:grid-cols-3 gap-2 text-sm">
              <div>
                <span class="text-text-secondary">借用人：</span>
                <span>{{ record.borrowerName }}</span>
              </div>
              <div>
                <span class="text-text-secondary">借出日期：</span>
                <span class="pixel-font">{{ record.borrowDate }}</span>
              </div>
              <div>
                <span class="text-text-secondary">预计归还：</span>
                <span class="pixel-font">{{ record.expectedReturnDate }}</span>
              </div>
              <div v-if="record.status !== 'returned'">
                <span class="text-text-secondary">剩余天数：</span>
                <span :class="getStatusInfo(record).daysClass" class="pixel-font font-bold">
                  {{ calcDaysRemaining(record.expectedReturnDate) }} 天
                </span>
              </div>
              <div v-else>
                <span class="text-text-secondary">实际归还：</span>
                <span class="pixel-font">{{ record.actualReturnDate }}</span>
              </div>
              <div v-if="record.status === 'returned'">
                <span class="text-text-secondary">借用时长：</span>
                <span class="pixel-font text-pixel-green">
                  {{ calcBorrowDuration(record.borrowDate, record.actualReturnDate || '') }} 天
                </span>
              </div>
            </div>
            <div v-if="record.notes" class="text-sm text-text-secondary">
              <span>备注：{{ record.notes }}</span>
            </div>
            <div v-if="activeTab === 'borrowed'" class="flex gap-2 pt-2">
              <button class="pixel-btn pixel-btn-success !py-2 !px-3 !text-xs" @click="openReturnDialog(record)">
                ✅ 登记归还
              </button>
              <button class="pixel-btn !py-2 !px-3 !text-xs" @click="openExtendDialog(record)">
                ⏰ 延长借期
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div v-if="showBorrowDialog" class="fixed inset-0 bg-black/70 flex items-center justify-center z-50" @click.self="showBorrowDialog = false">
      <div class="pixel-card p-6 max-w-lg w-full mx-4 max-h-[90vh] overflow-y-auto">
        <h3 class="text-bright-yellow mb-4">登记借出</h3>
        <div class="space-y-4">
          <div>
            <label class="block text-sm text-text-secondary mb-1 pixel-font">选择卡带</label>
            <select v-model="borrowForm.cartridgeId" class="pixel-input w-full">
              <option :value="0">请选择卡带</option>
              <option v-for="c in availableCartridges" :key="c.id" :value="c.id">{{ c.title }}</option>
            </select>
          </div>
          <div>
            <label class="block text-sm text-text-secondary mb-1 pixel-font">借用人姓名</label>
            <input v-model="borrowForm.borrowerName" type="text" class="pixel-input w-full" placeholder="请输入姓名" />
          </div>
          <div>
            <label class="block text-sm text-text-secondary mb-1 pixel-font">联系方式</label>
            <input v-model="borrowForm.borrowerContact" type="text" class="pixel-input w-full" placeholder="电话或微信" />
          </div>
          <div class="grid grid-cols-2 gap-4">
            <div>
              <label class="block text-sm text-text-secondary mb-1 pixel-font">借出日期</label>
              <input v-model="borrowForm.borrowDate" type="date" class="pixel-input w-full" />
            </div>
            <div>
              <label class="block text-sm text-text-secondary mb-1 pixel-font">预计归还</label>
              <input v-model="borrowForm.expectedReturnDate" type="date" class="pixel-input w-full" />
            </div>
          </div>
          <div>
            <label class="block text-sm text-text-secondary mb-1 pixel-font">备注</label>
            <textarea v-model="borrowForm.notes" class="pixel-input w-full" rows="2" placeholder="可选"></textarea>
          </div>
        </div>
        <div class="flex gap-4 justify-end mt-6">
          <button class="pixel-btn" @click="showBorrowDialog = false">取消</button>
          <button class="pixel-btn pixel-btn-primary" @click="handleBorrow">确认借出</button>
        </div>
      </div>
    </div>

    <div v-if="showExtendDialog" class="fixed inset-0 bg-black/70 flex items-center justify-center z-50" @click.self="showExtendDialog = false">
      <div class="pixel-card p-6 max-w-md w-full mx-4">
        <h3 class="text-bright-yellow mb-4">延长借期</h3>
        <p class="text-text-secondary mb-4">
          卡带：{{ selectedRecord?.cartridge?.title }}<br>
          当前到期日：{{ selectedRecord?.expectedReturnDate }}
        </p>
        <div>
          <label class="block text-sm text-text-secondary mb-1 pixel-font">新的预计归还日期</label>
          <input v-model="extendForm.expectedReturnDate" type="date" class="pixel-input w-full" />
        </div>
        <div class="flex gap-4 justify-end mt-6">
          <button class="pixel-btn" @click="showExtendDialog = false">取消</button>
          <button class="pixel-btn pixel-btn-primary" @click="handleExtend">确认延长</button>
        </div>
      </div>
    </div>

    <div v-if="showReturnDialog" class="fixed inset-0 bg-black/70 flex items-center justify-center z-50" @click.self="showReturnDialog = false">
      <div class="pixel-card p-6 max-w-md w-full mx-4">
        <h3 class="text-bright-yellow mb-4">登记归还</h3>
        <p class="text-text-secondary mb-4">
          卡带：{{ selectedRecord?.cartridge?.title }}<br>
          借用人：{{ selectedRecord?.borrowerName }}
        </p>
        <div>
          <label class="block text-sm text-text-secondary mb-1 pixel-font">实际归还日期</label>
          <input v-model="returnForm.actualReturnDate" type="date" class="pixel-input w-full" />
        </div>
        <div class="flex gap-4 justify-end mt-6">
          <button class="pixel-btn" @click="showReturnDialog = false">取消</button>
          <button class="pixel-btn pixel-btn-success" @click="handleReturn">确认归还</button>
        </div>
      </div>
    </div>
  </div>
</template>

<style>
.pixel-blink {
  animation: blink 1s infinite;
}
@keyframes blink {
  0%, 50% { opacity: 1; }
  51%, 100% { opacity: 0.5; }
}
</style>
