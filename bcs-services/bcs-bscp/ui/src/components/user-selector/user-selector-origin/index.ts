import type { App, Plugin } from 'vue';

import request from './request';
import UserSelector from './selector.vue';

// UserSelector.install = (Vue) => {
//   window.$vueApp.component(UserSelector.name, UserSelector);
// };

// export default UserSelector;

// export { request };

export interface OriginComponent {
  name: string;
  install?: Plugin;
}

const withInstall = <T extends OriginComponent>(component: T): T & Plugin  => {
  // eslint-disable-next-line no-param-reassign
  component.install = function (app: App, { prefix } = {}) {
    const pre = app.config.globalProperties.bkUIPrefix || prefix || 'Bk';
    app.component(pre + component.name, component);
  };
  return component as T & Plugin;
};

export { request };

const BkUserSelector = withInstall(UserSelector);
export default BkUserSelector;
