<script setup lang="ts">
import { ref, onMounted } from 'vue'
import api from '@/api'

const loading = ref(false)
const threshold = ref(80)
const warningUsers = ref<any[]>([])
const selectedUsers = ref<number[]>([])
const showConfirmDialog = ref(false)
const confirmAction = ref<string>('')
const confirmMessage = ref<string>('')
const confirmCallback = ref<(() => void) | null>(null)

const formatBytes = (bytes: number) => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

const fetchWarningUsers = async () => {
  loading.value = true
  try {
    const res = await api.get(`/api/v2/admin/traffic/warnings?threshold=${threshold.value}`)
    warningUsers.value = res.data.data || []
  } catch (e) {
    console.error('获取流量预警用户失败:', e)
  } finally {
    loading.value = false
  }
}

const resetUserTraffic = async (userId: number) => {
  confirmAction.value = 'reset'
  confirmMessage.value = '确定要重置该用户的流量吗？'
  confirmCallback.value = async () => {
    try {
      await api.post(`/api/v2/admin/traffic/reset/${userId}`)
      alert('流量重置成功')
      fetchWarningUsers()
    } catch (e) {
      alert('流量重置失败')
    }
  }
  showConfirmDialog.value = true
}

const batchResetTraffic = async () => {
  if (selectedUsers.value.length === 0) {
    alert('请选择要重置流量的用户')
    return
  }
  
  confirmAction.value = 'batchReset'
  confirmMessage.value = `确定要重置选中的 ${selectedUsers.value.length} 个用户的流量吗？`
  confirmCallback.value = async () => {
    try {
      for (const userId of selectedUsers.value) {
        await api.post(`/api/v2/admin/traffic/reset/${userId}`)
      }
      alert('批量重置成功')
      selectedUsers.value = []
      fetchWarningUsers()
    } catch (e) {
      alert('批量重置失败')
    }
  }
  showConfirmDialog.value = true
}

const sendWarning = async (userId: number) => {
  try {
    await api.post(`/api/v2/admin/traffic/warning/${userId}`)
    alert('预警通知已发送')
  } catch (e) {
    alert('发送预警通知失败')
  }
}

const batchSendWarnings = async () => {
  confirmAction.value = 'batchWarning'
  confirmMessage.value = `确定要向所有流量使用超过 ${threshold.value}% 的用户发送预警通知吗？`
  confirmCallback.value = async () => {
    try {
      const res = await api.post(`/api/v2/admin/traffic/warnings/send?threshold=${threshold.value}`)
      alert(`批量发送完成，成功 ${res.data.success}/${res.data.total} 个`)
      fetchWarningUsers()
    } catch (e) {
      alert('批量发送失败')
    }
  }
  showConfirmDialog.value = true
}

const autoBanUsers = async () => {
  confirmAction.value = 'autoBan'
  confirmMessage.value = '确定要自动封禁所有流量超限的用户吗？此操作不可撤销！'
  confirmCallback.value = async () => {
    try {
      const res = await api.post('/api/v2/admin/traffic/autoban')
      alert(`已封禁 ${res.data.count} 个超流量用户`)
      fetchWarningUsers()
    } catch (e) {
      alert('自动封禁失败')
    }
  }
  showConfirmDialog.value = true
}

const confirmDialogAction = () => {
  if (confirmCallback.value) {
    confirmCallback.value()
  }
  showConfirmDialog.value = false
}

const cancelDialogAction = () => {
  showConfirmDialog.value = false
}

const toggleSelectAll = () => {
  if (selectedUsers.value.length === warningUsers.value.length) {
    selectedUsers.value = []
  } else {
    selectedUsers.value = warningUsers.value.map(u => u.id)
  }
}

const toggleSelect = (userId: number) => {
  const index = selectedUsers.value.indexOf(userId)
  if (index > -1) {
    selectedUsers.value.splice(index, 1)
  } else {
    selectedUsers.value.push(userId)
  }
}

onMounted(fetchWarningUsers)
</script>

<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gray-900">流量管理</h1>
        <p class="text-gray-500 text-sm mt-1">管理流量预警用户和流量重置</p>
      </div>
      <button @click="fetchWarningUsers" :disabled="loading" class="px-4 py-2 text-sm text-gray-600 hover:bg-gray-100 rounded-xl transition-colors">
        {{ loading ? '加载中...' : '刷新' }}
      </button>
    </div>

    <!-- 流量统计概览 -->
    <div class="bg-white rounded-2xl p-6 shadow-sm">
      <h3 class="text-lg font-semibold mb-4">流量统计概览</h3>
      <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
        <div class="bg-gray-50 rounded-xl p-4">
          <div class="text-sm text-gray-500 mb-1">预警用户数</div>
          <div class="text-2xl font-bold text-orange-600">{{ warningUsers.length }}</div>
        </div>
        <div class="bg-gray-50 rounded-xl p-4">
          <div class="text-sm text-gray-500 mb-1">超限用户数</div>
          <div class="text-2xl font-bold text-red-600">{{ warningUsers.filter(u => u.is_over_limit).length }}</div>
        </div>
        <div class="bg-gray-50 rounded-xl p-4">
          <div class="text-sm text-gray-500 mb-1">流量阈值</div>
          <div class="flex items-center gap-2">
            <input 
              v-model.number="threshold" 
              type="number" 
              min="0" 
              max="100" 
              class="w-20 px-3 py-1 border border-gray-300 rounded-lg"
            />
            <span class="text-sm text-gray-600">%</span>
            <button @click="fetchWarningUsers" class="px-3 py-1 text-sm bg-indigo-600 text-white rounded-lg hover:bg-indigo-700">
              应用
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- 批量操作 -->
    <div class="bg-white rounded-2xl p-6 shadow-sm">
      <h3 class="text-lg font-semibold mb-4">批量操作</h3>
      <div class="flex flex-wrap gap-3">
        <button 
          @click="batchResetTraffic" 
          :disabled="selectedUsers.length === 0"
          class="px-4 py-2 bg-blue-600 text-white rounded-xl hover:bg-blue-700 disabled:bg-gray-300 disabled:cursor-not-allowed transition-colors"
        >
          🔄 批量重置流量 ({{ selectedUsers.length }})
        </button>
        <button 
          @click="batchSendWarnings"
          class="px-4 py-2 bg-orange-600 text-white rounded-xl hover:bg-orange-700 transition-colors"
        >
          📧 批量发送预警
        </button>
        <button 
          @click="autoBanUsers"
          class="px-4 py-2 bg-red-600 text-white rounded-xl hover:bg-red-700 transition-colors"
        >
          🚫 自动封禁超限用户
        </button>
      </div>
    </div>

    <!-- 流量预警用户列表 -->
    <div class="bg-white rounded-2xl p-6 shadow-sm">
      <h3 class="text-lg font-semibold mb-4">流量预警用户列表</h3>
      <div class="overflow-x-auto">
        <table class="w-full">
          <thead>
            <tr class="text-left text-sm text-gray-500 border-b border-gray-100">
              <th class="pb-3 font-medium">
                <input 
                  type="checkbox" 
                  :checked="selectedUsers.length === warningUsers.length && warningUsers.length > 0"
                  @change="toggleSelectAll"
                  class="rounded"
                />
              </th>
              <th class="pb-3 font-medium">用户</th>
              <th class="pb-3 font-medium text-right">已用流量</th>
              <th class="pb-3 font-medium text-right">总流量</th>
              <th class="pb-3 font-medium text-right">使用率</th>
              <th class="pb-3 font-medium text-right">状态</th>
              <th class="pb-3 font-medium text-right">操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="user in warningUsers" :key="user.id" class="border-b border-gray-50 hover:bg-gray-50">
              <td class="py-3">
                <input 
                  type="checkbox" 
                  :checked="selectedUsers.includes(user.id)"
                  @change="toggleSelect(user.id)"
                  class="rounded"
                />
              </td>
              <td class="py-3 text-sm text-gray-900">{{ user.email }}</td>
              <td class="py-3 text-sm text-gray-500 text-right">{{ formatBytes(user.total_used) }}</td>
              <td class="py-3 text-sm text-gray-500 text-right">{{ formatBytes(user.transfer_enable) }}</td>
              <td class="py-3 text-sm text-right">
                <span :class="[
                  'inline-flex items-center px-2 py-1 rounded-full text-xs font-medium',
                  user.usage_percent >= 100 ? 'bg-red-100 text-red-700' :
                  user.usage_percent >= 90 ? 'bg-orange-100 text-orange-700' :
                  'bg-yellow-100 text-yellow-700'
                ]">
                  {{ user.usage_percent.toFixed(1) }}%
                </span>
              </td>
              <td class="py-3 text-sm text-right">
                <span v-if="user.is_over_limit" class="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-red-100 text-red-700">
                  超限
                </span>
                <span v-else class="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-green-100 text-green-700">
                  正常
                </span>
              </td>
              <td class="py-3 text-right">
                <div class="flex items-center justify-end gap-2">
                  <button 
                    @click="resetUserTraffic(user.id)"
                    class="px-3 py-1 text-xs bg-blue-600 text-white rounded-lg hover:bg-blue-700"
                  >
                    重置
                  </button>
                  <button 
                    @click="sendWarning(user.id)"
                    class="px-3 py-1 text-xs bg-orange-600 text-white rounded-lg hover:bg-orange-700"
                  >
                    通知
                  </button>
                </div>
              </td>
            </tr>
            <tr v-if="warningUsers.length === 0">
              <td colspan="7" class="py-8 text-center text-gray-400">暂无预警用户</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- 确认对话框 -->
    <div v-if="showConfirmDialog" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div class="bg-white rounded-2xl p-6 max-w-md w-full mx-4">
        <h3 class="text-lg font-semibold mb-4">确认操作</h3>
        <p class="text-gray-600 mb-6">{{ confirmMessage }}</p>
        <div class="flex justify-end gap-3">
          <button 
            @click="cancelDialogAction"
            class="px-4 py-2 text-gray-600 hover:bg-gray-100 rounded-xl transition-colors"
          >
            取消
          </button>
          <button 
            @click="confirmDialogAction"
            class="px-4 py-2 bg-indigo-600 text-white rounded-xl hover:bg-indigo-700 transition-colors"
          >
            确认
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
