<template>
  <bcs-popover
    theme="light navigation-message"
    :offset="offset"
    placement="bottom"
    :arrow="false"
    :trigger="trigger"
    :tippy-options="{ boundary: 'window' }">
    <slot></slot>
    <template slot="content">
      <slot name="content"></slot>
    </template>
  </bcs-popover>
</template>
<script lang="ts">
import { defineComponent } from '@vue/composition-api';

export default defineComponent({
  name: 'PopoverSelector',
  props: {
    trigger: {
      type: String,
      default: 'click',
    },
    offset: {
      type: String,
      default: '0, 20',
    },
    placement: {
      type: String,
      default: 'bottom',
    },
  },
  setup() {
    return {};
  },
});
</script>
