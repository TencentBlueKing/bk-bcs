<template>
  <div class="biz-content">
    <BaseLayout title="Pods" kind="Pod" category="pods" type="workloads">
      <template
        #default="{
          curPageData,
          pageConf,
          handlePageChange,
          handlePageSizeChange,
          handleGetExtData,
          gotoDetail,
          handleSortChange,
          handleUpdateResource,
          handleDeleteResource,
          nameValue, handleClearSearchData
        }">
        <bk-table
          :data="curPageData"
          :pagination="pageConf"
          @page-change="handlePageChange"
          @page-limit-change="handlePageSizeChange"
          @sort-change="handleSortChange">
          <bk-table-column :label="$t('名称')" min-width="130" prop="metadata.name" sortable fixed="left">
            <template #default="{ row }">
              <bk-button class="bcs-button-ellipsis" text @click="gotoDetail(row)">{{ row.metadata.name }}</bk-button>
            </template>
          </bk-table-column>
          <bk-table-column :label="$t('命名空间')" width="150" prop="metadata.namespace" sortable></bk-table-column>
          <bk-table-column :label="$t('镜像')" min-width="200" :show-overflow-tooltip="false">
            <template #default="{ row }">
              <span v-bk-tooltips.top="(handleGetExtData(row.metadata.uid, 'images') || []).join('<br />')">
                {{ (handleGetExtData(row.metadata.uid, 'images') || []).join(', ') }}
              </span>
            </template>
          </bk-table-column>
          <bk-table-column label="Status" width="120" :resizable="false">
            <template #default="{ row }">
              <StatusIcon :status="handleGetExtData(row.metadata.uid, 'status')"></StatusIcon>
            </template>
          </bk-table-column>
          <bk-table-column label="Ready" width="100" :resizable="false">
            <template #default="{ row }">
              {{handleGetExtData(row.metadata.uid, 'readyCnt')}}/{{handleGetExtData(row.metadata.uid, 'totalCnt')}}
            </template>
          </bk-table-column>
          <bk-table-column label="Restarts" width="100" :resizable="false">
            <template #default="{ row }">{{handleGetExtData(row.metadata.uid, 'restartCnt')}}</template>
          </bk-table-column>
          <bk-table-column label="Host IP" width="140">
            <template #default="{ row }">{{row.status.hostIP || '--'}}</template>
          </bk-table-column>
          <bk-table-column label="Pod IPv4" width="140">
            <template #default="{ row }">{{handleGetExtData(row.metadata.uid, 'podIPv4') || '--'}}</template>
          </bk-table-column>
          <bk-table-column label="Pod IPv6" min-width="200">
            <template #default="{ row }">{{handleGetExtData(row.metadata.uid, 'podIPv6') || '--'}}</template>
          </bk-table-column>
          <bk-table-column label="Node" :resizable="false">
            <template #default="{ row }">{{row.spec.nodeName || '--'}}</template>
          </bk-table-column>
          <bk-table-column label="Age" sortable="custom" prop="createTime" :resizable="false">
            <template #default="{ row }">
              <span>{{handleGetExtData(row.metadata.uid, 'age')}}</span>
            </template>
          </bk-table-column>
          <bk-table-column :label="$t('编辑模式')" width="100">
            <template #default="{ row }">
              <span>
                {{handleGetExtData(row.metadata.uid, 'editMode') === 'form'
                  ? $t('表单') : 'YAML'}}
              </span>
            </template>
          </bk-table-column>
          <bk-table-column :label="$t('操作')" :resizable="false" width="180" fixed="right">
            <template #default="{ row }">
              <bk-button text @click="handleShowLog(row)">{{ $t('日志') }}</bk-button>
              <bk-button
                text class="ml10"
                @click="handleUpdateResource(row)">{{ $t('更新') }}</bk-button>
              <bk-button
                class="ml10" text
                @click="handleDeleteResource(row)">{{ $t('删除') }}</bk-button>
            </template>
          </bk-table-column>
          <template #empty>
            <BcsEmptyTableStatus :type="nameValue ? 'search-empty' : 'empty'" @clear="handleClearSearchData" />
          </template>
        </bk-table>
      </template>
    </BaseLayout>
    <BcsLog
      v-model="showLog"
      :cluster-id="clusterId"
      :namespace="currentRow.metadata.namespace"
      :name="currentRow.metadata.name">
    </BcsLog>
  </div>
</template>
<script lang="ts">
import { defineComponent, computed, ref } from 'vue';
import BaseLayout from '@/views/resource-view/common/base-layout';
import StatusIcon from '@/components/status-icon';
import BcsLog from '@/components/bcs-log/log-dialog.vue';
import $store from '@/store';

export default defineComponent({
  name: 'WorkloadPods',
  components: { BaseLayout, StatusIcon, BcsLog },
  setup() {
    const clusterId = computed(() => $store.getters.curClusterId);

    // 显示日志
    const showLog = ref(false);
    const currentRow = ref<Record<string, any>>({ metadata: {} });
    const handleShowLog = (row) => {
      currentRow.value = row;
      showLog.value = true;
    };

    return {
      clusterId,
      showLog,
      currentRow,
      handleShowLog,
    };
  },
});
</script>
<style lang="postcss" scoped>
@import './detail/pod-log.css';
/deep/ .base-layout {
    width: 100%;
}
</style>
