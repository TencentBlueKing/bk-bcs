<template>
  <div
    class="flex items-center cursor-pointer text-[#979ba5] h-[32px]"
    @click="handleClick">
    <span
      class="ml-[-6px] transition-all relative top-[-1px]"
      :style="collapse ? 'transform: rotate(-90deg);' : 'transform: rotate(0deg);'">
      <i class="bcs-icon bcs-icon-down-shape"></i>
    </span>
    <span class="ml-[8px]">{{ title }}</span>
  </div>
</template>
<script lang="ts">
import { defineComponent } from '@vue/composition-api';

export default defineComponent({
  name: 'CollapseTitle',
  props: {
    title: {
      type: String,
      default: '',
    },
    collapse: {
      type: Boolean,
      default: false,
    },
  },
  setup(props, ctx) {
    const handleClick = () => {
      ctx.emit('click');
    };
    return {
      handleClick,
    };
  },
});
</script>
