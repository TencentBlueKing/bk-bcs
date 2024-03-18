<template>
  <div class="matching-result">
    <div class="head-area">
      <div class="head-left">
        <bk-overflow-title v-if="rule" class="rule" type="tips">
          {{ rule.appName + rule.scopeContent }}
        </bk-overflow-title>
        <span class="result">匹配结果</span>
      </div>
      <div class="totle">共 {{ 0 }} 项</div>
    </div>
    <SearchInput v-model="searchStr" :placeholder="'请输入配置项名称'" @search="loadCredentialRulePreviewList" />
    <bk-table
      :empty-text="tableEmptyText"
      :data="tableData"
      :border="['outer']"
      :remote-pagination="true"
      :pagination="pagination"
      @page-limit-change="handlePageLimitChange"
      @page-value-change="loadCredentialRulePreviewList">
      <bk-table-column label="配置项" prop="name"></bk-table-column>
    </bk-table>
  </div>
</template>
<script setup lang="ts">
  import { computed, ref, watch } from 'vue';
  import { IPreviewRule } from '../../../../../types/credential';
  import { getCredentialPreview } from '../../../../api/credentials';
  import SearchInput from '../../../../components/search-input.vue';

  const props = defineProps<{
    rule: IPreviewRule | null;
    bkBizId: string;
  }>();

  watch(
    () => props.rule?.id,
    () => {
      loadCredentialRulePreviewList();
    },
  );

  const tableEmptyText = computed(() => (props.rule?.id ? '暂无数据' : '请先在左侧表单设置关联规则并预览'));

  const searchStr = ref('');
  const tableData = ref();
  const pagination = ref({
    count: 0,
    current: 1,
    limit: 10,
    small: true,
    showTotalCount: false,
    showLimit: false,
    align: 'center',
  });

  const loadCredentialRulePreviewList = async () => {
    const params = {
      start: (pagination.value.current - 1) * pagination.value.limit,
      limit: pagination.value.limit,
      app_name: props.rule!.appName,
      scope: props.rule!.scopeContent,
      search_value: searchStr.value,
    };
    const res = await getCredentialPreview(props.bkBizId, params);
    pagination.value.count = res.data.count;
    tableData.value = res.data.details;
  };

  // 更改每页条数
  const handlePageLimitChange = (val: number) => {
    pagination.value.limit = val;
    loadCredentialRulePreviewList();
  };
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
