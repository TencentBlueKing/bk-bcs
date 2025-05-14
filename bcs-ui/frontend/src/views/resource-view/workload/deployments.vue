<template>
  <BaseLayout title="Deployments" kind="Deployment" category="deployments" type="workloads">
    <template
      #default="{
        curPageData, pageConf, statusMap, updateStrategyMap, handlePageChange, handlePageSizeChange,
        handleGetExtData, handleSortChange, gotoDetail, handleUpdateResource, handleDeleteResource,
        handleEnlargeCapacity, statusFilters, handleShowViewConfig, handleFilterChange,
        handleGotoUpdateRecord, handleRestart, handleRollback, clusterNameMap, goNamespace, isViewEditable,
        isClusterMode, sourceTypeMap, resolveLink,
      }">
      <bk-table
        :data="curPageData"
        :pagination="pageConf"
        @page-change="handlePageChange"
        @page-limit-change="handlePageSizeChange"
        @sort-change="handleSortChange"
        @filter-change="handleFilterChange">
        <bk-table-column
          :label="$t('generic.label.name')"
          prop="metadata.name"
          min-width="100"
          sortable="custom"
          fixed="left">
          <template #default="{ row }">
            <bk-button
              :disabled="isViewEditable"
              class="bcs-button-ellipsis"
              text>
              <a :href="resolveLink(row)" @click.prevent="gotoDetail($event, resolveLink(row), row)">
                {{ row.metadata.name }}
              </a>
            </bk-button>
          </template>
        </bk-table-column>
        <bk-table-column :label="$t('cluster.labels.nameAndId')" v-if="!isClusterMode">
          <template #default="{ row }">
            <div class="flex flex-col py-[6px] h-[50px]">
              <span class="bcs-ellipsis">{{ clusterNameMap[handleGetExtData(row.metadata.uid, 'clusterID')] }}</span>
              <span class="bcs-ellipsis mt-[6px]">{{ handleGetExtData(row.metadata.uid, 'clusterID') }}</span>
            </div>
          </template>
        </bk-table-column>
        <bk-table-column
          :label="$t('k8s.namespace')"
          prop="metadata.namespace"
          min-width="100"
          sortable="custom">
          <template #default="{ row }">
            <bk-button
              class="bcs-button-ellipsis"
              text
              :disabled="isViewEditable"
              @click="goNamespace(row)">
              {{ row.metadata.namespace }}
            </bk-button>
          </template>
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
          column-key="status"
          :filters="statusFilters"
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
        <bk-table-column label="Ready" width="100" :render-header="renderReadyHeader" :resizable="false">
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
        <bk-table-column :label="$t('generic.label.source')" :show-overflow-tooltip="false">
          <template #default="{ row }">
            <sourceTableCell
              :row="row"
              :source-type-map="sourceTypeMap" />
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
        <bk-table-column
          :label="$t('generic.label.action')"
          :resizable="false"
          width="230"
          fixed="right"
          v-if="!isViewEditable">
          <template #default="{ row }">
            <bk-button
              text
              @click="handleUpdateResource(row)">{{ $t('generic.button.update') }}</bk-button>
            <bk-button
              class="ml10" text
              @click="handleEnlargeCapacity(row)">{{ $t('deploy.templateset.scale') }}</bk-button>
            <bk-button
              class="ml10" text
              @click="handleRestart(
                row,
                row.spec.strategy.type === 'Recreate'
                  ? updateStrategyMap[row.spec.strategy.type]
                  : $t('dashboard.workload.button.restart')
              )">
              {{ row.spec.strategy.type === 'Recreate'
                ? $t('k8s.updateStrategy.reCreate')
                : $t('dashboard.workload.button.restart') }}
            </bk-button>
            <bk-popover
              placement="bottom"
              theme="light dropdown"
              :arrow="false"
              trigger="click"
              class="ml-[10px]">
              <span :class="['bcs-icon-more-btn', { disabled: false }]">
                <i class="bcs-icon bcs-icon-more"></i>
              </span>
              <template #content>
                <ul class="bcs-dropdown-list">
                  <li class="bcs-dropdown-item" @click="handleGotoUpdateRecord(row)">
                    {{$t('deploy.helm.history')}}
                  </li>
                  <li class="bcs-dropdown-item" @click="handleRollback(row)">
                    {{$t('deploy.helm.rollback')}}
                  </li>
                  <li class="bcs-dropdown-item" @click="handleDeleteResource(row)">
                    {{$t('generic.button.delete')}}
                  </li>
                </ul>
              </template>
            </bk-popover>
          </template>
        </bk-table-column>
        <template #empty>
          <BcsEmptyTableStatus
            :button-text="$t('generic.button.resetSearch')"
            type="search-empty"
            @clear="handleShowViewConfig" />
        </template>
      </bk-table>
    </template>
  </BaseLayout>
</template>
<script>
import { defineComponent } from 'vue';

import StatusIcon from '../../../components/status-icon';
import sourceTableCell from '../common/source-table-cell.vue';

import LoadingIcon from '@/components/loading-icon.vue';
import $i18n from '@/i18n/i18n-setup';
import BaseLayout from '@/views/resource-view/common/base-layout';

export default defineComponent({
  name: 'DashboardDeploy',
  components: { BaseLayout, StatusIcon, LoadingIcon, sourceTableCell },
  setup() {
    const renderReadyHeader = (h, { column }) => h(
      'span',
      {
        class: 'bcs-border-tips',
        directives: [
          {
            name: 'bkTooltips',
            value: {
              content: $i18n.t('k8s.deploy.readyTips'),
            },
          },
        ],
      },
      column.label,
    );
    return {
      renderReadyHeader,
    };
  },
});
</script>
<style lang="postcss" scoped>
>>> .border-bottom-tips {
  display: inline-block;
  line-height: 18px;
  border-bottom: 1px dashed #979ba5;
}
</style>
