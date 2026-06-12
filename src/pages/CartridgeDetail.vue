<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { Message, Modal, Tabs, TabPane } from '@arco-design/web-vue'
import { cartridgeApi, borrowApi, wishlistApi } from '@/api'
import { ConditionLabels, StatusLabels } from '@/types'
import type { Cartridge, Playthrough, Review, BorrowRecord } from '@/types'

const route = useRoute()
const router = useRouter()

const cartridge = ref<Cartridge | null>(null)
const playthroughs = ref<Playthrough[]>([])
const review = ref<Review | null>(null)
const borrowRecords = ref<BorrowRecord[]>([])
const loading = ref(true)
const deleteConfirmVisible = ref(false)

const id = Number(route.params.id)

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

const statusBadgeClass = (status: string) => {
  const map: Record<string, string> = {
    borrowed: 'pixel-badge-warning',
    returned: 'pixel-badge-success',
    overdue: 'pixel-badge-danger'
  }
  return map[status] || ''
}

const renderStars = (rating: number) => {
  const stars = []
  for (let i = 1; i <= 5; i++) {
    stars.push(i <= rating ? '★' : '☆')
  }
  return stars.join('')
}

const formatDate = (dateStr: string | null) => {
  if (!dateStr) return '—'
  return dateStr.split('T')[0]
}

const loadData = async () => {
  try {
    loading.value = true
    const [cartRes, playRes, reviewRes, borrowRes] = await Promise.all([
      cartridgeApi.getById(id),
      cartridgeApi.getPlaythroughs(id),
      cartridgeApi.getReview(id),
      borrowApi.getList({ cartridgeId: id } as any)
    ])
    cartridge.value = cartRes.data
    playthroughs.value = playRes.data
    review.value = reviewRes.data
    borrowRecords.value = borrowRes.data
  } catch (error) {
    Message.error('加载数据失败')
  } finally {
    loading.value = false
  }
}

const handleDelete = () => {
  deleteConfirmVisible.value = true
}

const confirmDelete = async () => {
  try {
    await cartridgeApi.delete(id)
    Message.success('删除成功')
    router.push('/cartridges')
  } catch (error) {
    Message.error('删除失败')
  }
}

const handleAddPlaythrough = () => {
  router.push({ path: '/playthroughs/new', query: { cartridgeId: String(id) } })
}

const handleAddWishlist = async () => {
  try {
    await wishlistApi.create({ cartridgeId: id, priority: 'medium' })
    Message.success('已添加到待玩列表')
  } catch (error) {
    Message.error('添加失败')
  }
}

const handleAddBorrow = () => {
  router.push({ path: '/borrows/new', query: { cartridgeId: String(id) } })
}

onMounted(loadData)
</script>

<template>
  <div class="p-6 min-h-screen">
    <div v-if="loading" class="flex items-center justify-center py-20">
      <div class="pixel-font text-neon-blue text-xl animate-pulse">LOADING...</div>
    </div>

    <div v-else-if="!cartridge" class="text-center py-20">
      <div class="text-6xl mb-4">❓</div>
      <h3 class="text-bright-yellow mb-2">卡带不存在</h3>
      <button class="pixel-btn mt-4" @click="router.push('/cartridges')">返回列表</button>
    </div>

    <div v-else class="space-y-6">
      <div class="flex flex-wrap items-center justify-between gap-4">
        <button class="pixel-btn" @click="router.push('/cartridges')">◀ 返回</button>
        <div class="flex flex-wrap gap-3">
          <button class="pixel-btn" @click="router.push(`/cartridges/${id}/edit`)">✏️ 编辑</button>
          <button class="pixel-btn pixel-btn-danger" @click="handleDelete">🗑️ 删除</button>
          <button class="pixel-btn pixel-btn-success" @click="handleAddPlaythrough">🎮 通关记录</button>
          <button class="pixel-btn" @click="handleAddWishlist">📋 待玩</button>
          <button class="pixel-btn" @click="handleAddBorrow">🤝 借出</button>
        </div>
      </div>

      <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <div class="lg:col-span-1">
          <div class="cartridge-case !aspect-[3/4] !w-full max-w-md mx-auto">
            <div class="cartridge-label">
              <img v-if="cartridge.coverImage" :src="cartridge.coverImage" :alt="cartridge.title" />
              <div v-else class="flex flex-col items-center justify-center h-full">
                <div class="text-sm pixel-font text-bright-yellow leading-relaxed">{{ cartridge.title }}</div>
                <div class="text-xs text-neon-blue mt-2">{{ cartridge.platform }}</div>
              </div>
            </div>
          </div>
        </div>

        <div class="lg:col-span-2 space-y-6">
          <div class="pixel-card p-6">
            <h1 class="text-2xl text-bright-yellow glow-yellow mb-4 pixel-font">{{ cartridge.title }}</h1>
            <div class="flex flex-wrap gap-3 mb-6">
              <span class="pixel-badge">{{ cartridge.platform }}</span>
              <span class="pixel-badge" :class="conditionBadgeClass(cartridge.condition)">
                {{ ConditionLabels[cartridge.condition] }}
              </span>
            </div>

            <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
              <div class="flex items-start gap-2">
                <span class="pixel-badge !text-[10px] shrink-0">发行商</span>
                <span class="text-text-primary">{{ cartridge.publisher }}</span>
              </div>
              <div class="flex items-start gap-2">
                <span class="pixel-badge !text-[10px] shrink-0">发行年份</span>
                <span class="text-text-primary">{{ cartridge.releaseYear }}</span>
              </div>
              <div class="flex items-start gap-2">
                <span class="pixel-badge !text-[10px] shrink-0">购买价格</span>
                <span class="text-text-primary">¥{{ cartridge.purchasePrice }}</span>
              </div>
              <div class="flex items-start gap-2">
                <span class="pixel-badge !text-[10px] shrink-0">购买日期</span>
                <span class="text-text-primary">{{ formatDate(cartridge.purchaseDate) }}</span>
              </div>
              <div class="flex items-start gap-2">
                <span class="pixel-badge !text-[10px] shrink-0">地区</span>
                <span class="text-text-primary">{{ cartridge.region || '—' }}</span>
              </div>
            </div>

            <div v-if="cartridge.notes" class="mt-6">
              <span class="pixel-badge !text-[10px]">备注</span>
              <p class="mt-2 text-text-secondary">{{ cartridge.notes }}</p>
            </div>
          </div>

          <div class="pixel-card p-4">
            <Tabs type="card">
              <TabPane title="通关记录" key="playthroughs">
                <div v-if="playthroughs.length === 0" class="text-center py-8 text-text-secondary">
                  暂无通关记录
                </div>
                <div v-else class="space-y-4 max-h-96 overflow-y-auto scroll-hidden">
                  <div v-for="p in playthroughs" :key="p.id" class="pixel-border p-4">
                    <div class="flex justify-between items-center mb-2">
                      <span class="pixel-font text-neon-blue">第 {{ p.id }} 次通关</span>
                      <span class="text-text-secondary text-sm">{{ formatDate(p.completionDate) }}</span>
                    </div>
                    <div class="grid grid-cols-2 gap-2 text-sm">
                      <div>开始: {{ formatDate(p.startDate) }}</div>
                      <div>游戏时长: {{ p.playTimeHours }}h</div>
                      <div>难度: {{ renderStars(p.difficultyRating) }}</div>
                      <div>结局: {{ p.endingType }}</div>
                    </div>
                    <div v-if="p.notes" class="mt-2 text-text-secondary text-sm">{{ p.notes }}</div>
                  </div>
                </div>
              </TabPane>

              <TabPane title="游戏评价" key="review">
                <div v-if="!review" class="text-center py-8 text-text-secondary">
                  暂无评价
                </div>
                <div v-else class="space-y-4">
                  <div class="grid grid-cols-2 sm:grid-cols-4 gap-4">
                    <div class="text-center">
                      <div class="text-text-secondary text-sm mb-1">内容</div>
                      <div class="text-bright-yellow text-xl">{{ renderStars(review.contentRating) }}</div>
                    </div>
                    <div class="text-center">
                      <div class="text-text-secondary text-sm mb-1">玩法</div>
                      <div class="text-bright-yellow text-xl">{{ renderStars(review.gameplayRating) }}</div>
                    </div>
                    <div class="text-center">
                      <div class="text-text-secondary text-sm mb-1">画面</div>
                      <div class="text-bright-yellow text-xl">{{ renderStars(review.graphicsRating) }}</div>
                    </div>
                    <div class="text-center">
                      <div class="text-text-secondary text-sm mb-1">音效</div>
                      <div class="text-bright-yellow text-xl">{{ renderStars(review.soundRating) }}</div>
                    </div>
                  </div>
                  <div class="text-center py-2">
                    <span class="pixel-badge pixel-badge-success !text-base !px-6 !py-2">
                      综合评分: {{ review.overallRating.toFixed(1) }}
                    </span>
                  </div>
                  <div v-if="review.reviewText" class="pixel-border p-4">
                    <span class="pixel-badge !text-[10px]">评价</span>
                    <p class="mt-2 text-text-secondary">{{ review.reviewText }}</p>
                  </div>
                  <div v-if="review.storyNotes" class="pixel-border p-4">
                    <span class="pixel-badge !text-[10px]">剧情备注</span>
                    <p class="mt-2 text-text-secondary">{{ review.storyNotes }}</p>
                  </div>
                  <div v-if="review.easterEggs && review.easterEggs.length > 0" class="pixel-border p-4">
                    <span class="pixel-badge !text-[10px]">彩蛋记录</span>
                    <ul class="mt-2 space-y-1">
                      <li v-for="(egg, i) in review.easterEggs" :key="i" class="text-text-secondary">
                        <span class="text-pixel-pink mr-2">▸</span>{{ egg }}
                      </li>
                    </ul>
                  </div>
                </div>
              </TabPane>

              <TabPane title="借还记录" key="borrows">
                <div v-if="borrowRecords.length === 0" class="text-center py-8 text-text-secondary">
                  暂无借还记录
                </div>
                <div v-else class="space-y-4 max-h-96 overflow-y-auto scroll-hidden">
                  <div v-for="b in borrowRecords" :key="b.id" class="pixel-border p-4">
                    <div class="flex justify-between items-center mb-2">
                      <span class="pixel-font text-neon-blue">{{ b.borrowerName }}</span>
                      <span class="pixel-badge" :class="statusBadgeClass(b.status)">
                        {{ StatusLabels[b.status] }}
                      </span>
                    </div>
                    <div class="grid grid-cols-2 gap-2 text-sm">
                      <div>借出: {{ formatDate(b.borrowDate) }}</div>
                      <div>应还: {{ formatDate(b.expectedReturnDate) }}</div>
                      <div v-if="b.actualReturnDate">实还: {{ formatDate(b.actualReturnDate) }}</div>
                      <div v-if="b.borrowerContact">联系: {{ b.borrowerContact }}</div>
                    </div>
                    <div v-if="b.notes" class="mt-2 text-text-secondary text-sm">{{ b.notes }}</div>
                  </div>
                </div>
              </TabPane>
            </Tabs>
          </div>
        </div>
      </div>
    </div>

    <Modal
      v-model:visible="deleteConfirmVisible"
      title="确认删除"
      @ok="confirmDelete"
      okText="确认删除"
      cancelText="取消"
      okType="danger"
    >
      <p>确定要删除这张卡带吗？此操作无法撤销。</p>
    </Modal>
  </div>
</template>

<style scoped>
:deep(.arco-tabs-header) {
  border-bottom: 3px solid var(--neon-blue) !important;
}

:deep(.arco-tabs-tab) {
  font-family: var(--font-pixel) !important;
  font-size: 12px !important;
  color: var(--text-secondary) !important;
  padding: 12px 20px !important;
  border: none !important;
  margin-right: 4px !important;
  background: var(--dark-bg-2) !important;
}

:deep(.arco-tabs-tab:hover) {
  color: var(--neon-blue) !important;
}

:deep(.arco-tabs-tab-active) {
  color: var(--bright-yellow) !important;
  background: var(--dark-bg-3) !important;
  border-top: 3px solid var(--bright-yellow) !important;
  margin-top: -3px !important;
}

:deep(.arco-tabs-content) {
  padding-top: 16px !important;
}

:deep(.arco-modal) {
  background: var(--dark-bg-2) !important;
  border: 4px solid var(--neon-blue) !important;
  box-shadow: 0 0 30px rgba(0, 240, 255, 0.5) !important;
}

:deep(.arco-modal-header) {
  border-bottom: 2px solid var(--neon-blue) !important;
  font-family: var(--font-pixel) !important;
  color: var(--bright-yellow) !important;
}

:deep(.arco-modal-title) {
  font-family: var(--font-pixel) !important;
  color: var(--bright-yellow) !important;
}

:deep(.arco-modal-content) {
  color: var(--text-primary) !important;
}

:deep(.arco-btn-primary) {
  font-family: var(--font-pixel) !important;
  background: var(--neon-blue-dark) !important;
  border: 2px solid var(--neon-blue) !important;
  color: var(--dark-bg) !important;
}

:deep(.arco-btn-primary:hover) {
  background: var(--neon-blue) !important;
}

:deep(.arco-btn-status-danger) {
  font-family: var(--font-pixel) !important;
  background: var(--pixel-red) !important;
  border: 2px solid var(--pixel-pink) !important;
  color: white !important;
}

:deep(.arco-btn-default) {
  font-family: var(--font-pixel) !important;
  background: var(--dark-bg-3) !important;
  border: 2px solid var(--neon-blue) !important;
  color: var(--text-primary) !important;
}
</style>
