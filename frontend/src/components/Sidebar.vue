<script setup lang="ts">
import { computed } from 'vue';
import { useRequestStore } from '../stores/request';
import { NButton, NIcon, NScrollbar, NEmpty } from 'naive-ui';
import { FolderOpenOutline, DocumentTextOutline, AddOutline } from '@vicons/ionicons5';

const store = useRequestStore();

// Display current folder name (only the last part)
const folderName = computed(() => {
  if (!store.workDir) return 'No Folder Opened';
  const parts = store.workDir.split(/[\\/]/);
  return parts[parts.length - 1] || store.workDir;
});

const handleNewRequest = () => {
  store.createNewRequest();
};
</script>

<template>
  <div class="sidebar flex flex-col h-full bg-gray-900 text-gray-300 border-r border-gray-800 w-[250px] select-none">
    <!-- Header -->
    <div class="header flex items-center justify-between px-4 py-3 border-b border-gray-800">
      <span class="text-xs font-bold uppercase tracking-wider truncate flex-1 mr-2 text-gray-400" :title="store.workDir">
        {{ folderName }}
      </span>
      <n-button text @click="store.chooseDir" class="text-gray-400 hover:text-white">
        <template #icon>
          <n-icon :component="FolderOpenOutline" />
        </template>
      </n-button>
    </div>

    <!-- Actions -->
    <div class="p-2 border-b border-gray-800/50">
      <n-button secondary block size="small" @click="handleNewRequest" class="justify-start px-2 bg-gray-800 hover:bg-gray-700 text-gray-300 border-gray-700">
        <template #icon>
          <n-icon :component="AddOutline" />
        </template>
        New Request
      </n-button>
    </div>

    <!-- File List -->
    <div class="flex-1 overflow-hidden mt-1">
      <div class="px-4 py-2 text-[10px] font-bold text-gray-500 uppercase tracking-widest">Requests</div>
      <n-scrollbar>
        <div v-if="store.fileList.length > 0" class="flex flex-col gap-0.5">
          <div
            v-for="file in store.fileList"
            :key="file"
            @click="store.loadFrom(file)"
            class="group flex items-center px-4 py-1.5 cursor-pointer text-sm transition-colors duration-150 border-l-2"
            :class="[
              store.currentFileName === file 
                ? 'bg-gray-800 text-white border-blue-500'
                : 'border-transparent text-gray-400 hover:bg-gray-800 hover:text-gray-200'
            ]"
          >
            <n-icon 
              :component="DocumentTextOutline" 
              class="mr-2 transition-colors duration-150"
              :class="store.currentFileName === file ? 'text-blue-400' : 'text-gray-600 group-hover:text-gray-400'"
            />
            <span class="truncate">{{ file.replace('.json', '') }}</span>
          </div>
        </div>
        
        <!-- Empty States -->
        <div v-else-if="!store.workDir" class="px-4 py-10 opacity-60">
          <n-empty size="small" description="Open a folder" />
        </div>
        <div v-else class="px-4 py-10 opacity-60">
          <n-empty size="small" description="No requests found" />
        </div>
      </n-scrollbar>
    </div>
  </div>
</template>

<style scoped>
.sidebar {
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif;
}

/* Customize Scrollbar to blend in */
:deep(.n-scrollbar-rail) {
  background-color: transparent !important;
}
:deep(.n-scrollbar-rail .n-scrollbar-rail__scrollbar) {
  background-color: rgba(255, 255, 255, 0.2) !important;
}
</style>