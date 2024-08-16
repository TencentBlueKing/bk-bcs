<template>
  <BaseLayout
    title="GameStatefulSets"
    kind="GameStatefulSet"
    type="crd"
    category="custom_objects"
    :crd="crd"
    default-active-detail-type="yaml"
    :show-detail-tab="false"
    scope="Namespaced">
    <template
      #default="{
        curPageData, pageConf,
        handlePageChange, handlePageSizeChange,
        handleGetExtData, handleUpdateResource,
        handleDeleteResource, handleSortChange,
        gotoDetail, renderCrdHeader,
        getJsonPathValue, additionalColumns,
        webAnnotations, updateStrategyMap, statusMap,
        handleEnlargeCapacity, handleShowViewConfig,
        handleRestart, handleGotoUpdateRecord, handleRollback,
        clusterNameMap, goNamespace, isViewEditable,isClusterMode
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
              text
              :disabled="isViewEditable"
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
        <bk-table-column :label="$t('k8s.namespace')" prop="metadata.namespace" min-width="100" sortable>
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
        <bk-table-column :label="$t('k8s.updateStrategy.text')" width="150" :resizable="false">
          <template slot-scope="{ row }">
            <span v-if="row.spec.updateStrategy">
              {{ updateStrategyMap[row.spec.updateStrategy.type] || $t('k8s.updateStrategy.rollingUpdate') }}
            </span>
            <span v-else>{{ $t('k8s.updateStrategy.rollingUpdate') }}</span>
          </template>
        </bk-table-column>
        <bk-table-column :label="$t('k8s.statefulset.podManagementPolicy')" min-width="120">
          <template slot-scope="{ row }">
            {{ row.spec.podManagementPolicy || '--' }}
          </template>
        </bk-table-column>
        <bk-table-column :label="$t('generic.label.status')" min-width="100">
          <template slot-scope="{ row }">
            <StatusIcon status="running" v-if="handleGetExtData(row.metadata.uid, 'resStatus') === 'normal'">
              {{statusMap[handleGetExtData(row.metadata.uid, 'resStatus')] || '--'}}
            </StatusIcon>
            <LoadingIcon v-else>
              <span class="bcs-ellipsis">{{ statusMap[handleGetExtData(row.metadata.uid, 'resStatus')] || '--' }}</span>
            </LoadingIcon>
          </template>
        </bk-table-column>
        <bk-table-column
          v-for="item in additionalColumns"
          :key="item.name"
          :label="item.name"
          :prop="item.jsonPath"
          :render-header="renderCrdHeader">
          <template #default="{ row }">
            <span>
              {{ getJsonPathValue(row, item.jsonPath) || '--' }}
            </span>
          </template>
        </bk-table-column>
        <bk-table-column label="Age" sortable="custom" prop="createTime" :show-overflow-tooltip="false">
          <template #default="{ row }">
            <span v-bk-tooltips="{ content: handleGetExtData(row.metadata.uid, 'createTime') }">
              {{ handleGetExtData(row.metadata.uid, 'age') }}</span>
          </template>
        </bk-table-column>
        <bk-table-column :label="$t('generic.label.createdBy')">
          <template slot-scope="{ row }">
            <span>{{handleGetExtData(row.metadata.uid, 'creator') || '--'}}</span>
          </template>
        </bk-table-column>
        <bk-table-column :label="$t('generic.label.source')" show-overflow-tooltip>
          <template #default="{ row }">
            <span v-if="handleGetExtData(row.metadata.uid, 'createSource') === 'Template'">
              {{ `${handleGetExtData(row.metadata.uid, 'templateName') || '--'}:${
                handleGetExtData(row.metadata.uid, 'templateVersion') || '--'}` }}
            </span>
            <span v-else-if="handleGetExtData(row.metadata.uid, 'createSource') === 'Helm'">
              {{ handleGetExtData(row.metadata.uid, 'chart')
                ?`${handleGetExtData(row.metadata.uid, 'chart') || '--'}`
                : 'Helm' }}
            </span>
            <span v-else>{{ handleGetExtData(row.metadata.uid, 'createSource') }}</span>
          </template>
        </bk-table-column>
        <bk-table-column :label="$t('generic.label.editMode.text')" width="100">
          <template #default="{ row }">
            <span>
              {{handleGetExtData(row.metadata.uid, 'editMode') === 'form'
                ? $t('generic.label.editMode.form') : 'YAML'}}
            </span>
          </template>
        </bk-table-column>
        <bk-table-column
          :label="$t('generic.label.action')"
          :resizable="false"
          width="240"
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
              :disabled="row.spec.updateStrategy.type === 'InplaceUpdate'"
              @click="handleRestart(row)">
              <span
                v-bk-tooltips="{
                  content: $t('dashboard.workload.tips.inplaceUpdateNotSupportRestart'),
                  disabled: row.spec.updateStrategy.type !== 'InplaceUpdate'
                }">
                {{ $t('dashboard.workload.button.restart') }}
              </span>
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
                  <li
                    class="bcs-dropdown-item"
                    v-authority="{
                      clickable: webAnnotations.perms.items[row.metadata.uid]
                        ? webAnnotations.perms.items[row.metadata.uid].deleteBtn.clickable : true,
                      content: webAnnotations.perms.items[row.metadata.uid]
                        ? webAnnotations.perms.items[row.metadata.uid].deleteBtn.tip : '',
                      disablePerms: true
                    }"
                    @click="handleDeleteResource(row)">
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

import LoadingIcon from '@/components/loading-icon.vue';
import BaseLayout from '@/views/resource-view/common/base-layout';

export default defineComponent({
  name: 'GameStatefulSets',
  components: { BaseLayout, StatusIcon, LoadingIcon },
  props: {
    crd: {
      type: String,
      default: 'gamestatefulsets.tkex.tencent.com',
    },
  },
});
</script>
