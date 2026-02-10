import { defineStore } from 'pinia';
import { GetSettings, SaveSettings } from '../../wailsjs/go/main/App';
import { main } from '../../wailsjs/go/models';

export const useSettingsStore = defineStore('settings', {
    state: () => ({
        proxyUrl: '',
        insecure: false,
        timeout: 30,
    }),
    actions: {
        async load() {
            try {
                const cfg = await GetSettings();
                this.proxyUrl = cfg.proxyUrl;
                this.insecure = cfg.insecure;
                this.timeout = cfg.timeout;
            } catch (e) {
                console.error('Failed to load settings:', e);
            }
        },
        async save(cfgData?: { proxyUrl: string, insecure: boolean, timeout: number }) {
            try {
                if (cfgData) {
                    this.proxyUrl = cfgData.proxyUrl;
                    this.insecure = cfgData.insecure;
                    this.timeout = cfgData.timeout;
                }
                const cfg = new main.AppConfig({
                    proxyUrl: this.proxyUrl,
                    insecure: this.insecure,
                    timeout: this.timeout
                });
                await SaveSettings(cfg);
            } catch (e) {
                console.error('Failed to save settings:', e);
                throw e;
            }
        }
    }
});
