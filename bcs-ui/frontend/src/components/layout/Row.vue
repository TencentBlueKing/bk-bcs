<template>
  <div :class="type">
    <div class="left"><slot name="left"></slot></div>
    <div class="right"><slot name="right"></slot></div>
  </div>
</template>
<script lang="ts">
import { defineComponent, PropType } from 'vue';

export type RowType = 'space-between';
export default defineComponent({
  name: 'LayoutRow',
  props: {
    type: {
      type: String as PropType<RowType>,
      default: 'space-between',
    },
  },
  setup() {
    return {};
  },
});
</script>
<style lang="postcss">
.space-between {
  display: flex;
  align-items: center;
  justify-content: space-between;
  .left {
    display: flex;
    align-items: center;
  }
  .right {
    display: flex;
    align-items: center;
  }
}
</style>
