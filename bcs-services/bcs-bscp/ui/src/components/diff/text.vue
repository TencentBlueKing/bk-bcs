<script setup lang="ts">
  import hljs from 'highlight.js';

  const props = defineProps<{
    current: string,
    base: string
  }>()
</script>
<template>
  <section class="text-diff-wrapper">
    <bk-code-diff
      language="json"
      diff-format="side-by-side"
      theme="dark"
      :hljs="hljs"
      :old-content="base"
      :new-content="current"
    />
  </section>
</template>
<style lang="scss" scoped>
  .text-diff-wrapper {
    height: 100%;
  }
  .bk-code-diff {
    height: 100%;
  }
  :deep(.d2h-file-wrapper) {
    border: none;
  }
</style>