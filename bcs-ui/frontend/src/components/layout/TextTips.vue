<template>
  <span
    :class="{ 'bcs-border-tips': !!tooltips }"
    v-bk-tooltips="{
      disabled: !tooltips,
      content: tooltips
    }">
    {{ text }}
  </span>
</template>
<script setup lang="ts">
defineProps({
  text: String,
  tooltips: String,
});
</script>
