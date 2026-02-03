import { defineStore } from 'pinia';
import { ParseCurl, BuildCurl, SendRequest } from '../../wailsjs/go/main/App';
import { main } from '../../wailsjs/go/models';

export const useRequestStore = defineStore('request', {
    state: () => ({
        curlCode: '',
        request: new main.HttpRequest(),
        isLoading: false,
        response: new main.HttpResponse(),
    }),
    actions: {
        async syncFromCurl() {
            try {
                // ParseCurl returns a Promise<main.HttpRequest>
                const req = await ParseCurl(this.curlCode);
                this.request = req;
            } catch (e) {
                console.error('Failed to parse curl:', e);
            }
        },
        async syncToCurl() {
            try {
                // BuildCurl returns a Promise<string>
                const curl = await BuildCurl(this.request);
                this.curlCode = curl;
            } catch (e) {
                console.error('Failed to build curl:', e);
            }
        },
        async send() {
            this.isLoading = true;
            this.response = new main.HttpResponse(); // Reset response
            try {
                // SendRequest returns a Promise<main.HttpResponse>
                const res = await SendRequest(this.request);
                this.response = res;
            } catch (e) {
                console.error('Request failed:', e);
                // Handle error properly, maybe update response.error
                this.response.error = String(e);
            } finally {
                this.isLoading = false;
            }
        },
    },
});
