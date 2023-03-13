<template>
  <div class="flex">
    <ul class="bcs-border">
      <li
        :class="[
          'bcs-dropdown-item',
          {
            'pr-[5px]': item.children && item.children.length,
            'active': item.children && item.children.length && activeID === item.id
          }
        ]"
        v-for="item in list"
        :key="item.id"
        @click="handleClickItem(item)">
        <div class="flex place-content-center">
          <span>{{ item.label }}</span>
          <span class="text-[20px] ml-[5px]" v-if="item.children && item.children.length">
            <i class="bk-icon icon-angle-right"></i>
          </span>
        </div>
      </li>
    </ul>
    <BcsCascadeItem
      :list="activeChildren"
      class="ml-[-1px]"
      @click="handleClickItem"
      v-if="activeChildren.length" />
  </div>
</template>
<script lang="ts">
import { defineComponent, PropType, toRefs, ref, computed } from '@vue/composition-api';

export interface IData {
  id: string
  label: string
  children?: IData[]
}
export default defineComponent({
  name: 'BcsCascadeItem',
  props: {
    list: {
      type: Array as PropType<IData[]>,
      default: () => ([]),
    },
  },
  setup(props, ctx) {
    const { list } = toRefs(props);
    const activeID = ref('');
    const activeChildren = computed(() => list.value.find(item => item.id === activeID.value)?.children || []);
    const handleClickItem = (item: IData) => {
      activeID.value = item.id;
      ctx.emit('click', item);
    };

    return {
      activeID,
      activeChildren,
      handleClickItem,
    };
  },
});
</script>
