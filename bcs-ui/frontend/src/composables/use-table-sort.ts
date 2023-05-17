import { Ref, ref, computed } from 'vue';
import { sort } from '@/common/util';

export default function useTableSort(data: Ref<any[]>, extraDataFn?) {
  const sortData = ref({
    prop: '',
    order: '',
  });
  const handleSortChange = (data) => {
    sortData.value = {
      prop: data.prop,
      order: data.order,
    };
  };
  const sortTableData = computed<any[]>(() => {
    const { prop, order } = sortData.value;
    return prop ? sort(data.value, prop, order, extraDataFn) : data.value;
  });

  return {
    sortData,
    sortTableData,
    handleSortChange,
  };
}
