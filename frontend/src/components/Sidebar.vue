<script setup lang="ts">
import { computed, ref } from 'vue';
import { useRequestStore } from '../stores/request';
import {
  NButton, NIcon, NScrollbar, NEmpty, NModal, NCard, NInput, NSpace,
  useMessage, NCollapse, NCollapseItem, NBadge, NTooltip
} from 'naive-ui';
import {
  FolderOpenOutline, DocumentTextOutline, AddOutline,
  CloudDownloadOutline, EyeOutline, EyeOffOutline, SearchOutline
} from '@vicons/ionicons5';

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

const toggleShowDeleted = () => {
  store.showDeleted = !store.showDeleted;
};
</script>

<template>
  <div class="sidebar flex flex-col h-full bg-gray-900 text-gray-300 border-r border-gray-800 w-[280px] select-none">
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

    <!-- Actions & Search -->
    <div class="p-2 border-b border-gray-800/50 shrink-0 flex flex-col gap-2">
      <n-button secondary block size="small" @click="handleNewRequest" class="justify-start px-2 bg-gray-800 hover:bg-gray-700 text-gray-300 border-gray-700">
        <template #icon>
          <n-icon :component="AddOutline" />
        </template>
        New Request
      </n-button>

      <div class="flex items-center gap-1">
        <n-input
          v-model:value="store.searchKeyword"
          size="small"
          placeholder="Search requests..."
          clearable
          class="bg-gray-800"
        >
          <template #prefix>
            <n-icon :component="SearchOutline" />
          </template>
        </n-input>
        <n-button
          secondary
          size="small"
          @click="toggleShowDeleted"
          :class="store.showDeleted ? 'text-blue-400 bg-blue-500/10' : 'text-gray-500'"
          title="Toggle Deleted Items"
        >
          <template #icon>
            <n-icon :component="store.showDeleted ? EyeOutline : EyeOffOutline" />
          </template>
        </n-button>
      </div>
    </div>

    <!-- File Tree List -->
    <div class="flex-1 overflow-hidden mt-1">
      <n-scrollbar>
        <div v-if="Object.keys(store.fileTree).length > 0" class="pb-4">
          <n-collapse :default-expanded-names="Object.keys(store.fileTree)" arrow-placement="right">
            <n-collapse-item
              v-for="(files, folder) in store.fileTree"
              :key="folder"
              :name="folder"
              class="px-2"
            >
              <template #header>
                <div class="text-xs font-semibold text-gray-500 flex items-center truncate">
                  <span class="truncate">{{ folder }}</span>
                  <span class="ml-2 opacity-50 text-[10px]">({{ files.length }})</span>
                </div>
              </template>

              <div class="flex flex-col gap-0.5 mt-1">
                <div
                  v-for="file in files"
                  :key="file.fileName"
                  @click="store.loadFrom(file.fileName)"
                  class="group relative flex items-center px-2 py-1.5 cursor-pointer text-sm transition-colors duration-150 rounded border-l-2"
                  :class="[
                    store.currentFileName === file.fileName
                      ? 'bg-blue-500/10 text-blue-400 border-blue-500'
                      : 'border-transparent text-gray-400 hover:bg-gray-800 hover:text-gray-200'
                  ]"
                >
                  <n-tooltip trigger="hover" placement="right">
                    <template #trigger>
                      <div class="flex items-center flex-1 min-w-0">
                        <n-badge
                          v-if="file.meta.status === 'new'"
                          dot
                          type="success"
                          class="mr-2 shrink-0"
                        />
                        <n-icon
                          v-else
                          :component="DocumentTextOutline"
                          class="mr-2 shrink-0 transition-colors duration-150"
                          :class="store.currentFileName === file.fileName ? 'text-blue-400' : 'text-gray-600 group-hover:text-gray-400'"
                        />
                        <span
                          class="truncate flex-1"
                          :class="{ 'line-through text-gray-600': file.meta.status === 'deleted' }"
                        >
                          {{ file.meta.summary || file.meta.key || file.fileName.replace('.json', '') }}
                        </span>
                      </div>
                    </template>
                    <div class="text-xs">
                      <div>File: {{ file.fileName }}</div>
                      <div v-if="file.meta.key" class="opacity-70 mt-1">Key: {{ file.meta.key }}</div>
                    </div>
                  </n-tooltip>
                </div>
              </div>
            </n-collapse-item>
          </n-collapse>
        </div>

        <!-- Empty States -->
        <div v-else-if="!store.workDir" class="px-4 py-10 opacity-60">
          <n-empty size="small" description="Open a folder" />
        </div>
        <div v-else class="px-4 py-10 opacity-60">
          <n-empty size="small" description="No results found" />
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

/* Remove default padding/margin for collapse to make it tighter */
:deep(.n-collapse .n-collapse-item .n-collapse-item__header) {
  padding-top: 4px;
  padding-bottom: 4px;
}
:deep(.n-collapse .n-collapse-item .n-collapse-item__content-inner) {
  padding-top: 0 !important;
  padding-bottom: 8px !important;
}
</style>
