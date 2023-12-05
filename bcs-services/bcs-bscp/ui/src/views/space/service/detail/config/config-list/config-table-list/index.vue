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
            @uploaded="refreshConfigList"
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
      <div class="groups-info" v-if="versionData.status.released_groups.length > 0">
        <div v-for="group in versionData.status.released_groups" class="group-item" :key="group.id">
          {{ group.name }}
        </div>
      </div>
      <SearchInput
        v-model="searchStr"
        class="config-search-input"
        placeholder="配置文件名/创建人/修改人"
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
        @delete-config="refreshVariable"
      />
    </section>
  </section>
</template>
<script setup lang="ts">
import { ref } from 'vue';
import { storeToRefs } from 'pinia';
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

const props = defineProps<{
  bkBizId: string;
  appId: number;
}>();

const tableRef = ref();
const searchStr = ref('');
const useTemplate = ref(true);
const editVariablesRef = ref();

const refreshConfigList = () => {
  tableRef.value.refresh();
  refreshVariable();
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
  .groups-info {
    display: flex;
    align-items: center;
    flex-wrap: wrap;
    margin-left: 24px;
    .group-item {
      padding: 0 8px;
      line-height: 22px;
      color: #63656e;
      font-size: 12px;
      background: #f0f1f5;
      border-radius: 2px;
      margin-bottom: 2px;
      &:not(:last-of-type) {
        margin-right: 8px;
      }
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
