<template>
  <bcs-popover
    :theme="`light ${theme}`"
    :offset="offset"
    :placement="placement"
    :arrow="false"
    :trigger="trigger"
    :tippy-options="tippyOptions"
    :always="always"
    :on-hide="onHide"
    :on-show="onShow"
    :disabled="disabled"
    ref="popoverRef">
    <slot></slot>
    <template slot="content">
      <div class="py-[4px] bg-[#fff]">
        <slot name="content"></slot>
      </div>
    </template>
  </bcs-popover>
</template>
<script lang="ts">
import { defineComponent, ref } from 'vue';

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
    always: {
      type: Boolean,
      default: false,
    },
    tippyOptions: {
      type: Object,
      default: () => ({}),
    },
    onHide: Function,
    onShow: Function,
    disabled: {
      type: Boolean,
      default: false,
    },
    theme: {
      type: String,
      default: 'navigation-message',
    },
  },
  setup() {
    const popoverRef = ref<any>(null);
    const show = () => {
      popoverRef.value.showHandler();
    };
    const hide = () => {
      popoverRef.value.hideHandler();
    };
    return {
      popoverRef,
      show,
      hide,
    };
  },
});
</script>
