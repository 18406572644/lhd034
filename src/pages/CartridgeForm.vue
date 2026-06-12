<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { Message } from '@arco-design/web-vue'
import { cartridgeApi } from '@/api'
import { PlatformOptions, ConditionOptions, PlayStatusLabels } from '@/types'
import type { Cartridge } from '@/types'

const route = useRoute()
const router = useRouter()

const isEdit = route.name === 'cartridge-edit'
const id = isEdit ? Number(route.params.id) : 0

const loading = ref(false)
const uploading = ref(false)
const coverUploading = ref(false)

const form = reactive({
  title: '',
  platform: '',
  publisher: '',
  releaseYear: 0,
  condition: 'good' as Cartridge['condition'],
  status: 'unstarted' as Cartridge['status'],
  purchasePrice: 0,
  purchaseDate: '',
  region: '',
  coverImage: '',
  screenshots: [] as string[],
  notes: ''
})

const errors = reactive<Record<string, string>>({})

const publisherSuggestions = ref<string[]>([])
const coverPreview = ref('')
const screenshotPreviews = ref<string[]>([])

const formTitle = isEdit ? '编辑卡带' : '新增卡带'

const validate = () => {
  Object.keys(errors).forEach(key => delete errors[key])
  if (!form.title.trim()) errors.title = '请输入游戏标题'
  if (!form.platform) errors.platform = '请选择平台'
  if (form.releaseYear && (form.releaseYear < 1970 || form.releaseYear > new Date().getFullYear())) {
    errors.releaseYear = '请输入有效的年份'
  }
  if (form.purchasePrice < 0) errors.purchasePrice = '价格不能为负数'
  return Object.keys(errors).length === 0
}

const loadPublishers = async () => {
  try {
    const res = await cartridgeApi.getPublishers()
    publisherSuggestions.value = res.data
  } catch {
    // ignore
  }
}

const loadData = async () => {
  try {
    loading.value = true
    const res = await cartridgeApi.getById(id)
    const data = res.data
    Object.assign(form, {
      title: data.title,
      platform: data.platform,
      publisher: data.publisher,
      releaseYear: data.releaseYear,
      condition: data.condition,
      status: data.status || 'unstarted',
      purchasePrice: data.purchasePrice,
      purchaseDate: data.purchaseDate ? data.purchaseDate.split('T')[0] : '',
      region: data.region || '',
      coverImage: data.coverImage || '',
      screenshots: data.screenshots || [],
      notes: data.notes || ''
    })
    coverPreview.value = data.coverImage || ''
    screenshotPreviews.value = [...(data.screenshots || [])]
  } catch (error) {
    Message.error('加载数据失败')
  } finally {
    loading.value = false
  }
}

const handleCoverUpload = async (e: Event) => {
  const input = e.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return

  const reader = new FileReader()
  reader.onload = (ev) => {
    coverPreview.value = ev.target?.result as string
  }
  reader.readAsDataURL(file)

  try {
    coverUploading.value = true
    const res = await cartridgeApi.upload(file)
    form.coverImage = res.data.url
  } catch (error) {
    Message.error('封面上传失败')
    coverPreview.value = form.coverImage
  } finally {
    coverUploading.value = false
  }
  input.value = ''
}

const handleScreenshotUpload = async (e: Event) => {
  const input = e.target as HTMLInputElement
  const files = input.files
  if (!files || files.length === 0) return

  const uploadPromises = Array.from(files).map(async (file) => {
    const reader = new FileReader()
    reader.onload = (ev) => {
      screenshotPreviews.value.push(ev.target?.result as string)
    }
    reader.readAsDataURL(file)

    try {
      const res = await cartridgeApi.upload(file)
      return res.data.url
    } catch (error) {
      Message.error('截图上传失败')
      return null
    }
  })

  try {
    uploading.value = true
    const results = await Promise.all(uploadPromises)
    results.forEach(url => {
      if (url) form.screenshots.push(url)
    })
  } finally {
    uploading.value = false
  }
  input.value = ''
}

const removeScreenshot = (index: number) => {
  form.screenshots.splice(index, 1)
  screenshotPreviews.value.splice(index, 1)
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
      await cartridgeApi.update(id, data)
      Message.success('更新成功')
    } else {
      await cartridgeApi.create(data)
      Message.success('创建成功')
    }
    router.push('/cartridges')
  } catch (error) {
    Message.error(isEdit ? '更新失败' : '创建失败')
  } finally {
    loading.value = false
  }
}

const handleCancel = () => {
  router.push('/cartridges')
}

onMounted(async () => {
  await loadPublishers()
  if (isEdit) {
    await loadData()
  }
})
</script>

<template>
  <div class="p-6 max-w-6xl mx-auto">
    <div v-if="loading && isEdit" class="flex items-center justify-center py-20">
      <div class="pixel-font text-neon-blue text-xl animate-pulse">LOADING...</div>
    </div>

    <div v-else class="space-y-6">
      <div class="flex items-center justify-between">
        <h1 class="text-neon-blue glow-blue pixel-font">{{ formTitle }}</h1>
        <button class="pixel-btn" @click="handleCancel">◀ 返回</button>
      </div>

      <div class="pixel-card p-6">
        <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
          <div class="space-y-4">
            <div>
              <label class="block text-sm text-text-secondary mb-1 pixel-font">游戏标题 <span class="text-pixel-red">*</span></label>
              <input
                v-model="form.title"
                type="text"
                class="pixel-input w-full"
                :class="{ '!border-pixel-red': errors.title }"
                placeholder="输入游戏标题..."
              />
              <p v-if="errors.title" class="text-pixel-red text-sm mt-1">{{ errors.title }}</p>
            </div>

            <div>
              <label class="block text-sm text-text-secondary mb-1 pixel-font">平台 <span class="text-pixel-red">*</span></label>
              <select
                v-model="form.platform"
                class="pixel-input w-full"
                :class="{ '!border-pixel-red': errors.platform }"
              >
                <option value="">请选择平台</option>
                <option v-for="p in PlatformOptions" :key="p" :value="p">{{ p }}</option>
              </select>
              <p v-if="errors.platform" class="text-pixel-red text-sm mt-1">{{ errors.platform }}</p>
            </div>

            <div>
              <label class="block text-sm text-text-secondary mb-1 pixel-font">发行商</label>
              <input
                v-model="form.publisher"
                type="text"
                class="pixel-input w-full"
                placeholder="输入或选择发行商..."
                list="publisher-list"
              />
              <datalist id="publisher-list">
                <option v-for="p in publisherSuggestions" :key="p" :value="p">{{ p }}</option>
              </datalist>
            </div>

            <div class="grid grid-cols-2 gap-4">
              <div>
                <label class="block text-sm text-text-secondary mb-1 pixel-font">发行年份</label>
                <input
                  v-model.number="form.releaseYear"
                  type="number"
                  class="pixel-input w-full"
                  :class="{ '!border-pixel-red': errors.releaseYear }"
                  placeholder="如: 1998"
                />
                <p v-if="errors.releaseYear" class="text-pixel-red text-sm mt-1">{{ errors.releaseYear }}</p>
              </div>
              <div>
                <label class="block text-sm text-text-secondary mb-1 pixel-font">品相</label>
                <select v-model="form.condition" class="pixel-input w-full">
                  <option v-for="c in ConditionOptions" :key="c.value" :value="c.value">{{ c.label }}</option>
                </select>
              </div>
            </div>

            <div class="grid grid-cols-2 gap-4">
              <div>
                <label class="block text-sm text-text-secondary mb-1 pixel-font">购买价格</label>
                <input
                  v-model.number="form.purchasePrice"
                  type="number"
                  step="0.01"
                  class="pixel-input w-full"
                  :class="{ '!border-pixel-red': errors.purchasePrice }"
                  placeholder="¥0.00"
                />
                <p v-if="errors.purchasePrice" class="text-pixel-red text-sm mt-1">{{ errors.purchasePrice }}</p>
              </div>
              <div>
                <label class="block text-sm text-text-secondary mb-1 pixel-font">购买日期</label>
                <input
                  v-model="form.purchaseDate"
                  type="date"
                  class="pixel-input w-full"
                />
              </div>
            </div>

            <div>
              <label class="block text-sm text-text-secondary mb-1 pixel-font">游玩状态</label>
              <select v-model="form.status" class="pixel-input w-full">
                <option value="unstarted">未开始</option>
                <option value="playing">进行中</option>
                <option value="completed">已通关</option>
                <option value="shelved">搁置</option>
              </select>
            </div>

            <div>
              <label class="block text-sm text-text-secondary mb-1 pixel-font">地区</label>
              <input
                v-model="form.region"
                type="text"
                class="pixel-input w-full"
                placeholder="如: 日版、美版、欧版..."
              />
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

          <div class="space-y-6">
            <div>
              <label class="block text-sm text-text-secondary mb-2 pixel-font">封面图</label>
              <div class="pixel-border p-4 bg-dark-bg-2">
                <div class="cartridge-case !aspect-[3/4] !w-full max-w-xs mx-auto mb-4">
                  <div class="cartridge-label">
                    <img v-if="coverPreview" :src="coverPreview" alt="封面预览" />
                    <div v-else class="flex flex-col items-center justify-center h-full">
                      <div class="text-4xl mb-2">📷</div>
                      <div class="text-xs text-text-secondary">暂无封面</div>
                    </div>
                  </div>
                </div>
                <label class="pixel-btn w-full block text-center cursor-pointer">
                  <input
                    type="file"
                    accept="image/*"
                    class="hidden"
                    @change="handleCoverUpload"
                    :disabled="coverUploading"
                  />
                  {{ coverUploading ? '上传中...' : (form.coverImage ? '更换封面' : '上传封面') }}
                </label>
              </div>
            </div>

            <div>
              <label class="block text-sm text-text-secondary mb-2 pixel-font">游戏截图</label>
              <div class="pixel-border p-4 bg-dark-bg-2">
                <div v-if="screenshotPreviews.length > 0" class="grid grid-cols-3 gap-3 mb-4">
                  <div
                    v-for="(preview, index) in screenshotPreviews"
                    :key="index"
                    class="relative aspect-video bg-dark-bg-3 overflow-hidden"
                  >
                    <img :src="preview" class="w-full h-full object-cover" alt="截图预览" />
                    <button
                      class="absolute top-1 right-1 pixel-btn pixel-btn-danger !p-1 !text-xs"
                      @click="removeScreenshot(index)"
                    >
                      ✕
                    </button>
                  </div>
                </div>
                <div v-else class="text-center py-8 text-text-secondary mb-4">
                  <div class="text-3xl mb-2">🖼️</div>
                  <div class="text-sm">暂无截图</div>
                </div>
                <label class="pixel-btn w-full block text-center cursor-pointer">
                  <input
                    type="file"
                    accept="image/*"
                    multiple
                    class="hidden"
                    @change="handleScreenshotUpload"
                    :disabled="uploading"
                  />
                  {{ uploading ? '上传中...' : '添加截图' }}
                </label>
              </div>
            </div>
          </div>
        </div>

        <div class="flex justify-end gap-4 mt-8 pt-6 border-t-2 border-neon-blue/30">
          <button class="pixel-btn" @click="handleCancel" :disabled="loading">
            取消
          </button>
          <button class="pixel-btn pixel-btn-primary" @click="handleSubmit" :disabled="loading">
            {{ loading ? '保存中...' : '保存' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
:deep(input[type="date"]::-webkit-calendar-picker-indicator) {
  filter: invert(1) sepia(1) saturate(5) hue-rotate(170deg);
  cursor: pointer;
}

:deep(input[type="date"]) {
  color-scheme: dark;
}

:deep(input[type="number"]::-webkit-inner-spin-button),
:deep(input[type="number"]::-webkit-outer-spin-button) {
  opacity: 1;
}
</style>
