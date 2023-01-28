import { Ref, computed, ComputedRef, ref } from '@vue/composition-api';

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
export default function useTableSearch(data: Ref<any[]>, keys: Ref<any[]>): ITableSeachResult {
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
      return String(str).toLowerCase()
        .includes(searchValue.value.toLowerCase());
    }));
  });

  return {
    searchValue,
    tableDataMatchSearch,
  };
}
