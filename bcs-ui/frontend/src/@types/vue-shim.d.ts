// eslint-disable-next-line @typescript-eslint/no-unused-vars
import Vue from 'vue';

declare module 'vue/types/vue' {
  interface Vue {
    PROJECT_CONFIG: Record<string, string>;
    $INTERNAL: boolean;
    $bkInfo: any;
    $bkMessage: any;
    $bkNotify: any;
    $chainable: (obj: any, path: string) => any;
  }
}
