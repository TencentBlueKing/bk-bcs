<template>
  <div>
    <Row
      :class="[
        'h-[28px] text-[#313238] bg-[#F5F7FA] flex items-center px-[16px]',
        {
          'text-[12px]': size === 'small',
          'text-[14px]': size === 'middle'
        }
      ]">
      <template #left>{{ title }}</template>
      <template #right><slot name="title-right"></slot></template>
    </Row>
    <div
      :class="[
        'mt-[16px]',
        {
          'text-[12px]': size === 'small',
          'text-[14px]': size === 'middle'
        }
      ]">
      <slot></slot>
    </div>
  </div>
</template>
<script setup lang="ts">
import Row from './layout/Row.vue';
defineProps({
  title: String,
  size: {
    type: String,
    default: 'small',
  },
});
</script>
