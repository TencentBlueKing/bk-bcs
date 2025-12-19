<template>
  <BaseLayout title="StatefulSets" kind="StatefulSet" category="statefulsets" type="workloads">
    <template
      #default="{
        curPageData, pageConf, statusMap, updateStrategyMap, handlePageChange, handlePageSizeChange,
        handleGetExtData, gotoDetail, handleSortChange,handleUpdateResource,handleDeleteResource,
        handleEnlargeCapacity, handleShowViewConfig, handleRestart, handleGotoUpdateRecord,
        handleRollback, clusterNameMap, goNamespace, isViewEditable, isClusterMode, sourceTypeMap,
        resolveLink
      }">
      <bk-table
        :data="curPageData"
        :pagination="pageConf"
        @page-change="handlePageChange"
        @page-limit-change="handlePageSizeChange"
        @sort-change="handleSortChange">
        <bk-table-column :label="$t('generic.label.name')" prop="metadata.name" sortable fixed="left">
          <template #default="{ row }">
            <bk-button
              class="bcs-button-ellipsis"
              :disabled="isViewEditable"
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
        <bk-table-column :label="$t('k8s.namespace')" prop="metadata.namespace" sortable>
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
        <bk-table-column :label="$t('k8s.updateStrategy.text')" min-width="100">
          <template slot-scope="{ row }">
            {{ updateStrategyMap[$chainable(row.spec, 'updateStrategy.type')]
              || $t('k8s.updateStrategy.rollingUpdate') }}
          </template>
        </bk-table-column>
        <bk-table-column :label="$t('k8s.statefulset.podManagementPolicy')">
          <template slot-scope="{ row }">
            {{ row.spec.podManagementPolicy || '--' }}
          </template>
        </bk-table-column>
        <bk-table-column :label="$t('generic.label.status')" min-width="100">
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
          <template slot-scope="{ row }">
            <span :class="{ 'text-[#E38B02]': (row.status.readyReplicas || 0) < (row.spec.replicas || 0) }">
              {{row.status.readyReplicas || 0}} / {{row.spec.replicas || 0}}
            </span>
          </template>
        </bk-table-column>
        <bk-table-column label="Up-to-date" width="110" :resizable="false">
          <template slot-scope="{ row }">{{row.status.updatedReplicas || 0}}</template>
        </bk-table-column>
        <bk-table-column label="Age" sortable="custom" prop="createTime" :resizable="false">
          <template #default="{ row }">
            <span>{{handleGetExtData(row.metadata.uid, 'age')}}</span>
          </template>
        </bk-table-column>
        <bk-table-column :label="$t('generic.label.createdBy')">
          <template slot-scope="{ row }">
            <bk-user-display-name :user-id="handleGetExtData(row.metadata.uid, 'creator')"></bk-user-display-name>
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
          width="250"
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
              class="ml10"
              text
              v-if="$chainable(row.spec, 'updateStrategy.type') === 'OnDelete'"
              @click="gotoDetail(row)">
              {{ $t('dashboard.workload.button.onDelete') }}
            </bk-button>
            <bk-button
              class="ml10"
              text
              v-else
              @click="handleRestart(row)">
              {{ $t('dashboard.workload.button.restart') }}
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
import BaseLayout from '@/views/resource-view/common/base-layout';

export default defineComponent({
  name: 'DashboardStateful',
  components: { BaseLayout, StatusIcon, LoadingIcon, sourceTableCell },
});
</script>
