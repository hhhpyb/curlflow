<script setup lang="ts">
import { computed, ref } from 'vue';
import { useRequestStore } from '../stores/request';
import { NButton, NIcon, NScrollbar, NEmpty, NModal, NCard, NInput, NSpace, useMessage } from 'naive-ui';
import { FolderOpenOutline, DocumentTextOutline, AddOutline, CloudDownloadOutline } from '@vicons/ionicons5';

const store = useRequestStore();
const message = useMessage();

// Sync Modal State
const showSyncModal = ref(false);
const swaggerUrl = ref('');
const isSyncing = ref(false);

// Display current folder name (only the last part)
const folderName = computed(() => {
  if (!store.workDir) return 'No Folder Opened';
  const parts = store.workDir.split(/[\\/]/);
  return parts[parts.length - 1] || store.workDir;
});

const handleNewRequest = () => {
  store.createNewRequest();
};

const handleOpenSync = () => {
  if (!store.workDir) {
    message.warning('请先打开一个工作目录');
    return;
  }
  showSyncModal.value = true;
};

const handleStartSync = async () => {
  if (!swaggerUrl.value.trim()) {
    message.warning('请输入 Swagger URL');
    return;
  }

  isSyncing.value = true;
  try {
    await store.importSwagger(swaggerUrl.value.trim());
    showSyncModal.value = false;
    swaggerUrl.value = '';
  } catch (e) {
    // Error is handled in store
  } finally {
    isSyncing.value = false;
  }
};
</script>

<template>
  <div class="sidebar flex flex-col h-full bg-gray-900 text-gray-300 border-r border-gray-800 w-[250px] select-none">
    <!-- Header -->
    <div class="header flex items-center justify-between px-4 py-3 border-b border-gray-800 shrink-0">
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
    <div class="p-2 border-b border-gray-800/50 shrink-0">
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
        <div v-if="store.fileList.length > 0" class="flex flex-col gap-0.5 pb-4">
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

    <!-- Footer Sync Action -->
    <div class="p-4 border-t border-gray-800 shrink-0 bg-gray-900/50">
      <n-button 
        secondary 
        block 
        size="medium" 
        @click="handleOpenSync"
        class="bg-blue-500/10 hover:bg-blue-500/20 text-blue-400 border-blue-500/30"
      >
        <template #icon>
          <n-icon :component="CloudDownloadOutline" />
        </template>
        Sync Swagger
      </n-button>
    </div>

    <!-- Sync Modal -->
    <n-modal v-model:show="showSyncModal">
      <n-card
        style="width: 500px"
        title="Sync Swagger / OpenAPI"
        :bordered="false"
        size="huge"
        role="dialog"
        aria-modal="true"
      >
        <n-space vertical size="large">
          <div class="text-sm text-gray-400">
            请输入 Swagger/OpenAPI 的 JSON URL。系统将自动解析路径并同步到本地工作区。
          </div>
          
          <n-input 
            v-model:value="swaggerUrl" 
            placeholder="https://example.com/v2/api-docs" 
            @keyup.enter="handleStartSync"
          />
          
          <div class="p-3 bg-blue-500/5 border border-blue-500/10 rounded text-xs text-blue-300/80">
            <strong>同步说明：</strong> 此操作将创建新接口并更新现有接口的结构，但系统遵循“用户数据优先”原则，不会覆盖您手动填写的 Body 和 Header 值。
          </div>

          <div class="flex justify-end gap-2 mt-2">
            <n-button @click="showSyncModal = false">Cancel</n-button>
            <n-button 
              type="primary" 
              :loading="isSyncing" 
              @click="handleStartSync"
            >
              Start Sync
            </n-button>
          </div>
        </n-space>
      </n-card>
    </n-modal>
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