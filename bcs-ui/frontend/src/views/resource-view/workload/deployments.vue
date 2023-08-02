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
        <bk-table-column :label="$t('generic.label.name')" prop="metadata.name" min-width="100" sortable="custom">
          <template #default="{ row }">
            <bk-button class="bcs-button-ellipsis" text @click="gotoDetail(row)">{{ row.metadata.name }}</bk-button>
          </template>
        </bk-table-column>
        <bk-table-column
          :label="$t('k8s.namespace')"
          prop="metadata.namespace"
          min-width="100"
          sortable="custom">
        </bk-table-column>
        <bk-table-column :label="$t('k8s.updateStrategy.text')" min-width="115">
          <template slot-scope="{ row }">
            <span>
              <bk-popover placement="top" v-if="$chainable(row.spec, 'strategy.type') === 'RollingUpdate'">
                <span class="border-bottom-tips">{{ $t('k8s.updateStrategy.rollingUpdate') }}</span>
                <div slot="content" v-if="$chainable(row.spec, 'strategy.rollingUpdate.maxSurge')">
                  <p>
                    {{ $t(
                      'dashboard.workload.label.upgrade.maxSurg',
                      { num: row.spec.strategy.rollingUpdate.maxSurge }
                    ) }}
                  </p>
                  <p>
                    {{ $t(
                      'dashboard.workload.label.upgrade.maxUnavailable',
                      { num: row.spec.strategy.rollingUpdate.maxUnavailable }
                    ) }}
                  </p>
                </div>
                <div slot="content" v-else>
                  <p>{{ $t('dashboard.workload.label.upgrade.maxSurg', { num: '--' }) }}</p>
                  <p>{{ $t('dashboard.workload.label.upgrade.maxUnavailable', { num: '--' }) }}</p>
                </div>
              </bk-popover>
              <span v-else>
                {{ updateStrategyMap[row.spec.strategy.type] }}
              </span>
            </span>
          </template>
        </bk-table-column>
        <bk-table-column
          :label="$t('generic.label.status')"
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
        <bk-table-column label="Age" sortable="custom" prop="createTime" width="100" :resizable="false">
          <template slot-scope="{ row }">
            <span>{{handleGetExtData(row.metadata.uid, 'age')}}</span>
          </template>
        </bk-table-column>
        <bk-table-column :label="$t('generic.label.createdBy')">
          <template slot-scope="{ row }">
            <span>{{handleGetExtData(row.metadata.uid, 'creator') || '--'}}</span>
          </template>
        </bk-table-column>
        <bk-table-column :label="$t('generic.label.editMode.text')" width="100">
          <template slot-scope="{ row }">
            <span>
              {{handleGetExtData(row.metadata.uid, 'editMode') === 'form'
                ? $t('generic.label.editMode.form') : 'YAML'}}
            </span>
          </template>
        </bk-table-column>
        <bk-table-column :label="$t('generic.label.action')" :resizable="false" width="250">
          <template #default="{ row }">
            <bk-button
              text
              @click="handleUpdateResource(row)">{{ $t('generic.button.update') }}</bk-button>
            <bk-button
              class="ml10" text
              @click="handleEnlargeCapacity(row)">{{ $t('deploy.templateset.scale') }}</bk-button>
            <bk-button
              class="ml10" text
              @click="gotoDetail(row)">{{ $t('dashboard.workload.pods.delete') }}</bk-button>
            <bk-button
              class="ml10" text
              @click="handleDeleteResource(row)">{{ $t('generic.button.delete') }}</bk-button>
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
  name: 'DashboardDeploy',
  components: { BaseLayout, StatusIcon, LoadingIcon },
});
</script>
<style lang="postcss" scoped>
>>> .border-bottom-tips {
  display: inline-block;
  line-height: 18px;
  border-bottom: 1px dashed #979ba5;
}
</style>
