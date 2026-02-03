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
        activeEnvName: 'dev',
        variables: {} as Record<string, string>,
        // Internal storage for all environments
        allEnvs: {
            'dev': { variables: {} },
            'prod': { variables: {} }
        } as Record<string, Environment>
    }),

    getters: {
        envList: (state) => Object.keys(state.allEnvs),
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
                    this.activeEnvName = config.activeEnvName || 'dev';
                    
                    // Ensure active environment exists in the list
                    if (!this.allEnvs[this.activeEnvName]) {
                        // fallback if activeEnvName from file is invalid
                        const keys = Object.keys(this.allEnvs);
                        if (keys.length > 0) {
                            this.activeEnvName = keys[0];
                        } else {
                            this.initDefaults();
                        }
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
            if (this.allEnvs[name]) {
                this.activeEnvName = name;
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
        }
    }
});
