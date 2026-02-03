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

        createNewRequest() {
            this.request = new domain.HttpRequest();
            this.request.method = 'GET';
            this.curlCode = '';
            this.currentFileName = '';
            // Sync empty request to curl to have a fresh start
            this.syncToCurl(); 
            this.response = new domain.HttpResponse();
        },

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

        async saveCurrent(filename?: string) {
            if (!this.workDir) {
                throw new Error('No working directory selected. Please open a folder first.');
            }

            let targetName = '';

            // Scenario A: Save As (User provided a filename)
            if (filename && filename.trim() !== '') {
                targetName = filename;
            } 
            // Scenario B: Overwrite (Use existing filename)
            else if (this.currentFileName) {
                targetName = this.currentFileName;
            } 
            // Scenario C: Error (No filename available)
            else {
                throw new Error('Filename is required for a new request.');
            }

            try {
                const savedPath = await SaveRequest(this.workDir, targetName, this.request);
                if (savedPath) {
                    // Normalize filename (ensure .json extension is handled if backend didn't return full name, though backend path usually does)
                    // We trust the logic that if we provided "foo", and it saved "foo.json", we want to track "foo.json"
                    const actualName = targetName.toLowerCase().endsWith('.json') ? targetName : targetName + '.json';
                    
                    this.currentFileName = actualName;
                    await this.fetchFiles();
                }
                return savedPath;
            } catch (e) {
                console.error('Failed to save file:', e);
                throw e; // Re-throw to let UI handle the error display
            }
        },

        async loadFrom(filename: string) {
            if (!this.workDir) {
                console.warn('Cannot load file: No working directory selected');
                return;
            }
            try {
                const req = await LoadRequest(this.workDir, filename);
                // Check if req is valid (basic check)
                if (req) {
                    this.request = req;
                    this.currentFileName = filename;
                    // Sync the loaded object back to the Curl string
                    await this.syncToCurl();
                    // Clear previous response/errors when loading new file
                    this.response = new domain.HttpResponse();
                }
            } catch (e) {
                console.error(`Failed to load file ${filename}:`, e);
            }
        }
    },
});
