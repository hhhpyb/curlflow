<script setup lang="ts">
import { ref, computed } from 'vue'
import { NTag, NCode, NInput, NIcon } from 'naive-ui'
import { SearchOutline } from '@vicons/ionicons5'
import JsonSchemaTable from './JsonSchemaTable.vue'
import { domain } from '../../wailsjs/go/models'

const props = defineProps<{
  request: domain.RequestFile
}>()

const globalSearch = ref('')

const requestBody = computed(() => props.request.data.body || '{}')
const paramDocs = computed(() => props.request._meta.param_docs || {})

// Cast to any to access new fields until wailsjs is regenerated
const responseDocs = computed(() => (props.request._meta as any).response_docs || {})
const responseExample = computed(() => (props.request._meta as any).response_example || '{}')

const summary = computed(() => props.request._meta.summary)
const description = computed(() => props.request._meta.description)
const method = computed(() => props.request.data.method)
const url = computed(() => props.request.data.url)

</script>

<template>
  <div class="doc-preview h-full flex flex-col relative overflow-hidden">
    <!-- Sticky Search Header -->
    <div class="sticky top-0 z-10 bg-[#1e1e1e] p-4 pb-2 border-b border-gray-700/50">
      <n-input
        v-model:value="globalSearch"
        placeholder="Search in Request & Response params..."
        clearable
        size="medium"
      >
        <template #suffix>
          <n-icon :component="SearchOutline" />
        </template>
      </n-input>
    </div>

    <!-- Scrollable Content -->
    <div class="flex-1 overflow-y-auto p-4">
      <!-- Section 1: Basic Info -->
      <div class="mb-6">
        <div class="flex items-center gap-2 mb-2">
          <n-tag type="info" size="small">{{ method }}</n-tag>
          <div class="text-lg font-bold">{{ summary || 'No Summary' }}</div>
        </div>
        <div class="mb-2">
          <n-code :code="url" language="text" word-wrap />
        </div>
        <div v-if="description" class="text-gray-500 text-sm whitespace-pre-wrap">
          {{ description }}
        </div>
      </div>

      <!-- Section 2: Request Parameters -->
      <div class="mb-8">
        <h3 class="text-md font-bold mb-2 border-l-4 border-blue-500 pl-2">Request Body / Params</h3>
        <JsonSchemaTable 
          :json-content="requestBody" 
          :doc-map="paramDocs" 
          root-key="body" 
          :keyword="globalSearch"
        />
      </div>

      <!-- Section 3: Response Data -->
      <div class="mb-8">
        <h3 class="text-md font-bold mb-2 border-l-4 border-green-500 pl-2">Response Data</h3>
        <JsonSchemaTable 
          :json-content="responseExample" 
          :doc-map="responseDocs" 
          root-key="data" 
          :keyword="globalSearch"
        />
        
        <!-- Optional: Show Raw JSON Example -->
        <div v-if="responseExample && responseExample !== '{}'" class="mt-4">
          <div class="text-xs text-gray-500 mb-1">Response Example (Raw JSON):</div>
          <n-code 
            :code="responseExample" 
            language="json" 
          />
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.doc-preview {
  scroll-behavior: smooth;
}
</style>