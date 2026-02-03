<script setup lang="ts">
import { computed } from 'vue';
import { useRequestStore } from '../stores/request';
import { NButton, NIcon, NScrollbar, NEmpty } from 'naive-ui';
import { FolderOpenOutline, DocumentTextOutline, AddOutline } from '@vicons/ionicons5';
import { domain } from '../../wailsjs/go/models';

const store = useRequestStore();

// 计算当前显示的目录名（仅显示最后一级目录）
const folderName = computed(() => {
  if (!store.workDir) return 'No Folder Opened';
  const parts = store.workDir.split(/[\\/]/);
  return parts[parts.length - 1] || store.workDir;
});

const handleNewRequest = () => {
  store.request = new domain.HttpRequest();
  store.request.method = 'GET';
  store.currentFileName = '';
  store.syncToCurl();
};
</script>

<template>
  <div class="sidebar flex flex-col h-full bg-[#1e1e1e] text-gray-300 border-r border-[#333] w-[250px] select-none">
    <!-- Header -->
    <div class="header flex items-center justify-between px-4 py-3 border-b border-[#333]">
      <span class="text-xs font-bold uppercase tracking-wider truncate flex-1 mr-2" :title="store.workDir">
        {{ folderName }}
      </span>
      <n-button text @click="store.chooseDir" class="text-gray-400 hover:text-white">
        <template #icon>
          <n-icon :component="FolderOpenOutline" />
        </template>
      </n-button>
    </div>

    <!-- Actions -->
    <div class="p-2">
      <n-button secondary block size="small" @click="handleNewRequest" class="justify-start px-2">
        <template #icon>
          <n-icon :component="AddOutline" />
        </template>
        New Request
      </n-button>
    </div>

    <!-- File List -->
    <div class="flex-1 overflow-hidden mt-1">
      <div class="px-4 py-1 text-[11px] font-bold text-gray-500 uppercase">Requests</div>
      <n-scrollbar>
        <div v-if="store.fileList.length > 0" class="flex flex-col">
          <div
            v-for="file in store.fileList"
            :key="file"
            @click="store.loadFrom(file)"
            class="group flex items-center px-4 py-1.5 cursor-pointer text-sm transition-colors duration-200"
            :class="[
              store.currentFileName === file 
                ? 'bg-[#37373d] text-white border-l-2 border-blue-500'
                : 'hover:bg-[#2a2d2e] text-gray-400 hover:text-gray-200 border-l-2 border-transparent'
            ]"
          >
            <n-icon :component="DocumentTextOutline" class="mr-2 text-gray-500 group-hover:text-blue-400" />
            <span class="truncate">{{ file.replace('.json', '') }}</span>
          </div>
        </div>
        <div v-else-if="!store.workDir" class="px-4 py-10">
          <n-empty size="small" description="Open a folder to see files" />
        </div>
        <div v-else class="px-4 py-10">
          <n-empty size="small" description="No .json files found" />
        </div>
      </n-scrollbar>
    </div>
  </div>
</template>

<style scoped>
.sidebar {
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif;
}

/* 隐藏滚动条背景 */
:deep(.n-scrollbar-rail) {
  --n-scrollbar-bezier: cubic-bezier(0.4, 0, 0.2, 1);
}
</style>
