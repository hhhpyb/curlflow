<script setup lang="ts">
import {computed, onUnmounted, ref} from 'vue'
import {useMessage, NTabs, NTabPane, NDynamicInput} from 'naive-ui'
import {CloudDownloadOutline, PlayOutline} from '@vicons/ionicons5'
import CodeEditor from './CodeEditor.vue'
import RequestPanel from './RequestPanel.vue'
import Sidebar from './Sidebar.vue'
import {useRequestStore} from '../stores/request'

const message = useMessage()
const store = useRequestStore()
const activeTab = ref('body')

// Split pane logic
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
    activeTab.value = 'body'
  }
}

const responseOutput = computed(() => {
  if (store.response.error) {
    return `Error: ${store.response.error}`
  }
  // Check if response is empty (initial state)
  if (!store.response.body && !store.response.statusCode) {
    return '{\n  "status": "ready"\n}'
  }
  try {
    const jsonObj = JSON.parse(store.response.body)
    return JSON.stringify(jsonObj, null, 2)
  } catch {
    return store.response.body
  }
})

const responseHeaders = computed(() => {
  const list: { key: string; value: string }[] = []
  if (store.response.headers) {
    for (const [k, v] of Object.entries(store.response.headers)) {
      list.push({ key: k, value: v })
    }
  }
  return list
})
</script>

<template>
  <div class="h-screen w-screen flex bg-gray-900 text-gray-200 overflow-hidden">
    <Sidebar />
    
    <div class="flex-1 flex flex-col p-4 gap-4 min-w-0">
      <!-- Header -->
      <div class="flex items-center justify-between shrink-0">
        <div class="flex items-center gap-2">
          <n-icon size="24" color="#4ade80">
            <CloudDownloadOutline/>
          </n-icon>
          <span class="font-bold text-lg tracking-wide text-white">CurlFlow</span>
        </div>
        <n-button
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
      </div>

      <!-- Main Content Area -->
      <div class="flex-1 flex flex-col min-h-0 relative" ref="containerRef">
        <!-- Request Section (Top) -->
        <div class="flex flex-col gap-2 min-h-0" :style="{ height: `${requestHeightPercent}%` }">
          <div class="text-xs font-bold font-mono text-gray-500 uppercase tracking-widest flex items-center gap-2 shrink-0">
            <div class="w-1.5 h-1.5 rounded-full bg-blue-500"></div>
            Request
          </div>
          <RequestPanel class="flex-1 min-h-0"/>
        </div>

        <!-- Resizer Handle -->
        <div 
          class="h-2 w-full hover:bg-blue-500/50 cursor-row-resize flex items-center justify-center group transition-colors my-1 shrink-0"
          @mousedown="startDrag"
        >
          <div class="w-10 h-1 rounded-full bg-gray-700 group-hover:bg-blue-400 transition-colors"></div>
        </div>

        <!-- Response Section (Bottom) -->
        <div class="flex-1 flex flex-col gap-2 min-h-0 overflow-hidden">
          <div class="text-xs font-bold font-mono text-gray-500 uppercase tracking-widest flex items-center gap-2 shrink-0">
            <div class="w-1.5 h-1.5 rounded-full bg-green-500"></div>
            Response
            <span v-if="store.response.statusCode" class="ml-auto normal-case text-gray-400">
              Status: <span :class="store.response.statusCode < 400 ? 'text-green-400' : 'text-red-400'">{{ store.response.statusCode }}</span>
              <span class="mx-2 text-gray-700">|</span>
              Time: <span class="text-blue-400">{{ store.response.time }}ms</span>
            </span>
          </div>
          
          <div class="flex-1 flex flex-col bg-gray-800/50 rounded-lg border border-gray-700/50 p-1 min-h-0">
            <n-tabs v-model:value="activeTab" type="line" animated class="h-full flex flex-col">
              <n-tab-pane name="body" tab="Body" class="flex-1 h-0">
                <div class="h-full pt-2">
                  <CodeEditor
                      :model-value="responseOutput"
                      language="json"
                      :read-only="true"
                      height="100%"
                  />
                </div>
              </n-tab-pane>
              <n-tab-pane name="headers" tab="Headers" class="flex-1 h-0">
                <div class="h-full pt-2 overflow-auto px-2">
                  <n-dynamic-input
                    :value="responseHeaders"
                    preset="pair"
                    key-placeholder="Header Name"
                    value-placeholder="Header Value"
                    :show-button="false"
                  >
                    <template #default="{ value }">
                      <div class="flex gap-2 w-full mb-2">
                        <n-input :value="value.key" readonly placeholder="Key" class="flex-1" />
                        <n-input :value="value.value" readonly placeholder="Value" class="flex-2" />
                      </div>
                    </template>
                  </n-dynamic-input>
                  <div v-if="responseHeaders.length === 0" class="text-center py-10 text-gray-500 italic">
                    No headers received
                  </div>
                </div>
              </n-tab-pane>
            </n-tabs>
          </div>
        </div>
      </div>
    </div>
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
</style>
