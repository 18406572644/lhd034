<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { backupApi } from '@/api'
import type { BackupInfo, BackupConfig } from '@/types'

const backups = ref<BackupInfo[]>([])
const loading = ref(false)
const creating = ref(false)
const restoring = ref<string | null>(null)
const deleting = ref<string | null>(null)

const config = ref<BackupConfig>({
  enabled: false,
  frequency: 'daily',
  retention: 7
})

const configLoading = ref(false)
const configSaving = ref(false)

const formatSize = (bytes: number) => {
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  return `${(bytes / (1024 * 1024)).toFixed(2)} MB`
}

const formatDate = (dateStr: string) => {
  if (!dateStr) return '-'
  const date = new Date(dateStr)
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  })
}

const loadBackups = async () => {
  loading.value = true
  try {
    const res = await backupApi.getList()
    if (res.code === 0) {
      backups.value = res.data || []
    }
  } finally {
    loading.value = false
  }
}

const loadConfig = async () => {
  configLoading.value = true
  try {
    const res = await backupApi.getConfig()
    if (res.code === 0 && res.data) {
      config.value = res.data
    }
  } finally {
    configLoading.value = false
  }
}

const handleCreateBackup = async () => {
  creating.value = true
  try {
    const res = await backupApi.createBackup()
    if (res.code === 0) {
      await loadBackups()
    }
  } catch {
  } finally {
    creating.value = false
  }
}

const handleRestore = async (filename: string) => {
  if (!confirm(`确定要从备份 "${filename}" 恢复数据库吗？\n\n恢复前将自动创建当前数据库的快照。`)) {
    return
  }
  restoring.value = filename
  try {
    const res = await backupApi.restoreBackup(filename)
    if (res.code === 0) {
      alert(`数据库恢复成功！\n\n已创建恢复前快照: ${res.data?.snapshotFilename || ''}`)
      await loadBackups()
    }
  } catch {
  } finally {
    restoring.value = null
  }
}

const handleDelete = async (filename: string) => {
  if (!confirm(`确定要删除备份 "${filename}" 吗？此操作不可撤销。`)) {
    return
  }
  deleting.value = filename
  try {
    const res = await backupApi.deleteBackup(filename)
    if (res.code === 0) {
      await loadBackups()
    }
  } catch {
  } finally {
    deleting.value = null
  }
}

const handleSaveConfig = async () => {
  configSaving.value = true
  try {
    const res = await backupApi.updateConfig(config.value)
    if (res.code === 0 && res.data) {
      config.value = res.data
    }
  } catch {
  } finally {
    configSaving.value = false
  }
}

const frequencyOptions = [
  { value: 'daily', label: '每日' },
  { value: 'weekly', label: '每周' },
  { value: 'monthly', label: '每月' }
]

const retentionOptions = [
  { value: 3, label: '保留 3 份' },
  { value: 7, label: '保留 7 份' },
  { value: 14, label: '保留 14 份' },
  { value: 30, label: '保留 30 份' },
  { value: 90, label: '保留 90 份' }
]

const isSnapshot = (filename: string) => {
  return typeof filename === 'string' && filename.startsWith('pre_restore_snapshot_')
}

onMounted(() => {
  loadBackups()
  loadConfig()
})
</script>

<template>
  <div class="space-y-8">
    <div class="pixel-card p-6 pixel-fade-in">
      <h2 class="text-bright-yellow glow-yellow mb-6">⚙️ 备份管理</h2>
      <p class="text-text-secondary mb-6">管理数据库备份，保障数据安全</p>
    </div>

    <div class="pixel-card p-6 pixel-fade-in" style="animation-delay: 0.1s">
      <h3 class="text-neon-blue glow-blue mb-4">📦 手动备份</h3>
      <p class="text-text-secondary mb-4">立即创建当前数据库的完整备份</p>
      <button
        class="pixel-btn pixel-btn-primary"
        :disabled="creating"
        @click="handleCreateBackup"
      >
        {{ creating ? '备份中...' : '+ 创建备份' }}
      </button>
    </div>

    <div class="pixel-card p-6 pixel-fade-in" style="animation-delay: 0.2s">
      <h3 class="text-neon-blue glow-blue mb-4">⏰ 自动备份设置</h3>
      <div v-if="configLoading" class="text-text-secondary">加载中...</div>
      <div v-else class="space-y-4">
        <div class="flex items-center justify-between">
          <div>
            <div class="text-text-primary font-bold">启用自动备份</div>
            <div class="text-text-secondary text-sm">按设定周期自动创建数据库备份</div>
          </div>
          <label class="relative inline-flex items-center cursor-pointer">
            <input
              type="checkbox"
              v-model="config.enabled"
              class="sr-only peer"
            />
            <div class="w-14 h-7 bg-dark-bg-3 border-2 border-neon-blue rounded-full peer peer-checked:bg-pixel-green transition-all after:content-[''] after:absolute after:top-1 after:left-1 after:bg-dark-bg after:border-2 after:border-neon-blue after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:after:translate-x-7"></div>
          </label>
        </div>

        <div v-if="config.enabled" class="space-y-4 pt-4 border-t-2 border-dark-bg-3">
          <div>
            <label class="block text-text-primary mb-2">备份周期</label>
            <select
              v-model="config.frequency"
              class="pixel-input w-full"
            >
              <option v-for="opt in frequencyOptions" :key="opt.value" :value="opt.value">
                {{ opt.label }}
              </option>
            </select>
          </div>

          <div>
            <label class="block text-text-primary mb-2">保留数量</label>
            <select
              v-model.number="config.retention"
              class="pixel-input w-full"
            >
              <option v-for="opt in retentionOptions" :key="opt.value" :value="opt.value">
                {{ opt.label }}
              </option>
            </select>
            <p class="text-text-secondary text-sm mt-1">
              超过保留数量的旧备份将自动清理
            </p>
          </div>

          <button
            class="pixel-btn pixel-btn-success"
            :disabled="configSaving"
            @click="handleSaveConfig"
          >
            {{ configSaving ? '保存中...' : '保存设置' }}
          </button>
        </div>
      </div>
    </div>

    <div class="pixel-card p-6 pixel-fade-in" style="animation-delay: 0.3s">
      <div class="flex justify-between items-center mb-4">
        <h3 class="text-neon-blue glow-blue">📋 备份列表</h3>
        <button
          class="pixel-btn !py-2 !px-4 text-xs"
          @click="loadBackups"
        >
          🔄 刷新
        </button>
      </div>

      <div v-if="loading" class="text-center py-8 text-text-secondary">
        加载中...
      </div>

      <div v-else-if="!backups || backups.length === 0" class="text-center py-8 text-text-secondary">
        暂无备份记录
      </div>

      <div v-else class="space-y-3">
        <div
          v-for="backup in (backups || [])"
          :key="backup.filename"
          class="p-4 bg-dark-bg-2 border-2 border-neon-blue/30 hover:border-neon-blue transition-all"
        >
          <div class="flex justify-between items-start">
            <div class="flex-1">
              <div class="flex items-center gap-2 mb-1">
                <span class="text-text-primary font-bold">{{ backup.filename }}</span>
                <span
                  v-if="isSnapshot(backup.filename)"
                  class="pixel-badge pixel-badge-warning text-xs"
                >
                  恢复快照
                </span>
              </div>
              <div class="text-text-secondary text-sm space-x-4">
                <span>📅 {{ formatDate(backup.createdAt) }}</span>
                <span>💾 {{ formatSize(backup.size) }}</span>
              </div>
            </div>
            <div class="flex gap-2">
              <button
                class="pixel-btn !py-2 !px-3 text-xs pixel-btn-success"
                :disabled="restoring === backup.filename"
                @click="handleRestore(backup.filename)"
              >
                {{ restoring === backup.filename ? '恢复中...' : '恢复' }}
              </button>
              <button
                class="pixel-btn !py-2 !px-3 text-xs pixel-btn-danger"
                :disabled="deleting === backup.filename"
                @click="handleDelete(backup.filename)"
              >
                {{ deleting === backup.filename ? '删除中...' : '删除' }}
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.sr-only {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  white-space: nowrap;
  border-width: 0;
}

select {
  appearance: none;
  background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' fill='none' viewBox='0 0 20 20'%3E%3Cpath stroke='%2300F0FF' stroke-linecap='round' stroke-linejoin='round' stroke-width='2' d='M6 8l4 4 4-4'/%3E%3C/svg%3E");
  background-position: right 0.75rem center;
  background-repeat: no-repeat;
  background-size: 1.25rem;
  padding-right: 2.5rem;
}
</style>
