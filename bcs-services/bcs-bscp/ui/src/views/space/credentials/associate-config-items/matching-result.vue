<template>
  <div class="matching-result">
    <div class="head-area">
      <span class="title">匹配结果</span>
      <bk-input class="search-input" placeholder="关联规则/服务名称/配置文件名称"></bk-input>
    </div>
    <bk-table empty-text="请先在左侧表单设置关联规则并预览" :border="['outer']" :pagination="pagination">
      <bk-table-column label="关联规则" :filter="true"></bk-table-column>
      <bk-table-column label="服务名称" :filter="true"></bk-table-column>
      <bk-table-column label="配置文件" :filter="true"></bk-table-column>
    </bk-table>
  </div>
</template>
<script setup lang="ts">
  import { ref } from 'vue';

  const pagination = ref({
    count: 0,
    current: 1,
    limit: 10,
  });
</script>
<style lang="scss" scoped>
  .head-area {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 16px;
    .title {
      line-height: 16px;
      font-size: 12px;
      font-weight: 700;
      color: #63656e;
    }
    .search-input {
      width: 360px;
    }
  }
</style>
