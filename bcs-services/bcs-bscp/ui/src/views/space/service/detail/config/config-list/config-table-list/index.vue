<template>
  <bk-alert
    v-if="isFileType && conflictFileCount > 0"
    theme="warning"
    :title="
      t('模板套餐导入完成，存在 {n} 个冲突配置项，请修改配置项信息或删除对应模板套餐，否则无法生成版本。', {
        n: conflictFileCount,
      })
    " />
  <section class="config-list-wrapper">
    <div class="operate-area">
      <div class="operate-btns">
        <template v-if="versionData.status.publish_status === 'editing'">
          <CreateConfig
            :bk-biz-id="props.bkBizId"
            :app-id="props.appId"
            @created="refreshConfigList"
            @imported="refreshConfigList"
            @uploaded="refreshConfigList(true)" />
          <EditVariables v-if="isFileType" ref="editVariablesRef" :bk-biz-id="props.bkBizId" :app-id="props.appId" />
        </template>
        <ViewVariables
          v-else-if="isFileType"
          :bk-biz-id="props.bkBizId"
          :app-id="props.appId"
          :verision-id="versionData.id" />
        <ConfigExport
          :bk-biz-id="props.bkBizId"
          :app-id="props.appId"
          :version-id="versionData.id"
          :version-name="versionData.spec.name" />
        <BatchOperationBtn
          v-if="versionData.status.publish_status === 'editing'"
          :bk-biz-id="props.bkBizId"
          :app-id="props.appId"
          :selected-ids="selectedIds"
          :is-file-type="isFileType"
          :selected-items="selectedItems"
          @deleted="handleBatchDeleted" />
      </div>
      <SearchInput
        v-model="searchStr"
        class="config-search-input"
        :width="280"
        :placeholder="t('配置文件名/创建人/修改人')" />
    </div>
    <section class="config-list-table">
      <TableWithTemplates
        v-if="isFileType"
        ref="tableRef"
        :bk-biz-id="props.bkBizId"
        :app-id="props.appId"
        :search-str="searchStr"
        @clear-str="clearStr"
        @delete-config="refreshVariable"
        @update-selected-ids="selectedIds = $event"
        @update-selected-items="selectedItems = $event" />
      <TableWithKv
        v-else
        ref="tableRef"
        :bk-biz-id="props.bkBizId"
        :app-id="props.appId"
        :search-str="searchStr"
        @clear-str="clearStr"
        @update-selected-ids="selectedIds = $event" />
    </section>
  </section>
</template>
<script setup lang="ts">
  import { ref } from 'vue';
  import { storeToRefs } from 'pinia';
  import { useI18n } from 'vue-i18n';
  import useConfigStore from '../../../../../../../store/config';
  import useServiceStore from '../../../../../../../store/service';
  import SearchInput from '../../../../../../../components/search-input.vue';
  import CreateConfig from './create-config/index.vue';
  import EditVariables from './variables/edit-variables.vue';
  import ViewVariables from './variables/view-variables.vue';
  import TableWithTemplates from './tables/table-with-templates.vue';
  import TableWithKv from './tables/table-with-kv.vue';
  import ConfigExport from './config-export.vue';
  import BatchOperationBtn from './batch-operation-btn.vue';

  const configStore = useConfigStore();
  const serviceStore = useServiceStore();
  const { versionData, conflictFileCount } = storeToRefs(configStore);
  const { isFileType } = storeToRefs(serviceStore);
  const { t } = useI18n();

  const props = defineProps<{
    bkBizId: string;
    appId: number;
  }>();

  const tableRef = ref();
  const searchStr = ref('');
  const editVariablesRef = ref();
  const selectedIds = ref<number[]>([]);
  const selectedItems = ref<any[]>([]);

  const refreshConfigList = (isBatchUpload = false) => {
    if (isFileType.value) {
      tableRef.value.refresh(isBatchUpload);
      refreshVariable();
    } else {
      tableRef.value.refresh();
    }
  };

  const refreshVariable = () => {
    editVariablesRef.value.getVariableList();
  };

  const clearStr = () => {
    searchStr.value = '';
  };

  // 批量删除配置项回调
  const handleBatchDeleted = () => {
    tableRef.value.refreshAfterBatchSet();
  };

  defineExpose({
    refreshConfigList,
  });
</script>
<style lang="scss" scoped>
  .config-list-wrapper {
    position: relative;
    padding: 0 24px 24px 24px;
    height: 100%;
  }
  .operate-area {
    display: flex;
    align-items: center;
    padding: 16px 0;
    .operate-btns {
      display: flex;
      align-items: center;
      :deep(.create-config-btn) {
        margin-right: 8px;
      }
      :deep(.batch-delete-btn) {
        margin-left: 8px;
      }
    }
    .config-search-input {
      margin-left: auto;
    }
  }
  .config-list-table {
    height: calc(100% - 64px);
  }
</style>
