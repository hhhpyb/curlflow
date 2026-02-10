<script setup lang="ts">
import { ref, computed, watch } from 'vue';
import { 
  NModal, NTabs, NTabPane, NForm, NFormItem, NInput, 
  NButton, NIcon, NSelect, NDynamicInput, NEmpty,
  useMessage, useDialog, NDivider, NInputNumber, NSwitch
} from 'naive-ui';
import { Add as AddIcon, Trash as TrashIcon, SettingsOutline, ShieldCheckmarkOutline, ListOutline, GlobeOutline } from '@vicons/ionicons5';
import { useRequestStore } from '../stores/request';
import { useEnvStore } from '../stores/env';
import { useSettingsStore } from '../stores/settings';
import { useAppStore } from '../stores/app';
import AuthEditor from './AuthEditor.vue';
import { useI18n } from 'vue-i18n';

const props = defineProps<{
  show: boolean;
}>();

const emit = defineEmits<{
  (e: 'update:show', value: boolean): void;
}>();

const { t } = useI18n();
const store = useRequestStore();
const envStore = useEnvStore();
const settingsStore = useSettingsStore();
const appStore = useAppStore();
const message = useMessage();
const dialog = useDialog();

const internalShow = computed({
  get: () => props.show,
  set: (val) => emit('update:show', val),
});

// --- Tab States ---
const projectData = ref({ 
  name: '',
  swagger_url: '',
  proxy_url: '',
  description: '',
  auth: { type: 'noauth', data: {} }
});

const appSettings = ref({
  proxyUrl: '',
  insecure: false,
  timeout: 30
});

const currentEnv = ref<string>('');
const kvPairs = ref<{ key: string; value: string }[]>([]);
const showCreateEnv = ref(false);
const newEnvName = ref('');

// --- Initialization ---
watch(() => props.show, async (visible) => {
  if (visible) {
    // 1. Load Project Config
    await store.loadProjectConfig();
    const config = JSON.parse(JSON.stringify(store.projectConfig));
    if (!config.auth || !config.auth.type) {
      config.auth = { type: 'noauth', data: {} };
    }
    projectData.value = config;

    // 2. Load Environments
    syncEnvFromStore();

    // 3. Load Global Settings
    await settingsStore.load();
    appSettings.value = { ...settingsStore.$state };
  }
});

// --- Environments Logic ---
const envOptions = computed(() => {
  return envStore.envList.map((name) => ({ label: name, value: name }));
});

function syncEnvFromStore() {
  currentEnv.value = envStore.activeEnvName;
  const currentVars = envStore.variables;
  kvPairs.value = Object.entries(currentVars).map(([k, v]) => ({
    key: k,
    value: v,
  }));
}

function saveKvToStore(envName: string) {
  if (!envName) return;
  if (envStore.allEnvs[envName]) {
     const newVars: Record<string, string> = {};
     kvPairs.value.forEach(item => {
       if (item.key) {
         newVars[item.key] = item.value;
       }
     });
     envStore.allEnvs[envName].variables = newVars;
     if (envName === envStore.activeEnvName) {
       envStore.updateCurrentVariables();
     }
  }
}

function handleEnvChange(val: string) {
  saveKvToStore(envStore.activeEnvName);
  envStore.setActiveEnv(val);
  syncEnvFromStore();
}

function handleCreateEnv() {
  const name = newEnvName.value.trim();
  if (!name) return;
  if (envStore.envList.includes(name)) {
    message.warning('Environment already exists');
    return;
  }
  saveKvToStore(currentEnv.value);
  envStore.createEnv(name);
  envStore.setActiveEnv(name);
  newEnvName.value = '';
  showCreateEnv.value = false;
  syncEnvFromStore();
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
      envStore.deleteEnv(toDelete);
      syncEnvFromStore();
      message.success(`Deleted ${toDelete}`);
    }
  });
}

// --- Final Save ---
const handleFinalSave = async () => {
  saveKvToStore(currentEnv.value);
  try {
    await Promise.all([
      store.saveProjectConfig(projectData.value as any),
      envStore.saveEnvs(),
      settingsStore.save(appSettings.value)
    ]);
    message.success('All settings saved');
    internalShow.value = false;
  } catch (e) {
    message.error('Failed to save settings');
  }
};
</script>

<template>
  <n-modal
    v-model:show="internalShow"
    preset="card"
    title="Project Settings"
    style="width: 700px"
    :bordered="false"
    size="huge"
  >
    <n-tabs type="line" animated>
      <!-- Tab 1: General -->
      <n-tab-pane name="general" tab="General">
        <template #tab>
          <div class="flex items-center gap-2">
            <n-icon><SettingsOutline /></n-icon> General
          </div>
        </template>
        <n-form label-placement="left" label-width="120" class="mt-4">
          <n-form-item label="Project Name">
            <n-input v-model:value="projectData.name" placeholder="My Awesome Project" />
          </n-form-item>
          <n-form-item label="Swagger URL">
            <n-input v-model:value="projectData.swagger_url" placeholder="https://api.example.com/swagger.json" />
            <template #feedback>
              URL used for syncing requests from OpenAPI/Swagger.
            </template>
          </n-form-item>
          <n-form-item label="Proxy URL">
            <n-input v-model:value="projectData.proxy_url" placeholder="http://localhost:7890" />
            <template #feedback>
              Project-specific proxy server.
            </template>
          </n-form-item>
          <n-form-item label="Description">
            <n-input 
              v-model:value="projectData.description" 
              type="textarea" 
              placeholder="Brief description of the project" 
              :autosize="{ minRows: 2 }"
            />
          </n-form-item>
          <n-form-item :label="t('common.language')">
            <n-select
              :value="appStore.language"
              :options="[
                { label: 'English', value: 'en-US' },
                { label: '简体中文', value: 'zh-CN' }
              ]"
              @update:value="appStore.setLanguage"
            />
          </n-form-item>
        </n-form>
      </n-tab-pane>

      <!-- Tab 2: Global Auth -->
      <n-tab-pane name="auth" tab="Global Auth">
        <template #tab>
          <div class="flex items-center gap-2">
            <n-icon><ShieldCheckmarkOutline /></n-icon> Global Auth
          </div>
        </template>
        <div class="mt-2 h-[400px] flex flex-col">
          <div class="mb-4 p-3 bg-blue-500/10 border border-blue-500/20 rounded text-xs text-blue-300">
            This configuration is stored in project.json and serves as the root-level authorization for the entire project.
            Requests can inherit these settings by choosing "Inherit from parent".
          </div>
          <div class="border border-gray-700 rounded flex-1 min-h-0 bg-gray-800/30">
             <AuthEditor v-model:auth="projectData.auth" :show-inherit="false" />
          </div>
        </div>
      </n-tab-pane>

      <!-- Tab 3: Environments -->
      <n-tab-pane name="envs" tab="Environments">
        <template #tab>
          <div class="flex items-center gap-2">
            <n-icon><ListOutline /></n-icon> Environments
          </div>
        </template>
        <div class="flex flex-col gap-4 mt-4">
          <div class="flex items-center gap-3">
            <n-select
              v-model:value="currentEnv"
              :options="envOptions"
              placeholder="Select Environment"
              class="flex-1"
              @update:value="handleEnvChange"
            />
            <n-button @click="showCreateEnv = true">
              <template #icon><n-icon><AddIcon /></n-icon></template>
              New
            </n-button>
            <n-button type="error" ghost @click="handleDeleteEnv" :disabled="!currentEnv">
              <template #icon><n-icon><TrashIcon /></n-icon></template>
            </n-button>
          </div>

          <div class="border border-gray-700 rounded p-4 h-[300px] overflow-y-auto bg-gray-800/50">
            <n-empty v-if="!currentEnv" description="No environment selected" />
            <n-dynamic-input
              v-else
              v-model:value="kvPairs"
              preset="pair"
              key-placeholder="Variable Key"
              value-placeholder="Value"
            />
          </div>
        </div>
      </n-tab-pane>

      <!-- Tab 4: App Settings -->
      <n-tab-pane name="app" tab="App Settings">
        <template #tab>
          <div class="flex items-center gap-2">
            <n-icon><GlobeOutline /></n-icon> App Settings
          </div>
        </template>
        <n-form label-placement="left" label-width="120" class="mt-4">
          <n-form-item label="Global Proxy">
            <n-input v-model:value="appSettings.proxyUrl" placeholder="http://localhost:7890" />
          </n-form-item>
          <n-form-item label="Timeout (s)">
            <n-input-number v-model:value="appSettings.timeout" :min="1" :max="300" class="w-full" />
          </n-form-item>
          <n-form-item label="Insecure SSL">
            <n-switch v-model:value="appSettings.insecure" />
          </n-form-item>
        </n-form>
      </n-tab-pane>
    </n-tabs>

    <template #footer>
      <div class="flex justify-end gap-2">
        <n-button @click="internalShow = false">{{ t('common.cancel') }}</n-button>
        <n-button type="primary" @click="handleFinalSave">{{ t('common.save') }}</n-button>
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
    <n-input v-model:value="newEnvName" placeholder="Environment Name" @keyup.enter="handleCreateEnv" />
  </n-modal>
</template>