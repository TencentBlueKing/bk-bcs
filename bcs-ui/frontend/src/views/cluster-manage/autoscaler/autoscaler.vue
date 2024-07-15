<!-- eslint-disable max-len -->
<template>
  <div class="autoscaler-management">
    <!-- 自动扩缩容配置 -->
    <section class="autoscaler" ref="autoscalerRef">
      <div class="group-header">
        <div>
          <span class="group-header-title">{{$t('cluster.ca.title.caConfig')}}</span>
          <span class="switch-autoscaler">
            {{$t('cluster.ca.name')}}
            <bcs-switcher
              size="small"
              v-model="autoscalerData.enableAutoscale"
              :disabled="autoscalerData.status === 'UPDATING'"
              :pre-check="handleToggleAutoScaler"
            ></bcs-switcher>
          </span>
        </div>
        <bcs-button
          theme="primary"
          :disabled="autoscalerData.status === 'UPDATING'"
          @click="handleEditAutoScaler">{{$t('cluster.ca.button.edit')}}</bcs-button>
      </div>
      <div
        :class="[
          'grid gap-[16px]',
          `grid-cols-${cols}`
        ]"
        v-bkloading="{ isLoading: configLoading }">
        <LayoutGroup :title="$t('cluster.ca.basic.title')" class="mb10">
          <AutoScalerFormItem
            :list="basicScalerConfig"
            :autoscaler-data="autoscalerData"
            width="100%">
          </AutoScalerFormItem>
        </LayoutGroup>
        <LayoutGroup :title="$t('cluster.ca.unreadyConfig.title')" class="mb10">
          <i18n
            path="cluster.ca.unreadyConfig.path"
            class="text-[#979BA5] leading-[32px]">
            <span place="0" class="text-[#313238]">{{ autoscalerData.okTotalUnreadyCount }}</span>
            <span place="1" class="text-[#313238]">{{ autoscalerData.maxTotalUnreadyPercentage }}%</span>
          </i18n>
          <span
            class="ml-[5px] text-[16px] text-[#979ba5]"
            v-bk-tooltips="$t('cluster.ca.unreadyConfig.desc')">
            <i class="bk-icon icon-info-circle"></i>
          </span>
        </LayoutGroup>
        <LayoutGroup :title="$t('cluster.ca.autoScalerConfig.title')" class="mb10">
          <AutoScalerFormItem
            :list="autoScalerConfig"
            :autoscaler-data="autoscalerData"
            width="100%">
            <template #suffix="{ data }">
              <span
                class="text-[#699DF4] ml-[5px] h-[20px] flex items-center"
                style="border-bottom: 1px dashed #699DF4;"
                v-if="['bufferResourceCpuRatio', 'bufferResourceMemRatio'].includes(data.prop)"
                v-bk-tooltips="data.prop === 'bufferResourceCpuRatio'
                  ? `${Number((overview.cpu_usage.request || 0)).toFixed(2)}${$t('units.suffix.cores')} / ${Number(Math.ceil(overview.cpu_usage.total) || 0).toFixed(2)}${$t('units.suffix.cores')}`
                  : `${formatBytes(overview.memory_usage.request_bytes || 0, 2)} / ${formatBytes(overview.memory_usage.total_bytes || 0, 2)}`">
                {{
                  $t('cluster.ca.metric.curUsagePath', {
                    val: data.prop === 'bufferResourceCpuRatio'
                      ? conversionPercentUsed(overview.cpu_usage.request, overview.cpu_usage.total)
                      : conversionPercentUsed(overview.memory_usage.request_bytes, overview.memory_usage.total_bytes)
                  })
                }}
              </span>
              <span
                class="text-[#699DF4] ml-[5px] h-[20px] flex items-center"
                style="border-bottom: 1px dashed #699DF4;"
                v-else-if="data.prop === 'bufferResourceRatio'"
                v-bk-tooltips="overview.pod_usage
                  ?`${Number((overview.pod_usage.used || 0))} / ${Number(Math.ceil(overview.pod_usage.total) || 0)}`
                  : '--'">
                {{
                  $t('cluster.ca.metric.curUsagePath', {
                    val: overview.pod_usage
                      ?conversionPercentUsed(overview.pod_usage.used, overview.pod_usage.total)
                      : '--'
                  })
                }}
              </span>
            </template>
          </AutoScalerFormItem>
        </LayoutGroup>
        <LayoutGroup collapsible class="mb15" :expanded="!!autoscalerData.isScaleDownEnable">
          <template #title>
            <span class="flex items-center">
              <span>{{$t('cluster.ca.autoScalerDownConfig.title')}}</span>
              <span class="flex items-center ml-[8px]" @click.stop>
                <span
                  :class="['px-[8px] inline-block leading-[20px] text-[#979BA5]', {
                    '!text-[#2DCB56] bg-[#F2FFF4]': autoscalerData.isScaleDownEnable && autoscalerData.enableAutoscale
                  }]">
                  {{ autoscalerData.isScaleDownEnable && autoscalerData.enableAutoscale ? $t('cluster.ca.status.on') : $t('cluster.ca.status.off') }}
                </span>
                <bcs-divider direction="vertical" class="!mr-[10px]"></bcs-divider>
                <span
                  v-bk-tooltips="{
                    disabled: autoscalerData.enableAutoscale,
                    content: $t('cluster.ca.tips.cannotEnableScalerDownConfig')
                  }">
                  <bk-button
                    text
                    class="text-[12px]"
                    :disabled="autoscalerData.status === 'UPDATING' || !autoscalerData.enableAutoscale"
                    @click="handleChangeScalerDown">
                    {{ autoscalerData.isScaleDownEnable && autoscalerData.enableAutoscale ? $t('generic.button.close') : $t('generic.button.on') }}
                  </bk-button>
                </span>
              </span>
            </span>
          </template>
          <AutoScalerFormItem
            :list="autoScalerDownConfig"
            :autoscaler-data="autoscalerData"
            width="100%">
          </AutoScalerFormItem>
        </LayoutGroup>
        <LayoutGroup collapsible class="mb15" :expanded="isPodsPriorityEnable">
          <template #title>
            <span class="flex items-center">
              <span>{{$t('cluster.ca.podsPriorityConfig.title')}}</span>
              <span class="flex items-center ml-[8px]" @click.stop>
                <span
                  :class="['px-[8px] inline-block leading-[20px] text-[#979BA5]', {
                    '!text-[#2DCB56] bg-[#F2FFF4]': isPodsPriorityEnable && autoscalerData.enableAutoscale
                  }]">
                  {{ isPodsPriorityEnable && autoscalerData.enableAutoscale ? $t('cluster.ca.status.on') : $t('cluster.ca.status.off') }}
                </span>
                <bcs-divider direction="vertical" class="!mr-[10px]"></bcs-divider>
                <span
                  v-bk-tooltips="{
                    disabled: autoscalerData.enableAutoscale,
                    content: $t('cluster.ca.tips.cannotEnablePodsPriorityConfig')
                  }">
                  <bk-button
                    text
                    class="text-[12px]"
                    :disabled="autoscalerData.status === 'UPDATING' || !autoscalerData.enableAutoscale"
                    @click="handleTogglePodsPriorityDialog">
                    {{ isPodsPriorityEnable && autoscalerData.enableAutoscale ? $t('generic.button.close') : $t('generic.button.on') }}
                  </bk-button>
                </span>
              </span>
            </span>
          </template>
          <AutoScalerFormItem
            :list="[{
              prop: 'expendablePodsPriorityCutoff',
              name: $t('cluster.ca.podsPriorityConfig.expendablePodsPriorityCutoff.title'),
              desc: $t('cluster.ca.podsPriorityConfig.expendablePodsPriorityCutoff.desc'),
            }]"
            :autoscaler-data="autoscalerData" />
        </LayoutGroup>
      </div>
    </section>
    <!-- 资源池配置 -->
    <section class="group-border-top" v-if="curCluster.provider === 'tencentCloud'">
      <div class="group-header">
        <div class="group-header-title">{{$t('tkeCa.label.poolManage')}}</div>
      </div>
      <bk-form class="bcs-form-content px-[24px] pb-[10px]" :label-width="210">
        <bk-form-item :label="$t('tkeCa.label.provider.text')" :desc="$t('tkeCa.label.provider.desc')">
          <template v-if="!isEditDevicePool">
            <span class="break-all">
              {{ (autoscalerData.devicePoolProvider || 'yunti') === 'yunti'
                ? $t('tkeCa.label.provider.yunti') : $t('tkeCa.label.provider.self') }}
            </span>
            <span
              class="hover:text-[#3a84ff] cursor-pointer ml-[8px]"
              @click="isEditDevicePool = true">
              <i class="bk-icon icon-edit-line"></i>
            </span>
          </template>
          <template v-else>
            <bcs-select
              :clearable="false"
              searchable
              class="w-[200px]"
              :value="autoscalerData.devicePoolProvider || 'yunti'"
              @change="handleChangeDevicePool">
              <bcs-option id="yunti" :name="$t('tkeCa.label.provider.yunti')"></bcs-option>
              <bcs-option
                id="self"
                :name="$t('tkeCa.label.provider.self')"
                :disabled="disableSelfDevicePool"
                v-bk-tooltips="{
                  content: $t('tkeCa.tips.notSupportSelfPool'),
                  disabled: !disableSelfDevicePool,
                  placement: 'left'
                }">
              </bcs-option>
            </bcs-select>
            <bcs-button
              text
              class="text-[12px] ml-[10px] h-[32px]"
              @click="handleSaveDevicePoolChange">{{ $t('generic.button.save') }}</bcs-button>
            <bcs-button
              text
              class="text-[12px] ml-[10px] h-[32px]"
              @click="isEditDevicePool = false">{{ $t('generic.button.cancel') }}</bcs-button>
          </template>
        </bk-form-item>
      </bk-form>
    </section>
    <!-- 节点规格配置 -->
    <section class="group-border-top">
      <div class="group-header">
        <div class="group-header-title">{{$t('cluster.ca.title.nodePoolManage')}}</div>
        <div class="flex">
          <bcs-button theme="primary" icon="plus" @click="handleCreatePool">{{$t('cluster.ca.button.createNodePool')}}</bcs-button>
          <bcs-button class="ml5" @click="handleShowRecord({})">{{$t('cluster.ca.button.record')}}</bcs-button>
        </div>
      </div>
      <bcs-table
        :data="curPageData"
        :pagination="pagination"
        v-bkloading="{ isLoading: nodepoolLoading }"
        @page-change="pageChange"
        @page-limit-change="pageSizeChange">
        <bcs-table-column :label="$t('cluster.ca.nodePool.label.nameAndID')" min-width="200" show-overflow-tooltip>
          <template #default="{ row }">
            <div class="bk-primary bk-button-normal bk-button-text" @click="handleGotoDetail(row)">
              <span class="bcs-ellipsis">{{`${row.name}（${row.nodeGroupID}）`}}</span>
            </div>
          </template>
        </bcs-table-column>
        <bcs-table-column :label="$t('cluster.ca.nodePool.label.nodeQuota')" align="right" width="100">
          <template #default="{ row }">
            {{ row.autoScaling.maxSize }}
          </template>
        </bcs-table-column>
        <bcs-table-column :label="$t('cluster.ca.nodePool.label.nodeCounts')" align="right" width="100">
          <template #default="{ row }">
            <bcs-button
              text
              :disabled="row.autoScaling.desiredSize === 0"
              @click="handleShowNodeManage(row)">
              <div class="min-w-[80px] text-right">{{row.autoScaling.desiredSize}}</div>
            </bcs-button>
          </template>
        </bcs-table-column>
        <bcs-table-column :label="$t('generic.ipSelector.label.serverModel')">
          <template #default="{ row }">
            {{ row.launchTemplate.instanceType }}
          </template>
        </bcs-table-column>
        <bcs-table-column :label="$t('cluster.ca.nodePool.label.system')" show-overflow-tooltip>
          <template #default="{ row }">{{ row.nodeOS || '' }}</template>
        </bcs-table-column>
        <bcs-table-column :label="$t('tkeCa.label.provider.text')" show-overflow-tooltip v-if="curCluster.provider === 'tencentCloud'">
          <template #default="{ row }">
            {{ (row.extraInfo && row.extraInfo.resourcePoolType ? row.extraInfo.resourcePoolType : 'yunti') === 'yunti'
              ? $t('tkeCa.label.provider.yunti')
              : $t('tkeCa.label.provider.self') }}
          </template>
        </bcs-table-column>
        <bcs-table-column :label="$t('cluster.ca.nodePool.label.status')">
          <template #default="{ row }">
            <LoadingIcon v-if="['CREATING', 'DELETING', 'UPDATING'].includes(row.status)">
              {{ statusTextMap[row.status] }}
            </LoadingIcon>
            <StatusIcon status="unknown" v-else-if="!row.enableAutoscale && row.status === 'RUNNING'">
              {{$t('cluster.ca.status.off')}}
            </StatusIcon>
            <StatusIcon
              :status="row.status"
              :status-color-map="statusColorMap"
              v-else>
              {{ statusTextMap[row.status] || $t('generic.status.unknown') }}
            </StatusIcon>
          </template>
        </bcs-table-column>
        <bcs-table-column :label="$t('generic.label.action')" width="170">
          <template #default="{ row }">
            <div class="operate">
              <bcs-button text @click="handleShowRecord(row)">{{$t('cluster.ca.button.record')}}</bcs-button>
              <bcs-popover
                placement="bottom"
                theme="light dropdown"
                :arrow="false"
                class="ml15"
                :disabled="row.status === 'DELETING'"
                trigger="click"
                :ref="row.nodeGroupID">
                <span class="more-icon"><i class="bcs-icon bcs-icon-more"></i></span>
                <div class="bg-[#fff]" slot="content">
                  <ul>
                    <li class="dropdown-item" @click="handleAddNode(row)">{{$t('cluster.nodeList.create.text')}}</li>
                    <li
                      :class="['dropdown-item', {
                        disabled: (row.enableAutoscale && disabledAutoscaler)
                          || ['CREATING', 'DELETING', 'UPDATING'].includes(row.status)
                      }]"
                      v-bk-tooltips="{
                        content: $t('cluster.ca.tips.noEnableNodePool'),
                        disabled: !(row.enableAutoscale && disabledAutoscaler)
                      }"
                      @click="handleToggleNodeScaler(row)">
                      {{row.enableAutoscale ? $t('cluster.ca.nodePool.action.off') : $t('cluster.ca.nodePool.action.on')}}
                    </li>
                    <li
                      class="dropdown-item"
                      @click="handleClonePool(row)">
                      {{$t('cluster.ca.nodePool.action.clone')}}
                    </li>
                    <li class="dropdown-item" @click="handleEditPool(row)">{{$t('cluster.ca.nodePool.action.edit')}}</li>
                    <li
                      :class="['dropdown-item', { disabled: disabledDelete || !!row.autoScaling.desiredSize }]"
                      v-bk-tooltips="{
                        content: !!row.autoScaling.desiredSize
                          ? $t('cluster.ca.tips.notEmptyNodes')
                          : $t('cluster.ca.tips.needMoreThanOneNodePoolOn'),
                        disabled: !(disabledDelete || !!row.autoScaling.desiredSize),
                        placement: 'left'
                      }"
                      @click="handleDeletePool(row)">{{$t('cluster.ca.nodePool.action.delete.text')}}</li>
                  </ul>
                </div>
              </bcs-popover>
            </div>
          </template>
        </bcs-table-column>
      </bcs-table>
    </section>
    <!-- 节点数量 -->
    <bcs-dialog
      theme="primary"
      header-position="left"
      :title="$t('cluster.ca.nodePool.nodes.title')"
      :width="800"
      v-model="showNodeManage"
      @cancel="handleNodeManageCancel">
      <bcs-alert type="info" :title="$t('cluster.ca.nodePool.nodes.desc')"></bcs-alert>
      <bcs-form class="form-content mt15" :label-width="100">
        <bcs-form-item class="form-content-item" :label="$t('cluster.ca.nodePool.label.name')">
          <span>{{ currentOperateRow.name }}</span>
        </bcs-form-item>
        <bcs-form-item class="form-content-item" :label="$t('cluster.ca.nodePool.label.nodeQuota')">
          <span>
            {{
              currentOperateRow.autoScaling
                ? currentOperateRow.autoScaling.maxSize
                : '--'
            }}
          </span>
        </bcs-form-item>
        <bcs-form-item class="form-content-item" :label="$t('cluster.ca.nodePool.label.nodeCounts')">
          <span>
            {{
              currentOperateRow.autoScaling
                ? currentOperateRow.autoScaling.desiredSize
                : '--'
            }}
          </span>
        </bcs-form-item>
      </bcs-form>
      <Row class="mt-[10px]">
        <template #left>
          <bcs-button
            theme="primary"
            icon="plus"
            class="mr10"
            @click="handleAddNode(currentOperateRow)">
            {{$t('cluster.nodeList.create.text')}}
          </bcs-button>
          <bcs-dropdown-menu
            :disabled="!selections.length"
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
            <ul slot="dropdown-content">
              <li class="bcs-dropdown-item" @click="handleBatchEnableNodes">{{$t('generic.button.uncordon.text')}}</li>
              <li class="bcs-dropdown-item" @click="handleBatchStopNodes">{{$t('generic.button.cordon.text')}}</li>
              <li
                :class="['bcs-dropdown-item', { disabled: podDisabled }]"
                v-bk-tooltips="{
                  content: $t('generic.button.drain.tips'),
                  disabled: !podDisabled,
                  placement: 'right'
                }"
                @click="handleBatchPodScheduler">
                {{$t('generic.button.drain.text')}}
              </li>
              <li
                :class="['bcs-dropdown-item', { disabled: disableBatchDelete }]"
                v-bk-tooltips="{
                  content: $t('cluster.ca.nodePool.nodes.action.delete.tips'),
                  disabled: !disableBatchDelete,
                  placement: 'right'
                }"
                @click="handleBatchDeleteNodes">
                {{$t('generic.button.delete')}}
              </li>
            </ul>
          </bcs-dropdown-menu>
        </template>
        <template #right>
          <bcs-input
            right-icon="bk-icon icon-search"
            clearable
            class="w-[360px]"
            :placeholder="$t('generic.placeholder.searchIp')"
            v-model.trim="searchIpData">
          </bcs-input>
        </template>
      </Row>
      <bcs-table
        class="mt15"
        v-bkloading="{ isLoading: nodeListLoading }"
        :max-height="nodeListLoading ? 200 : ''"
        :data="nodeCurPageData"
        :pagination="nodePagination"
        @filter-change="handleNodeFilterChange"
        @page-change="nodePageChange"
        @page-limit-change="nodePageSizeChange">
        <template #prepend>
          <transition name="fade">
            <div class="flex items-center justify-center h-[30px] bg-[#ebecf0]" v-if="selectType !== CheckType.Uncheck">
              <i18n path="cluster.nodeList.msg.selectedData">
                <span place="num" class="font-bold">{{selections.length}}</span>
              </i18n>
              <bk-button
                class="text-[12px] ml-[5px]"
                text
                v-if="selectType === CheckType.AcrossChecked"
                @click="handleClearSelection">
                {{ $t('cluster.nodeList.button.cancelSelectAll') }}
              </bk-button>
              <bk-button
                class="text-[12px] ml-[5px]"
                text
                v-else
                @click="handleSelectionAll">
                <i18n path="cluster.nodeList.msg.selectedAllData">
                  <span place="num" class="font-bold">{{nodePagination.count}}</span>
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
              :checked="selections.some(item => item.innerIP === row.innerIP && item.nodeID === row.nodeID)"
              :disabled="['INITIALIZATION', 'DELETING', 'APPLYING'].includes(row.status)"
              @change="(value) => handleRowCheckChange(value, row)"
            />
          </template>
        </bcs-table-column>
        <bcs-table-column :label="$t('cluster.ca.nodePool.nodes.label.name')" prop="innerIP">
          <template #default="{ row }">
            {{ row.innerIP || '--' }}
          </template>
        </bcs-table-column>
        <bcs-table-column
          :label="$t('generic.label.status')"
          :filters="filtersStatus"
          :filtered-value="filtersStatusValue"
          column-key="status"
          prop="status">
          <template #default="{ row }">
            <LoadingIcon v-if="['DELETING', 'INITIALIZATION', 'APPLYING'].includes(row.status)">
              {{ nodeStatusMap[row.status] }}
            </LoadingIcon>
            <StatusIcon
              :status="row.status"
              :status-color-map="nodeColorMap"
              v-else>
              {{ nodeStatusMap[row.status] }}
            </StatusIcon>
          </template>
        </bcs-table-column>
        <bcs-table-column :label="$t('generic.label.action')" width="120">
          <template #default="{ row }">
            <div class="operate">
              <template v-if="row.status === 'APPLY-FAILURE'">
                <bk-button text @click="handleDeleteNodeGroupNode(row)">
                  {{ $t('cluster.ca.nodePool.nodes.action.delete.text') }}
                </bk-button>
              </template>
              <template v-else-if="row.status !== 'APPLYING'">
                <bcs-button
                  text
                  :disabled="['DELETING', 'INITIALIZATION'].includes(row.status)"
                  @click="handleToggleCordon(row)">
                  {{row.unSchedulable ? $t('generic.button.uncordon.text') : $t('generic.button.cordon.text')}}
                </bcs-button>
                <bcs-popover
                  placement="bottom"
                  theme="light dropdown"
                  :arrow="false"
                  :disabled="['DELETING', 'INITIALIZATION'].includes(row.status)"
                  trigger="click"
                  class="ml15">
                  <span
                    :class="['more-icon', { 'disabled': ['DELETING', 'INITIALIZATION'].includes(row.status) }]">
                    <i class="bcs-icon bcs-icon-more"></i>
                  </span>
                  <div slot="content">
                    <ul>
                      <li class="dropdown-item" @click="handleNodeDrain(row)">{{$t('generic.button.drain.text')}}</li>
                      <li
                        :class="['dropdown-item', { disabled: !row.unSchedulable }]"
                        v-bk-tooltips="{
                          content: $t('cluster.ca.nodePool.nodes.action.delete.tips'),
                          disabled: row.unSchedulable,
                          placement: 'left'
                        }"
                        @click="handleDeleteNodeGroupNode(row)"
                      >{{$t('cluster.ca.nodePool.nodes.action.delete.text')}}</li>
                    </ul>
                  </div>
                </bcs-popover>
              </template>
            </div>
          </template>
        </bcs-table-column>
      </bcs-table>
    </bcs-dialog>
    <!-- 扩缩容记录 -->
    <bcs-dialog
      theme="primary"
      header-position="left"
      :title="$t('cluster.ca.button.record')"
      :width="1200"
      v-model="showRecord"
      @cancel="handleRecordCancel">
      <div class="mb15 flex-between">
        <bcs-date-picker
          :shortcuts="shortcuts"
          type="datetimerange"
          shortcut-close
          :use-shortcut-text="false"
          :clearable="false"
          v-model="timeRange"
          @change="handleTimeRangeChange">
        </bcs-date-picker>
        <bcs-input
          class="w-[360px]"
          right-icon="bk-icon icon-search"
          clearable
          :placeholder="$t('generic.ipSelector.placeholder.searchIp')"
          v-model="searchIp">
        </bcs-input>
      </div>
      <bcs-table
        v-bkloading="{ isLoading: recordLoading }"
        :data="recordList"
        :pagination="recordPagination"
        row-key="taskID"
        :expand-row-keys="expandRowKeys"
        :key="currentOperateRow.nodeGroupID"
        @page-change="recordPageChange"
        @page-limit-change="recordPageSizeChange"
        @expand-change="handleExpandChange"
        @filter-change="handleFilterChange">
        <bcs-table-column type="expand" width="30">
          <template #default="{ row }">
            <bcs-table
              :data="row.task ? row.task.stepSequence : []"
              :outer-border="false"
              :header-cell-style="{ background: '#fff', borderRight: 'none' }">
              <bcs-table-column :label="$t('cluster.ca.nodePool.records.label.stepName')" width="150" show-overflow-tooltip>
                <template #default="{ row: key }">
                  <div class="flex items-center">
                    <span class="bcs-ellipsis">{{ row.task.steps[key].taskName }}</span>
                    <i
                      class="bcs-icon bcs-icon-fenxiang bcs-icon-btn text-[#3a84ff] flex-1 ml-[4px]"
                      v-if="row.task.steps[key].params && row.task.steps[key].params.taskUrl"
                      @click="handleGotoSops(row.task.steps[key].params.taskUrl)"></i>
                  </div>
                </template>
              </bcs-table-column>
              <bcs-table-column :label="$t('cluster.ca.nodePool.records.label.stepMsg')">
                <template #default="{ row: key }">
                  <bcs-popover
                    :disabled="!row.task.steps[key].message || taskStatusColorMap[row.task.steps[key].status] === 'green'">
                    <span class="select-all bcs-ellipsis">{{ row.task.steps[key].message || '--' }}</span>
                    <template #content>
                      {{ row.task.steps[key].message }}
                    </template>
                  </bcs-popover>
                </template>
              </bcs-table-column>
              <bcs-table-column :label="$t('cluster.ca.nodePool.records.label.startTime')" width="180" show-overflow-tooltip>
                <template #default="{ row: key }">
                  {{ row.task.steps[key].start }}
                </template>
              </bcs-table-column>
              <bcs-table-column :label="$t('cluster.ca.nodePool.records.label.endTime')" width="180" show-overflow-tooltip>
                <template #default="{ row: key }">
                  {{ row.task.steps[key].end }}
                </template>
              </bcs-table-column>
              <!-- 设置宽度为了保持和外面表格对齐 -->
              <bcs-table-column :label="$t('generic.label.status')" width="240">
                <template #default="{ row: key }">
                  <LoadingIcon v-if="['RUNNING'].includes(row.task.steps[key].status)">
                    {{ taskStatusMap[row.task.steps[key].status] }}
                  </LoadingIcon>
                  <StatusIcon
                    :status-color-map="taskStatusColorMap"
                    :status="row.task.steps[key].status"
                    v-else>
                    {{ taskStatusMap[row.task.steps[key].status] }}
                  </StatusIcon>
                </template>
              </bcs-table-column>
            </bcs-table>
          </template>
        </bcs-table-column>
        <bcs-table-column
          :label="$t('cluster.ca.nodePool.records.label.eventType')"
          width="150"
          prop="taskType"
          column-key="taskType"
          :filters="filters.taskType"
          :filter-multiple="false"
          filter-searchable
          show-overflow-tooltip>
          <template #default="{ row }">
            {{row.task ? row.task.taskName : '--'}}
          </template>
        </bcs-table-column>
        <bcs-table-column :label="$t('cluster.ca.nodePool.records.label.eventMsg')" prop="message">
          <template #default="{ row }">
            <bcs-popover>
              <span class="select-all bcs-ellipsis">{{ row.message || '--' }}</span>
              <template #content>
                {{ row.message }}
              </template>
            </bcs-popover>
          </template>
        </bcs-table-column>
        <bcs-table-column
          :label="$t('cluster.ca.nodePool.text')"
          prop="resourceID"
          :filtered-value="filterValues.resourceID"
          :filters="filters.resourceID"
          :filter-multiple="false"
          filter-searchable
          column-key="resourceID"
          :key="JSON.stringify(filterValues.resourceID)"
          show-overflow-tooltip>
        </bcs-table-column>
        <bcs-table-column :label="$t('cluster.ca.nodePool.records.label.startTime')" width="180" prop="createTime" show-overflow-tooltip></bcs-table-column>
        <bcs-table-column :label="$t('cluster.ca.nodePool.records.label.endTime')" width="180" show-overflow-tooltip>
          <template #default="{ row }">
            {{row.task ? row.task.end : '--'}}
          </template>
        </bcs-table-column>
        <bcs-table-column
          :label="$t('generic.label.status')"
          width="150"
          prop="status"
          column-key="status"
          filter-searchable
          :filter-multiple="false"
          :filters="filters.status">
          <template #default="{ row }">
            <template v-if="row.task">
              <LoadingIcon v-if="['RUNNING'].includes(row.task.status)">
                {{ taskStatusMap[row.task.status] }}
              </LoadingIcon>
              <StatusIcon
                :status-color-map="taskStatusColorMap"
                :status="row.task.status"
                v-else>
                {{ taskStatusMap[row.task.status] }}
              </StatusIcon>
            </template>
            <span v-else>--</span>
          </template>
        </bcs-table-column>
        <bcs-table-column :label="$t('generic.label.action')" width="120">
          <template #default="{ row }">
            <!-- <bcs-button
              text
              :disabled="!(row.task && taskStatusColorMap[row.task.status] === 'red')"
              @click="handleRetryTask(row)">{{$t('cluster.ca.nodePool.records.action.retry')}}</bcs-button> -->
            <bcs-button
              text
              :disabled="!(row.task && row.task.nodeIPList && row.task.nodeIPList.length)"
              @click="handleShowIPList(row)">{{$t('cluster.ca.nodePool.records.action.ipList')}}</bcs-button>
          </template>
        </bcs-table-column>
      </bcs-table>
    </bcs-dialog>
    <!-- IP列表 -->
    <bcs-dialog
      theme="primary"
      header-position="left"
      :title="$t('cluster.ca.nodePool.records.action.ipList')"
      v-model="showIPList">
      <div class="relative">
        <bk-button text size="small" class="absolute z-10 right-[10px] top-[8px]" @click="handleCopyAllIP">
          <i class="bcs-icon bcs-icon-copy"></i>
          {{ $t('generic.button.copyAll') }}
        </bk-button>
        <bcs-table
          :key="ipTableKey"
          :data="currentOperateRow.task ? currentOperateRow.task.nodeIPList : []"
          :max-height="600">
          <bcs-table-column label="IP">
            <template #default="{ row }">{{ row }}</template>
          </bcs-table-column>
        </bcs-table>
      </div>
    </bcs-dialog>
    <!-- 低优先级Pod配置 -->
    <bcs-dialog
      theme="primary"
      header-position="left"
      :title="$t('cluster.ca.podsPriorityConfig.expendablePodsPriorityCutoff.on')"
      :width="480"
      :loading="podsPriorityLoading"
      v-model="showPodsPriorityDialog"
      @confirm="handleSetPodsPriority">
      <div class="flex items-start">
        <i class="bk-icon icon-info-circle mr-[8px] relative top-[2px]"></i>
        <i18n path="cluster.ca.podsPriorityConfig.path" class="text-[12px]">
          <span class="text-[#FF9C01] font-bold">{{ $t('units.op.le') }}</span>
          <span>{{ $t('cluster.ca.podsPriorityConfig.expendablePodsPriorityCutoff.list') }}</span>
        </i18n>
      </div>
      <div class="flex items-center mt-[20px] ml-[22px]">
        <span class="mr-[20px]">{{ $t('cluster.ca.podsPriorityConfig.expendablePodsPriorityCutoff.title') }}</span>
        <bcs-input
          class="w-[100px]"
          type="number"
          :min="-2147483647"
          :max="-1"
          v-model="curPodsPriority">
        </bcs-input>
        <span class="text-[#979BA5] ml-[8px]">(-2147483647 ~ -1)</span>
      </div>
    </bcs-dialog>
  </div>
</template>
<script lang="ts">
import { computed, defineComponent, getCurrentInstance, onBeforeUnmount, onMounted, ref, watch } from 'vue';

import useNode from '../node-list/use-node';

import AutoScalerFormItem from './form-item.vue';

import { updateClusterAutoScalingProviders } from '@/api/modules/cluster-manager';
import { clusterOverview } from '@/api/modules/monitor';
import $bkMessage from '@/common/bkmagic';
import { copyText, formatBytes } from '@/common/util';
import { CheckType } from '@/components/across-check.vue';
import $bkInfo from '@/components/bk-magic-2.0/bk-info';
import Row from '@/components/layout/Row.vue';
import LoadingIcon from '@/components/loading-icon.vue';
import StatusIcon from '@/components/status-icon';
import { ICluster, useConfig, useProject } from '@/composables/use-app';
import useAutoCols from '@/composables/use-auto-cols';
import useDebouncedRef from '@/composables/use-debounce';
import useInterval from '@/composables/use-interval';
import usePage from '@/composables/use-page';
import useTableAcrossCheck from '@/composables/use-table-across-check';
import $i18n from '@/i18n/i18n-setup';
import $router from '@/router';
import $store from '@/store/index';
import { useClusterInfo } from '@/views/cluster-manage/cluster/use-cluster';
import LayoutGroup from '@/views/cluster-manage/components/layout-group.vue';

export default defineComponent({
  name: 'AutoScaler',
  components: { StatusIcon, LoadingIcon, LayoutGroup, AutoScalerFormItem, Row },
  props: {
    clusterId: {
      type: String,
      default: '',
    },
  },
  setup(props) {
    const configLoading = ref(false);
    const autoscalerData = ref<Record<string, any>>({});
    const basicScalerConfig = ref([
      {
        prop: 'status',
        name: $i18n.t('generic.label.status'),
      },
      {
        prop: 'scanInterval',
        name: $i18n.t('cluster.ca.basic.scanInterval.label'),
        unit: $i18n.t('units.suffix.seconds'),
      },
      // {
      //   prop: 'scaleOutModuleName',
      //   name: $i18n.t('cluster.ca.basic.module.label'),
      //   desc: $i18n.t('cluster.ca.basic.module.desc'),
      // },
    ]);
    const autoScalerConfig = ref([
      {
        prop: 'expander',
        name: $i18n.t('cluster.ca.autoScalerConfig.expander.title'),
        isBasicProp: true,
        desc: $i18n.t('cluster.ca.autoScalerConfig.expander.desc'),
      },
      {
        prop: 'bufferResourceCpuRatio',
        name: $i18n.t('cluster.ca.autoScalerConfig.bufferResourceCpuRatio.title'),
        isBasicProp: true,
        unit: '%',
        desc: $i18n.t('cluster.ca.autoScalerConfig.bufferResourceCpuRatio.desc'),
      },
      {
        prop: 'bufferResourceMemRatio',
        name: $i18n.t('cluster.ca.autoScalerConfig.bufferResourceMemRatio.title'),
        isBasicProp: true,
        unit: '%',
        desc: $i18n.t('cluster.ca.autoScalerConfig.bufferResourceMemRatio.desc'),
      },
      {
        prop: 'bufferResourceRatio',
        name: $i18n.t('cluster.ca.autoScalerConfig.bufferResourceRatio.title'),
        unit: '%',
        desc: $i18n.t('cluster.ca.autoScalerConfig.bufferResourceRatio.desc'),
      },
      {
        prop: 'maxNodeProvisionTime',
        name: $i18n.t('cluster.ca.autoScalerConfig.maxNodeProvisionTime.title'),
        unit: $i18n.t('units.suffix.seconds'),
        desc: $i18n.t('cluster.ca.autoScalerConfig.maxNodeProvisionTime.desc'),
      },
    ]);
    const autoScalerDownConfig = ref([
      {
        prop: 'scaleDownUtilizationThreahold',
        name: $i18n.t('cluster.ca.autoScalerDownConfig.scaleDownUtilizationThreahold.title'),
        isBasicProp: true,
        unit: '%',
        desc: $i18n.t('cluster.ca.autoScalerDownConfig.scaleDownUtilizationThreahold.desc'),
      },
      {
        prop: 'scaleDownUnneededTime',
        name: $i18n.t('cluster.ca.autoScalerDownConfig.scaleDownUnneededTime.title'),
        isBasicProp: true,
        unit: $i18n.t('units.suffix.seconds'),
        desc: $i18n.t('cluster.ca.autoScalerDownConfig.scaleDownUnneededTime.desc'),
        suffix: $i18n.t('cluster.ca.autoScalerDownConfig.scaleDownUnneededTime.suffix'),
      },
      {
        prop: 'maxGracefulTerminationSec',
        name: $i18n.t('cluster.ca.autoScalerDownConfig.maxGracefulTerminationSec.title'),
        isBasicProp: true,
        unit: $i18n.t('units.suffix.seconds'),
        desc: $i18n.t('cluster.ca.autoScalerDownConfig.maxGracefulTerminationSec.desc'),
      },
      {
        prop: 'scaleDownDelayAfterAdd',
        name: $i18n.t('cluster.ca.autoScalerDownConfig.scaleDownDelayAfterAdd.title'),
        unit: $i18n.t('units.suffix.seconds'),
        desc: $i18n.t('cluster.ca.autoScalerDownConfig.scaleDownDelayAfterAdd.desc'),
      },
      {
        prop: 'scaleDownDelayAfterDelete',
        name: $i18n.t('cluster.ca.autoScalerDownConfig.scaleDownDelayAfterDelete.title'),
        unit: $i18n.t('units.suffix.seconds'),
        desc: $i18n.t('cluster.ca.autoScalerDownConfig.scaleDownDelayAfterDelete.desc'),
      },
      {
        prop: 'scaleDownUnreadyTime',
        name: $i18n.t('cluster.ca.autoScalerDownConfig.scaleDownUnreadyTime.title'),
        unit: $i18n.t('units.suffix.seconds'),
      },
      {
        prop: 'maxEmptyBulkDelete',
        name: $i18n.t('cluster.ca.autoScalerDownConfig.maxEmptyBulkDelete.title'),
        unit: $i18n.t('units.suffix.units'),
      },
      {
        prop: 'skipNodesWithLocalStorage',
        name: $i18n.t('cluster.ca.autoScalerDownConfig.skipNodesWithLocalStorage.title'),
        invert: true,
        desc: $i18n.t('cluster.ca.autoScalerDownConfig.skipNodesWithLocalStorage.desc'),
      },
    ]);
    const { _INTERNAL_ } = useConfig();
    const getAutoScalerConfig = async () => {
      if (!props.clusterId) return;
      autoscalerData.value = await $store.dispatch('clustermanager/clusterAutoScaling', {
        $clusterId: props.clusterId,
        provider: _INTERNAL_.value ? 'selfProvisionCloud' : '',
      });
      // 设置模块名称 -- 方便展示（会污染数据）
      autoscalerData.value.scaleOutModuleName = autoscalerData.value.module?.scaleOutModuleName || '--';
      if (autoscalerData.value.status !== 'UPDATING') {
        stop();
      }
    };
    const handleGetAutoScalerConfig = async () => {
      configLoading.value = true;
      await getAutoScalerConfig();
      if (autoscalerData.value.status === 'UPDATING') {
        start();
      }
      configLoading.value = false;
    };
    const { start, stop } = useInterval(getAutoScalerConfig, 5000); // 轮询
    // 自动扩容开启｜关闭
    const user = computed(() => $store.state.user);
    // eslint-disable-next-line @typescript-eslint/no-misused-promises
    const handleToggleAutoScaler = async value => new Promise(async (resolve, reject) => {
      if (!clusterData.value?.clusterBasicSettings?.module?.workerModuleID) {
        $bkInfo({
          type: 'warning',
          clsName: 'custom-info-confirm',
          title: $i18n.t('cluster.ca.tips.noModule'),
          defaultInfo: true,
          okText: $i18n.t('cluster.ca.button.edit'),
          confirmFn: () => {
            $router.replace({ query: { clusterId: props.clusterId, active: 'node' } });
          },
          cancelFn: () => {
            // eslint-disable-next-line prefer-promise-reject-errors
            reject(false);
          },
        });
      } else if (!autoscalerData.value.enableAutoscale
        && (!nodepoolList.value.length
        || nodepoolList.value.every(item => !item.enableAutoscale)
        || nodepoolList.value.every(item => item.status !== 'RUNNING')
        )) {
        // 开启时前置判断是否存在节点规格 或 节点规格都是未开启状态时，要提示至少开启一个
        $bkInfo({
          type: 'warning',
          clsName: 'custom-info-confirm',
          title: !nodepoolList.value.length
            ? $i18n.t('cluster.ca.tips.emptyNodePool')
            : $i18n.t('cluster.ca.tips.notEnableAnyNodePool'),
          defaultInfo: true,
          okText: $i18n.t('cluster.ca.button.create'),
          confirmFn: () => {
            handleCreatePool();
          },
          cancelFn: () => {
            // eslint-disable-next-line prefer-promise-reject-errors
            reject(false);
          },
        });
      } else {
        // 开启或关闭扩缩容
        const result = await $store.dispatch('clustermanager/toggleClusterAutoScalingStatus', {
          enable: value,
          provider: _INTERNAL_.value ? 'selfProvisionCloud' : '',
          $clusterId: props.clusterId,
          updater: user.value.username,
        });
        if (result) {
          $bkMessage({
            theme: 'success',
            message: $i18n.t('generic.msg.success.ok'),
          });
          handleGetAutoScalerConfig();
          resolve(true);
        } else {
          // eslint-disable-next-line prefer-promise-reject-errors
          reject(false);
        }
      }
    });
    // eslint-disable-next-line @typescript-eslint/no-misused-promises
    const handleChangeScalerDown = async () => new Promise(async (resolve, reject) => {
      const value = !autoscalerData.value.isScaleDownEnable;
      $bkInfo({
        type: 'warning',
        clsName: 'custom-info-confirm',
        title: value ? $i18n.t('cluster.ca.title.confirmOnScalerDownConfig') : $i18n.t('cluster.ca.title.confirmOffScalerDownConfig'),
        defaultInfo: true,
        confirmFn: async () => {
          configLoading.value = true;
          const result = await $store.dispatch('clustermanager/updateClusterAutoScaling', {
            ...autoscalerData.value,
            isScaleDownEnable: value,
            updater: user.value.username,
            $clusterId: props.clusterId,
          });
          configLoading.value = false;
          if (result) {
            $bkMessage({
              theme: 'success',
              message: $i18n.t('generic.msg.success.update'),
            });
            handleGetAutoScalerConfig();
            resolve(true);
          } else {
            // eslint-disable-next-line prefer-promise-reject-errors
            reject(false);
          }
        },
        cancelFn: () => {
          // eslint-disable-next-line prefer-promise-reject-errors
          reject(false);
        },
      });
    });
    // 低优先级Pod配置
    const podsPriorityLoading = ref(false);
    const showPodsPriorityDialog = ref(false);
    const curPodsPriority = ref(-10);
    const isPodsPriorityEnable = computed(() => autoscalerData.value?.expendablePodsPriorityCutoff !== -2147483648);
    const handleTogglePodsPriorityDialog = () => {
      if (!isPodsPriorityEnable.value) {
        // 开启
        curPodsPriority.value = -10;// 开启时默认设置为 -10
        showPodsPriorityDialog.value = true;
      } else {
        // 关闭
        curPodsPriority.value  = -2147483648;
        handleSetPodsPriority();
      }
    };
    const handleSetPodsPriority = async () => {
      podsPriorityLoading.value = true;
      const result = await $store.dispatch('clustermanager/updateClusterAutoScaling', {
        ...autoscalerData.value,
        expendablePodsPriorityCutoff: curPodsPriority.value,
        updater: user.value.username,
        $clusterId: props.clusterId,
      });
      if (result) {
        $bkMessage({
          theme: 'success',
          message: $i18n.t('generic.msg.success.modify'),
        });
        handleGetAutoScalerConfig();
        showPodsPriorityDialog.value = false;
      }
      podsPriorityLoading.value = false;
    };

    // 资源池设置
    const isEditDevicePool = ref(false);
    const disableSelfDevicePool = ref(true);
    const newProvider = ref<'yunti'|'self'|''>('');
    const handleChangeDevicePool = async (value) => {
      newProvider.value = value;
    };
    const handleSaveDevicePoolChange = async () => {
      if (!newProvider.value) {
        isEditDevicePool.value = false;
        return;
      };
      configLoading.value = true;
      const result = await updateClusterAutoScalingProviders({
        $clusterId: props.clusterId,
        $provider: newProvider.value,
      }).then(() => true)
        .catch(() => false);
      configLoading.value = false;
      if (result) {
        $bkMessage({
          theme: 'success',
          message: $i18n.t('generic.msg.success.modify'),
        });
        isEditDevicePool.value = false;
        handleGetAutoScalerConfig();
      }
    };
    const { projectID, curProject } = useProject();
    const curCluster = computed<ICluster>(() => ($store.state as any).cluster.clusterList
      ?.find(item => item.clusterID === props.clusterId) || {});
    // 判断是否支持self资源池
    const getSelfDevicePool = async () => {
      if (curCluster.value.provider !== 'tencentCloud') return;

      const data = await $store.dispatch('clustermanager/cloudInstanceTypes', {
        $cloudID: curCluster.value.provider,
        region: curCluster.value.region,
        accountID: curCluster.value.cloudAccountID,
        projectID: projectID.value,
        version: 'v2',
        provider: 'self',
        resourceType: 'online',
        bizID: curProject.value?.businessID,
      });
      disableSelfDevicePool.value = !data?.length;
    };

    // 节点池
    const statusTextMap = { // 节点规格状态
      RUNNING: $i18n.t('generic.status.ready'),
      CREATING: $i18n.t('generic.status.creating'),
      DELETING: $i18n.t('generic.status.deleting'),
      UPDATING: $i18n.t('generic.status.updating'),
      DELETED: $i18n.t('generic.status.deleted'),
      'CREATE-FAILURE': $i18n.t('generic.status.createFailed'),
      'UPDATE-FAILURE': $i18n.t('generic.status.updateFailed'),
      'DELETE-FAILURE': $i18n.t('generic.status.deleteFailed'),
    };
    const statusColorMap = {
      RUNNING: 'green',
      DELETED: 'gray',
      'CREATE-FAILURE': 'red',
      'UPDATE-FAILURE': 'red',
      'DELETE-FAILURE': 'red',
    };
    const nodepoolList = ref<any[]>([]);
    const nodepoolLoading = ref(false);
    const {
      pagination,
      curPageData,
      pageChange,
      pageSizeChange,
    } = usePage(nodepoolList);
    const getNodePoolList = async () => {
      const list = await $store.dispatch('clustermanager/nodeGroup', {
        clusterID: props.clusterId,
      });
      const promiseList = list.map(item => $store.dispatch('clustermanager/nodeGroupNodeList', {
        $nodeGroupID: item.nodeGroupID,
        output: 'wide',
      }));
      const data = await Promise.all(promiseList);
      nodepoolList.value = list.map((item, index) => {
        // 节点数量动态从节点列表获取
        item.autoScaling.desiredSize = data[index]?.length;
        return item;
      });
      // if (!nodepoolList.value.some(pool => [
      //   'CREATING',
      //   'DELETING',
      //   'UPDATING',
      // ].includes(pool.status))) {
      //   stopPoolInterval();
      // } else {
      //   startPoolInterval();
      // }
    };
    const handleGetNodePoolList = async () => {
      nodepoolLoading.value = true;
      await getNodePoolList();
      if (!!nodepoolList.value.length) {
        startPoolInterval();
      }
      // if (nodepoolList.value.some(pool => [
      //   'CREATING',
      //   'DELETING',
      //   'UPDATING',
      // ].includes(pool.status))) {
      //   startPoolInterval();
      // }
      nodepoolLoading.value = false;
    };
    const { start: startPoolInterval, stop: stopPoolInterval } = useInterval(getNodePoolList, 5000); // 轮询
    // 节点规格详情
    const handleGotoDetail = (row) => {
      $router.push({
        name: 'nodePoolDetail',
        params: {
          clusterId: props.clusterId,
          nodeGroupID: row.nodeGroupID,
        },
      }).catch((err) => {
        console.warn(err);
      });
    };
    // 至少保证一个节点规格处于开启状态
    const disabledAutoscaler = computed(() => autoscalerData.value.enableAutoscale
                    && nodepoolList.value.filter(item => item.enableAutoscale).length <= 1);
    // 单节点开启和关闭弹性伸缩
    const { proxy } = getCurrentInstance() || { proxy: null };
    const handleToggleNodeScaler = async (row) => {
      if (nodepoolLoading.value || ['CREATING', 'DELETING', 'UPDATING'].includes(row.status)) return;

      const $refs = proxy?.$refs || {};
      $refs[row.nodeGroupID] && ($refs[row.nodeGroupID] as any).hideHandler();
      nodepoolLoading.value = true;
      let result = false;
      if (row.enableAutoscale) {
        // 关闭时校验是否时最后一个开启状态
        if (disabledAutoscaler.value) {
          nodepoolLoading.value = false;
          return;
        }
        // 关闭
        result = await $store.dispatch('clustermanager/disableNodeGroupAutoScale', {
          $nodeGroupID: row.nodeGroupID,
        });
      } else {
        // 包年包月模式不建议开启节点池
        if (curCluster.value.provider === 'tencentPublicCloud' && row.launchTemplate.instanceChargeType === 'PREPAID') {
          $bkInfo({
            type: 'warning',
            clsName: 'custom-info-confirm',
            title: $i18n.t('cluster.ca.nodePool.action.on'),
            subTitle: $i18n.t('tke.tips.prepaidOfEnableCA'),
            defaultInfo: true,
            confirmFn: async () => {
              // 启用
              const result = await $store.dispatch('clustermanager/enableNodeGroupAutoScale', {
                $nodeGroupID: row.nodeGroupID,
              });
              if (result) {
                $bkMessage({
                  theme: 'success',
                  message: $i18n.t('generic.msg.success.ok'),
                });
                await handleGetNodePoolList();
              }
            },
          });
        } else {
          // 启用
          result = await $store.dispatch('clustermanager/enableNodeGroupAutoScale', {
            $nodeGroupID: row.nodeGroupID,
          });
        }
      }
      if (result) {
        $bkMessage({
          theme: 'success',
          message: $i18n.t('generic.msg.success.ok'),
        });
        await handleGetNodePoolList();
      }
      nodepoolLoading.value = false;
    };
    // 当前操作行
    const currentOperateRow = ref<Record<string, any>>({});
    // 删除node pool
    const disabledDelete = computed(() =>
    // 至少保证一个节点规格
      autoscalerData.value.enableAutoscale
                    && nodepoolList.value.length <= 1);
    const handleDeletePool = (row) => {
      if (disabledDelete.value || !!row.autoScaling.desiredSize) return;

      $bkInfo({
        type: 'warning',
        clsName: 'custom-info-confirm',
        title: $i18n.t('cluster.ca.nodePool.action.delete.title', { name: `${row.nodeGroupID}（${row.name}）` }),
        defaultInfo: true,
        confirmFn: async () => {
          const result = await $store.dispatch('clustermanager/deleteNodeGroup', {
            $nodeGroupID: row.nodeGroupID,
            operator: user.value.username,
          });
          if (result) {
            $bkMessage({
              theme: 'success',
              message: $i18n.t('generic.msg.success.ok'),
            });
            handleGetNodePoolList();
          }
        },
      });
    };
    // 节点规格节点数量管理
    const nodeStatusMap = {
      INITIALIZATION: $i18n.t('generic.status.initializing'),
      RUNNING: $i18n.t('generic.status.ready'),
      DELETING: $i18n.t('generic.status.deleting'),
      DELETED: $i18n.t('generic.status.deleted'),
      'DELETE-FAILURE': $i18n.t('generic.status.deleteFailed'),
      'ADD-FAILURE': $i18n.t('cluster.ca.nodePool.nodes.status.scaleFailed'),
      'REMOVE-FAILURE': $i18n.t('cluster.ca.nodePool.nodes.status.downFailed'),
      REMOVABLE: $i18n.t('generic.status.removable'),
      NOTREADY: $i18n.t('generic.status.notReady'),
      UNKNOWN: $i18n.t('generic.status.unknown1'),
      'REMOVE-CA-FAILURE': $i18n.t('cluster.ca.nodePool.nodes.status.removeFailed'),
      APPLYING: $i18n.t('cluster.nodeList.status.applying'),
      'APPLY-FAILURE': $i18n.t('cluster.nodeList.status.applyFailure'),
    };
    const nodeColorMap = {
      RUNNING: 'green',
      'DELETE-FAILURE': 'red',
      'ADD-FAILURE': 'red',
      'REMOVE-FAILURE': 'red',
      'REMOVE-CA-FAILURE': 'red',
      'APPLY-FAILURE': 'red',
    };
    const nodeListLoading = ref(false);
    const nodeList = ref<any[]>([]);
    // 节点数量状态筛选
    const filtersStatus = computed(() => nodeList.value.reduce((pre, node) => {
      const exit = pre.find(item => item.value === node.status);
      if (!exit) {
        pre.push({
          text: nodeStatusMap[node.status] || node.status,
          value: node.status,
        });
      }
      return pre;
    }, []));
    const filtersStatusValue = ref<string[]>([]);
    const searchIpData = ref('');
    const handleNodeFilterChange = (data) => {
      filtersStatusValue.value = data?.status || [];
    };
    const filterNodeList = computed(() => nodeList.value
      .filter(node => (!filtersStatusValue.value.length || filtersStatusValue.value.includes(node.status))
          && (!searchIpData.value || searchIpData.value.split(',')?.includes(node.innerIP) || searchIpData.value.split(' ')?.includes(node.innerIP))));
    const {
      pagination: nodePagination,
      curPageData: nodeCurPageData,
      pageChange: nodePageChange,
      pageSizeChange: nodePageSizeChange,
    } = usePage(filterNodeList);
    watch(searchIpData, () => {
      nodePageChange(1);
    });
    const showNodeManage = ref(false);
    const handleNodeManageCancel = () => {
      currentOperateRow.value = {};
      nodeList.value = [];
      handleClearSelection();
      stopNodeInterval();
      // 刷新节点规格
      handleGetNodePoolList();
    };
    const handleShowNodeManage = (row) => {
      currentOperateRow.value = row;
      showNodeManage.value = true;
      handleGetNodeList();
    };
    const getNodeList = async () => {
      if (!currentOperateRow.value?.nodeGroupID) {
        stopNodeInterval();
        return;
      }
      nodeList.value = await $store.dispatch('clustermanager/nodeGroupNodeList', {
        $nodeGroupID: currentOperateRow.value.nodeGroupID,
        output: 'wide',
      });
      if (nodeList.value.some(node => ['DELETING', 'INITIALIZATION'].includes(node.status))) {
        startNodeInterval();
      } else {
        stopNodeInterval();
      }
    };
    const handleGetNodeList = async () => {
      nodeListLoading.value = true;
      await getNodeList();
      nodeListLoading.value = false;
    };
    const { start: startNodeInterval, stop: stopNodeInterval } = useInterval(getNodeList, 5000); // 轮询
    const handleNodeDrain = async (row) => {
      if (nodeListLoading.value) return;

      $bkInfo({
        type: 'warning',
        clsName: 'custom-info-confirm',
        title: $i18n.t('generic.button.drain.title'),
        subTitle: $i18n.t('generic.button.drain.subTitle', { ip: row.innerIP }),
        defaultInfo: true,
        confirmFn: async () => {
          // POD迁移
          nodeListLoading.value = true;
          const result = await $store.dispatch('clustermanager/clusterNodeDrain', {
            innerIPs: [row.innerIP],
            clusterID: props.clusterId,
            updater: user.value.username,
          });
          if (result) {
            $bkMessage({
              theme: 'success',
              message: $i18n.t('cluster.ca.nodePool.nodes.msg.drainSuccess'),
            });
            await getNodeList();
          }
          nodeListLoading.value = false;
        },
      });
    };
    const {
      batchDeleteNodes,
      handleCordonNodes,
      handleUncordonNodes,
      schedulerNode,
    } = useNode();
    const handleDeleteNodeGroupNode = async (row) => {
      if (nodeListLoading.value || (!row.unSchedulable && row.status !== 'APPLY-FAILURE')) return;

      $bkInfo({
        type: 'warning',
        clsName: 'custom-info-confirm',
        title: $i18n.t('cluster.ca.nodePool.nodes.action.delete.title'),
        subTitle: $i18n.t('cluster.ca.nodePool.nodes.action.delete.subTitle', { ip: row.innerIP }),
        defaultInfo: true,
        confirmFn: async () => {
          // 删除节点组节点
          nodeListLoading.value = true;
          const result = await batchDeleteNodes({
            $clusterId: props.clusterId,
            operator: user.value.username,
            nodeIPs: row.innerIP,
            virtualNodeIDs: !row.innerIP ? row.nodeID : '',
          });
          if (result) {
            $bkMessage({
              theme: 'success',
              message: $i18n.t('generic.msg.success.ok'),
            });
            await Promise.all([
              handleGetNodePoolList(),
              getNodeList(),
            ]);
          }
          nodeListLoading.value = false;
        },
      });
    };
    const handleToggleCordon = async (row) => {
      // 停止调度和允许调度
      nodeListLoading.value = true;
      let result = false;
      if (row.unSchedulable) {
        // 允许调度
        result = await $store.dispatch('clustermanager/nodeUnCordon', {
          innerIPs: [row.innerIP],
          clusterID: props.clusterId,
          updater: user.value.username,
        });
      } else {
        // 停止调度
        result = await $store.dispatch('clustermanager/nodeCordon', {
          innerIPs: [row.innerIP],
          clusterID: props.clusterId,
          updater: user.value.username,
        });
      }
      if (result) {
        $bkMessage({
          theme: 'success',
          message: $i18n.t('generic.msg.success.cordon'),
        });
        await getNodeList();
      }
      nodeListLoading.value = false;
    };
    // 节点批量操作
    const showBatchMenu = ref(false);
    const filterFailureTableData = computed(() => filterNodeList.value
      .filter(item => !['INITIALIZATION', 'DELETING', 'APPLYING'].includes(item.status)));
    const filterFailureCurTableData = computed(() => nodeCurPageData.value
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
    // 添加节点
    const handleAddNode = (row) => {
      const { href } = $router.resolve({
        name: 'addClusterNode',
        params: {
          clusterId: props.clusterId,
        },
        query: {
          source: 'nodePool',
          nodePool: row?.nodeGroupID,
        },
      });
      window.open(href);
    };
    const podDisabled = computed(() => !selections.value.every(select => select.status === 'REMOVABLE'));
    const disableBatchDelete = computed(() => selections.value.some(item => item.status === 'RUNNING'));
    // 弹窗二次确认
    const bkComfirmInfo = ({ title, subTitle, callback }) => {
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
    // 批量允许调度
    const handleBatchEnableNodes = () => {
      if (!selections.value.length) return;

      bkComfirmInfo({
        title: $i18n.t('generic.button.uncordon.title2'),
        subTitle: $i18n.t('generic.button.uncordon.subTitle2', {
          ip: selections.value[0].innerIP,
          num: selections.value.length,
        }),
        callback: async () => {
          const result = await handleUncordonNodes({
            clusterID: props.clusterId,
            nodes: selections.value.map(item => item.innerIP),
          });
          if (result) {
            handleGetNodeList();
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
          ip: selections.value[0].innerIP,
          num: selections.value.length,
        }),
        callback: async () => {
          const result = await handleCordonNodes({
            clusterID: props.clusterId,
            nodes: selections.value.map(item => item.innerIP),
          });
          if (result) {
            handleGetNodeList();
            handleResetCheckStatus();
          }
        },
      });
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
      bkComfirmInfo({
        title: $i18n.t('generic.button.drain.title'),
        subTitle: $i18n.t('generic.button.drain.subTitle2', {
          num: selections.value.length,
          ip: selections.value[0].innerIP,
        }),
        callback: async () => {
          await schedulerNode({
            clusterId: props.clusterId,
            nodes: selections.value.map(item => item.innerIP),
          });
        },
      });
    };
    // 批量删除节点
    const handleBatchDeleteNodes = () => {
      if (disableBatchDelete.value) return;
      bkComfirmInfo({
        title: $i18n.t('cluster.ca.nodePool.nodes.action.delete.title'),
        subTitle: $i18n.t('cluster.nodeList.button.delete.subTitle', {
          num: selections.value.length,
          ip: selections.value[0].innerIP || selections.value[0].nodeID,
        }),
        callback: async () => {
          const nodeIPs: string[] = [];
          const virtualNodeIDs: string[] = [];
          selections.value.forEach((row) => {
            if (row.innerIP) {
              nodeIPs.push(row.innerIP);
            } else if (row.nodeID) {
              virtualNodeIDs.push(row.nodeID);
            }
          });
          const result = await batchDeleteNodes({
            $clusterId: props.clusterId,
            nodeIPs: nodeIPs.join(','),
            virtualNodeIDs: virtualNodeIDs.join(','),
            operator: user.value.username,
          });
          if (result) {
            $bkMessage({
              theme: 'success',
              message: $i18n.t('generic.msg.success.ok'),
            });
            handleGetNodeList();
            handleResetCheckStatus();
          }
        },
      });
    };

    // 扩缩容记录
    const taskStatusMap = {
      INITIALIZING: $i18n.t('generic.status.initializing'),
      RUNNING: $i18n.t('generic.status.doing'),
      SUCCESS: $i18n.t('generic.status.taskSuccess'),
      FAILURE: $i18n.t('generic.status.taskFailed'),
      TIMEOUT: $i18n.t('generic.status.taskTimeout'),
      FORCETERMINATE: $i18n.t('generic.status.taskKill'),
      NOTSTARTED: $i18n.t('generic.status.taskNotEnable'),
    };
    const taskTypeMap = {
      CreateNodeGroup: $i18n.t('cluster.ca.nodePool.records.taskType.create'),
      UpdateNodeGroup: $i18n.t('cluster.ca.nodePool.records.taskType.update'),
      DeleteNodeGroup: $i18n.t('cluster.ca.nodePool.action.delete.text'),
      SwitchNodeGroupAutoScaling: $i18n.t('cluster.ca.nodePool.records.taskType.enable'),
      UpdateNodeGroupDesiredNode: $i18n.t('cluster.ca.nodePool.records.taskType.scale'),
      CleanNodeGroupNodes: $i18n.t('cluster.ca.nodePool.records.taskType.down'),
    };
    const filters = computed(() => ({
      taskType: Object.keys(taskTypeMap).map(key => ({ text: taskTypeMap[key], value: key })),
      status: [
        {
          text: $i18n.t('generic.status.initializing'),
          value: 'INITIALIZING',
        },
        {
          text: $i18n.t('generic.status.doing'),
          value: 'RUNNING',
        },
        {
          text: $i18n.t('generic.status.taskSuccess'),
          value: 'SUCCESS',
        },
        {
          text: $i18n.t('generic.status.taskFailed'),
          value: 'FAILURE',
        },
        {
          text: $i18n.t('generic.status.taskTimeout'),
          value: 'TIMEOUT',
        },
      ],
      resourceID: nodepoolList.value.map(item => ({
        text: `${item.nodeGroupID}( ${item.name} )`,
        value: item.nodeGroupID,
      })),
    }));
    const filterValues = ref<{
      taskType: string[]
      status: string[]
      resourceID: string[]
    }>({
      taskType: [],
      status: [],
      resourceID: [],
    });
    const taskStatusColorMap = {
      INITIALIZING: 'green',
      RUNNING: 'green',
      SUCCESS: 'green',
      FAILURE: 'red',
      TIMEOUT: 'red',
      FORCETERMINATE: 'red',
      NOTSTARTED: 'gray',
    };
    const shortcuts = ref([
      {
        text: $i18n.t('units.time.today'),
        value() {
          const end = new Date();
          const start = new Date(end.getFullYear(), end.getMonth(), end.getDate());
          return [start, end];
        },
      },
      {
        text: $i18n.t('units.time.lastDays'),
        value() {
          const end = new Date();
          const start = new Date();
          start.setTime(start.getTime() - 3600 * 1000 * 24 * 7);
          return [start, end];
        },
      },
      {
        text: $i18n.t('units.time.last15Days'),
        value() {
          const end = new Date();
          const start = new Date();
          start.setTime(start.getTime() - 3600 * 1000 * 24 * 15);
          return [start, end];
        },
      },
      {
        text: $i18n.t('units.time.last30Days'),
        value() {
          const end = new Date();
          const start = new Date();
          start.setTime(start.getTime() - 3600 * 1000 * 24 * 30);
          return [start, end];
        },
      },
    ]);
    const expandRowKeys = ref<any[]>([]);
    const timeRange = ref<Date[]>([]);
    const showRecord = ref(false);
    const recordLoading = ref(false);
    const recordList = ref<any[]>([]);
    const recordPagination = ref({
      current: 1,
      limit: 10,
      count: 0,
      showTotalCount: true,
    });
    const searchIp = useDebouncedRef<string>('');
    watch(searchIp, () => {
      reSearchRecordList();
    });

    const reSearchRecordList = () => {
      recordPagination.value.current = 1;
      handleGetRecordList();
    };
    const recordPageChange = (page) => {
      recordPagination.value.current = page;
      handleGetRecordList();
    };
    const recordPageSizeChange = (limit) => {
      recordPagination.value.current = 1;
      recordPagination.value.limit = limit;
      handleGetRecordList();
    };
    const handleTimeRangeChange = () => {
      reSearchRecordList();
    };
    const handleShowRecord = (row) => {
      const end = new Date();
      const start = new Date();
      start.setTime(start.getTime() - 3600 * 1000 * 24 * 7);
      timeRange.value = [
        start,
        end,
      ];
      filterValues.value = {
        taskType: [],
        status: [],
        resourceID: [row.nodeGroupID],
      };
      currentOperateRow.value = row;
      filterValues.value = {
        taskType: [],
        status: [],
        resourceID: row.nodeGroupID ? [row.nodeGroupID] : [],
      };
      showRecord.value = true;
      handleGetRecordList();
    };
    const handleRecordCancel = () => {
      currentOperateRow.value = {};
      recordPagination.value = {
        current: 1,
        limit: 10,
        count: 0,
        showTotalCount: true,
      };
      expandRowKeys.value = [];
    };
    const { start: startTaskPool, stop: stopTaskPool } = useInterval(getRecordList, 5000); // 轮询
    async function getRecordList() {
      const { taskType = [], resourceID = [], status = [] } = filterValues.value;
      const { results = [], count = 0 } = await $store.dispatch('clustermanager/clusterAutoScalingLogsV2', {
        resourceType: 'nodegroup',
        resourceID: resourceID?.[0],
        startTime: Math.floor(new Date(timeRange.value[0]).getTime() / 1000) || '',
        endTime: Math.floor(new Date(timeRange.value[1]).getTime() / 1000) || '',
        limit: recordPagination.value.limit,
        page: recordPagination.value.current,
        status: status?.[0],
        taskType: taskType?.[0],
        clusterID: props.clusterId,
        ipList: searchIp.value.split(' ').join(','),
      });
      recordList.value = results.map(item => ({
        ...item,
        taskID: item.taskID || Math.random() * 1000,
      }));
      if (recordList.value.some(row => row.task?.status === 'RUNNING')) {
        startTaskPool();
      } else {
        stopTaskPool();
      }
      recordPagination.value.count = count;
    };
    const handleGetRecordList = async () => {
      recordLoading.value = true;
      await getRecordList();
      recordLoading.value = false;
    };
    const handleExpandChange = (row) => {
      const index = expandRowKeys.value.findIndex(key => key === row.taskID);
      if (index > -1) {
        expandRowKeys.value.splice(index, 1);
      } else {
        expandRowKeys.value.push(row.taskID);
      }
    };
    const handleFilterChange = async (filters) => {
      Object.keys(filters).forEach((key) => {
        filterValues.value[key] = filters[key];
      });
      recordPagination.value.current = 1;
      recordLoading.value = true;
      await getRecordList();
      recordLoading.value = false;
    };

    // 编辑自动扩缩容
    const handleEditAutoScaler = () => {
      $router.push({
        name: 'autoScalerConfig',
        params: {
          clusterId: props.clusterId,
        },
      });
    };
    // 新建节点规格
    const handleCreatePool = () => {
      $router.push({
        name: 'nodePool',
        params: {
          clusterId: props.clusterId,
        },
        query: {
          provider: autoscalerData.value.devicePoolProvider || 'yunti',
        },
      }).catch((err) => {
        console.warn(err);
      });
    };
    // 克隆节点
    const handleClonePool = (row) => {
      $router.push({
        name: 'nodePool',
        params: {
          clusterId: props.clusterId,
          nodeGroupID: row.nodeGroupID,
        },
        query: {
          provider: autoscalerData.value.devicePoolProvider || 'yunti',
        },
      }).catch((err) => {
        console.warn(err);
      });
    };
    // 编辑节点规格
    const handleEditPool = (row) => {
      $router.push({
        name: 'editNodePool',
        params: {
          clusterId: props.clusterId,
          nodeGroupID: row.nodeGroupID,
        },
        query: {
          provider: row.extraInfo?.resourcePoolType || 'yunti',
        },
      }).catch((err) => {
        console.warn(err);
      });
    };
    // 集群详情
    const { clusterOS, clusterData, getClusterDetail } = useClusterInfo();

    // 重试任务
    // const handleRetryTask = async (row) => {
    //   const result = await $store.dispatch('clustermanager/taskRetry', {
    //     $taskId: row.taskID,
    //     updater: user.value.username,
    //   });
    //   result && handleGetRecordList();
    // };
    // ip列表
    const ipTableKey = ref('');
    const showIPList = ref(false);
    const handleShowIPList = (row) => {
      ipTableKey.value = String(Math.random() * 100);// 修复dialog嵌套表格自适应问题
      showIPList.value = true;
      currentOperateRow.value = row;
    };
    // 复制所有IP
    const handleCopyAllIP = () => {
      const ipList = currentOperateRow.value.task.nodeIPList || [];
      if (!ipList.length) return;

      copyText(ipList.join('\n'));
      $bkMessage({
        theme: 'success',
        message: $i18n.t('generic.msg.success.copy'),
      });
    };

    // 跳转标准运维
    const handleGotoSops = (url: string) => {
      window.open(url);
    };

    // 集群装箱率
    const overview = ref<any>({
      cpu_usage: {},
      memory_usage: {},
    });
    const handleGetClusterOverview = async () => {
      overview.value = await clusterOverview({ $clusterId: props.clusterId });
    };
    const conversionPercentUsed = (used, total) => {
      if (!total || parseFloat(total) === 0) {
        return 0;
      }

      let ret: any = parseFloat(used || '0') / parseFloat(total) * 100;
      if (ret !== 0 && ret !== 100) {
        ret = ret.toFixed(2);
      }
      return ret;
    };

    // 自适应两栏布局
    const autoscalerRef = ref();
    const { cols } = useAutoCols(autoscalerRef, [
      {
        min: 0,
        max: 840,
        cols: 1,
      },
      {
        min: 840,
        max: Infinity,
        cols: 2,
      }]);

    onMounted(() => {
      handleGetAutoScalerConfig();
      handleGetNodePoolList();
      getClusterDetail(props.clusterId, true);
      handleGetClusterOverview();
      getSelfDevicePool();
    });
    onBeforeUnmount(() => {
      stop();
      stopPoolInterval();
      stopNodeInterval();
      stopTaskPool();
    });
    return {
      curCluster,
      searchIp,
      isPodsPriorityEnable,
      podsPriorityLoading,
      showPodsPriorityDialog,
      curPodsPriority,
      handleSetPodsPriority,
      handleTogglePodsPriorityDialog,
      filters,
      filterValues,
      ipTableKey,
      showIPList,
      handleCopyAllIP,
      conversionPercentUsed,
      formatBytes,
      overview,
      expandRowKeys,
      clusterOS,
      disabledAutoscaler,
      disabledDelete,
      currentOperateRow,
      nodeListLoading,
      showNodeManage,
      configLoading,
      nodepoolLoading,
      pagination,
      curPageData,
      pageChange,
      pageSizeChange,
      nodeStatusMap,
      nodeColorMap,
      nodePagination,
      nodeCurPageData,
      handleNodeManageCancel,
      nodePageChange,
      nodePageSizeChange,
      handleNodeDrain,
      handleDeleteNodeGroupNode,
      handleToggleCordon,
      autoscalerData,
      basicScalerConfig,
      autoScalerConfig,
      autoScalerDownConfig,
      statusColorMap,
      statusTextMap,
      showRecord,
      timeRange,
      recordLoading,
      handleTimeRangeChange,
      shortcuts,
      recordPagination,
      recordList,
      recordPageChange,
      recordPageSizeChange,
      handleRecordCancel,
      handleGotoDetail,
      handleShowNodeManage,
      handleToggleNodeScaler,
      handleDeletePool,
      handleToggleAutoScaler,
      handleGetNodeList,
      handleShowRecord,
      handleEditAutoScaler,
      handleCreatePool,
      handleChangeScalerDown,
      handleEditPool,
      taskStatusMap,
      taskStatusColorMap,
      // handleRetryTask,
      handleExpandChange,
      handleGotoSops,
      handleShowIPList,
      handleFilterChange,
      autoscalerRef,
      cols,
      isEditDevicePool,
      disableSelfDevicePool,
      handleChangeDevicePool,
      handleSaveDevicePoolChange,
      showBatchMenu,
      selectType,
      selections,
      handleResetCheckStatus,
      renderSelection,
      handleRowCheckChange,
      handleSelectionAll,
      handleClearSelection,
      CheckType,
      handleAddNode,
      podDisabled,
      disableBatchDelete,
      handleBatchEnableNodes,
      handleBatchStopNodes,
      handleBatchPodScheduler,
      handleBatchDeleteNodes,
      filtersStatus,
      filtersStatusValue,
      handleNodeFilterChange,
      searchIpData,
      handleClonePool,
    };
  },
});
</script>
<style lang="postcss" scoped>
.autoscaler-management {
    padding: 0px 32px 20px 32px;
    font-size: 12px;
    .group-header {
        display: flex;
        align-items: center;
        justify-content: space-between;
        padding: 20px 0;
        &-title {
            font-size: 14px;
            font-weight: bold;
            color: #63656E;
        }
    }
    .autoscaler {
        .switch-autoscaler {
            margin-left: 16px;
            padding-left: 16px;
            border-left: 1px solid #DCDEE5;
            color: #63656E;
        }
    }
    .group-border-top {
        border-top: 1px solid #DCDEE5;
    }
    .disabled {
        color: #C4C6CC;
    }
    >>> .bcs-form-content {
      display: flex;
      flex-wrap: wrap;
      &-item {
          margin-top: 0;
          font-size: 12px;
          width: 100%;
      }
      .bk-label {
          font-size: 12px;
          color: #979BA5;
          text-align: left;
      }
      .bk-form-content {
          font-size: 12px;
          color: #313238;
          display: flex;
          align-items: center;
      }
    }
}
.flex-between {
    display: flex;
    align-items: center;
    justify-content: space-between;
}
.operate {
    display: flex;
    align-items: center;
    >>> .more-icon {
        display: flex;
        align-items: center;
        justify-content: center;
        color: #63656E;
        font-size: 18px;
        cursor: pointer;
        margin-top: 2px;
        color: #3A84FF;
        &:hover:not(.disabled) {
            background: #eaf2ff;
            border-radius: 50%;
        }
        &.disabled {
          cursor: not-allowed;
          color: #dcdee5;
        }
    }
}
>>> .form-content {
    display: flex;
    flex-wrap: wrap;
    &-item {
        margin-top: 0;
        font-size: 12px;
        width: 100%;
    }
    .bk-label {
        font-size: 12px;
        color: #979BA5;
        text-align: left;
    }
    .bk-form-content {
        font-size: 12px;
        color: #313238;
        display: flex;
        align-items: center;
    }
}
>>> .dropdown-item {
    height: 32px;
    line-height: 32px;
    padding: 0 16px;
    color: #63656e;
    font-size: 12px;
    text-decoration: none;
    white-space: nowrap;
    cursor: pointer;
    &:hover:not(.disabled) {
        background-color: #eaf3ff;
        color: #3a84ff;
    }
    &.disabled {
        color: #C4C6CC;
        cursor: not-allowed;
    }
}
.mw88 {
    min-width: 88px;
}
.delete-content {
    display: flex;
    flex-direction: column;
    align-items: center;
}
.delete-title {
    color: #313238;
    font-size: 20px;
}
>>> .custom-header-cell {
    text-decoration: underline;
    text-decoration-style: dashed;
    text-underline-position: under;
}
</style>
