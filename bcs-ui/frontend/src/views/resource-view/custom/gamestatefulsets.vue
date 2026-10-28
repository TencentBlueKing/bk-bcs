<template>
  <BaseLayout
    title="GameStatefulSets"
    kind="GameStatefulSet"
    type="crd"
    category="custom_objects"
    :crd="`${resource}.${group}`"
    default-active-detail-type="yaml"
    :show-detail-tab="false"
    :crd-options="{
      group: group,
      version: version,
      resource: resource,
      namespaced: namespaced,
    }"
    :table-setting-fields="getTableSettingFields"
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
        clusterNameMap, goNamespace, isViewEditable,isClusterMode,
        sourceTypeMap, resolveLink,
        tableSetting, isColumnRender, handleSettingChange
      }">
      <bk-table
        :data="curPageData"
        :size="tableSetting.size"
        :pagination="pageConf"
        @page-change="handlePageChange"
        @page-limit-change="handlePageSizeChange"
        @sort-change="handleSortChange">
        <bk-table-column
          :label="$t('generic.label.name')"
          prop="metadata.name"
          sortable
          fixed="left"
          key="name">
          <template #default="{ row }">
            <bk-button
              class="bcs-button-ellipsis"
              text
              :disabled="isViewEditable">
              <a :href="resolveLink(row)" @click.prevent="gotoDetail($event, resolveLink(row), row)">
                {{ row.metadata.name }}
              </a>
            </bk-button>
          </template>
        </bk-table-column>
        <bk-table-column
          :label="$t('cluster.labels.nameAndId')"
          key="cluster"
          v-if="!isClusterMode && isColumnRender('cluster')">
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
          sortable
          key="namespace"
          v-if="isColumnRender('namespace')">
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
        <bk-table-column
          :label="$t('k8s.updateStrategy.text')"
          width="150"
          :resizable="false"
          key="updateStrategy"
          v-if="isColumnRender('updateStrategy')">
          <template slot-scope="{ row }">
            <span v-if="row.spec.updateStrategy">
              {{ updateStrategyMap[row.spec.updateStrategy.type] || $t('k8s.updateStrategy.rollingUpdate') }}
            </span>
            <span v-else>{{ $t('k8s.updateStrategy.rollingUpdate') }}</span>
          </template>
        </bk-table-column>
        <bk-table-column
          :label="$t('k8s.statefulset.podManagementPolicy')"
          min-width="120"
          key="podManagementPolicy"
          v-if="isColumnRender('podManagementPolicy')">
          <template slot-scope="{ row }">
            {{ row.spec.podManagementPolicy || '--' }}
          </template>
        </bk-table-column>
        <bk-table-column
          :label="$t('generic.label.status')"
          min-width="100"
          key="status"
          v-if="isColumnRender('status')">
          <template slot-scope="{ row }">
            <StatusIcon status="running" v-if="handleGetExtData(row.metadata.uid, 'resStatus') === 'normal'">
              {{statusMap[handleGetExtData(row.metadata.uid, 'resStatus')] || '--'}}
            </StatusIcon>
            <LoadingIcon v-else>
              <span class="bcs-ellipsis">{{ statusMap[handleGetExtData(row.metadata.uid, 'resStatus')] || '--' }}</span>
            </LoadingIcon>
          </template>
        </bk-table-column>
        <template v-for="item in additionalColumns">
          <bk-table-column
            v-if="isColumnRender(getAdditionalColumnId(item))"
            :key="getAdditionalColumnId(item)"
            :label="item.name"
            :prop="item.jsonPath"
            :render-header="renderCrdHeader">
            <template #default="{ row }">
              <span>
                {{ getJsonPathValue(row, item.jsonPath) || '--' }}
              </span>
            </template>
          </bk-table-column>
        </template>
        <bk-table-column
          label="Age"
          sortable="custom"
          prop="createTime"
          :show-overflow-tooltip="false"
          key="age"
          v-if="isColumnRender('age')">
          <template #default="{ row }">
            <span v-bk-tooltips="{ content: handleGetExtData(row.metadata.uid, 'createTime') }">
              {{ handleGetExtData(row.metadata.uid, 'age') }}</span>
          </template>
        </bk-table-column>
        <bk-table-column
          :label="$t('generic.label.createdBy')"
          key="creator"
          v-if="isColumnRender('creator')">
          <template slot-scope="{ row }">
            <bk-user-display-name :user-id="handleGetExtData(row.metadata.uid, 'creator')"></bk-user-display-name>
          </template>
        </bk-table-column>
        <bk-table-column
          :label="$t('generic.label.source')"
          :show-overflow-tooltip="false"
          key="source"
          v-if="isColumnRender('source')">
          <template #default="{ row }">
            <sourceTableCell
              :row="row"
              :source-type-map="sourceTypeMap" />
          </template>
        </bk-table-column>
        <bk-table-column
          :label="$t('generic.label.editMode.text')"
          width="100"
          key="editMode"
          v-if="isColumnRender('editMode')">
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
          key="action"
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
              :disabled="$chainable(row.spec, 'updateStrategy.type') === 'InplaceUpdate'"
              @click="handleRestart(row)">
              <span
                v-bk-tooltips="{
                  content: $t('dashboard.workload.tips.inplaceUpdateNotSupportRestart'),
                  disabled: $chainable(row.spec, 'updateStrategy.type') !== 'InplaceUpdate'
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
        <bk-table-column type="setting" :resizable="false" key="setting">
          <bcs-table-setting-content
            :fields="tableSetting.fields"
            :selected="tableSetting.selectedFields"
            :size="tableSetting.size"
            @setting-change="handleSettingChange">
          </bcs-table-setting-content>
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
  name: 'GameStatefulSets',
  components: { BaseLayout, StatusIcon, LoadingIcon, sourceTableCell },
  props: {
    group: {
      type: String,
      default: 'tkex.tencent.com',
    },
    version: {
      type: String,
      default: 'v1alpha1',
    },
    resource: {
      type: String,
      default: 'gamestatefulsets',
    },
    namespaced: {
      type: Boolean,
      default: true,
    },
  },
  setup() {
    // CRD 动态列ID
    const getAdditionalColumnId = item => `additional_${item.name}`;
    // 列显隐设置字段
    const getTableSettingFields = ({ additionalColumns = [], isClusterMode }) => [
      { id: 'name', label: $i18n.t('generic.label.name'), disabled: true },
      ...(!isClusterMode ? [{ id: 'cluster', label: $i18n.t('cluster.labels.nameAndId') }] : []),
      { id: 'namespace', label: $i18n.t('k8s.namespace') },
      { id: 'updateStrategy', label: $i18n.t('k8s.updateStrategy.text') },
      { id: 'podManagementPolicy', label: $i18n.t('k8s.statefulset.podManagementPolicy') },
      { id: 'status', label: $i18n.t('generic.label.status') },
      ...additionalColumns.map(item => ({ id: getAdditionalColumnId(item), label: item.name })),
      { id: 'age', label: 'Age' },
      { id: 'creator', label: $i18n.t('generic.label.createdBy') },
      { id: 'source', label: $i18n.t('generic.label.source') },
      { id: 'editMode', label: $i18n.t('generic.label.editMode.text') },
    ];

    return {
      getAdditionalColumnId,
      getTableSettingFields,
    };
  },
});
</script>
