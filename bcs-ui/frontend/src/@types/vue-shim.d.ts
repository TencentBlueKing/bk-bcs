// eslint-disable-next-line @typescript-eslint/no-unused-vars
import Vue from 'vue';

import VueI18n from './vue-i18n';

declare module 'vue/types/vue' {
  interface Vue {
    PROJECT_CONFIG: Record<string, string>;
    $INTERNAL: boolean;
    $bkInfo: any;
    $bkMessage: any;
    $bkNotify: any;
    $chainable: (obj: any, path: string) => any;
    $t: typeof VueI18n.prototype.t;
  }
}

declare module 'vue-router' {
  interface RouteMeta {
    menuId?: string // 父菜单ID
    id?: string // 当前菜单ID
    title?: string // 标题
  }
}
