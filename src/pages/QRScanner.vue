<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { Message } from '@arco-design/web-vue'
import { parseQRContent, generateQRDataURL } from '@/lib/qrcode'
import { cartridgeApi } from '@/api'
import { ConditionLabels } from '@/types'
import type { Cartridge } from '@/types'
import jsQR from 'jsqr'

const router = useRouter()

const videoRef = ref<HTMLVideoElement | null>(null)
const canvasRef = ref<HTMLCanvasElement | null>(null)
const stream = ref<MediaStream | null>(null)
const scanning = ref(false)
const scanError = ref('')
const lastResult = ref<{ valid: boolean; cartridgeId: number | null }>({
  valid: false,
  cartridgeId: null
})
const foundCartridge = ref<Cartridge | null>(null)
const facingMode = ref<'user' | 'environment'>('environment')
const manualInput = ref('')
const showingManual = ref(false)

let animationFrameId: number | null = null
let lastScanTime = 0

const scanResultText = computed(() => {
  if (!lastResult.value.cartridgeId) return ''
  if (lastResult.value.valid) {
    return `✓ 识别成功！卡带 ID: ${lastResult.value.cartridgeId}`
  }
  return `✗ 无法识别的二维码 (ID: ${lastResult.value.cartridgeId})`
})

const startScanner = async () => {
  try {
    scanError.value = ''
    const constraints: MediaStreamConstraints = {
      video: {
        facingMode: facingMode.value,
        width: { ideal: 1280 },
        height: { ideal: 720 }
      }
    }
    stream.value = await navigator.mediaDevices.getUserMedia(constraints)
    if (videoRef.value) {
      videoRef.value.srcObject = stream.value
      await videoRef.value.play()
    }
    scanning.value = true
    requestAnimationFrame(scanFrame)
  } catch (e: any) {
    console.error('摄像头启动失败:', e)
    scanError.value = e?.message || '无法访问摄像头，请确保已授权摄像头权限'
    scanning.value = false
  }
}

const stopScanner = () => {
  if (stream.value) {
    stream.value.getTracks().forEach(track => track.stop())
    stream.value = null
  }
  if (animationFrameId !== null) {
    cancelAnimationFrame(animationFrameId)
    animationFrameId = null
  }
  scanning.value = false
}

const scanFrame = () => {
  if (!scanning.value || !videoRef.value || !canvasRef.value) {
    animationFrameId = requestAnimationFrame(scanFrame)
    return
  }

  const now = Date.now()
  if (now - lastScanTime < 100) {
    animationFrameId = requestAnimationFrame(scanFrame)
    return
  }
  lastScanTime = now

  const video = videoRef.value
  const canvas = canvasRef.value
  const ctx = canvas.getContext('2d', { willReadFrequently: true })

  if (!ctx) {
    animationFrameId = requestAnimationFrame(scanFrame)
    return
  }

  if (video.readyState === video.HAVE_ENOUGH_DATA) {
    canvas.width = video.videoWidth
    canvas.height = video.videoHeight
    ctx.drawImage(video, 0, 0, canvas.width, canvas.height)

    const imageData = ctx.getImageData(0, 0, canvas.width, canvas.height)
    const code = jsQR(imageData.data, imageData.width, imageData.height)

    if (code) {
      const result = parseQRContent(code.data)
      if (result.valid && result.cartridgeId) {
        if (lastResult.value.cartridgeId !== result.cartridgeId) {
          lastResult.value = result
          loadCartridge(result.cartridgeId)
        }
      }
    }
  }

  animationFrameId = requestAnimationFrame(scanFrame)
}

const loadCartridge = async (id: number) => {
  try {
    const res = await cartridgeApi.getById(id)
    foundCartridge.value = res.data || null
  } catch {
    foundCartridge.value = null
  }
}

const switchCamera = async () => {
  stopScanner()
  facingMode.value = facingMode.value === 'user' ? 'environment' : 'user'
  await startScanner()
}

const goToDetail = () => {
  if (lastResult.value.cartridgeId) {
    router.push(`/cartridges/${lastResult.value.cartridgeId}`)
  }
}

const handleManualScan = () => {
  if (!manualInput.value.trim()) {
    Message.warning('请输入二维码内容或卡带ID')
    return
  }
  const input = manualInput.value.trim()
  let result = parseQRContent(input)
  if (!result.valid && /^\d+$/.test(input)) {
    const id = parseInt(input, 10)
    result = { valid: true, cartridgeId: id }
  }
  if (result.valid && result.cartridgeId) {
    lastResult.value = result
    loadCartridge(result.cartridgeId)
    Message.success('识别成功')
  } else {
    Message.error('无法识别的内容')
    lastResult.value = { valid: false, cartridgeId: null }
    foundCartridge.value = null
  }
}

const resetScan = () => {
  lastResult.value = { valid: false, cartridgeId: null }
  foundCartridge.value = null
}

onMounted(() => {
  // 不自动启动，需要用户点击开始
})

onUnmounted(() => {
  stopScanner()
})
</script>

<template>
  <div class="p-6 space-y-6">
    <div class="flex flex-col md:flex-row md:items-center md:justify-between gap-4">
      <div>
        <h1 class="text-neon-blue glow-blue">二维码扫描</h1>
        <p class="text-text-secondary">
          扫描卡带二维码快速定位到对应卡带详情页
        </p>
      </div>
      <div class="flex gap-3">
        <button
          class="pixel-btn"
          @click="showingManual = !showingManual"
        >
          ⌨️ 手动输入
        </button>
        <button
          v-if="scanning"
          class="pixel-btn pixel-btn-danger"
          @click="stopScanner"
        >
          ⏹️ 停止扫描
        </button>
        <button
          v-else
          class="pixel-btn pixel-btn-primary"
          @click="startScanner"
        >
          📷 开始扫描
        </button>
      </div>
    </div>

    <div v-if="showingManual" class="pixel-card p-4">
      <h3 class="text-bright-yellow mb-3">手动输入</h3>
      <p class="text-text-secondary text-sm mb-3">
        如果摄像头无法使用，可以手动输入二维码内容或卡带ID
      </p>
      <div class="flex gap-3">
        <input
          v-model="manualInput"
          type="text"
          class="pixel-input flex-1"
          placeholder="输入二维码内容或卡带ID..."
          @keyup.enter="handleManualScan"
        />
        <button class="pixel-btn pixel-btn-primary" @click="handleManualScan">
          🔍 查询
        </button>
      </div>
    </div>

    <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
      <div class="pixel-card p-4">
        <h3 class="text-bright-yellow mb-4">📷 扫描区域</h3>

        <div
          class="relative bg-dark-bg-3 border-4 border-neon-blue"
          style="aspect-ratio: 4/3; overflow: hidden;"
        >
          <video
            ref="videoRef"
            class="absolute inset-0 w-full h-full object-cover"
            playsinline
            muted
          />
          <canvas ref="canvasRef" class="hidden" />

          <div
            v-if="!scanning"
            class="absolute inset-0 flex flex-col items-center justify-center text-text-secondary"
          >
            <div class="text-6xl mb-4">📷</div>
            <p v-if="scanError" class="text-pixel-red text-center px-4 mb-4">
              {{ scanError }}
            </p>
            <p v-else>点击"开始扫描"启动摄像头</p>
            <button
              v-if="scanError"
              class="pixel-btn pixel-btn-primary mt-4"
              @click="startScanner"
            >
              🔄 重试
            </button>
          </div>

          <div
            v-if="scanning"
            class="absolute inset-0 pointer-events-none"
          >
            <div class="absolute inset-0 flex items-center justify-center">
              <div
                class="w-2/3 h-2/3 border-4 border-bright-yellow/70 relative"
                style="box-shadow: 0 0 30px rgba(255, 217, 61, 0.3);"
              >
                <div class="absolute top-0 left-0 w-8 h-8 border-t-4 border-l-4 border-bright-yellow"></div>
                <div class="absolute top-0 right-0 w-8 h-8 border-t-4 border-r-4 border-bright-yellow"></div>
                <div class="absolute bottom-0 left-0 w-8 h-8 border-b-4 border-l-4 border-bright-yellow"></div>
                <div class="absolute bottom-0 right-0 w-8 h-8 border-b-4 border-r-4 border-bright-yellow"></div>
              </div>
            </div>
            <div class="absolute top-4 left-1/2 -translate-x-1/2">
              <span class="pixel-badge pixel-badge-success animate-pulse">
                扫描中...
              </span>
            </div>
          </div>
        </div>

        <div class="mt-4 flex justify-center gap-3">
          <button
            v-if="scanning"
            class="pixel-btn !text-xs"
            @click="switchCamera"
          >
            🔄 切换摄像头
          </button>
          <button
            v-if="lastResult.cartridgeId"
            class="pixel-btn"
            @click="resetScan"
          >
            🔄 重新扫描
          </button>
        </div>
      </div>

      <div class="pixel-card p-4">
        <h3 class="text-bright-yellow mb-4">📋 扫描结果</h3>

        <div v-if="!lastResult.cartridgeId" class="text-center py-12 text-text-secondary">
          <div class="text-5xl mb-4">🔍</div>
          <p>将二维码对准扫描框</p>
          <p class="text-sm mt-2">支持卡带档案系统生成的二维码标签</p>
        </div>

        <div v-else class="space-y-4">
          <div
            class="p-4 pixel-border"
            :class="lastResult.valid ? '!border-pixel-green' : '!border-pixel-red'"
          >
            <div class="flex items-center gap-2 mb-2">
              <span class="text-2xl">{{ lastResult.valid ? '✅' : '❌' }}</span>
              <span
                class="pixel-font text-sm"
                :class="lastResult.valid ? 'text-pixel-green' : 'text-pixel-red'"
              >
                {{ lastResult.valid ? '有效二维码' : '无效二维码' }}
              </span>
            </div>
            <p v-if="scanResultText" class="text-text-secondary text-sm">
              {{ scanResultText }}
            </p>
          </div>

          <div v-if="foundCartridge" class="space-y-3">
            <div class="flex gap-4">
              <div class="w-24 h-32 shrink-0">
                <div class="cartridge-case !w-full !h-full">
                  <div class="cartridge-label">
                    <img
                      v-if="foundCartridge.coverImage"
                      :src="foundCartridge.coverImage"
                      :alt="foundCartridge.title"
                    />
                    <div v-else class="flex flex-col items-center justify-center h-full">
                      <div class="text-[10px] pixel-font text-bright-yellow leading-relaxed text-center px-1">
                        {{ foundCartridge.title }}
                      </div>
                    </div>
                  </div>
                </div>
              </div>
              <div class="flex-1 space-y-2">
                <h4 class="pixel-font text-bright-yellow text-sm">
                  {{ foundCartridge.title }}
                </h4>
                <div class="flex flex-wrap gap-2">
                  <span class="pixel-badge !text-[8px]">
                    {{ foundCartridge.platform }}
                  </span>
                  <span class="pixel-badge !text-[8px]">
                    {{ ConditionLabels[foundCartridge.condition] }}
                  </span>
                </div>
                <p class="text-text-secondary text-sm">
                  发行商: {{ foundCartridge.publisher }}
                </p>
                <p class="text-text-secondary text-sm">
                  年份: {{ foundCartridge.releaseYear }}
                </p>
              </div>
            </div>

            <div class="flex gap-3 pt-2">
              <button
                class="pixel-btn pixel-btn-primary flex-1"
                @click="goToDetail"
              >
                📄 查看详情
              </button>
            </div>
          </div>

          <div v-else-if="lastResult.cartridgeId" class="text-center py-6 text-text-secondary">
            <div class="animate-pulse">⏳ 正在加载卡带信息...</div>
          </div>
        </div>
      </div>
    </div>

    <div class="pixel-card p-4">
      <h3 class="text-bright-yellow mb-3">💡 使用提示</h3>
      <ul class="space-y-2 text-text-secondary text-sm">
        <li>• 将二维码对准扫描框，保持光线充足</li>
        <li>• 支持扫描卡带档案系统生成的所有二维码标签</li>
        <li>• 移动端建议使用后置摄像头（自动启用）</li>
        <li>• 如果无法识别，可以使用「手动输入」功能</li>
        <li>• 扫描成功后点击「查看详情」跳转到卡带页面</li>
      </ul>
    </div>
  </div>
</template>

<style scoped>
@keyframes scan-line {
  0% { top: 0; }
  100% { top: 100%; }
}
</style>
