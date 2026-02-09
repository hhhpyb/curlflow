import { defineStore } from 'pinia';
import { markRaw } from 'vue';
import { EventsOn } from '../../wailsjs/runtime';
import { WsConnect, WsDisconnect, WsSend } from '../../wailsjs/go/main/App';

export interface LogEntry {
    timestamp: number;
    direction: 'sent' | 'received' | 'system' | 'error';
    content: string;
}

export interface WebSocketSession {
    isConnected: boolean;
    url: string;
    messages: LogEntry[];
}

export const useWebSocketStore = defineStore('websocket', {
    state: () => ({
        sessions: {} as Record<string, WebSocketSession>,
        isListening: false,
    }),
    getters: {
        getSession: (state) => (sessionID: string) => {
            return state.sessions[sessionID];
        }
    },
    actions: {
        initSession(sessionID: string) {
            if (!this.sessions[sessionID]) {
                this.sessions[sessionID] = {
                    isConnected: false,
                    url: '',
                    messages: []
                };
            }
        },

        async connect(sessionID: string, url: string, headers: Record<string, string> = {}) {
            this.initSession(sessionID);
            this.sessions[sessionID].url = url;
            
            try {
                const res = await WsConnect(sessionID, url, headers);
                if (res !== 'success') {
                    this.addLog(sessionID, 'error', `Connection failed: ${res}`);
                }
            } catch (e: any) {
                this.addLog(sessionID, 'error', `Error calling backend: ${e}`);
            }
        },

        async disconnect(sessionID: string) {
            if (!this.sessions[sessionID]) return;
            try {
                await WsDisconnect(sessionID);
            } catch (e: any) {
                console.error(e);
            }
        },

        async sendMessage(sessionID: string, message: string) {
            if (!this.sessions[sessionID]?.isConnected) {
                this.addLog(sessionID, 'error', 'Cannot send: Not connected');
                return;
            }
            try {
                this.addLog(sessionID, 'sent', message);
                const res = await WsSend(sessionID, message);
                if (res !== 'success') {
                    this.addLog(sessionID, 'error', `Send failed: ${res}`);
                }
            } catch (e: any) {
                this.addLog(sessionID, 'error', `Error sending: ${e}`);
            }
        },

        addLog(sessionID: string, direction: LogEntry['direction'], content: string) {
            if (!this.sessions[sessionID]) this.initSession(sessionID);
            
            const session = this.sessions[sessionID];
            session.messages.push(markRaw({
                timestamp: Date.now(),
                direction,
                content
            }));

            if (session.messages.length > 500) {
                session.messages.shift();
            }
        },

        handleBatchMessage(payload: { session_id: string; messages: any[] }) {
            const { session_id, messages } = payload;
            if (!session_id || !messages || messages.length === 0) return;

            this.initSession(session_id);
            const session = this.sessions[session_id];

            // Bulk create log entries
            const newEntries = messages.map(msg => markRaw({
                timestamp: Date.now(),
                direction: 'received' as const,
                content: msg.data
            }));

            // Bulk push
            session.messages.push(...newEntries);

            // Hard limit truncation
            if (session.messages.length > 500) {
                // Keep the last 500
                session.messages = session.messages.slice(-500);
            }
        },

        handleEvent(payload: any) {
            // payload: { session_id, type, data }
            const { session_id, type, data } = payload;
            if (!session_id) return;

            this.initSession(session_id);
            const session = this.sessions[session_id];

            switch (type) {
                case 'connected':
                    session.isConnected = true;
                    this.addLog(session_id, 'system', data);
                    break;
                case 'disconnected':
                    session.isConnected = false;
                    this.addLog(session_id, 'system', data);
                    break;
                case 'message':
                    this.addLog(session_id, 'received', data);
                    break;
                case 'error':
                    this.addLog(session_id, 'error', data);
                    break;
            }
        },

        setupGlobalListener() {
            if (this.isListening) return;
            this.isListening = true;
            
            // @ts-ignore
            EventsOn('ws:event', (payload: any) => {
                this.handleEvent(payload);
            });

            // @ts-ignore
            EventsOn('ws:batch-message', (payload: any) => {
                this.handleBatchMessage(payload);
            });
        }
    }
});
