/// <reference types="vite/client" />

interface Window {
    $message: import('naive-ui').MessageApi;
}

declare module '*.vue' {
    import type {DefineComponent} from 'vue'
    const component: DefineComponent<{}, {}, any>
    export default component
}
