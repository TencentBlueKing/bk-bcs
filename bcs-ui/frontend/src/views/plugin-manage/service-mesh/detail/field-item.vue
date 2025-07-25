<template>
  <div class="text-[12px]">
    <div class="text-[#979BA5] leading-[20px] mb-[4px]">{{ label || '--' }}</div>
    <slot>
      <div class="leading-[20px] break-words">{{ value || '--' }}</div>
    </slot>
  </div>
</template>
<script setup lang="ts">
defineProps({
  label: {
    type: String,
    default: '',
  },
  value: {
    type: [String, Number],
    default: '',
  },
});
</script>
