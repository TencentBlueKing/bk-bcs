<template>
  <div :id="id">
    <div
      :class="[
        'flex items-center cursor-pointer text-[12px] px-[8px] hover:bg-[#F5F7FA]',
        active ? 'bg-[#E1ECFF] text-[#3A84FF]' : ''
      ]"
      v-bind="$attrs"
      v-if="showTitle"
      @click="handleTitleClick">
      <LoadingIcon v-if="loading" />
      <span
        class="transition-all relative top-[-1px] text-[#979BA5] mr-[6px]"
        :style="collapse ? 'transform: rotate(-90deg);' : 'transform: rotate(0deg);'"
        v-else
        @click="handleToggleCollapse">
        <i class="bcs-icon bcs-icon-down-shape"></i>
      </span>
      <slot name="title">
        {{ title }}
      </slot>
    </div>
    <div v-show="!collapse">
      <slot></slot>
    </div>
  </div>
</template>
<script setup lang="ts">
import LoadingIcon from '@/components/loading-icon.vue';
const props = defineProps({
  // 是否折叠
  collapse: {
    type: Boolean,
    default: true,
  },
  active: {
    type: Boolean,
    default: false,
  },
  title: {
    type: String,
    default: '',
  },
  showTitle: {
    type: Boolean,
    default: true,
  },
  id: {
    type: String,
  },
  loading: {
    type: Boolean,
    default: false,
  },
});
const emits = defineEmits(['collapse-change', 'title-click']);


const handleToggleCollapse = () => {
  emits('collapse-change', !props.collapse);
};

const handleTitleClick = () => {
  emits('title-click', props.collapse);
};
</script>
