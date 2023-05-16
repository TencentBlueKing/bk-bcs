import { Ref, computed, ComputedRef, ref } from 'vue';
import { padIPv6, validateIPv6 } from '@/common/util';

export interface ITableSeachResult {
  searchValue: Ref<string>;
  tableDataMatchSearch: ComputedRef<any[]>;
}

/**
 * 搜索
 * @param data
 * @param keys
 * @returns
 */
export default function useTableSearch(data: Ref<any[]>, keys: Ref<string[]>): ITableSeachResult {
  const searchValue = ref('');
  const tableDataMatchSearch = computed(() => {
    if (!searchValue.value) return data.value;

    return data.value.filter(item => keys.value.some((key) => {
      const tmpKey = String(key).split('.');
      const str = tmpKey.reduce((pre, key) => {
        if (typeof pre === 'object') {
          return pre[key];
        }
        return pre;
      }, item);
      if (validateIPv6(str)) {
        return padIPv6(str).includes(padIPv6(searchValue.value));
      }
      return String(str).toLowerCase()
        .includes(searchValue.value.toLowerCase());
    }));
  });

  return {
    searchValue,
    tableDataMatchSearch,
  };
}
