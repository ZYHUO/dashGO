<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import api from '@/api'

interface Host {
  id: number
  name: string
  token: string
  ip: string
  status: number
  last_heartbeat?: number
  system_info?: Record<string, any>
}

interface ServerNode {
  id: number
  host_id: number
  name: string
  type: string
  listen_port: number
  group_ids: number[]
  rate: number
  show: boolean
  protocol_settings?: Record<string, any>
  tls_settings?: Record<string, any>
}

const hosts = ref<Host[]>([])
const nodes = ref<ServerNode[]>([])
const selectedHost = ref<Host | null>(null)
const loading = ref(false)

// Modals
const showHostModal = ref(false)
const showNodeModal = ref(false)
const showConfigModal = ref(false)
const showTokenModal = ref(false)

const newHostName = ref('')
const currentToken = ref('')
const configData = ref('')

const editingNode = ref<Partial<ServerNode>>({})

const nodeTypes = [
  { value: 'shadowsocks', label: 'Shadowsocks 2022' },
  { value: 'vless', label: 'VLESS Reality' },
  { value: 'vmess', label: 'VMess' },
  { value: 'trojan', label: 'Trojan' },
  { value: 'hysteria2', label: 'Hysteria2' },
  { value: 'tuic', label: 'TUIC v5' },
  { value: 'anytls', label: 'AnyTLS' },
  { value: 'shadowtls', label: 'ShadowTLS v3' },
  { value: 'naive', label: 'NaiveProxy' },
]

const fetchHosts = async () => {
  loading.value = true
  try {
    const res = await api.get('/api/v2/admin/hosts')
    hosts.value = res.data.data || []
  } finally {
    loading.value = false
  }
}

const fetchNodes = async (hostId: number) => {
  const res = await api.get('/api/v2/admin/nodes', { params: { host_id: hostId } })
  nodes.value = res.data.data || []
}

const selectHost = (host: Host) => {
  selectedHost.value = host
  fetchNodes(host.id)
}

const createHost = async () => {
  if (!newHostName.value) return
  try {
    const res = await api.post('/api/v2/admin/host', { name: newHostName.value })
    currentToken.value = res.data.data.token
    showHostModal.value = false
    showTokenModal.value = true
    newHostName.value = ''
    fetchHosts()
  } catch (e: any) {
    alert(e.response?.data?.error || '创建失败')
  }
}

const deleteHost = async (host: Host) => {
  if (!confirm(`确定删除主机 "${host.name}"？将同时删除所有节点。`)) return
  await api.delete(`/api/v2/admin/host/${host.id}`)
  if (selectedHost.value?.id === host.id) {
    selectedHost.value = null
    nodes.value = []
  }
  fetchHosts()
}

const resetToken = async (host: Host) => {
  if (!confirm('重置后需要重新配置 Agent')) return
  const res = await api.post(`/api/v2/admin/host/${host.id}/reset_token`)
  currentToken.value = res.data.data.token
  showTokenModal.value = true
  fetchHosts()
}

const showConfig = async (host: Host) => {
  const res = await api.get(`/api/v2/admin/host/${host.id}/config`)
  configData.value = JSON.stringify(res.data.data, null, 2)
  showConfigModal.value = true
}

// 节点操作
const openNodeModal = async (node?: ServerNode) => {
  if (node) {
    editingNode.value = { ...node }
  } else {
    // 获取默认配置
    const res = await api.get('/api/v2/admin/node/default', { params: { type: 'shadowsocks' } })
    editingNode.value = {
      host_id: selectedHost.value!.id,
      type: 'shadowsocks',
      group_ids: [1],
      rate: 1,
      show: true,
      ...res.data.data
    }
  }
  showNodeModal.value = true
}

const onTypeChange = async () => {
  const res = await api.get('/api/v2/admin/node/default', { params: { type: editingNode.value.type } })
  const defaults = res.data.data
  editingNode.value = {
    ...editingNode.value,
    name: defaults.name || editingNode.value.name,
    listen_port: defaults.listen_port || editingNode.value.listen_port,
    protocol_settings: defaults.protocol_settings || {},
    tls_settings: defaults.tls_settings || {},
  }
}

const saveNode = async () => {
  try {
    if (editingNode.value.id) {
      await api.put(`/api/v2/admin/node/${editingNode.value.id}`, editingNode.value)
    } else {
      await api.post('/api/v2/admin/node', editingNode.value)
    }
    showNodeModal.value = false
    fetchNodes(selectedHost.value!.id)
  } catch (e: any) {
    alert(e.response?.data?.error || '保存失败')
  }
}

const deleteNode = async (node: ServerNode) => {
  if (!confirm(`确定删除节点 "${node.name}"？`)) return
  await api.delete(`/api/v2/admin/node/${node.id}`)
  fetchNodes(selectedHost.value!.id)
}

const formatTime = (ts?: number) => ts ? new Date(ts * 1000).toLocaleString() : '-'
const getStatusClass = (status: number) => status === 1 ? 'text-green-600' : 'text-gray-400'
const getStatusText = (status: number) => status === 1 ? '在线' : '离线'

// 一键安装脚本
const installScript = computed(() => {
  if (!currentToken.value) return ''
  const panel = window.location.origin
  return `curl -sL https://raw.githubusercontent.com/ZYHUO/xboard-go/main/agent/install.sh | bash -s -- ${panel} ${currentToken.value}`
})

// 手动安装命令
const agentCommand = computed(() => {
  if (!currentToken.value) return ''
  return `./xboard-agent -panel ${window.location.origin} -token ${currentToken.value}`
})

const copyToClipboard = (text: string) => {
  navigator.clipboard.writeText(text)
  alert('已复制到剪贴板')
}

onMounted(fetchHosts)
</script>

<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gray-900">主机管理</h1>
        <p class="text-gray-500 mt-1">管理运行 sing-box 的主机，自动下发配置</p>
      </div>
      <button @click="showHostModal = true" class="px-4 py-2 bg-primary-500 text-white rounded-xl hover:bg-primary-600">
        添加主机
      </button>
    </div>

    <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
      <!-- 主机列表 -->
      <div class="lg:col-span-1 bg-white rounded-2xl shadow-sm border border-gray-100">
        <div class="p-4 border-b border-gray-100">
          <h2 class="font-medium">主机列表</h2>
        </div>
        <div class="divide-y divide-gray-100">
          <div v-for="host in hosts" :key="host.id"
            @click="selectHost(host)"
            :class="selectedHost?.id === host.id ? 'bg-primary-50 border-l-4 border-primary-500' : 'hover:bg-gray-50 border-l-4 border-transparent'"
            class="p-4 cursor-pointer transition">
            <div class="flex items-center justify-between">
              <div>
                <div class="font-medium">{{ host.name }}</div>
                <div class="text-sm text-gray-500">{{ host.ip || '等待连接' }}</div>
                <div class="text-xs" :class="getStatusClass(host.status)">
                  {{ getStatusText(host.status) }}
                  <span v-if="host.last_heartbeat" class="text-gray-400 ml-1">
                    {{ formatTime(host.last_heartbeat) }}
                  </span>
                </div>
              </div>
              <div class="flex flex-col gap-1 text-xs">
                <button @click.stop="showConfig(host)" class="text-blue-600 hover:underline">配置</button>
                <button @click.stop="resetToken(host)" class="text-orange-600 hover:underline">重置</button>
                <button @click.stop="deleteHost(host)" class="text-red-600 hover:underline">删除</button>
              </div>
            </div>
          </div>
          <div v-if="hosts.length === 0" class="p-8 text-center text-gray-400">
            暂无主机
          </div>
        </div>
      </div>

      <!-- 节点列表 -->
      <div class="lg:col-span-2 bg-white rounded-2xl shadow-sm border border-gray-100">
        <div class="p-4 border-b border-gray-100 flex items-center justify-between">
          <h2 class="font-medium">{{ selectedHost ? `${selectedHost.name} - 节点` : '请选择主机' }}</h2>
          <button v-if="selectedHost" @click="openNodeModal()" 
            class="px-3 py-1.5 bg-primary-500 text-white rounded-lg text-sm hover:bg-primary-600">
            添加节点
          </button>
        </div>
        
        <div v-if="selectedHost">
          <div v-if="nodes.length === 0" class="p-8 text-center text-gray-400">
            暂无节点，点击添加
          </div>
          <div v-else class="divide-y divide-gray-100">
            <div v-for="node in nodes" :key="node.id" class="p-4 flex items-center justify-between">
              <div>
                <div class="font-medium">{{ node.name }}</div>
                <div class="text-sm text-gray-500">
                  {{ node.type }} · 端口 {{ node.listen_port }} · {{ node.rate }}x倍率
                </div>
                <div class="text-xs text-gray-400">
                  用户组: {{ node.group_ids?.join(', ') || '无' }}
                </div>
              </div>
              <div class="flex items-center gap-3">
                <span :class="node.show ? 'bg-green-100 text-green-600' : 'bg-gray-100 text-gray-500'" 
                  class="px-2 py-0.5 rounded text-xs">
                  {{ node.show ? '显示' : '隐藏' }}
                </span>
                <button @click="openNodeModal(node)" class="text-primary-600 text-sm hover:underline">编辑</button>
                <button @click="deleteNode(node)" class="text-red-600 text-sm hover:underline">删除</button>
              </div>
            </div>
          </div>
        </div>
        <div v-else class="p-8 text-center text-gray-400">
          请先选择一个主机
        </div>
      </div>
    </div>

    <!-- 添加主机 Modal -->
    <Teleport to="body">
      <div v-if="showHostModal" class="fixed inset-0 z-50 flex items-center justify-center p-4">
        <div class="absolute inset-0 bg-black/30" @click="showHostModal = false"></div>
        <div class="relative bg-white rounded-2xl shadow-xl w-full max-w-md p-6">
          <h3 class="text-lg font-bold mb-4">添加主机</h3>
          <input v-model="newHostName" type="text" placeholder="主机名称" 
            class="w-full px-4 py-2 border border-gray-200 rounded-xl mb-4" />
          <div class="flex gap-3">
            <button @click="showHostModal = false" class="flex-1 px-4 py-2 border border-gray-200 rounded-xl">取消</button>
            <button @click="createHost" class="flex-1 px-4 py-2 bg-primary-500 text-white rounded-xl">创建</button>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- Token 显示 Modal -->
    <Teleport to="body">
      <div v-if="showTokenModal" class="fixed inset-0 z-50 flex items-center justify-center p-4">
        <div class="absolute inset-0 bg-black/30" @click="showTokenModal = false"></div>
        <div class="relative bg-white rounded-2xl shadow-xl w-full max-w-2xl p-6">
          <h3 class="text-lg font-bold mb-4">🎉 主机创建成功</h3>
          
          <!-- 一键安装 -->
          <div class="mb-6">
            <div class="flex items-center justify-between mb-2">
              <span class="text-sm font-medium text-gray-700">一键安装（推荐）</span>
              <button @click="copyToClipboard(installScript)" class="text-xs text-primary-600 hover:underline">复制</button>
            </div>
            <div class="bg-gray-900 text-green-400 p-4 rounded-xl font-mono text-xs break-all cursor-pointer hover:bg-gray-800"
              @click="copyToClipboard(installScript)">
              {{ installScript }}
            </div>
            <p class="text-xs text-gray-500 mt-2">在服务器上执行此命令，自动安装 Agent 和 sing-box</p>
          </div>

          <!-- 手动安装 -->
          <div class="mb-6">
            <div class="flex items-center justify-between mb-2">
              <span class="text-sm font-medium text-gray-700">手动安装</span>
              <button @click="copyToClipboard(agentCommand)" class="text-xs text-primary-600 hover:underline">复制</button>
            </div>
            <div class="bg-gray-900 text-green-400 p-4 rounded-xl font-mono text-xs break-all cursor-pointer hover:bg-gray-800"
              @click="copyToClipboard(agentCommand)">
              {{ agentCommand }}
            </div>
          </div>

          <p class="text-sm text-orange-600 mb-4">⚠️ Token 仅显示一次，请妥善保存</p>
          <button @click="showTokenModal = false" class="w-full px-4 py-2 bg-primary-500 text-white rounded-xl">
            我已保存
          </button>
        </div>
      </div>
    </Teleport>

    <!-- 节点编辑 Modal -->
    <Teleport to="body">
      <div v-if="showNodeModal" class="fixed inset-0 z-50 flex items-center justify-center p-4">
        <div class="absolute inset-0 bg-black/30" @click="showNodeModal = false"></div>
        <div class="relative bg-white rounded-2xl shadow-xl w-full max-w-2xl p-6 max-h-[90vh] overflow-y-auto">
          <h3 class="text-lg font-bold mb-4">{{ editingNode.id ? '编辑节点' : '添加节点' }}</h3>
          
          <div class="space-y-4">
            <div class="grid grid-cols-2 gap-4">
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-1">节点名称</label>
                <input v-model="editingNode.name" type="text" class="w-full px-4 py-2 border border-gray-200 rounded-xl" />
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-1">协议类型</label>
                <select v-model="editingNode.type" @change="onTypeChange" class="w-full px-4 py-2 border border-gray-200 rounded-xl">
                  <option v-for="t in nodeTypes" :key="t.value" :value="t.value">{{ t.label }}</option>
                </select>
              </div>
            </div>
            
            <div class="grid grid-cols-3 gap-4">
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-1">监听端口</label>
                <input v-model.number="editingNode.listen_port" type="number" class="w-full px-4 py-2 border border-gray-200 rounded-xl" />
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-1">倍率</label>
                <input v-model.number="editingNode.rate" type="number" step="0.1" class="w-full px-4 py-2 border border-gray-200 rounded-xl" />
              </div>
              <div class="flex items-center pt-6">
                <input v-model="editingNode.show" type="checkbox" id="nodeShow" class="mr-2" />
                <label for="nodeShow" class="text-sm">显示节点</label>
              </div>
            </div>

            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">用户组 ID（逗号分隔）</label>
              <input 
                :value="editingNode.group_ids?.join(',')"
                @input="editingNode.group_ids = ($event.target as HTMLInputElement).value.split(',').map(id => parseInt(id.trim())).filter(id => !isNaN(id))"
                type="text" placeholder="1,2,3"
                class="w-full px-4 py-2 border border-gray-200 rounded-xl" />
            </div>

            <!-- 协议设置 -->
            <div v-if="editingNode.type === 'shadowsocks'">
              <label class="block text-sm font-medium text-gray-700 mb-1">加密方式</label>
              <select v-model="editingNode.protocol_settings!.method" class="w-full px-4 py-2 border border-gray-200 rounded-xl">
                <option value="2022-blake3-aes-128-gcm">2022-blake3-aes-128-gcm</option>
                <option value="2022-blake3-aes-256-gcm">2022-blake3-aes-256-gcm</option>
                <option value="2022-blake3-chacha20-poly1305">2022-blake3-chacha20-poly1305</option>
              </select>
            </div>

            <div v-if="editingNode.type === 'vless'">
              <label class="block text-sm font-medium text-gray-700 mb-1">Reality 目标域名</label>
              <input v-model="editingNode.tls_settings!.server_name" type="text" placeholder="addons.mozilla.org"
                class="w-full px-4 py-2 border border-gray-200 rounded-xl" />
              <p class="text-xs text-gray-500 mt-1">Agent 会自动生成 Reality 密钥对</p>
            </div>

            <!-- VMess 设置 -->
            <div v-if="editingNode.type === 'vmess'" class="space-y-3">
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-1">传输方式</label>
                <select v-model="editingNode.transport_settings!.type" class="w-full px-4 py-2 border border-gray-200 rounded-xl">
                  <option value="ws">WebSocket</option>
                  <option value="grpc">gRPC</option>
                  <option value="tcp">TCP</option>
                </select>
              </div>
              <div v-if="editingNode.transport_settings?.type === 'ws'">
                <label class="block text-sm font-medium text-gray-700 mb-1">WebSocket 路径</label>
                <input v-model="editingNode.transport_settings!.path" type="text" placeholder="/vmess"
                  class="w-full px-4 py-2 border border-gray-200 rounded-xl" />
              </div>
            </div>

            <!-- TUIC 设置 -->
            <div v-if="editingNode.type === 'tuic'">
              <label class="block text-sm font-medium text-gray-700 mb-1">拥塞控制</label>
              <select v-model="editingNode.protocol_settings!.congestion_control" class="w-full px-4 py-2 border border-gray-200 rounded-xl">
                <option value="bbr">BBR</option>
                <option value="cubic">Cubic</option>
                <option value="new_reno">New Reno</option>
              </select>
            </div>

            <!-- Hysteria2 设置 -->
            <div v-if="editingNode.type === 'hysteria2'" class="grid grid-cols-2 gap-3">
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-1">上行带宽 (Mbps)</label>
                <input v-model.number="editingNode.protocol_settings!.up_mbps" type="number" placeholder="100"
                  class="w-full px-4 py-2 border border-gray-200 rounded-xl" />
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-1">下行带宽 (Mbps)</label>
                <input v-model.number="editingNode.protocol_settings!.down_mbps" type="number" placeholder="100"
                  class="w-full px-4 py-2 border border-gray-200 rounded-xl" />
              </div>
            </div>

            <!-- ShadowTLS 设置 -->
            <div v-if="editingNode.type === 'shadowtls'" class="space-y-3">
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-1">握手服务器</label>
                <input v-model="editingNode.protocol_settings!.handshake_server" type="text" placeholder="addons.mozilla.org"
                  class="w-full px-4 py-2 border border-gray-200 rounded-xl" />
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-1">内层加密方式</label>
                <select v-model="editingNode.protocol_settings!.detour_method" class="w-full px-4 py-2 border border-gray-200 rounded-xl">
                  <option value="2022-blake3-aes-128-gcm">2022-blake3-aes-128-gcm</option>
                  <option value="2022-blake3-aes-256-gcm">2022-blake3-aes-256-gcm</option>
                </select>
              </div>
            </div>

            <!-- Trojan/Naive TLS 设置 -->
            <div v-if="['trojan', 'naive'].includes(editingNode.type!)">
              <label class="block text-sm font-medium text-gray-700 mb-1">TLS 域名 (ACME 自动申请证书)</label>
              <input v-model="editingNode.tls_settings!.server_name" type="text" placeholder="your-domain.com"
                class="w-full px-4 py-2 border border-gray-200 rounded-xl" />
              <p class="text-xs text-gray-500 mt-1">需要将域名解析到服务器 IP</p>
            </div>
          </div>

          <div class="flex gap-3 mt-6">
            <button @click="showNodeModal = false" class="flex-1 px-4 py-2 border border-gray-200 rounded-xl">取消</button>
            <button @click="saveNode" class="flex-1 px-4 py-2 bg-primary-500 text-white rounded-xl">保存</button>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- 配置预览 Modal -->
    <Teleport to="body">
      <div v-if="showConfigModal" class="fixed inset-0 z-50 flex items-center justify-center p-4">
        <div class="absolute inset-0 bg-black/30" @click="showConfigModal = false"></div>
        <div class="relative bg-white rounded-2xl shadow-xl w-full max-w-4xl p-6 max-h-[90vh] overflow-y-auto">
          <h3 class="text-lg font-bold mb-4">配置预览</h3>
          <pre class="bg-gray-900 text-green-400 p-4 rounded-xl text-sm overflow-x-auto max-h-96">{{ configData }}</pre>
          <button @click="showConfigModal = false" class="mt-4 w-full px-4 py-2 bg-gray-500 text-white rounded-xl">关闭</button>
        </div>
      </div>
    </Teleport>
  </div>
</template>
