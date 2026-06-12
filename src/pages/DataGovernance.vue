<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { Message } from '@arco-design/web-vue'
import { dataQualityApi } from '@/api'
import type { DataQualityReport, DuplicateGroup, AnomalyItem } from '@/types'

const loading = ref(false)
const report = ref<DataQualityReport | null>(null)

const activeTab = ref<'overview' | 'duplicates' | 'missing' | 'anomalies'>('overview')

const scanLoading = ref(false)

const fixDuplicateDialog = ref(false)
const currentDuplicateGroup = ref<DuplicateGroup | null>(null)
const keepId = ref<number>(0)

const fixMissingDialog = ref(false)
const fixMissingField = ref('')
const fixMissingValue = ref('')

const severityColors: Record<string, string> = {
  high: 'text-red-400',
  medium: 'text-yellow-400',
  low: 'text-green-400'
}

const severityLabels: Record<string, string> = {
  high: '高',
  medium: '中',
  low: '低'
}

const loadReport = async () => {
  loading.value = true
  try {
    const res = await dataQualityApi.getReport()
    if (res.code === 0) {
      report.value = res.data
    }
  } catch (e) {
    Message.error('加载数据质量报告失败')
  } finally {
    loading.value = false
  }
}

const rescan = async () => {
  scanLoading.value = true
  try {
    const res = await dataQualityApi.getReport()
    if (res.code === 0) {
      report.value = res.data
      Message.success('扫描完成')
    }
  } catch (e) {
    Message.error('扫描失败')
  } finally {
    scanLoading.value = false
  }
}

const openFixDuplicate = (group: DuplicateGroup) => {
  currentDuplicateGroup.value = group
  keepId.value = group.cartridges[0]?.id || 0
  fixDuplicateDialog.value = true
}

const executeFixDuplicate = async () => {
  if (!keepId.value) {
    Message.warning('请选择要保留的卡带')
    return
  }
  
  try {
    const res = await dataQualityApi.fixDuplicates(keepId.value)
    if (res.code === 0) {
      Message.success(`成功合并，删除 ${(res.data as any).deletedCount} 条重复记录`)
      fixDuplicateDialog.value = false
      loadReport()
    }
  } catch (e) {
    Message.error('修复失败')
  }
}

const openFixMissing = (field: string) => {
  fixMissingField.value = field
  fixMissingValue.value = ''
  fixMissingDialog.value = true
}

const executeFixMissing = async () => {
  if (!fixMissingValue.value) {
    Message.warning('请输入填充值')
    return
  }
  
  try {
    const res = await dataQualityApi.fixMissingField(fixMissingField.value, fixMissingValue.value)
    if (res.code === 0) {
      Message.success(`成功修复 ${(res.data as any).updatedCount} 条记录`)
      fixMissingDialog.value = false
      loadReport()
    }
  } catch (e) {
    Message.error('修复失败')
  }
}

const getScoreColor = (score: number) => {
  if (score >= 80) return 'text-green-400'
  if (score >= 60) return 'text-yellow-400'
  return 'text-red-400'
}

const getScoreLabel = (score: number) => {
  if (score >= 90) return '优秀'
  if (score >= 80) return '良好'
  if (score >= 60) return '一般'
  return '较差'
}

onMounted(() => {
  loadReport()
})
</script>

<template>
  <div class="p-6 space-y-6">
    <div class="flex flex-col md:flex-row md:items-center md:justify-between gap-4">
      <div>
        <h1 class="text-neon-blue glow-blue">数据治理中心</h1>
        <p class="text-text-secondary">数据质量检测与修复</p>
      </div>
      <button 
        class="pixel-btn pixel-btn-primary" 
        @click="rescan"
        :disabled="scanLoading"
      >
        {{ scanLoading ? '扫描中...' : '🔄 重新扫描' }}
      </button>
    </div>

    <div v-if="loading" class="text-center py-16">
      <div class="text-4xl animate-pulse">🔍</div>
      <p class="text-text-secondary mt-4">正在分析数据质量...</p>
    </div>

    <div v-else-if="!report" class="text-center py-16">
      <div class="text-4xl mb-4">📊</div>
      <p class="text-text-secondary">暂无数据</p>
    </div>

    <div v-else class="space-y-6">
      <div class="grid grid-cols-1 md:grid-cols-4 gap-4">
        <div class="pixel-card p-4 text-center">
          <div class="text-3xl mb-2">📚</div>
          <div class="text-2xl pixel-font text-bright-yellow">{{ report.totalCartridges }}</div>
          <div class="text-sm text-text-secondary">总卡带数</div>
        </div>
        <div class="pixel-card p-4 text-center">
          <div class="text-3xl mb-2">✨</div>
          <div class="text-2xl pixel-font" :class="getScoreColor(report.completenessScore)">
            {{ report.completenessScore.toFixed(1) }}
          </div>
          <div class="text-sm text-text-secondary">
            完整度评分 ({{ getScoreLabel(report.completenessScore) }})
          </div>
        </div>
        <div class="pixel-card p-4 text-center">
          <div class="text-3xl mb-2">🔄</div>
          <div class="text-2xl pixel-font" :class="report.duplicateCount > 0 ? 'text-yellow-400' : 'text-green-400'">
            {{ report.duplicateCount }}
          </div>
          <div class="text-sm text-text-secondary">重复记录</div>
        </div>
        <div class="pixel-card p-4 text-center">
          <div class="text-3xl mb-2">⚠️</div>
          <div class="text-2xl pixel-font" :class="report.anomalyCount > 0 ? 'text-red-400' : 'text-green-400'">
            {{ report.anomalyCount }}
          </div>
          <div class="text-sm text-text-secondary">异常值</div>
        </div>
      </div>

      <div class="pixel-card p-1">
        <div class="flex border-b-4 border-dark-bg-3">
          <button
            v-for="tab in [
              { key: 'overview', label: '📊 总览' },
              { key: 'duplicates', label: '🔄 重复检测' },
              { key: 'missing', label: '📝 缺失字段' },
              { key: 'anomalies', label: '⚠️ 异常值' }
            ]"
            :key="tab.key"
            class="px-6 py-3 pixel-font text-sm transition-colors"
            :class="activeTab === tab.key 
              ? 'text-bright-yellow border-b-4 border-bright-yellow -mb-1' 
              : 'text-text-secondary hover:text-text-primary'"
            @click="activeTab = tab.key as any"
          >
            {{ tab.label }}
          </button>
        </div>

        <div class="p-4">
          <div v-if="activeTab === 'overview'" class="space-y-6">
            <div>
              <h3 class="pixel-font text-bright-yellow mb-4">数据完整度</h3>
              <div class="h-4 bg-dark-bg-3 rounded overflow-hidden">
                <div 
                  class="h-full transition-all duration-500"
                  :class="getScoreColor(report.completenessScore).replace('text-', 'bg-')"
                  :style="{ width: `${report.completenessScore}%` }"
                ></div>
              </div>
              <div class="flex justify-between mt-2 text-sm text-text-secondary">
                <span>0%</span>
                <span class="pixel-font" :class="getScoreColor(report.completenessScore)">
                  {{ report.completenessScore.toFixed(1) }}%
                </span>
                <span>100%</span>
              </div>
            </div>

            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div class="pixel-border p-4">
                <h4 class="pixel-font text-neon-blue mb-3">📝 字段缺失情况</h4>
                <div class="space-y-2">
                  <div 
                    v-for="field in report.missingFields.slice(0, 5)" 
                    :key="field.field"
                    class="flex justify-between items-center text-sm"
                  >
                    <span class="text-text-primary">{{ field.label }}</span>
                    <span 
                      class="pixel-badge !text-[10px]"
                      :class="field.missingCount > 0 ? 'pixel-badge-warning' : ''"
                    >
                      {{ field.missingCount }} 条缺失 ({{ field.missingRate.toFixed(1) }}%)
                    </span>
                  </div>
                </div>
                <button 
                  class="pixel-btn !py-1 !px-3 !text-xs mt-3"
                  @click="activeTab = 'missing'"
                >
                  查看全部 →
                </button>
              </div>

              <div class="pixel-border p-4">
                <h4 class="pixel-font text-neon-blue mb-3">⚠️ 异常值分布</h4>
                <div v-if="report.anomalies.length === 0" class="text-green-400 text-sm">
                  ✅ 未发现异常值
                </div>
                <div v-else class="space-y-2">
                  <div class="text-sm">
                    <span class="text-red-400">● 高风险: </span>
                    <span>{{ report.anomalies.filter(a => a.severity === 'high').length }} 项</span>
                  </div>
                  <div class="text-sm">
                    <span class="text-yellow-400">● 中风险: </span>
                    <span>{{ report.anomalies.filter(a => a.severity === 'medium').length }} 项</span>
                  </div>
                  <div class="text-sm">
                    <span class="text-green-400">● 低风险: </span>
                    <span>{{ report.anomalies.filter(a => a.severity === 'low').length }} 项</span>
                  </div>
                </div>
                <button 
                  class="pixel-btn !py-1 !px-3 !text-xs mt-3"
                  @click="activeTab = 'anomalies'"
                >
                  查看全部 →
                </button>
              </div>
            </div>

            <div v-if="report.duplicateGroups.length > 0" class="pixel-border p-4">
              <h4 class="pixel-font text-neon-blue mb-3">🔄 重复卡带 (前3组)</h4>
              <div class="space-y-3">
                <div 
                  v-for="group in report.duplicateGroups.slice(0, 3)" 
                  :key="group.key"
                  class="pixel-border p-3"
                >
                  <div class="flex justify-between items-center">
                    <div>
                      <span class="text-bright-yellow pixel-font text-sm">{{ group.title }}</span>
                      <span class="text-text-secondary text-xs ml-2">
                        {{ group.platform }} · {{ group.releaseYear }}
                      </span>
                    </div>
                    <span class="pixel-badge pixel-badge-warning !text-[10px]">
                      {{ group.count }} 条重复
                    </span>
                  </div>
                </div>
              </div>
              <button 
                class="pixel-btn !py-1 !px-3 !text-xs mt-3"
                @click="activeTab = 'duplicates'"
              >
                查看全部 →
              </button>
            </div>
          </div>

          <div v-if="activeTab === 'duplicates'" class="space-y-4">
            <div v-if="report.duplicateGroups.length === 0" class="text-center py-12">
              <div class="text-4xl mb-4">✅</div>
              <p class="text-green-400 pixel-font">未检测到重复卡带</p>
            </div>

            <div v-else class="space-y-4">
              <div 
                v-for="group in report.duplicateGroups" 
                :key="group.key"
                class="pixel-border p-4"
              >
                <div class="flex justify-between items-start mb-4">
                  <div>
                    <h4 class="pixel-font text-bright-yellow">{{ group.title }}</h4>
                    <p class="text-sm text-text-secondary">
                      平台: {{ group.platform }} · 年份: {{ group.releaseYear }}
                    </p>
                  </div>
                  <div class="flex items-center gap-3">
                    <span class="pixel-badge pixel-badge-warning">
                      {{ group.count }} 条重复
                    </span>
                    <button 
                      class="pixel-btn pixel-btn-primary !py-1 !px-3 !text-xs"
                      @click="openFixDuplicate(group)"
                    >
                      🔧 修复
                    </button>
                  </div>
                </div>

                <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3">
                  <div 
                    v-for="cartridge in group.cartridges"
                    :key="cartridge.id"
                    class="pixel-border p-2 bg-dark-bg-3"
                  >
                    <div class="text-sm text-text-secondary mb-1">ID: {{ cartridge.id }}</div>
                    <div class="text-bright-yellow text-sm truncate">{{ cartridge.title }}</div>
                    <div class="text-xs text-text-secondary mt-1">
                      品相: {{ cartridge.condition }} · 价格: ¥{{ cartridge.purchasePrice }}
                    </div>
                    <div class="text-xs text-text-secondary">
                      添加时间: {{ cartridge.createdAt?.split('T')[0] || '未知' }}
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <div v-if="activeTab === 'missing'" class="space-y-4">
            <div class="overflow-x-auto">
              <table class="w-full">
                <thead>
                  <tr class="border-b-2 border-dark-bg-3">
                    <th class="text-left py-3 px-4 pixel-font text-sm text-neon-blue">字段</th>
                    <th class="text-center py-3 px-4 pixel-font text-sm text-neon-blue">缺失数量</th>
                    <th class="text-center py-3 px-4 pixel-font text-sm text-neon-blue">缺失率</th>
                    <th class="text-right py-3 px-4 pixel-font text-sm text-neon-blue">操作</th>
                  </tr>
                </thead>
                <tbody>
                  <tr 
                    v-for="field in report.missingFields" 
                    :key="field.field"
                    class="border-b border-dark-bg-3 hover:bg-dark-bg-3"
                  >
                    <td class="py-3 px-4 text-text-primary">{{ field.label }}</td>
                    <td class="py-3 px-4 text-center">
                      <span 
                        class="pixel-badge"
                        :class="field.missingCount > 0 ? 'pixel-badge-warning' : ''"
                      >
                        {{ field.missingCount }}
                      </span>
                    </td>
                    <td class="py-3 px-4 text-center">
                      <div class="inline-block w-24 h-2 bg-dark-bg-3 rounded overflow-hidden align-middle mr-2">
                        <div 
                          class="h-full bg-yellow-400"
                          :style="{ width: `${field.missingRate}%` }"
                        ></div>
                      </div>
                      <span class="text-sm text-text-secondary">{{ field.missingRate.toFixed(1) }}%</span>
                    </td>
                    <td class="py-3 px-4 text-right">
                      <button 
                        class="pixel-btn !py-1 !px-3 !text-xs"
                        :disabled="field.missingCount === 0"
                        @click="openFixMissing(field.field)"
                      >
                        🔧 一键修复
                      </button>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>

          <div v-if="activeTab === 'anomalies'" class="space-y-4">
            <div v-if="report.anomalies.length === 0" class="text-center py-12">
              <div class="text-4xl mb-4">✅</div>
              <p class="text-green-400 pixel-font">未检测到异常值</p>
            </div>

            <div v-else class="overflow-x-auto">
              <table class="w-full">
                <thead>
                  <tr class="border-b-2 border-dark-bg-3">
                    <th class="text-left py-3 px-4 pixel-font text-sm text-neon-blue">严重程度</th>
                    <th class="text-left py-3 px-4 pixel-font text-sm text-neon-blue">卡带</th>
                    <th class="text-left py-3 px-4 pixel-font text-sm text-neon-blue">字段</th>
                    <th class="text-left py-3 px-4 pixel-font text-sm text-neon-blue">当前值</th>
                    <th class="text-left py-3 px-4 pixel-font text-sm text-neon-blue">异常原因</th>
                  </tr>
                </thead>
                <tbody>
                  <tr 
                    v-for="anomaly in report.anomalies" 
                    :key="`${anomaly.id}-${anomaly.field}`"
                    class="border-b border-dark-bg-3 hover:bg-dark-bg-3"
                  >
                    <td class="py-3 px-4">
                      <span 
                        class="pixel-badge !text-[10px]"
                        :class="anomaly.severity === 'high' ? 'pixel-badge-danger' : anomaly.severity === 'medium' ? 'pixel-badge-warning' : ''"
                      >
                        {{ severityLabels[anomaly.severity] }}
                      </span>
                    </td>
                    <td class="py-3 px-4 text-bright-yellow">{{ anomaly.title }}</td>
                    <td class="py-3 px-4 text-text-primary">{{ anomaly.label }}</td>
                    <td class="py-3 px-4" :class="severityColors[anomaly.severity]">
                      {{ anomaly.value }}
                    </td>
                    <td class="py-3 px-4 text-text-secondary text-sm">{{ anomaly.reason }}</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </div>
      </div>

      <div class="text-sm text-text-secondary">
        最后扫描时间: {{ report.lastScanTime ? new Date(report.lastScanTime).toLocaleString('zh-CN') : '未知' }}
      </div>
    </div>

    <div
      v-if="fixDuplicateDialog && currentDuplicateGroup"
      class="fixed inset-0 bg-black/70 flex items-center justify-center z-50"
      @click.self="fixDuplicateDialog = false"
    >
      <div class="pixel-card p-6 max-w-lg w-full mx-4 max-h-[80vh] overflow-y-auto">
        <h3 class="text-bright-yellow mb-4">修复重复卡带</h3>
        <p class="text-text-secondary mb-4">
          选择要保留的卡带，其他重复记录将被删除
        </p>

        <div class="space-y-3 mb-6">
          <label
            v-for="cartridge in currentDuplicateGroup.cartridges"
            :key="cartridge.id"
            class="flex items-start gap-3 p-3 pixel-border cursor-pointer"
            :class="{ '!border-bright-yellow': keepId === cartridge.id }"
            @click="keepId = cartridge.id"
          >
            <div
              class="w-5 h-5 border-2 mt-0.5 flex-shrink-0"
              :class="keepId === cartridge.id ? 'bg-bright-yellow border-bright-yellow' : 'border-neon-blue'"
            >
              <span v-if="keepId === cartridge.id" class="text-dark-bg text-xs flex items-center justify-center h-full">✓</span>
            </div>
            <div class="flex-1">
              <div class="text-bright-yellow pixel-font text-sm">{{ cartridge.title }}</div>
              <div class="text-xs text-text-secondary mt-1">
                ID: {{ cartridge.id }} · 品相: {{ cartridge.condition }} · 价格: ¥{{ cartridge.purchasePrice }}
              </div>
              <div class="text-xs text-text-secondary">
                添加时间: {{ cartridge.createdAt?.split('T')[0] || '未知' }}
              </div>
            </div>
          </label>
        </div>

        <div class="flex gap-4 justify-end">
          <button class="pixel-btn" @click="fixDuplicateDialog = false">取消</button>
          <button class="pixel-btn pixel-btn-danger" @click="executeFixDuplicate">
            确认合并删除
          </button>
        </div>
      </div>
    </div>

    <div
      v-if="fixMissingDialog"
      class="fixed inset-0 bg-black/70 flex items-center justify-center z-50"
      @click.self="fixMissingDialog = false"
    >
      <div class="pixel-card p-6 max-w-sm w-full mx-4">
        <h3 class="text-bright-yellow mb-4">一键修复缺失字段</h3>
        <p class="text-text-secondary mb-4">
          将为所有缺失该字段的卡带填充以下值
        </p>

        <div class="mb-6">
          <label class="block text-sm text-text-secondary mb-1 pixel-font">填充值</label>
          <input
            v-model="fixMissingValue"
            type="text"
            class="pixel-input w-full"
            placeholder="请输入填充值"
          />
        </div>

        <div class="flex gap-4 justify-end">
          <button class="pixel-btn" @click="fixMissingDialog = false">取消</button>
          <button class="pixel-btn pixel-btn-primary" @click="executeFixMissing">
            确认填充
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
