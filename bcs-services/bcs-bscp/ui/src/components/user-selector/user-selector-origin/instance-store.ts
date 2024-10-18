import { type ComponentPublicInstance } from 'vue';

const store = {};

export default {
  setInstance(group: string, id: string, proxy: ComponentPublicInstance) {
    if (!store[group]) {
      store[group] = {};
    }

    store[group][id] = proxy;
  },
  getInstance(group: string, id: string) {
    return store[group][id];
  },
};
