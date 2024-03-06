<template>
  <div class="matching-result">
    <div class="head-area">
      <div class="head-left">
        <bk-overflow-title v-if="rule" class="rule" type="tips">
          {{ rule.appName + rule.scopeContent }}
        </bk-overflow-title>
        <span class="result">匹配结果</span>
      </div>
      <div class="totle">共{{ 0 }}项</div>
    </div>
    <bk-input class="search-input" placeholder="配置项名称">
      <template #suffix>
        <Search class="search-input-icon" />
      </template>
    </bk-input>
    <bk-table empty-text="请先在左侧表单设置关联规则并预览" :border="['outer']" :pagination="pagination">
      <bk-table-column label="配置项"></bk-table-column>
    </bk-table>
  </div>
</template>
<script setup lang="ts">
  import { ref } from 'vue';
  import { IPreviewRule } from '../../../../../types/credential';
  import { Search } from 'bkui-vue/lib/icon';

  defineProps<{
    rule: IPreviewRule | null;
  }>();

  const pagination = ref({
    count: 0,
    current: 1,
    limit: 10,
  });
</script>
<style lang="scss" scoped>
  .matching-result {
    width: 100%;
    .head-area {
      display: flex;
      align-items: center;
      justify-content: space-between;
      margin-bottom: 12px;
      color: #63656e;
      font-size: 12px;
      .head-left {
        display: flex;
        align-items: center;
        .rule {
          line-height: 16px;
          max-width: 150px;
        }
        .result {
          font-weight: 700;
        }
      }
      .totle {
        width: 70px;
        height: 24px;
        background: #eaebf0;
        border-radius: 12px;
        text-align: center;
        line-height: 24px;
      }
    }
    .search-input {
      margin-bottom: 16px;
      .search-input-icon {
        padding-right: 10px;
        color: #979ba5;
        background: #ffffff;
      }
    }
  }
</style>
