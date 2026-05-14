<template>
  <div class="bg-[#f0f1f4] flex items-center p-[4px] text-[12px]">
    <div
      v-for="(item, index) in items"
      :key="item.value || index"
      v-bk-tooltips="{
        content: item.tips,
        disabled: !item.tips
      }"
      :class="[
        'py-[4px] px-[8px] rounded-sm transition-all relative cursor-pointer',
        {
          'bg-[#fff] text-[#3a84ff]': modelValue === item.value && !item.disabled,
          'cursor-not-allowed opacity-50': item.disabled,
          'before:absolute before:top-1/2 before:-translate-y-1/2 before:-right-px before:w-px before:h-[12px]':
            modelValue !== item.value && index < items.length - 1,
          'before:bg-[#d4dee5]': modelValue !== item.value && index < items.length - 1
        }
      ]"
      @click="handleClick(item)">
      <span :class="{ 'border-b border-dashed border-current': item.tips }">{{ item.label }}</span>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent, PropType } from 'vue';

export interface TabItem {
  label: string;
  value: string | number;
  disabled?: boolean;
  tips?: string;
}

export default defineComponent({
  name: 'TabButtonGroup',
  model: {
    prop: 'modelValue',
    event: 'update:modelValue',
  },
  props: {
    modelValue: {
      type: [String, Number] as PropType<string | number>,
      required: true,
    },
    items: {
      type: Array as PropType<TabItem[]>,
      default: () => [],
    },
  },
  setup(props, { emit }) {
    const handleClick = (item: TabItem) => {
      if (item.disabled) {
        return;
      }
      if (props.modelValue !== item.value) {
        emit('update:modelValue', item.value);
        emit('change', item.value);
      }
    };

    return {
      handleClick,
    };
  },
});
</script>
