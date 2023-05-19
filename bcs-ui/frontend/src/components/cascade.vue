<template>
  <bcs-dropdown-menu trigger="click" @hide="handleHideDropDown" @show="handleShowDropDown">
    <template #dropdown-trigger>
      <slot></slot>
    </template>
    <template #dropdown-content>
      <BcsCascadeItem ref="cascadeRef" :list="list" @click="handleClickItem" />
    </template>
  </bcs-dropdown-menu>
</template>
<script lang="ts">
import {  defineComponent, PropType, ref } from 'vue';
import BcsCascadeItem from './cascade-item.vue';

export interface IData {
  id: string
  label: string
  children?: IData[]
}

export default defineComponent({
  name: 'BcsCascade',
  components: { BcsCascadeItem },
  props: {
    list: {
      type: Array as PropType<IData[]>,
      default: () => ([]),
    },
  },
  setup(props, ctx) {
    const cascadeRef = ref<any>(null);
    const handleClickItem = (item: IData) => {
      ctx.emit('click', item);
      cascadeRef.value?.hide();
    };
    const handleShowDropDown = () => {
      ctx.emit('show');
    };
    const handleHideDropDown = () => {
      cascadeRef.value?.hide();
      ctx.emit('hide');
    };
    return {
      cascadeRef,
      handleClickItem,
      handleShowDropDown,
      handleHideDropDown,
    };
  },
});
</script>
