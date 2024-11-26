<template>
  <div class="bg-[#fff] h-full overflow-y-auto overflow-x-hidden">
    <bcs-exception
      type="empty"
      scene="part"
      v-if="list.length === 0">
    </bcs-exception>
    <div
      v-else
      :class="[
        'h-full py-[0 8px] border-r-[1px] border-solid bg-[#fff] w-[230px] pt-[12px]'
      ]">
      <div
        v-for="(item, index) in list"
        :key="index"
        :class="[
          'flex items-center cursor-pointer leading-[20px]',
          'h-[32px] text-[12px] px-[12px]',
          activeIndex === index ? 'bg-[#e1ecff] text-[#3a84ff] border-r-2 border-[#3a84ff]' : ''
        ]"
        @click="handleAnchor(item, index)">
        <span
          :class="[
            'rounded-full w-2.5 h-2.5 bg-[red] border-2 border-white flex-shrink-0',
            validArray[index] ? 'visible' : 'invisible'
          ]"></span>
        <span
          :class="[
            'rounded-full w-4 h-4 leading-[1rem] text-center text-[#fff] mx-2 flex-shrink-0',
            activeIndex === index ? 'bg-[#3a84ff] text-[#fff]' : 'bg-[#979ba5]'
          ]">{{ index + 1 }}</span>
        <span class="bcs-ellipsis" v-bk-overflow-tips>
          <slot :item="item">
            {{ item.name || $t('templateFile.label.untitled') }}
          </slot>
        </span>
      </div>
    </div>
  </div>
</template>
<script setup lang="ts">
import { PropType } from 'vue';

defineProps({
  list: {
    type: Array as PropType<{name: string, offset?: number}[]>,
    default: () => [],
  },
  activeIndex: {
    type: Number,
    default: 0,
  },
  validArray: {
    type: Array,
    default: () => [],
  },
});

const emits = defineEmits(['cellClick']);

function handleAnchor(item, index) {
  emits('cellClick', { item, index });
}
</script>
