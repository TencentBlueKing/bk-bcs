import {  computed, Ref, unref, watch } from '@vue/composition-api';
import useApp from './use-app';

export interface IStorageConfig {
  autoInitValue?: boolean
  engine?: Storage
  scope?: 'project' | 'cluster' | undefined
}

/**
 * 浏览器缓存
 * @param ref
 * @param key
 * @param config
 */
export function useStorage(
  ref: Ref<any>,
  key: string,
  config: IStorageConfig | Ref<IStorageConfig> = {
    autoInitValue: true,
    engine: localStorage,
    scope: undefined,
  },
) {
  const { autoInitValue = true, engine = localStorage, scope } = unref(config);
  const { projectID, curClusterId } = useApp;

  const storageKey = computed(() => {
    let itemKey: string | undefined = key;
    if (scope === 'project') {
      itemKey = `${projectID.value}-${key}`;
    } else if (scope === 'cluster') {
      itemKey = curClusterId.value ? `${curClusterId.value}-${key}` : undefined;
    }

    return itemKey;
  });

  function setRefValue() {
    if (storageKey.value === undefined) return;
    try {
      ref.value = engine.getItem(storageKey.value);
    } catch (e) {
      console.error(e);
    }
  }

  watch(ref, () => {
    if (storageKey.value === undefined) return;
    try {
      engine.setItem(storageKey.value, ref.value);
    } catch (e) {
      console.error(e);
    }
  });

  if (autoInitValue && !ref.value) {
    setRefValue();
  }
}
