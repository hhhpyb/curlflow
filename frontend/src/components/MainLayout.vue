<script setup lang="ts">
import {ref, watch, onMounted } from 'vue'
import { useMessage } from 'naive-ui'
import { PlayOutline, CloudDownloadOutline } from '@vicons/ionicons5'
// 注意路径变化：CodeEditor 和当前文件在同一个目录
import CodeEditor from './CodeEditor.vue'

// 注意路径变化：需要往上跳两级才能找到 wailsjs
import { ParseCurl, SendRequest } from '../../wailsjs/go/main/App'
import { main } from '../../wailsjs/go/models'

const message = useMessage()

const curlInput = ref('curl -X GET https://httpbin.org/get')
const currentRequest = ref(new main.HttpRequest())
const responseOutput = ref('{\n  "status": "ready"\n}')
const isLoading = ref(false)

const parseCurl = async (curl: string) => {
  if (!curl) return
  try {
    const res = await ParseCurl(curl)
    currentRequest.value = res
    console.log('curlInput',res)
  } catch (e) {
    console.error(e)
  }
}

watch(curlInput, (newVal) => {
  parseCurl(newVal)
})

onMounted(() => {
  parseCurl(curlInput.value)
})

const handleRun = async () => {
  console.log('handle run', currentRequest.value)
  if (!currentRequest.value.url) {
    message.warning("URL不能为空")
    return
  }

  isLoading.value = true
  responseOutput.value = "Sending..."

  try {
    const res = await SendRequest(currentRequest.value)

    if (res.error) {
      responseOutput.value = `Error: ${res.error}`
      message.error("请求失败")
    } else {
      try {
        const jsonObj = JSON.parse(res.body)
        responseOutput.value = JSON.stringify(jsonObj, null, 2)
      } catch {
        responseOutput.value = res.body
      }
      message.success(`Status: ${res.statusCode} | Time: ${res.time}ms`)
    }
  } catch (e) {
    responseOutput.value = `System Error: ${e}`
  } finally {
    isLoading.value = false
  }
}
</script>

<template>
  <div class="h-screen w-screen flex flex-col bg-gray-900 text-gray-200 p-4 gap-4">
    <div class="flex items-center justify-between">
      <div class="flex items-center gap-2">
        <n-icon size="24" color="#4ade80"><CloudDownloadOutline /></n-icon>
        <span class="font-bold text-lg tracking-wide">CurlFlow</span>
      </div>
      <n-button
          type="primary"
          size="small"
          :loading="isLoading"
          @click="handleRun"
      >
        <template #icon><n-icon><PlayOutline /></n-icon></template>
        Run
      </n-button>
    </div>

    <div class="flex-1 flex gap-4 min-h-0">
      <div class="flex-1 flex flex-col gap-2 min-w-0">
        <div class="text-xs font-mono text-gray-400 uppercase">Request (Curl)</div>
        <CodeEditor
            v-model="curlInput"
            language="shell"
            height="100%"
        />
      </div>

      <div class="flex-1 flex flex-col gap-2 min-w-0">
        <div class="text-xs font-mono text-gray-400 uppercase">Response</div>
        <CodeEditor
            v-model="responseOutput"
            language="json"
            :read-only="true"
            height="100%"
        />
      </div>
    </div>
  </div>
</template>