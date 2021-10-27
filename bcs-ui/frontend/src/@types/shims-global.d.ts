import Vue, { VNode } from 'vue'
import { ComponentRenderProxy } from '@vue/composition-api'

declare global {
    namespace JSX {
        interface Element extends VNode {}
        interface ElementClass extends ComponentRenderProxy {}
        interface IntrinsicElements {
            [elem: string]: any;
        }
    }

    interface Window {
        [key: string]: any;
    }

    export const SITE_URL: string
}
