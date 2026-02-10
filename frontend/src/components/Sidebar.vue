<script setup lang="ts">
import { computed, ref, nextTick, onMounted, Component, h } from 'vue';
import { useRequestStore } from '../stores/request';
import { useHistoryStore } from '../stores/history';
import {
  NButton, NIcon, NScrollbar, NEmpty, NModal, NCard, NInput, NSpace,
  useMessage, NCollapse, NCollapseItem, NBadge, NDivider, useDialog,
  NDropdown, NTag
} from 'naive-ui';
import type { DropdownOption } from 'naive-ui';
import { GetRecentProjects, RemoveProject } from '../../wailsjs/go/main/App';
import {
  FolderOpenOutline, DocumentTextOutline, AddOutline,
  CloudDownloadOutline, EyeOutline, EyeOffOutline, SearchOutline,
  ChevronForwardOutline, ChevronDownOutline, FlaskOutline, TrashOutline,
  CopyOutline, DuplicateOutline, TimeOutline, FolderOutline, CheckmarkOutline, CloseOutline,
  SyncOutline, ChevronUpOutline, ChevronDown, DesktopOutline
} from '@vicons/ionicons5';
import { useEnvStore } from '../stores/env';
import { useI18n } from 'vue-i18n';

const { t } = useI18n();
const store = useRequestStore();
const envStore = useEnvStore();
const historyStore = useHistoryStore();
const message = useMessage();
const dialog = useDialog();

// Sidebar View Mode: 'collections' | 'history'
const activeView = ref<'collections' | 'history'>('collections');

onMounted(() => {
  // Setup event listeners for history
  historyStore.setupListeners();
  store.fetchRecentProjects(); // Load projects on mount
});

// Watch for workDir changes to load history
import { watch } from 'vue';
watch(() => store.workDir, (newVal) => {
  if (newVal) {
    historyStore.fetchHistory();
  }
});

// ==========================================
// Collections Logic (Legacy Sidebar Logic)
// ==========================================

// Context Menu State
const showDropdown = ref(false);
const x = ref(0);
const y = ref(0);
const dropdownTargetFile = ref('');

const renderIcon = (icon: Component) => {
  return () => h(NIcon, null, { default: () => h(icon) });
};

const dropdownOptions = [
  {
    label: 'Copy as cURL',
    key: 'copy-curl',
    icon: renderIcon(CopyOutline)
  },
  {
    label: 'Duplicate (⌘ + D)',
    key: 'duplicate',
    icon: renderIcon(DuplicateOutline)
  },
  {
    label: 'Delete  (⌘ + Delete)',
    key: 'delete',
    icon: renderIcon(TrashOutline),
    props: {
      style: 'color: #e88080'
    }
  }
];

const handleContextMenu = (e: MouseEvent, fileName: string) => {
  showDropdown.value = false;
  nextTick().then(() => {
    showDropdown.value = true;
    x.value = e.clientX;
    y.value = e.clientY;
    dropdownTargetFile.value = fileName;
  });
};

const handleSelect = (key: string) => {
  showDropdown.value = false;
  if (!dropdownTargetFile.value) return;

  switch (key) {
    case 'copy-curl':
      store.copyCurlByFilename(dropdownTargetFile.value);
      break;
    case 'duplicate':
      store.duplicateFile(dropdownTargetFile.value);
      break;
    case 'delete':
      store.deleteFileByFilename(dropdownTargetFile.value);
      break;
  }
};

// Sync Modal State
const showSyncModal = ref(false);
const isSyncing = ref(false);

const handlePurge = () => {
  dialog.warning({
    title: 'Confirm Purge',
    content: 'Are you sure you want to permanently delete all files marked as "deleted"? This action cannot be undone.',
    positiveText: 'Yes, Delete',
    negativeText: 'Cancel',
    onPositiveClick: async () => {
      await store.purgeDeleted();
      showSyncModal.value = false;
    }
  });
};

const toggleNode = (fileName: string) => {
  store.toggleExpand(fileName);
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
    store.projectConfig.swagger_url = store.swaggerUrl.trim();
    await store.saveProjectConfig();
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

const getCaseLabel = (fileName: string, mainFileName: string) => {
  const prefix = mainFileName.replace('.json', '') + '_case_';
  return fileName.replace(prefix, '').replace('.json', '');
};

// ==========================================
// History Logic
// ==========================================

const formatTime = (ts: number) => {
  const date = new Date(ts * 1000);
  const m = (date.getMonth() + 1).toString().padStart(2, '0');
  const d = date.getDate().toString().padStart(2, '0');
  const hh = date.getHours().toString().padStart(2, '0');
  const mm = date.getMinutes().toString().padStart(2, '0');
  return `${m}-${d} ${hh}:${mm}`;
};

const handleLoadHistory = (entry: any) => {
  historyStore.loadEntry(entry);
};

const handleClearHistory = () => {
  dialog.warning({
    title: 'Clear History',
    content: 'Are you sure you want to clear all history records for this project?',
    positiveText: 'Clear',
    negativeText: 'Cancel',
    onPositiveClick: async () => {
      await historyStore.clearHistory();
    }
  });
};

const getMethodColor = (method: string) => {
  switch (method.toUpperCase()) {
    case 'GET': return 'info';
    case 'POST': return 'success';
    case 'PUT': return 'warning';
    case 'DELETE': return 'error';
    default: return 'default';
  }
};

const getMethodClass = (method: string) => {
  const m = (method || 'GET').toUpperCase();
  switch (m) {
    case 'GET': return 'text-green-500';
    case 'POST': return 'text-orange-500';
    case 'PUT': return 'text-blue-500';
    case 'DELETE': return 'text-red-500';
    case 'PATCH': return 'text-purple-500';
    default: return 'text-gray-500';
  }
};

// ==========================================
// Secondary Toolbar Logic
// ==========================================
const envSelectorOptions = computed(() => {
  const options = envStore.envList.map(name => ({
    label: name,
    key: name
  }));
  return [{ label: 'No Environment', key: '' }, ...options];
});

const handleEnvSelect = (key: string) => {
  envStore.setActiveEnv(key);
  envStore.saveEnvs();
};

const handleSyncToolbar = async () => {
  if (!store.workDir) return;
  
  if (!store.projectConfig.swagger_url) {
    // If no URL configured, show the full sync modal to let user input it
    await store.loadProjectConfig();
    showSyncModal.value = true;
    return;
  }

  // Quick Sync
  isSyncing.value = true;
  try {
    await store.importSwagger(store.projectConfig.swagger_url);
    message.success('Swagger Synced');
  } catch (e) {
    // Error handled in store
  } finally {
    isSyncing.value = false;
  }
};

const renderEnvLabel = (option: DropdownOption) => {
  const isCurrent = envStore.activeEnvName === option.key;
  return h('div', { class: 'flex items-center gap-2 py-0.5 min-w-[120px]' }, [
    isCurrent 
      ? h(NIcon, { component: CheckmarkOutline, class: 'text-green-500 shrink-0', size: '14' }) 
      : h('div', { class: 'w-[14px] shrink-0' }),
    h('span', { class: isCurrent ? 'text-green-400 font-medium' : 'text-gray-300' }, option.label as string)
  ]);
};

// ==========================================
// Project Dropdown Custom Rendering
// ==========================================
const renderProjectLabel = (option: any) => {
  if (option.type === 'divider') return null;

  // Render "Open Folder..." with its icon
  if (option.key === 'open-folder') {
    return h('div', { class: 'flex items-center gap-2 py-1' }, [
      h(NIcon, { component: FolderOpenOutline, class: 'text-gray-400' }),
      h('span', { class: 'font-medium' }, option.label)
    ]);
  }

  const isCurrent = store.workDir === option.path;

  return h('div', { 
    class: 'flex items-center justify-between w-full min-w-[300px] py-1 group/item',
  }, [
    h('div', { class: 'flex items-center gap-3 flex-1 min-w-0' }, [
      isCurrent 
        ? h(NIcon, { component: CheckmarkOutline, class: 'text-blue-500 shrink-0', size: '16' }) 
        : h('div', { class: 'w-4 shrink-0' }),
      h('div', { class: 'flex flex-col min-w-0 flex-1 leading-tight' }, [
        h('span', { class: 'text-[13px] font-semibold truncate text-gray-200' }, option.label),
        h('span', { class: 'text-[10px] text-gray-500 truncate mt-0.5 font-mono opacity-80' }, option.path),
      ])
    ]),
    h('div', { class: 'w-8 flex justify-end shrink-0 ml-2' }, [
      h(NButton, { 
        text: true, 
        size: 'tiny',
        class: 'opacity-0 group-hover/item:opacity-100 transition-opacity hover:text-red-400 text-gray-500 p-1',
        onClick: (e: MouseEvent) => {
          e.stopPropagation();
          store.removeProject(option.path);
        }
      }, { icon: () => h(NIcon, { component: CloseOutline, size: '14' }) })
    ])
  ]);
};
</script>

<template>
  <div class="flex flex-col h-full w-full overflow-hidden bg-[#0f111a] text-gray-300 border-r border-gray-800">
    
    <!-- Row 1: Header (Traffic Lights + Project) -->
    <div 
      class="flex items-center w-full h-[40px] shrink-0 select-none bg-gray-900 border-b border-gray-800"
      style="--wails-draggable: drag;"
    >
      <!-- Traffic Light Spacer -->
      <div class="w-[70px] h-full flex-shrink-0"></div>

      <!-- Project Selector -->
      <n-dropdown 
        trigger="click" 
        :options="store.projectOptions" 
        :render-label="renderProjectLabel"
        @select="store.handleProjectSelect" 
        placement="bottom-start"
        class="bg-gray-800 border border-gray-700"
      >
        <div class="flex items-center gap-1.5 px-2 py-1 hover:bg-white/10 rounded cursor-pointer no-drag transition-colors" style="--wails-draggable: no-drag">
          <span class="text-[13px] font-semibold text-gray-200 truncate max-w-[140px]">{{ folderName }}</span>
          <n-icon :component="ChevronDownOutline" size="10" class="opacity-50" />
        </div>
      </n-dropdown>

      <div class="flex-1"></div>
    </div>

    <div class="flex flex-1 overflow-hidden">
      <!-- Activity Bar (Far Left) -->
      <div class="activity-bar w-11 flex flex-col items-center py-4 bg-gray-900 border-r border-gray-800 z-10 shrink-0">
        <div 
          class="icon-btn mb-4" 
          :class="{ active: activeView === 'collections' }"
          @click="activeView = 'collections'"
          title="Collections"
        >
          <n-icon :component="FolderOutline" size="20" />
        </div>
        
        <div 
          class="icon-btn" 
          :class="{ active: activeView === 'history' }"
          @click="activeView = 'history'"
          title="History"
        >
          <n-icon :component="TimeOutline" size="20" />
        </div>
      </div>

      <!-- Side Panel Content -->
      <div class="side-panel flex-1 flex flex-col h-full overflow-hidden bg-gray-900 w-64 min-w-[240px]">
        
        <!-- View: Collections (File Tree) -->
        <div v-if="activeView === 'collections'" class="flex flex-col h-full w-full">

            <!-- Row 2: Search & Create -->
            <div class="flex items-center px-2 gap-2 h-[36px] shrink-0 pt-1">
            <n-input
                v-model:value="store.searchKeyword"
                round
                size="small"
                :placeholder="t('common.search')"
                class="flex-1 bg-gray-800/50"
                :bordered="false"
                clearable
                >
                <template #prefix>
                    <n-icon :component="SearchOutline" class="opacity-50" />
                </template>
                </n-input>
                
                <n-button 
                circle 
                size="small" 
                @click="handleNewRequest" 
                title="New Request" 
                class="bg-white/5 hover:bg-white/10 text-gray-400 hover:text-white"
                >
                <template #icon><n-icon :component="AddOutline" size="18" /></template>
                </n-button>
            </div>

            <!-- Row 3: Toolbar -->
            <div class="flex items-center justify-end px-2 gap-1 h-[32px] shrink-0 border-b border-gray-800">
            <!-- Sync Swagger -->
            <n-tooltip trigger="hover">
                <template #trigger>
                <n-button quaternary size="tiny" @click="handleSyncToolbar" :disabled="isSyncing" class="text-gray-500 hover:text-blue-400">
                    <template #icon><n-icon :component="SyncOutline" :class="{ 'animate-spin': isSyncing }" /></template>
                </n-button>
                </template>
                Sync Swagger
            </n-tooltip>

            <!-- Toggle Deleted -->
            <n-tooltip trigger="hover">
                <template #trigger>
                <n-button quaternary size="tiny" @click="toggleShowDeleted" :class="store.showDeleted ? 'text-blue-400' : 'text-gray-500'">
                    <template #icon><n-icon :component="store.showDeleted ? EyeOutline : EyeOffOutline" /></template>
                </n-button>
                </template>
                {{ store.showDeleted ? 'Hide Deleted' : 'Show Deleted' }}
            </n-tooltip>

            <!-- Env Switcher -->
            <n-dropdown trigger="click" :options="envSelectorOptions" :render-label="renderEnvLabel" @select="handleEnvSelect">
                <n-button quaternary size="tiny" class="text-gray-500 hover:text-green-400" title="Switch Environment">
                    <template #icon><n-icon :component="DesktopOutline" /></template>
                </n-button>
            </n-dropdown>

            <div class="w-px h-3 bg-gray-700 mx-1"></div>

            <!-- Expand/Collapse -->
                <n-button quaternary size="tiny" @click="store.expandAll()" title="Expand All" class="text-gray-500 hover:text-gray-300">
                <template #icon><n-icon :component="ChevronDown" /></template>
                </n-button>
                
                <n-button quaternary size="tiny" @click="store.collapseAll()" title="Collapse All" class="text-gray-500 hover:text-gray-300">
                <template #icon><n-icon :component="ChevronUpOutline" /></template>
                </n-button>
            </div>

          <!-- Three-Level File Tree -->
          <div class="flex-1 overflow-hidden mt-0">
            <n-scrollbar trigger="hover">
              <div v-if="Object.keys(store.fileTree).length > 0" class="pb-4">
                <n-collapse :expanded-names="Array.from(store.expandedKeys)" @update:expanded-names="(names) => store.expandedKeys = new Set(names)" arrow-placement="right">
                  <!-- Level 1: Folder (Tag) -->
                  <n-collapse-item
                    v-for="(nodes, folder) in store.fileTree"
                    :key="folder"
                    :name="folder"
                    class="px-2"
                  >
                    <template #header>
                      <div class="text-xs font-bold text-gray-500 flex items-center truncate max-w-full">
                        <span class="truncate uppercase tracking-tighter">{{ folder }}</span>
                        <span class="ml-2 opacity-40 text-[9px] shrink-0">[{{ nodes.length }}]</span>
                      </div>
                    </template>

                    <div class="flex flex-col gap-0.5 mt-1">
                      <!-- Level 2: Interface Node -->
                      <div v-for="node in nodes" :key="node.mainFile.fileName" class="interface-group">
                        <!-- Interface Entry Row -->
                        <div
                          :id="'file-item-' + node.mainFile.fileName"
                          class="group flex items-center px-2 py-1.5 cursor-pointer text-sm transition-all duration-150 rounded border-l-2 file-item relative"
                          :class="[
                            store.currentFileName === node.mainFile.fileName
                              ? 'bg-blue-500/10 text-blue-400 border-blue-500'
                              : 'border-transparent text-gray-400 hover:bg-gray-800 hover:text-gray-200'
                          ]"
                          @click="store.loadFrom(node.mainFile.fileName)"
                          @contextmenu.prevent="handleContextMenu($event, node.mainFile.fileName)"
                        >
                          <!-- Expand Icon - Toggle Only -->
                          <div 
                            v-if="node.children.length > 0"
                            class="mr-1 hover:bg-white/10 rounded p-0.5 flex items-center transition-colors shrink-0 z-10"
                            @click.stop="toggleNode(node.mainFile.fileName)"
                          >
                            <n-icon :component="store.expandedKeys.has(node.mainFile.fileName) ? ChevronDownOutline : ChevronForwardOutline" size="12" />
                          </div>
                          <div v-else class="w-[18px] shrink-0" />
    
                          <!-- Main Label -->
                          <div class="flex items-center flex-1 min-w-0 pointer-events-none">
                            <n-badge v-if="node.mainFile.meta.status === 'new'" dot type="success" class="mr-2 shrink-0" />
                            <div 
                              v-else 
                              class="w-9 text-[10px] font-bold shrink-0 text-left"
                              :class="getMethodClass(node.mainFile.method)"
                            >
                              {{ (node.mainFile.method || 'GET').toUpperCase() }}
                            </div>
                            <span class="file-name" :class="{ 'line-through text-gray-600': node.mainFile.meta.status === 'deleted' }">
                              {{ node.mainFile.meta.summary || node.mainFile.meta.key || node.mainFile.fileName.replace('.json', '') }}
                            </span>
                          </div>
                        </div>

                        <!-- Level 3: Test Cases (Children) -->
                        <div v-if="store.expandedKeys.has(node.mainFile.fileName) && node.children.length > 0" class="ml-6 mt-0.5 flex flex-col gap-0.5 border-l border-gray-800 pl-1">
                          <div
                            v-for="child in node.children"
                            :key="child.fileName"
                            :id="'file-item-' + child.fileName"
                            class="flex items-center px-2 py-1 cursor-pointer rounded transition-all file-item group/child"
                            :class="[
                              store.currentFileName === child.fileName
                                ? 'bg-blue-500/10 text-blue-300'
                                : 'text-gray-500 hover:bg-gray-800 hover:text-gray-300'
                            ]"
                            @click="store.loadFrom(child.fileName)"
                            @contextmenu.prevent="handleContextMenu($event, child.fileName)"
                          >
                            <div 
                              class="w-9 text-[10px] font-bold shrink-0 text-left opacity-70"
                              :class="getMethodClass(child.method)"
                            >
                              {{ (child.method || 'GET').toUpperCase() }}
                            </div>
                            <span 
                              class="text-[11px] file-name flex-1"
                              :class="{ 'line-through text-gray-700': child.meta.status === 'deleted' }"
                            >
                              {{ getCaseLabel(child.fileName, node.mainFile.fileName) }}
                            </span>
                            <n-badge v-if="child.meta.status === 'new'" dot type="success" class="ml-auto shrink-0" />
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
        </div>

        <!-- View: History List -->
        <div v-else-if="activeView === 'history'" class="flex flex-col h-full w-full">
          <div class="header flex items-center justify-between px-4 py-3 border-b border-gray-800 shrink-0 h-[46px]">
            <span class="text-xs font-bold uppercase tracking-wider text-gray-400">History</span>
            <n-button text size="tiny" @click="handleClearHistory" class="text-gray-500 hover:text-red-400">
              <template #icon>
                <n-icon :component="TrashOutline" />
              </template>
            </n-button>
          </div>

          <div class="flex-1 overflow-hidden bg-gray-900/50">
            <n-scrollbar trigger="hover">
              <div v-if="historyStore.list.length > 0" class="flex flex-col">
                <div 
                  v-for="entry in historyStore.list" 
                  :key="entry.id"
                  class="px-3 py-2 border-b border-gray-800 hover:bg-gray-800 cursor-pointer group"
                  @click="handleLoadHistory(entry)"
                >
                  <div class="flex items-center gap-2 mb-1">
                    <n-tag :type="getMethodColor(entry.request.method)" size="small" class="font-bold text-[10px] px-1 h-5">
                        {{ entry.request.method }}
                    </n-tag>
                    <span class="text-[10px] text-gray-500 ml-auto">{{ formatTime(entry.executed_at) }}</span>
                  </div>
                  <div class="text-xs text-gray-300 break-all line-clamp-2 font-mono opacity-80 group-hover:opacity-100">
                    {{ entry.request.url }}
                  </div>
                </div>
              </div>
              <div v-else class="px-4 py-10 opacity-40">
                  <n-empty size="small" description="No history yet" />
              </div>
            </n-scrollbar>
          </div>
        </div>
      </div>
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

          <n-divider title-placement="left" style="margin-top: 24px; margin-bottom: 12px;">
            <span class="text-[10px] text-gray-500 uppercase tracking-widest">Maintenance Zone</span>
          </n-divider>

          <n-button 
            type="error" 
            ghost 
            block 
            @click="handlePurge"
            class="mt-2"
          >
            <template #icon>
              <n-icon :component="TrashOutline" />
            </template>
            Purge Deleted Files (清理已删除接口)
          </n-button>
        </n-space>
      </n-card>
    </n-modal>

    <!-- Context Menu -->
    <n-dropdown
      placement="bottom-start"
      trigger="manual"
      :x="x"
      :y="y"
      :options="dropdownOptions"
      :show="showDropdown"
      :on-clickoutside="() => showDropdown = false"
      @select="handleSelect"
    />
  </div>
</template>

<style scoped>
.main-layout {
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif;
}

.activity-bar .icon-btn {
  @apply text-gray-500 cursor-pointer p-2 transition-colors duration-200;
}
.activity-bar .icon-btn:hover {
  @apply text-gray-300;
}
.activity-bar .icon-btn.active {
  @apply text-blue-400;
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
.file-item:active {
  @apply opacity-70 scale-[0.98];
}
.file-name {
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: 100%;
  display: block;
}
</style>