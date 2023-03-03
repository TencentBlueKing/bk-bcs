<template>
  <BaseLayout title="Deployments" kind="Deployment" category="deployments" type="workloads">
    <template
      #default="{
        curPageData, pageConf, statusMap, updateStrategyMap, handlePageChange, handlePageSizeChange,
        handleGetExtData, handleSortChange, gotoDetail, handleUpdateResource, handleDeleteResource,
        handleEnlargeCapacity, statusFilters, statusFilterMethod, nameValue, handleClearSearchData
      }">
      <bk-table
        :data="curPageData"
        :pagination="pageConf"
        @page-change="handlePageChange"
        @page-limit-change="handlePageSizeChange"
        @sort-change="handleSortChange">
        <bk-table-column :label="$t('名称')" prop="metadata.name" min-width="100" sortable>
          <template #default="{ row }">
            <bk-button class="bcs-button-ellipsis" text @click="gotoDetail(row)">{{ row.metadata.name }}</bk-button>
          </template>
        </bk-table-column>
        <bk-table-column :label="$t('命名空间')" prop="metadata.namespace" min-width="100" sortable></bk-table-column>
        <bk-table-column :label="$t('升级策略')" min-width="100">
          <template slot-scope="{ row }">
            <span>
              <bk-popover placement="top" v-if="$chainable(row.spec, 'strategy.type') === 'RollingUpdate'">
                <span>{{ $t('滚动升级') }}</span>
                <div slot="content" v-if="$chainable(row.spec, 'strategy.rollingUpdate.maxSurge')">
                  <p>
                    {{ $t(`最大调度Pod数量（maxSurge）: ${String(row.spec.strategy.rollingUpdate.maxSurge).split('%')[0]}%`) }}
                  </p>
                  <p>
                    {{ $t(`最大不可用数量（maxUnavailable）: ${String(row.spec.strategy.rollingUpdate.maxUnavailable)
                      .split('%')[0]}%`) }}
                  </p>
                </div>
                <div slot="content" v-else>
                  <p>{{ $t('最大调度Pod数量（maxSurge）: --') }}</p>
                  <p>{{ $t('最大不可用数量（maxUnavailable）: --') }}</p>
                </div>
              </bk-popover>
              <span v-else>
                {{ updateStrategyMap[row.spec.strategy.type] }}
              </span>
            </span>
          </template>
        </bk-table-column>
        <bk-table-column
          :label="$t('状态')"
          prop="status"
          :filters="statusFilters"
          :filter-method="statusFilterMethod"
          filter-multiple
          min-width="100">
          <template slot-scope="{ row }">
            <StatusIcon status="running" v-if="handleGetExtData(row.metadata.uid, 'status') === 'normal'">
              {{statusMap[handleGetExtData(row.metadata.uid, 'status')] || '--'}}
            </StatusIcon>
            <LoadingIcon v-else>
              <span class="bcs-ellipsis">{{ statusMap[handleGetExtData(row.metadata.uid, 'status')] || '--' }}</span>
            </LoadingIcon>
          </template>
        </bk-table-column>
        <bk-table-column label="Ready" width="100" :resizable="false">
          <template slot-scope="{ row }">{{row.status.readyReplicas || 0}} / {{row.spec.replicas}}</template>
        </bk-table-column>
        <bk-table-column label="Up-to-date" width="110" :resizable="false">
          <template slot-scope="{ row }">{{row.status.updatedReplicas || 0}}</template>
        </bk-table-column>
        <bk-table-column label="Available" width="100" :resizable="false">
          <template slot-scope="{ row }">{{row.status.availableReplicas || 0}}</template>
        </bk-table-column>
        <bk-table-column label="Age" width="100" :resizable="false">
          <template slot-scope="{ row }">
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
import { defineComponent } from '@vue/composition-api';
import BaseLayout from '@/views/dashboard/common/base-layout';
import StatusIcon from '../common/status-icon';
import LoadingIcon from '@/components/loading-icon.vue';

export default defineComponent({
  name: 'DashboardDeploy',
  components: { BaseLayout, StatusIcon, LoadingIcon },
});
</script>
