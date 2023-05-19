<template>
  <ul>
    <li
      :class="[
        'bcs-dropdown-item',
        {
          'pr-[5px]': item.children && item.children.length,
          'active': item.children && item.children.length && activeID === item.id,
          'disabled': item.disabled
        }
      ]"
      v-for="item, index in list"
      :key="item.id"
      @mouseenter="handleMouseEnter(item, index)"
      @click="handleClickItem(item, index)">
      <PopoverSelector
        trigger="manual"
        placement="right-start"
        offset="10, -4"
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
import { defineComponent, PropType, ref } from 'vue';
import PopoverSelector from './popover-selector.vue';

export interface IData {
  id: string
  label: string
  children?: IData[]
  disabled?: boolean
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
      if (item.disabled) return;
      activeID.value = item.id;
      if (!item.children?.length) return;
      popoverRef.value[index]?.show();
    };
    const handleClickItem = (item: IData, index) => {
      if (item.disabled) return;

      popoverRef.value[index]?.show();
      if (item.children?.length) return;
      ctx.emit('click', item);
      popoverRef.value[index]?.hide();
    };
    const hide = () => {
      setTimeout(() => {
        props.list.forEach((_, index) => {
          popoverRef.value[index]?.hide();
        });
        childrenItemRef.value?.[0]?.hide();
        activeID.value = '';
      }, 0);
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
