import { defineStore } from 'pinia';
import { 
    ParseCurl, 
    BuildCurl, 
    SendRequest,
    SelectWorkDir,
    GetFileList,
    SaveRequest,
    LoadRequest
} from '../../wailsjs/go/main/App';
import { domain } from '../../wailsjs/go/models';

export const useRequestStore = defineStore('request', {
    state: () => ({
        curlCode: '',
        request: new domain.HttpRequest(),
        isLoading: false,
        response: new domain.HttpResponse(),
        
        // File Management State
        workDir: '',
        fileList: [] as string[],
        currentFileName: '',
    }),
    actions: {
        async syncFromCurl() {
            try {
                const req = await ParseCurl(this.curlCode);
                this.request = req;
            } catch (e) {
                console.error('Failed to parse curl:', e);
            }
        },
        async syncToCurl() {
            try {
                const curl = await BuildCurl(this.request);
                this.curlCode = curl;
            } catch (e) {
                console.error('Failed to build curl:', e);
            }
        },
        async send() {
            this.isLoading = true;
            this.response = new domain.HttpResponse();
            try {
                const res = await SendRequest(this.request);
                this.response = res;
            } catch (e) {
                console.error('Request failed:', e);
                this.response.error = String(e);
            } finally {
                this.isLoading = false;
            }
        },

        // ================= File Actions =================

        async chooseDir() {
            try {
                const dir = await SelectWorkDir();
                if (dir) {
                    this.workDir = dir;
                    this.currentFileName = ''; // Reset current file when changing dir
                    await this.fetchFiles();
                }
            } catch (e) {
                console.error('Failed to select directory:', e);
            }
        },

        async fetchFiles() {
            if (!this.workDir) return;
            try {
                const files = await GetFileList(this.workDir);
                this.fileList = files || [];
            } catch (e) {
                console.error('Failed to list files:', e);
                this.fileList = [];
            }
        },

        async saveTo(filename: string) {
            if (!this.workDir) {
                console.warn('No working directory selected');
                return;
            }
            try {
                const savedPath = await SaveRequest(this.workDir, filename, this.request);
                if (savedPath) {
                    // Update current filename only if save successful
                    const actualName = filename.toLowerCase().endsWith('.json') ? filename : filename + '.json';
                    this.currentFileName = actualName;
                    await this.fetchFiles();
                }
            } catch (e) {
                console.error('Failed to save file:', e);
            }
        },

        async loadFrom(filename: string) {
            if (!this.workDir) return;
            try {
                const req = await LoadRequest(this.workDir, filename);
                if (req) {
                    this.request = req;
                    this.currentFileName = filename;
                    // Sync the loaded object back to the Curl string
                    await this.syncToCurl();
                    // Clear previous response/errors when loading new file
                    this.response = new domain.HttpResponse();
                }
            } catch (e) {
                console.error('Failed to load file:', e);
            }
        }
    },
});