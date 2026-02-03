<script setup lang="ts">
import { computed, ref } from 'vue'
import { 
    NModal, NCard, NSpace, NInput, NInputNumber, 
    NSwitch, NButton, NFormItem, useMessage 
} from 'naive-ui'
import { useSettingsStore } from '../stores/settings'

const props = defineProps<{
    show: boolean
}>()

const emit = defineEmits<{
    (e: 'update:show', value: boolean): void
}>()

const store = useSettingsStore()
const message = useMessage()
const loading = ref(false)

const sslVerified = computed({
    get: () => !store.insecure,
    set: (val) => store.insecure = !val
})

const handleSave = async () => {
    loading.value = true
    try {
        await store.save()
        message.success('Settings saved')
        emit('update:show', false)
    } catch (e) {
        message.error('Failed to save settings')
    } finally {
        loading.value = false
    }
}
</script>

<template>
    <n-modal 
        :show="show" 
        @update:show="(val) => emit('update:show', val)"
    >
        <n-card
            style="width: 500px"
            title="Settings"
            :bordered="false"
            size="huge"
            role="dialog"
            aria-modal="true"
        >
            <n-space vertical size="large">
                <n-form-item label="Proxy URL">
                    <n-input 
                        v-model:value="store.proxyUrl" 
                        placeholder="http://127.0.0.1:7890" 
                    />
                </n-form-item>
                
                <n-form-item label="Request Timeout (seconds)">
                    <n-input-number 
                        v-model:value="store.timeout" 
                        :min="1" 
                        placeholder="30"
                    />
                </n-form-item>
                
                <n-form-item label="SSL Verification">
                    <n-space align="center">
                        <n-switch v-model:value="sslVerified" />
                        <span class="text-xs text-gray-500">
                            {{ sslVerified ? 'Enabled (Secure)' : 'Disabled (Insecure)' }}
                        </span>
                    </n-space>
                </n-form-item>

                <div class="flex justify-end gap-2 mt-4">
                    <n-button @click="emit('update:show', false)">Cancel</n-button>
                    <n-button type="primary" :loading="loading" @click="handleSave">
                        Save
                    </n-button>
                </div>
            </n-space>
        </n-card>
    </n-modal>
</template>
