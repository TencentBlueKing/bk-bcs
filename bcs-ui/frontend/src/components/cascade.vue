<template>
  <PopoverSelector trigger="mouseenter" placement="bottom-start" offset="0, 15">
    <slot></slot>
    <template #content>
      <BcsCascadeItem :list="list" @click="handleClickItem" />
    </template>
  </PopoverSelector>
</template>
<script lang="ts">
import {  defineComponent, PropType } from '@vue/composition-api';
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
    const handleClickItem = (item: IData) => {
      ctx.emit('click', item);
    };
    return {
      handleClickItem,
    };
  },
});
</script>
