<script setup lang="ts">

  const props = withDefaults(defineEmits<{
    editable?: boolean;
  }>(), {
    editable: true
  })
</script>
<template>
  <section class="script-content"></section>
</template>