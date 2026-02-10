<script setup lang="ts">
import { computed, h, ref } from 'vue'
import { useRequestStore } from '../stores/request'
import { 
    NTabs, NTabPane, NTag, NSpin, NEmpty, NDataTable, 
    NButton, NIcon, useMessage, NModal, NCard 
} from 'naive-ui'
import { CopyOutline, CloudOfflineOutline, InformationCircleOutline } from '@vicons/ionicons5'
import CodeEditor from './CodeEditor.vue'

const store = useRequestStore()
const message = useMessage()
const showErrorDetail = ref(false)

interface Props {
    previewContent?: string
}
const props = defineProps<Props>()

// --- Computed Properties ---

const isShowingPreview = computed(() => !!props.previewContent)

const hasResponse = computed(() => {
    if (isShowingPreview.value) return true
    return store.response && store.response.statusCode > 0
})

const hasError = computed(() => {
    if (isShowingPreview.value) return false
    return !!store.response.error
})

const statusType = computed(() => {
    if (isShowingPreview.value) return 'info'
    const code = store.response.statusCode
    if (code >= 200 && code < 300) return 'success'
    if (code >= 300 && code < 400) return 'warning'
    return 'error'
})

const responseSize = computed(() => {
    const len = displayedBody.value ? displayedBody.value.length : 0
    if (len < 1024) return `${len} B`
    return `${(len / 1024).toFixed(2)} KB`
})

const headersList = computed(() => {
    if (isShowingPreview.value) {
        return [{ key: 'Content-Type', value: 'application/json (Example)' }]
    }
    if (!store.response.headers) return []
    return Object.entries(store.response.headers).map(([key, value]) => ({
        key,
        value
    }))
})

// Auto-detect language for syntax highlighting
const editorLanguage = computed(() => {
    if (isShowingPreview.value) return 'json'
    if (!store.response.headers) return 'json'
    // Case-insensitive header lookup
    const keys = Object.keys(store.response.headers)
    const typeKey = keys.find(k => k.toLowerCase() === 'content-type')
    const ct = typeKey ? store.response.headers[typeKey] : ''
    
    if (ct.includes('html')) return 'html'
    if (ct.includes('xml')) return 'xml'
    return 'json'
})

// The actual body to display (Prioritize previewContent)
const displayedBody = computed(() => {
    if (isShowingPreview.value) return props.previewContent || ''
    return store.response.body || ''
})

// Formatting Body (Prettify JSON if applicable)
const formattedBody = computed(() => {
    const body = displayedBody.value
    if (!body) return ''
    if (editorLanguage.value === 'json') {
        try {
            return JSON.stringify(JSON.parse(body), null, 2)
        } catch {
            return body
        }
    }
    return body
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
            <template v-if="isShowingPreview">
                <n-tag type="info" size="small" :bordered="false" class="font-mono font-bold">
                    EXAMPLE PREVIEW
                </n-tag>
                <span class="text-xs text-gray-500 italic">Showing documentation example</span>
            </template>
            <template v-else>
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
        </template>
        
        <!-- Error State -->
        <template v-else-if="hasError">
             <n-tag type="error" size="small" class="shrink-0">Error</n-tag>
             <div class="flex items-center gap-2 min-w-0 flex-1">
                 <span class="text-xs text-red-400 truncate" :title="store.response.error">{{ store.response.error }}</span>
                 <n-button text size="tiny" class="text-red-400 hover:text-red-300 shrink-0" @click="showErrorDetail = true">
                    <n-icon size="16"><InformationCircleOutline /></n-icon>
                 </n-button>
             </div>
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

    <!-- Error Detail Modal -->
    <n-modal v-model:show="showErrorDetail">
        <n-card
            style="width: 600px; max-width: 90vw;"
            title="Error Details"
            :bordered="false"
            size="large"
            role="dialog"
            aria-modal="true"
        >
            <div class="bg-gray-100 dark:bg-gray-800 p-4 rounded border border-gray-200 dark:border-gray-700">
                <pre class="text-xs text-red-600 dark:text-red-400 whitespace-pre-wrap break-all font-mono">{{ store.response.error }}</pre>
            </div>
            <div class="flex justify-end mt-4">
                <n-button @click="showErrorDetail = false">Close</n-button>
            </div>
        </n-card>
    </n-modal>
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