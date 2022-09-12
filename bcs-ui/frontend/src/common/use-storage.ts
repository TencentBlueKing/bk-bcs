import {  computed, Ref, unref, watch } from '@vue/composition-api';
import useApp from './use-app';

export interface IStorageConfig {
  autoInitValue: true
  engine: Storage
  scope: 'project' | 'cluster' | undefined
}

/**
 * 浏览器缓存
 * @param key
 * @param namespace
 * @param config
 */
export function useStorage(
  key: Ref<any>,
  namespace: string,
  config: IStorageConfig | Ref<IStorageConfig>,
) {
  const { autoInitValue, engine = localStorage, scope } = unref(config);
  const { projectID, curClusterId } = useApp;

  const storageKey = computed(() => {
    let itemKey = namespace;
    if (scope === 'project') {
      itemKey = `${projectID.value}-${namespace}`;
    } else if (scope === 'cluster') {
      itemKey = `${curClusterId.value}-${namespace}`;
    }

    return itemKey;
  });

  function setKeys() {
    try {
      key.value = engine.getItem(storageKey.value);
    } catch (e) {
      console.error(e);
    }
  }

  watch(key, () => {
    try {
      engine.setItem(storageKey.value, key.value);
    } catch (e) {
      console.error(e);
    }
  });

  if (autoInitValue) {
    setKeys();
  }
}
