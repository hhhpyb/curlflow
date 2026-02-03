<script setup lang="ts">
import {computed, ref} from 'vue'
import {useMessage, NTabs, NTabPane, NDynamicInput} from 'naive-ui'
import {CloudDownloadOutline, PlayOutline} from '@vicons/ionicons5'
import CodeEditor from './CodeEditor.vue'
import RequestPanel from './RequestPanel.vue'
import {useRequestStore} from '../stores/request'

const message = useMessage()
const store = useRequestStore()
const activeTab = ref('body')

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
  <div class="h-screen w-screen flex flex-col bg-gray-900 text-gray-200 p-4 gap-4 overflow-hidden">
    <div class="flex items-center justify-between">
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

    <div class="flex-1 flex gap-4 min-h-0">
      <!-- Request Section -->
      <div class="flex-1 flex flex-col gap-2 min-w-0">
        <div class="text-xs font-bold font-mono text-gray-500 uppercase tracking-widest flex items-center gap-2">
          <div class="w-1.5 h-1.5 rounded-full bg-blue-500"></div>
          Request
        </div>
        <RequestPanel/>
      </div>

      <!-- Response Section -->
      <div class="flex-1 flex flex-col gap-2 min-w-0">
        <div class="text-xs font-bold font-mono text-gray-500 uppercase tracking-widest flex items-center gap-2">
          <div class="w-1.5 h-1.5 rounded-full bg-green-500"></div>
          Response
          <span v-if="store.response.statusCode" class="ml-auto normal-case text-gray-400">
            Status: <span :class="store.response.statusCode < 400 ? 'text-green-400' : 'text-red-400'">{{ store.response.statusCode }}</span>
            <span class="mx-2 text-gray-700">|</span>
            Time: <span class="text-blue-400">{{ store.response.time }}ms</span>
          </span>
        </div>
        
        <div class="flex-1 flex flex-col bg-gray-800/50 rounded-lg border border-gray-700/50 p-1">
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
