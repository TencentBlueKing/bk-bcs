<template>
  <bcs-select
    :popover-min-width="300"
    searchable
    :value="!multiple && Array.isArray(value) ? value.join(',') : value"
    :loading="isLoading"
    :multiple="multiple"
    class="flex-1 bg-[#fff] w-0"
    @change="handleChange">
    <bcs-option v-for="v in valuesSuggestList" :key="v" :id="v" :name="v" />
  </bcs-select>
</template>
<script setup lang="ts">
// import { isEqual } from 'lodash';
import { PropType, ref, watch } from 'vue';

import useViewConfig from './use-view-config';

const props = defineProps({
  value: {
    type: [Array, String] as PropType<string|string[]>,
    default: () => [],
  },
  label: {
    type: String,
    default: '',
  },
  clusterNamespaces: {
    type: Array as PropType<IClusterNamespace[]>,
    default: () => [],
  },
  multiple: {
    type: Boolean,
    default: true,
  },
});
const emits = defineEmits(['change', 'input']);

const { valuesSuggest } = useViewConfig();

const isLoading = ref(false);
const valuesSuggestList = ref<string[]>([]);
watch(
  [
    () => props.label,
    () => props.clusterNamespaces,
  ],
  async () => {
    // if (isEqual(newValue, oldValue)) return;
    isLoading.value = true;
    valuesSuggestList.value = await valuesSuggest({
      label: props.label,
      clusterNamespaces: props.clusterNamespaces,
    });
    // todo 判断值是否在列表中
    if (Array.isArray(props.value) && props.value?.some(v => !valuesSuggestList.value.includes(v))) {
      handleChange([]);
    }
    isLoading.value = false;
  },
  { immediate: true, deep: true },
);

const handleChange = (v: string|string[]) => {
  const data = typeof v === 'string' ? [v] : v;
  emits('change', data);
  emits('input', data);
};
</script>
