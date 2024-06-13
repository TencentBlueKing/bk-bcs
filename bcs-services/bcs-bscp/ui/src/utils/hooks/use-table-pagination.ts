import { ref, onBeforeMount } from 'vue';

export default function useTablePagination(
  tableId: string,
  options: { [key: string]: boolean | string | number } = {},
) {
  const pagination = ref({
    count: 0,
    current: 1,
    limit: 10,
    ...options,
  });

  const updatePagination = (key: 'count' | 'current' | 'limit', value: number) => {
    pagination.value[key] = value;
    if (key === 'limit') {
      pagination.value.current = 1;
      const tablePagination = JSON.parse(localStorage.getItem('tablePagination') || '{}');
      tablePagination[tableId] = { limit: value };
      localStorage.setItem('tablePagination', JSON.stringify(tablePagination));
    }
  };

  onBeforeMount(() => {
    const tablePagination = JSON.parse(localStorage.getItem('tablePagination') || '{}');
    if (tablePagination[tableId]?.limit) {
      pagination.value.limit = tablePagination[tableId].limit;
    }
  });

  return {
    pagination,
    updatePagination,
  };
}
