<script setup lang="ts">
import { ref, watch, nextTick } from 'vue';
import { NTable, NInput, NButton, NIcon } from 'naive-ui';
import { TrashOutline } from '@vicons/ionicons5';

interface HeaderRow {
  key: string;
  value: string;
  id: string;
}

const props = defineProps<{
  modelValue: Record<string, string>;
  meta: any | null; // domain.MetaData
}>();

const emit = defineEmits<{
  (e: 'update:modelValue', value: Record<string, string>): void;
}>();

const headersList = ref<HeaderRow[]>([]);
const isInternalChange = ref(false);

// 生成简易唯一标识
const generateId = () => Math.random().toString(36).substring(2, 9);

// 确保列表末尾始终有一个空行
const ensureEmptyRow = () => {
  if (
    headersList.value.length === 0 || 
    headersList.value[headersList.value.length - 1].key !== '' || 
    headersList.value[headersList.value.length - 1].value !== ''
  ) {
    headersList.value.push({ key: '', value: '', id: generateId() });
  }
};

// Object -> Array (Prop to List)
const syncPropToList = (val: Record<string, string>) => {
  const newList: HeaderRow[] = [];
  if (val) {
    Object.entries(val).forEach(([key, value]) => {
      newList.push({ key, value, id: generateId() });
    });
  }
  
  // 比较内容，避免不必要的更新导致的焦点丢失
  const currentSimplified = headersList.value
    .filter(h => h.key !== '')
    .map(h => `${h.key}:${h.value}`)
    .join('|');
  const nextSimplified = Object.entries(val || {})
    .map(([k, v]) => `${k}:${v}`)
    .join('|');

  if (currentSimplified !== nextSimplified) {
    headersList.value = newList;
    ensureEmptyRow();
  }
};

// Array -> Object (List to Prop)
const syncListToProp = () => {
  const map: Record<string, string> = {};
  headersList.value.forEach(row => {
    if (row.key.trim()) {
      map[row.key] = row.value;
    }
  });
  emit('update:modelValue', map);
};

// 监听外部变化
watch(
  () => props.modelValue,
  (newVal) => {
    if (isInternalChange.value) return;
    syncPropToList(newVal);
  },
  { immediate: true, deep: true }
);

// 监听内部表格变化
watch(
  headersList,
  () => {
    isInternalChange.value = true;
    syncListToProp();
    ensureEmptyRow();
    
    nextTick(() => {
      isInternalChange.value = false;
    });
  },
  { deep: true }
);

const removeRow = (index: number) => {
  if (index === headersList.value.length - 1) return;
  headersList.value.splice(index, 1);
};
</script>

<template>
  <div class="headers-editor">
    <n-table size="small" :bordered="false" :single-line="false" class="bg-transparent">
      <thead>
        <tr>
          <th class="w-1/3 bg-gray-800/50">Header Key</th>
          <th class="w-1/3 bg-gray-800/50">Value</th>
          <th class="bg-gray-800/50">Description</th>
          <th class="w-12 bg-gray-800/50"></th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="(row, index) in headersList" :key="row.id">
          <td class="p-0">
            <n-input
              v-model:value="row.key"
              placeholder="Header Key (e.g. Content-Type)"
              size="small"
              :bordered="false"
              class="bg-transparent font-mono text-green-400"
            />
          </td>
          <td class="p-0">
            <n-input
              v-model:value="row.value"
              placeholder="Value"
              size="small"
              :bordered="false"
              class="bg-transparent"
            />
          </td>
          <td class="align-middle text-gray-500 text-xs px-3">
            {{ props.meta?.param_docs?.['header.' + row.key] || '-' }}
          </td>
          <td class="text-center p-0">
            <n-button
              v-if="index !== headersList.length - 1"
              text
              size="tiny"
              type="error"
              @click="removeRow(index)"
            >
              <template #icon>
                <n-icon :component="TrashOutline" />
              </template>
            </n-button>
          </td>
        </tr>
      </tbody>
    </n-table>
  </div>
</template>

<style scoped>
:deep(.n-input) {
  --n-bezier: none !important;
  --n-border: none !important;
  --n-border-hover: none !important;
  --n-border-focus: none !important;
  --n-box-shadow-focus: none !important;
}

:deep(.n-input .n-input__border),
:deep(.n-input .n-input__state-border) {
  display: none;
}

tr:hover td {
  background-color: rgba(255, 255, 255, 0.03) !important;
}
</style>
