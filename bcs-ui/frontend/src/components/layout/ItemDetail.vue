<template>
  <div class="bcs-detail-item">
    <div class="bcs-detail-item-label"></div>
  </div>
</template>
<script lang="ts">
import { defineComponent } from '@vue/composition-api';

export default defineComponent({});
</script>
<style lang="postcss" scoped>
.bcs-detail-item {
  display: flex;
  align-items: flex-start;
  font-size: 14px;
  justify-content: flex-end;
  line-height: 32px;
  &-label {
    align-items: center;
    color: #b2b5bd;
    display: flex;
    flex: 0 0 auto;
    justify-content: flex-end;
    padding-right: 2px;
    text-align: right;
    white-space: pre;
    word-break: keep-all;
  }
  &-content {
    color: #63656e;
    flex: 1;
    word-break: break-all;
  }
}
</style>
