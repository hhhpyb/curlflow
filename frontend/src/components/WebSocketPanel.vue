<script setup lang="ts">
import { computed, nextTick, onMounted, ref, watch } from 'vue';
import { useWebSocketStore } from '../stores/websocket';
import { useRequestStore } from '../stores/request';
import { NButton, NIcon, NScrollbar, NSplit, NInput } from 'naive-ui';
import {
  PaperPlaneOutline, CloudOfflineOutline, CloudDoneOutline,
  CodeSlashOutline, TimeOutline, TrashOutline, ArrowDownOutline
} from '@vicons/ionicons5';
import CodeEditor from './CodeEditor.vue';

const props = defineProps<{
  requestId: string; // Unique Session ID
  request: any; // Reactive Request Object from parent
}>();

const wsStore = useWebSocketStore();
const requestStore = useRequestStore();

// UI State
const autoScroll = ref(true);
const virtualListRef = ref<any>(null);
const splitSize = ref(0.7); // 70% messages, 30% composer
const showNewMsgIndicator = ref(false);

// Get current session state from store
const session = computed(() => wsStore.getSession(props.requestId));

// Sync input with request body for persistence
// request.body is our "Draft"
const messageInput = computed({
  get: () => props.request.body || '',
  set: (val) => {
    props.request.body = val;
    // Trigger sync to curl for auto-save detection?
    // Not strictly necessary for curl generation but good for dirty check
    requestStore.syncToCurl(); 
  }
});

// Watch for incoming messages
watch(() => session.value?.messages.length, (newLen, oldLen) => {
  if (autoScroll.value) {
    scrollToBottom();
  } else {
    // Show indicator if user is scrolled up and new message arrives
    if (newLen && oldLen && newLen > oldLen) {
      showNewMsgIndicator.value = true;
    }
  }
});

const scrollToBottom = () => {
  nextTick(() => {
    if (virtualListRef.value) {
        // Support both scrollbar and potential virtual list API
        if (typeof virtualListRef.value.scrollTo === 'function') {
            virtualListRef.value.scrollTo({ 
                top: 99999, 
                position: 'bottom', // For some virtual list components
                behavior: 'smooth' 
            });
        }
    }
    showNewMsgIndicator.value = false;
  });
};

const handleScroll = (e: Event) => {
  const target = e.target as HTMLElement;
  const isAtBottom = target.scrollHeight - target.scrollTop - target.clientHeight < 50;
  autoScroll.value = isAtBottom;
  if (isAtBottom) showNewMsgIndicator.value = false;
};

// 编辑器加载完成后的配置
const handleMount = (editor: any, monaco: any) => {
  // 添加发送消息的命令
  // KeyCode.Enter = 3
  editor.addCommand(monaco.KeyCode.Enter, () => {
    handleSend();
  }, '!suggestWidgetVisible && !findWidgetVisible');

  // 允许 Shift+Enter 换行
  editor.addCommand(monaco.KeyMod.Shift | monaco.KeyCode.Enter, () => {
    const position = editor.getPosition();
    editor.executeEdits('', [
      {
        range: new monaco.Range(position.lineNumber, position.column, position.lineNumber, position.column),
        text: '\n',
        forceMoveMarkers: true,
      },
    ]);
  });
};

const handleSend = async () => {
  if (!messageInput.value.trim()) return;
  await wsStore.sendMessage(props.requestId, messageInput.value);
  
  // Force scroll to bottom after sending
  scrollToBottom();
};

const handleFormatJSON = () => {
  try {
    const obj = JSON.parse(messageInput.value);
    messageInput.value = JSON.stringify(obj, null, 2);
  } catch (e) {
    // Ignore
  }
};

const formatTime = (ts: number) => {
  return new Date(ts).toLocaleTimeString([], { hour12: false, hour: '2-digit', minute: '2-digit', second: '2-digit' });
};

const clearLogs = () => {
  if (session.value) {
    session.value.messages = [];
  }
};

onMounted(() => {
  wsStore.setupGlobalListener();
  wsStore.initSession(props.requestId);
  // Initial scroll
  scrollToBottom();
});
</script>

<template>
  <div class="ws-panel flex flex-col h-full bg-[#1e1e1e] relative">
    
    <n-split direction="vertical" v-model:size="splitSize" :min="0.2" :max="0.8">
      
      <!-- Top: Message Stream -->
      <template #1>
        <div class="h-full flex flex-col relative bg-[#1e1e1e]">
          <n-scrollbar 
            ref="virtualListRef" 
            class="px-4 py-2" 
            @scroll="handleScroll"
          >
            <div v-if="!session?.messages.length" class="text-center text-gray-600 mt-20 text-sm select-none">
              <div class="mb-2 opacity-50"><n-icon size="40" :component="CloudOfflineOutline" /></div>
              <div>Ready to connect</div>
            </div>

            <div class="flex flex-col gap-4 pb-4">
              <div 
                v-for="(msg, idx) in session?.messages" 
                :key="idx"
                class="flex flex-col max-w-[85%]"
                :class="{
                  'self-end items-end': msg.direction === 'sent',
                  'self-start items-start': msg.direction === 'received',
                  'self-center items-center max-w-full w-full': ['system', 'error'].includes(msg.direction)
                }"
              >
                <!-- System / Error -->
                <div v-if="['system', 'error'].includes(msg.direction)" 
                     class="text-xs py-1 px-3 rounded-full bg-opacity-20 flex items-center gap-2 my-2 select-text"
                     :class="msg.direction === 'error' ? 'bg-red-500 text-red-400' : 'bg-gray-500 text-gray-400'"
                >
                  <n-icon :component="msg.direction === 'error' ? CloudOfflineOutline : TimeOutline" />
                  <span>{{ msg.content }}</span>
                  <span class="opacity-50 ml-2 font-mono text-[10px]">{{ formatTime(msg.timestamp) }}</span>
                </div>

                <!-- Chat Bubbles -->
                <template v-else>
                  <div class="flex items-end gap-2 mb-1 opacity-60 text-[10px] select-none">
                     <span v-if="msg.direction === 'sent'">You</span>
                     <span v-else>Server</span>
                     <span>{{ formatTime(msg.timestamp) }}</span>
                  </div>
                  <div 
                    class="px-3 py-2 rounded-2xl text-sm font-mono break-all whitespace-pre-wrap shadow-sm select-text relative group"
                    :class="msg.direction === 'sent' 
                      ? 'bg-[#0f5324] text-[#d1e7dd] rounded-tr-sm' 
                      : 'bg-[#2a2d35] text-[#e2e8f0] rounded-tl-sm'"
                  >
                    {{ msg.content }}
                  </div>
                </template>
              </div>
            </div>
          </n-scrollbar>

          <!-- Overlay Tools -->
          <div class="absolute top-2 right-4 z-10">
            <n-button circle size="tiny" secondary type="error" @click="clearLogs" title="Clear Logs">
              <template #icon><n-icon :component="TrashOutline" /></template>
            </n-button>
          </div>

          <!-- New Message Toast -->
          <div 
            v-if="showNewMsgIndicator"
            class="absolute bottom-4 left-1/2 transform -translate-x-1/2 bg-blue-600 text-white px-3 py-1 rounded-full shadow-lg text-xs cursor-pointer flex items-center gap-1 hover:bg-blue-500 transition-colors z-20"
            @click="scrollToBottom"
          >
            <n-icon :component="ArrowDownOutline" />
            New messages
          </div>
        </div>
      </template>

      <!-- Bottom: Composer -->
      <template #2>
        <div class="h-full bg-[#252526] border-t border-gray-800 flex flex-col">
          <!-- Toolbar -->
          <div class="flex items-center justify-between px-2 py-1 bg-[#2d2d30] border-b border-gray-800 shrink-0">
            <span class="text-xs font-bold text-gray-500 uppercase tracking-wider ml-1">Payload</span>
            <div class="flex gap-1">
               <n-button text size="tiny" @click="handleFormatJSON" class="text-blue-400 hover:text-blue-300 mr-2" title="Format JSON">
                <template #icon><n-icon :component="CodeSlashOutline" /></template>
              </n-button>
              <n-button type="primary" size="small" @click="handleSend" :disabled="!session?.isConnected">
                <template #icon><n-icon :component="PaperPlaneOutline" /></template>
                Send
              </n-button>
            </div>
          </div>
          
          <!-- Editor -->
          <div class="flex-1 min-h-0 relative">
             <CodeEditor
                v-model:model-value="messageInput"
                language="json"
                height="100%"
                @mount="handleMount"
              />
          </div>
        </div>
      </template>

    </n-split>
  </div>
</template>

<style scoped>
:deep(.n-split-pane-1) {
  overflow: hidden;
}
:deep(.n-split-pane-2) {
  overflow: hidden;
}
</style>