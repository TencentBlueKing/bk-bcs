<template>
  <div class="bk-spin-loading" :class="extCls">
    <div class="rotate rotate1"></div>
    <div class="rotate rotate2"></div>
    <div class="rotate rotate3"></div>
    <div class="rotate rotate4"></div>
    <div class="rotate rotate5"></div>
    <div class="rotate rotate6"></div>
    <div class="rotate rotate7"></div>
    <div class="rotate rotate8"></div>
  </div>
</template>

<script>
export default {
  props: {
    extCls: {
      type: String,
      default: 'bk-spin-loading-primary',
    },
  },
};
</script>
