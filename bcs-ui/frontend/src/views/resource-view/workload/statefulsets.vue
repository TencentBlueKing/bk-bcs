<template>
  <BaseLayout title="StatefulSets" kind="StatefulSet" category="statefulsets" type="workloads">
    <template
      #default="{ curPageData, pageConf, statusMap, updateStrategyMap, handlePageChange, handlePageSizeChange,
                  handleGetExtData, gotoDetail, handleSortChange,handleUpdateResource,handleDeleteResource,
                  handleEnlargeCapacity,nameValue, handleClearSearchData }">
      <bk-table
        :data="curPageData"
        :pagination="pageConf"
        @page-change="handlePageChange"
        @page-limit-change="handlePageSizeChange"
        @sort-change="handleSortChange">
        <bk-table-column :label="$t('名称')" prop="metadata.name" sortable>
          <template #default="{ row }">
            <bk-button class="bcs-button-ellipsis" text @click="gotoDetail(row)">{{ row.metadata.name }}</bk-button>
          </template>
        </bk-table-column>
        <bk-table-column :label="$t('命名空间')" prop="metadata.namespace" sortable></bk-table-column>
        <bk-table-column :label="$t('升级策略')" min-width="100">
          <template slot-scope="{ row }">
            {{ updateStrategyMap[$chainable(row.spec, 'updateStrategy.type')] || $t('滚动升级') }}
          </template>
        </bk-table-column>
        <bk-table-column :label="$t('Pod 管理策略')">
          <template slot-scope="{ row }">
            {{ row.spec.podManagementPolicy || '--' }}
          </template>
        </bk-table-column>
        <bk-table-column :label="$t('状态')" min-width="60">
          <template slot-scope="{ row }">
            <StatusIcon status="running" v-if="handleGetExtData(row.metadata.uid, 'status') === 'normal'">
              {{statusMap[handleGetExtData(row.metadata.uid, 'status')] || '--'}}
            </StatusIcon>
            <LoadingIcon v-else>
              <span class="bcs-ellipsis">{{ statusMap[handleGetExtData(row.metadata.uid, 'status')] || '--' }}</span>
            </LoadingIcon>
          </template>
        </bk-table-column>
        <bk-table-column label="Ready" width="110" :resizable="false">
          <template slot-scope="{ row }">{{row.status.readyReplicas || 0}} / {{row.spec.replicas || 0}}</template>
        </bk-table-column>
        <bk-table-column label="Up-to-date" width="110" :resizable="false">
          <template slot-scope="{ row }">{{row.status.updatedReplicas || 0}}</template>
        </bk-table-column>
        <bk-table-column label="Age" sortable="custom" prop="createTime" :resizable="false">
          <template #default="{ row }">
            <span>{{handleGetExtData(row.metadata.uid, 'age')}}</span>
          </template>
        </bk-table-column>
        <bk-table-column :label="$t('创建人')">
          <template slot-scope="{ row }">
            <span>{{handleGetExtData(row.metadata.uid, 'creator') || '--'}}</span>
          </template>
        </bk-table-column>
        <bk-table-column :label="$t('编辑模式')" width="100">
          <template slot-scope="{ row }">
            <span>
              {{handleGetExtData(row.metadata.uid, 'editMode') === 'form'
                ? $t('表单') : 'YAML'}}
            </span>
          </template>
        </bk-table-column>
        <bk-table-column :label="$t('操作')" :resizable="false" width="240">
          <template #default="{ row }">
            <bk-button
              text
              @click="handleUpdateResource(row)">{{ $t('更新') }}</bk-button>
            <bk-button
              class="ml10" text
              @click="handleEnlargeCapacity(row)">{{ $t('扩缩容') }}</bk-button>
            <bk-button
              class="ml10" text
              @click="gotoDetail(row)">{{ $t('重新调度') }}</bk-button>
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
</template>
<script>
import { defineComponent } from 'vue';
import BaseLayout from '@/views/resource-view/common/base-layout';
import StatusIcon from '../../../components/status-icon';
import LoadingIcon from '@/components/loading-icon.vue';

export default defineComponent({
  name: 'DashboardStateful',
  components: { BaseLayout, StatusIcon, LoadingIcon },
});
</script>
