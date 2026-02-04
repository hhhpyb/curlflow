import { defineStore } from 'pinia';
import { 
    ParseCurl, 
    BuildCurl, 
    SendRequest,
    SelectWorkDir,
    GetFileList,
    GetFileSummaries,
    SaveRequest,
    SaveFullRequest,
    LoadRequest,
    SyncSwagger,
    DeleteFile
} from '../../wailsjs/go/main/App';
import { domain, storage } from '../../wailsjs/go/models';
import { useEnvStore } from './env';

export interface FileSummary {
    fileName: string;
    meta: domain.MetaData;
}

export interface InterfaceNode {
    mainFile: storage.FileSummary;
    children: storage.FileSummary[];
}

export const useRequestStore = defineStore('request', {
    state: () => ({
        curlCode: '',
        request: new domain.HttpRequest(),
        meta: null as domain.MetaData | null,
        isLoading: false,
        response: new domain.HttpResponse(),
        
        // File Management State
        workDir: '',
        fileList: [] as storage.FileSummary[],
        currentFileName: '',

        // Search and View settings
        searchKeyword: '',
        showDeleted: false,
    }),
    getters: {
        /**
         * Computes a grouped tree of interfaces and their test cases.
         * Returns: Record<FolderName, InterfaceNode[]>
         */
        fileTree(state): Record<string, InterfaceNode[]> {
            // 1. Initial Filtering
            const filtered = state.fileList.filter(item => {
                if (!state.showDeleted && item.meta.status === 'deleted') return false;

                if (state.searchKeyword) {
                    const keyword = state.searchKeyword.toLowerCase();
                    return (
                        (item.meta.summary || '').toLowerCase().includes(keyword) ||
                        (item.meta.key || '').toLowerCase().includes(keyword) ||
                        item.fileName.toLowerCase().includes(keyword)
                    );
                }
                return true;
            });

            // 2. Group by meta.id to aggregate cases
            const groups: Record<string, storage.FileSummary[]> = {};
            filtered.forEach(item => {
                const id = item.meta.id || 'no-id';
                if (!groups[id]) groups[id] = [];
                groups[id].push(item);
            });

            // 3. Build InterfaceNodes and group by Folder (Tag)
            const tree: Record<string, InterfaceNode[]> = {};

            Object.values(groups).forEach(group => {
                // Sort by filename length to find the "Main" file (usually shortest, e.g., "get_user.json" vs "get_user_case_1.json")
                group.sort((a, b) => a.fileName.length - b.fileName.length);
                
                const mainFile = group[0];
                const children = group.slice(1);
                
                const tag = (mainFile.meta.tags && mainFile.meta.tags.length > 0)
                    ? mainFile.meta.tags[0]
                    : 'Uncategorized';

                if (!tree[tag]) tree[tag] = [];
                
                tree[tag].push({
                    mainFile,
                    children
                });
            });

            // Sort interfaces within each tag by summary or filename
            Object.keys(tree).forEach(tag => {
                tree[tag].sort((a, b) => {
                    const labelA = a.mainFile.meta.summary || a.mainFile.fileName;
                    const labelB = b.mainFile.meta.summary || b.mainFile.fileName;
                    return labelA.localeCompare(labelB);
                });
            });

            return tree;
        },

        /**
         * Returns all files (test cases) belonging to the same API (same meta.id).
         */
        relatedCases(state): storage.FileSummary[] {
            if (!state.meta || !state.meta.id) return [];

            return state.fileList
                .filter(f => f.meta.id === state.meta?.id)
                .sort((a, b) => {
                    // Main case (shortest name) first, then alphabetical
                    if (a.fileName.length !== b.fileName.length) {
                        return a.fileName.length - b.fileName.length;
                    }
                    return a.fileName.localeCompare(b.fileName);
                });
        }
    },
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
            const envStore = useEnvStore();
            try {
                // Process variables in curl code before sending
                const processedCurl = envStore.processString(this.curlCode);
                console.log('--- Debug: Processed Curl ---');
                console.log(processedCurl);

                // Re-parse to ensure domain.HttpRequest is up to date with replaced values
                const finalRequest = await ParseCurl(processedCurl);
                console.log('--- Debug: Final Request Object ---');
                console.log(finalRequest);
                
                const res = await SendRequest(finalRequest);
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
            this.meta = null;
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
                    localStorage.setItem('curlflow_workDir', dir);
                    this.currentFileName = ''; // Reset current file when changing dir
                    await this.fetchFiles();
                    
                    // Reload environments for the new directory
                    const envStore = useEnvStore();
                    await envStore.loadEnvs();
                    console.log('WorkDir selected, environments loaded from:', dir);
                }
            } catch (e) {
                console.error('Failed to select directory:', e);
            }
        },

        async init() {
            const savedDir = localStorage.getItem('curlflow_workDir');
            if (savedDir) {
                this.workDir = savedDir;
                await this.fetchFiles();
                const envStore = useEnvStore();
                await envStore.loadEnvs();
                return true;
            }
            return false;
        },

        async fetchFiles() {
            if (!this.workDir) return;
            try {
                // Use the new GetFileSummaries method for efficient sidebar loading
                const summaries = await GetFileSummaries(this.workDir);
                this.fileList = summaries || [];
            } catch (e) {
                console.error('Failed to list file summaries:', e);
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

            // Defense: Prevent loading environment config as a request
            if (filename.toLowerCase() === 'environments.json') {
                console.warn('Security: Attempted to load environments.json as a request file. Operation blocked.');
                return;
            }

            try {
                // Cast result to any first to avoid type mismatch with old Wails definition if it hasn't updated
                const res = await LoadRequest(this.workDir, filename) as any;
                
                // Check if res is valid
                if (res) {
                    // New structure: res has _meta and data
                    // We assign data to this.request to keep UI working
                    this.request = res.data || new domain.HttpRequest();
                    this.meta = res._meta || null;

                    this.currentFileName = filename;
                    // Sync the loaded object back to the Curl string
                    await this.syncToCurl();
                    // Clear previous response/errors when loading new file
                    this.response = new domain.HttpResponse();
                }
            } catch (e) {
                console.error(`Failed to load file ${filename}:`, e);
            }
        },

        async importSwagger(url: string) {
            if (!this.workDir) {
                const errMsg = 'No working directory selected. Please open a folder first.';
                // @ts-ignore
                if (window.$message) window.$message.error(errMsg);
                return;
            }

            try {
                this.isLoading = true;
                const result = await SyncSwagger(this.workDir, url);
                
                // Show feedback via Naive UI message (assuming window.$message is available)
                // @ts-ignore
                if (window.$message) {
                    // @ts-ignore
                    window.$message.success(result, { duration: 5000 });
                }

                // Refresh file list to show new/updated/deleted status
                await this.fetchFiles();
            } catch (e) {
                console.error('Failed to sync swagger:', e);
                // @ts-ignore
                if (window.$message) window.$message.error(`Sync Failed: ${e}`);
            } finally {
                this.isLoading = false;
            }
        },

        async createCase(caseName: string) {
            if (!this.workDir || !this.currentFileName || !this.meta) {
                throw new Error('Cannot create case: No active request file.');
            }

            const cleanCaseName = caseName.trim().replace(/[^a-zA-Z0-9_-]/g, '_');
            if (!cleanCaseName) throw new Error('Invalid case name.');

            // 1. Determine main filename (strip extension and any existing case suffix)
            let baseName = this.currentFileName.replace(/\.json$/i, '');
            baseName = baseName.split('_case_')[0];

            const newFileName = `${baseName}_case_${cleanCaseName}.json`;

            // 2. Construct new metadata (Deep copy tags and keep same ID)
            const newMeta = { 
                ...this.meta,
                tags: [...(this.meta.tags || [])], // Ensure tags are copied
                status: 'active',
                last_synced_at: Math.floor(Date.now() / 1000),
                summary: `${this.meta.summary || baseName} (${caseName})`
            } as any;

            // 3. Construct complete RequestFile object
            const reqFile = {
                _meta: newMeta,
                data: this.request
            } as any;

            try {
                // 4. Use SaveFullRequest to ensure our custom meta is preserved
                const savedPath = await SaveFullRequest(this.workDir, newFileName, reqFile);
                if (savedPath) {
                    await this.fetchFiles();
                    await this.loadFrom(newFileName);
                }
            } catch (e) {
                console.error('Failed to create case:', e);
                throw e;
            }
        },

        async deleteCurrentFile() {
            if (!this.workDir || !this.currentFileName) return;

            try {
                await DeleteFile(this.workDir, this.currentFileName);
                
                const deletedName = this.currentFileName;
                const alternatives = this.relatedCases.filter(f => f.fileName !== deletedName);
                
                await this.fetchFiles();

                if (alternatives.length > 0) {
                    await this.loadFrom(alternatives[0].fileName);
                } else {
                    this.createNewRequest();
                }
            } catch (e) {
                console.error('Failed to delete file:', e);
                throw e;
            }
        }
    },
});
