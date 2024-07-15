<template>
  <div
    :class="['transition-all rounded-sm',active ? 'bg-[#FFF3E1]' : '']"
    @mouseenter="handleMouseEnter"
    @mouseleave="handleMouseLeave">
    <div class="flex items-center justify-between text-[12px] mb-[6px] leading-[20px]">
      <span>{{ title }}</span>
      <span class="text-[#979BA5]" v-if="isHover">
        <i class="bcs-icon bcs-icon-close3-shape cursor-pointer ml-[10px]" v-if="deletable" @click="deleteField"></i>
      </span>
    </div>
    <slot></slot>
  </div>
</template>
<script setup lang="ts">
import { ref } from 'vue';
defineProps({
  title: {
    type: String,
  },
  top: {
    type: Boolean,
    default: false,
  },
  deletable: {
    type: Boolean,
    default: true,
  },
  active: {
    type: Boolean,
    default: false,
  },
});

const emits = defineEmits(['delete']);

const isHover = ref(false);
const handleMouseEnter = () => {
  isHover.value = true;
};
const handleMouseLeave = () => {
  isHover.value = false;
};

const deleteField = () => {
  emits('delete');
};
</script>
<style scoped lang="postcss">
.is-top {
  display: inline-flex;
  transform: rotate(180deg);
}
</style>
