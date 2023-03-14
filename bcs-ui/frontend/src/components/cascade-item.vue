<template>
  <ul class="bcs-border">
    <li
      :class="[
        'bcs-dropdown-item',
        {
          'pr-[5px]': item.children && item.children.length,
          'active': item.children && item.children.length && activeID === item.id
        }
      ]"
      v-for="item, index in list"
      :key="item.id"
      @mouseenter="handleMouseEnter(item, index)"
      @click="handleClickItem(item, index)">
      <PopoverSelector
        trigger="manual"
        placement="right-start"
        offset="20, -4"
        :tippy-options="{ flip: false }"
        ref="popoverRef">
        <div
          class="flex place-content-center">
          <span>{{ item.label }}</span>
          <span class="text-[20px] ml-[5px]" v-if="item.children && item.children.length">
            <i class="bk-icon icon-angle-right"></i>
          </span>
        </div>
        <template #content>
          <BcsCascadeItem
            :list="item.children"
            v-if="item.children && item.children.length && activeID === item.id"
            ref="childrenItemRef"
            @click="handleClickItem"
          />
        </template>
      </PopoverSelector>
    </li>
  </ul>
</template>
<script lang="ts">
import { defineComponent, PropType, ref } from '@vue/composition-api';
import PopoverSelector from './popover-selector.vue';

export interface IData {
  id: string
  label: string
  children?: IData[]
}
export default defineComponent({
  name: 'BcsCascadeItem',
  components: { PopoverSelector },
  props: {
    list: {
      type: Array as PropType<IData[]>,
      default: () => ([]),
    },
  },
  setup(props, ctx) {
    const popoverRef = ref<any[]>([]);
    const childrenItemRef = ref<any>(null);
    const activeID = ref('');
    const handleMouseEnter = (item: IData, index) => {
      activeID.value = item.id;
      if (!item.children?.length) return;
      popoverRef.value[index]?.show();
    };
    const handleClickItem = (item: IData, index) => {
      if (item.children?.length) return;
      ctx.emit('click', item);
      popoverRef.value[index]?.hide();
    };
    const hide = () => {
      props.list.forEach((_, index) => {
        popoverRef.value[index]?.hide();
      });
      childrenItemRef.value?.[0]?.hide();
      activeID.value = '';
    };

    return {
      popoverRef,
      childrenItemRef,
      activeID,
      handleClickItem,
      handleMouseEnter,
      hide,
    };
  },
});
</script>
