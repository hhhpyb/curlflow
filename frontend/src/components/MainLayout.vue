<script setup lang="ts">
import {computed} from 'vue'
import {useMessage} from 'naive-ui'
import {CloudDownloadOutline, PlayOutline} from '@vicons/ionicons5'
import CodeEditor from './CodeEditor.vue'
import RequestPanel from './RequestPanel.vue'
import {useRequestStore} from '../stores/request'

const message = useMessage()
const store = useRequestStore()

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
</script>

<template>
  <div class="h-screen w-screen flex flex-col bg-gray-900 text-gray-200 p-4 gap-4">
    <div class="flex items-center justify-between">
      <div class="flex items-center gap-2">
        <n-icon size="24" color="#4ade80">
          <CloudDownloadOutline/>
        </n-icon>
        <span class="font-bold text-lg tracking-wide">CurlFlow</span>
      </div>
      <n-button
          type="primary"
          size="small"
          :loading="store.isLoading"
          @click="handleRun"
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
      <div class="flex-1 flex flex-col gap-2 min-w-0">
        <div class="text-xs font-mono text-gray-400 uppercase">Request</div>
        <RequestPanel/>
      </div>

      <div class="flex-1 flex flex-col gap-2 min-w-0">
        <div class="text-xs font-mono text-gray-400 uppercase">Response</div>
        <CodeEditor
            :model-value="responseOutput"
            language="json"
            :read-only="true"
            height="100%"
        />
      </div>
    </div>
  </div>
</template>
