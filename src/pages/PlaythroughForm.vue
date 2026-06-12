<script setup lang="ts">
import { ref, reactive, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { DatePicker, Message } from '@arco-design/web-vue'
import { cartridgeApi, playthroughApi } from '@/api'
import { EndingTypeOptions } from '@/types'
import type { Cartridge, Playthrough } from '@/types'

const route = useRoute()
const router = useRouter()

const isEdit = route.name === 'playthrough-edit'
const id = isEdit ? Number(route.params.id) : 0

const loading = ref(false)
const cartridges = ref<Cartridge[]>([])
const hoverDifficulty = ref(0)

const form = reactive({
  cartridgeId: 0,
  startDate: '',
  completionDate: '',
  playTimeHours: 0,
  difficultyRating: 3 as Playthrough['difficultyRating'],
  endingType: '标准结局',
  multipleEndings: false,
  achievedEndings: [] as string[],
  notes: ''
})

const errors = reactive<Record<string, string>>({})
const newEnding = ref('')

const formTitle = computed(() => isEdit ? '编辑通关记录' : '添加通关记录')

const selectedCartridge = computed(() =>
  cartridges.value.find(c => c.id === form.cartridgeId)
)

const validate = () => {
  Object.keys(errors).forEach(key => delete errors[key])
  if (!form.cartridgeId) errors.cartridgeId = '请选择卡带'
  if (!form.completionDate) errors.completionDate = '请输入通关日期'
  if (form.playTimeHours < 0) errors.playTimeHours = '游玩时长不能为负数'
  return Object.keys(errors).length === 0
}

const loadCartridges = async () => {
  try {
    const res = await cartridgeApi.getList({ pageSize: 100 })
    cartridges.value = res.data.items
  } catch {
    Message.error('加载卡带列表失败')
  }
}

const loadData = async () => {
  try {
    loading.value = true
    const res = await playthroughApi.getById(id)
    const data = res.data
    Object.assign(form, {
      cartridgeId: data.cartridgeId,
      startDate: data.startDate ? data.startDate.split('T')[0] : '',
      completionDate: data.completionDate ? data.completionDate.split('T')[0] : '',
      playTimeHours: data.playTimeHours,
      difficultyRating: data.difficultyRating,
      endingType: data.endingType,
      multipleEndings: data.multipleEndings,
      achievedEndings: data.achievedEndings || [],
      notes: data.notes || ''
    })
  } catch (error) {
    Message.error('加载数据失败')
  } finally {
    loading.value = false
  }
}

const setDifficulty = (rating: number) => {
  form.difficultyRating = rating as Playthrough['difficultyRating']
}

const addEnding = () => {
  if (newEnding.value.trim() && !form.achievedEndings.includes(newEnding.value.trim())) {
    form.achievedEndings.push(newEnding.value.trim())
    newEnding.value = ''
  }
}

const removeEnding = (index: number) => {
  form.achievedEndings.splice(index, 1)
}

const handleSubmit = async () => {
  if (!validate()) {
    Message.warning('请检查表单信息')
    return
  }
  try {
    loading.value = true
    const data = { ...form }
    if (isEdit) {
      await playthroughApi.update(id, data)
      Message.success('更新成功')
    } else {
      await playthroughApi.create(data)
      Message.success('创建成功')
    }
    router.push('/playthroughs')
  } catch (error) {
    Message.error(isEdit ? '更新失败' : '创建失败')
  } finally {
    loading.value = false
  }
}

const handleCancel = () => {
  router.push('/playthroughs')
}

onMounted(async () => {
  await loadCartridges()
  if (isEdit) {
    await loadData()
  }
})
</script>

<template>
  <div class="p-6 max-w-4xl mx-auto">
    <div v-if="loading && isEdit" class="flex items-center justify-center py-20">
      <div class="pixel-font text-neon-blue text-xl animate-pulse">LOADING...</div>
    </div>

    <div v-else class="space-y-6">
      <div class="flex items-center justify-between">
        <h1 class="text-bright-yellow glow-yellow pixel-font">{{ formTitle }}</h1>
        <button class="pixel-btn" @click="handleCancel">◀ 返回</button>
      </div>

      <div class="pixel-card p-6">
        <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
          <div class="space-y-4">
            <div>
              <label class="block text-sm text-text-secondary mb-1 pixel-font">选择卡带 <span class="text-pixel-red">*</span></label>
              <select
                v-model.number="form.cartridgeId"
                class="pixel-input w-full"
                :class="{ '!border-pixel-red': errors.cartridgeId }"
              >
                <option :value="0">请选择卡带</option>
                <option v-for="c in cartridges" :key="c.id" :value="c.id">{{ c.title }}</option>
              </select>
              <p v-if="errors.cartridgeId" class="text-pixel-red text-sm mt-1">{{ errors.cartridgeId }}</p>
              <div v-if="selectedCartridge" class="mt-3 flex items-center gap-3 p-3 bg-dark-bg-2 border-2 border-neon-blue/50">
                <div class="w-16 h-16 flex-shrink-0 overflow-hidden border-2 border-bright-yellow">
                  <img v-if="selectedCartridge.coverImage" :src="selectedCartridge.coverImage" class="w-full h-full object-cover" />
                  <div v-else class="w-full h-full flex items-center justify-center text-2xl">🎮</div>
                </div>
                <div class="flex-1 min-w-0">
                  <div class="pixel-font text-sm text-bright-yellow truncate">{{ selectedCartridge.title }}</div>
                  <div class="text-xs text-text-secondary">{{ selectedCartridge.platform }} · {{ selectedCartridge.releaseYear }}</div>
                </div>
              </div>
            </div>

            <div>
              <label class="block text-sm text-text-secondary mb-1 pixel-font">开始日期</label>
              <DatePicker
                v-model="form.startDate"
                format="YYYY-MM-DD"
                class="w-full"
              />
            </div>

            <div>
              <label class="block text-sm text-text-secondary mb-1 pixel-font">通关日期 <span class="text-pixel-red">*</span></label>
              <DatePicker
                v-model="form.completionDate"
                format="YYYY-MM-DD"
                class="w-full"
                :class="{ '!border-pixel-red': errors.completionDate }"
              />
              <p v-if="errors.completionDate" class="text-pixel-red text-sm mt-1">{{ errors.completionDate }}</p>
            </div>

            <div>
              <label class="block text-sm text-text-secondary mb-1 pixel-font">游玩时长（小时）</label>
              <input
                v-model.number="form.playTimeHours"
                type="number"
                min="0"
                step="0.5"
                class="pixel-input w-full"
                :class="{ '!border-pixel-red': errors.playTimeHours }"
                placeholder="如: 15.5"
              />
              <p v-if="errors.playTimeHours" class="text-pixel-red text-sm mt-1">{{ errors.playTimeHours }}</p>
            </div>
          </div>

          <div class="space-y-4">
            <div>
              <label class="block text-sm text-text-secondary mb-1 pixel-font">难度评分</label>
              <div class="flex items-center gap-2">
                <button
                  v-for="i in 5"
                  :key="i"
                  type="button"
                  class="text-3xl transition-all hover:scale-125"
                  :class="i <= (hoverDifficulty || form.difficultyRating) ? 'text-pixel-orange' : 'text-text-secondary opacity-30'"
                  @mouseenter="hoverDifficulty = i"
                  @mouseleave="hoverDifficulty = 0"
                  @click="setDifficulty(i)"
                >★</button>
                <span class="ml-2 text-sm text-text-secondary">{{ form.difficultyRating }} 星</span>
              </div>
            </div>

            <div>
              <label class="block text-sm text-text-secondary mb-1 pixel-font">结局类型</label>
              <select v-model="form.endingType" class="pixel-input w-full">
                <option v-for="e in EndingTypeOptions" :key="e" :value="e">{{ e }}</option>
              </select>
            </div>

            <div>
              <label class="pixel-btn flex items-center gap-3 cursor-pointer !py-3">
                <input type="checkbox" v-model="form.multipleEndings" class="w-5 h-5" />
                <span>是否多结局</span>
              </label>
            </div>

            <div v-if="form.multipleEndings">
              <label class="block text-sm text-text-secondary mb-1 pixel-font">已达成结局</label>
              <div class="flex gap-2 mb-2">
                <input
                  v-model="newEnding"
                  type="text"
                  class="pixel-input flex-1"
                  placeholder="输入结局名称..."
                  @keyup.enter="addEnding"
                />
                <button type="button" class="pixel-btn pixel-btn-primary" @click="addEnding">添加</button>
              </div>
              <div class="flex flex-wrap gap-2">
                <span
                  v-for="(ending, index) in form.achievedEndings"
                  :key="index"
                  class="pixel-badge flex items-center gap-1"
                >
                  {{ ending }}
                  <button type="button" class="ml-1 hover:text-pixel-red" @click="removeEnding(index)">✕</button>
                </span>
                <span v-if="form.achievedEndings.length === 0" class="text-text-secondary text-sm">暂无达成结局</span>
              </div>
            </div>

            <div>
              <label class="block text-sm text-text-secondary mb-1 pixel-font">备注</label>
              <textarea
                v-model="form.notes"
                class="pixel-input w-full min-h-[100px] resize-y"
                placeholder="输入备注信息..."
              ></textarea>
            </div>
          </div>
        </div>

        <div class="flex justify-end gap-4 mt-8 pt-6 border-t-2 border-neon-blue/30">
          <button class="pixel-btn" @click="handleCancel" :disabled="loading">取消</button>
          <button class="pixel-btn pixel-btn-primary" @click="handleSubmit" :disabled="loading">
            {{ loading ? '保存中...' : '保存' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
:deep(.arco-picker) {
  background: var(--dark-bg-2);
  border: 3px solid var(--neon-blue);
  color: var(--text-primary);
  font-family: var(--font-body);
  font-size: 18px;
  padding: 10px 16px;
  height: auto;
}

:deep(.arco-picker:hover) {
  border-color: var(--bright-yellow);
}

:deep(.arco-picker-focus),
:deep(.arco-picker:focus-within) {
  border-color: var(--bright-yellow);
  box-shadow: 0 0 10px rgba(255, 217, 61, 0.3);
}

:deep(.arco-picker-input input) {
  background: transparent;
  color: var(--text-primary);
  font-family: var(--font-body);
  font-size: 18px;
}

:deep(.arco-picker-input input::placeholder) {
  color: var(--text-secondary);
}

:deep(.arco-picker-suffix svg) {
  color: var(--neon-blue);
}

:deep(.arco-checkbox-input) {
  background: var(--dark-bg-3);
  border: 2px solid var(--neon-blue);
}

:deep(.arco-checkbox-checked .arco-checkbox-input) {
  background: var(--neon-blue);
  border-color: var(--bright-yellow);
}
</style>
