<template>
  <n-modal
    v-model:show="internalShow"
    preset="card"
    title="Environment Manager"
    style="width: 600px"
    :bordered="false"
    size="huge"
    role="dialog"
    aria-modal="true"
  >
    <div class="flex flex-col gap-4">
      <!-- Top Bar: Select Env + New Env -->
      <div class="flex items-center gap-3">
        <n-select
          v-model:value="currentEnv"
          :options="envOptions"
          placeholder="Select Environment"
          class="flex-1"
          @update:value="handleEnvChange"
        />
        <n-button @click="showCreateEnv = true">
          <template #icon>
            <n-icon><add-icon /></n-icon>
          </template>
          New
        </n-button>
        <n-button type="error" ghost @click="handleDeleteEnv" :disabled="!currentEnv">
          <template #icon>
             <n-icon><trash-icon /></n-icon>
          </template>
        </n-button>
      </div>

      <!-- Variables Editor -->
      <div class="border border-gray-200 dark:border-gray-700 rounded p-4 max-h-[400px] overflow-y-auto">
        <n-empty v-if="!currentEnv" description="No environment selected" />
        <n-dynamic-input
          v-else
          v-model:value="kvPairs"
          preset="pair"
          key-placeholder="Variable Key (e.g. base_url)"
          value-placeholder="Value (e.g. http://localhost)"
        />
      </div>
    </div>

    <!-- Footer Actions -->
    <template #footer>
      <div class="flex justify-end gap-2">
        <n-button @click="close">Cancel</n-button>
        <n-button type="primary" @click="handleSave">Save & Close</n-button>
      </div>
    </template>
  </n-modal>

  <!-- Create New Environment Dialog -->
  <n-modal
    v-model:show="showCreateEnv"
    preset="dialog"
    title="Create New Environment"
    positive-text="Create"
    negative-text="Cancel"
    @positive-click="handleCreateEnv"
    @negative-click="showCreateEnv = false"
  >
    <n-input
      v-model:value="newEnvName"
      placeholder="Environment Name (e.g. staging)"
      @keyup.enter="handleCreateEnv"
    />
  </n-modal>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue';
import { useEnvStore } from '../stores/env';
import { 
  NModal, NSelect, NButton, NIcon, NDynamicInput, NInput, NEmpty, useMessage, useDialog
} from 'naive-ui';
import { Add as AddIcon, Trash as TrashIcon } from '@vicons/ionicons5';

const props = defineProps<{
  show: boolean;
}>();

const emit = defineEmits<{
  (e: 'update:show', value: boolean): void;
}>();

const store = useEnvStore();
const message = useMessage();
const dialog = useDialog();

// Internal state for modal visibility
const internalShow = computed({
  get: () => props.show,
  set: (val) => emit('update:show', val),
});

// UI State
const currentEnv = ref<string>('');
const kvPairs = ref<{ key: string; value: string }[]>([]);
const showCreateEnv = ref(false);
const newEnvName = ref('');

// Computed options for Select
const envOptions = computed(() => {
  return store.envList.map((name) => ({ label: name, value: name }));
});

// --- Synchronization Logic ---

// When modal opens or active env changes externally, sync local state
watch(
  () => props.show,
  (visible) => {
    if (visible) {
      syncFromStore();
    }
  }
);

// Also sync when the user changes selection in the dropdown
function handleEnvChange(val: string) {
  // Before switching, we might want to save temp changes to the previous env in store memory?
  // For simplicity, we assume "Save" button commits everything to disk, 
  // but switching dropdown should update the Store's activeEnv immediately 
  // so we can edit the new one.
  
  // First, save current KV pairs to the *previous* env in the store (in memory)
  // so we don't lose edits when just switching dropdowns.
  saveKvToStore(store.activeEnvName);

  // Switch active env in store
  store.setActiveEnv(val);
  
  // Reload KV pairs from the new env
  syncFromStore();
}

function syncFromStore() {
  currentEnv.value = store.activeEnvName;
  const currentVars = store.variables; // this is a reactive Record<string, string>
  
  // Convert Record to Array for Dynamic Input
  kvPairs.value = Object.entries(currentVars).map(([k, v]) => ({
    key: k,
    value: v,
  }));
}

function saveKvToStore(envName: string) {
  if (!envName) return;
  
  // 1. Clear existing vars for this env in store (to handle deletions)
  // Since store.setVariable appends/updates, we need a way to reset.
  // The store doesn't have a "replaceVariables" method, but we can access state directly 
  // or add a method. Accessing state directly is easier here given the Store definition.
  if (store.allEnvs[envName]) {
     const newVars: Record<string, string> = {};
     kvPairs.value.forEach(item => {
       if (item.key) {
         newVars[item.key] = item.value;
       }
     });
     store.allEnvs[envName].variables = newVars;
     // If this is the active env, update the 'variables' shortcut state too
     if (envName === store.activeEnvName) {
       store.updateCurrentVariables();
     }
  }
}

// --- Actions ---

function handleSave() {
  // Save current KV pairs to store memory first
  saveKvToStore(currentEnv.value);
  
  // Persist to disk
  store.saveEnvs()
    .then(() => {
      message.success('Environments saved');
      close();
    })
    .catch(() => {
      message.error('Failed to save environments');
    });
}

function close() {
  internalShow.value = false;
}

function handleCreateEnv() {
  const name = newEnvName.value.trim();
  if (!name) return;
  
  if (store.envList.includes(name)) {
    message.warning('Environment already exists');
    return;
  }

  // Save current work before switching
  saveKvToStore(currentEnv.value);

  store.createEnv(name);
  store.setActiveEnv(name);
  
  newEnvName.value = '';
  showCreateEnv.value = false;
  
  syncFromStore();
  message.success(`Created environment: ${name}`);
}

function handleDeleteEnv() {
  dialog.warning({
    title: 'Confirm Delete',
    content: `Are you sure you want to delete environment "${currentEnv.value}"?`,
    positiveText: 'Delete',
    negativeText: 'Cancel',
    onPositiveClick: () => {
      const toDelete = currentEnv.value;
      store.deleteEnv(toDelete);
      syncFromStore(); // Store auto-switches active env if current is deleted
      message.success(`Deleted ${toDelete}`);
    }
  });
}
</script>
