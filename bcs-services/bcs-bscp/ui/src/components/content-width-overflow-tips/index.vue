<template>
  <bk-popover placement="top" :disabled="disabled">
    <div ref="containerRef" class="overflow-popover-container">
      <slot></slot>
    </div>
    <template #content>
      <slot></slot>
    </template>
  </bk-popover>
</template>
<script lang="ts" setup>
  import { ref, onMounted, onBeforeUnmount } from 'vue';

  const containerRef = ref();
  const disabled = ref(true);

  onMounted(() => {
    calcPopover();
    // 监听容器宽度变化，重新设置popover激活态
    const observer = new ResizeObserver(calcPopover);
    observer.observe(containerRef.value);
    onBeforeUnmount(() => {
      containerRef.value && observer?.unobserve(containerRef.value);
      observer?.disconnect();
    });
  });

  // 计算内容宽度是否超出容器宽度，超出则激活popover
  const calcPopover = () => {
    const contentEl = containerRef.value.firstElementChild;

    if (contentEl) {
      console.log(contentEl.scrollWidth, containerRef.value.clientWidth);
    }
    if (contentEl && contentEl.scrollWidth > containerRef.value.clientWidth) {
      disabled.value = false;
    } else {
      disabled.value = true;
    }
  };
</script>
<style lang="scss" scoped>
  .overflow-popover-container {
    overflow: hidden;
  }
</style>
