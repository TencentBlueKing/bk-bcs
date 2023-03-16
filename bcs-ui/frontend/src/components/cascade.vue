<template>
  <PopoverSelector
    trigger="manual"
    placement="bottom-start"
    offset="0, 5"
    :on-hide="handleHidePopover"
    ref="popoverRef">
    <div @mouseenter="handleShowPopover" @click="handleShowPopover">
      <slot></slot>
    </div>
    <template #content>
      <BcsCascadeItem ref="cascadeRef" :list="list" @click="handleClickItem" />
    </template>
  </PopoverSelector>
</template>
<script lang="ts">
import {  defineComponent, PropType, ref } from '@vue/composition-api';
import PopoverSelector from './popover-selector.vue';
import BcsCascadeItem from './cascade-item.vue';

export interface IData {
  id: string
  label: string
  children?: IData[]
}

export default defineComponent({
  name: 'BcsCascade',
  components: { PopoverSelector, BcsCascadeItem },
  props: {
    list: {
      type: Array as PropType<IData[]>,
      default: () => ([]),
    },
  },
  setup(props, ctx) {
    const popoverRef = ref<any>(null);
    const cascadeRef = ref<any>(null);
    const handleClickItem = (item: IData) => {
      ctx.emit('click', item);
      cascadeRef.value?.hide();
    };
    const handleShowPopover = () => {
      popoverRef.value?.show();
    };
    const handleHidePopover = () => {
      cascadeRef.value?.hide();
    };
    return {
      popoverRef,
      cascadeRef,
      handleClickItem,
      handleShowPopover,
      handleHidePopover,
    };
  },
});
</script>
