<script setup lang="ts">
import { h, ref, watch, computed } from 'vue'
import { NDataTable, NTag, NInput, NIcon } from 'naive-ui'
import { SearchOutline } from '@vicons/ionicons5'
import type { DataTableColumns } from 'naive-ui'

interface Props {
  bodyJson: string
  paramDocs?: Record<string, string>
}

const props = withDefaults(defineProps<Props>(), {
  bodyJson: '',
  paramDocs: () => ({})
})

interface RowData {
  key: string
  field: string
  type: string
  description: string
  children?: RowData[]
}

const tableData = ref<RowData[]>([])
const searchKeyword = ref('')

const columns: DataTableColumns<RowData> = [
  {
    title: 'Field',
    key: 'field',
    width: 200
  },
  {
    title: 'Type',
    key: 'type',
    width: 120,
    render(row) {
      return h(
        NTag,
        { size: 'small', type: 'info', bordered: false },
        { default: () => row.type }
      )
    }
  },
  {
    title: 'Description',
    key: 'description',
    ellipsis: {
      tooltip: true
    }
  }
]

const getType = (val: any): string => {
  if (val === null) return 'Null'
  if (Array.isArray(val)) return 'Array'
  const t = typeof val
  return t.charAt(0).toUpperCase() + t.slice(1)
}

const buildTree = (data: any, parentPath: string = ''): RowData[] => {
  if (data === null || typeof data !== 'object') {
    return []
  }

  const rows: RowData[] = []
  
  if (Array.isArray(data)) {
    // Array handling: only process the first element as a structure example
    if (data.length > 0) {
      const index = 0
      const val = data[index]
      const currentPath = parentPath ? `${parentPath}.${index}` : `${index}`
      const docKey = `body.${currentPath}`
      const description = props.paramDocs?.[docKey] || ''

      const row: RowData = {
        key: currentPath,
        field: '[0]', // Indicate this is the first item of an array
        type: getType(val),
        description
      }

      if (val && typeof val === 'object') {
        const children = buildTree(val, currentPath)
        if (children.length > 0) {
          row.children = children
        }
      }
      rows.push(row)
    }
  } else {
    // Object handling
    for (const key of Object.keys(data)) {
      const val = data[key]
      const currentPath = parentPath ? `${parentPath}.${key}` : key
      const docKey = `body.${currentPath}`
      const description = props.paramDocs?.[docKey] || ''

      const row: RowData = {
        key: currentPath,
        field: key,
        type: getType(val),
        description
      }

      if (val && typeof val === 'object') {
        const children = buildTree(val, currentPath)
        if (children.length > 0) {
          row.children = children
        }
      }
      rows.push(row)
    }
  }

  return rows
}

const filterTree = (nodes: RowData[], keyword: string): RowData[] => {
  if (!keyword) return nodes
  const lowerKeyword = keyword.toLowerCase()

  const traverse = (node: RowData): RowData | null => {
    const isMatch = 
      node.field.toLowerCase().includes(lowerKeyword) || 
      node.description.toLowerCase().includes(lowerKeyword)
    
    let children: RowData[] = []
    if (node.children) {
      children = node.children
        .map(n => traverse(n))
        .filter((n): n is RowData => n !== null)
    }

    if (isMatch || children.length > 0) {
      return {
        ...node,
        children: children.length > 0 ? children : undefined
      }
    }
    return null
  }

  return nodes
    .map(n => traverse(n))
    .filter((n): n is RowData => n !== null)
}

const filteredTableData = computed(() => {
  return filterTree(tableData.value, searchKeyword.value)
})

watch(
  [() => props.bodyJson, () => props.paramDocs],
  () => {
    try {
      if (!props.bodyJson.trim()) {
        tableData.value = []
        return
      }
      const parsed = JSON.parse(props.bodyJson)
      tableData.value = buildTree(parsed, '')
    } catch (e) {
      // Quietly fail or empty list on invalid JSON
      tableData.value = []
    }
  },
  { immediate: true }
)
</script>

<template>
  <div class="body-docs-viewer h-full flex flex-col">
    <div class="mb-2">
      <n-input
        v-model:value="searchKeyword"
        placeholder="Search field or description..."
        clearable
        size="small"
      >
        <template #suffix>
          <n-icon :component="SearchOutline" />
        </template>
      </n-input>
    </div>
    <n-data-table
      :columns="columns"
      :data="filteredTableData"
      :bordered="false"
      size="small"
      default-expand-all
      :row-key="(row: RowData) => row.key"
      flex-height
      class="flex-1"
      :style="{ height: '100%' }"
    />
  </div>
</template>