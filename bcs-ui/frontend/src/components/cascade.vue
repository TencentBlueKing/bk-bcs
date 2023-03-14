<template>
  <PopoverSelector trigger="manual" placement="bottom-start" offset="0, 15" ref="popoverRef">
    <div @mouseenter="handleMouseEnter">
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
    const handleMouseEnter = () => {
      popoverRef.value?.show();
    };
    return {
      popoverRef,
      cascadeRef,
      handleClickItem,
      handleMouseEnter,
    };
  },
});
</script>
