<template>
  <div class="matching-result">
    <div class="head-area">
      <div class="head-left">
        <bk-overflow-title v-if="rule" class="rule" type="tips">
          {{ rule.appName + rule.scopeContent }}
        </bk-overflow-title>
        <span class="result">{{ t('匹配结果') }}</span>
      </div>
      <div class="totle">{{ t('共') }} {{ pagination.count }} {{ t('项') }}</div>
    </div>
    <SearchInput v-model="searchStr" :placeholder="inputPlaceholder" @search="loadCredentialRulePreviewList" />
    <bk-loading :loading="listLoading">
      <bk-table
        :empty-text="tableEmptyText"
        :data="tableData"
        :border="['outer']"
        :remote-pagination="true"
        :pagination="pagination"
        :key="isFileType"
        @page-value-change="loadCredentialRulePreviewList">
        <bk-table-column :label="isFileType ? t('配置文件名') : t('配置项')">
          <template #default="{ row }">
            <div v-if="row.name">
              {{ isFileType ? fileAP(row) : row.name }}
            </div>
          </template>
        </bk-table-column>
        <template #empty>
          <TableEmpty :empty-title="tableEmptyText" :is-search-empty="isSearchEmpty" @clear="handleClearSearchStr" />
        </template>
      </bk-table>
    </bk-loading>
  </div>
</template>
<script setup lang="ts">
  import { computed, ref, watch } from 'vue';
  import { IPreviewRule } from '../../../../../types/credential';
  import { getCredentialPreview } from '../../../../api/credentials';
  import useTablePagination from '../../../../utils/hooks/use-table-pagination';
  import SearchInput from '../../../../components/search-input.vue';
  import TableEmpty from '../../../../components/table/table-empty.vue';
  import { useI18n } from 'vue-i18n';

  const { t } = useI18n();

  const { pagination } = useTablePagination('clientPullRecord', {
    small: true,
    showTotalCount: false,
    showLimit: false,
    align: 'center',
  });

  const props = defineProps<{
    rule: IPreviewRule | null;
    bkBizId: string;
  }>();

  const isFileType = ref(false);
  const isSearchEmpty = ref(false);
  const listLoading = ref(false);

  watch(
    () => props.rule,
    (val) => {
      if (val) {
        loadCredentialRulePreviewList();
      }
    },
    { deep: true },
  );

  // 配置文件名
  const fileAP = computed(() => ({ name, path }: { name: string; path: string }) => {
    if (path.endsWith('/')) {
      return `${path}${name}`;
    }
    return `${path}/${name}`;
  });

  const inputPlaceholder = computed(() => (isFileType.value ? t('请输入配置文件名') : t('请输入配置项名称')));

  const tableEmptyText = computed(() => {
    return props.rule?.appName ? t('没有匹配到配置项') : t('请先在左侧表单设置关联规则并预览');
  });

  const searchStr = ref('');
  const tableData = ref();

  const loadCredentialRulePreviewList = async () => {
    listLoading.value = true;
    isSearchEmpty.value = searchStr.value !== '';
    const params = {
      start: (pagination.value.current - 1) * pagination.value.limit,
      limit: pagination.value.limit,
      app_name: props.rule!.appName,
      scope: props.rule!.scopeContent,
      search_value: searchStr.value,
    };
    try {
      const res = await getCredentialPreview(props.bkBizId, params);
      pagination.value.count = res.data.count;
      tableData.value = res.data.details;
      isFileType.value = !!tableData.value[0]?.path;
    } catch (error) {
      console.error(error);
    } finally {
      listLoading.value = false;
    }
  };

  const handleClearSearchStr = () => {
    searchStr.value = '';
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
        line-height: 16px;
        .rule {
          max-width: 150px;
        }
        .result {
          font-weight: 700;
        }
      }
      .totle {
        padding: 0 16px;
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
