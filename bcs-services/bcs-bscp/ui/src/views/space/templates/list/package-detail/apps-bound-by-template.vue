<template>
  <bk-sideslider :title="t('被引用')" :width="640" :quick-close="true" :is-show="isShow" @closed="close">
    <div class="top-area">
      <div class="config-name">{{ props.config.name }}</div>
      <SearchInput :placeholder="t('服务名称/配置文件版本')" v-model="searchStr" :width="320" @search="handleSearch" />
    </div>
    <div class="apps-table">
      <bk-table
        :border="['outer']"
        :data="appList"
        :remote-pagination="true"
        :pagination="pagination"
        @page-limit-change="handlePageLimitChange"
        @page-value-change="getList">
        <bk-table-column :label="t('配置文件版本')" prop="template_revision_name"></bk-table-column>
        <bk-table-column :label="t('引用此配置文件的服务')">
          <template #default="{ row }">
            <bk-link v-if="row.app_id" class="link-btn" theme="primary" target="_blank" :href="getHref(row.app_id)">
              {{ row.app_name }}
            </bk-link>
          </template>
        </bk-table-column>
        <template #empty>
          <table-empty :is-search-empty="isSearchEmpty" @clear="handleClearSearchStr"></table-empty>
        </template>
      </bk-table>
    </div>
    <div class="action-btn">
      <bk-button @click="close">{{ t('关闭') }}</bk-button>
    </div>
  </bk-sideslider>
</template>
<script lang="ts" setup>
  import { ref, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { useRouter } from 'vue-router';
  import { getUnNamedVersionAppsBoundByTemplate } from '../../../../../api/template';
  import { ICommonQuery } from '../../../../../../types/index';
  import { IAppBoundByTemplateDetailItem } from '../../../../../../types/template';
  import useTablePagination from '../../../../../utils/hooks/use-table-pagination';
  import SearchInput from '../../../../../components/search-input.vue';
  import tableEmpty from '../../../../../components/table/table-empty.vue';

  const router = useRouter();
  const { t } = useI18n();

  const { pagination, updatePagination } = useTablePagination('appBoundByTemplate');

  const props = defineProps<{
    show: boolean;
    spaceId: string;
    currentTemplateSpace: number;
    config: { id: number; name: string };
  }>();

  const emits = defineEmits(['update:show']);

  const isShow = ref(false);
  const appList = ref<IAppBoundByTemplateDetailItem[]>([]);
  const searchStr = ref('');
  const loading = ref(false);
  const isSearchEmpty = ref(false);

  watch(
    () => props.show,
    (val) => {
      isShow.value = val;
      if (val) {
        searchStr.value = '';
        getList();
      }
    },
  );

  const getList = async () => {
    loading.value = true;
    const params: ICommonQuery = {
      start: (pagination.value.current - 1) * pagination.value.limit,
      limit: pagination.value.limit,
    };
    if (searchStr.value) {
      params.search_fields = 'app_name,template_revision_name';
      params.search_value = searchStr.value;
    }
    const res = await getUnNamedVersionAppsBoundByTemplate(
      props.spaceId,
      props.currentTemplateSpace,
      props.config.id,
      params,
    );
    appList.value = res.details;
    pagination.value.count = res.count;
    loading.value = false;
  };

  const getHref = (id: number) => {
    const { href } = router.resolve({ name: 'service-config', params: { spaceId: props.spaceId, appId: id } });
    return href;
  };

  const handleSearch = (val: string) => {
    isSearchEmpty.value = true;
    searchStr.value = val;
    pagination.value.current = 1;
    getList();
  };

  const handlePageLimitChange = (val: number) => {
    updatePagination('limit', val);
    getList();
  };

  const close = () => {
    emits('update:show', false);
  };

  const handleClearSearchStr = () => {
    searchStr.value = '';
    isSearchEmpty.value = false;
    getList();
  };
</script>
<style lang="scss" scoped>
  .top-area {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin: 24px 0;
    padding: 0 24px;
    height: 32px;
    .config-name {
      margin-right: 8px;
      font-size: 14px;
      font-weight: 700;
      color: #63656e;
      overflow: hidden;
      white-space: nowrap;
      text-overflow: ellipsis;
    }
    .search-input {
      flex-shrink: 0;
    }
  }
  .apps-table {
    padding: 0 24px;
    height: calc(100vh - 180px);
    overflow: auto;
  }
  .link-btn {
    font-size: 12px;
    color: #3a84ff;
  }
  .action-btn {
    padding: 8px 24px;
    background: #fafbfd;
    box-shadow: 0 -1px 0 0 #dcdee5;
    .bk-button {
      min-width: 88px;
    }
  }
</style>
