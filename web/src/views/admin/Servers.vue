<script setup lang="ts">
import { ref, onMounted } from 'vue'
import api from '@/api'

interface Server {
  id: number
  name: string
  type: string
  host: string
  port: string
  server_port: number
  rate: number
  show: boolean
  tags: string[]
  group_id: number[]
  host_id?: number | null
  protocol_settings?: Record<string, any>
}

interface Host {
  id: number
  name: string
  ip: string
  status: number
}

interface ServerStatus {
  online: boolean
  server?: string
  api_version?: string
  users_count?: number
  stats?: {
    uplink_bytes: number
    downlink_bytes: number
    tcp_sessions: number
    udp_sessions: number
  }
  error?: string
}

const servers = ref<Server[]>([])
const hosts = ref<Host[]>([])
const serverStatuses = ref<Record<number, ServerStatus>>({})
const loading = ref(false)
const showModal = ref(false)
const editingServer = ref<Partial<Server> | null>(null)
const syncing = ref<number | null>(null)

const serverTypes = [
  { value: 'shadowsocks', label: 'Shadowsocks' },
  { value: 'vmess', label: 'VMess' },
  { value: 'vless', label: 'VLESS' },
  { value: 'trojan', label: 'Trojan' },
  { value: 'hysteria', label: 'Hysteria2' },
  { value: 'tuic', label: 'TUIC' },
]

const fetchServers = async () => {
  loading.value = true
  try {
    const res = await api.get('/api/v2/admin/servers')
    servers.value = res.data.data || []
    servers.value.forEach(server => {
      fetchServerStatus(server.id)
    })
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

const fetchHosts = async () => {
  try {
    const res = await api.get('/api/v2/admin/hosts')
    hosts.value = res.data.data || []
  } catch (e) {
    console.error(e)
  }
}

const getHostName = (hostId: number | null | undefined) => {
  if (!hostId) return '未绑定'
  const host = hosts.value.find(h => h.id === hostId)
  return host ? host.name : '未知主机'
}

const fetchServerStatus = async (serverId: number) => {
  // 直接设置为在线，实际状态由 Agent 心跳管理
  serverStatuses.value[serverId] = { online: true }
}

const syncServer = async (serverId: number) => {
  syncing.value = serverId
  try {
    await api.post(`/api/v2/admin/server/${serverId}/sync`)
    alert('同步请求已发送，Agent 将自动更新配置')
  } catch (e: any) {
    alert(e.response?.data?.error || '同步失败')
  } finally {
    syncing.value = null
  }
}

const openCreateModal = () => {
  editingServer.value = {
    name: '',
    type: 'shadowsocks',
    host: '',
    port: '',
    server_port: 8388,
    rate: 1,
    show: true,
    group_id: [],
    host_id: null,
    protocol_settings: {
      cipher: 'aes-256-gcm'
    }
  }
  showModal.value = true
}

const openEditModal = (server: Server) => {
  editingServer.value = { 
    ...server,
    protocol_settings: server.protocol_settings || { cipher: 'aes-256-gcm' }
  }
  showModal.value = true
}

const saveServer = async () => {
  if (!editingServer.value) return
  try {
    if (editingServer.value.id) {
      await api.put(`/api/v2/admin/server/${editingServer.value.id}`, editingServer.value)
    } else {
      await api.post('/api/v2/admin/server', editingServer.value)
    }
    showModal.value = false
    fetchServers()
  } catch (e: any) {
    alert(e.response?.data?.error || '保存失败')
  }
}

const deleteServer = async (server: Server) => {
  if (!confirm(`确定要删除节点 "${server.name}" 吗？`)) return
  try {
    await api.delete(`/api/v2/admin/server/${server.id}`)
    fetchServers()
  } catch (e: any) {
    alert(e.response?.data?.error || '删除失败')
  }
}

const formatBytes = (bytes: number) => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

onMounted(() => {
  fetchServers()
  fetchHosts()
})
</script>

<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gray-900">节点管理</h1>
        <p class="text-gray-500 mt-1">管理代理节点</p>
      </div>
      <button @click="openCreateModal" class="px-4 py-2 bg-primary-500 text-white rounded-xl hover:bg-primary-600 transition">
        添加节点
      </button>
    </div>

    <div class="bg-white rounded-xl shadow-sm overflow-hidden">
      <div v-if="loading" class="text-center py-12 text-gray-500">加载中...</div>

      <table v-else class="w-full">
        <thead class="bg-gray-50 border-b border-gray-200">
          <tr>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">名称</th>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">类型</th>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">地址</th>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">绑定主机</th>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">状态</th>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">流量</th>
            <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase">操作</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-gray-200">
          <tr v-for="server in servers" :key="server.id" class="hover:bg-gray-50">
            <td class="px-6 py-4">
              <div class="font-medium text-gray-900">{{ server.name }}</div>
              <div class="text-xs text-gray-400">{{ server.rate }}x 倍率</div>
            </td>
            <td class="px-6 py-4">
              <span class="px-2 py-1 bg-blue-100 text-blue-600 rounded-full text-xs">{{ server.type }}</span>
            </td>
            <td class="px-6 py-4 text-sm text-gray-500">{{ server.host }}:{{ server.port }}</td>
            <td class="px-6 py-4">
              <span v-if="server.host_id" class="px-2 py-1 bg-green-100 text-green-600 rounded-full text-xs">
                {{ getHostName(server.host_id) }}
              </span>
              <span v-else class="text-gray-400 text-xs">未绑定</span>
            </td>
            <td class="px-6 py-4">
              <div v-if="serverStatuses[server.id]">
                <span v-if="serverStatuses[server.id].online" class="flex items-center gap-1.5">
                  <span class="w-2 h-2 bg-green-500 rounded-full animate-pulse"></span>
                  <span class="text-green-600 text-sm">在线</span>
                </span>
                <span v-else class="flex items-center gap-1.5">
                  <span class="w-2 h-2 bg-red-500 rounded-full"></span>
                  <span class="text-red-600 text-sm">离线</span>
                </span>
              </div>
              <span v-else class="text-gray-400 text-sm">检测中...</span>
            </td>
            <td class="px-6 py-4 text-sm">
              <div v-if="serverStatuses[server.id]?.stats">
                <div class="text-gray-600">↑ {{ formatBytes(serverStatuses[server.id].stats!.uplink_bytes) }}</div>
                <div class="text-gray-600">↓ {{ formatBytes(serverStatuses[server.id].stats!.downlink_bytes) }}</div>
              </div>
              <span v-else class="text-gray-400">-</span>
            </td>
            <td class="px-6 py-4 text-right space-x-2">
              <button @click="syncServer(server.id)" :disabled="syncing === server.id" class="text-green-600 hover:text-green-700 text-sm disabled:opacity-50">
                {{ syncing === server.id ? '同步中...' : '同步' }}
              </button>
              <button @click="openEditModal(server)" class="text-primary-600 hover:text-primary-700 text-sm">编辑</button>
              <button @click="deleteServer(server)" class="text-red-600 hover:text-red-700 text-sm">删除</button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Modal -->
    <Teleport to="body">
      <div v-if="showModal" class="fixed inset-0 z-50 flex items-center justify-center p-4">
        <div class="absolute inset-0 bg-black/30" @click="showModal = false"></div>
        <div class="relative bg-white rounded-2xl shadow-xl w-full max-w-lg p-6 max-h-[90vh] overflow-y-auto">
          <h3 class="text-lg font-bold mb-4">{{ editingServer?.id ? '编辑节点' : '添加节点' }}</h3>
          
          <div class="space-y-4">
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">名称</label>
              <input v-model="editingServer!.name" type="text" class="w-full px-4 py-2 border border-gray-200 rounded-xl" />
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">类型</label>
              <select v-model="editingServer!.type" class="w-full px-4 py-2 border border-gray-200 rounded-xl">
                <option v-for="t in serverTypes" :key="t.value" :value="t.value">{{ t.label }}</option>
              </select>
            </div>
            <div class="grid grid-cols-2 gap-4">
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-1">地址</label>
                <input v-model="editingServer!.host" type="text" class="w-full px-4 py-2 border border-gray-200 rounded-xl" />
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-1">端口</label>
                <input v-model="editingServer!.port" type="text" class="w-full px-4 py-2 border border-gray-200 rounded-xl" />
              </div>
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">倍率</label>
              <input v-model.number="editingServer!.rate" type="number" step="0.1" class="w-full px-4 py-2 border border-gray-200 rounded-xl" />
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">服务端口</label>
              <input v-model.number="editingServer!.server_port" type="number" class="w-full px-4 py-2 border border-gray-200 rounded-xl" placeholder="sing-box 监听端口" />
              <p class="text-xs text-gray-400 mt-1">sing-box 实际监听的端口</p>
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">绑定主机</label>
              <select v-model="editingServer!.host_id" class="w-full px-4 py-2 border border-gray-200 rounded-xl">
                <option :value="null">不绑定（手动配置）</option>
                <option v-for="h in hosts" :key="h.id" :value="h.id">
                  {{ h.name }} ({{ h.ip || '未知IP' }})
                </option>
              </select>
              <p class="text-xs text-gray-400 mt-1">绑定后将自动部署到主机</p>
            </div>
            <div class="flex items-center gap-2">
              <input v-model="editingServer!.show" type="checkbox" id="show" class="rounded" />
              <label for="show" class="text-sm text-gray-700">显示节点</label>
            </div>

            <!-- Shadowsocks 特定配置 -->
            <div v-if="editingServer!.type === 'shadowsocks'" class="border-t pt-4 mt-4">
              <h4 class="text-sm font-medium text-gray-700 mb-3">Shadowsocks 配置</h4>
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-1">加密方式</label>
                <select v-model="editingServer!.protocol_settings!.cipher" class="w-full px-4 py-2 border border-gray-200 rounded-xl">
                  <option value="aes-256-gcm">aes-256-gcm</option>
                  <option value="aes-128-gcm">aes-128-gcm</option>
                  <option value="chacha20-ietf-poly1305">chacha20-ietf-poly1305</option>
                  <option value="2022-blake3-aes-128-gcm">2022-blake3-aes-128-gcm</option>
                  <option value="2022-blake3-aes-256-gcm">2022-blake3-aes-256-gcm</option>
                  <option value="2022-blake3-chacha20-poly1305">2022-blake3-chacha20-poly1305</option>
                </select>
              </div>
            </div>
          </div>

          <div class="flex gap-3 mt-6">
            <button @click="showModal = false" class="flex-1 px-4 py-2 border border-gray-200 text-gray-600 rounded-xl hover:bg-gray-50">取消</button>
            <button @click="saveServer" class="flex-1 px-4 py-2 bg-primary-500 text-white rounded-xl hover:bg-primary-600">保存</button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>
