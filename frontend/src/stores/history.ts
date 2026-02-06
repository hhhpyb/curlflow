import { defineStore } from 'pinia';
import { GetHistoryList, ClearHistory } from '../../wailsjs/go/main/App';
import { history, domain } from '../../wailsjs/go/models';
import { useRequestStore } from './request';
import { EventsOn } from '../../wailsjs/runtime';

export const useHistoryStore = defineStore('history', {
    state: () => ({
        list: [] as history.HistoryEntry[],
    }),
    actions: {
        async fetchHistory() {
            const requestStore = useRequestStore();
            if (!requestStore.workDir) return;

            try {
                const res = await GetHistoryList(requestStore.workDir);
                this.list = res || [];
            } catch (e) {
                console.error('Failed to fetch history:', e);
            }
        },
        async clearHistory() {
            const requestStore = useRequestStore();
            if (!requestStore.workDir) return;

            try {
                await ClearHistory(requestStore.workDir);
                this.list = [];
            } catch (e) {
                console.error('Failed to clear history:', e);
            }
        },
        loadEntry(entry: history.HistoryEntry) {
            const requestStore = useRequestStore();
            
            // Treat as new request
            requestStore.request = JSON.parse(JSON.stringify(entry.request)); // Deep copy
            requestStore.meta = {
                id: crypto.randomUUID(),
                status: 'new',
                source: 'user',
                tags: []
            } as any;
            
            requestStore.currentFileName = ''; // Clear filename so "Save" triggers "Save As"
            requestStore.response = new domain.HttpResponse(); // Reset response
            requestStore.syncToCurl(); // Update Editor
            requestStore.smartFocus(); // Smart switch tab
        },
        setupListeners() {
            // Listen for backend event
            EventsOn('history_updated', () => {
                this.fetchHistory();
            });
        }
    }
});
