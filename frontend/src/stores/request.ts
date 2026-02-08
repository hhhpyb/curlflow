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
    DeleteFile,
    GetProjectConfig,
    SaveProjectConfig,
    GetEnvConfig,
    PurgeDeletedFiles,
    GetLastOpenedProject,
    OpenProject
} from '../../wailsjs/go/main/App';
import {domain, main, storage} from '../../wailsjs/go/models';
import { useEnvStore } from './env';

export interface FileSummary {
    fileName: string;
    meta: domain.MetaData;
    method?: string;
}

export interface InterfaceNode {
    mainFile: storage.FileSummary;
    children: storage.FileSummary[];
}

// ==========================================
// Helper: Clean up Windows CMD escape characters
// ==========================================
function preprocessCurl(curl: string): string {
    if (!curl) return "";
    let s = curl.trim();
    // 1. Handle Windows CMD line continuation (caret + newline)
    s = s.replace(/\^\s*[\r\n]+/g, " ");
    
    // 2. Handle standard newlines
    s = s.replace(/[\r\n]+/g, " ");

    // 3. Handle Windows Escapes safely using placeholders
    // CMD uses ^" for a standard quote, and ^\^" for an escaped quote inside JSON
    const SAFE_ESCAPED_QUOTE = "___ESCAPED_QUOTE___";

    s = s.split("^\\^\"").join(SAFE_ESCAPED_QUOTE); // Handle ^\^" first
    s = s.split('^"').join('"');                    // Handle ^"
    s = s.split("^{").join("{");                    // Handle ^{
    s = s.split("^}").join("}");                    // Handle ^}
    
    // Restore the placeholder to a standard escaped quote \"
    s = s.split(SAFE_ESCAPED_QUOTE).join('\\"');

    // 4. Remove any remaining carets that aren't part of valid escapes
    s = s.replace(/\^/g, "");

    return s.trim();
}

// ==========================================
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
        swaggerUrl: '',
        
        // Environment Config
        envConfig: {
            activeEnvName: 'dev',
            environments: {} as Record<string, { variables: Record<string, string> }>
        },

        // Path Variables: Extracted from URL {key}
        pathParams: {} as Record<string, string>,

        // Search and View settings
        searchKeyword: '',
        showDeleted: false,
        activeEditorTab: 'Params',
    }),
    getters: {
        /**
         * Computes a grouped tree of interfaces and their test cases.
         * Returns: Record<FolderName, InterfaceNode[]>
         */
        fileTree(state): Record<string, InterfaceNode[]> {
            // ... (existing implementation)
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
                // Sort by filename length to find the "Main" file
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

            // Sort interfaces within each tag
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
         * Returns a flat list of files following the visual order in the sidebar.
         */
        flatFileList(): storage.FileSummary[] {
            const tree = this.fileTree;
            const flat: storage.FileSummary[] = [];
            
            // Sort tags alphabetically, with 'Uncategorized' usually at the end or following sort
            const sortedTags = Object.keys(tree).sort();

            sortedTags.forEach(tag => {
                tree[tag].forEach(node => {
                    flat.push(node.mainFile);
                    node.children.forEach(child => {
                        flat.push(child);
                    });
                });
            });

            return flat;
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
        },

        /**
         * Returns a list of existing folder names for AutoComplete.
         */
        folderOptions(): string[] {
            // fileTree is Record<string, InterfaceNode[]> where key is the folder name
            return Object.keys(this.fileTree).filter(tag => tag !== 'Uncategorized');
        }
    },
    actions: {
        async syncFromCurl() {
            try {
                this.curlCode = preprocessCurl(this.curlCode);
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
                
                // --- Path Variable Substitution ---
                // Replace {key} in URL with values from pathParams
                if (finalRequest.url && Object.keys(this.pathParams).length > 0) {
                    let url = finalRequest.url;
                    for (const [key, val] of Object.entries(this.pathParams)) {
                        // Use negative lookahead/lookbehind to ensure we only match single braces {key}
                        // and not double braces {{key}}
                        const regex = new RegExp(`(?<!\\{)\\{${key}\\}(?!\\})`, 'g');
                        url = url.replace(regex, val);
                    }
                    finalRequest.url = url;
                }
                
                console.log('--- Debug: Final Request Object (After Path Params) ---');
                console.log(finalRequest);
                
                const res = await SendRequest(finalRequest, this.workDir);
                this.response = res;
            } catch (e) {
                console.error('Request failed:', e);
                this.response.error = String(e);
            } finally {
                this.isLoading = false;
            }
        },

        // ================= File Actions =================

        async updateRequestMeta(metaInfo: { summary: string, tag: string, description: string }) {
            if (!this.meta) return;

            // 1. Update local state
            this.meta.summary = metaInfo.summary;
            this.meta.description = metaInfo.description;
            this.meta.tags = metaInfo.tag ? [metaInfo.tag] : [];

            // 2. Save to disk
            try {
                await this.saveCurrent();
                // 3. Refresh file list to reflect folder changes in sidebar
                await this.fetchFiles();
            } catch (e) {
                console.error('Failed to update request meta:', e);
                throw e;
            }
        },

        createNewRequest() {
            this.request = new domain.HttpRequest();
            this.request.method = 'GET';
            this.request.url = '';
            this.request.headers = {};
            this.request.body = '';
            
            // Initialize meta with source = 'user' for new requests
            this.meta = {
                id: crypto.randomUUID(),
                status: 'active',
                source: 'user',
                tags: []
            } as any;
            
            this.curlCode = '';
            this.currentFileName = '';
            this.pathParams = {};
            
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

        async openProject(dir: string) {
            try {
                // Explicitly switch to a known path (e.g. from history list)
                if (dir) {
                    // Notify backend to update "Recent" list (move to top)
                    await OpenProject(dir);

                    this.workDir = dir;
                    localStorage.setItem('curlflow_workDir', dir);
                    this.currentFileName = ''; 
                    await this.fetchFiles();
                    
                    const envStore = useEnvStore();
                    await envStore.loadEnvs();
                }
            } catch (e) {
                console.error('Failed to open project:', e);
            }
        },

        async init() {
            try {
                // 1. Try to get the last opened project from the backend (System of Truth)
                const lastProject = await GetLastOpenedProject();
                if (lastProject) {
                    this.workDir = lastProject;
                    await this.fetchFiles();
                    const envStore = useEnvStore();
                    await envStore.loadEnvs();
                    return true;
                }
            } catch (e) {
                console.error('Failed to get last opened project from backend:', e);
            }

            // 2. Fallback to localStorage if backend fails or returns empty (Legacy)
            const savedDir = localStorage.getItem('curlflow_workDir');
            if (savedDir) {
                this.workDir = savedDir;
                // Sync back to backend
                try {
                    await OpenProject(savedDir); 
                } catch (e) {}
                
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
                // Also load environment config when fetching files
                await this.loadEnvConfig();
                
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
                // IMPORTANT: Use SaveFullRequest instead of SaveRequest to preserve metadata changes
                // We construct the full RequestFile object here
                const reqFile = {
                    _meta: this.meta || {
                        id: crypto.randomUUID(), // Generate a new ID if meta is missing
                        status: 'active',
                        source: 'user', // Default source for manually saved requests
                    },
                    data: this.request
                };

                const savedPath = await SaveFullRequest(this.workDir, targetName, reqFile as any);
                if (savedPath) {
                    const actualName = targetName.toLowerCase().endsWith('.json') ? targetName : targetName + '.json';
                    this.currentFileName = actualName;
                    
                    // Update local meta if it was newly created
                    if (!this.meta) {
                        this.meta = reqFile._meta as any;
                    }
                    
                    await this.fetchFiles();
                }
                return savedPath;
            } catch (e) {
                console.error('Failed to save file:', e);
                throw e;
            }
        },

        async loadFrom(filename: string) {
            if (!this.workDir) {
                console.warn('Cannot load file: No working directory selected');
                return;
            }

            // Defense: Prevent loading environment config as a request
            if (filename.toLowerCase() === '.curlflow/environments.json' || filename.toLowerCase() === 'environments.json') {
                console.warn('Security: Attempted to load environments configuration as a request file. Operation blocked.');
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
                    
                    // Ensure headers is at least an empty object for UI components
                    if (!this.request.headers) {
                        this.request.headers = {};
                    }
                    
                    this.meta = res._meta || null;

                    this.currentFileName = filename;
                    // Sync the loaded object back to the Curl string
                    await this.syncToCurl();
                    // Clear previous response/errors when loading new file
                    this.response = new domain.HttpResponse();

                    // Smart focus on the most relevant tab
                    this.smartFocus();
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
                source: 'user', // Manual cases are always treated as user data
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

        async loadProjectConfig() {
            if (!this.workDir) return;
            try {
                const config = await GetProjectConfig(this.workDir);
                this.swaggerUrl = config.swagger_url || '';
            } catch (e) {
                console.error('Failed to load project config:', e);
            }
        },

        async saveProjectConfig(url: string) {
            if (!this.workDir) return;
            try {
                const result = await SaveProjectConfig(this.workDir, url);
                if (result === 'success') {
                    this.swaggerUrl = url;
                }
            } catch (e) {
                console.error('Failed to save project config:', e);
            }
        },


        async purgeDeleted() {
            if (!this.workDir) return;
            try {
                const result = await PurgeDeletedFiles(this.workDir);
                
                // @ts-ignore
                if (window.$message) window.$message.success(result);

                // If the currently opened request is one of the purged files, clear the editor
                if (this.meta && this.meta.status === 'deleted') {
                    this.createNewRequest();
                }

                // Refresh the sidebar to reflect the changes
                await this.fetchFiles();
            } catch (e) {
                console.error('Failed to purge deleted files:', e);
                // @ts-ignore
                if (window.$message) window.$message.error(`Purge Failed: ${e}`);
            }
        },

        async loadEnvConfig() {
            if (!this.workDir) return;
            try {
                const config = await GetEnvConfig(this.workDir);
                this.envConfig = config || { activeEnvName: '', environments: {} };
            } catch (e) {
                console.error('Failed to load env config:', e);
            }
        },

        async duplicateFile(fileName: string) {
            if (!this.workDir) return;
            try {
                const res = await LoadRequest(this.workDir, fileName) as any;
                if (res) {
                    const baseName = fileName.replace(/\.json$/i, '');
                    const newFileName = `${baseName}_copy.json`;

                    const newReqFile = {
                        _meta: {
                            ...res._meta,
                            id: crypto.randomUUID(),
                            summary: (res._meta.summary || baseName) + ' - Copy'
                        },
                        data: res.data
                    };

                    await SaveFullRequest(this.workDir, newFileName, newReqFile as any);
                    await this.fetchFiles();

                    // @ts-ignore
                    if (window.$message) window.$message.success('File duplicated');
                }
            } catch (e) {
                console.error('Failed to duplicate file:', e);
                // @ts-ignore
                if (window.$message) window.$message.error(`Duplicate Failed: ${e}`);
            }
        },

        async copyCurlByFilename(fileName: string) {
            if (!this.workDir) return;
            try {
                const res = await LoadRequest(this.workDir, fileName) as any;
                if (res && res.data) {
                    const curl = await BuildCurl(res.data);
                    await navigator.clipboard.writeText(curl);
                    // @ts-ignore
                    if (window.$message) window.$message.success('Copied to clipboard');
                }
            } catch (e) {
                console.error('Failed to copy curl:', e);
                // @ts-ignore
                if (window.$message) window.$message.error('Failed to copy to clipboard');
            }
        },

        async deleteFileByFilename(fileName?: string) {
            const target = fileName || this.currentFileName;
            if (!target || !this.workDir) return;

            try {
                await DeleteFile(this.workDir, target);

                // If the deleted file is the current one, clear the editor
                if (target === this.currentFileName) {
                    this.createNewRequest();
                }

                await this.fetchFiles();
                // @ts-ignore
                if (window.$message) window.$message.success('File marked as deleted');
            } catch (e) {
                console.error('Failed to delete file:', e);
                // @ts-ignore
                if (window.$message) window.$message.error(`Delete Failed: ${e}`);
            }
        },

        /**
         * Automatically switches to the most relevant tab based on the current request content.
         */
        smartFocus() {
            const bodyMethods = ['POST', 'PUT', 'PATCH'];
            const url = this.request.url || '';
            const body = this.request.body || '';

            // Priority 1: Body (for relevant methods and non-empty body)
            if (bodyMethods.includes((this.request.method || 'GET').toUpperCase()) && body.trim().length > 0) {
                this.activeEditorTab = 'Body';
                return;
            }

            // Priority 2: Path Variables (if URL contains {variable})
            if (url.includes('{') && url.includes('}')) {
                this.activeEditorTab = 'Path';
                return;
            }

            // Priority 3: Query Params (if URL contains ?) or Default
            // Both lead to 'Params' tab
            this.activeEditorTab = 'Params';
        },

        selectNextFile() {
            const list = this.flatFileList;
            if (list.length === 0) return;

            const currentIndex = list.findIndex(f => f.fileName === this.currentFileName);
            let nextIndex = 0;
            if (currentIndex !== -1) {
                nextIndex = (currentIndex + 1) % list.length;
            }
            
            this.loadFrom(list[nextIndex].fileName);
        },

        selectPrevFile() {
            const list = this.flatFileList;
            if (list.length === 0) return;

            const currentIndex = list.findIndex(f => f.fileName === this.currentFileName);
            let prevIndex = list.length - 1;
            if (currentIndex !== -1) {
                prevIndex = (currentIndex - 1 + list.length) % list.length;
            }

            this.loadFrom(list[prevIndex].fileName);
        }
    },
});
