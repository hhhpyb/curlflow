<script setup lang="ts">
import { computed, ref } from 'vue';
import { useRequestStore } from '../stores/request';
import {
  NButton, NIcon, NScrollbar, NEmpty, NModal, NCard, NInput, NSpace,
  useMessage, NCollapse, NCollapseItem, NBadge
} from 'naive-ui';
import {
  FolderOpenOutline, DocumentTextOutline, AddOutline,
  CloudDownloadOutline, EyeOutline, EyeOffOutline, SearchOutline,
  ChevronForwardOutline, ChevronDownOutline, FlaskOutline
} from '@vicons/ionicons5';

const store = useRequestStore();
const message = useMessage();

// Sync Modal State
const showSyncModal = ref(false);
const isSyncing = ref(false);

// Track expanded interface nodes (multi-case groups)
const expandedNodes = ref<Set<string>>(new Set());

const toggleNode = (fileName: string) => {
  if (expandedNodes.value.has(fileName)) {
    expandedNodes.value.delete(fileName);
  } else {
    expandedNodes.value.add(fileName);
  }
};

const folderName = computed(() => {
  if (!store.workDir) return 'No Folder Opened';
  const parts = store.workDir.split(/[\\/]/);
  return parts[parts.length - 1] || store.workDir;
});

const handleNewRequest = () => {
  store.createNewRequest();
};

const handleOpenSync = async () => {
  if (!store.workDir) {
    message.warning('请先打开一个工作目录');
    return;
  }
  await store.loadProjectConfig();
  showSyncModal.value = true;
};

const handleStartSync = async () => {
  if (!store.swaggerUrl.trim()) {
    message.warning('请输入 Swagger URL');
    return;
  }

  isSyncing.value = true;
  try {
    // Save the URL first
    await store.saveProjectConfig(store.swaggerUrl.trim());
    // Then import
    await store.importSwagger(store.swaggerUrl.trim());
    showSyncModal.value = false;
  } catch (e) {
    // Error is handled in store
  } finally {
    isSyncing.value = false;
  }
};

const toggleShowDeleted = () => {
  store.showDeleted = !store.showDeleted;
};

/**
 * 解析用例名称：从 get_user_case_error.json 中提取 "error"
 */
const getCaseLabel = (fileName: string, mainFileName: string) => {
  const prefix = mainFileName.replace('.json', '') + '_case_';
  return fileName.replace(prefix, '').replace('.json', '');
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
          placeholder="Search..."
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
        >
          <template #icon>
            <n-icon :component="store.showDeleted ? EyeOutline : EyeOffOutline" />
          </template>
        </n-button>
      </div>
    </div>

    <!-- Three-Level File Tree -->
    <div class="flex-1 overflow-hidden mt-1">
      <n-scrollbar>
        <div v-if="Object.keys(store.fileTree).length > 0" class="pb-4">
          <n-collapse :default-expanded-names="Object.keys(store.fileTree)" arrow-placement="right">
            <!-- Level 1: Folder (Tag) -->
            <n-collapse-item
              v-for="(nodes, folder) in store.fileTree"
              :key="folder"
              :name="folder"
              class="px-2"
            >
              <template #header>
                <div class="text-xs font-bold text-gray-500 flex items-center truncate">
                  <span class="truncate uppercase tracking-tighter">{{ folder }}</span>
                  <span class="ml-2 opacity-40 text-[9px]">[{{ nodes.length }}]</span>
                </div>
              </template>

              <div class="flex flex-col gap-0.5 mt-1">
                <!-- Level 2: Interface Node -->
                <div v-for="node in nodes" :key="node.mainFile.fileName" class="interface-group">
                  <div
                    class="group flex items-center px-2 py-1.5 cursor-pointer text-sm transition-colors duration-150 rounded border-l-2"
                    :class="[
                      store.currentFileName === node.mainFile.fileName
                        ? 'bg-blue-500/10 text-blue-400 border-blue-500'
                        : 'border-transparent text-gray-400 hover:bg-gray-800 hover:text-gray-200'
                    ]"
                  >
                    <!-- Expand Icon for Cases -->
                    <div 
                      v-if="node.children.length > 0"
                      class="mr-1 hover:bg-white/10 rounded p-0.5 flex items-center transition-colors"
                      @click.stop="toggleNode(node.mainFile.fileName)"
                    >
                      <n-icon :component="expandedNodes.has(node.mainFile.fileName) ? ChevronDownOutline : ChevronForwardOutline" size="12" />
                    </div>
                    <div v-else class="w-[18px]" />

                    <!-- Main Label -->
                    <div class="flex items-center flex-1 min-w-0" @click="store.loadFrom(node.mainFile.fileName)">
                      <n-badge v-if="node.mainFile.meta.status === 'new'" dot type="success" class="mr-2 shrink-0" />
                      <n-icon
                        v-else
                        :component="DocumentTextOutline"
                        class="mr-2 shrink-0 opacity-50"
                        :class="store.currentFileName === node.mainFile.fileName ? 'text-blue-400 opacity-100' : ''"
                      />
                      <span class="truncate" :class="{ 'line-through text-gray-600': node.mainFile.meta.status === 'deleted' }">
                        {{ node.mainFile.meta.summary || node.mainFile.meta.key || node.mainFile.fileName.replace('.json', '') }}
                      </span>
                    </div>
                  </div>

                  <!-- Level 3: Test Cases (Children) -->
                  <div v-if="expandedNodes.has(node.mainFile.fileName) && node.children.length > 0" class="ml-6 mt-0.5 flex flex-col gap-0.5 border-l border-gray-800 pl-1">
                    <div
                      v-for="child in node.children"
                      :key="child.fileName"
                      @click="store.loadFrom(child.fileName)"
                      class="flex items-center px-2 py-1 cursor-pointer rounded transition-colors"
                      :class="[
                        store.currentFileName === child.fileName
                          ? 'bg-blue-500/10 text-blue-300'
                          : 'text-gray-500 hover:bg-gray-800 hover:text-gray-300'
                      ]"
                    >
                      <n-icon :component="FlaskOutline" size="10" class="mr-2 opacity-40" />
                      <span 
                        class="text-[11px] truncate"
                        :class="{ 'line-through text-gray-700': child.meta.status === 'deleted' }"
                      >
                        {{ getCaseLabel(child.fileName, node.mainFile.fileName) }}
                      </span>
                      <n-badge v-if="child.meta.status === 'new'" dot type="success" class="ml-auto" />
                    </div>
                  </div>
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
      >
        <n-space vertical size="large">
          <n-input v-model:value="store.swaggerUrl" placeholder="Swagger URL" @keyup.enter="handleStartSync" />
          <div class="flex justify-end gap-2">
            <n-button @click="showSyncModal = false">Cancel</n-button>
            <n-button type="primary" :loading="isSyncing" @click="handleStartSync">Start Sync</n-button>
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
:deep(.n-collapse .n-collapse-item .n-collapse-item__header) {
  padding-top: 6px;
  padding-bottom: 6px;
}
:deep(.n-collapse .n-collapse-item .n-collapse-item__content-inner) {
  padding-top: 0 !important;
  padding-bottom: 12px !important;
}
.interface-group {
  margin-bottom: 2px;
}
</style>