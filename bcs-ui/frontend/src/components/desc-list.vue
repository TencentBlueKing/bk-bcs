<template>
  <div class="p-[16px] max-w-[800px]">
    <div class="h-[22px] text-[14px] text-[#313238] font-bold">{{ title }}</div>
    <div class="flex flex-wrap mt-[16px] text-[12px]">
      <slot></slot>
    </div>
  </div>
</template>
<script setup lang="ts">
defineProps({
  title: String,
});
</script>
