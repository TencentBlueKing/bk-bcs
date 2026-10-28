import { computed, ComputedRef, isRef, Ref, ref, watch } from 'vue';

import { IField } from './use-table-setting';

import { USER_TABLE_SETTING } from '@/common/constant';
import $store from '@/store';

type MaybeRef<T> = T | Ref<T> | ComputedRef<T>;

// 显式收窄类型，避免 unref 无法解包 ComputedRef 导致类型推断错误
const toValue = <T>(source: MaybeRef<T>): T => (isRef(source) ? source.value : source) as T;

/**
 * 表格列显隐设置（按用户维度持久化到 localStorage）
 * @param fields 列配置（支持响应式，动态列变化时自动更新）
 * @param storageKey 业务唯一标识，例如资源 kind
 * 说明：未设置过的字段默认显示（defaultChecked 为 false 时默认隐藏），disabled 字段始终显示
 */
export default function useUserTableSetting(fields: MaybeRef<IField[]>, storageKey: MaybeRef<string>) {
  const username = computed(() => $store.state.user?.username || '');
  const key = computed(() => `${USER_TABLE_SETTING}${toValue(storageKey)}_${username.value}`);

  const readStorage = (): Record<string, boolean> => {
    try {
      return JSON.parse(localStorage.getItem(key.value) || '{}') || {};
    } catch (_) {
      return {};
    }
  };

  const storageFields = ref<Record<string, boolean>>(readStorage());
  // 用户信息异步加载或业务标识变化时重新读取
  watch(key, () => {
    storageFields.value = readStorage();
  });

  const size = ref('medium');
  const fieldList = computed<IField[]>(() => toValue(fields) || []);
  const selectedFields = computed(() => fieldList.value.filter((item) => {
    if (item.disabled) return true;
    if (item.id in storageFields.value) return storageFields.value[item.id];
    return item.defaultChecked !== false;
  }));

  const tableSetting = computed(() => ({
    size: size.value,
    fields: fieldList.value,
    selectedFields: selectedFields.value,
  }));

  const isColumnRender = (id: string) => selectedFields.value.some(item => item.id === id);

  const handleSettingChange = ({ size: newSize, fields: selected }: { size: string, fields: IField[] }) => {
    size.value = newSize;
    // 合并旧数据，保留当前未出现的动态列设置
    const data = fieldList.value.reduce<Record<string, boolean>>((pre, cur) => {
      pre[cur.id] = selected.some(item => item.id === cur.id);
      return pre;
    }, { ...storageFields.value });
    storageFields.value = data;
    localStorage.setItem(key.value, JSON.stringify(data));
  };

  return {
    tableSetting,
    isColumnRender,
    handleSettingChange,
  };
}
