<script setup lang="ts">
import {computed, onMounted, onUnmounted, ref, watch} from 'vue'
import {useMessage, useDialog, NTabs, NTabPane, NDynamicInput, NButton, NIcon, NInput, NModal, NCard, NSpace, NSelect, NBadge, NSplit} from 'naive-ui'
import {
  CloudDownloadOutline,
  PlayOutline,
  SaveOutline,
  SettingsOutline,
  ListOutline,
  InformationCircleOutline,
  PencilOutline,
  CloudOfflineOutline,
  CloudDoneOutline
} from '@vicons/ionicons5'
import CodeEditor from './CodeEditor.vue'
import QueryParamsEditor from './QueryParamsEditor.vue'
import PathVariablesEditor from './PathVariablesEditor.vue'
import HeadersEditor from './HeadersEditor.vue'
import CaseBar from './CaseBar.vue'
import Sidebar from './Sidebar.vue'
import EnvManager from './EnvManager.vue'
import ResponsePanel from './ResponsePanel.vue'
import SettingsModal from './SettingsModal.vue'
import RequestMetaModal from './RequestMetaModal.vue'
import BodyDocsViewer from './BodyDocsViewer.vue'
import WebSocketPanel from './WebSocketPanel.vue'
import {useRequestStore} from '../stores/request'
import {useEnvStore} from '../stores/env'
import {useSettingsStore} from '../stores/settings'
import { useWebSocketStore } from '../stores/websocket'

const message = useMessage()
const dialog = useDialog()
// @ts-ignore
window.$message = message
// @ts-ignore
window.$dialog = dialog
const store = useRequestStore()
const envStore = useEnvStore()
const settingsStore = useSettingsStore()
const wsStore = useWebSocketStore()

// Modal States
const showEnvModal = ref(false)
const showSettingsModal = ref(false)
const showMetaModal = ref(false)

// WS Connection State Helper
const isWsConnected = computed(() => {
  if (store.request.method !== 'WS') return false
  const sessionId = store.meta?.id || 'temp-session'
  return wsStore.getSession(sessionId)?.isConnected || false
})

// Horizontal split logic
const horizontalSplitSize = ref(0.2) // Default 20%
const handleHorizontalDragEnd = () => {
  localStorage.setItem('horizontal-split-size', horizontalSplitSize.value.toString())
}

onMounted(async () => {
  // Load global settings
  settingsStore.load()
  
  // Restore horizontal split size
  const savedSize = localStorage.getItem('horizontal-split-size')
  if (savedSize) {
    horizontalSplitSize.value = parseFloat(savedSize)
  }

  const restored = await store.init()
  if (restored) {
    if (store.workDir) {
       await envStore.loadEnvs()
    }
  }
})

// Split pane logic (Vertical)
const requestHeightPercent = ref(50) // Initial height percentage for the top panel
const isDragging = ref(false)
const containerRef = ref<HTMLElement | null>(null)

const startDrag = () => {
  isDragging.value = true
  document.addEventListener('mousemove', onDrag)
  document.addEventListener('mouseup', stopDrag)
  document.body.style.cursor = 'row-resize'
  document.body.style.userSelect = 'none'
}

const onDrag = (e: MouseEvent) => {
  if (!isDragging.value || !containerRef.value) return
  
  const containerRect = containerRef.value.getBoundingClientRect()
  const relativeY = e.clientY - containerRect.top
  const percentage = (relativeY / containerRect.height) * 100
  
  // Limit the resizing range (e.g., between 10% and 90%)
  if (percentage >= 10 && percentage <= 90) {
    requestHeightPercent.value = percentage
  }
}

const stopDrag = () => {
  isDragging.value = false
  document.removeEventListener('mousemove', onDrag)
  document.removeEventListener('mouseup', stopDrag)
  document.body.style.cursor = ''
  document.body.style.userSelect = ''
}

onUnmounted(() => {
  // Cleanup in case component is destroyed while dragging
  document.removeEventListener('mousemove', onDrag)
  document.removeEventListener('mouseup', stopDrag)
})

const handleSave = async () => {
  if (!store.workDir) {
    message.warning("请先选择一个工作目录 (Click 'Open' in Sidebar)")
    return
  }

  if (store.currentFileName) {
    try {
      await store.saveCurrent()
      message.success("保存成功")
    } catch (e: any) {
      message.error(e.message || "保存失败")
    }
  } else {
    showMetaModal.value = true
  }
}

const handleMetaSave = async (data: { summary: string; tag: string; description: string }) => {
  try {
    // 1. Update Store State (Initializes meta if missing)
    await store.updateRequestMeta({
      summary: data.summary,
      tag: data.tag,
      description: data.description
    })
    
    // 2. Perform Save
    // saveCurrent will now handle unique filename generation if currentFileName is empty
    await store.saveCurrent()
    
    // 3. Close Modal
    showMetaModal.value = false

    message.success("保存成功")
  } catch (e: any) {
    console.error('handleMetaSave error:', e)
    message.error(e.message || "保存失败")
  }
}

const handleRun = async () => {
  if (!store.request.url) {
    message.warning("URL不能为空")
    return
  }

  await store.send()

  if (store.response.error) {
    message.error("请求失败")
  } else {
    message.success(`Status: ${store.response.statusCode} | Time: ${store.response.time}ms`)
  }
}

const handleWsAction = () => {
  const sessionId = store.meta?.id || 'temp-session'
  if (isWsConnected.value) {
    wsStore.disconnect(sessionId)
  } else {
    if (!store.request.url) {
      message.warning("URL不能为空")
      return
    }
    // TODO: Pass headers from store.request.headers
    wsStore.connect(sessionId, store.request.url, store.request.headers || {})
  }
}

const handleEnvChange = (val: string) => {
  envStore.setActiveEnv(val)
  envStore.saveEnvs()
}

const methodOptions = [
  { label: 'GET', value: 'GET' },
  { label: 'POST', value: 'POST' },
  { label: 'PUT', value: 'PUT' },
  { label: 'DELETE', value: 'DELETE' },
  { label: 'PATCH', value: 'PATCH' },
  { label: 'HEAD', value: 'HEAD' },
  { label: 'OPTIONS', value: 'OPTIONS' },
  { label: 'WS', value: 'WS', style: { color: '#a78bfa', fontWeight: 'bold' } }
]

const handleRequestBaseChange = () => {
  store.syncToCurl()
}

// ================= Env Replacement Logic =================
const possibleReplacement = ref<string | null>(null)
watch(
  () => store.curlCode,
  (newCode) => {
    if (newCode && newCode.trim().length > 10) {
      possibleReplacement.value = envStore.reverseReplace(newCode)
    } else {
      possibleReplacement.value = null
    }
  }
)

const applyReplacement = () => {
  if (possibleReplacement.value) {
    store.curlCode = possibleReplacement.value
    store.syncFromCurl()
    possibleReplacement.value = null
    message.success('Applied environment variables')
  }
}

// ================= Body Logic =================
const handleBodyChange = (val: string) => {
  store.request.body = val
  store.syncToCurl()
}

const formatBody = () => {
  if (!store.request.body) return
  try {
    const jsonObj = JSON.parse(store.request.body)
    store.request.body = JSON.stringify(jsonObj, null, 2)
    store.syncToCurl()
    message.success('JSON formatted')
  } catch (e) {
    message.warning('Invalid JSON content')
  }
}

// ================= Curl Logic =================
const handleCurlChange = (val: string) => {
  store.curlCode = val
  store.syncFromCurl()
}

const copyRealCurl = async () => {
  const curl = store.curlCode;
  if (!curl) return;

  // Use envStore as the source of truth for active environment
  const activeEnvName = envStore.activeEnvName;
  const currentVars = store.envConfig.environments[activeEnvName]?.variables || {};

  // Regex replacement: find all {{key}}
  const replacedCurl = curl.replace(/{{(.*?)}}/g, (match, key) => {
    const trimmedKey = key.trim();
    // Check if the key exists in the current environment's variables
    return currentVars[trimmedKey] !== undefined ? currentVars[trimmedKey] : match;
  });

  try {
    await navigator.clipboard.writeText(replacedCurl);
    message.success(`已复制 Curl (环境: ${activeEnvName || '无'})`);
  } catch (err) {
    message.error("复制失败");
  }
}

// ================= Global Shortcuts =================
const handleGlobalKeydown = (e: KeyboardEvent) => {
  // Save: Ctrl+S or Cmd+S
  if ((e.ctrlKey || e.metaKey) && e.key.toLowerCase() === 's') {
    e.preventDefault()
    handleSave()
  }
  
  // Run: Ctrl+Enter or Cmd+Enter
  if ((e.ctrlKey || e.metaKey) && e.key === 'Enter') {
    if (!store.isLoading) {
      e.preventDefault()
      handleRun()
    }
  }
}

onMounted(async () => {
  // Load global settings
  settingsStore.load()
  
  const restored = await store.init()
  if (restored) {
    if (store.workDir) {
       await envStore.loadEnvs()
    }
  }

  window.addEventListener('keydown', handleGlobalKeydown)
})

onUnmounted(() => {
  // Cleanup
  document.removeEventListener('mousemove', onDrag)
  document.removeEventListener('mouseup', stopDrag)
  window.removeEventListener('keydown', handleGlobalKeydown)
})
</script>

<template>
  <div class="h-screen w-screen bg-gray-900 text-gray-200 overflow-hidden">
    <n-split direction="horizontal" v-model:size="horizontalSplitSize" :min="0.15" :max="0.4" @drag-end="handleHorizontalDragEnd" class="h-full">
      <template #1>
        <div class="split-pane sidebar-container h-full overflow-hidden">
          <Sidebar />
        </div>
      </template>
      <template #2>
        <div class="split-pane content-container h-full p-4 flex flex-col gap-4 min-w-0 overflow-hidden">
          <!-- Header -->
          <div class="flex items-center justify-between shrink-0">
            <div class="flex items-center gap-2">
              <n-icon size="24" color="#4ade80">
                <CloudDownloadOutline/>
              </n-icon>
              <span class="font-bold text-lg tracking-wide text-white">CurlFlow</span>
            </div>
            <div class="flex items-center gap-2">
              <n-select
                  :value="envStore.activeEnvName"
                  :options="envStore.envOptions"
                  size="small"
                  placeholder="Select Env"
                  style="width: 150px"
                  @update:value="handleEnvChange"
              />
                            <n-button 
                              v-if="store.currentFileName"
                              secondary
                              size="small"
                              @click="showMetaModal = true"
                              class="px-2 text-gray-300"
                              title="Edit Interface Info"
                            >
              
                <template #icon>
                  <n-icon>
                    <InformationCircleOutline/>
                  </n-icon>
                </template>
              </n-button>
              <n-button
                  secondary
                  size="small"
                  @click="handleSave"
                  class="px-4 text-gray-300"
              >
                <template #icon>
                  <n-icon>
                    <SaveOutline/>
                  </n-icon>
                </template>
                Save
              </n-button>
              <n-button
                  secondary
                  size="small"
                  @click="showEnvModal = true"
                  class="px-2 text-gray-300"
                  title="Environment Variables"
              >
                <template #icon>
                  <n-icon>
                    <ListOutline/>
                  </n-icon>
                </template>
              </n-button>
              <n-button
                  secondary
                  size="small"
                  @click="showSettingsModal = true"
                  class="px-2 text-gray-300"
                  title="Global Settings"
              >
                <template #icon>
                  <n-icon>
                    <SettingsOutline/>
                  </n-icon>
                </template>
              </n-button>
              
              <!-- Dynamic Action Button -->
              <n-button
                  v-if="store.request.method !== 'WS'"
                  type="primary"
                  size="small"
                  :loading="store.isLoading"
                  @click="handleRun"
                  class="px-6"
              >
                <template #icon>
                  <n-icon>
                    <PlayOutline/>
                  </n-icon>
                </template>
                Run
              </n-button>

              <n-button
                  v-else
                  :type="isWsConnected ? 'error' : 'primary'"
                  size="small"
                  @click="handleWsAction"
                  class="px-6"
              >
                <template #icon>
                  <n-icon :component="isWsConnected ? CloudOfflineOutline : CloudDoneOutline" />
                </template>
                {{ isWsConnected ? 'Disconnect' : 'Connect' }}
              </n-button>

            </div>
          </div>

          <!-- Main Content Area -->
          <div class="flex-1 flex flex-col min-h-0 relative" ref="containerRef">
            <!-- Request Section (Top) -->
            <div class="flex flex-col gap-2 min-h-0" :style="store.request.method === 'WS' ? { height: '100%' } : { height: `${requestHeightPercent}%` }">
              <div class="text-xs font-bold font-mono text-gray-500 uppercase tracking-widest flex items-center justify-between shrink-0">
                <div class="flex items-center gap-2">
                  <div class="w-1.5 h-1.5 rounded-full bg-blue-500"></div>
                  Request
                  <span v-if="store.meta?.summary" class="ml-2 text-gray-400 normal-case font-sans truncate max-w-[300px]">
                    {{ store.meta.summary }}
                  </span>
                  <n-button 
                    v-if="store.currentFileName"
                    text 
                    size="tiny" 
                    @click="showMetaModal = true" 
                    class="ml-1 text-gray-600 hover:text-blue-400"
                  >
                    <template #icon>
                      <n-icon :component="PencilOutline" />
                    </template>
                  </n-button>
                </div>
              </div>
              
              <div class="flex flex-col flex-1 min-h-0 bg-gray-800 rounded-lg border border-gray-700/50 p-3 overflow-hidden">
                <!-- Optional CaseBar -->
                <CaseBar v-if="store.meta && store.meta.id" class="mb-3" />

                <!-- URL Bar -->
                <div class="mb-4">
                  <n-input-group>
                    <n-select
                      v-model:value="store.request.method"
                      :options="methodOptions"
                      :style="{ width: '120px' }"
                      @update:value="handleRequestBaseChange"
                    />
                    <n-input
                      v-model:value="store.request.url"
                      placeholder="https://api.example.com/v1/resource"
                      @update:value="handleRequestBaseChange"
                      class="flex-1"
                    />
                  </n-input-group>
                </div>

                <!-- WebSocket Panel (Conditional) -->
                <div v-if="store.request.method === 'WS'" class="flex-1 flex flex-col min-h-0 bg-[#1e1e1e] rounded overflow-hidden">
                  <WebSocketPanel 
                    :requestId="store.meta?.id || 'temp-session'" 
                    :request="store.request"
                  />
                </div>

                <!-- HTTP Request Tabs (Conditional) -->
                <div v-else-if="store.request" class="flex-1 flex flex-col min-h-0">
                  <n-tabs
                    v-model:value="store.activeEditorTab"
                    type="line"
                    size="small"
                    class="flex-1 flex flex-col overflow-hidden"
                    display-directive="show"
                    pane-style="height: 100%; overflow: hidden; display: flex; flex-direction: column;"
                  >
                    <!-- Tab 1: Path -->
                    <n-tab-pane name="Path">
                      <template #tab>
                        <div class="flex items-center gap-1">
                          Path
                          <n-badge 
                            :value="Object.keys(store.pathParams).length" 
                            :show="Object.keys(store.pathParams).length > 0" 
                            type="info"
                            :offset="[4, -4]"
                            size="small"
                          />
                        </div>
                      </template>
                      <div class="py-2 h-full overflow-auto">
                        <PathVariablesEditor
                          :url="store.request.url"
                          v-model:modelValue="store.pathParams"
                          :meta="store.meta"
                        />
                      </div>
                    </n-tab-pane>

                    <!-- Tab 2: Query -->
                    <n-tab-pane name="Params" tab="Params">
                      <div class="py-2 h-full overflow-auto">
                        <QueryParamsEditor
                          v-model:url="store.request.url"
                          :meta="store.meta"
                          @update:url="handleRequestBaseChange"
                        />
                      </div>
                    </n-tab-pane>

                    <!-- Tab 3: Headers -->
                    <n-tab-pane name="Headers" tab="Headers">
                      <div class="py-2 h-full overflow-auto">
                        <HeadersEditor
                          v-model:modelValue="store.request.headers"
                          :meta="store.meta"
                          @update:modelValue="handleRequestBaseChange"
                        />
                      </div>
                    </n-tab-pane>

                    <!-- Tab 3: Body -->
                    <n-tab-pane name="Body" tab="Body">
                      <div class="flex flex-col h-full pt-2 gap-2">
                        <div class="flex justify-end px-1">
                          <n-button size="tiny" secondary type="info" @click="formatBody">
                            Format JSON
                          </n-button>
                        </div>
                        <div class="flex-1 min-h-0">
                          <CodeEditor
                            :model-value="store.request.body"
                            language="json"
                            height="100%"
                            @update:model-value="handleBodyChange"
                          />
                        </div>
                      </div>
                    </n-tab-pane>

                    <!-- Tab 4: Docs -->
                    <n-tab-pane v-if="store.request.method !== 'GET'" name="Docs" tab="Docs">
                      <div class="py-2 h-full overflow-auto">
                        <BodyDocsViewer 
                          :body-json="store.request.body" 
                          :param-docs="store.meta?.param_docs || {}" 
                        />
                      </div>
                    </n-tab-pane>

                    <!-- Tab 5: Raw Curl -->
                    <n-tab-pane name="Curl" tab="Curl">
                      <div class="h-full pt-2 flex flex-col gap-2">
                        <div class="flex justify-end px-1 gap-2">
                          <n-button size="tiny" secondary type="success" @click="copyRealCurl">
                            <template #icon>
                              <n-icon><SaveOutline /></n-icon>
                            </template>
                            Copy with Env
                          </n-button>
                        </div>
                        <n-alert v-if="possibleReplacement" type="info" show-icon class="mb-1">
                          Detected values matching environment variables.
                          <n-button size="tiny" type="primary" secondary @click="applyReplacement" class="ml-2">
                            Replace with {{ envStore.activeEnvName }}
                          </n-button>
                        </n-alert>
                        <CodeEditor
                          :model-value="store.curlCode"
                          language="shell"
                          height="100%"
                          @update:model-value="handleCurlChange"
                        />
                      </div>
                    </n-tab-pane>
                  </n-tabs>
                </div>
              </div>
            </div>

            <!-- Resizer Handle -->
            <div 
              v-if="store.request.method !== 'WS'"
              class="h-2 w-full hover:bg-blue-500/50 cursor-row-resize flex items-center justify-center group transition-colors my-1 shrink-0"
              @mousedown="startDrag"
            >
              <div class="w-10 h-1 rounded-full bg-gray-700 group-hover:bg-blue-400 transition-colors"></div>
            </div>

            <!-- Response Section (Bottom) -->
            <div v-if="store.request.method !== 'WS'" class="flex-1 min-h-0 overflow-hidden mt-1">
              <ResponsePanel class="h-full rounded-lg border border-gray-700/50" />
            </div>
          </div>
        </div>
      </template>
    </n-split>

    <!-- Modals -->
    <RequestMetaModal
      v-model:show="showMetaModal"
      :initial-data="store.meta"
      :current-file-name="store.currentFileName"
      :existing-tags="store.folderOptions"
      @save="handleMetaSave"
    />

    <EnvManager v-model:show="showEnvModal" />
    <SettingsModal v-model:show="showSettingsModal" />
  </div>
</template>

<style scoped>
:deep(.n-tabs-pane-wrapper) {
  flex: 1;
  height: 0;
}
:deep(.n-tab-pane) {
  height: 100%;
}
.split-pane {
  width: 100%;
}
</style>