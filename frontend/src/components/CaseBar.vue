<script setup lang="ts">
import { ref } from 'vue';
import { useRequestStore } from '../stores/request';
import { NButton, NIcon, NTabs, NTab, NPopover, NInput, NSpace, useMessage } from 'naive-ui';
import { AddOutline, BookmarksOutline } from '@vicons/ionicons5';

const store = useRequestStore();
const message = useMessage();

const showAddPopover = ref(false);
const newCaseName = ref('');

/**
 * 根据文件名计算 Tab 显示名称
 * 主用例显示 "Default"，子用例显示后缀名
 */
const getCaseLabel = (fileName: string) => {
  // relatedCases 已经按长度排序，第一个通常是主用例
  const mainName = store.relatedCases[0]?.fileName;

  if (fileName === mainName) return 'Default';
  
  if (fileName.includes('_case_')) {
    const parts = fileName.replace('.json', '').split('_case_');
    return parts[parts.length - 1];
  }
  
  return fileName.replace('.json', '');
};

const handleAddCase = async () => {
  if (!newCaseName.value.trim()) {
    message.warning('请输入用例名称');
    return;
  }
  try {
    await store.createCase(newCaseName.value.trim());
    newCaseName.value = '';
    showAddPopover.value = false;
    message.success('用例创建成功');
  } catch (e: any) {
    message.error('创建失败: ' + e.message);
  }
};
</script>

<template>
  <div v-if="store.meta && store.meta.id" class="case-bar flex items-center px-3 bg-[#111] border-b border-gray-800 h-10 select-none">
    <!-- Title Area -->
    <div class="flex items-center text-gray-500 text-[10px] mr-4 shrink-0">
      <n-icon :component="BookmarksOutline" class="mr-1 text-xs" />
      <span class="font-bold uppercase tracking-wider">Test Cases</span>
    </div>

    <!-- Tabs Area -->
    <div class="flex-1 overflow-hidden h-full">
      <n-tabs
        type="card"
        size="small"
        :value="store.currentFileName"
        @update:value="store.loadFrom"
        class="case-tabs h-full"
      >
        <n-tab
          v-for="file in store.relatedCases"
          :key="file.fileName"
          :name="file.fileName"
        >
          <span class="max-w-[120px] truncate text-xs">{{ getCaseLabel(file.fileName) }}</span>
        </n-tab>
      </n-tabs>
    </div>

    <!-- Actions Area -->
    <div class="ml-2 shrink-0">
      <n-popover
        v-model:show="showAddPopover"
        trigger="click"
        placement="bottom-end"
        :width="240"
        style="padding: 12px"
      >
        <template #trigger>
          <n-button circle size="tiny" type="primary" secondary title="New Test Case">
            <template #icon><n-icon :component="AddOutline" /></template>
          </n-button>
        </template>
        
        <div class="case-popover-content">
          <div class="text-[10px] font-bold mb-3 text-gray-500 uppercase tracking-widest">Create New Case</div>
          <n-space vertical size="medium">
            <n-input 
              v-model:value="newCaseName" 
              placeholder="e.g. error_response" 
              size="small"
              autofocus
              @keyup.enter="handleAddCase"
            />
            <div class="flex justify-end gap-2">
              <n-button size="tiny" quaternary @click="showAddPopover = false">Cancel</n-button>
              <n-button size="tiny" type="primary" @click="handleAddCase">Create</n-button>
            </div>
          </n-space>
        </div>
      </n-popover>
    </div>
  </div>
</template>

<style scoped>
.case-tabs :deep(.n-tabs-tab-pad) {
  padding: 0 10px !important;
}

.case-tabs :deep(.n-tabs-wrapper) {
  height: 100%;
}

.case-tabs :deep(.n-tabs-nav) {
  height: 100%;
}

.case-tabs :deep(.n-tabs-tab) {
  background-color: transparent !important;
  border: none !important;
  height: 100% !important;
  transition: all 0.2s;
  color: #666;
}

.case-tabs :deep(.n-tabs-tab:hover) {
  color: #aaa !important;
}

.case-tabs :deep(.n-tabs-tab.n-tabs-tab--active) {
  color: #3b82f6 !important;
  background-color: rgba(59, 130, 246, 0.05) !important;
  font-weight: 600;
}

.case-tabs :deep(.n-tabs-bar) {
  height: 2px !important;
  bottom: 0 !important;
  background-color: #3b82f6 !important;
}

.case-bar {
  /* 确保整体布局严丝合缝 */
  box-sizing: border-box;
}
</style>
