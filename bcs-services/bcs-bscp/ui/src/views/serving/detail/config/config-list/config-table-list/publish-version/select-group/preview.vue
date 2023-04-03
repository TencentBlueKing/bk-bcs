<script setup lang="ts">
</script>
<template>
  <div class="roll-back-preview">
    <h3 class="title">
      回滚预览
      <span class="tips">回滚后，相关分组将退回至以下各版本</span>
    </h3>
    <div class="version-list-wrapper">
      <bk-exception scene="part" type="empty">没有数据</bk-exception>
    </div>
  </div>
</template>
<style lang="scss" scoped>
  .title {
    margin: 0 0 16px;
    line-height: 19px;
    font-size: 14px;
    font-weight: 700;
    color: #63656e;
    .tips {
      margin-left: 16px;
      line-height: 20px;
      color: #979ba5;
      font-size: 12px;
      font-weight: 400;
    }
  }
</style>