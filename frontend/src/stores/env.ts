import { defineStore } from 'pinia';
import { useRequestStore } from './request';
import { LoadConfig, SaveConfig } from '../../wailsjs/go/main/App';

interface Environment {
    variables: Record<string, string>;
}

interface EnvConfig {
    activeEnvName: string;
    environments: Record<string, Environment>;
}

export const useEnvStore = defineStore('env', {
    state: () => ({
        activeEnvName: '',
        variables: {} as Record<string, string>,
        // Internal storage for all environments
        allEnvs: {
            'dev': { variables: {} },
            'prod': { variables: {} }
        } as Record<string, Environment>
    }),

    getters: {
        envList: (state) => Object.keys(state.allEnvs),
        envOptions: (state) => {
            const options = Object.keys(state.allEnvs).map(name => ({
                label: name,
                value: name
            }));
            return [{ label: 'No Environment', value: '' }, ...options];
        }
    },

    actions: {
        async loadEnvs() {
            const requestStore = useRequestStore();
            if (!requestStore.workDir) {
                // Initialize defaults if no workDir
                this.initDefaults();
                return;
            }

            try {
                const configStr = await LoadConfig(requestStore.workDir, "environments.json");
                if (configStr) {
                    const config = JSON.parse(configStr) as EnvConfig;
                    this.allEnvs = config.environments || {};
                    
                    // Restore from localStorage if valid
                    const savedEnv = localStorage.getItem('curlflow_activeEnv');
                    if (savedEnv !== null && (savedEnv === '' || this.allEnvs[savedEnv])) {
                        this.activeEnvName = savedEnv;
                    } else {
                        this.activeEnvName = config.activeEnvName || '';
                    }
                    
                    // Final safety check: if current activeEnvName is invalid and not empty
                    if (this.activeEnvName !== '' && !this.allEnvs[this.activeEnvName]) {
                        const keys = Object.keys(this.allEnvs);
                        this.activeEnvName = keys.length > 0 ? keys[0] : '';
                    }
                } else {
                    this.initDefaults();
                }
            } catch (e) {
                // File likely doesn't exist, use defaults
                console.log('environments.json not found or invalid, using defaults');
                this.initDefaults();
            }
            this.updateCurrentVariables();
        },

        initDefaults() {
            this.allEnvs = {
                'dev': { variables: { 'base_url': 'http://localhost:8080' } },
                'prod': { variables: { 'base_url': 'https://api.example.com' } }
            };
            this.activeEnvName = 'dev';
            this.updateCurrentVariables();
        },

        async saveEnvs() {
            const requestStore = useRequestStore();
            if (!requestStore.workDir) return;

            const fullConfig: EnvConfig = {
                activeEnvName: this.activeEnvName,
                environments: this.allEnvs
            };

            try {
                await SaveConfig(requestStore.workDir, "environments.json", JSON.stringify(fullConfig, null, 2));
            } catch (e) {
                console.error('Failed to save environments:', e);
            }
        },

        updateCurrentVariables() {
            if (this.allEnvs[this.activeEnvName]) {
                this.variables = this.allEnvs[this.activeEnvName].variables;
            } else {
                this.variables = {};
            }
        },

        setVariable(key: string, value: string) {
            if (!this.allEnvs[this.activeEnvName]) {
                this.allEnvs[this.activeEnvName] = { variables: {} };
            }
            this.allEnvs[this.activeEnvName].variables[key] = value;
            this.updateCurrentVariables();
        },

        getVariable(key: string): string {
            return this.variables[key] || '';
        },

        setActiveEnv(name: string) {
            if (name === '' || this.allEnvs[name]) {
                this.activeEnvName = name;
                localStorage.setItem('curlflow_activeEnv', name);
                this.updateCurrentVariables();
            }
        },

        createEnv(name: string) {
            if (!this.allEnvs[name]) {
                this.allEnvs[name] = { variables: {} };
                // Optionally switch to it? No, just create.
            }
        },

        deleteEnv(name: string) {
            if (this.allEnvs[name]) {
                delete this.allEnvs[name];
                if (this.activeEnvName === name) {
                    const keys = Object.keys(this.allEnvs);
                    this.activeEnvName = keys.length > 0 ? keys[0] : '';
                    this.updateCurrentVariables();
                }
            }
        },

        processString(input: string): string {
            if (!input) return input;
            // Regex to find {{ variableName }}
            return input.replace(/{{\s*(\w+)\s*}}/g, (match, key) => {
                const val = this.variables[key];
                // If variable exists, replace it. Otherwise keep the original {{key}} string.
                return val !== undefined ? val : match;
            });
        },

        // Reverse check: Find values in the string that match current environment variables
        // Returns the string with values replaced by {{key}}, or null if no changes needed
        reverseReplace(input: string): string | null {
            if (!input) return null;
            
            let result = input;
            let changed = false;
            
            // Sort variables by value length (descending) to avoid partial replacements
            // e.g. if we have var1="http://api" and var2="http://api/v2", we should match var2 first
            const entries = Object.entries(this.variables)
                .filter(([_, val]) => val && val.length > 2) // Ignore very short values to avoid false positives
                .sort((a, b) => b[1].length - a[1].length);

            for (const [key, val] of entries) {
                if (result.includes(val)) {
                    // Replace all occurrences
                    // Escape regex special characters in 'val'
                    const escapedVal = val.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
                    const regex = new RegExp(escapedVal, 'g');
                    result = result.replace(regex, `{{${key}}}`);
                    changed = true;
                }
            }
            
            return changed ? result : null;
        }
    }
});
