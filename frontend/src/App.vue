<script setup lang="ts">
import { onMounted, onUnmounted, computed } from 'vue'
import { darkTheme, zhCN, dateZhCN, enUS, dateEnUS } from 'naive-ui'
import MainLayout from './components/MainLayout.vue'
import { useRequestStore } from './stores/request'
import { useAppStore } from './stores/app'

const store = useRequestStore()
const appStore = useAppStore()

const naiveLocale = computed(() => {
  return appStore.language === 'zh-CN' ? zhCN : enUS
})

const naiveDateLocale = computed(() => {
  return appStore.language === 'zh-CN' ? dateZhCN : dateEnUS
})

const handleGlobalShortcuts = async (e: KeyboardEvent) => {
  const isCmdOrCtrl = e.metaKey || e.ctrlKey
  const target = e.target as HTMLElement
  const isInputActive = ['INPUT', 'TEXTAREA'].includes(target.tagName) || target.isContentEditable

  // 1. Cmd+R / Ctrl+R: Sync Swagger (Global, but check loading)
  if (isCmdOrCtrl && e.key.toLowerCase() === 'r') {
    e.preventDefault()
    if (store.isLoading) return
    if (!store.swaggerUrl) await store.loadProjectConfig()

    if (!store.swaggerUrl) {
      // @ts-ignore
      window.$message?.warning?.('Swagger URL not configured. Please open Sync dialog first.')
      return
    }
    store.importSwagger(store.swaggerUrl)
    return
  }

  // 2. Cmd+F / Ctrl+F: Focus Search (Global)
  if (isCmdOrCtrl && e.key.toLowerCase() === 'f') {
    e.preventDefault()
    const inputEl = document.querySelector('#sidebar-search-input input') as HTMLInputElement
    if (inputEl) {
      inputEl.focus()
      inputEl.select()
    }
    return
  }

  // --- Shortcuts that are BLOCKED if user is typing in an input ---
  if (isInputActive) return

  // 3. ArrowUp / ArrowDown: Navigation
  if (e.key === 'ArrowUp' || e.key === 'ArrowDown') {
    e.preventDefault()
    if (e.key === 'ArrowUp') {
      store.selectPrevFile()
    } else {
      store.selectNextFile()
    }
    setTimeout(() => {
      const el = document.getElementById(`file-item-${store.currentFileName}`)
      if (el) el.scrollIntoView({ block: 'nearest', behavior: 'smooth' })
    }, 50)
    return
  }

  // 4. Cmd+D / Ctrl+D: Duplicate File
  if (isCmdOrCtrl && e.key.toLowerCase() === 'd') {
    e.preventDefault()
    if (store.currentFileName) {
      await store.duplicateFile(store.currentFileName)
    }
    return
  }

  // 5. Cmd+Backspace / Cmd+Delete: Delete File
  if (isCmdOrCtrl && (e.key === 'Backspace' || e.key === 'Delete')) {
    e.preventDefault()
    if (!store.currentFileName) return

    // @ts-ignore
    if (window.$dialog) {
      // @ts-ignore
      window.$dialog.warning({
        title: 'Delete File',
        content: `Are you sure you want to delete "${store.currentFileName}"?`,
        positiveText: 'Delete',
        negativeText: 'Cancel',
        positiveButtonProps: { type: 'error' },
        onPositiveClick: async () => {
          await store.deleteFileByFilename(store.currentFileName)
        }
      })
    }
    return
  }
}

onMounted(() => {
  window.addEventListener('keydown', handleGlobalShortcuts)
})

onUnmounted(() => {
  window.removeEventListener('keydown', handleGlobalShortcuts)
})
</script>

<template>
  <n-config-provider 
    :theme="darkTheme" 
    :locale="naiveLocale" 
    :date-locale="naiveDateLocale"
  >
    <n-global-style/>
    <n-message-provider>
      <n-dialog-provider>
        <MainLayout/>
      </n-dialog-provider>
    </n-message-provider>
  </n-config-provider>
</template>