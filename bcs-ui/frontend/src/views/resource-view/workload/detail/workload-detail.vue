<template>
  <div class="workload-detail" v-bkloading="{ isLoading }">
    <div class="workload-detail-info">
      <div class="workload-info-basic">
        <div class="basic-left">
          <span class="name mr20 select-all">{{ metadata.name }}</span>
          <div class="basic-wrapper">
            <div
              v-for="item in basicInfoList"
              :key="item.label"
              class="basic-item">
              <span class="label">{{ item.label }}</span>
              <span class="value">{{ item.value }}</span>
            </div>
            <div
              v-for="item in additionalColumns"
              :key="item.name"
              class="basic-item">
              <span class="label">{{ item.name }}</span>
              <span class="value">{{ getJsonPathValue(detail && detail.manifest, item.jsonPath) }}</span>
            </div>
          </div>
        </div>
        <div class="btns">
          <bk-button theme="primary" @click="handleShowYamlPanel">{{ $t('dashboard.workload.button.yaml') }}</bk-button>
          <template v-if="!hiddenOperate">
            <bk-button
              theme="primary"
              @click="handleUpdateResource">{{$t('generic.button.update')}}</bk-button>
            <bk-button
              theme="danger"
              v-authority="{
                actionId: 'namespace_scoped_delete',
                clickable: hasPerms,
                content: pagePerms.deleteBtn.tip,
                disablePerms: true,
                resourceName: metadata.name,
                permCtx: {
                  resource_type: 'namespace',
                  name: metadata.namespace,
                  project_id: curProject.project_id,
                  cluster_id: clusterId,
                }
              }"
              @click="handleDeleteResource">{{$t('generic.button.delete')}}</bk-button>
          </template>
        </div>
      </div>
      <div class="workload-main-info">
        <div class="info-item">
          <span class="label">{{ $t('cluster.labels.name') }}</span>
          <span class="value">{{ clusterNameMap[clusterId] }}</span>
        </div>
        <div class="info-item">
          <span class="label">{{ $t('k8s.namespace') }}</span>
          <span class="value select-all">{{ metadata.namespace }}</span>
        </div>
        <div class="info-item">
          <span class="label">{{ $t('k8s.image') }}</span>
          <span class="value select-all" v-bk-overflow-tips="getImagesTips(manifestExt.images)">
            {{ manifestExt.images && manifestExt.images.join(', ') }}</span>
        </div>
        <div class="info-item">
          <span class="label">UID</span>
          <span class="value select-all">{{ metadata.uid }}</span>
        </div>
        <div class="info-item">
          <span class="label">{{ $t('cluster.labels.createdAt') }}</span>
          <span class="value">{{ timeFormat(manifestExt.createTime) }}</span>
        </div>
        <div class="info-item">
          <span class="label">{{ $t('k8s.age') }}</span>
          <span class="value">{{ manifestExt.age }}</span>
        </div>
        <div class="info-item" v-if="showUpdateStrategy">
          <span class="label">{{ $t('k8s.updateStrategy.text') }}</span>
          <span class="value">
            {{ updateStrategyMap[updateStrategy.type] || $t('k8s.updateStrategy.rollingUpdate') }}
          </span>
        </div>
        <template v-if="category === 'deployments'">
          <div class="info-item">
            <span class="label">{{ $t('k8s.deployment.maxSurge') }}</span>
            <span class="value" v-if="$chainable(spec, 'strategy.rollingUpdate.maxSurge')">
              {{ spec.strategy.rollingUpdate.maxSurge }}
            </span>
            <span class="value" v-else>--</span>
          </div>
          <div class="info-item">
            <span class="label">{{ $t('k8s.deployment.maxUnavailable') }}</span>
            <span class="value" v-if="$chainable(spec, 'strategy.rollingUpdate.maxUnavailable')">
              {{ spec.strategy.rollingUpdate.maxUnavailable }}
            </span>
            <span class="value" v-else>--</span>
          </div>
          <div class="info-item">
            <span class="label">{{ $t('k8s.deployment.minReadySeconds') }}</span>
            <span class="value" v-if="Number.isInteger(spec.minReadySeconds)">{{ spec.minReadySeconds }}s</span>
            <span class="value" v-else>--</span>
          </div>
          <div class="info-item">
            <span class="label">{{ $t('k8s.deployment.progressDeadlineSeconds') }}</span>
            <span class="value" v-if="Number.isInteger(spec.progressDeadlineSeconds)">
              {{ spec.progressDeadlineSeconds }}s</span>
            <span class="value" v-else>--</span>
          </div>
        </template>
        <template v-else-if="category === 'statefulsets'">
          <div class="info-item">
            <span class="label">{{ $t('dashboard.workload.pods.podManagementPolicy') }}</span>
            <span class="value">{{ spec.podManagementPolicy || '--' }}</span>
          </div>
          <div class="info-item">
            <span class="label">{{ $t('k8s.statefulset.partition') }}</span>
            <span class="value" v-if="Number.isInteger($chainable(spec, 'updateStrategy.rollingUpdate.partition'))">
              {{ spec.updateStrategy.rollingUpdate.partition }}s</span>
            <span class="value" v-else>--</span>
          </div>
        </template>
      </div>
    </div>
    <div class="workload-detail-body">
      <div class="workload-metric" v-bkloading="{ isLoading: podLoading }">
        <Metric
          :title="$t('metrics.cpuUsage')"
          metric="cpu_usage"
          :params="params"
          category="pods"
          colors="#30d878"
          unit="percent-number">
        </Metric>
        <Metric
          :title="$t('metrics.memUsage1')"
          metric="memory_used"
          :params="params"
          unit="byte"
          category="pods"
          colors="#3a84ff"
          :desc="$t('dashboard.workload.tips.containerMemoryWorkingSetBytesOom')">
        </Metric>
        <Metric
          :title="$t('k8s.networking')"
          :metric="['network_receive', 'network_transmit']"
          :params="params"
          category="pods"
          unit="byte"
          :colors="['#853cff', '#30d878']"
          :suffix="[$t('metrics.network.receive'), $t('metrics.network.transmit')]">
        </Metric>
      </div>
      <bcs-tab class="workload-tab" :active.sync="activePanel" type="card" :label-height="42">
        <bcs-tab-panel name="pod" label="Pods" v-bkloading="{ isLoading: podLoading }" render-directive="if">
          <div class="pod-info-header">
            <!-- <bk-button
              v-if="showBatchDispatch"
              :loading="batchBtnLoading"
              :disabled="!selectPods.length"
              @click="handelShowRescheduleDialog">
              {{ $t('dashboard.workload.pods.multiDelete') }}
            </bk-button> -->
            <!-- 占位 -->
            <div></div>
            <bk-input
              v-model="searchValue"
              :placeholder="$t('dashboard.workload.pods.search')"
              class="search-input"
              right-icon="bk-icon icon-search"
              clearable>
            </bk-input>
          </div>
          <bcs-table
            :data="curPodTablePageData"
            ref="podTableRef"
            row-key="metadata.uid"
            :pagination="podTablePagination"
            @page-change="podTablePageChang"
            @page-limit-change="podTablePageSizeChange"
            @select="handleSelectPod"
            @select-all="handleSelectAllPod"
            @sort-change="handleSortChange"
            @filter-change="handleFilterChange"
          >
            <!-- <bcs-table-column
              v-if="showBatchDispatch"
              type="selection"
              width="60"
              reserve-selection
              :selectable="handlePodSelectable">
            </bcs-table-column> -->
            <bcs-table-column
              :label="$t('generic.label.name')"
              min-width="130" prop="metadata.name" sortable show-overflow-tooltip fixed="left">
              <template #default="{ row }">
                <bk-button
                  class="bcs-button-ellipsis"
                  text @click="gotoPodDetail(row)">{{ row.metadata.name }}</bk-button>
              </template>
            </bcs-table-column>
            <bcs-table-column :label="$t('k8s.image')" min-width="200" :show-overflow-tooltip="false">
              <template #default="{ row }">
                <span
                  class="select-all"
                  v-bk-tooltips.top="(handleGetExtData(row.metadata.uid, 'images') || []).join('<br />')">
                  {{ (handleGetExtData(row.metadata.uid, 'images') || []).join(', ') }}
                </span>
              </template>
            </bcs-table-column>
            <bcs-table-column
              label="Status"
              width="120"
              column-key="status"
              prop="status"
              :filters="podStatusFilters"
              filter-multiple
              show-overflow-tooltip>
              <template #default="{ row }">
                <StatusIcon :status="handleGetExtData(row.metadata.uid, 'status')"></StatusIcon>
              </template>
            </bcs-table-column>
            <bcs-table-column label="Ready" width="100">
              <template #default="{ row }">
                {{handleGetExtData(row.metadata.uid, 'readyCnt')}}/{{handleGetExtData(row.metadata.uid, 'totalCnt')}}
              </template>
            </bcs-table-column>
            <bcs-table-column label="Readiness Gates">
              <template #default="{ row }">
                <span
                  :class="{ 'bcs-border-tips inline-flex': getReadinessGates(row).content }"
                  v-bk-tooltips="getReadinessGates(row).content">
                  {{ getReadinessGates(row).rate }}
                </span>
              </template>
            </bcs-table-column>
            <bcs-table-column label="Restarts" width="100">
              <template #default="{ row }">{{handleGetExtData(row.metadata.uid, 'restartCnt')}}</template>
            </bcs-table-column>
            <bcs-table-column label="Host IP" min-width="140" show-overflow-tooltip>
              <template #default="{ row }">{{row.status.hostIP || '--'}}</template>
            </bcs-table-column>
            <bcs-table-column label="Pod IPv4" width="140" show-overflow-tooltip>
              <template #default="{ row }">{{handleGetExtData(row.metadata.uid, 'podIPv4') || '--'}}</template>
            </bcs-table-column>
            <bcs-table-column label="Pod IPv6" min-width="200" show-overflow-tooltip>
              <template #default="{ row }">{{handleGetExtData(row.metadata.uid, 'podIPv6') || '--'}}</template>
            </bcs-table-column>
            <bcs-table-column label="Node" show-overflow-tooltip>
              <template #default="{ row }">{{row.spec.nodeName || '--'}}</template>
            </bcs-table-column>
            <bcs-table-column label="Age" sortable prop="createTime">
              <template #default="{ row }">
                <span>{{handleGetExtData(row.metadata.uid, 'age')}}</span>
              </template>
            </bcs-table-column>
            <bcs-table-column :label="$t('generic.label.action')" width="200" fixed="right">
              <template #default="{ row }">
                <bk-button
                  text :disabled="handleGetExtData(row.metadata.uid, 'status') === 'Terminating'"
                  @click="handleShowLog(row)">{{ $t('generic.button.log1') }}</bk-button>
                <bk-button
                  class="ml10" :disabled="handleGetExtData(row.metadata.uid, 'status') === 'Terminating'"
                  text @click="handleReschedule(row)">{{ $t('dashboard.workload.pods.delete') }}</bk-button>
              </template>
            </bcs-table-column>
            <template #empty>
              <BcsEmptyTableStatus :type="searchValue ? 'search-empty' : 'empty'" @clear="searchValue = ''" />
            </template>
          </bcs-table>
        </bcs-tab-panel>
        <bcs-tab-panel name="event" :label="$t('generic.label.event')" render-directive="if">
          <EventTable
            :kinds="kind === 'Deployment' ? [kind,'ReplicaSet', 'Pod'] : [kind, 'Pod']"
            :cluster-id="clusterId"
            :namespace="namespace"
            :name="kindsNames"
            v-if="!loading && !clusterMap[clusterId]?.is_shared">
          </EventTable>
          <EventQueryTable
            class="min-h-[360px]"
            hide-cluster-and-namespace
            :kinds="kind === 'Deployment' ? [kind,'ReplicaSet', 'Pod'] : [kind, 'Pod']"
            :cluster-id="clusterId"
            :namespace="namespace"
            :name="kindsNames"
            :reset-page-when-name-change="false"
            v-else-if="!loading && !!clusterMap[clusterId]?.is_shared">
          </EventQueryTable>
        </bcs-tab-panel>
        <bcs-tab-panel name="label" :label="$t('k8s.label')" render-directive="if">
          <bk-table :data="labels">
            <bk-table-column label="Key" prop="key"></bk-table-column>
            <bk-table-column label="Value" prop="value"></bk-table-column>
          </bk-table>
        </bcs-tab-panel>
        <bcs-tab-panel name="annotations" :label="$t('k8s.annotation')" render-directive="if">
          <bk-table :data="annotations">
            <bk-table-column label="Key" prop="key"></bk-table-column>
            <bk-table-column label="Value" prop="value"></bk-table-column>
          </bk-table>
        </bcs-tab-panel>
        <bcs-tab-panel
          name="selector"
          :label="$t('k8s.selector')"
          render-directive="if"
          v-if="['deployments', 'statefulsets'].includes(category)">
          <bk-table :data="selectors">
            <bk-table-column label="Key" prop="key"></bk-table-column>
            <bk-table-column label="Value" prop="value"></bk-table-column>
          </bk-table>
        </bcs-tab-panel>
      </bcs-tab>
    </div>
    <bcs-sideslider quick-close :title="metadata.name" :is-show.sync="showYamlPanel" :width="800">
      <template #content>
        <CodeEditor
          v-full-screen="{ tools: ['fullscreen', 'copy'], content: yaml }"
          width="100%"
          height="100%"
          readonly
          :options="{
            roundedSelection: false,
            scrollBeyondLastLine: false,
            renderLineHighlight: 'none',
          }"
          :value="yaml">
        </CodeEditor>
      </template>
    </bcs-sideslider>
    <BcsLog
      v-model="showLog"
      :cluster-id="clusterId"
      :namespace="currentRow.metadata.namespace"
      :name="currentRow.metadata.name">
    </BcsLog>
    <bcs-dialog
      v-model="podRescheduleShow"
      :title="$t('dashboard.workload.pods.confirm')"
      @confirm="handleConfirmReschedule">
      <template v-if="isBatchReschedule">
        <div v-for="pod in selectPods" :key="pod['metadata']['uid']">
          {{ pod['metadata']['name'] }}
        </div>
      </template>
      <template v-else>
        <div v-for="pod in curPodRowData" :key="pod['metadata']['uid']">
          {{ pod['metadata']['name'] }}
        </div>
      </template>
    </bcs-dialog>
  </div>
</template>
<script lang="ts">
/* eslint-disable camelcase */
import { bkOverflowTips } from 'bk-magic-vue';
import { computed, defineComponent, onBeforeUnmount, onMounted, ref, toRefs, watch } from 'vue';

import EventTable from './bk-monitor-event.vue';
import detailBasicList from './detail-basic';
import useDetail from './use-detail';

import { userPermsByAction } from '@/api/modules/user-manager';
import $bkMessage from '@/common/bkmagic';
import { timeFormat } from '@/common/util';
import BcsLog from '@/components/bcs-log/log-dialog.vue';
import Metric from '@/components/metric.vue';
import CodeEditor from '@/components/monaco-editor/new-editor.vue';
import StatusIcon from '@/components/status-icon';
import { useCluster, useProject } from '@/composables/use-app';
import useInterval from '@/composables/use-interval';
import usePage from '@/composables/use-page';
import useSearch from '@/composables/use-search';
import useTableSort from '@/composables/use-table-sort';
import fullScreen from '@/directives/full-screen';
import $i18n from '@/i18n/i18n-setup';
import $store from '@/store';
import EventQueryTable from '@/views/project-manage/event-query/event-query-table.vue';

export interface IDetail {
  manifest: any;
  manifestExt: any;
}

export default defineComponent({
  name: 'WorkloadDetail',
  components: {
    StatusIcon,
    Metric,
    CodeEditor,
    BcsLog,
    EventTable,
    EventQueryTable,
  },
  directives: {
    bkOverflowTips,
    'full-screen': fullScreen,
  },
  props: {
    namespace: {
      type: String,
      default: '',
      required: true,
    },
    // workload类型
    category: {
      type: String,
      default: '',
      required: true,
    },
    // kind类型
    kind: {
      type: String,
      default: '',
      required: true,
    },
    // 名称
    name: {
      type: String,
      default: '',
      required: true,
    },
    crd: {
      type: String,
      default: '',
      required: true,
    },
    clusterId: {
      type: String,
      default: '',
      required: true,
    },
    // 是否隐藏 更新 和 删除操作（兼容集群管理应用详情）
    hiddenOperate: {
      type: Boolean,
      default: false,
    },
  },
  setup(props, ctx) {
    const { clusterId } = toRefs(props);
    const { clusterNameMap, clusterMap } = useCluster();
    const { curProject } = useProject();
    const updateStrategyMap = ref({
      RollingUpdate: $i18n.t('k8s.updateStrategy.rollingUpdate'),
      InplaceUpdate: $i18n.t('k8s.updateStrategy.inplaceUpdate'),
      OnDelete: $i18n.t('k8s.updateStrategy.onDelete'),
      Recreate: $i18n.t('k8s.updateStrategy.reCreate'),
    });
    const curType = props.category === 'custom_objects' ? 'crd' : 'workloads';
    const {
      isLoading,
      detail,
      activePanel,
      labels,
      annotations,
      selectors,
      updateStrategy,
      spec,
      metadata,
      manifestExt,
      pagePerms,
      additionalColumns,
      yaml,
      showYamlPanel,
      getJsonPathValue,
      handleGetDetail,
      handleGetCustomObjectDetail,
      handleShowYamlPanel,
      handleUpdateResource,
      handleDeleteResource,
    } = useDetail({
      ...props,
      defaultActivePanel: 'pod',
      type: curType,
    });
    const podLoading = ref(false);
    const workloadPods = ref<IDetail|null>(null);
    const basicInfoList = detailBasicList({
      category: props.category,
      detail,
    });
    const podTableRef = ref();
    const podRescheduleShow = ref(false);
    const isBatchReschedule = ref(false);
    const curPodRowData = ref<any>([]);
    // 表格选中的pods数据
    const selectPods = ref<any[]>([]);
    // pods数据
    const allPodsData = computed(() => (workloadPods.value?.manifest?.items || []).map(item => ({
      ...item,
      images: (handleGetExtData(item.metadata?.uid, 'images') || []).join(''),
      podIPv6: handleGetExtData(item.metadata?.uid, 'podIPv6'),
      podIPv4: handleGetExtData(item.metadata?.uid, 'podIPv4'),
    })));
    // 状态列表
    const podStatusFilters = computed(() => allPodsData.value.reduce((pre, item) => {
      const itemStatus = handleGetExtData(item.metadata.uid, 'status');
      const exist = pre.find(status => status === itemStatus);
      if (!exist) {
        pre.push(itemStatus);
      }
      return pre;
    }, []).map(status => ({
      text: status,
      value: status,
    })));
    const filters = ref<Record<string, string[]>>({});
    const handleFilterChange = (data) => {
      filters.value = data;
    };
    // 排序
    const { handleSortChange, sortTableData: pods } = useTableSort(allPodsData, item => ({
      createTime: handleGetExtData(item.metadata?.uid, 'createTime'),
    }));
    // 表头过滤
    const filterPodsByStatus = computed(() => pods.value.filter((item) => {
      const status = handleGetExtData(item.metadata.uid, 'status');
      return !filters.value?.status?.length || filters.value.status.includes(status);
    }));
    // pods过滤 'images', 'podIPv6', 'podIPv4' 这个三个参数在ext里面，在全量数据那里处理过
    const keys = ref(['metadata.name', 'status.hostIP', 'status.podIP', 'spec.nodeName', 'images', 'podIPv6', 'podIPv4']);
    const { searchValue, tableDataMatchSearch } = useSearch(filterPodsByStatus, keys);
    // pods分页
    const {
      pageChange: podTablePageChang,
      pageSizeChange: podTablePageSizeChange,
      curPageData: curPodTablePageData,
      pagination: podTablePagination,
    } = usePage(tableDataMatchSearch);
    watch(searchValue, () => {
      podTablePageChang(1);
    });
    // 当前行是否可以勾选
    const handlePodSelectable = row => handleGetExtData(row.metadata.uid, 'status') !== 'Terminating';
    // 是否展示升级策略
    const showUpdateStrategy = computed(() => ['deployments', 'statefulsets', 'custom_objects'].includes(props.category));
    // 是否展示批量调度功能
    const showBatchDispatch = computed(() => ['Deployment', 'StatefulSet', 'GameDeployment', 'GameStatefulSet'].includes(props.kind));
    // 获取pod manifestExt数据
    const handleGetExtData = (uid, prop) => workloadPods.value?.manifestExt?.[uid]?.[prop];
    // 指标参数
    const params = computed<Record<string, any>|null>(() => {
      const list = curPodTablePageData.value.map(item => item.metadata.name);
      return list.length
        ? { pod_name_list: list, $namespaceId: props.namespace, $clusterId: clusterId.value }
        : null;
    });
    const getReadinessGates = (row) => {
      const readinessGates = handleGetExtData(row.metadata.uid, 'readinessGates');
      const keys = Object.keys(readinessGates || {});
      // <none> 会被当作dom渲染不出来
      return {
        rate: `${keys.filter(key => readinessGates[key] === 'True').length}/${keys.length}`,
        content: keys.map(key => `${key}: ${readinessGates[key] === '<none>' ? 'none' : readinessGates[key]} `)
          .join('<br/>'),
      };
    };

    // 跳转pod详情
    const gotoPodDetail = (row) => {
      ctx.emit('pod-detail', row);
    };

    // 获取镜像tips
    const getImagesTips = (images) => {
      if (!images) {
        return {
          content: '',
        };
      }
      return {
        allowHTML: true,
        maxWidth: 480,
        content: images.join('<br />'),
      };
    };

    const handleSelectPod = (selection) => {
      selectPods.value = selection;
    };
    const handleSelectAllPod = (selection) => {
      selectPods.value = selection;
    };

    const handleGetPodsData = async () => {
      if (!clusterId.value) return;
      // 获取工作负载下对应的pod数据
      const matchLabels = detail.value?.manifest?.spec?.selector?.matchLabels || {};
      const labelSelector = Object.keys(matchLabels).reduce((pre, key, index) => {
        pre += `${index > 0 ? ',' : ''}${key}=${matchLabels[key]}`;
        return pre;
      }, '');

      const data = await $store.dispatch('dashboard/listWorkloadPods', {
        $namespaceId: props.namespace,
        $clusterId: clusterId.value,
        labelSelector,
        ownerKind: props.kind,
        ownerName: props.name,
        format: 'manifest',
      });

      return data;
    };
    // 获取工作负载下的pods数据
    const handleGetWorkloadPods = async () => {
      podLoading.value = true;
      workloadPods.value = await handleGetPodsData();
      podLoading.value = false;
    };

    // 显示日志
    const showLog = ref(false);
    const currentRow = ref<Record<string, any>>({ metadata: {} });
    const handleShowLog = (row) => {
      currentRow.value = row;
      showLog.value = true;
    };
    // 批量调度-打开弹框
    const handelShowRescheduleDialog = () => {
      isBatchReschedule.value = true;
      podRescheduleShow.value = true;
    };
    // 单个重新调度-打开弹框
    const handleReschedule = async (row) => {
      curPodRowData.value = [row];
      isBatchReschedule.value = false;
      podRescheduleShow.value = true;
    };

    // 确认调度
    const handleConfirmReschedule = async () => {
      podLoading.value = true;
      if (isBatchReschedule.value) {
        await handleBatchReschedulePod();
      } else if (['CronJob', 'Job'].includes(props.kind)) {
        await handleDeletePod();
      } else {
        await handleReschedulePod();
      }
      await handleGetWorkloadPods();
      podLoading.value = false;
    };

    // 删除Pod
    const handleDeletePod = async () => {
      const { name, namespace } = curPodRowData.value[0].metadata;
      const result = await $store.dispatch('dashboard/resourceDelete', {
        $namespaceId: namespace,
        $type: 'workloads',
        $category: 'pods',
        $clusterId: clusterId.value,
        $name: name,
      });
      result && $bkMessage({
        theme: 'success',
        message: $i18n.t('generic.msg.success.delete'),
      });
    };

    // 单个重新调度
    const handleReschedulePod = async () => {
      const { name } = curPodRowData.value[0].metadata;
      const result = await $store.dispatch('dashboard/reschedulePod', {
        $namespaceId: props.namespace,
        $podId: name,
        $clusterId: clusterId.value,
      });
      result && $bkMessage({
        theme: 'success',
        message: $i18n.t('generic.msg.success.cordon'),
      });
      selectPods.value = [];
      podTableRef.value?.clearSelection();
    };

    // 批量重新调度
    const batchBtnLoading = ref(false);
    const handleBatchReschedulePod = async () => {
      batchBtnLoading.value = true;
      const matchLabels = detail.value?.manifest?.spec?.selector?.matchLabels || {};
      const labelSelector = Object.keys(matchLabels).reduce((pre, key, index) => {
        pre += `${index > 0 ? ',' : ''}${key}=${matchLabels[key]}`;
        return pre;
      }, '');
      let result = false;
      if (['Deployment', 'Statefulset'].includes(props.kind)) {
        result = await $store.dispatch('dashboard/batchReschedulePod', {
          $namespace: props.namespace,
          $name: metadata.value.name,
          $category: props.category,
          $clusterId: clusterId.value,
          podNames: selectPods.value.map(pod => pod.metadata.name),
          labelSelector,
        });
      } else if (['GameDeployment', 'GameStatefulSet'].includes(props.kind)) {
        result = await $store.dispatch('dashboard/batchRescheduleCrdPod', {
          $crdName: props.crd,
          $cobjName: metadata.value.name,
          $clusterId: clusterId.value,
          podNames: selectPods.value.map(pod => pod.metadata.name),
          namespace: props.namespace,
          labelSelector,
        });
      }
      if (result) {
        $bkMessage({
          theme: 'success',
          message: $i18n.t('generic.msg.success.cordon'),
        });
        selectPods.value = [];
        podTableRef.value?.clearSelection();
      }
      batchBtnLoading.value = false;
    };

    // 事件列表
    const kindsNames = computed(() => [
      props.name,
      ...pods.value.map(item => item.metadata.name),
      ...rsNames.value,
    ]);

    // 刷新Pod状态
    const handleRefreshPodsStatus = async () => {
      workloadPods.value = await handleGetPodsData();
      // 获取详情
      if (props.category === 'custom_objects') {
        await handleGetCustomObjectDetail(false);
      } else {
        await handleGetDetail(false);
      }
    };
    const { start, stop } = useInterval(handleRefreshPodsStatus, 8000);

    // 获取RS资源
    const rsData = ref<any>({});
    const rsNames = computed(() => rsData.value?.manifest?.items?.map(item => item.metadata.name) || []);
    const handleGetRSData = async () => {
      if (props.kind !== 'Deployment') return;
      rsData.value = await $store.dispatch('dashboard/getReplicasets', {
        $namespaceId: props.namespace,
        $clusterId: clusterId.value,
        ownerName: props.name,
      });
    };

    // 删除按钮鉴权
    const hasPerms = ref(false);
    async function handelDelButtonPerms() {
      if (!clusterId.value || !props.namespace || !curProject.value?.project_id) return;
      const res = await userPermsByAction({
        $actionId: 'namespace_scoped_delete',
        perm_ctx: {
          resource_type: 'namespace',
          cluster_id: clusterId.value,
          name: props.namespace,
          project_id: curProject.value?.project_id,
        },
      }).catch(() => ({}));
      hasPerms.value = res.perms?.namespace_scoped_delete;
    };

    const loading = ref(true);
    onMounted(async () => {
      loading.value = true;
      // 详情接口前置
      if (props.category === 'custom_objects') {
        await handleGetCustomObjectDetail();
      } else {
        await handleGetDetail();
      }
      await Promise.all([
        handleGetWorkloadPods(),
        handleGetRSData(),
        handelDelButtonPerms(),
      ]);
      loading.value = false;
      // 开启轮询
      start();
    });
    onBeforeUnmount(() => {
      stop();
    });

    return {
      loading,
      batchBtnLoading,
      updateStrategyMap,
      isLoading,
      detail,
      updateStrategy,
      showUpdateStrategy,
      showBatchDispatch,
      spec,
      metadata,
      manifestExt,
      pagePerms,
      additionalColumns,
      basicInfoList,
      activePanel,
      params,
      podTableRef,
      selectPods,
      searchValue,
      labels,
      annotations,
      selectors,
      podLoading,
      yaml,
      showYamlPanel,
      kindsNames,
      timeFormat,
      handleShowYamlPanel,
      gotoPodDetail,
      handleGetExtData,
      getImagesTips,
      handleUpdateResource,
      handleDeleteResource,
      handleReschedule,
      getJsonPathValue,
      handleSelectPod,
      handleSelectAllPod,
      handleBatchReschedulePod,
      handleReschedulePod,
      handleConfirmReschedule,
      curPodTablePageData,
      podTablePagination,
      podTablePageChang,
      podTablePageSizeChange,
      podRescheduleShow,
      isBatchReschedule,
      curPodRowData,
      handelShowRescheduleDialog,
      handlePodSelectable,
      showLog,
      currentRow,
      handleShowLog,
      handleSortChange,
      getReadinessGates,
      clusterMap,
      clusterNameMap,
      podStatusFilters,
      handleFilterChange,
      curProject,
      hasPerms,
    };
  },
});
</script>
<style lang="postcss" scoped>
@import './workload-detail.css';
</style>
