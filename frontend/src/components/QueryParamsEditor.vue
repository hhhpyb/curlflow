<script setup lang="ts">
import { ref, watch, nextTick } from 'vue';
import { NTable, NInput, NButton, NIcon } from 'naive-ui';
import { TrashOutline } from '@vicons/ionicons5';
import { storage } from '../../wailsjs/go/models';

interface ParamRow {
  key: string;
  value: string;
  id: string;
}

const props = defineProps<{
  url: string;
  meta: any | null; // domain.MetaData
}>();

const emit = defineEmits<{
  (e: 'update:url', value: string): void;
}>();

const params = ref<ParamRow[]>([]);
const isInternalChange = ref(false);

// 生成简易唯一标识
const generateId = () => Math.random().toString(36).substring(2, 9);

// 确保列表末尾始终有一个空行
const ensureEmptyRow = () => {
  if (params.value.length === 0 || params.value[params.value.length - 1].key !== '' || params.value[params.value.length - 1].value !== '') {
    params.value.push({ key: '', value: '', id: generateId() });
  }
};

// 解析 URL -> Table
const parseUrlToParams = (url: string) => {
  if (!url) {
    params.value = [];
    ensureEmptyRow();
    return;
  }

  const parts = url.split('?');
  if (parts.length < 2) {
    params.value = [];
    ensureEmptyRow();
    return;
  }

  const queryString = parts[1];
  const searchParams = new URLSearchParams(queryString);
  const newParams: ParamRow[] = [];

  searchParams.forEach((value, key) => {
    newParams.push({ key, value, id: generateId() });
  });

  params.value = newParams;
  ensureEmptyRow();
};

// 构建 Table -> URL
const buildUrlFromParams = () => {
  const basePath = props.url.split('?')[0] || '';
  if (!basePath && !props.url) return;

  const validParams = params.value.filter(p => p.key.trim() !== '');
  
  if (validParams.length === 0) {
    emit('update:url', basePath);
    return;
  }

  const searchParams = new URLSearchParams();
  validParams.forEach(p => {
    searchParams.append(p.key, p.value);
  });

  const queryString = searchParams.toString();
  const newUrl = basePath + (queryString ? '?' + queryString : '');
  
  emit('update:url', newUrl);
};

// 监听外部 URL 变化
watch(
  () => props.url,
  (newUrl) => {
    if (isInternalChange.value) return;
    parseUrlToParams(newUrl);
  },
  { immediate: true }
);

// 监听内部表格变化
watch(
  params,
  () => {
    isInternalChange.value = true;
    buildUrlFromParams();
    ensureEmptyRow();
    
    // 在下一个 tick 恢复标识，防止死循环
    nextTick(() => {
      isInternalChange.value = false;
    });
  },
  { deep: true }
);

const removeRow = (index: number) => {
  // 不允许删除最后一行（空行）
  if (index === params.value.length - 1) return;
  params.value.splice(index, 1);
};
</script>

<template>
  <div class="query-params-editor">
    <n-table size="small" :bordered="false" :single-line="false" class="bg-transparent">
      <thead>
        <tr>
          <th class="w-1/3 bg-gray-800/50">Key</th>
          <th class="w-1/3 bg-gray-800/50">Value</th>
          <th class="bg-gray-800/50">Description</th>
          <th class="w-12 bg-gray-800/50"></th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="(row, index) in params" :key="row.id">
          <td class="p-0">
            <n-input
              v-model:value="row.key"
              placeholder="Key"
              size="small"
              :bordered="false"
              class="bg-transparent font-mono text-blue-400"
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
            {{ props.meta?.param_docs?.['query.' + row.key] || '-' }}
          </td>
          <td class="text-center p-0">
            <n-button
              v-if="index !== params.length - 1"
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

/* 斑马纹效果增强 */
tr:hover td {
  background-color: rgba(255, 255, 255, 0.03) !important;
}
</style>
