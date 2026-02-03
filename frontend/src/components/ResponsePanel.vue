<script setup lang="ts">
import { computed, h } from 'vue'
import { useRequestStore } from '../stores/request'
import { 
    NTabs, NTabPane, NTag, NSpin, NEmpty, NDataTable, 
    NButton, NIcon, useMessage 
} from 'naive-ui'
import { CopyOutline, CloudOfflineOutline } from '@vicons/ionicons5'
import CodeEditor from './CodeEditor.vue'

const store = useRequestStore()
const message = useMessage()

// --- Computed Properties ---

const hasResponse = computed(() => {
    return store.response && store.response.statusCode > 0
})

const hasError = computed(() => {
    return !!store.response.error
})

const statusType = computed(() => {
    const code = store.response.statusCode
    if (code >= 200 && code < 300) return 'success'
    if (code >= 300 && code < 400) return 'warning'
    return 'error'
})

const responseSize = computed(() => {
    const len = store.response.body ? store.response.body.length : 0
    if (len < 1024) return `${len} B`
    return `${(len / 1024).toFixed(2)} KB`
})

const headersList = computed(() => {
    if (!store.response.headers) return []
    return Object.entries(store.response.headers).map(([key, value]) => ({
        key,
        value
    }))
})

// Auto-detect language for syntax highlighting
const editorLanguage = computed(() => {
    if (!store.response.headers) return 'json'
    // Case-insensitive header lookup
    const keys = Object.keys(store.response.headers)
    const typeKey = keys.find(k => k.toLowerCase() === 'content-type')
    const ct = typeKey ? store.response.headers[typeKey] : ''
    
    if (ct.includes('html')) return 'html'
    if (ct.includes('xml')) return 'xml'
    return 'json'
})

// Formatting Body (Prettify JSON if applicable)
const formattedBody = computed(() => {
    if (!store.response.body) return ''
    if (editorLanguage.value === 'json') {
        try {
            return JSON.stringify(JSON.parse(store.response.body), null, 2)
        } catch {
            return store.response.body
        }
    }
    return store.response.body
})

// --- Actions ---

const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text).then(() => {
        message.success('Copied')
    }).catch(() => {
        message.error('Failed to copy')
    })
}

// --- Table Config ---

const headerColumns = [
    { title: 'Name', key: 'key', width: 200, ellipsis: true },
    { 
        title: 'Value', 
        key: 'value',
        render(row: any) {
            return h('div', { class: 'flex items-center justify-between group gap-2' }, [
                h('span', { class: 'truncate font-mono text-xs text-gray-300' }, row.value),
                h(NButton, {
                    size: 'tiny',
                    quaternary: true,
                    circle: true,
                    class: 'opacity-0 group-hover:opacity-100 transition-opacity shrink-0',
                    onClick: () => copyToClipboard(row.value)
                }, { icon: () => h(NIcon, null, { default: () => h(CopyOutline) }) })
            ])
        }
    }
]
</script>

<template>
  <div class="flex flex-col h-full w-full bg-gray-900 border-l border-gray-800">
    <!-- Status Bar -->
    <div class="flex items-center gap-4 px-4 py-2 bg-gray-800/50 border-b border-gray-800 shrink-0 h-10">
        <!-- Loading State -->
        <template v-if="store.isLoading">
            <n-spin size="small" stroke="#4ade80" />
            <span class="text-xs text-gray-400 animate-pulse">Sending Request...</span>
        </template>
        
        <!-- Response Data State -->
        <template v-else-if="hasResponse">
            <n-tag :type="statusType" size="small" :bordered="false" class="font-mono font-bold">
                {{ store.response.statusCode }}
            </n-tag>
            <div class="flex items-center gap-3 text-xs text-gray-400">
                <span class="flex items-center gap-1">
                    <span class="text-gray-500">Time:</span>
                    <span class="text-green-400 font-mono">{{ store.response.time }}ms</span>
                </span>
                <span class="w-[1px] h-3 bg-gray-700"></span>
                <span class="flex items-center gap-1">
                    <span class="text-gray-500">Size:</span>
                    <span class="text-blue-400 font-mono">{{ responseSize }}</span>
                </span>
            </div>
        </template>
        
        <!-- Error State -->
        <template v-else-if="hasError">
             <n-tag type="error" size="small">Error</n-tag>
             <span class="text-xs text-red-400 truncate">{{ store.response.error }}</span>
        </template>
        
        <!-- Idle State -->
        <template v-else>
             <div class="flex items-center gap-2 text-gray-600">
                <div class="w-2 h-2 rounded-full bg-gray-600"></div>
                <span class="text-xs font-medium uppercase tracking-wider">Ready</span>
             </div>
        </template>
    </div>

    <!-- Content Area -->
    <div class="flex-1 min-h-0 relative">
        <!-- Empty Placeholder -->
        <div v-if="!hasResponse && !store.isLoading && !hasError" class="absolute inset-0 flex items-center justify-center text-gray-700">
            <n-empty description="No Response">
                <template #icon>
                    <n-icon :component="CloudOfflineOutline" />
                </template>
            </n-empty>
        </div>

        <!-- Tabs -->
        <n-tabs v-else type="line" animated class="h-full flex flex-col custom-tabs">
            <n-tab-pane name="body" tab="Body" class="h-full p-0 flex-1">
                <CodeEditor 
                    :model-value="formattedBody"
                    :language="editorLanguage"
                    read-only
                    height="100%"
                />
            </n-tab-pane>
            <n-tab-pane name="headers" tab="Headers" class="h-full flex-1 overflow-hidden">
                <div class="h-full overflow-auto">
                    <n-data-table
                        :columns="headerColumns"
                        :data="headersList"
                        :bordered="false"
                        size="small"
                        class="bg-transparent"
                    />
                </div>
            </n-tab-pane>
        </n-tabs>
    </div>
  </div>
</template>

<style scoped>
/* Ensure tabs take full height */
:deep(.n-tabs-pane-wrapper) {
    flex: 1;
    overflow: hidden;
}
:deep(.n-tabs-nav) {
    background-color: rgba(31, 41, 55, 0.3); /* gray-800/30 */
    padding-left: 1rem;
}
:deep(.n-data-table .n-data-table-td) {
    background-color: transparent;
    color: #e5e7eb;
}
:deep(.n-data-table .n-data-table-th) {
    background-color: rgba(31, 41, 55, 0.5);
    color: #9ca3af;
}
:deep(.n-data-table:hover .n-data-table-td) {
    background-color: transparent !important;
}
</style>