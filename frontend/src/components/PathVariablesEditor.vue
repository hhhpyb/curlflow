<script setup lang="ts">
import { computed, watch } from 'vue';
import { NTable, NInput, NEmpty } from 'naive-ui';

const props = defineProps<{
  url: string;
  meta: any | null; // domain.MetaData
  modelValue: Record<string, string>;
}>();

const emit = defineEmits<{
  (e: 'update:modelValue', value: Record<string, string>): void;
}>();

// 计算提取到的 keys
const extractedKeys = computed(() => {
  if (!props.url) return [];
  
  const keys = new Set<string>();
  // 使用 Lookbehind/Lookahead 排除 {{...}}，只匹配 {...}
  // 注意：需要环境支持 Lookbehind (Chrome 62+, Safari 16.4+, Edge 79+, Node 9+)
  const regex = /(?<!\{)\{([a-zA-Z0-9_-]+)\}(?!\})/g;
  
  const matches = props.url.matchAll(regex);
  for (const match of matches) {
    if (match[1]) {
      keys.add(match[1]);
    }
  }
  return Array.from(keys);
});

// 监听 extractedKeys 变化，同步 modelValue
// 策略：以 URL 中提取的 keys 为准。
// 1. URL 中存在的 key -> 保留原值 或 初始化为空
// 2. URL 中不存在的 key -> 移除 (清理旧数据)
watch(extractedKeys, (newKeys) => {
  const nextParams: Record<string, string> = {};
  let hasChanges = false;
  
  const currentKeys = Object.keys(props.modelValue || {});

  // 检查是否需要更新 (简单的 keys 集合比较)
  const newKeySet = new Set(newKeys);
  const currentKeySet = new Set(currentKeys);
  
  if (newKeySet.size !== currentKeySet.size) {
    hasChanges = true;
  } else {
    for (const k of newKeys) {
      if (!currentKeySet.has(k)) {
        hasChanges = true;
        break;
      }
    }
  }

  if (hasChanges) {
    newKeys.forEach(key => {
      nextParams[key] = props.modelValue[key] || '';
    });
    // Emit 新对象
    emit('update:modelValue', nextParams);
  }
}, { immediate: true });

// 处理输入框变化
const handleInput = (key: string, value: string) => {
  const newParams = { ...props.modelValue, [key]: value };
  emit('update:modelValue', newParams);
};
</script>

<template>
  <div class="path-variables-editor">
    <div v-if="extractedKeys.length > 0">
      <n-table size="small" :bordered="false" :single-line="false" class="bg-transparent">
        <thead>
          <tr>
            <th class="w-1/4 bg-gray-800/50 text-gray-400">Key</th>
            <th class="w-1/3 bg-gray-800/50 text-gray-400">Value</th>
            <th class="bg-gray-800/50 text-gray-400">Description</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="key in extractedKeys" :key="key">
            <td class="align-middle font-mono text-blue-400 select-all">{{ key }}</td>
            <td>
              <n-input 
                :value="props.modelValue[key] || ''" 
                @update:value="(val) => handleInput(key, val)"
                size="small" 
                placeholder="Value" 
                :bordered="false"
                class="bg-transparent"
              />
            </td>
            <td class="text-gray-500 text-xs align-middle px-3">
              {{ props.meta?.param_docs?.['path.' + key] || '-' }}
            </td>
          </tr>
        </tbody>
      </n-table>
    </div>
    <div v-else class="py-6">
      <n-empty size="small" description="No path variables found in URL">
        <template #extra>
          <span class="text-xs text-gray-500">
            Add <code class="bg-gray-800 px-1 rounded">{variable}</code> to the URL to extract params.
          </span>
        </template>
      </n-empty>
    </div>
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
