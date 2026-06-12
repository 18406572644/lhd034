<script setup lang="ts">
import { ref, watch, onMounted, computed } from 'vue'
import { generateQRDataURL, generateQRContent } from '@/lib/qrcode'
import { ConditionLabels } from '@/types'
import type { Cartridge } from '@/types'

const props = defineProps<{
  cartridge: Cartridge
  size?: 'small' | 'medium' | 'large'
  showTitle?: boolean
  forPrint?: boolean
}>()

const emit = defineEmits<{
  (e: 'print'): void
}>()

const qrDataURL = ref('')
const qrSize = computed(() => {
  switch (props.size || 'medium') {
    case 'small': return 100
    case 'large': return 220
    default: return 150
  }
})

const loadQR = async () => {
  qrDataURL.value = await generateQRDataURL(props.cartridge.id, qrSize.value)
}

const handlePrint = () => {
  emit('print')
}

const qrRawContent = computed(() => generateQRContent(props.cartridge.id))

onMounted(loadQR)
watch(() => props.cartridge.id, loadQR)
</script>

<template>
  <div
    class="qr-label"
    :class="[
      `qr-label-${size || 'medium'}`,
      { 'qr-label-print': forPrint }
    ]"
  >
    <div class="qr-label-inner">
      <div class="qr-code-wrapper">
        <img
          v-if="qrDataURL"
          :src="qrDataURL"
          :alt="`卡带 ${cartridge.id} 二维码`"
          class="qr-code-img"
        />
      </div>

      <div v-if="showTitle !== false" class="qr-info">
        <div class="qr-title pixel-font" :title="cartridge.title">
          {{ cartridge.title }}
        </div>
        <div class="qr-meta">
          <span class="qr-platform">{{ cartridge.platform }}</span>
          <span class="qr-condition">{{ ConditionLabels[cartridge.condition] }}</span>
        </div>
        <div class="qr-id pixel-font">
          ID: {{ cartridge.id }}
        </div>
      </div>
    </div>

    <div v-if="!forPrint" class="qr-actions">
      <button class="pixel-btn !text-[10px] !px-3 !py-2" @click="handlePrint">
        🖨️ 打印标签
      </button>
    </div>

    <div class="qr-raw" v-if="forPrint">
      {{ qrRawContent }}
    </div>
  </div>
</template>

<style scoped>
.qr-label {
  display: flex;
  flex-direction: column;
  align-items: center;
  background: linear-gradient(145deg, var(--dark-bg-2) 0%, var(--dark-bg) 100%);
  border: 3px solid var(--neon-blue);
  padding: 12px;
  gap: 8px;
}

.qr-label-print {
  background: white !important;
  border: 2px solid #333 !important;
  color: #111 !important;
  page-break-inside: avoid;
}

.qr-label-small {
  max-width: 160px;
}

.qr-label-medium {
  max-width: 220px;
}

.qr-label-large {
  max-width: 300px;
}

.qr-label-inner {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  width: 100%;
}

.qr-code-wrapper {
  padding: 6px;
  background: white;
  border: 2px solid var(--neon-blue);
}

.qr-label-print .qr-code-wrapper {
  border: 2px solid #333 !important;
}

.qr-code-img {
  display: block;
  image-rendering: pixelated;
}

.qr-info {
  width: 100%;
  text-align: center;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.qr-title {
  font-size: 10px;
  color: var(--bright-yellow);
  line-height: 1.4;
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  word-break: break-all;
}

.qr-label-print .qr-title {
  color: #111 !important;
}

.qr-meta {
  display: flex;
  justify-content: center;
  gap: 6px;
  flex-wrap: wrap;
}

.qr-platform,
.qr-condition {
  font-family: var(--font-pixel);
  font-size: 8px;
  padding: 2px 6px;
  border: 1px solid var(--neon-blue);
  color: var(--neon-blue);
  background: var(--dark-bg-3);
}

.qr-label-print .qr-platform,
.qr-label-print .qr-condition {
  border-color: #333 !important;
  color: #333 !important;
  background: #f5f5f5 !important;
}

.qr-id {
  font-size: 8px;
  color: var(--text-secondary);
  letter-spacing: 1px;
}

.qr-label-print .qr-id {
  color: #666 !important;
}

.qr-actions {
  margin-top: 4px;
}

.qr-raw {
  font-family: monospace;
  font-size: 7px;
  color: var(--text-secondary);
  word-break: break-all;
  text-align: center;
  margin-top: 4px;
  opacity: 0.7;
}

.qr-label-print .qr-raw {
  color: #999 !important;
}

@media print {
  .qr-actions {
    display: none !important;
  }
}
</style>
