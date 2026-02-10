<script setup lang="ts">
import { computed } from 'vue';

const props = defineProps<{
    auth: {
        type: string;
        data: Record<string, string>;
    };
    showInherit?: boolean; // Option to show "Inherit" type (e.g. for requests but not for project root)
}>();

const emit = defineEmits<{
    (e: 'update:auth', value: any): void;
}>();

// Computed for easier binding
const authModel = computed({
    get: () => props.auth,
    set: (val) => emit('update:auth', val)
});

// Options for the select
const authOptions = computed(() => {
    const options = [
        { label: 'No Auth', value: 'noauth' },
        { label: 'Bearer Token', value: 'bearer' },
        { label: 'Basic Auth', value: 'basic' },
        { label: 'API Key', value: 'apikey' }
    ];
    if (props.showInherit) {
        options.unshift({ label: 'Inherit from parent', value: 'inherit' });
    }
    return options;
});

// Handle type change to clear/init data
const handleTypeChange = (newType: string) => {
    authModel.value.type = newType;
    if (newType !== 'inherit') {
        authModel.value.data = {};
    }
    
    // Initialize default keys for better UX
    if (newType === 'apikey') {
        authModel.value.data = { key: '', value: '', addTo: 'header' };
    } else if (newType === 'basic') {
        authModel.value.data = { username: '', password: '' };
    } else if (newType === 'bearer') {
        authModel.value.data = { token: '' };
    }
};

const apiKeyLocationOptions = [
    { label: 'Header', value: 'header' },
    { label: 'Query Params', value: 'query' }
];
</script>

<template>
    <div class="flex flex-col gap-4 p-4 h-full overflow-y-auto">
        <!-- Auth Type Selector -->
        <div class="flex items-center gap-4">
            <span class="w-24 text-gray-500">Type</span>
            <n-select 
                v-model:value="authModel.type" 
                :options="authOptions" 
                @update:value="handleTypeChange"
                class="w-full max-w-xs"
            />
        </div>

        <n-divider />

        <!-- Dynamic Form -->
        <div class="flex-1">
            <!-- Inherit -->
            <div v-if="authModel.type === 'inherit'" class="text-gray-400 italic">
                <p>Inheriting authorization from parent configuration.</p>
                <p class="text-xs mt-2 opacity-70">
                    Resolution Order: Main Case (Folder) > Project Settings
                </p>
            </div>

            <!-- No Auth -->
            <div v-if="authModel.type === 'noauth'" class="text-gray-400 italic">
                This request does not use any authorization.
            </div>

            <!-- Bearer Token -->
            <div v-if="authModel.type === 'bearer'" class="flex flex-col gap-3">
                <div class="flex items-center gap-4">
                    <span class="w-24 text-gray-500">Token</span>
                    <n-input 
                        v-model:value="authModel.data.token" 
                        type="textarea"
                        placeholder="e.g. eyJhbGciOi..." 
                        autosize
                    />
                </div>
            </div>

            <!-- Basic Auth -->
            <div v-if="authModel.type === 'basic'" class="flex flex-col gap-3">
                <div class="flex items-center gap-4">
                    <span class="w-24 text-gray-500">Username</span>
                    <n-input v-model:value="authModel.data.username" placeholder="Username" />
                </div>
                <div class="flex items-center gap-4">
                    <span class="w-24 text-gray-500">Password</span>
                    <n-input 
                        v-model:value="authModel.data.password" 
                        type="password" 
                        show-password-on="click"
                        placeholder="Password" 
                    />
                </div>
            </div>

            <!-- API Key -->
            <div v-if="authModel.type === 'apikey'" class="flex flex-col gap-3">
                <div class="flex items-center gap-4">
                    <span class="w-24 text-gray-500">Key</span>
                    <n-input v-model:value="authModel.data.key" placeholder="Key" />
                </div>
                <div class="flex items-center gap-4">
                    <span class="w-24 text-gray-500">Value</span>
                    <n-input v-model:value="authModel.data.value" placeholder="Value" />
                </div>
                <div class="flex items-center gap-4">
                    <span class="w-24 text-gray-500">Add to</span>
                    <n-select 
                        v-model:value="authModel.data.addTo" 
                        :options="apiKeyLocationOptions" 
                    />
                </div>
            </div>
        </div>
    </div>
</template>
