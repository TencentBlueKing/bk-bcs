<template>
  <div>
    <div
      class="flex items-center mb-[8px]"
      v-for="item, index in data"
      :key="index">
      <bcs-input
        :disabled="disabled"
        class="flex-1"
        v-model="item.key">
      </bcs-input>
      <span class="text-[#FF9C01] px-[8px]">=</span>
      <bcs-input
        :disabled="disabled"
        class="flex-1"
        v-model="item.value">
      </bcs-input>
      <i
        :class="[
          'text-[16px] text-[#C4C6CC] bk-icon icon-plus-circle-shape ml-[16px] cursor-pointer',
          { '!cursor-not-allowed !text-[#EAEBF0]': disabled },
        ]"
        @click="handleAddLabel">
      </i>
      <i
        :class="[
          'text-[16px] text-[#C4C6CC] bk-icon icon-minus-circle-shape ml-[10px] cursor-pointer',
          { '!cursor-not-allowed !text-[#EAEBF0]': data.length === 1 || disabled },
        ]"
        @click="handleDeleteLabel(index)">
      </i>
    </div>
  </div>
</template>
<script setup lang="ts">
import { PropType, watch, ref, onBeforeMount } from 'vue';
import { cloneDeep } from 'lodash';

const props = defineProps({
  value: {
    type: Array as PropType<Array<{ key: string, value: string }>>,
    default: () => [],
  },
  disabled: {
    type: Boolean,
    default: false,
  },
});

const emits = defineEmits(['change']);

const data = ref<Array<{ key: string, value: string }>>([]);

const watchOnce = watch(() => props.value, () => {
  handleSetData();
  watchOnce();
});

watch(data, () => {
  emits('change', data.value);
});

const handleSetData = () => {
  if (props.value.length) {
    data.value = cloneDeep(props.value);
  } else {
    data.value = [
      {
        key: '',
        value: '',
      },
    ];
  }
};

const handleAddLabel = () => {
  if (props.disabled) return;
  data.value.push({
    key: '',
    value: '',
  });
};

const handleDeleteLabel = (index: number) => {
  if (data.value.length === 1 || props.disabled) return;
  data.value.splice(index, 1);
};

onBeforeMount(() => {
  handleSetData();
});
</script>
