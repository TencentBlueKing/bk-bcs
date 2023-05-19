import { ref, watch } from 'vue';

export function usePage(pageOrLimitChange?: Function) {
  const pagination = ref({
    current: 1,
    count: 0,
    limit: 20,
  });

  watch(() => [pagination.value.current, pagination.value.limit], () => {
    pageOrLimitChange?.();
  });

  function handlePageChange(page: number) {
    pagination.value.current = page;
  }
  function handlePageLimitChange(limit: number) {
    pagination.value.current = 1;
    pagination.value.limit = limit;
  }
  return {
    pagination,
    handlePageChange,
    handlePageLimitChange,
  };
}
