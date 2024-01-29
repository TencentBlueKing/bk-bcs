<template>
  <div class="bcs-body">
    <slot name="header"></slot>
    <section class="content" ref="contentRef">
      <slot></slot>
    </section>
  </div>
</template>
<script setup lang="ts">
import { ref } from 'vue';

import useCalcHeight from '@/composables/use-calc-height';

const contentRef = ref();
useCalcHeight({
  el: contentRef,
  calc: ['#bcs-notice-com', '.bk-navigation-header', '.bcs-body .header-nav'],
});
</script>
<style lang="postcss" scoped>
.bcs-body {
    .content {
        overflow: auto;
        padding: 24px;
    }
}
</style>
