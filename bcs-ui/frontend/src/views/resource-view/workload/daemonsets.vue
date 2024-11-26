<template>
  <BaseLayout title="DaemonSets" kind="DaemonSet" category="daemonsets" type="workloads">
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
        handleShowViewConfig,
        handleRestart,
        handleGotoUpdateRecord,
        handleRollback,
        updateStrategyMap,
        clusterNameMap,
        goNamespace,
        isViewEditable,
        isClusterMode,
        sourceTypeMap
      }">
      <bk-table
        :data="curPageData"
        :pagination="pageConf"
        @page-change="handlePageChange"
        @page-limit-change="handlePageSizeChange"
        @sort-change="handleSortChange">
        <bk-table-column :label="$t('generic.label.name')" prop="metadata.name" min-width="100" sortable fixed="left">
          <template #default="{ row }">
            <bk-button
              class="bcs-button-ellipsis"
              :disabled="isViewEditable"
              text
              @click="gotoDetail(row)">{{ row.metadata.name }}</bk-button>
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
          sortable>
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
        <bk-table-column :label="$t('k8s.image')" min-width="280" :show-overflow-tooltip="false">
          <template slot-scope="{ row }">
            <span v-bk-tooltips.top="(handleGetExtData(row.metadata.uid, 'images') || []).join('<br />')">
              {{ (handleGetExtData(row.metadata.uid, 'images') || []).join(', ') }}
            </span>
          </template>
        </bk-table-column>
        <bk-table-column label="Desired" width="110" :resizable="false">
          <template slot-scope="{ row }">{{row.status.desiredNumberScheduled || 0}}</template>
        </bk-table-column>
        <bk-table-column label="Current" width="110" :resizable="false">
          <template slot-scope="{ row }">{{row.status.currentNumberScheduled || 0}}</template>
        </bk-table-column>
        <bk-table-column label="Ready" width="110" :resizable="false">
          <template slot-scope="{ row }">{{row.status.numberReady || 0}}</template>
        </bk-table-column>
        <bk-table-column label="Up-to-date" width="110" :resizable="false">
          <template slot-scope="{ row }">{{row.status.updatedNumberScheduled || 0}}</template>
        </bk-table-column>
        <bk-table-column label="Available" width="110" :resizable="false">
          <template slot-scope="{ row }">{{row.status.numberAvailable || 0}}</template>
        </bk-table-column>
        <bk-table-column label="Age" sortable="custom" prop="createTime" :resizable="false">
          <template #default="{ row }">
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
            <div class="flex items-center">
              <bk-popover
                class="size-[16px] mr-[4px]"
                :content="sourceTypeMap?.[handleGetExtData(row.metadata.uid, 'createSource')]?.iconText"
                :tippy-options="{ interactive: false }">
                <i
                  class="text-[14px] p-[1px]"
                  :class="sourceTypeMap?.[handleGetExtData(row.metadata.uid, 'createSource')]?.iconClass"></i>
              </bk-popover>
              <span
                v-bk-overflow-tips="{ interactive: false }"
                class="bcs-ellipsis" v-if="handleGetExtData(row.metadata.uid, 'createSource') === 'Template'">
                {{ `${handleGetExtData(row.metadata.uid, 'templateName') || '--'}:${
                  handleGetExtData(row.metadata.uid, 'templateVersion') || '--'}` }}
              </span>
              <span
                v-bk-overflow-tips="{ interactive: false }" class="bcs-ellipsis"
                v-else-if="handleGetExtData(row.metadata.uid, 'createSource') === 'Helm'">
                {{ handleGetExtData(row.metadata.uid, 'chart')
                  ?`${handleGetExtData(row.metadata.uid, 'chart') || '--'}`
                  : 'Helm' }}
              </span>
              <span
                v-bk-overflow-tips="{ interactive: false }" class="bcs-ellipsis"
                v-else>{{ handleGetExtData(row.metadata.uid, 'createSource') }}</span>
            </div>
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
          width="200"
          fixed="right"
          v-if="!isViewEditable">
          <template #default="{ row }">
            <bk-button
              text
              @click="handleUpdateResource(row)">{{ $t('generic.button.update') }}</bk-button>
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

import BaseLayout from '@/views/resource-view/common/base-layout';

export default defineComponent({
  name: 'WorkloadDaemonsets',
  components: { BaseLayout },
});
</script>
