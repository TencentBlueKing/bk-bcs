<template>
  <section class="config-list-wrapper">
    <div class="operate-area">
      <div class="operate-btns">
        <template v-if="versionData.status.publish_status === 'editing'">
          <CreateConfig
            :bk-biz-id="props.bkBizId"
            :app-id="props.appId"
            @created="refreshConfigList"
            @imported="refreshConfigList"
            @uploaded="refreshConfigList(true)"
          />
          <EditVariables v-if="isFileType" ref="editVariablesRef" :bk-biz-id="props.bkBizId" :app-id="props.appId" />
        </template>
        <ViewVariables
          v-else-if="isFileType"
          :bk-biz-id="props.bkBizId"
          :app-id="props.appId"
          :verision-id="versionData.id"
        />
      </div>
      <SearchInput
        v-model="searchStr"
        class="config-search-input"
        :placeholder="t('配置文件名/创建人/修改人')"
        :width="280"
      />
    </div>
    <section class="config-list-table">
      <template v-if="isFileType">
        <TableWithTemplates
          v-if="useTemplate"
          ref="tableRef"
          :bk-biz-id="props.bkBizId"
          :app-id="props.appId"
          :search-str="searchStr"
          @clear-str="clearStr"
          @delete-config="refreshVariable"
        />
        <TableWithPagination
          v-else
          ref="tableRef"
          :bk-biz-id="props.bkBizId"
          :app-id="props.appId"
          :search-str="searchStr"
        />
      </template>
      <TableWithKv
        v-else
        ref="tableRef"
        :bk-biz-id="props.bkBizId"
        :app-id="props.appId"
        :search-str="searchStr"
        @clear-str="clearStr"
      />
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
import TableWithPagination from './tables/table-with-pagination.vue';
import TableWithKv from './tables/table-with-kv.vue';

const configStore = useConfigStore();
const serviceStore = useServiceStore();
const { versionData } = storeToRefs(configStore);
const { isFileType } = storeToRefs(serviceStore);
const { t } = useI18n();

const props = defineProps<{
  bkBizId: string;
  appId: number;
}>();

const tableRef = ref();
const searchStr = ref('');
const useTemplate = ref(true);
const editVariablesRef = ref();

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

defineExpose({
  refreshConfigList,
});
</script>
<style lang="scss" scoped>
.config-list-wrapper {
  position: relative;
  padding: 0 24px;
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
  }
  .config-search-input {
    margin-left: auto;
  }
}
.config-list-table {
  max-height: calc(100% - 64px);
  overflow: auto;
}
</style>
