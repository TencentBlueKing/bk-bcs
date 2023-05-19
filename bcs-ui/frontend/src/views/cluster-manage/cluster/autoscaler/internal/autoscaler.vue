<!-- eslint-disable max-len -->
<template>
  <div class="autoscaler-management">
    <!-- 自动扩缩容配置 -->
    <section class="autoscaler">
      <div class="group-header">
        <div>
          <span class="group-header-title">{{$t('Cluster Autoscaler配置')}}</span>
          <span class="switch-autoscaler">
            {{$t('Cluster Autoscaler')}}
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
          @click="handleEditAutoScaler">{{$t('编辑配置')}}</bcs-button>
      </div>
      <div v-bkloading="{ isLoading: configLoading }">
        <LayoutGroup :title="$t('基本配置')" class="mb10">
          <AutoScalerFormItem
            :list="basicScalerConfig"
            :autoscaler-data="autoscalerData">
          </AutoScalerFormItem>
        </LayoutGroup>
        <LayoutGroup :title="$t('扩缩容暂停配置')" class="mb10">
          <i18n
            path="NotReady节点数大于 {0} 个且超过集群总节点数 {1} 时暂停自动扩缩容"
            class="text-[#979BA5] leading-[32px]">
            <span place="0" class="text-[#313238]">{{ autoscalerData.okTotalUnreadyCount }}</span>
            <span place="1" class="text-[#313238]">{{ autoscalerData.maxTotalUnreadyPercentage }}%</span>
          </i18n>
          <span
            class="ml-[5px] text-[16px] text-[#979ba5]"
            v-bk-tooltips="$t('自动扩缩容保护机制，如果NotReady节点数量或比例过大，自动扩容上来的节点也有可能会是NotReady状态，导致业务成本增加，当NotReady节点数不符合暂停触发条件时自动恢复自动扩缩容')">
            <i class="bk-icon icon-info-circle"></i>
          </span>
        </LayoutGroup>
        <LayoutGroup :title="$t('自动扩容配置')" class="mb10">
          <AutoScalerFormItem
            :list="autoScalerConfig"
            :autoscaler-data="autoscalerData">
            <template #suffix="{ data }">
              <span
                class="text-[#699DF4] ml-[5px] h-[20px] flex items-center"
                style="border-bottom: 1px dashed #699DF4;"
                v-if="['bufferResourceCpuRatio', 'bufferResourceMemRatio'].includes(data.prop)"
                v-bk-tooltips="data.prop === 'bufferResourceCpuRatio'
                  ? `${Number((overview.cpu_usage.request || 0)).toFixed(2)}${$t('核')} / ${Number(Math.ceil(overview.cpu_usage.total) || 0).toFixed(2)}${$t('核')}`
                  : `${formatBytes(overview.memory_usage.request_bytes || 0, 2)} / ${formatBytes(overview.memory_usage.total_bytes || 0, 2)}`">
                {{
                  $t('( 当前使用率 {val} % )', {
                    val: data.prop === 'bufferResourceCpuRatio'
                      ? conversionPercentUsed(overview.cpu_usage.request, overview.cpu_usage.total)
                      : conversionPercentUsed(overview.memory_usage.request_bytes, overview.memory_usage.total_bytes)
                  })
                }}
              </span>
            </template>
          </AutoScalerFormItem>
        </LayoutGroup>
        <LayoutGroup collapsible class="mb15" :expanded="!!autoscalerData.isScaleDownEnable">
          <template #title>
            <span class="flex items-center">
              <span>{{$t('自动缩容配置')}}</span>
              <span class="flex items-center ml-[8px]" @click.stop>
                <span
                  :class="['px-[8px] inline-block leading-[20px] text-[#979BA5]', {
                    '!text-[#2DCB56] bg-[#F2FFF4]': autoscalerData.isScaleDownEnable && autoscalerData.enableAutoscale
                  }]">
                  {{ autoscalerData.isScaleDownEnable && autoscalerData.enableAutoscale ? $t('已开启') : $t('已关闭') }}
                </span>
                <bcs-divider direction="vertical" class="!mr-[10px]"></bcs-divider>
                <span
                  v-bk-tooltips="{
                    disabled: autoscalerData.enableAutoscale,
                    content: $t('集群自动扩缩容已关闭，无法单独开启自动缩容配置')
                  }">
                  <bk-button
                    text
                    class="text-[12px]"
                    :disabled="autoscalerData.status === 'UPDATING' || !autoscalerData.enableAutoscale"
                    @click="handleChangeScalerDown">
                    {{ autoscalerData.isScaleDownEnable && autoscalerData.enableAutoscale ? $t('关闭') : $t('开启') }}
                  </bk-button>
                </span>
              </span>
            </span>
          </template>
          <AutoScalerFormItem
            :list="autoScalerDownConfig"
            :autoscaler-data="autoscalerData">
          </AutoScalerFormItem>
        </LayoutGroup>
        <LayoutGroup collapsible class="mb15" :expanded="isPodsPriorityEnable">
          <template #title>
            <span class="flex items-center">
              <span>{{$t('低优先级 Pod 配置')}}</span>
              <span class="flex items-center ml-[8px]" @click.stop>
                <span
                  :class="['px-[8px] inline-block leading-[20px] text-[#979BA5]', {
                    '!text-[#2DCB56] bg-[#F2FFF4]': isPodsPriorityEnable && autoscalerData.enableAutoscale
                  }]">
                  {{ isPodsPriorityEnable && autoscalerData.enableAutoscale ? $t('已开启') : $t('已关闭') }}
                </span>
                <bcs-divider direction="vertical" class="!mr-[10px]"></bcs-divider>
                <span
                  v-bk-tooltips="{
                    disabled: autoscalerData.enableAutoscale,
                    content: $t('集群自动扩缩容已关闭，无法单独开启低优先级 Pod 配置')
                  }">
                  <bk-button
                    text
                    class="text-[12px]"
                    :disabled="autoscalerData.status === 'UPDATING' || !autoscalerData.enableAutoscale"
                    @click="handleTogglePodsPriorityDialog">
                    {{ isPodsPriorityEnable && autoscalerData.enableAutoscale ? $t('关闭') : $t('开启') }}
                  </bk-button>
                </span>
              </span>
            </span>
          </template>
          <AutoScalerFormItem
            :list="[{
              prop: 'expendablePodsPriorityCutoff',
              name: $t('Pod 优先级阈值'),
              desc: $t('当优先级低于此值的 pod，pending不会触发扩容，缩容时会直接kill，不会等待优雅退出时间'),
            }]"
            :autoscaler-data="autoscalerData" />
        </LayoutGroup>
      </div>
    </section>
    <!-- 节点规格配置 -->
    <section class="nodepool">
      <div class="group-header">
        <div class="group-header-title">{{$t('节点规格管理')}}</div>
        <div class="flex">
          <bcs-button theme="primary" icon="plus" @click="handleCreatePool">{{$t('新建节点规格')}}</bcs-button>
          <bcs-button class="ml5" @click="handleShowRecord({})">{{$t('扩缩容记录')}}</bcs-button>
        </div>
      </div>
      <bcs-table
        :data="curPageData"
        :pagination="pagination"
        v-bkloading="{ isLoading: nodepoolLoading }"
        @page-change="pageChange"
        @page-limit-change="pageSizeChange">
        <bcs-table-column :label="$t('节点规格 ID (名称)')" min-width="200" show-overflow-tooltip>
          <template #default="{ row }">
            <div class="bk-primary bk-button-normal bk-button-text" @click="handleGotoDetail(row)">
              <span class="bcs-ellipsis">{{`${row.nodeGroupID}（${row.name}）`}}</span>
            </div>
          </template>
        </bcs-table-column>
        <bcs-table-column :label="$t('节点配额')" align="right" width="120">
          <template #default="{ row }">
            {{ row.autoScaling.maxSize }}
          </template>
        </bcs-table-column>
        <bcs-table-column :label="$t('节点数量')" align="right" width="120">
          <template #default="{ row }">
            <bcs-button
              text
              :disabled="row.autoScaling.desiredSize === 0"
              @click="handleShowNodeManage(row)">
              <div class="min-w-[80px] text-right">{{row.autoScaling.desiredSize}}</div>
            </bcs-button>
          </template>
        </bcs-table-column>
        <bcs-table-column :label="$t('机型')">
          <template #default="{ row }">
            {{ row.launchTemplate.instanceType }}
          </template>
        </bcs-table-column>
        <bcs-table-column :label="$t('操作系统')" show-overflow-tooltip>
          <template #default>{{clusterOS}}</template>
        </bcs-table-column>
        <bcs-table-column :label="$t('节点规格状态')">
          <template #default="{ row }">
            <LoadingIcon v-if="['CREATING', 'DELETING', 'UPDATING'].includes(row.status)">
              {{ statusTextMap[row.status] }}
            </LoadingIcon>
            <StatusIcon status="unknown" v-else-if="!row.enableAutoscale && row.status === 'RUNNING'">
              {{$t('已关闭')}}
            </StatusIcon>
            <StatusIcon
              :status="row.status"
              :status-color-map="statusColorMap"
              v-else>
              {{ statusTextMap[row.status] || $t('未知') }}
            </StatusIcon>
          </template>
        </bcs-table-column>
        <bcs-table-column :label="$t('操作')" width="170">
          <template #default="{ row }">
            <div class="operate">
              <bcs-button text @click="handleShowRecord(row)">{{$t('扩缩容记录')}}</bcs-button>
              <bcs-popover
                placement="bottom"
                theme="light dropdown"
                :arrow="false"
                class="ml15"
                :disabled="row.status === 'DELETING'"
                trigger="click"
                :ref="row.nodeGroupID">
                <span class="more-icon"><i class="bcs-icon bcs-icon-more"></i></span>
                <div slot="content">
                  <ul>
                    <li
                      :class="['dropdown-item', {
                        disabled: (row.enableAutoscale && disabledAutoscaler)
                          || ['CREATING', 'DELETING', 'UPDATING'].includes(row.status)
                      }]"
                      v-bk-tooltips="{
                        content: $t('Cluster Autoscaler 需要至少一个节点规格开启，请停用 Cluster Autoscaler 后再关闭'),
                        disabled: !(row.enableAutoscale && disabledAutoscaler)
                      }"
                      @click="handleToggleNodeScaler(row)">
                      {{row.enableAutoscale ? $t('关闭节点规格') : $t('启用节点规格')}}
                    </li>
                    <li class="dropdown-item" @click="handleEditPool(row)">{{$t('编辑节点规格')}}</li>
                    <li
                      :class="['dropdown-item', { disabled: disabledDelete || !!row.autoScaling.desiredSize }]"
                      v-bk-tooltips="{
                        content: !!row.autoScaling.desiredSize
                          ? $t('请删除节点后再删除节点规格')
                          : $t('Cluster Autoscaler 需要至少一个节点规格，请停用 Cluster Autoscaler 后再删除'),
                        disabled: !(disabledDelete || !!row.autoScaling.desiredSize),
                        placements: 'left'
                      }"
                      @click="handleDeletePool(row)">{{$t('删除节点规格')}}</li>
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
      :title="$t('管理节点数量')"
      :width="700"
      v-model="showNodeManage"
      @cancel="handleNodeManageCancel">
      <bcs-alert type="info" :title="$t('注意：若节点规格已开启自动伸缩， 则数量将会随集群负载自动调整')"></bcs-alert>
      <bcs-form class="form-content mt15" :label-width="100">
        <bcs-form-item class="form-content-item" :label="$t('节点规格名称')">
          <span>{{ currentOperateRow.name }}</span>
        </bcs-form-item>
        <bcs-form-item class="form-content-item" :label="$t('节点配额')">
          <span>
            {{
              currentOperateRow.autoScaling
                ? currentOperateRow.autoScaling.maxSize
                : '--'
            }}
          </span>
        </bcs-form-item>
        <bcs-form-item class="form-content-item" :label="$t('节点数量')">
          <span>
            {{
              currentOperateRow.autoScaling
                ? currentOperateRow.autoScaling.desiredSize
                : '--'
            }}
          </span>
        </bcs-form-item>
      </bcs-form>
      <bcs-table
        class="mt15"
        v-bkloading="{ isLoading: nodeListLoading }"
        :data="nodeCurPageData"
        :pagination="nodePagination"
        @page-change="nodePageChange"
        @page-limit-change="nodePageSizeChange">
        <bcs-table-column :label="$t('节点名称')" prop="innerIP"></bcs-table-column>
        <bcs-table-column :label="$t('状态')">
          <template #default="{ row }">
            <LoadingIcon v-if="['DELETING', 'INITIALIZATION'].includes(row.status)">
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
        <bcs-table-column :label="$t('操作')" width="120">
          <template #default="{ row }">
            <div class="operate">
              <bcs-button
                text
                :disabled="['DELETING', 'INITIALIZATION'].includes(row.status)"
                @click="handleToggleCordon(row)">
                {{row.unSchedulable ? $t('允许调度') : $t('停止调度')}}
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
                    <li class="dropdown-item" @click="handleNodeDrain(row)">{{$t('pod驱逐')}}</li>
                    <li
                      :class="['dropdown-item', { disabled: !row.unSchedulable }]"
                      v-bk-tooltips="{
                        content: $t('请先停止调度'),
                        disabled: row.unSchedulable
                      }"
                      @click="handleDeleteNodeGroupNode(row)"
                    >{{$t('删除节点')}}</li>
                  </ul>
                </div>
              </bcs-popover>
            </div>
          </template>
        </bcs-table-column>
      </bcs-table>
    </bcs-dialog>
    <!-- 扩缩容记录 -->
    <bcs-dialog
      theme="primary"
      header-position="left"
      :title="$t('扩缩容记录')"
      :width="1200"
      v-model="showRecord"
      @cancel="handleRecordCancel">
      <div class="mb15 flex-between">
        <div></div>
        <bcs-date-picker
          :shortcuts="shortcuts"
          type="datetimerange"
          shortcut-close
          :use-shortcut-text="false"
          :clearable="false"
          v-model="timeRange"
          @change="handleTimeRangeChange">
        </bcs-date-picker>
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
              <bcs-table-column :label="$t('步骤名称')" width="150" show-overflow-tooltip>
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
              <bcs-table-column :label="$t('步骤信息')" show-overflow-tooltip>
                <template #default="{ row: key }">
                  {{ row.task.steps[key].message || '--' }}
                </template>
              </bcs-table-column>
              <bcs-table-column :label="$t('开始时间')" width="180" show-overflow-tooltip>
                <template #default="{ row: key }">
                  {{ row.task.steps[key].start }}
                </template>
              </bcs-table-column>
              <bcs-table-column :label="$t('结束时间')" width="180" show-overflow-tooltip>
                <template #default="{ row: key }">
                  {{ row.task.steps[key].end }}
                </template>
              </bcs-table-column>
              <!-- 设置宽度为了保持和外面表格对齐 -->
              <bcs-table-column :label="$t('状态')" width="240">
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
          :label="$t('事件类型')"
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
        <bcs-table-column :label="$t('事件信息')" prop="message" show-overflow-tooltip></bcs-table-column>
        <bcs-table-column
          :label="$t('节点规格')"
          prop="resourceID"
          :filtered-value="filterValues.resourceID"
          :filters="filters.resourceID"
          :filter-multiple="false"
          filter-searchable
          column-key="resourceID"
          :key="JSON.stringify(filterValues.resourceID)"
          show-overflow-tooltip>
        </bcs-table-column>
        <bcs-table-column :label="$t('开始时间')" width="180" prop="createTime" show-overflow-tooltip></bcs-table-column>
        <bcs-table-column :label="$t('结束时间')" width="180" show-overflow-tooltip>
          <template #default="{ row }">
            {{row.task ? row.task.end : '--'}}
          </template>
        </bcs-table-column>
        <bcs-table-column
          :label="$t('状态')"
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
        <bcs-table-column :label="$t('操作')" width="120">
          <template #default="{ row }">
            <!-- <bcs-button
              text
              :disabled="!(row.task && taskStatusColorMap[row.task.status] === 'red')"
              @click="handleRetryTask(row)">{{$t('重试')}}</bcs-button> -->
            <bcs-button
              text
              :disabled="!(row.task && row.task.nodeIPList && row.task.nodeIPList.length)"
              @click="handleShowIPList(row)">{{$t('IP列表')}}</bcs-button>
          </template>
        </bcs-table-column>
      </bcs-table>
    </bcs-dialog>
    <!-- IP列表 -->
    <bcs-dialog
      theme="primary"
      header-position="left"
      :title="$t('IP列表')"
      v-model="showIPList">
      <bk-table
        :key="ipTableKey"
        :data="currentOperateRow.task ? currentOperateRow.task.nodeIPList : []">
        <bcs-table-column label="IP">
          <template #default="{ row }">{{ row }}</template>
        </bcs-table-column>
      </bk-table>
    </bcs-dialog>
    <!-- 低优先级Pod配置 -->
    <bcs-dialog
      theme="primary"
      header-position="left"
      :title="$t('开启低优先级 Pod 配置')"
      :width="480"
      :loading="podsPriorityLoading"
      v-model="showPodsPriorityDialog"
      @confirm="handleSetPodsPriority">
      <div class="flex items-start">
        <i class="bk-icon icon-info-circle mr-[8px] relative top-[2px]"></i>
        <i18n path="低优先级Pod配置提示语" class="text-[12px]">
          <span class="text-[#FF9C01] font-bold">{{ $t('低于') }}</span>
          <span>{{ $t('以下') }}</span>
        </i18n>
      </div>
      <div class="flex items-center mt-[20px] ml-[22px]">
        <span class="mr-[20px]">{{ $t('Pod 优先级阈值') }}</span>
        <bcs-input
          class="w-[74px]"
          type="number"
          :min="-2147483648"
          :max="-1"
          :show-controls="false"
          v-model="curPodsPriority">
        </bcs-input>
        <span class="text-[#979BA5] ml-[8px]">(-2147483648 - -1)</span>
      </div>
    </bcs-dialog>
  </div>
</template>
<script lang="ts">
import { defineComponent, onMounted, ref, computed, onBeforeUnmount, getCurrentInstance } from 'vue';
import $i18n from '@/i18n/i18n-setup';
import $router from '@/router';
import $store from '@/store/index';
import StatusIcon from '@/components/status-icon';
import LoadingIcon from '@/components/loading-icon.vue';
import usePage from '@/composables/use-page';
import useInterval from '@/composables/use-interval';
import LayoutGroup from '@/views/cluster-manage/components/layout-group.vue';
import AutoScalerFormItem from '../tencent/form-item.vue';
import { useClusterInfo } from '@/views/cluster-manage/cluster/use-cluster';
import { clusterOverview } from '@/api/modules/monitor';
import { formatBytes } from '@/common/util';
import $bkMessage from '@/common/bkmagic';
import $bkInfo from '@/components/bk-magic-2.0/bk-info';

export default defineComponent({
  name: 'AutoScaler',
  components: { StatusIcon, LoadingIcon, LayoutGroup, AutoScalerFormItem },
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
        name: $i18n.t('状态'),
      },
      {
        prop: 'scanInterval',
        name: $i18n.t('扩缩容检测时间间隔'),
        unit: $i18n.t('秒'),
      },
      {
        prop: 'scaleOutModuleName',
        name: $i18n.t('扩容后转移模块'),
        desc: $i18n.t('扩容节点后节点转移到关联业务的CMDB模块'),
      },
    ]);
    const autoScalerConfig = ref([
      {
        prop: 'expander',
        name: $i18n.t('扩容算法'),
        isBasicProp: true,
        desc: $i18n.t('random：在有多个节点规格时，随机选择节点规格<br/>least-waste：在有多个节点规格时，以最小浪费原则选择，选择有最少可用资源的节点规格<br/>most-pods：在有多个节点规格时，选择容量最大（可以创建最多Pod）的节点规格'),
      },
      {
        prop: 'bufferResourceCpuRatio',
        name: $i18n.t('触发扩容资源阈值 (CPU)'),
        isBasicProp: true,
        unit: '%',
        desc: $i18n.t('集群整体CPU资源(Request)使用率超过该阈值触发扩容, 无论内存资源使用率是否达到阈值'),
      },
      {
        prop: 'bufferResourceMemRatio',
        name: $i18n.t('触发扩容资源阈值 (内存)'),
        isBasicProp: true,
        unit: '%',
        desc: $i18n.t('集群整体内存资源(Request)使用率超过该阈值触发扩容, 无论CPU资源使用率是否达到阈值'),
      },
      {
        prop: 'bufferResourceRatio',
        name: $i18n.t('触发扩容资源阈值 (Pods)'),
        unit: '%',
        desc: $i18n.t('集群整体内存资源Pod数量使用率超过该阈值触发扩容, 无论CPU / 内存资源使用率是否达到阈值'),
      },
      {
        prop: 'maxNodeProvisionTime',
        name: $i18n.t('等待节点提供最长时间'),
        unit: $i18n.t('秒'),
        desc: $i18n.t('如果节点规格在设置的时间范围内没有提供可用资源，会导致此次自动扩容失败'),
      },
      // {
      //   prop: 'scaleUpFromZero',
      //   name: $i18n.t('(没有ready节点时) 允许自动扩容'),
      // },
    ]);
    const autoScalerDownConfig = ref([
      {
        prop: 'scaleDownUtilizationThreahold',
        name: $i18n.t('触发缩容资源阈值 (CPU/内存)'),
        isBasicProp: true,
        unit: '%',
        desc: $i18n.t('取整范围0% ~ 80%，节点的CPU和内存资源(Request)必须同时低于设定阈值后会驱逐该节点上的Pod执行缩容流程，如果只考虑缩容空节点，可以把此值设置为0'),
      },
      {
        prop: 'scaleDownUnneededTime',
        name: $i18n.t('节点连续空闲'),
        isBasicProp: true,
        unit: $i18n.t('秒'),
        desc: $i18n.t('节点从第一次被标记空闲状态到设定时间内一直处于空闲状态才会被缩容，防止节点资源使用率短时间内波动造成频繁扩缩容操作'),
        suffix: $i18n.t('后执行缩容'),
      },
      {
        prop: 'maxGracefulTerminationSec',
        name: $i18n.t('等待 Pod 退出最长时间'),
        isBasicProp: true,
        unit: $i18n.t('秒'),
        desc: $i18n.t('缩容节点时，等待 pod 停止的最长时间（不会遵守 terminationGracefulPeriodSecond，超时强杀）'),
      },
      {
        prop: 'scaleDownDelayAfterAdd',
        name: $i18n.t('扩容后判断缩容时间间隔'),
        unit: $i18n.t('秒'),
        desc: $i18n.t('扩容节点后多久才继续缩容判断，如果业务自定义初始化任务所需时间比较长，需要适当上调此值'),
      },
      {
        prop: 'scaleDownDelayAfterDelete',
        name: $i18n.t('连续两次缩容时间间隔'),
        unit: $i18n.t('秒'),
        desc: $i18n.t('缩容节点后多久再继续缩容节点，默认设置为0，代表与扩缩容检测时间间隔设置的值相同'),
      },
      // {
      //   prop: 'scaleDownDelayAfterFailure',
      //   name: $i18n.t('缩容失败后重试时间间隔'),
      //   unit: $i18n.t('秒'),
      // },
      {
        prop: 'scaleDownUnreadyTime',
        name: $i18n.t('NotReady节点缩容等待时间'),
        unit: $i18n.t('秒'),
      },
      {
        prop: 'maxEmptyBulkDelete',
        name: $i18n.t('单次缩容最大节点数'),
        unit: $i18n.t('个'),
      },
      {
        prop: 'skipNodesWithLocalStorage',
        name: $i18n.t('允许缩容已使用本地存储的节点'),
        invert: true,
        desc: $i18n.t('如果设置为 “是”，则表示已使用本地存储的节点将允许被缩容，例如已使用过empryDir / HostPath的节点将允许被缩容'),
      },
    ]);
    const getAutoScalerConfig = async () => {
      if (!props.clusterId) return;
      autoscalerData.value = await $store.dispatch('clustermanager/clusterAutoScaling', {
        $clusterId: props.clusterId,
        provider: 'selfProvisionCloud',
      });
      // 方便展示（会污染数据）
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
      if (!autoscalerData.value.module?.scaleOutModuleID) {
        $bkInfo({
          type: 'warning',
          clsName: 'custom-info-confirm',
          title: $i18n.t('弹性伸缩需要配置扩容后转移模块'),
          defaultInfo: true,
          okText: $i18n.t('编辑配置'),
          confirmFn: () => {
            handleEditAutoScaler();
          },
          cancelFn: () => {
            // eslint-disable-next-line prefer-promise-reject-errors
            reject(false);
          },
        });
      } else if (!autoscalerData.value.enableAutoscale
                        && (!nodepoolList.value.length || nodepoolList.value.every(item => !item.enableAutoscale))) {
        // 开启时前置判断是否存在节点规格 或 节点规格都是未开启状态时，要提示至少开启一个
        $bkInfo({
          type: 'warning',
          clsName: 'custom-info-confirm',
          title: !nodepoolList.value.length
            ? $i18n.t('没有检测到可用节点规格，请先创建节点规格')
            : $i18n.t('请至少启用 1 个节点规格的自动扩缩容功能或创建新的节点规格'),
          defaultInfo: true,
          okText: $i18n.t('立即新建'),
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
          provider: 'selfProvisionCloud',
          $clusterId: props.clusterId,
          updater: user.value.username,
        });
        if (result) {
          $bkMessage({
            theme: 'success',
            message: $i18n.t('操作成功'),
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
        title: value ? $i18n.t('确定开启自动缩容配置') : $i18n.t('确定关闭自动缩容配置'),
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
              message: $i18n.t('更新成功'),
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
        curPodsPriority.value = autoscalerData.value?.expendablePodsPriorityCutoff;
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
          message: $i18n.t('修改成功'),
        });
        handleGetAutoScalerConfig();
        showPodsPriorityDialog.value = false;
      }
      podsPriorityLoading.value = false;
    };

    const statusTextMap = { // 节点规格状态
      RUNNING: $i18n.t('正常'),
      CREATING: $i18n.t('创建中'),
      DELETING: $i18n.t('删除中'),
      UPDATING: $i18n.t('更新中'),
      DELETED: $i18n.t('已删除'),
      'CREATE-FAILURE': $i18n.t('创建失败'),
      'UPDATE-FAILURE': $i18n.t('更新失败'),
      'DELETE-FAILURE': $i18n.t('删除失败'),
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
        // 启用
        result = await $store.dispatch('clustermanager/enableNodeGroupAutoScale', {
          $nodeGroupID: row.nodeGroupID,
        });
      }
      if (result) {
        $bkMessage({
          theme: 'success',
          message: $i18n.t('操作成功'),
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
        title: $i18n.t('确定删除节点规格 {name} ', { name: `${row.nodeGroupID}（${row.name}）` }),
        defaultInfo: true,
        confirmFn: async () => {
          const result = await $store.dispatch('clustermanager/deleteNodeGroup', {
            $nodeGroupID: row.nodeGroupID,
            operator: user.value.username,
          });
          if (result) {
            $bkMessage({
              theme: 'success',
              message: $i18n.t('操作成功'),
            });
            handleGetNodePoolList();
          }
        },
      });
    };
    // 节点规格节点数量管理
    const nodeStatusMap = {
      INITIALIZATION: $i18n.t('初始化中'),
      RUNNING: $i18n.t('正常'),
      DELETING: $i18n.t('删除中'),
      DELETED: $i18n.t('已删除'),
      'DELETE-FAILURE': $i18n.t('删除失败'),
      'ADD-FAILURE': $i18n.t('扩容节点失败'),
      'REMOVE-FAILURE': $i18n.t('缩容节点失败'),
      REMOVABLE: $i18n.t('不可调度'),
      NOTREADY: $i18n.t('不正常'),
      UNKNOWN: $i18n.t('未知状态'),
      'REMOVE-CA-FAILURE': $i18n.t('缩容成功, 下架失败'),
    };
    const nodeColorMap = {
      RUNNING: 'green',
      'DELETE-FAILURE': 'red',
      'ADD-FAILURE': 'red',
      'REMOVE-FAILURE': 'red',
      'REMOVE-CA-FAILURE': 'red',
    };
    const nodeListLoading = ref(false);
    const nodeList = ref<any[]>([]);
    const {
      pagination: nodePagination,
      curPageData: nodeCurPageData,
      pageChange: nodePageChange,
      pageSizeChange: nodePageSizeChange,
    } = usePage(nodeList);
    const showNodeManage = ref(false);
    const handleNodeManageCancel = () => {
      currentOperateRow.value = {};
      nodeList.value = [];
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
        title: $i18n.t('确认Pod驱逐'),
        subTitle: $i18n.t('确认要对节点 {ip} 上的Pod进行驱逐', { ip: row.innerIP }),
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
              message: $i18n.t('POD迁移成功'),
            });
            await getNodeList();
          }
          nodeListLoading.value = false;
        },
      });
    };
    const handleDeleteNodeGroupNode = async (row) => {
      if (nodeListLoading.value || !row.unSchedulable) return;

      $bkInfo({
        type: 'warning',
        clsName: 'custom-info-confirm',
        title: $i18n.t('确认删除节点'),
        subTitle: $i18n.t('确认删除节点 {ip}', { ip: row.innerIP }),
        defaultInfo: true,
        confirmFn: async () => {
          // 删除节点组节点
          nodeListLoading.value = true;
          const result = await $store.dispatch('clustermanager/deleteNodeGroupNode', {
            $nodeGroupID: currentOperateRow.value.nodeGroupID,
            nodes: row.innerIP,
            clusterID: props.clusterId,
            operator: user.value.username,
          });
          if (result) {
            $bkMessage({
              theme: 'success',
              message: $i18n.t('操作成功'),
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
          message: $i18n.t('调度成功'),
        });
        await getNodeList();
      }
      nodeListLoading.value = false;
    };

    // 扩缩容记录
    const taskStatusMap = {
      INITIALIZING: $i18n.t('初始化中'),
      RUNNING: $i18n.t('执行中'),
      SUCCESS: $i18n.t('执行成功'),
      FAILURE: $i18n.t('执行失败'),
      TIMEOUT: $i18n.t('执行超时'),
      FORCETERMINATE: $i18n.t('强制终止'),
      NOTSTARTED: $i18n.t('未启动'),
    };
    const taskTypeMap = {
      CreateNodeGroup: $i18n.t('创建节点规格'),
      UpdateNodeGroup: $i18n.t('更新节点规格'),
      DeleteNodeGroup: $i18n.t('删除节点规格'),
      SwitchNodeGroupAutoScaling: $i18n.t('开启/关闭节点规格'),
      UpdateNodeGroupDesiredNode: $i18n.t('扩容节点规格'),
      CleanNodeGroupNodes: $i18n.t('缩容节点规格'),
    };
    const filters = computed(() => ({
      taskType: Object.keys(taskTypeMap).map(key => ({ text: taskTypeMap[key], value: key })),
      status: [
        {
          text: $i18n.t('初始化中'),
          value: 'INITIALIZING',
        },
        {
          text: $i18n.t('执行中'),
          value: 'RUNNING',
        },
        {
          text: $i18n.t('执行成功'),
          value: 'SUCCESS',
        },
        {
          text: $i18n.t('执行失败'),
          value: 'FAILURE',
        },
        {
          text: $i18n.t('执行超时'),
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
        text: $i18n.t('今天'),
        value() {
          const end = new Date();
          const start = new Date(end.getFullYear(), end.getMonth(), end.getDate());
          return [start, end];
        },
      },
      {
        text: $i18n.t('近7天'),
        value() {
          const end = new Date();
          const start = new Date();
          start.setTime(start.getTime() - 3600 * 1000 * 24 * 7);
          return [start, end];
        },
      },
      {
        text: $i18n.t('近15天'),
        value() {
          const end = new Date();
          const start = new Date();
          start.setTime(start.getTime() - 3600 * 1000 * 24 * 15);
          return [start, end];
        },
      },
      {
        text: $i18n.t('近30天'),
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
      handleGetRecordList();
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
      });
    };
    // 集群详情
    const { clusterOS, getClusterDetail } = useClusterInfo();

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

    onMounted(() => {
      handleGetAutoScalerConfig();
      handleGetNodePoolList();
      getClusterDetail(props.clusterId, true);
      handleGetClusterOverview();
    });
    onBeforeUnmount(() => {
      stop();
      stopPoolInterval();
      stopNodeInterval();
      stopTaskPool();
    });
    return {
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
    .nodepool {
        border-top: 1px solid #DCDEE5;
    }
    .disabled {
        color: #C4C6CC;
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
