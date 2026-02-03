<script setup lang="ts">
import { ref, watch, nextTick } from 'vue'
import { useRequestStore } from '../stores/request'
import CodeEditor from './CodeEditor.vue'
import { 
  NTabs, 
  NTabPane, 
  NInput, 
  NInputGroup, 
  NSelect, 
  NDynamicInput,
  NIcon,
  NButton,
  useMessage
} from 'naive-ui'
import { DocumentTextOutline, ListOutline, CodeSlashOutline, OptionsOutline } from '@vicons/ionicons5'

const store = useRequestStore()
const message = useMessage()

const methodOptions = [
  { label: 'GET', value: 'GET' },
  { label: 'POST', value: 'POST' },
  { label: 'PUT', value: 'PUT' },
  { label: 'DELETE', value: 'DELETE' },
  { label: 'PATCH', value: 'PATCH' },
  { label: 'HEAD', value: 'HEAD' },
  { label: 'OPTIONS', value: 'OPTIONS' }
]

// ================= Headers Logic =================
const headersList = ref<{ key: string; value: string }[]>([])
const ignoreHeaderUpdate = ref(false)

// Sync Store -> Local UI (When ParseCurl or manual edits happen)
watch(
  () => store.request.headers,
  (newHeaders) => {
    if (ignoreHeaderUpdate.value) return
    const list: { key: string; value: string }[] = []
    if (newHeaders) {
      Object.entries(newHeaders).forEach(([key, value]) => {
        list.push({ key, value })
      })
    }
    // Only update if different to avoid cursor jumps or infinite loops
    if (JSON.stringify(list) !== JSON.stringify(headersList.value)) {
      headersList.value = list
    }
  },
  { deep: true, immediate: true }
)

// Sync Local UI -> Store (When User edits headers)
const handleHeadersChange = (val: any) => {
  ignoreHeaderUpdate.value = true
  const map: Record<string, string> = {}
  if (Array.isArray(val)) {
    val.forEach((item: any) => {
      if (item.key) map[item.key] = item.value
    })
  }
  store.request.headers = map
  store.syncToCurl()
  nextTick(() => {
    ignoreHeaderUpdate.value = false
  })
}

// ================= Body Logic =================
const requestBody = ref(store.request.body)
watch(() => store.request.body, (newBody) => {
  requestBody.value = newBody
})

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

// ================= Base Info =================
const handleRequestBaseChange = () => {
  store.syncToCurl()
}

const handleCurlChange = (val: string) => {
  store.curlCode = val
  store.syncFromCurl()
}
</script>

<style scoped>
:deep(.n-tabs-pane-wrapper) {
  flex: 1;
  height: 0;
}
:deep(.n-tab-pane) {
  height: 100%;
}
</style>

<template>
  <div class="flex flex-col h-full gap-4">
    <!-- Top Bar: Method & URL -->
    <div class="bg-gray-800/30 p-2 rounded-lg border border-gray-700/50">
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

    <!-- Tabs Area -->
    <div class="flex-1 flex flex-col bg-gray-800/50 rounded-lg border border-gray-700/50 p-1 min-h-0">
      <n-tabs type="line" animated class="h-full flex flex-col">
        <!-- Tab 1: Raw Curl -->
        <n-tab-pane name="curl" display-directive="show">
          <template #tab>
            <div class="flex items-center gap-1.5">
              <n-icon><CodeSlashOutline /></n-icon>
              Raw Curl
            </div>
          </template>
          <div class="h-full pt-2">
            <CodeEditor
              :model-value="store.curlCode"
              language="shell"
              height="100%"
              @update:model-value="handleCurlChange"
            />
          </div>
        </n-tab-pane>

        <!-- Tab 2: Headers -->
        <n-tab-pane name="headers" display-directive="show">
          <template #tab>
            <div class="flex items-center gap-1.5">
              <n-icon><ListOutline /></n-icon>
              Headers
            </div>
          </template>
          <div class="h-full pt-2 overflow-auto px-2">
            <n-dynamic-input
              v-model:value="headersList"
              preset="pair"
              key-placeholder="Header Name"
              value-placeholder="Header Value"
              @update:value="handleHeadersChange"
            />
          </div>
        </n-tab-pane>

        <!-- Tab 3: Body -->
        <n-tab-pane name="body" display-directive="show">
          <template #tab>
            <div class="flex items-center gap-1.5">
              <n-icon><DocumentTextOutline /></n-icon>
              Body
            </div>
          </template>
          <div class="flex flex-col h-full pt-2 gap-2">
            <div class="flex justify-end px-1">
              <n-button size="tiny" secondary type="info" @click="formatBody">
                <template #icon>
                  <n-icon><OptionsOutline /></n-icon>
                </template>
                Format JSON
              </n-button>
            </div>
            <div class="flex-1 min-h-0">
              <CodeEditor
                :model-value="requestBody"
                language="json"
                height="100%"
                @update:model-value="handleBodyChange"
              />
            </div>
          </div>
        </n-tab-pane>
      </n-tabs>
    </div>
  </div>
</template>
