<!-- eslint-disable max-len -->
<template>
  <div class="cluster-node bcs-content-wrapper pb-[20px]">
    <bcs-alert type="info" class="cluster-node-tip">
      <div slot="title">
        {{$t('集群就绪后，您可以创建命名空间、推送项目镜像到仓库，然后通过服务配置模板集部署服务。')}}
        <i18n
          path="当前集群已添加节点数（含Master） {nodes}，还可添加节点数 {realRemainNodesCount}，当容器网络资源超额使用时，会触发容器网络自动扩容，扩容后最多可以添加 {maxRemainNodesCount} 个节点。"
          v-if="maxRemainNodesCount > 0">
          <span place="nodes" class="num">{{nodesCount}}</span>
          <span place="realRemainNodesCount" class="num">{{realRemainNodesCount}}</span>
          <span place="maxRemainNodesCount" class="num">{{maxRemainNodesCount}}</span>
        </i18n>
      </div>
    </bcs-alert>
    <!-- 操作栏 -->
    <div class="cluster-node-operate">
      <div class="left">
        <template v-if="fromCluster">
          <span v-bk-tooltips="{ disabled: !isImportCluster, content: $t('导入集群，节点管理功能不可用') }">
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
              :disabled="isImportCluster"
              @click="handleAddNode">
              {{$t('添加节点')}}
            </bcs-button>
          </span>
        </template>
        <template v-if="$INTERNAL && curSelectedCluster.providerType === 'tke' && fromCluster">
          <apply-host
            class="mr10"
            :title="$t('申请Node节点')"
            :cluster-id="localClusterId"
            :is-backfill="true" />
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
                <span class="text-[14px]">{{$t('批量')}}</span>
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
            <li @click="handleBatchEnableNodes">{{$t('允许调度')}}</li>
            <li @click="handleBatchStopNodes">{{$t('停止调度')}}</li>
            <li
              :disabled="isImportCluster"
              v-bk-tooltips="{
                disabled: !isImportCluster,
                content: $t('导入集群，节点管理功能不可用')
              }"
              @click="handleBatchReAddNodes">{{$t('失败重试')}}</li>
            <div
              class="h-[32px]"
              v-bk-tooltips="{ content: $t('IP状态为停止调度才能做POD驱逐操作'), disabled: !podDisabled, placement: 'right' }">
              <li :disabled="podDisabled" @click="handleBatchPodScheduler">{{$t('pod驱逐')}}</li>
            </div>
            <li @click="handleBatchSetLabels">{{$t('设置标签')}}</li>
            <div
              class="h-[32px]"
              v-bk-tooltips="{
                content: $t('请先停止节点调度'),
                disabled: !selections.some(item => item.status === 'RUNNING'),
                placement: 'right'
              }">
              <li
                :disabled="isImportCluster || selections.some(item => item.status === 'RUNNING')"
                v-bk-tooltips="{
                  disabled: !isImportCluster,
                  content: !isImportCluster ? $t('导入集群，节点管理功能不可用') : $t('请先停止节点调度')
                }"
                @click="handleBatchDeleteNodes">{{$t('删除')}}</li>
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
              <span class="text-[14px]">{{$t('复制')}}</span>
              <i :class="['bk-icon icon-angle-down', { 'icon-flip': showCopyMenu }]"></i>
            </div>
          </bcs-button>
        </BcsCascade>
      </div>
      <div class="right">
        <ClusterSelect
          class="mr10 w-[254px]"
          v-model="localClusterId"
          @change="handleClusterChange"
          v-if="!hideClusterSelect"
        />
        <bcs-search-select
          clearable
          class="search-select bg-[#fff]"
          :data="searchSelectData"
          :show-condition="false"
          :show-popover-tag-change="false"
          :placeholder="$t('搜索IP、标签、污点、注解、状态、可用区、节点来源、所属节点规格')"
          default-focus
          v-model="searchSelectValue"
          @change="searchSelectChange"
          @clear="handleClearSearchSelect">
        </bcs-search-select>
      </div>
    </div>
    <!-- 节点列表 -->
    <div class="mt-[20px] px-[20px]" v-bkloading="{ isLoading: tableLoading }">
      <bcs-table
        :size="tableSetting.size"
        :data="curPageData"
        ref="tableRef"
        :key="tableKey"
        :pagination="pagination"
        @filter-change="handleFilterChange"
        @page-change="pageChange"
        @page-limit-change="pageSizeChange">
        <template #prepend>
          <transition name="fade">
            <div class="selection-tips" v-if="selectType !== CheckType.Uncheck">
              <i18n path="已选择 {num} 条数据">
                <span place="num" class="tips-num">{{selections.length}}</span>
              </i18n>
              <bk-button
                ext-cls="tips-btn"
                text
                v-if="selectType === CheckType.AcrossChecked"
                @click="handleClearSelection">
                {{ $t('取消选择所有数据') }}
              </bk-button>
              <bk-button
                ext-cls="tips-btn"
                text
                v-else
                @click="handleSelectionAll">
                <i18n path="选择所有 {num} 条">
                  <span place="num" class="tips-num">{{pagination.count}}</span>
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
              :checked="selections.some(item => item.nodeName === row.nodeName)"
              :disabled="!row.nodeName"
              @change="(value) => handleRowCheckChange(value, row)"
            />
          </template>
        </bcs-table-column>
        <bcs-table-column :label="$t('节点名')" min-width="120" prop="nodeName" fixed="left" show-overflow-tooltip>
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
        <bcs-table-column label="IPv4" width="150" prop="innerIP" sortable show-overflow-tooltip></bcs-table-column>
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
          :label="$t('节点来源')"
          :filters="filtersDataSource.nodeSource"
          :filtered-value="filteredValue.nodeSource"
          column-key="nodeSource"
          prop="nodeSource"
          min-width="130"
          v-if="isColumnRender('nodeSource')">
          <template #default="{ row }">
            {{ row.nodeGroupID ? $t('节点规格') : $t('手动添加') }}
          </template>
        </bcs-table-column>
        <bcs-table-column
          :label="$t('所属节点规格')"
          min-width="130"
          show-overflow-tooltip
          v-if="isColumnRender('nodeGroupID')">
          <template #default="{ row }">{{ row.nodeGroupName || '--' }}</template>
        </bcs-table-column>
        <bcs-table-column
          :label="$t('状态')"
          :filters="filtersDataSource.status"
          :filtered-value="filteredValue.status"
          min-width="160"
          column-key="status"
          prop="status"
          show-overflow-tooltip>
          <template #default="{ row }">
            <LoadingIcon
              v-if="['INITIALIZATION', 'DELETING'].includes(row.status)"
            >
              <span class="bcs-ellipsis">{{ nodeStatusMap[row.status.toLowerCase()] }}</span>
            </LoadingIcon>
            <StatusIcon
              :status="row.status"
              :status-color-map="nodeStatusColorMap"
              v-else
            >
              <span class="bcs-ellipsis flex-1">{{ nodeStatusMap[row.status.toLowerCase()] }}</span>
            </StatusIcon>
          </template>
        </bcs-table-column>
        <bcs-table-column
          :label="$t('可用区')"
          :filters="filtersDataSource.zoneID"
          :filtered-value="filteredValue.zoneID"
          min-width="160"
          column-key="zoneID"
          prop="zoneName"
          show-overflow-tooltip
          v-if="isColumnRender('zoneID')">
          <template #default="{ row }">
            {{ row.zoneName || '--' }}
          </template>
        </bcs-table-column>
        <bcs-table-column
          :label="$t('容器数量')"
          min-width="100"
          align="right"
          prop="container_count"
          key="container_count"
          v-if="isColumnRender('container_count')">
          <template #default="{ row }">
            {{
              nodeMetric[row.nodeName]
                ? nodeMetric[row.nodeName].container_count || '--'
                : '--'
            }}
          </template>
        </bcs-table-column>
        <bcs-table-column
          :label="$t('Pod数量')"
          min-width="100"
          align="right"
          prop="pod_count"
          key="pod_count"
          v-if="isColumnRender('pod_count')">
          <template #default="{ row }">
            {{
              nodeMetric[row.nodeName]
                ? nodeMetric[row.nodeName].pod_count || '--'
                : '--'
            }}
          </template>
        </bcs-table-column>
        <bcs-table-column min-width="200" :label="$t('标签')" key="labels" v-if="isColumnRender('labels')">
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
        <bcs-table-column min-width="200" :label="$t('污点')" key="taint" v-if="isColumnRender('taint')">
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
        <bcs-table-column min-width="200" :label="$t('注解')" key="annotations" v-if="isColumnRender('annotations')">
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
          :sort-method="(pre, next) => sortMethod(pre, next, item.prop)"
          :key="item.prop"
          sortable
          align="center"
          min-width="120">
          <template #default="{ row }">
            <template v-if="['RUNNING', 'REMOVABLE'].includes(row.status)">
              <LoadingCell v-if="!nodeMetric[row.nodeName]" />
              <RingCell
                :percent="nodeMetric[row.nodeName][item.prop]"
                :fill-color="item.color"
                v-if="nodeMetric[row.nodeName]" />
            </template>
            <template v-else>--</template>
          </template>
        </bcs-table-column>
        <bcs-table-column :label="$t('操作')" width="160" :resizable="false" fixed="right">
          <template #default="{ row }">
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
              v-if="row.status !== 'REMOVE-CA-FAILURE'">
              <bk-button
                class="mr10"
                text
                @click="handleStopNode(row)"
                v-if="row.status === 'RUNNING'">
                {{ $t('停止调度') }}
              </bk-button>
              <template v-else-if="row.status === 'REMOVABLE'">
                <bk-button text class="mr10" @click="handleEnableNode(row)">
                  {{ $t('允许调度') }}
                </bk-button>
                <bk-button text class="mr10" @click="handleSchedulerNode(row)">
                  {{ $t('pod驱逐') }}
                </bk-button>
              </template>
              <bk-button
                class="mr10"
                text
                v-if="['INITIALIZATION', 'DELETING', 'REMOVE-FAILURE', 'ADD-FAILURE'].includes(row.status)"
                :disabled="!row.inner_ip"
                @click="handleShowLog(row)"
              >
                {{$t('查看日志')}}
              </bk-button>
              <bk-button
                text
                class="mr10"
                v-if="['REMOVE-FAILURE', 'ADD-FAILURE'].includes(row.status)"
                :disabled="!row.inner_ip"
                @click="handleRetry(row)"
              >{{ $t('重试') }}</bk-button>
              <bk-popover
                placement="bottom"
                theme="light dropdown"
                :arrow="false"
                :disabled="['INITIALIZATION', 'DELETING'].includes(row.status)"
                trigger="click">
                <span :class="['bcs-icon-more-btn', { disabled: ['INITIALIZATION', 'DELETING'].includes(row.status) }]">
                  <i class="bcs-icon bcs-icon-more"></i>
                </span>
                <template #content>
                  <ul class="bcs-dropdown-list">
                    <template v-if="row.status === 'RUNNING'">
                      <li class="bcs-dropdown-item" @click="handleSetLabel(row)">
                        {{$t('设置标签')}}
                      </li>
                      <li class="bcs-dropdown-item" @click="handleSetTaint(row)">
                        {{$t('设置污点')}}
                      </li>
                    </template>
                    <li
                      :class="['bcs-dropdown-item', { disabled: isImportCluster }]"
                      v-bk-tooltips="{
                        disabled: !isImportCluster,
                        content: $t('导入集群，节点管理功能不可用')
                      }"
                      v-if="['REMOVE-FAILURE', 'ADD-FAILURE', 'REMOVABLE', 'NOTREADY'].includes(row.status)"
                      :disabled="!row.inner_ip"
                      @click="handleDeleteNode(row)"
                    >
                      {{ $t('删除') }}
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
              {{$t('查看日志')}}
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
    <!-- 设置标签 -->
    <bcs-sideslider
      :is-show.sync="setLabelConf.isShow"
      :width="750"
      :before-close="handleBeforeClose"
      quick-close>
      <template #header>
        <span>{{setLabelConf.title}}</span>
        <span class="sideslider-tips">{{$t('标签有助于整理你的资源')}}</span>
      </template>
      <template #content>
        <KeyValue
          class="key-value-content"
          :model-value="setLabelConf.data"
          :loading="setLabelConf.btnLoading"
          :key-desc="setLabelConf.keyDesc"
          :key-rules="[
            {
              message: $i18n.t('仅支持字母，数字和字符(-_./)，且需以字母数字开头和结尾'),
              validator: KEY_REGEXP
            }
          ]"
          :value-rules="[
            {
              message: $i18n.t('仅支持字母，数字和字符(-_./)，且需以字母数字开头和结尾'),
              validator: VALUE_REGEXP
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
      :title="$t('设置污点')"
      :width="750"
      :before-close="handleBeforeClose"
      quick-close>
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
    <!-- 查看日志 -->
    <bk-sideslider
      :is-show.sync="logSideDialogConf.isShow"
      :title="logSideDialogConf.title"
      :width="860"
      @hidden="closeLog"
      :quick-close="true">
      <div slot="content">
        <div class="log-wrapper" v-bkloading="{ isLoading: logSideDialogConf.loading }">
          <TaskList :data="logSideDialogConf.taskData"></TaskList>
        </div>
      </div>
    </bk-sideslider>
    <!-- 确认删除 -->
    <ConfirmDialog
      v-model="showConfirmDialog"
      :title="removeNodeDialogTitle"
      :sub-title="$t('此操作无法撤回，请确认：')"
      :tips="deleteNodeNoticeList"
      :ok-text="$t('删除')"
      :cancel-text="$t('关闭')"
      :confirm="confirmDelNode"
      @ancel="cancelDelNode" />
    <!-- IP选择器 -->
    <IpSelector v-model="showIpSelector" @confirm="chooseServer"></IpSelector>
  </div>
</template>
<script lang="ts">
import { defineComponent, ref, onMounted, watch, set, computed } from 'vue';
import StatusIcon from '@/components/status-icon';
import ClusterSelect from '@/components/cluster-selector/cluster-select.vue';
import LoadingIcon from '@/components/loading-icon.vue';
import { KEY_REGEXP, VALUE_REGEXP } from '@/common/constant';
import useNode from './use-node';
import useTableSetting from '../../../composables/use-table-setting';
import usePage from '@/composables/use-page';
import useTableSearchSelect, { ISearchSelectData } from '../../../composables/use-table-search-select';
import useTableAcrossCheck from '../../../composables/use-table-across-check';
import { CheckType } from '@/components/across-check.vue';
import RingCell from '@/views/cluster-manage/components/ring-cell.vue';
import LoadingCell from '@/views/cluster-manage/components/loading-cell.vue';
import { copyText, padIPv6 } from '@/common/util';
import useInterval from '@/composables/use-interval';
import KeyValue, { IData } from '@/components/key-value.vue';
import TaintContent from '../components/taint.vue';
import ConfirmDialog from '@/components/comfirm-dialog.vue';
import ApplyHost from '@/views/cluster-manage/components/apply-host.vue';
import { TranslateResult } from 'vue-i18n';
import IpSelector from '@/components/ip-selector/selector-dialog.vue';
import TaskList from '../components/task-list.vue';
import { ICluster, useCluster } from '@/composables/use-app';
import BcsCascade from '@/components/cascade.vue';
import useSideslider from '@/composables/use-sideslider';
import $bkInfo from '@/components/bk-magic-2.0/bk-info';
import $bkMessage from '@/common/bkmagic';
import $store from '@/store';
import $router from '@/router';
import $i18n from '@/i18n/i18n-setup';

export default defineComponent({
  name: 'NodeList',
  components: {
    StatusIcon,
    LoadingIcon,
    ClusterSelect,
    RingCell,
    LoadingCell,
    KeyValue,
    TaintContent,
    ConfirmDialog,
    ApplyHost,
    IpSelector,
    TaskList,
    BcsCascade,
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

    const { reset, setChanged, handleBeforeClose } = useSideslider();
    const nodeStatusColorMap = {
      initialization: 'blue',
      running: 'green',
      deleting: 'blue',
      'add-failure': 'red',
      'remove-failure': 'red',
      'remove-ca-failure': 'red',
      removable: '',
      notready: 'red',
      unknown: '',
    };
    const nodeStatusMap = {
      initialization: window.i18n.t('初始化中'),
      running: window.i18n.t('正常'),
      deleting: window.i18n.t('删除中'),
      'add-failure': window.i18n.t('上架失败'),
      'remove-failure': window.i18n.t('下架失败'),
      'remove-ca-failure': window.i18n.t('缩容成功,下架失败'),
      removable: window.i18n.t('不可调度'),
      notready: window.i18n.t('不正常'),
      unknown: window.i18n.t('未知状态'),
    };
    // 表格表头搜索项配置
    const filtersDataSource = computed(() => ({
      status: Object.keys(nodeStatusMap).map(key => ({
        text: nodeStatusMap[key],
        value: key.toUpperCase(),
      })),
      nodeSource: [
        {
          text: $i18n.t('手动添加'),
          value: 'custom',
        },
        {
          text: $i18n.t('节点规格'),
          value: 'nodepool',
        },
      ],
      zoneID: zoneList.value,
    }));
    // 表格搜索项选中值
    const filteredValue = ref<Record<string, string[]>>({
      status: [],
      nodeSource: [],
      zoneID: [],
    });
    // searchSelect数据源配置
    const searchSelectDataSource = computed<ISearchSelectData[]>(() => [
      {
        name: $i18n.t('IP地址'),
        id: 'ip',
        placeholder: $i18n.t('多IP用空格符分割'),
      },
      {
        name: $i18n.t('状态'),
        id: 'status',
        multiable: true,
        children: Object.keys(nodeStatusMap).map(key => ({
          id: key.toUpperCase(),
          name: nodeStatusMap[key],
        })),
      },
      {
        name: $i18n.t('可用区'),
        id: 'zoneID',
        multiable: true,
        children: zoneList.value,
      },
      {
        name: $i18n.t('标签'),
        id: 'labels',
        multiable: true,
        children: labels.value.map(label => ({
          id: label,
          name: label,
        })),
      },
      {
        name: $i18n.t('污点'),
        id: 'taints',
        multiable: true,
        children: taints.value,
      },
      {
        name: $i18n.t('注解'),
        id: 'annotations',
        multiable: true,
        children: annotations.value.map(label => ({
          id: label,
          name: label,
        })),
      },
      {
        name: $i18n.t('节点来源'),
        id: 'nodeSource',
        multiable: true,
        children: [
          {
            id: 'custom',
            name: $i18n.t('手动添加'),
          },
          {
            id: 'nodepool',
            name: $i18n.t('节点规格'),
          },
        ],
      },
      {
        name: $i18n.t('所属节点规格'),
        id: 'nodeGroupID',
        multiable: true,
        children: tableData.value.reduce<any[]>((pre, item) => {
          if (item.nodeGroupID && pre.every(data => data.id !== item.nodeGroupID)) {
            pre.push({
              id: item.nodeGroupID,
              name: item.nodeGroupName,
            });
          }
          return pre;
        }, []),
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
    });

    // 表格设置字段配置
    const fields = [
      {
        id: 'innerIPv6',
        label: 'IPv6',
      },
      {
        id: 'nodeSource',
        label: $i18n.t('节点来源'),
        defaultChecked: true,
      },
      {
        id: 'nodeGroupID',
        label: $i18n.t('所属节点规格'),
        defaultChecked: true,
      },
      {
        id: 'zoneID',
        label: $i18n.t('可用区'),
        defaultChecked: true,
      },
      {
        id: 'container_count',
        label: $i18n.t('容器数量'),
        disabled: true,
      },
      {
        id: 'pod_count',
        label: $i18n.t('Pod数量'),
        disabled: true,
      },
      {
        id: 'labels',
        label: $i18n.t('标签'),
      },
      {
        id: 'taint',
        label: $i18n.t('污点'),
      },
      {
        id: 'annotations',
        label: $i18n.t('注解'),
      },
      {
        id: 'cpu_usage',
        label: $i18n.t('CPU使用率'),
        disabled: true,
      },
      {
        id: 'memory_usage',
        label: $i18n.t('内存使用率'),
        disabled: true,
      },
      {
        id: 'disk_usage',
        label: $i18n.t('磁盘使用率'),
        disabled: true,
      },
      {
        id: 'diskio_usage',
        label: $i18n.t('磁盘IO使用率'),
        disabled: true,
      },
    ];
    // 表格指标列配置
    const metricColumnConfig = ref([
      {
        label: $i18n.t('CPU使用率'),
        prop: 'cpu_usage',
        color: '#3ede78',
      },
      {
        label: $i18n.t('内存使用率'),
        prop: 'memory_usage',
        color: '#3a84ff',
      },
      {
        label: $i18n.t('CPU装箱率'),
        prop: 'cpu_request_usage',
        color: '#3ede78',
      },
      {
        label: $i18n.t('内存装箱率'),
        prop: 'memory_request_usage',
        color: '#3a84ff',
      },
      {
        label: $i18n.t('磁盘使用率'),
        prop: 'disk_usage',
        color: '#853cff',
      },
      {
        label: $i18n.t('磁盘IO'),
        prop: 'diskio_usage',
        color: '#853cff',
      },
    ]);
    const {
      tableSetting,
      handleSettingChange,
      isColumnRender,
    } = useTableSetting(fields);

    const sortMethod = (pre, next, prop) => {
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
      schedulerNode,
      deleteNode,
      addNode,
      getNodeOverview,
      retryTask,
      setNodeLabels,
    } = useNode();

    const tableLoading = ref(false);
    const localClusterId = ref(props.clusterId || $store.getters.curClusterId);
    const { clusterList } = useCluster();
    const curSelectedCluster = computed<Partial<ICluster>>(() => clusterList.value
      .find(item => item.clusterID === localClusterId.value) || {});
    // 导入集群
    const isImportCluster = computed(() => curSelectedCluster.value.clusterCategory === 'importer');
    // 全量表格数据
    const tableData = ref<any[]>([]);

    // 可用区
    const zoneList = computed(() => tableData.value.reduce((pre, row) => {
      if (!row.zoneID) return pre;
      const data = pre.find(item => item.value === row.zoneID);
      if (!data) {
        pre.push({
          value: row.zoneID,
          text: `${row.zoneName} (1)`,
          id: row.zoneID,
          name: `${row.zoneName} (1)`,
          count: 1,
        });
      } else {
        data.count += 1;
        data.text = `${row.zoneName} (${data.count})`;
        data.name = `${row.zoneName} (${data.count})`;
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
              const splitCode = String(v).indexOf('|') > -1 ? '|' : ' ';
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
          value: new Set(tmp.map(t => padIPv6(t))),
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
      .filter(item => !!item.nodeName));
    const filterFailureCurTableData = computed(() => curPageData.value.filter(item => !!item.nodeName));
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
        label: $i18n.t('复制勾选IP'),
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
        label: $i18n.t('复制所有IP'),
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
        message: $i18n.t('成功复制 {num} 个IP', { num: ipData.length }),
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

    // 设置标签（批量设置标签的交互有点奇怪，后续优化）
    const setLabelConf = ref<{
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
    const handleSetLabel = async (selections: Record<string, any>[] | Record<string, any>) => {
      setLabelConf.value.isShow = true;
      const rows = Array.isArray(selections) ? selections : [selections];
      // 批量设置时暂时只展示相同Key的项
      const labelArr = rows.reduce<any[]>((pre, row) => {
        const label = row.labels;
        Object.keys(label).forEach((key) => {
          const index = pre.findIndex(item => item.key === key);
          if (index > -1) {
            pre[index].value = '';
            pre[index].repeat += 1;
            pre[index].placeholder = $i18n.t('不变');
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

      set(setLabelConf, 'value', Object.assign(setLabelConf.value, {
        data: labelArr,
        rows,
        title: rows.length > 1 ? $i18n.t('批量设置标签') : $i18n.t('设置标签'),
        keyDesc: rows.length > 1 ? $i18n.t('批量设置只展示相同Key的标签') : '',
      }));
      reset();
    };
    const handleLabelEditCancel = () => {
      set(setLabelConf, 'value', Object.assign(setLabelConf.value, {
        isShow: false,
        keyDesc: '',
        rows: [],
        title: '',
        data: {},
      }));
    };
    const mergeLaels = (_originLabels, _newLabels) => {
      const originLabels = JSON.parse(JSON.stringify(_originLabels));
      const newLabels = JSON.parse(JSON.stringify(_newLabels));
      // 批量编辑
      if (setLabelConf.value.rows.length > 1) {
        const oldLabels = setLabelConf.value.data;
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
      setLabelConf.value.btnLoading = true;
      const result = await setNodeLabels({
        clusterID: localClusterId.value,
        nodes: setLabelConf.value.rows.map(item => ({
          nodeName: item.nodeName,
          labels: mergeLaels(item.labels, labels),
        })),
      });
      setLabelConf.value.btnLoading = false;
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
      subTitle: TranslateResult;
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
        title: $i18n.t('确认对节点 {ip} 停止调度', { ip: row.nodeName }),
        subTitle: $i18n.t('如果有使用Ingress及LoadBalancer类型的Service，节点停止调度后，Service Controller会剔除LB到nodePort的映射'),
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
        title: $i18n.t('确认允许调度'),
        subTitle: $i18n.t('确认对节点 {ip} 允许调度', { ip: row.nodeName }),
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
    const handleSchedulerNode = (row) => {
      bkComfirmInfo({
        title: $i18n.t('确认Pod驱逐'),
        subTitle: $i18n.t('确认要对节点 {ip} 上的Pod进行驱逐', { ip: row.nodeName }),
        callback: async () => {
          await schedulerNode({
            clusterId: row.cluster_id,
            nodes: [row.nodeName],
          });
          // result && handleGetNodeData()
        },
      });
    };
    // 节点删除
    const showConfirmDialog = ref(false);
    const deleteNodeNoticeList = ref([
      $i18n.t('当前节点上正在运行的容器会被调度到其它可用节点'),
      $i18n.t('清理容器服务系统组件'),
      $i18n.t('节点删除后服务器如不再使用请尽快回收，避免产生不必要的成本'),
    ]);
    const curDeleteRows = ref<any[]>([]);
    const removeNodeDialogTitle = ref<any>('');
    const handleDeleteNode = async (row) => {
      if (isImportCluster.value) return;
      curDeleteRows.value = [row];
      removeNodeDialogTitle.value = $i18n.t('确认要删除节点【{innerIp}】？', {
        innerIp: row.inner_ip,
      });
      showConfirmDialog.value = true;
    };
    const cancelDelNode = () => {
      curDeleteRows.value = [];
    };
    const delNode = async (clusterId: string, nodeIps: string[]) => {
      const result = await deleteNode({
        clusterId,
        nodeIps,
      });
      result && handleGetNodeData();
      handleResetPage();
      handleResetCheckStatus();
    };
    const confirmDelNode = async () => {
      await delNode(localClusterId.value, curDeleteRows.value.map(item => item.inner_ip));
    };
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
    const handleRetry = async (row) => {
      tableLoading.value = true;
      await retryTask({
        clusterId: row.cluster_id,
        nodeIP: row.inner_ip,
      });
      tableLoading.value = false;
    };
    // 批量允许调度
    const showBatchMenu = ref(false);
    const handleBatchEnableNodes = () => {
      if (!selections.value.length) return;

      bkComfirmInfo({
        title: $i18n.t('请确认是否批量允许调度'),
        subTitle: $i18n.t('请确认是否允许 {ip} 等 {num} 个IP调度', {
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
        title: $i18n.t('请确认是否批量停止调度'),
        subTitle: $i18n.t('请确认是否停止 {ip} 等 {num} 个IP调度', {
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
      if (!selections.value.length || isImportCluster.value) return;

      bkComfirmInfo({
        title: $i18n.t('确认重新添加节点'),
        subTitle: $i18n.t('请确认是否对 {ip} 等 {num} 个IP进行操作系统初始化和安装容器服务相关组件操作', {
          num: selections.value.length,
          ip: selections.value[0].inner_ip }),
        callback: async () => {
          await addClusterNode(localClusterId.value, selections.value.map(item => item.inner_ip));
        },
      });
    };
    // 批量设置标签
    const handleBatchSetLabels = () => {
      if (!selections.value.length) return;

      handleSetLabel(selections.value);
    };
    // 批量删除节点
    const handleBatchDeleteNodes = () => {
      if (isImportCluster.value) return;
      bkComfirmInfo({
        title: $i18n.t('确认删除节点'),
        subTitle: $i18n.t('确认是否删除 {ip} 等 {num} 个节点', {
          num: selections.value.length,
          ip: selections.value[0].inner_ip,
        }),
        callback: async () => {
          await delNode(localClusterId.value, selections.value.map(item => item.inner_ip));
        },
      });
    };
    // 批量Pod驱逐
    const handleBatchPodScheduler = () => {
      if (!selections.value.length) return;

      if (selections.value.length > 10) {
        $bkMessage({
          theme: 'warning',
          message: $i18n.t('最多只能批量驱逐10个节点'),
        });
        return;
      }
      bkComfirmInfo({
        title: $i18n.t('确认Pod驱逐'),
        subTitle: $i18n.t('确认要对 {ip} 等 {num} 个节点上的Pod进行驱逐', {
          num: selections.value.length,
          ip: selections.value[0].nodeName,
        }),
        callback: async () => {
          await schedulerNode({
            clusterId: localClusterId.value,
            nodes: selections.value.map(item => item.nodeName),
          });
          // result && handleGetNodeData()
        },
      });
    };
    // 添加节点
    const showIpSelector = ref(false);
    const handleAddNode = () => {
      $router.push({
        name: 'addClusterNode',
        params: {
          clusterId: props.clusterId,
        },
      });
    };
    const chooseServer = (data) => {
      if (!data.length) return;
      bkComfirmInfo({
        title: $i18n.t('确认添加节点'),
        subTitle: $i18n.t('请确认是否对 {ip} 等 {num} 个IP进行操作系统初始化和安装容器服务相关组件操作', {
          ip: data[0].bk_host_innerip,
          num: data.length,
        }),
        callback: async () => {
          await addClusterNode(localClusterId.value, data.map(item => item.bk_host_innerip));
          showIpSelector.value = false;
        },
      });
    };
    // 查看日志
    const logSideDialogConf = ref({
      isShow: false,
      title: '',
      taskData: [],
      row: null,
      loading: false,
    });
    const handleShowLog = async (row) => {
      logSideDialogConf.value.isShow = true;
      logSideDialogConf.value.title = row.inner_ip;
      logSideDialogConf.value.row = row;
      logSideDialogConf.value.loading = true;
      await getTaskTableData(row);
      logSideDialogConf.value.loading = false;
    };
    const getTaskTableData = async (row) => {
      const { taskData, latestTask } = await getTaskData({
        clusterId: row.cluster_id,
        nodeIP: row.inner_ip,
      });
      logSideDialogConf.value.taskData = taskData || [];
      if (['RUNNING', 'INITIALZING'].includes(latestTask?.status)) {
        logIntervalStart();
      } else {
        logIntervalStop();
      }
    };
    const { stop: logIntervalStop, start: logIntervalStart } = useInterval(async () => {
      const row = logSideDialogConf.value.row as any;
      if (!row) {
        logIntervalStop();
        return;
      }
      const { taskData, latestTask } = await getTaskData({
        clusterId: row.cluster_id,
        nodeIP: row.inner_ip,
      });
      logSideDialogConf.value.taskData = taskData || [];
      if (!['RUNNING', 'INITIALZING'].includes(latestTask?.status)) {
        logIntervalStop();
      }
    }, 5000);
    const closeLog = () => {
      logSideDialogConf.value.row = null;
      logIntervalStop();
    };

    // 获取节点指标
    const nodeMetric = ref({});
    const handleGetNodeOverview = async () => {
      const data = curPageData.value.filter(item => !nodeMetric.value[item.nodeName]
        && ['RUNNING', 'REMOVABLE'].includes(item.status));
      const promiseList: Promise<any>[] = [];
      for (const row of data) {
        (function (item) {
          promiseList.push(getNodeOverview({
            nodeIP: item.nodeName,
            clusterId: localClusterId.value,
          }).then((data) => {
            set(nodeMetric.value, item.nodeName, data);
          }));
        }(row));
      }
      await Promise.all(promiseList);
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
    const podDisabled = computed(() => !selections.value.every(select => select.status === 'REMOVABLE'));

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

    // eslint-disable-next-line max-len
    const nodesCount = computed(() => tableData.value.length + Object.keys(curSelectedCluster.value?.master || {}).length);
    const getCidrIpNum = (cidr) => {
      const mask = Number(cidr.split('/')[1] || 0);
      if (mask <= 0) {
        return 0;
      }
      return Math.pow(2, 32 - mask);
    };
    // 当前CIDR可添加节点数
    const realRemainNodesCount = computed(() => {
      const {
        maxNodePodNum,
        maxServiceNum,
        clusterIPv4CIDR,
        multiClusterCIDR = [],
      } = curSelectedCluster.value?.networkSettings || {};
      const totalCidrStep = [clusterIPv4CIDR, ...multiClusterCIDR].reduce((pre, cidr) => {
        pre += getCidrIpNum(cidr);
        return pre;
      }, 0);
      return Math.floor((totalCidrStep - maxServiceNum - maxNodePodNum * nodesCount.value) / maxNodePodNum);
    });
    // 扩容后最大节点数量
    const maxRemainNodesCount = computed(() => {
      const {
        cidrStep,
        maxNodePodNum,
        maxServiceNum,
        clusterIPv4CIDR,
        multiClusterCIDR = [],
      } = curSelectedCluster.value?.networkSettings || {};
      let totalCidrStep = 0;
      if (multiClusterCIDR.length < 3) {
        totalCidrStep = (5 - multiClusterCIDR.length) * cidrStep + multiClusterCIDR.reduce((pre, cidr) => {
          pre += getCidrIpNum(cidr);
          return pre;
        }, 0);
      } else {
        totalCidrStep = [clusterIPv4CIDR, ...multiClusterCIDR].reduce((pre, cidr) => {
          pre += getCidrIpNum(cidr);
          return pre;
        }, 0);
      }
      return Math.floor((totalCidrStep - maxServiceNum - maxNodePodNum * nodesCount.value) / maxNodePodNum);
    });

    onMounted(async () => {
      await handleGetNodeData();
      if (tableData.value.length) {
        start();
      }
    });
    return {
      showConfirmDialog,
      copyList,
      metricColumnConfig,
      removeNodeDialogTitle,
      nodesCount,
      realRemainNodesCount,
      maxRemainNodesCount,
      curSelectedCluster,
      logSideDialogConf,
      showIpSelector,
      deleteNodeNoticeList,
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
      setLabelConf,
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
      handleSetLabel,
      sortMethod,
      handleLabelEditCancel,
      handleLabelEditConfirm,
      handleConfirmTaintDialog,
      handleHideTaintDialog,
      handleSetTaint,
      handleEnableNode,
      handleStopNode,
      handleDeleteNode,
      handleSchedulerNode,
      confirmDelNode,
      cancelDelNode,
      handleRetry,
      handleBatchEnableNodes,
      handleBatchStopNodes,
      handleBatchReAddNodes,
      handleBatchSetLabels,
      handleBatchDeleteNodes,
      handleAddNode,
      chooseServer,
      handleClusterChange,
      handleShowLog,
      closeLog,
      handleBatchPodScheduler,
      podDisabled,
      webAnnotations,
      curProject,
      isImportCluster,
      KEY_REGEXP,
      VALUE_REGEXP,
      showBatchMenu,
      showCopyMenu,
      setChanged,
      handleBeforeClose,
    };
  },
});
</script>
<style lang="postcss" scoped>
.cluster-node-tip {
    margin: 20px;
    .num {
        font-weight: 700;
    }
}
.cluster-node-operate {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 0 20px;
    .left {
        display: flex;
        align-items: center;
        .add-node {
            min-width: 120px;
        }
    }
    .right {
        display: flex;
        .search-select {
            width: 460px;
        }
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
