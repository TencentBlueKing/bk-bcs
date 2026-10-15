<!-- eslint-disable max-len -->
<template>
  <div
    :class="[
      'cluster-node',
      { 'px-[24px] py-[16px]': !fromCluster }
    ]">
    <bcs-alert type="info" class="cluster-node-tip">
      <div slot="title">
        {{$t('cluster.nodeList.article1')}}
        <i18n v-if="isShowTableAlertAfter" path="cluster.nodeList.article2">
          <template #nodes>
            <span class="num">{{ clusterData.extraInfo?.clusterCurNodeNum || '--' }}</span>
          </template>
          <template #realRemainNodesCount>
            <span class="num">{{ clusterData.extraInfo?.clusterSupNodeNum || '--' }}</span>
          </template>
        </i18n>
        <template v-if="clusterData.provider === 'tencentCloud'">
          <i18n path="cluster.nodeList.article3">
            <template #maxRemainNodesCount>
              <span class="num">{{ clusterData.extraInfo?.clusterMaxNodeNum || '--' }}</span>
            </template>
          </i18n>
        </template>
      </div>
    </bcs-alert>
    <!-- 修改节点转移模块 -->
    <div class="flex items-center text-[12px]">
      <div class="text-[#979BA5] bcs-border-tips" v-bk-tooltips="$t('tke.tips.transferNodeCMDBModule')">
        {{ $t('tke.label.nodeModule.text') }}
      </div>
      <span class="mx-[4px]">:</span>
      <template v-if="isEditModule">
        <div class="flex items-center">
          <TopoSelector
            :placeholder="$t('generic.placeholder.select')"
            :cluster-id="clusterId"
            v-model="curModuleID"
            class="w-[360px]"
            @change="handleWorkerModuleChange"
            @node-data-change="handleNodeChange" />
          <span
            class="text-[12px] text-[#3a84ff] ml-[8px] cursor-pointer"
            text
            @click="handleSaveWorkerModule">{{ $t('generic.button.save') }}</span>
          <span
            class="text-[12px] text-[#3a84ff] ml-[8px] cursor-pointer"
            text
            @click="isEditModule = false">{{ $t('generic.button.cancel') }}</span>
        </div>
      </template>
      <template v-else>
        <span>
          {{ clusterData.clusterBasicSettings && clusterData.clusterBasicSettings.module
            ? clusterData.clusterBasicSettings.module.workerModuleName || '--'
            : '--' }}
        </span>
        <span
          :class="[
            'hover:text-[#3a84ff] cursor-pointer ml-[8px]',
            isGkeManagedCluster ? '!cursor-not-allowed' : ''
          ]"
          @click="handleEditWorkerModule">
          <i class="bk-icon icon-edit-line"></i>
        </span>
      </template>
    </div>
    <bcs-divider></bcs-divider>
    <!-- 操作栏 -->
    <div class="cluster-node-operate">
      <div class="left">
        <template v-if="fromCluster">
          <span
            v-bk-tooltips="{
              disabled: !isKubeConfigOrAgentImportCluster,
              content: $t('cluster.nodeList.tips.disableImportClusterAction')
            }">
            <bcs-button
              theme="primary"
              icon="plus"
              class="add-node mr10"
              v-authority="{
                clickable: webAnnotations.perms[localClusterId]
                  && webAnnotations.perms[localClusterId].cluster_manage,
                actionId: 'cluster_manage',
                resourceName: curSelectedCluster.clusterName,
                disablePerms: true,
                permCtx: {
                  project_id: curProject.project_id,
                  cluster_id: localClusterId
                }
              }"
              :disabled="disableAddNodeBtn"
              @click="handleAddNode">
              {{$t('cluster.nodeList.create.text')}}
            </bcs-button>
          </span>
        </template>
        <bcs-dropdown-menu
          :disabled="!selections.length"
          :class="['mr10', { 'from-cluster': fromCluster }]"
          trigger="click"
          @hide="showBatchMenu = false"
          @show="showBatchMenu = true">
          <template #dropdown-trigger>
            <bcs-button>
              <div class="h-[30px]">
                <span class="text-[14px]">{{$t('cluster.nodeList.button.batch')}}</span>
                <i :class="['bk-icon icon-angle-down', { 'icon-flip': showBatchMenu }]"></i>
              </div>
            </bcs-button>
          </template>
          <ul
            class="bk-dropdown-list"
            slot="dropdown-content"
            v-authority="{
              clickable: webAnnotations.perms[localClusterId]
                && webAnnotations.perms[localClusterId].cluster_manage,
              actionId: 'cluster_manage',
              resourceName: curSelectedCluster.clusterName,
              disablePerms: true,
              permCtx: {
                project_id: curProject.project_id,
                cluster_id: localClusterId
              }
            }">
            <li @click="handleBatchEnableNodes">{{$t('generic.button.uncordon.text')}}</li>
            <li @click="handleBatchStopNodes">{{$t('generic.button.cordon.text')}}</li>
            <!-- 'REMOVE-FAILURE', 'ADD-FAILURE' 才支持删除 -->
            <li
              :disabled="isKubeConfigOrAgentImportCluster
                || selections.some(item => !['REMOVE-FAILURE', 'ADD-FAILURE'].includes(item.status))"
              v-bk-tooltips="{
                disabled: !isKubeConfigOrAgentImportCluster,
                content: $t('cluster.nodeList.tips.disableImportClusterAction')
              }"
              @click="handleBatchReAddNodes">{{$t('cluster.nodeList.button.retry')}}</li>
            <div
              class="h-[32px]"
              v-bk-tooltips="{ content: $t('generic.button.drain.tips'), disabled: !podDisabled, placement: 'right' }">
              <li :disabled="podDisabled" @click="handleBatchPodScheduler">{{$t('generic.button.drain.text')}}</li>
            </div>
            <div
              v-bk-tooltips="{
                content: $t('cluster.nodeList.tips.disableBatchSettingNodes'),
                disabled: !selections.some(item => !['RUNNING'].includes(item.status)),
                placement: 'right'
              }">
              <li
                :disabled="selections.some(item => !['RUNNING'].includes(item.status))"
                @click="handleBatchSetNode('labels')">
                {{$t('cluster.nodeList.button.setLabel')}}
              </li>
            </div>
            <div
              v-bk-tooltips="{
                content: $t('cluster.nodeList.tips.disableBatchSettingNodes'),
                disabled: !selections.some(item => !['RUNNING'].includes(item.status)),
                placement: 'right'
              }">
              <li
                :disabled="selections.some(item => !['RUNNING'].includes(item.status))"
                @click="handleBatchSetNode('taints')">
                {{$t('cluster.nodeList.button.setTaint')}}
              </li>
            </div>
            <div
              v-bk-tooltips="{
                content: $t('cluster.nodeList.tips.disableBatchSettingNodes'),
                disabled: !selections.some(item => !['RUNNING'].includes(item.status)),
                placement: 'right'
              }">
              <li
                :disabled="selections.some(item => !['RUNNING'].includes(item.status))"
                @click="handleBatchSetNode('annotations')">
                {{$t('dashboard.ns.action.setAnnotation')}}
              </li>
            </div>
            <div
              class="h-[32px]"
              v-bk-tooltips="{
                content: disableBatchDeleteTips,
                disabled: !disableBatchDelete,
                placement: 'right'
              }">
              <li
                :disabled="disableBatchDelete"
                @click="handleBatchDeleteNodes">{{$t('generic.button.delete')}}</li>
            </div>
          </ul>
        </bcs-dropdown-menu>
        <BcsCascade
          :list="copyList"
          @click="handleCopy"
          @hide="showCopyMenu = false"
          @show="showCopyMenu = true">
          <bcs-button>
            <div class="h-[30px]">
              <span class="text-[14px]">{{$t('cluster.nodeList.button.copy.text')}}</span>
              <i :class="['bk-icon icon-angle-down', { 'icon-flip': showCopyMenu }]"></i>
            </div>
          </bcs-button>
        </BcsCascade>
      </div>
      <div class="right flex-1 ml-[10px]">
        <ClusterSelect
          class="mr10 w-[254px]"
          v-model="localClusterId"
          @change="handleClusterChange"
          v-if="!hideClusterSelect"
        />
        <bcs-search-select
          clearable
          class="bg-[#fff] flex-1"
          :data="searchSelectData"
          :show-condition="false"
          :show-popover-tag-change="false"
          :placeholder="$t('cluster.nodeList.placeholder.searchNode')"
          default-focus
          v-model="searchSelectValue"
          @change="searchSelectChange"
          @clear="handleClearSearchSelect">
        </bcs-search-select>
        <div
          class="flex items-center justify-center w-[32px] h-[32px] text-[12px] cursor-pointer ml-[10px] bcs-border"
          @click="handleGetNodeData">
          <i class="bcs-icon bcs-icon-reset"></i>
        </div>
      </div>
    </div>
    <!-- 节点列表 -->
    <div class="mt-[20px]" v-bkloading="{ isLoading: tableLoading }">
      <bcs-table
        :size="tableSetting.size"
        :data="curPageData"
        ref="tableRef"
        ext-cls="empty-center"
        empty-block-class-name="bcs-table-empty-dynamic-width"
        :key="tableKey"
        :pagination="pagination"
        :row-key="(row) => row.nodeName || row.nodeID"
        @filter-change="handleFilterChange"
        @page-change="pageChange"
        @page-limit-change="pageSizeChange">
        <template #prepend>
          <transition name="fade">
            <div class="selection-tips" v-if="selectType !== CheckType.Uncheck">
              <i18n path="cluster.nodeList.msg.selectedData">
                <template #num>
                  <span class="tips-num">{{selections.length}}</span>
                </template>
              </i18n>
              <bk-button
                ext-cls="tips-btn"
                text
                v-if="selectType === CheckType.AcrossChecked"
                @click="handleClearSelection">
                {{ $t('cluster.nodeList.button.cancelSelectAll') }}
              </bk-button>
              <bk-button
                ext-cls="tips-btn"
                text
                v-else
                @click="handleSelectionAll">
                <i18n path="cluster.nodeList.msg.selectedAllData">
                  <template #num>
                    <span class="tips-num">{{pagination.count}}</span>
                  </template>
                </i18n>
              </bk-button>
            </div>
          </transition>
        </template>
        <bcs-table-column
          :render-header="renderSelection"
          width="70"
          :resizable="false"
          fixed="left">
          <template #default="{ row }">
            <bcs-checkbox
              :checked="selections.some(item => item.nodeName === row.nodeName && item.nodeID === row.nodeID)"
              :disabled="['INITIALIZATION', 'DELETING', 'APPLYING'].includes(row.status)"
              @change="(value) => handleRowCheckChange(value, row)"
            />
          </template>
        </bcs-table-column>
        <bcs-table-column :label="$t('cluster.nodeList.label.name')" min-width="120" prop="nodeName" fixed="left" show-overflow-tooltip>
          <template #default="{ row }">
            <bcs-button
              :disabled="['INITIALIZATION', 'DELETING'].includes(row.status) || !row.nodeName"
              text
              v-authority="{
                clickable: webAnnotations.perms[localClusterId]
                  && webAnnotations.perms[localClusterId].cluster_view,
                actionId: 'cluster_view',
                resourceName: curSelectedCluster.clusterName,
                disablePerms: true,
                permCtx: {
                  project_id: curProject.project_id,
                  cluster_id: localClusterId
                }
              }"
              @click="handleGoOverview(row)">
              <span class="bcs-ellipsis">{{ row.nodeName || '--' }}</span>
            </bcs-button>
          </template>
        </bcs-table-column>
        <bcs-table-column label="IPv4" width="150" prop="innerIP" sortable show-overflow-tooltip>
          <template #default="{ row }">
            {{ row.innerIP || '--' }}
          </template>
        </bcs-table-column>
        <bcs-table-column
          label="IPv6"
          prop="innerIPv6"
          min-width="200"
          sortable
          key="innerIPv6"
          show-overflow-tooltip
          v-if="isColumnRender('innerIPv6')">
          <template #default="{ row }">
            {{ row.innerIPv6 || '--' }}
          </template>
        </bcs-table-column>
        <bcs-table-column
          :label="$t('cluster.nodeList.label.source.text')"
          :filters="filtersDataSource.nodeSource"
          :filtered-value="filteredValue.nodeSource"
          column-key="nodeSource"
          prop="nodeSource"
          min-width="130"
          v-if="isColumnRender('nodeSource')">
          <template #default="{ row }">
            {{ row.nodeGroupID ? $t('cluster.ca.nodePool.text') : $t('cluster.nodeList.label.source.add') }}
          </template>
        </bcs-table-column>
        <bcs-table-column
          :label="$t('cluster.nodeList.label.nodePool')"
          :filters="filtersDataSource.nodeGroup"
          :filtered-value="filteredValue.nodeGroup"
          column-key="nodeGroupID"
          prop="nodeGroupID"
          min-width="130"
          show-overflow-tooltip
          v-if="isColumnRender('nodeGroupID')">
          <template #default="{ row }">{{ row.nodeGroupName || '--' }}</template>
        </bcs-table-column>
        <bcs-table-column
          :label="$t('generic.label.status')"
          :filters="filtersDataSource.status"
          :filtered-value="filteredValue.status"
          min-width="160"
          column-key="status"
          prop="status"
          show-overflow-tooltip>
          <template #default="{ row }">
            <LoadingIcon
              v-if="['INITIALIZATION', 'DELETING', 'APPLYING'].includes(row.status)"
            >
              <span class="bcs-ellipsis">{{ nodeStatusMap[row.status.toLowerCase()] }}</span>
            </LoadingIcon>
            <StatusIcon
              :status="row.status"
              :status-color-map="nodeStatusColorMap"
              v-else
            >
              <span
                :class="['bcs-ellipsis flex-1', row.failedReason ? 'bcs-border-tips !flex-none' : '']"
                v-bk-tooltips="{
                  content: row.failedReason,
                  disabled: !row.failedReason
                }">
                {{ nodeStatusMap[row.status.toLowerCase()] }}</span>
            </StatusIcon>
          </template>
        </bcs-table-column>
        <bcs-table-column
          :label="$t('cluster.ca.nodePool.create.az.title')"
          :filters="filtersDataSource.zoneID"
          :filtered-value="filteredValue.zoneID"
          min-width="160"
          column-key="zoneID"
          prop="zoneName"
          show-overflow-tooltip
          v-if="isColumnRender('zoneID')">
          <template #default="{ row }">
            {{ row.zoneName || row.zoneID ||'--' }}
          </template>
        </bcs-table-column>
        <!-- 容器数量 -->
        <bcs-table-column
          :label="$t('dashboard.workload.container.counts')"
          min-width="100"
          align="right"
          prop="container_count"
          key="container_count"
          v-if="isColumnRender('container_count')">
          <template #default="{ row }">
            <template v-if="['RUNNING', 'REMOVABLE'].includes(row.status)">
              <LoadingIcon v-if="!nodeMetric[row.nodeName] && overviewsLoading" class="justify-end">
                {{ `${$t('generic.status.loading')}...` }}
              </LoadingIcon>
              <span v-else>
                {{nodeMetric[row.nodeName]?.container_count || '--'}}
              </span>
            </template>
            <span v-else>--</span>
          </template>
        </bcs-table-column>
        <!-- Pod数量 -->
        <bcs-table-column
          :label="$t('cluster.nodeList.label.podCounts')"
          min-width="100"
          align="right"
          prop="pod_count"
          key="pod_count"
          v-if="isColumnRender('pod_count')">
          <template #default="{ row }">
            <template v-if="['RUNNING', 'REMOVABLE'].includes(row.status)">
              <LoadingIcon v-if="!nodeMetric[row.nodeName] && overviewsLoading" class="justify-end">
                {{ `${$t('generic.status.loading')}...` }}
              </LoadingIcon>
              <span v-else>
                {{nodeMetric[row.nodeName]?.pod_count || '--'}}
              </span>
            </template>
            <span v-else>--</span>
          </template>
        </bcs-table-column>
        <bcs-table-column min-width="200" :label="$t('k8s.label')" key="labels" v-if="isColumnRender('labels')">
          <template #default="{ row }">
            <span v-if="!row.labels || !Object.keys(row.labels).length">--</span>
            <bcs-popover v-else :delay="300" placement="top" class="popover">
              <div class="row-label">
                <span class="label" v-for="key in Object.keys(row.labels)" :key="key">
                  {{ `${key}=${row.labels[key]}` }}
                </span>
              </div>
              <template slot="content">
                <div class="labels-tips">
                  <div v-for="key in Object.keys(row.labels)" :key="key">
                    <span>{{ `${key}=${row.labels[key]}` }}</span>
                  </div>
                </div>
              </template>
            </bcs-popover>
          </template>
        </bcs-table-column>
        <bcs-table-column min-width="200" :label="$t('k8s.taint')" key="taint" v-if="isColumnRender('taint')">
          <template #default="{ row }">
            <span v-if="!row.taints || !row.taints.length">--</span>
            <bcs-popover v-else :delay="300" placement="top" class="popover">
              <div class="row-label">
                <span class="label" v-for="(taint, index) in row.taints" :key="index">
                  {{
                    `${taint.key}=${taint.value && taint.effect
                      ? taint.value + ' : ' + taint.effect
                      : taint.value || taint.effect}`
                  }}
                </span>
              </div>
              <template slot="content">
                <div class="labels-tips">
                  <div class="label" v-for="(taint, index) in row.taints" :key="index">
                    {{
                      `${taint.key}=${taint.value && taint.effect
                        ? taint.value + ' : ' + taint.effect
                        : taint.value || taint.effect}`
                    }}
                  </div>
                </div>
              </template>
            </bcs-popover>
          </template>
        </bcs-table-column>
        <bcs-table-column min-width="200" :label="$t('k8s.annotation')" key="annotations" v-if="isColumnRender('annotations')">
          <template #default="{ row }">
            <span v-if="!row.annotations || !Object.keys(row.annotations).length">--</span>
            <bcs-popover v-else :delay="300" placement="top" class="popover">
              <div class="row-label">
                <span class="label" v-for="key in Object.keys(row.annotations)" :key="key">
                  {{ `${key}=${row.annotations[key]}` }}
                </span>
              </div>
              <template slot="content">
                <div class="labels-tips">
                  <div v-for="key in Object.keys(row.annotations)" :key="key">
                    <span>{{ `${key}=${row.annotations[key]}` }}</span>
                  </div>
                </div>
              </template>
            </bcs-popover>
          </template>
        </bcs-table-column>
        <!-- 指标 -->
        <bcs-table-column
          v-for="item in metricColumnConfig"
          :label="item.label"
          :sort-method="(pre, next) => sortMetricMethod(pre, next, item.prop)"
          :key="item.prop"
          sortable
          align="center"
          min-width="120">
          <template #default="{ row }">
            <template v-if="['RUNNING', 'REMOVABLE'].includes(row.status)">
              <LoadingIcon v-if="!nodeMetric[row.nodeName] && overviewsLoading" class="justify-center">
                {{ `${$t('generic.status.loading')}...` }}
              </LoadingIcon>
              <template v-else>
                <RingCell
                  :percent="nodeMetric[row.nodeName]?.[item.prop]"
                  :fill-color="item.color"
                  :key="row.nodeName"
                  v-bk-tooltips="{
                    disabled: !item.percent,
                    content: nodeMetric[row.nodeName]?.[`${item.prop}_tips`]
                  }"
                  v-if="nodeMetric[row.nodeName] && nodeMetric[row.nodeName]?.[item.prop]" />
                <span v-bk-tooltips="{ content: $t('generic.msg.error.data') }" v-else>--</span>
              </template>
            </template>
            <template v-else>--</template>
          </template>
        </bcs-table-column>
        <bcs-table-column :label="$t('generic.label.action')" width="220" :resizable="false" fixed="right">
          <template #default="{ row }">
            <bk-button class="mr10" text v-if="row.status === 'APPLY-FAILURE'" @click="handleDeleteNode(row)">
              {{ $t('generic.button.delete') }}
            </bk-button>
            <div
              class="node-operate-wrapper"
              v-authority="{
                clickable: webAnnotations.perms[localClusterId]
                  && webAnnotations.perms[localClusterId].cluster_manage,
                actionId: 'cluster_manage',
                resourceName: curSelectedCluster.clusterName,
                disablePerms: true,
                permCtx: {
                  project_id: curProject.project_id,
                  cluster_id: localClusterId
                }
              }"
              v-else-if="!['APPLYING'].includes(row.status)">
              <bk-button
                class="mr10"
                text
                @click="handleStopNode(row)"
                v-if="row.status === 'RUNNING'">
                {{ $t('generic.button.cordon.text') }}
              </bk-button>
              <bk-button v-else-if="row.status === 'REMOVABLE'" text class="mr10" @click="handleEnableNode(row)">
                {{ $t('generic.button.uncordon.text') }}
              </bk-button>
              <span
                v-bk-tooltips="{
                  content: $t('generic.button.drain.tips'),
                  disabled: ['REMOVABLE', 'REMOVE-FAILURE', 'REMOVE-CA-FAILURE'].includes(row.status),
                  placement: 'top',
                  interactive: false
                }">
                <bk-button
                  text
                  :disabled="!['REMOVE-FAILURE', 'REMOVE-CA-FAILURE', 'REMOVABLE'].includes(row.status)"
                  class="mr10"
                  @click="handleSchedulerNode(row)">
                  {{ $t('generic.button.drain.text') }}
                </bk-button>
              </span>
              <bk-button
                class="mr10"
                text
                v-if="['INITIALIZATION', 'DELETING', 'REMOVE-FAILURE', 'ADD-FAILURE', 'REMOVE-CA-FAILURE'].includes(row.status)"
                :disabled="!row.inner_ip"
                @click="handleShowLog(row)"
              >
                {{$t('generic.button.log')}}
              </bk-button>
              <bk-button
                text
                class="mr10"
                v-if="['REMOVE-FAILURE', 'ADD-FAILURE'].includes(row.status)"
                :disabled="!row.inner_ip || isCloudSelfNode(row)"
                @click="handleRetry(row)"
              >{{ $t('cluster.ca.nodePool.records.action.retry') }}</bk-button>
              <bk-popover
                placement="bottom"
                theme="light dropdown"
                :arrow="false"
                :disabled="['INITIALIZATION', 'DELETING', 'REMOVE-CA-FAILURE'].includes(row.status)"
                trigger="click">
                <span :class="['bcs-icon-more-btn', { disabled: ['INITIALIZATION', 'DELETING', 'REMOVE-CA-FAILURE'].includes(row.status) }]">
                  <i class="bcs-icon bcs-icon-more"></i>
                </span>
                <template #content>
                  <ul class="bcs-dropdown-list bg-[#fff]">
                    <template v-if="row.status === 'RUNNING'">
                      <li class="bcs-dropdown-item" @click="handleMetadata(row, 'labels')">
                        {{$t('cluster.nodeList.button.setLabel')}}
                      </li>
                      <li class="bcs-dropdown-item" @click="handleSetTaint(row)">
                        {{$t('cluster.nodeList.button.setTaint')}}
                      </li>
                      <li class="bcs-dropdown-item" @click="handleMetadata(row, 'annotations')">
                        {{$t('dashboard.ns.action.setAnnotation')}}
                      </li>
                    </template>
                    <li
                      :class="[
                        'bcs-dropdown-item',
                        { disabled: isKubeConfigOrAgentImportCluster || isCloudSelfNode(row) || isGkeManagedCluster
                          || !['REMOVE-FAILURE', 'ADD-FAILURE', 'REMOVABLE', 'NOTREADY'].includes(row.status)
                        }
                      ]"
                      v-bk-tooltips="{
                        disabled: !isKubeConfigOrAgentImportCluster
                          && !isCloudSelfNode(row)
                          && ['REMOVE-FAILURE', 'ADD-FAILURE', 'REMOVABLE', 'NOTREADY'].includes(row.status),
                        content: getDisableDeleteTips(row),
                        placement: 'right'
                      }"
                      :disabled="!row.inner_ip"
                      @click="handleDeleteNode(row)"
                    >
                      {{ $t('generic.button.delete') }}
                    </li>
                  </ul>
                </template>
              </bk-popover>
            </div>
            <bk-button
              class="mr10"
              text
              v-else
              @click="handleShowLog(row)">
              {{$t('generic.button.log')}}
            </bk-button>
          </template>
        </bcs-table-column>
        <bcs-table-column type="setting" :resizable="false">
          <bcs-table-setting-content
            :fields="tableSetting.fields"
            :selected="tableSetting.selectedFields"
            :size="tableSetting.size"
            @setting-change="handleSettingChange">
          </bcs-table-setting-content>
        </bcs-table-column>
        <template #empty>
          <BcsEmptyTableStatus :type="searchSelectValue.length ? 'search-empty' : 'empty'" @clear="searchSelectValue = []" />
        </template>
      </bcs-table>
    </div>
    <!-- 设置标签/注解 -->
    <bcs-sideslider
      :is-show.sync="setMetaConf.isShow"
      :width="750"
      :before-close="handleBeforeClose"
      quick-close
      transfer>
      <template #header>
        <span>{{setMetaConf.title}}</span>
        <span v-if="metaType === 'labels'" class="sideslider-tips">{{$t('cluster.nodeList.msg.labelDesc')}}</span>
      </template>
      <template #content>
        <KeyValue
          class="key-value-content"
          :model-value="setMetaConf.data"
          :loading="setMetaConf.btnLoading"
          :key-desc="setMetaConf.keyDesc"
          :key-rules="[
            {
              message: $i18n.t('generic.validate.labelKey1'),
              validator: KEY_REGEXP
            }
          ]"
          :value-rules="[
            {
              message: $i18n.t('generic.validate.labelKey1'),
              validator: metaType === 'labels' ? VALUE_REGEXP : ''
            }
          ]"
          :min-items="0"
          @data-change="setChanged(true)"
          @cancel="handleLabelEditCancel"
          @confirm="handleLabelEditConfirm"
        ></KeyValue>
      </template>
    </bcs-sideslider>
    <!-- 设置污点 -->
    <bcs-sideslider
      :is-show.sync="taintConfig.isShow"
      :title="`${$t('cluster.nodeList.button.setTaint')}(${taintConfig.nodes[0]?.nodeName})`"
      :width="750"
      :before-close="handleBeforeClose"
      quick-close
      transfer>
      <template #content>
        <TaintContent
          :cluster-id="localClusterId"
          :nodes="taintConfig.nodes"
          @data-change="setChanged(true)"
          @confirm="handleConfirmTaintDialog"
          @cancel="handleHideTaintDialog"
        />
      </template>
    </bcs-sideslider>
    <!-- pod 驱逐 -->
    <bcs-sideslider
      :is-show.sync="podConfig.isShow"
      :title="$t('generic.button.drain.text')"
      :width="960"
      :before-close="handleBeforeClose"
      quick-close
      transfer>
      <template #content>
        <PodDrain
          :nodes="podConfig.nodes"
          :cluster-id="clusterId"
          @cancel="handleHidePodDrain"
          @success="handleSuccess" />
      </template>
    </bcs-sideslider>
    <!-- 查看日志 -->
    <bk-sideslider
      :is-show.sync="logSideDialogConf.isShow"
      :title="logSideDialogConf.title || ' '"
      :width="960"
      @hidden="closeLog"
      :quick-close="true"
      transfer>
      <template #header>
        <div class="flex items-center">
          <TaskIcon :font-size="'16px'" :status="logSideDialogConf.status" />
          <span>{{logSideDialogConf.title || ' '}}</span>
        </div>
      </template>
      <div slot="content">
        <div v-bkloading="{ isLoading: logSideDialogConf.loading }">
          <TaskLog
            :data="logSideDialogConf.taskData"
            :enable-auto-refresh="['LOADING', 'WAITING'].includes(logSideDialogConf.status)"
            enable-statistics
            type="multi-task"
            :status="logSideDialogConf.status"
            :title="logSideDialogConf.title"
            :loading="logSideDialogConf.loading"
            :height="'calc(100vh - 92px)'"
            :rolling-loading="false"
            :show-step-retry-fn="handleStepRetry"
            :show-step-skip-fn="handleStepSkip"
            @refresh="handleShowLog(logSideDialogConf.row)"
            @auto-refresh="handleAutoRefresh"
            @download="getDownloadTaskRecords"
            @retry="(data) => handleRetry(logSideDialogConf.row, data)"
            @skip="handleSkip" />
        </div>
      </div>
    </bk-sideslider>
    <!-- 删除节点 -->
    <bcs-dialog
      v-model="showDeleteDialog"
      :show-footer="false"
      render-directive="if">
      <DeleteNode
        :title="$t('cluster.ca.nodePool.nodes.action.delete.title')"
        :sub-title="curCheckedNodes.length > 1
          ? $i18n.t('cluster.nodeList.button.delete.subTitle', {
            num: curCheckedNodes.length,
            ip: curCheckedNodes[0].innerIP || curCheckedNodes[0].nodeID,
          })
          : $t('cluster.ca.nodePool.nodes.action.delete.subTitle', { ip: curCheckedNodes[0].innerIP })"
        :is-loading="deleting"
        v-if="curCheckedNodes.length"
        @confirm="delNode"
        @cancel="showDeleteDialog = false">
        <div class="flex items-center justify-center mt-[8px]" v-if="curSelectedCluster.provider === 'tencentPublicCloud'">
          <bcs-checkbox
            v-model="deleteMode"
            true-value="terminate"
            false-value="retain">
            {{ $t('tke.label.retain') }}
          </bcs-checkbox>
        </div>
        <div v-if="curSelectedCluster.provider === 'tencentCloud'">
          <span>{{ $t('cluster.ca.nodePool.nodes.action.delete.article1') }}</span>
          <div>{{ $t('cluster.ca.nodePool.nodes.action.delete.article2') }}</div>
          <div>{{ $t('cluster.ca.nodePool.nodes.action.delete.article3') }}</div>
          <div>{{ $t('cluster.ca.nodePool.nodes.action.delete.article4') }}</div>
        </div>
      </DeleteNode>
    </bcs-dialog>
  </div>
</template>
<script lang="ts">
import { computed, defineComponent, onActivated, onBeforeMount, onBeforeUnmount, onMounted, ref, set, watch } from 'vue';
import { TranslateResult } from 'vue-i18n';

import TaskLog from '@blueking/task-log/vue2';

import useTableAcrossCheck from '../../../composables/use-table-across-check';
import useTableSearchSelect, { ISearchSelectData } from '../../../composables/use-table-search-select';
import useTableSetting from '../../../composables/use-table-setting';
import { useClusterInfo, useClusterList, useTask } from '../cluster/use-cluster';
import PodDrain from '../components/pod-drain.vue';
import TaintContent from '../components/taint.vue';

import DeleteNode from './delete-node.vue';
import useNode from './use-node';

import '@blueking/task-log/vue2/vue2.css';
import { clusterTaskRecords, setClusterModule, taskLogsDownloadURL } from '@/api/modules/cluster-manager';
import { parseUrl } from '@/api/request';
import $bkMessage from '@/common/bkmagic';
import { KEY_REGEXP, VALUE_REGEXP } from '@/common/constant';
import { copyText, formatBytes, padIPv6 } from '@/common/util';
import { CheckType } from '@/components/across-check.vue';
import $bkInfo from '@/components/bk-magic-2.0/bk-info';
import BcsCascade from '@/components/cascade.vue';
import ClusterSelect from '@/components/cluster-selector/cluster-select.vue';
import KeyValue, { IData } from '@/components/key-value.vue';
import LoadingIcon from '@/components/loading-icon.vue';
import StatusIcon from '@/components/status-icon';
import TaskIcon from '@/components/task-icon.vue';
import { ICluster, useAppData } from '@/composables/use-app';
import useInterval from '@/composables/use-interval';
import usePage from '@/composables/use-page';
import useSideslider from '@/composables/use-sideslider';
import $i18n from '@/i18n/i18n-setup';
import $router from '@/router';
import $store from '@/store';
import TopoSelector from '@/views/cluster-manage/autoscaler/components/topo-select-tree.vue';
import RingCell from '@/views/cluster-manage/components/ring-cell.vue';

interface IMetricData {
  container_count: string
  cpu_request: string
  cpu_request_usage: string
  cpu_total: string
  cpu_usage: string
  cpu_used: string
  disk_total: string
  disk_usage: string
  disk_used: string
  diskio_usage: string
  memory_request: string
  memory_request_usage: string
  memory_total: string
  memory_usage: string
  memory_used: string
  pod_count: string
  pod_total: string
}

type NodeMetricType = Record<string, IMetricData>;

type MetaType = 'labels'|'taints'|'annotations';

export default defineComponent({
  name: 'NodeList',
  components: {
    StatusIcon,
    LoadingIcon,
    ClusterSelect,
    RingCell,
    KeyValue,
    TaintContent,
    PodDrain,
    TaskLog,
    BcsCascade,
    TopoSelector,
    DeleteNode,
    TaskIcon,
  },
  props: {
    clusterId: {
      type: String,
      default: '',
    },
    fromCluster: {
      type: Boolean,
      default: false,
    },
    hideClusterSelect: {
      type: Boolean,
      default: false,
    },
  },
  setup(props) {
    const webAnnotations = computed(() => $store.state.cluster.clusterWebAnnotations);
    const curProject = computed(() => $store.state.curProject);

    const { flagsMap } = useAppData();

    // 修改节点转移模块设置
    const { clusterData, getClusterDetail } = useClusterInfo();// clusterData和curCluster一样，就是多了云上的数据信息
    const isGkeManagedCluster = computed(() => clusterData.value.provider === 'gcpCloud'
      && clusterData.value.extraInfo?.clusterType === 'autopilot');
    const isEditModule = ref(false);
    const curModuleID = ref();
    const curNodeModule = ref<Record<string, any>>({});
    const handleEditWorkerModule = () => {
      if (isGkeManagedCluster.value) return;
      curModuleID.value = Number(clusterData.value.clusterBasicSettings?.module?.workerModuleID);
      isEditModule.value = true;
    };
    const handleWorkerModuleChange = (moduleID) => {
      curModuleID.value = moduleID;
    };
    const handleNodeChange = (node) => {
      curNodeModule.value = node;
    };
    const handleSaveWorkerModule = async () => {
      // if (String(curModuleID.value) === String(clusterData.value.clusterBasicSettings?.module?.workerModuleID)) {
      //   isEditModule.value = false;
      //   return;
      // };

      $bkInfo({
        type: 'warning',
        clsName: 'custom-info-confirm',
        title: $i18n.t('tke.title.confirmUpdateNodeCMDBModule'),
        subTitle: $i18n.t('tke.title.confirmUpdateNodeCMDBModuleSubTitle', [curNodeModule.value.path]),
        defaultInfo: true,
        confirmFn: async () => {
          tableLoading.value = true;
          const result = await setClusterModule({
            $clusterId: props.clusterId,
            module: {
              workerModuleID: String(curModuleID.value),
            },
            operator: $store.state.user?.username,
          }).then(() => true)
            .catch(() => false);
          if (result) {
            await getClusterDetail(curSelectedCluster.value.clusterID || '', true);
            $bkMessage({
              theme: 'success',
              message: $i18n.t('generic.msg.success.modify'),
            });
            isEditModule.value = false;
          }
          tableLoading.value = false;
        },
      });
    };

    // 侧滑关闭交互
    const { reset, setChanged, handleBeforeClose } = useSideslider();
    const nodeStatusColorMap = {
      initialization: 'blue',
      running: 'green',
      deleting: 'blue',
      'add-failure': 'red',
      'remove-failure': 'red',
      'remove-ca-failure': 'red',
      'apply-failure': 'red',
      removable: '',
      notready: 'red',
      unknown: '',
    };
    const nodeStatusMap = {
      initialization: window.i18n.t('generic.status.initializing'),
      running: window.i18n.t('generic.status.ready'),
      deleting: window.i18n.t('generic.status.deleting'),
      'add-failure': window.i18n.t('cluster.nodeList.status.addNodeFailed'),
      'remove-failure': window.i18n.t('cluster.nodeList.status.deleteNodeFailed'),
      'remove-ca-failure': window.i18n.t('cluster.nodeList.status.scaleOKButRemoveFailed'),
      'apply-failure': window.i18n.t('cluster.nodeList.status.applyFailure'),
      applying: window.i18n.t('cluster.nodeList.status.applying'),
      removable: window.i18n.t('generic.status.removable'),
      notready: window.i18n.t('generic.status.notReady'),
      unknown: window.i18n.t('generic.status.unknown1'),
    };
    // 表格表头搜索项配置
    const filtersDataSource = computed(() => ({
      status: status.value,
      nodeSource: [
        {
          text: $i18n.t('cluster.nodeList.label.source.add'),
          value: 'custom',
        },
        {
          text: $i18n.t('cluster.ca.nodePool.text'),
          value: 'nodepool',
        },
      ],
      nodeGroup: nodeGroupList.value,
      zoneID: zoneList.value,
    }));
    // 表格搜索项选中值
    const filteredValue = ref({
      status: [],
      nodeSource: [],
      nodeGroup: [],
      zoneID: [],
    });
    // searchSelect数据源配置
    const searchSelectDataSource = computed<ISearchSelectData[]>(() => [
      {
        name: $i18n.t('generic.label.ip'),
        id: 'ip',
        placeholder: $i18n.t('generic.placeholder.ipInput'),
      },
      {
        name: $i18n.t('generic.label.status'),
        id: 'status',
        multiable: true,
        children: status.value,
      },
      {
        name: $i18n.t('cluster.ca.nodePool.create.az.title'),
        id: 'zoneID',
        multiable: true,
        children: zoneList.value,
      },
      {
        name: $i18n.t('k8s.label'),
        id: 'labels',
        multiable: true,
        children: labels.value.map(label => ({
          id: label,
          name: label,
        })),
      },
      {
        name: $i18n.t('k8s.taint'),
        id: 'taints',
        multiable: true,
        children: taints.value,
      },
      {
        name: $i18n.t('k8s.annotation'),
        id: 'annotations',
        multiable: true,
        children: annotations.value.map(label => ({
          id: label,
          name: label,
        })),
      },
      {
        name: $i18n.t('cluster.nodeList.label.source.text'),
        id: 'nodeSource',
        multiable: true,
        children: [
          {
            id: 'custom',
            name: $i18n.t('cluster.nodeList.label.source.add'),
          },
          {
            id: 'nodepool',
            name: $i18n.t('cluster.ca.nodePool.text'),
          },
        ],
      },
      {
        name: $i18n.t('cluster.nodeList.label.nodePool'),
        id: 'nodeGroupID',
        multiable: true,
        children: nodeGroupList.value,
      },
    ]);
    // 表格搜索联动
    const {
      tableKey,
      searchSelectData,
      searchSelectValue,
      handleFilterChange,
      handleSearchSelectChange,
      handleClearSearchSelect,
    } = useTableSearchSelect({
      searchSelectDataSource,
      filteredValue,
    });
    const searchSelectChange = (list) => {
      handleResetCheckStatus();
      handleSearchSelectChange(list);
    };

    watch(searchSelectValue, () => {
      handleResetPage();
    }, { deep: true });

    // 表格设置字段配置
    const fields = [
      {
        id: 'innerIPv6',
        label: 'IPv6',
      },
      {
        id: 'nodeSource',
        label: $i18n.t('cluster.nodeList.label.source.text'),
        defaultChecked: true,
      },
      {
        id: 'nodeGroupID',
        label: $i18n.t('cluster.nodeList.label.nodePool'),
        defaultChecked: true,
      },
      {
        id: 'zoneID',
        label: $i18n.t('cluster.ca.nodePool.create.az.title'),
        defaultChecked: true,
      },
      {
        id: 'container_count',
        label: $i18n.t('dashboard.workload.container.counts'),
        // disabled: true,
      },
      {
        id: 'pod_count',
        label: $i18n.t('cluster.nodeList.label.podCounts'),
        // disabled: true,
      },
      {
        id: 'labels',
        label: $i18n.t('k8s.label'),
      },
      {
        id: 'taint',
        label: $i18n.t('k8s.taint'),
      },
      {
        id: 'annotations',
        label: $i18n.t('k8s.annotation'),
      },
      {
        id: 'pod_usage',
        label: $i18n.t('metrics.podUsage'),
        // disabled: true,
      },
      {
        id: 'cpu_usage',
        label: $i18n.t('metrics.cpuUsage'),
        // disabled: true,
      },
      {
        id: 'memory_usage',
        label: $i18n.t('metrics.memUsage'),
        // disabled: true,
      },
      {
        label: $i18n.t('metrics.cpuRequestUsage.text'),
        id: 'cpu_request_usage',
      },
      {
        label: $i18n.t('metrics.memRequestUsage.text'),
        id: 'memory_request_usage',
      },
      {
        id: 'disk_usage',
        label: $i18n.t('metrics.diskUsage'),
        // disabled: true,
      },
      {
        id: 'diskio_usage',
        label: $i18n.t('metrics.diskIOUsage2'),
        // disabled: true,
      },
    ];
    // 表格指标列配置
    const metricColumnConfig = computed(() => {
      const data = [
        {
          label: $i18n.t('metrics.podUsage'),
          prop: 'pod_usage',
          color: '#3a84ff',
          percent: ['pod_count', 'pod_total'], // 分子和分母
          unit: 'int',
        },
        {
          label: $i18n.t('metrics.cpuUsage'),
          prop: 'cpu_usage',
          color: '#3ede78',
          percent: ['cpu_used', 'cpu_total'], // 分子和分母
          unit: 'cpu',
        },
        {
          label: $i18n.t('metrics.memUsage'),
          prop: 'memory_usage',
          color: '#3a84ff',
          percent: ['memory_used', 'memory_total'], // 分子和分母
          unit: 'byte',
        },
        {
          label: $i18n.t('metrics.cpuRequestUsage.text'),
          prop: 'cpu_request_usage',
          percent: ['cpu_request', 'cpu_total'], // 分子和分母
          color: '#3ede78',
          unit: 'cpu',
        },
        {
          label: $i18n.t('metrics.memRequestUsage.text'),
          prop: 'memory_request_usage',
          percent: ['memory_request', 'memory_total'], // 分子和分母
          color: '#3a84ff',
          unit: 'byte',
        },
        {
          label: $i18n.t('metrics.diskUsage'),
          prop: 'disk_usage',
          color: '#853cff',
          percent: ['disk_used', 'disk_total'], // 分子和分母
          unit: 'byte',
        },
        {
          label: $i18n.t('metrics.diskIOUsage'),
          prop: 'diskio_usage',
          color: '#853cff',
        },
      ];
      return data.filter(item => tableSetting.value.selectedFields.some(field => field.id === item.prop));
    });

    const {
      tableSetting,
      handleSettingChange,
      isColumnRender,
    } = useTableSetting(fields);

    const sortMetricMethod = (pre, next, prop) => {
      const preNumber = parseFloat(nodeMetric.value[pre.nodeName]?.[prop] || 0);
      const nextNumber = parseFloat(nodeMetric.value[next.nodeName]?.[prop] || 0);
      if (preNumber > nextNumber) {
        return -1;
      } if (preNumber < nextNumber) {
        return 1;
      }
      return 0;
    };

    const {
      getNodeList,
      getTaskData,
      handleCordonNodes,
      handleUncordonNodes,
      // schedulerNode,
      addNode,
      retryTask,
      setNodeLabels,
      setNodeAnnotations,
      batchDeleteNodes,
      getAllNodeOverview,
    } = useNode();

    const tableLoading = ref(false);
    const localClusterId = ref(props.clusterId);
    const { clusterList } = useClusterList();
    const curSelectedCluster = computed<Partial<ICluster>>(() => clusterList.value
      .find(item => item.clusterID === localClusterId.value) || {});
    const isImportCluster = computed(() => curSelectedCluster.value.clusterCategory === 'importer');
    // kubeConfig导入集群 或 控制面导入集群
    const isKubeConfigOrAgentImportCluster = computed(() => curSelectedCluster.value.clusterCategory === 'importer'
      && (curSelectedCluster.value.importCategory === 'kubeConfig' || curSelectedCluster.value.importCategory === 'machine'));
    const disableAddNodeBtn = computed(() =>
      // kubeConfig导入集群, 控制面导入集群 和未开启原生k8s集群时禁用掉添加节点按钮
      isKubeConfigOrAgentImportCluster.value || (!flagsMap.value.k8s && clusterData.value.provider === 'bluekingCloud') || isGkeManagedCluster.value);
    // cloud私有节点
    const isCloudSelfNode = row => curSelectedCluster.value.clusterCategory === 'importer'
      && (curSelectedCluster.value.provider === 'gcpCloud' || curSelectedCluster.value.provider === 'azureCloud'
      || curSelectedCluster.value.provider === 'huaweiCloud' || curSelectedCluster.value.provider === 'awsCloud')
      && !row.nodeGroupID;
    // 全量表格数据
    const tableData = ref<any[]>([]);

    // 状态
    const status = computed(() => tableData.value.reduce((pre, item) => {
      if (!pre.find(data => data.id === item.status)) {
        pre.push({
          id: item.status,
          name: nodeStatusMap[item.status?.toLocaleLowerCase()] || item.status,
          text: nodeStatusMap[item.status?.toLocaleLowerCase()] || item.status,
          value: item.status,
        });
      }
      return pre;
    }, []));
    // 节点池
    const nodeGroupList = computed(() => tableData.value.reduce<any[]>((pre, item) => {
      if (item.nodeGroupID && pre.every(data => data.id !== item.nodeGroupID)) {
        pre.push({
          id: item.nodeGroupID,
          name: item.nodeGroupName,
          text: item.nodeGroupName,
          value: item.nodeGroupID,
        });
      }
      return pre;
    }, []));
    // 可用区
    const zoneList = computed(() => tableData.value.reduce((pre, row) => {
      if (!row.zoneID) return pre;
      const data = pre.find(item => item.value === row.zoneID);
      if (!data) {
        pre.push({
          value: row.zoneID,
          text: `${row.zoneName || row.zoneID} (1)`,
          // 兼容两种数据源
          id: row.zoneID,
          name: `${row.zoneName || row.zoneID} (1)`,
          count: 1,
        });
      } else {
        data.count += 1;
        data.text = `${row.zoneName || row.zoneID} (${data.count})`;
        data.name = `${row.zoneName || row.zoneID} (${data.count})`;
      }
      return pre;
    }, []));
    // 标签
    const labels = computed(() => {
      const data: string[] = [];
      tableData.value.forEach((item) => {
        Object.keys(item.labels || {}).forEach((key) => {
          const label = `${key}=${item.labels[key]}`;
          const index = data.indexOf(label);
          index === -1 && data.push(label);
        });
      });
      return data;
    });
    // 注解
    const annotations = computed(() => {
      const data: string[] = [];
      tableData.value.forEach((item) => {
        Object.keys(item.annotations || {}).forEach((key) => {
          const label = `${key}=${item.annotations[key]}`;
          const index = data.indexOf(label);
          index === -1 && data.push(label);
        });
      });
      return data;
    });
    // 污点
    const taints = computed(() => {
      const data = tableData.value.reduce((pre, row) => {
        row.taints?.forEach((item) => {
          pre[`${item.key}=${item.value}:${item.effect}`] = true;
        });
        return pre;
      }, {});

      return Object.keys(data).map(item => ({ id: item, name: item }));
    });

    const parseSearchSelectValue = computed(() => {
      const searchValues: { id: string; value: Set<any> }[] = [];
      searchSelectValue.value.forEach((item) => {
        let tmp: string[] = [];
        if (Array.isArray(item.values)) {
          // 下拉选择搜索
          if (item.id === 'ip') {
            // 处理IP字段多值情况
            item.values.forEach((v) => {
              const splitCode = String(v?.id).indexOf('|') > -1 ? '|' : ' ';
              tmp.push(...v.id.trim().split(splitCode));
            });
          } else {
            tmp = item.values.map(v => v.id);
          }
        } else {
          // 自定义IP模糊搜索
          tmp = String(item.id).split(' ');
        }
        searchValues.push({
          id: item.id,
          value: new Set(tmp.map(t => padIPv6(t?.trim()))),
        });
      });
      return searchValues;
    });
    // 过滤后的表格数据(todo: 搜索性能优化)
    const filterTableData = computed(() => {
      if (!parseSearchSelectValue.value.length) return tableData.value;

      return tableData.value.filter(row => parseSearchSelectValue.value.every((item) => {
        if (['labels', 'annotations'].includes(item.id)) {
          return Object.keys(row[item.id]).some(key => item.value.has(`${key}=${row[item.id][key]}`));
        }
        if (item.id === 'taints') {
          return row.taints.some(taint => item.value.has(`${taint.key}=${taint.value}:${taint.effect}`));
        }
        if (item.id in row) {
          return item.value.has(row[item.id]);
        }
        return item.value.has(row.innerIP) || item.value.has(padIPv6(row.innerIPv6));
      }));
    });
    // 分页后的表格数据
    const {
      curPageData,
      pagination,
      pageChange,
      pageSizeChange,
      handleResetPage,
      pageConf,
    } = usePage(filterTableData);

    // 跨页全选
    const filterFailureTableData = computed(() => filterTableData.value
      .filter(item => !['INITIALIZATION', 'DELETING', 'APPLYING'].includes(item.status)));
    const filterFailureCurTableData = computed(() => curPageData.value
      .filter(item => !['INITIALIZATION', 'DELETING', 'APPLYING'].includes(item.status)));
    const {
      selectType,
      selections,
      handleResetCheckStatus,
      renderSelection,
      handleRowCheckChange,
      handleSelectionAll,
      handleClearSelection,
    } = useTableAcrossCheck({
      tableData: filterFailureTableData,
      curPageData: filterFailureCurTableData,
    });
    // kubeConfig导入、选中节点含有运行中状态、含有非节点池节点不让删除
    const disableBatchDelete = computed(() => isKubeConfigOrAgentImportCluster.value
    || selections.value.some(item => item.status === 'RUNNING')
    || (isImportCluster.value && selections.value.some(item => !item.nodeGroupID)));

    const disableBatchDeleteTips = computed(() => {
      if (isKubeConfigOrAgentImportCluster.value) {
        return $i18n.t('cluster.nodeList.tips.disableImportClusterAction');
      }
      if ((isImportCluster.value && selections.value.some(item => !item.nodeGroupID))) {
        return $i18n.t('cluster.nodeList.tips.hasNotNodePoolNode');
      }
      return $i18n.t('cluster.ca.nodePool.nodes.action.delete.tips');
    });

    const getDisableDeleteTips = (row) => {
      if (isKubeConfigOrAgentImportCluster.value) {
        return $i18n.t('cluster.nodeList.tips.disableImportClusterAction');
      }
      if (row.clusterCategory === 'importer' && !row.nodeGroupID) {
        return $i18n.t('cluster.nodeList.tips.notNodePoolNode');
      }
      return $i18n.t('cluster.ca.nodePool.nodes.action.delete.tips');
    };

    const handleGoOverview = (row) => {
      $router.push({
        name: 'clusterNodeOverview',
        params: {
          nodeName: row.nodeName,
          clusterId: row.cluster_id || localClusterId.value,
        },
      });
    };

    // IP复制
    const showCopyMenu = ref(false);
    const copyList = computed(() => [
      {
        id: 'checked',
        label: $i18n.t('cluster.nodeList.button.copy.checkedIP'),
        disabled: !selections.value.length,
        children: [
          {
            id: 'checked-ipv4',
            label: 'IPv4',
          },
          {
            id: 'checked-ipv6',
            label: 'IPv6',
          },
        ],
      },
      {
        id: 'all',
        label: $i18n.t('cluster.nodeList.button.copy.allIP'),
        children: [
          {
            id: 'all-ipv4',
            label: 'IPv4',
          },
          {
            id: 'all-ipv6',
            label: 'IPv6',
          },
        ],
      },
    ]);
    const handleCopy = (item) => {
      let ipData: string[] = [];
      switch (item.id) {
        case 'checked-ipv4':
          ipData = selections.value.map(data => data.innerIP).filter(ip => !!ip);
          break;
        case 'checked-ipv6':
          ipData = selections.value.map(data => data.innerIPv6).filter(ip => !!ip);
          break;
        case 'all-ipv4':
          ipData = tableData.value.map(data => data.innerIP).filter(ip => !!ip);
          break;
        case 'all-ipv6':
          ipData = tableData.value.map(item => item.innerIPv6).filter(ip => !!ip);
          break;
      }
      copyText(ipData.join('\n'));
      $bkMessage({
        theme: 'success',
        message: $i18n.t('generic.msg.success.copyIP', { num: ipData.length }),
      });
    };

    // 设置污点
    const taintConfig = ref<{
      isShow: boolean;
      nodes: any[];
    }>({
      isShow: false,
      nodes: [],
    });
    const handleSetTaint = (row) => {
      taintConfig.value.isShow = true;
      taintConfig.value.nodes = [row];
      reset();
    };
    const handleConfirmTaintDialog = () => {
      handleGetNodeData();
      handleHideTaintDialog();
    };
    const handleHideTaintDialog = () => {
      taintConfig.value.isShow = false;
      taintConfig.value.nodes = [];
    };

    // 设置标签/注解（批量设置标签的交互有点奇怪，后续优化）
    const setMetaConf = ref<{
      isShow: boolean;
      btnLoading: boolean;
      keyDesc: any;
      rows: any[];
      data: IData[];
      title: string;
    }>({
      isShow: false,
      btnLoading: false,
      keyDesc: '',
      rows: [],
      data: [],
      title: '',
    });
    const metaType = ref<'labels' | 'annotations'>('labels');
    const handleMetadata = async (selections: Record<string, any>[] | Record<string, any>, type: 'labels' | 'annotations') => {
      metaType.value = type;
      setMetaConf.value.isShow = true;
      const rows = Array.isArray(selections) ? selections : [selections];
      // 批量设置时暂时只展示相同Key的项
      const labelArr = rows.reduce<any[]>((pre, row) => {
        const label = row[type];
        Object.keys(label).forEach((key) => {
          const index = pre.findIndex(item => item.key === key);
          if (index > -1) {
            pre[index].value = '';
            pre[index].repeat += 1;
            pre[index].placeholder = $i18n.t('generic.placeholder.unChange');
          } else {
            pre.push({
              key,
              value: label[key],
              repeat: 1,
            });
          }
        });
        return pre;
      }, []).filter(item => item.repeat === rows.length);

      let title = '';
      if (rows.length > 1 && metaType.value === 'labels') {
        title = $i18n.t('cluster.nodeList.title.batchSetLabel.text'); // 批量设置标签
      } else if (rows.length > 1 && metaType.value === 'annotations') {
        title = $i18n.t('cluster.nodeList.title.batchSetAnnotation.text'); // 批量设置注解
      } else if (rows.length === 1 && metaType.value === 'labels') {
        title = `${$i18n.t('cluster.nodeList.button.setLabel')}(${rows[0]?.nodeName})`; // 设置标签
      } else if (rows.length === 1 && metaType.value === 'annotations') {
        title = `${$i18n.t('dashboard.ns.action.setAnnotation')}(${rows[0]?.nodeName})`; // 设置注解
      }
      set(setMetaConf, 'value', Object.assign(setMetaConf.value, {
        data: labelArr,
        rows,
        title,
        keyDesc: rows.length > 1 ? $i18n.t('cluster.nodeList.title.batchSetLabel.desc') : '',
      }));
      reset();
    };
    const handleLabelEditCancel = () => {
      set(setMetaConf, 'value', Object.assign(setMetaConf.value, {
        isShow: false,
        keyDesc: '',
        rows: [],
        title: '',
        data: {},
      }));
    };
    const mergeLabels = (_originLabels, _newLabels) => {
      const originLabels = JSON.parse(JSON.stringify(_originLabels));
      const newLabels = JSON.parse(JSON.stringify(_newLabels));
      // 批量编辑
      if (setMetaConf.value.rows.length > 1) {
        const oldLabels = setMetaConf.value.data;
        oldLabels.forEach((item) => {
          // eslint-disable-next-line no-prototype-builtins
          if (!newLabels.hasOwnProperty(item.key)) {
            // 删除去除的key
            delete originLabels[item.key];
          } else if (!newLabels[item.key]) {
            // 未修改的Key保持不变
            delete newLabels[item.key];
          }
        });
        return Object.assign({}, originLabels, newLabels);
      }
      return newLabels;
    };
    const handleLabelEditConfirm = async (labels) => {
      setMetaConf.value.btnLoading = true;
      let result = false;
      if (metaType.value === 'labels') {
        result = await setNodeLabels({
          clusterID: localClusterId.value,
          nodes: setMetaConf.value.rows.map(item => ({
            nodeName: item.nodeName,
            labels: mergeLabels(item.labels, labels),
          })),
        });
      } else if (metaType.value === 'annotations') {
        result = await setNodeAnnotations({
          clusterID: localClusterId.value,
          nodes: setMetaConf.value.rows.map(item => ({
            nodeName: item.nodeName,
            annotations: mergeLabels(item.annotations, labels),
          })),
        });
      }
      setMetaConf.value.btnLoading = false;
      if (result) {
        handleLabelEditCancel();
        handleResetCheckStatus();
        handleGetNodeData();
      }
    };

    // 弹窗二次确认
    const bkComfirmInfo = ({
      title, subTitle, callback,
    }: {
      title: TranslateResult;
      subTitle?: TranslateResult;
      callback: Function;
    }) => {
      $bkInfo({
        type: 'warning',
        clsName: 'custom-info-confirm',
        subTitle,
        title,
        defaultInfo: true,
        confirmFn: async () => {
          await callback();
        },
      });
    };

    // 停止调度
    const handleStopNode = (row) => {
      bkComfirmInfo({
        title: $i18n.t('generic.button.cordon.title', { ip: row.nodeName }),
        callback: async () => {
          const result = await handleCordonNodes({
            clusterID: row.cluster_id,
            nodes: [row.nodeName],
          });
          result && handleGetNodeData();
        },
      });
    };
    // 允许调度
    const handleEnableNode = (row) => {
      bkComfirmInfo({
        title: $i18n.t('generic.button.uncordon.title'),
        subTitle: $i18n.t('generic.button.uncordon.subTitle', { ip: row.nodeName }),
        callback: async () => {
          const result = await handleUncordonNodes({
            clusterID: row.cluster_id,
            nodes: [row.nodeName],
          });
          result && handleGetNodeData();
        },
      });
    };


    // Pod驱逐
    const podConfig = ref<{
      isShow: boolean;
      nodes: any[];
    }>({
      isShow: false,
      nodes: [],
    });
    const handleSchedulerNode = (row) => {
      podConfig.value.isShow = true;
      podConfig.value.nodes = [row];
    };
    function handleHidePodDrain() {
      podConfig.value.isShow = false;
      podConfig.value.nodes = [];
    }
    function handleSuccess() {
      handleHidePodDrain();
      handleGetNodeData();
    }

    // 节点删除
    const deleting = ref(false);
    const user = computed(() => $store.state.user);
    const deleteMode = ref<'retain'|'terminate'>('retain');
    const showDeleteDialog = ref(false);
    const curCheckedNodes = ref<any[]>([]);
    const handleHideDeleteDialog = () => {
      showDeleteDialog.value = false;
      curCheckedNodes.value = [];
    };
    const handleDeleteNode = async (row) => {
      if (isKubeConfigOrAgentImportCluster.value || isCloudSelfNode(row) || isGkeManagedCluster.value
        || !['REMOVE-FAILURE', 'ADD-FAILURE', 'REMOVABLE', 'NOTREADY', 'APPLY-FAILURE'].includes(row.status)) return;

      curCheckedNodes.value = [row];
      showDeleteDialog.value = true;
      // $bkInfo({
      //   type: 'warning',
      //   clsName: 'custom-info-confirm',
      //   title: $i18n.t('cluster.ca.nodePool.nodes.action.delete.title'),
      //   subTitle: $i18n.t('cluster.ca.nodePool.nodes.action.delete.subTitle', { ip: row.innerIP }),
      //   defaultInfo: true,
      //   confirmFn: async () => {
      //     await delNode(row.clusterID, [row]);
      //   },
      // });
    };
    const delNode = async () => {
      if (!curCheckedNodes.value.length) return;

      if (curCheckedNodes.value.length > 100) {
        $bkMessage({
          theme: 'warning',
          message: $i18n.t('cluster.validate.maxNumberOfDeletion'),
        });
        return;
      }
      const nodeIPs: string[] = [];
      const virtualNodeIDs: string[] = [];
      curCheckedNodes.value.forEach((row) => {
        if (row.innerIP) {
          nodeIPs.push(row.innerIP);
        } else if (row.nodeID) {
          virtualNodeIDs.push(row.nodeID);
        }
      });
      deleting.value = true;
      const result = await batchDeleteNodes({
        $clusterId: localClusterId.value,
        nodeIPs: nodeIPs.join(','),
        virtualNodeIDs: virtualNodeIDs.join(','),
        deleteMode: deleteMode.value,
        operator: user.value.username,
      });
      deleting.value = false;
      if (result) {
        $bkMessage({
          theme: 'success',
          message: $i18n.t('generic.msg.success.ok'),
        });
        handleGetNodeData();
        handleResetPage();
        handleResetCheckStatus();
        handleHideDeleteDialog();
      }
    };
    // 添加节点（任务重试时会调用重新添加接口）
    const addClusterNode = async (clusterId: string, nodeIps: string[]) => {
      stop();
      const result = await addNode({
        clusterId,
        nodeIps,
      });
      result && await handleGetNodeData();
      if (tableData.value.length) {
        start();
      }
    };
    // 节点重试
    const handleRetry = async (row, data?) => {
      if (data && !data?.step?.allowRetry) {
        $bkMessage({
          theme: 'warning',
          message: $i18n.t('cluster.title.allowRetry'),
        });
        return;
      }
      $bkInfo({
        type: 'warning',
        title: $i18n.t('cluster.title.retryTask'),
        clsName: 'custom-info-confirm default-info',
        subTitle: row.taskName || row.name,
        confirmFn: async () => {
          tableLoading.value = true;
          logSideDialogConf.value.loading = true;
          const result = await retryTask({
            clusterId: row.cluster_id,
            nodeIP: row.inner_ip,
          });
          if (result) {
            await handleGetNodeData();
            logSideDialogConf.value.isShow && await getTaskStepData(row);
          }
          logSideDialogConf.value.loading = false;
          tableLoading.value = false;
        },
      });
    };
    // 跳过任务
    const { skipTask } = useTask();
    const handleSkip = (row) => {
      if (!row?.step?.allowSkip) {
        $bkMessage({
          theme: 'warning',
          message: $i18n.t('cluster.title.cantSkip'),
        });
        return;
      }
      $bkInfo({
        type: 'warning',
        title: $i18n.t('cluster.title.skipTask'),
        clsName: 'custom-info-confirm default-info',
        subTitle: row.taskName || row.name,
        confirmFn: async () => {
          const result = await skipTask(logSideDialogConf.value.taskID);
          if (result) {
            $bkMessage({
              theme: 'success',
              message: $i18n.t('generic.msg.success.deliveryTask'),
            });
            logSideDialogConf.value.isShow = false;
            handleGetNodeData();
          }
        },
      });
    };
    // 重试按钮显示逻辑
    function handleStepRetry(item) {
      return item?.step?.status === 'FAILED' && item?.step?.allowRetry;
    }
    // 跳过按钮显示逻辑
    function handleStepSkip(item) {
      return item?.step?.status === 'FAILED' && item?.step?.allowSkip;
    }


    // 批量允许调度
    const showBatchMenu = ref(false);
    const handleBatchEnableNodes = () => {
      if (!selections.value.length) return;

      bkComfirmInfo({
        title: $i18n.t('generic.button.uncordon.title2'),
        subTitle: $i18n.t('generic.button.uncordon.subTitle2', {
          ip: selections.value[0].nodeName,
          num: selections.value.length,
        }),
        callback: async () => {
          const result = await handleUncordonNodes({
            clusterID: localClusterId.value,
            nodes: selections.value.map(item => item.nodeName),
          });
          if (result) {
            handleGetNodeData();
            handleResetCheckStatus();
          }
        },
      });
    };
    // 批量停止调度
    const handleBatchStopNodes = () => {
      if (!selections.value.length) return;

      bkComfirmInfo({
        title: $i18n.t('generic.button.cordon.title1'),
        subTitle: $i18n.t('generic.button.cordon.subTitle2', {
          ip: selections.value[0].nodeName,
          num: selections.value.length,
        }),
        callback: async () => {
          const result = await handleCordonNodes({
            clusterID: localClusterId.value,
            nodes: selections.value.map(item => item.nodeName),
          });
          if (result) {
            handleGetNodeData();
            handleResetCheckStatus();
          }
        },
      });
    };
    // 重新添加节点
    const handleBatchReAddNodes = () => {
      if (!selections.value.length
      || isKubeConfigOrAgentImportCluster.value
      || selections.value.some(item => !['REMOVE-FAILURE', 'ADD-FAILURE'].includes(item.status))) return;

      bkComfirmInfo({
        title: $i18n.t('cluster.nodeList.title.confirmReAddNode'),
        subTitle: $i18n.t('cluster.nodeList.create.button.confirmAdd.article1', {
          num: selections.value.length,
          ip: selections.value[0].inner_ip }),
        callback: async () => {
          await addClusterNode(localClusterId.value, selections.value.map(item => item.inner_ip));
        },
      });
    };
    // 批量设置标签和污点
    const handleBatchSetNode = (type: MetaType) => {
      if (!selections.value.length) return;

      $router.push({
        name: 'batchSettingNode',
        params: {
          clusterId: props.clusterId,
          type,
        },
        query: {
          nodeNameList: selections.value.map(item => item.nodeName).join(','),
        },
      });
    };
    // 批量删除节点
    const handleBatchDeleteNodes = () => {
      if (disableBatchDelete.value) return;

      curCheckedNodes.value = [...selections.value];
      showDeleteDialog.value = true;
      // bkComfirmInfo({
      //   title: $i18n.t('cluster.ca.nodePool.nodes.action.delete.title'),
      //   subTitle: $i18n.t('cluster.nodeList.button.delete.subTitle', {
      //     num: selections.value.length,
      //     ip: selections.value[0].innerIP || selections.value[0].nodeID,
      //   }),
      //   callback: async () => {
      //     await delNode(localClusterId.value, selections.value);
      //   },
      // });
    };
    // 批量Pod驱逐
    const handleBatchPodScheduler = () => {
      if (!selections.value.length) return;

      if (selections.value.length > 10) {
        $bkMessage({
          theme: 'warning',
          message: $i18n.t('cluster.nodeList.validate.max10NodeDrain'),
        });
        return;
      }
      // 改为 侧栏
      podConfig.value.isShow = true;
      podConfig.value.nodes = selections.value;
    };
    // 添加节点
    const handleAddNode = () => {
      $router.push({
        name: 'addClusterNode',
        params: {
          clusterId: props.clusterId,
        },
      });
    };
    // 查看日志
    const logSideDialogConf = ref<{
      isShow: boolean;
      title: string;
      taskID: string;
      taskData: any[];
      row: any;
      loading: boolean;
      status: string;
    }>({
      isShow: false,
      title: '',
      taskID: '', // 最新任务ID
      taskData: [],
      row: null, // 当前节点
      loading: false,
      status: '',
    });
    const handleShowLog = async (row) => {
      logSideDialogConf.value.isShow = true;
      logSideDialogConf.value.title = row.nodeName || row.innerIP;
      logSideDialogConf.value.row = row;
      logSideDialogConf.value.loading = true;
      await getTaskStepData(row);
      logSideDialogConf.value.loading = false;
    };

    const getTaskStepData = async (row) => {
      let taskID = '';
      if (row.taskID) {
        taskID = row.taskID;
      } else {
        const { latestTask } = await getTaskData({
          clusterId: row.cluster_id,
          nodeIP: row.inner_ip,
        });
        taskID = latestTask?.taskID || '';
      }
      if (!taskID) {
        logSideDialogConf.value.taskData = [];
        logSideDialogConf.value.status = '';
        return;
      }
      const { status, step } = await clusterTaskRecords({
        taskID,
      }).catch(() => ({ status: '', step: [] }));
      logSideDialogConf.value.taskData = step;
      logSideDialogConf.value.status = status;
      logSideDialogConf.value.taskID = taskID;
    };

    const closeLog = () => {
      logSideDialogConf.value.row = null;
      autoRefresh.stop();
    };

    // 获取节点指标
    const nodeMetric = ref<NodeMetricType>({});
    // 格式化分子和分母 & 重新计算使用率
    const formatMetricData = (data: IMetricData) => {
      const newData = data;
      metricColumnConfig.value.forEach((item) => {
        const [prop1, prop2] = item.percent || [];
        if (!prop1 || !prop2) return;

        const numerator = data[prop1];
        const denominator = data[prop2];
        if (!numerator || !denominator) {
          // 分支或分母不存在时设置使用率为空
          newData[item.prop] = '';
        } else {
          let usedOfTotal = '';
          switch (item.unit) {
            case 'byte':
              usedOfTotal = `${formatBytes(numerator)} / ${formatBytes(denominator)}`;
              break;
            case 'int':
              usedOfTotal = `${Math.ceil(numerator)} ${$i18n.t('units.suffix.units')} / ${Math.ceil(denominator)} ${$i18n.t('units.suffix.units')}`;
              break;
            case 'cpu':
              usedOfTotal = `${Number(numerator).toFixed(2)} ${$i18n.t('units.suffix.cores')} / ${Number(denominator).toFixed(2)} ${$i18n.t('units.suffix.cores')}`;
              break;
            default:
              usedOfTotal = `${numerator} / ${denominator}`;
          }
          newData[`${item.prop}_tips`] = usedOfTotal;// hack 悬浮tips展示的分子和分母
          newData[item.prop] = ((numerator / denominator) * 100).toFixed(2);// 重新计算百分比(接口不准确)
        }
      });
      return newData;
    };
    const overviewsLoading = ref(false);
    const handleGetNodeOverview = async () => {
      const data = curPageData.value.filter(item => !nodeMetric.value[item.nodeName]
        && ['RUNNING', 'REMOVABLE'].includes(item.status));
      const nodes = data.map(item => item.nodeName) as string[];
      overviewsLoading.value = true;
      const result = await getAllNodeOverview({
        clusterId: localClusterId.value,
        nodes,
      }).catch(() => {});
      Object.keys(result).forEach((key) => {
        set(nodeMetric.value, key, formatMetricData(result[key]));
      });
      overviewsLoading.value = false;
    };
    watch(curPageData, async () => {
      await handleGetNodeOverview();
    });
    // 切换集群
    const handleGetNodeData = async () => {
      handleResetCheckStatus();
      tableLoading.value = true;
      tableData.value = await getNodeList(localClusterId.value);
      tableLoading.value = false;
    };

    const handleClusterChange = async () => {
      stop();
      await handleGetNodeData();
      handleResetPage();
      handleResetCheckStatus();
      if (tableData.value.length) {
        start();
      }
    };
    const podDisabled = computed(() => !selections.value.every(select => ['REMOVABLE', 'REMOVE-FAILURE', 'REMOVE-CA-FAILURE'].includes(select.status)));

    // 是否显示表头alert的后半段
    const isShowTableAlertAfter = computed(() => !['awsCloud', 'azureCloud', 'gcpCloud'].includes(clusterData.value.provider));

    watch(pageConf, () => {
      // 非跨页全选在分页变更时重置selections
      if (![
        CheckType.AcrossChecked,
        CheckType.HalfAcrossChecked,
      ].includes(selectType.value)) {
        handleResetCheckStatus();
      }
    });

    const { stop, start } = useInterval(async () => {
      const data = await getNodeList(localClusterId.value);
      // todo 解决轮询导致指令不断更新，popover消失问题
      if (JSON.stringify(data) !== JSON.stringify(tableData.value)) {
        tableData.value = data;
      }
    }, 5000);

    // 自动刷新
    const autoRefresh = useInterval(async () => {
      const row = logSideDialogConf.value.row as any;
      if (!row) {
        autoRefresh.stop();
        return;
      }
      await getTaskStepData(row);
    }, 5000);
    function handleAutoRefresh(v: boolean) {
      if (v) {
        autoRefresh.start();
      } else {
        autoRefresh.stop();
      }
    };

    const timeRange = ref<Date[]>([]);
    // 下载集群操作日志
    async function getDownloadTaskRecords() {
      if (!props.clusterId) return;
      const { url } = parseUrl('get', taskLogsDownloadURL, {
        startTime: Math.ceil(new Date(timeRange.value[0]).getTime() / 1000),
        endTime: Math.ceil(new Date(timeRange.value[1]).getTime() / 1000),
        clusterID: props.clusterId,
        limit: pagination.value.limit,
        page: pagination.value.current,
        v2: true,
      });
      url && window.open(url);
    }

    onBeforeMount(() => {
      // 初始化默认时间
      const end = new Date();
      const start = new Date();
      start.setTime(start.getTime() - 3600 * 1000 * 24 * 7);
      timeRange.value = [
        start,
        end,
      ];
    });

    onMounted(async () => {
      getClusterDetail(curSelectedCluster.value.clusterID || '', true);
      await handleGetNodeData();
      if (tableData.value.length) {
        start();
      }
    });
    onActivated(() => {
      handleGetNodeData();
      if (tableData.value.length) {
        start();
      }
    });
    onBeforeUnmount(() => {
      autoRefresh.stop();
      stop();
    });
    return {
      copyList,
      metricColumnConfig,
      curSelectedCluster,
      clusterData, // 全量数据
      logSideDialogConf,
      searchSelectData,
      searchSelectValue,
      tableKey,
      filtersDataSource,
      filteredValue,
      selectType,
      selections,
      pagination,
      curPageData,
      nodeStatusColorMap,
      nodeStatusMap,
      tableSetting,
      taintConfig,
      setMetaConf,
      podConfig,
      tableLoading,
      localClusterId,
      CheckType,
      nodeMetric,
      renderSelection,
      pageChange,
      pageSizeChange,
      isColumnRender,
      handleSettingChange,
      handleGetNodeData,
      handleSelectionAll,
      handleClearSelection,
      handleRowCheckChange,
      handleFilterChange,
      searchSelectChange,
      handleClearSearchSelect,
      handleGoOverview,
      handleCopy,
      handleMetadata,
      sortMetricMethod,
      handleLabelEditCancel,
      handleLabelEditConfirm,
      handleConfirmTaintDialog,
      handleHideTaintDialog,
      handleSetTaint,
      handleEnableNode,
      handleStopNode,
      handleDeleteNode,
      handleSchedulerNode,
      handleRetry,
      handleBatchEnableNodes,
      handleBatchStopNodes,
      handleBatchReAddNodes,
      handleBatchSetNode,
      handleBatchDeleteNodes,
      handleAddNode,
      handleClusterChange,
      handleShowLog,
      closeLog,
      handleBatchPodScheduler,
      podDisabled,
      webAnnotations,
      curProject,
      isKubeConfigOrAgentImportCluster,
      KEY_REGEXP,
      VALUE_REGEXP,
      showBatchMenu,
      showCopyMenu,
      setChanged,
      handleBeforeClose,
      disableBatchDelete,
      disableBatchDeleteTips,
      isEditModule,
      curModuleID,
      handleEditWorkerModule,
      handleWorkerModuleChange,
      handleSaveWorkerModule,
      handleNodeChange,
      showDeleteDialog,
      curCheckedNodes,
      handleHideDeleteDialog,
      delNode,
      deleteMode,
      deleting,
      isCloudSelfNode,
      disableAddNodeBtn,
      isGkeManagedCluster,
      isShowTableAlertAfter,
      handleAutoRefresh,
      getDownloadTaskRecords,
      handleSkip,
      handleStepRetry,
      handleStepSkip,
      handleHidePodDrain,
      handleSuccess,
      overviewsLoading,
      getDisableDeleteTips,
      metaType,
    };
  },
});
</script>
<style lang="postcss" scoped>
.cluster-node-tip {
    margin-bottom: 20px;
    .num {
        font-weight: 700;
    }
}
.cluster-node-operate {
    display: flex;
    align-items: center;
    justify-content: space-between;
    .left {
        display: flex;
        align-items: center;
        .add-node {
            min-width: 120px;
        }
    }
    .right {
        display: flex;
    }
}
/deep/ .bk-dropdown-list {
    min-width: 100px;
    max-height: unset;
    li {
        display: block;
        height: 32px;
        line-height: 33px;
        padding: 0 16px;
        color: #63656e;
        font-size: 14px;
        cursor: pointer;
        white-space: nowrap;
        &:hover {
            background-color: #eaf3ff;
            color: #3a84ff;
        }
        &[disabled] {
            pointer-events: none;
            color: #c3cdd7;
            cursor: not-allowed;
        }
    }
}
/deep/ .bk-table-column-setting {
    .bk-tooltip-ref {
        display: flex;
        align-items: center;
        justify-content: center;
    }
}
.tips-enter-active {
  transition: opacity .5s;
}
.tips-enter,
.tips-leave-to {
  opacity: 0;
}
.selection-tips {
    height: 30px;
    background: #ebecf0;
    display: flex;
    align-items: center;
    justify-content: center;
    .tips-num {
        font-weight: bold;
    }
    .tips-btn {
        font-size: 12px;
        margin-left: 5px;
    }
}
.row-label {
    display: flex;
    align-items: center;
    flex-wrap: wrap;
    height: 22px;
    .label {
        display: inline-block;
        align-self: center;
        background: #f0f1f5;
        border-radius: 2px;
        line-height: 22px;
        padding: 0 8px;
        margin-right: 6px;
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap
    }
}
.sideslider-tips {
    color: #c3cdd7;
    font-size: 12px;
    font-weight: normal;
    margin-left: 10px;
}
.key-value-content {
    padding: 30px;
}
.log-wrapper {
    padding: 20px;
}
.labels-tips {
    max-height: 260px;
    overflow: auto;
}
.popover {
    width: 100%;
    /deep/ .bk-tooltip-ref {
        display: block;
    }
}
/deep/ .bk-dropdown-menu.disabled > div:nth-child(2) {
  background-color: #f5f7fa !important;
}
/deep/ .from-cluster.bk-dropdown-menu.disabled > div:nth-child(2) {
  background-color: #fff !important;
}
</style>
