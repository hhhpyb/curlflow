<script setup lang="ts">
import { ref, watch } from 'vue'
import { useRequestStore } from '../stores/request'
import CodeEditor from './CodeEditor.vue'
import { 
  NTabs, 
  NTabPane, 
  NInput, 
  NInputGroup, 
  NSelect, 
  NDynamicInput, 
  NFormItem 
} from 'naive-ui'

const store = useRequestStore()

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
// Naive UI Dynamic Input works best with arrays, store has map.
const headersList = ref<{ key: string; value: string }[]>([])

// 1. Sync Store -> Local UI (When ParseCurl happens)
watch(
  () => store.request.headers,
  (newHeaders) => {
    // Avoid re-mapping if we are the ones who triggered the change (optimization)
    // But for safety and simplicity, we map.
    const list: { key: string; value: string }[] = []
    if (newHeaders) {
      for (const k in newHeaders) {
        list.push({ key: k, value: newHeaders[k] })
      }
    }
    headersList.value = list
  },
  { deep: true, immediate: true }
)

// 2. Sync Local UI -> Store (When User edits headers)
const handleHeadersChange = (val: any) => {
  const map: Record<string, string> = {}
  if (Array.isArray(val)) {
    val.forEach((item: any) => {
      if (item.key) map[item.key] = item.value
    })
  }
  store.request.headers = map
  store.syncToCurl()
}

// ================= Event Handlers =================

// When Method or URL changes
const handleRequestBaseChange = () => {
  store.syncToCurl()
}

// When Curl Code changes
const handleCurlChange = (val: string) => {
  store.curlCode = val
  store.syncFromCurl()
}

// When Body changes
const handleBodyChange = (val: string) => {
  store.request.body = val
  store.syncToCurl()
}
</script>

<style scoped>
/* 强制 n-tabs 的内容区域撑满高度 */
:deep(.n-tabs-pane-wrapper) {
  height: 100%;
}
:deep(.n-tab-pane) {
  height: 100%;
}
</style>

<template>
  <div class="flex flex-col h-full gap-4">
    <!-- Top Bar: Method & URL -->
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
      />
    </n-input-group>

    <!-- Tabs Area -->
    <n-tabs type="line" animated class="flex-1 h-0">
      <!-- Tab 1: Raw Curl -->
      <n-tab-pane name="curl" tab="Raw Curl" display-directive="show">
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
      <n-tab-pane name="headers" tab="Headers" display-directive="show">
        <div class="h-full pt-2 overflow-auto">
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
      <n-tab-pane name="body" tab="Body" display-directive="show">
        <div class="h-full pt-2">
          <CodeEditor
            :model-value="store.request.body"
            language="json"
            height="100%"
            @update:model-value="handleBodyChange"
          />
        </div>
      </n-tab-pane>
    </n-tabs>
  </div>
</template>
