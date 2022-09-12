import { VNode } from 'vue';
import { ComponentRenderProxy } from '@vue/composition-api';

declare global {
  namespace JSX {
    type Element = VNode;
    type ElementClass = ComponentRenderProxy;
    interface IntrinsicElements {
      [elem: string]: any;
    }
  }

  interface Window {
    [key: string]: any;
  }

  export const SITE_URL: string;
}
