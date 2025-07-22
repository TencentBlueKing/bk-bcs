<template>
  <div>
    <div class="bg-[#F0F1F5] px-[8px] py-[4px] rounded-sm flex items-center">
      <slot name="label">
        <span>{{ label }}</span>
      </slot>
      <bcs-switcher
        v-if="!readonly"
        class="ml-[10px]"
        :value="active"
        size="small"
        @change="handleChange"></bcs-switcher>
    </div>
    <div class="border border-t-0 p-[16px]" v-show="active || readonly">
      <slot></slot>
    </div>
  </div>
</template>
<script lang="ts">
import { defineComponent, ref, watch } from 'vue';

export default defineComponent({
  name: 'ContentSwitcher',
  model: {
    prop: 'modelValue',
    event: 'update:modelValue',
  },
  props: {
    label: {
      type: String,
      default: '',
    },
    modelValue: {
      type: Boolean,
      default: false,
    },
    readonly: {
      type: Boolean,
      default: false,
    },
  },
  setup(props, ctx) {
    const active = ref<boolean>(props.modelValue);
    const handleChange = (value: boolean) => {
      ctx.emit('update:modelValue', value);
      ctx.emit('change', value);
    };

    watch(() => props.modelValue, () => {
      active.value = props.modelValue;
    });

    return {
      active,
      handleChange,
    };
  },
});
</script>
