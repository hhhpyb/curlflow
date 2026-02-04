<script setup lang="ts">
import {computed, onMounted, onUnmounted, ref} from 'vue'
import {useMessage, NTabs, NTabPane, NDynamicInput, NButton, NIcon, NInput, NModal, NCard, NSpace, NSelect} from 'naive-ui'
import {CloudDownloadOutline, PlayOutline, SaveOutline, SettingsOutline, ListOutline} from '@vicons/ionicons5'
import CodeEditor from './CodeEditor.vue'
import RequestPanel from './RequestPanel.vue'
import CaseBar from './CaseBar.vue'
import Sidebar from './Sidebar.vue'
import EnvManager from './EnvManager.vue'
import ResponsePanel from './ResponsePanel.vue'
import SettingsModal from './SettingsModal.vue'
import {useRequestStore} from '../stores/request'
import {useEnvStore} from '../stores/env'
import {useSettingsStore} from '../stores/settings'

const message = useMessage()
window.$message = message
const store = useRequestStore()
const envStore = useEnvStore()
const settingsStore = useSettingsStore()

// Environment Manager State
const showEnvModal = ref(false)
const showSettingsModal = ref(false)

onMounted(async () => {
  // Load global settings
  settingsStore.load()
  
  const restored = await store.init()
  if (restored) {
    if (store.workDir) {
       await envStore.loadEnvs()
    }
  }
})

// Save Modal State
const showSaveModal = ref(false)
const newFilename = ref('')

// Split pane logic
const requestHeightPercent = ref(50) // Initial height percentage for the top panel
const isDragging = ref(false)
const containerRef = ref<HTMLElement | null>(null)

const startDrag = () => {
  isDragging.value = true
  document.addEventListener('mousemove', onDrag)
  document.addEventListener('mouseup', stopDrag)
  document.body.style.cursor = 'row-resize'
  document.body.style.userSelect = 'none'
}

const onDrag = (e: MouseEvent) => {
  if (!isDragging.value || !containerRef.value) return
  
  const containerRect = containerRef.value.getBoundingClientRect()
  const relativeY = e.clientY - containerRect.top
  const percentage = (relativeY / containerRect.height) * 100
  
  // Limit the resizing range (e.g., between 10% and 90%)
  if (percentage >= 10 && percentage <= 90) {
    requestHeightPercent.value = percentage
  }
}

const stopDrag = () => {
  isDragging.value = false
  document.removeEventListener('mousemove', onDrag)
  document.removeEventListener('mouseup', stopDrag)
  document.body.style.cursor = ''
  document.body.style.userSelect = ''
}

onUnmounted(() => {
  // Cleanup in case component is destroyed while dragging
  document.removeEventListener('mousemove', onDrag)
  document.removeEventListener('mouseup', stopDrag)
})

const handleSave = async () => {
  if (!store.workDir) {
    message.warning("请先选择一个工作目录 (Click 'Open' in Sidebar)")
    return
  }

  // If already has a filename, save directly
  if (store.currentFileName) {
    try {
      await store.saveCurrent()
      message.success("保存成功")
    } catch (e: any) {
      message.error(e.message || "保存失败")
    }
  } else {
    // Open modal for new file
    newFilename.value = ''
    showSaveModal.value = true
  }
}

const confirmSave = async () => {
  if (!newFilename.value.trim()) {
    message.warning("文件名不能为空")
    return
  }
  
  try {
    await store.saveCurrent(newFilename.value)
    message.success("保存成功")
    showSaveModal.value = false
  } catch (e: any) {
    message.error(e.message || "保存失败")
  }
}

const handleRun = async () => {
  if (!store.request.url) {
    message.warning("URL不能为空")
    return
  }

  await store.send()

  if (store.response.error) {
    message.error("请求失败")
  } else {
    message.success(`Status: ${store.response.statusCode} | Time: ${store.response.time}ms`)
  }
}

const handleEnvChange = (val: string) => {
  envStore.setActiveEnv(val)
  envStore.saveEnvs()
}
</script>

<template>
  <div class="h-screen w-screen flex bg-gray-900 text-gray-200 overflow-hidden">
    <Sidebar />
    
    <div class="flex-1 flex flex-col p-4 gap-4 min-w-0">
      <!-- Header -->
      <div class="flex items-center justify-between shrink-0">
        <div class="flex items-center gap-2">
          <n-icon size="24" color="#4ade80">
            <CloudDownloadOutline/>
          </n-icon>
          <span class="font-bold text-lg tracking-wide text-white">CurlFlow</span>
        </div>
        <div class="flex items-center gap-2">
          <n-select
              :value="envStore.activeEnvName"
              :options="envStore.envOptions"
              size="small"
              placeholder="Select Env"
              style="width: 150px"
              @update:value="handleEnvChange"
          />
          <n-button
              secondary
              size="small"
              @click="handleSave"
              class="px-4 text-gray-300"
          >
            <template #icon>
              <n-icon>
                <SaveOutline/>
              </n-icon>
            </template>
            Save
          </n-button>
          <n-button
              secondary
              size="small"
              @click="showEnvModal = true"
              class="px-2 text-gray-300"
              title="Environment Variables"
          >
            <template #icon>
              <n-icon>
                <ListOutline/>
              </n-icon>
            </template>
          </n-button>
          <n-button
              secondary
              size="small"
              @click="showSettingsModal = true"
              class="px-2 text-gray-300"
              title="Global Settings"
          >
            <template #icon>
              <n-icon>
                <SettingsOutline/>
              </n-icon>
            </template>
          </n-button>
          <n-button
              type="primary"
              size="small"
              :loading="store.isLoading"
              @click="handleRun"
              class="px-6"
          >
            <template #icon>
              <n-icon>
                <PlayOutline/>
              </n-icon>
            </template>
            Run
          </n-button>
        </div>
      </div>

      <!-- Main Content Area -->
      <div class="flex-1 flex flex-col min-h-0 relative" ref="containerRef">
        <!-- Request Section (Top) -->
        <div class="flex flex-col gap-2 min-h-0" :style="{ height: `${requestHeightPercent}%` }">
          <div class="text-xs font-bold font-mono text-gray-500 uppercase tracking-widest flex items-center justify-between shrink-0">
            <div class="flex items-center gap-2">
              <div class="w-1.5 h-1.5 rounded-full bg-blue-500"></div>
              Request
            </div>
          </div>
          <div class="flex flex-col flex-1 min-h-0 bg-gray-800 rounded-lg border border-gray-700/50 overflow-hidden">
            <CaseBar v-if="store.meta && store.meta.id" />
            <RequestPanel class="flex-1 min-h-0"/>
          </div>
        </div>

        <!-- Resizer Handle -->
        <div 
          class="h-2 w-full hover:bg-blue-500/50 cursor-row-resize flex items-center justify-center group transition-colors my-1 shrink-0"
          @mousedown="startDrag"
        >
          <div class="w-10 h-1 rounded-full bg-gray-700 group-hover:bg-blue-400 transition-colors"></div>
        </div>

        <!-- Response Section (Bottom) -->
        <div class="flex-1 min-h-0 overflow-hidden mt-1">
          <ResponsePanel class="h-full rounded-lg border border-gray-700/50" />
        </div>
      </div>
    </div>

    <!-- Save Modal -->
    <n-modal v-model:show="showSaveModal">
      <n-card
        style="width: 400px"
        title="保存请求"
        :bordered="false"
        size="huge"
        role="dialog"
        aria-modal="true"
      >
        <template #header-extra>
          <n-icon size="20" class="cursor-pointer" @click="showSaveModal = false">
            <!-- Close icon could go here -->
          </n-icon>
        </template>
        <n-space vertical>
          <n-input 
            v-model:value="newFilename" 
            placeholder="请输入文件名 (例如: my-request.json)" 
            @keyup.enter="confirmSave"
            autofocus
          />
          <div class="flex justify-end gap-2 mt-4">
            <n-button @click="showSaveModal = false">取消</n-button>
            <n-button type="primary" @click="confirmSave">保存</n-button>
          </div>
        </n-space>
      </n-card>
    </n-modal>

    <EnvManager v-model:show="showEnvModal" />
    <SettingsModal v-model:show="showSettingsModal" />
  </div>
</template>

<style scoped>
:deep(.n-tabs-pane-wrapper) {
  flex: 1;
  height: 0;
}
:deep(.n-tab-pane) {
  height: 100%;
}
</style>