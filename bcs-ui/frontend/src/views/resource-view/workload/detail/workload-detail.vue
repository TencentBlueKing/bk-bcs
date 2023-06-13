<template>
  <div class="workload-detail bcs-content-wrapper" v-bkloading="{ isLoading }">
    <div class="workload-detail-info">
      <div class="workload-info-basic">
        <div class="basic-left">
          <span class="name mr20">{{ metadata.name }}</span>
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
          <bk-button theme="primary" @click="handleShowYamlPanel">{{ $t('查看YAML配置') }}</bk-button>
          <template v-if="!hiddenOperate">
            <bk-button
              theme="primary"
              @click="handleUpdateResource">{{$t('更新')}}</bk-button>
            <bk-button
              theme="danger"
              v-authority="{
                clickable: webAnnotations.perms
                  && webAnnotations.perms.page.deleteBtn ? webAnnotations.perms.page.deleteBtn.clickable : true,
                content: webAnnotations.perms
                  && webAnnotations.perms.page.deleteBtn ? webAnnotations.perms.page.deleteBtn.tip : '',
                disablePerms: true
              }"
              @click="handleDeleteResource">{{$t('删除')}}</bk-button>
          </template>
        </div>
      </div>
      <div class="workload-main-info">
        <div class="info-item">
          <span class="label">{{ $t('命名空间') }}</span>
          <span class="value">{{ metadata.namespace }}</span>
        </div>
        <div class="info-item">
          <span class="label">{{ $t('镜像') }}</span>
          <span class="value" v-bk-overflow-tips="getImagesTips(manifestExt.images)">
            {{ manifestExt.images && manifestExt.images.join(', ') }}</span>
        </div>
        <div class="info-item">
          <span class="label">UID</span>
          <span class="value">{{ metadata.uid }}</span>
        </div>
        <div class="info-item">
          <span class="label">{{ $t('创建时间') }}</span>
          <span class="value">{{ timeZoneTransForm(manifestExt.createTime) }}</span>
        </div>
        <div class="info-item">
          <span class="label">{{ $t('存在时间') }}</span>
          <span class="value">{{ manifestExt.age }}</span>
        </div>
        <div class="info-item" v-if="showUpdateStrategy">
          <span class="label">{{ $t('升级策略') }}</span>
          <span class="value">{{ updateStrategyMap[updateStrategy.type] || $t('滚动升级') }}</span>
        </div>
      </div>
      <div class="workload-main-info" v-if="category === 'deployments'">
        <div class="info-item">
          <span class="label">{{ $t('最大调度Pod数量') }}</span>
          <span class="value" v-if="$chainable(spec, 'strategy.rollingUpdate.maxSurge')">
            {{ String(spec.strategy.rollingUpdate.maxSurge).split('%')[0] }}%</span>
          <span class="value" v-else>--</span>
        </div>
        <div class="info-item">
          <span class="label">{{ $t('最大不可用数量') }}</span>
          <span class="value" v-if="$chainable(spec, 'strategy.rollingUpdate.maxUnavailable')">
            {{ String(spec.strategy.rollingUpdate.maxUnavailable).split('%')[0] }}%</span>
          <span class="value" v-else>--</span>
        </div>
        <div class="info-item">
          <span class="label">{{ $t('最小就绪时间') }}</span>
          <span class="value" v-if="Number.isInteger(spec.minReadySeconds)">{{ spec.minReadySeconds }}s</span>
          <span class="value" v-else>--</span>
        </div>
        <div class="info-item">
          <span class="label">{{ $t('进程截止时间') }}</span>
          <span class="value" v-if="Number.isInteger(spec.progressDeadlineSeconds)">
            {{ spec.progressDeadlineSeconds }}s</span>
          <span class="value" v-else>--</span>
        </div>
      </div>
      <div class="workload-main-info" v-if="category === 'statefulsets'">
        <div class="info-item">
          <span class="label">{{ $t('Pod管理策略') }}</span>
          <span class="value">{{ spec.podManagementPolicy || '--' }}</span>
        </div>
        <div class="info-item">
          <span class="label">{{ $t('分区滚动更新') }}</span>
          <span class="value" v-if="Number.isInteger($chainable(spec, 'updateStrategy.rollingUpdate.partition'))">
            {{ spec.updateStrategy.rollingUpdate.partition }}s</span>
          <span class="value" v-else>--</span>
        </div>
      </div>
    </div>
    <div class="workload-detail-body">
      <div class="workload-metric" v-bkloading="{ isLoading: podLoading }">
        <Metric :title="$t('CPU使用率')" metric="cpu_usage" :params="params" category="pods" colors="#30d878"></Metric>
        <Metric
          :title="$t('内存使用量')"
          metric="memory_used"
          :params="params"
          unit="byte"
          category="pods"
          colors="#3a84ff"
          :desc="$t('container_memory_working_set_bytes，limit限制时oom判断依据')">
        </Metric>
        <Metric
          :title="$t('网络')"
          :metric="['network_receive', 'network_transmit']"
          :params="params"
          category="pods"
          unit="byte"
          :colors="['#853cff', '#30d878']"
          :suffix="[$t('入流量'), $t('出流量')]">
        </Metric>
      </div>
      <bcs-tab class="workload-tab" :active.sync="activePanel" type="card" :label-height="42">
        <bcs-tab-panel name="pod" label="Pods" v-bkloading="{ isLoading: podLoading }">
          <div class="pod-info-header">
            <bk-button
              v-if="showBatchDispatch"
              :loading="batchBtnLoading"
              :disabled="!selectPods.length"
              @click="handelShowRescheduleDialog">
              {{ $t('批量重新调度') }}
            </bk-button>
            <!-- 占位 -->
            <div v-else></div>
            <bk-input
              v-model="searchValue"
              :placeholder="$t('请入名称、镜像、IP、Node搜索')"
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
          >
            <bcs-table-column
              v-if="showBatchDispatch"
              type="selection"
              width="60"
              reserve-selection
              :selectable="handlePodSelectable">
            </bcs-table-column>
            <bcs-table-column
              :label="$t('名称')"
              min-width="130" prop="metadata.name" sortable show-overflow-tooltip fixed="left">
              <template #default="{ row }">
                <bk-button
                  class="bcs-button-ellipsis"
                  text @click="gotoPodDetail(row)">{{ row.metadata.name }}</bk-button>
              </template>
            </bcs-table-column>
            <bcs-table-column :label="$t('镜像')" min-width="200" :show-overflow-tooltip="false">
              <template #default="{ row }">
                <span v-bk-tooltips.top="(handleGetExtData(row.metadata.uid, 'images') || []).join('<br />')">
                  {{ (handleGetExtData(row.metadata.uid, 'images') || []).join(', ') }}
                </span>
              </template>
            </bcs-table-column>
            <bcs-table-column label="Status" width="120" show-overflow-tooltip>
              <template #default="{ row }">
                <StatusIcon :status="handleGetExtData(row.metadata.uid, 'status')"></StatusIcon>
              </template>
            </bcs-table-column>
            <bcs-table-column label="Ready" width="100">
              <template #default="{ row }">
                {{handleGetExtData(row.metadata.uid, 'readyCnt')}}/{{handleGetExtData(row.metadata.uid, 'totalCnt')}}
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
            <bcs-table-column :label="$t('操作')" width="140" fixed="right">
              <template #default="{ row }">
                <bk-button
                  text :disabled="handleGetExtData(row.metadata.uid, 'status') === 'Terminating'"
                  @click="handleShowLog(row)">{{ $t('日志') }}</bk-button>
                <bk-button
                  class="ml10" :disabled="handleGetExtData(row.metadata.uid, 'status') === 'Terminating'"
                  text @click="handleReschedule(row)">{{ $t('重新调度') }}</bk-button>
              </template>
            </bcs-table-column>
            <template #empty>
              <BcsEmptyTableStatus :type="searchValue ? 'search-empty' : 'empty'" @clear="searchValue = ''" />
            </template>
          </bcs-table>
        </bcs-tab-panel>
        <bcs-tab-panel name="event" :label="$t('事件')">
          <EventQueryTableVue
            class="min-h-[360px]"
            hide-cluster-and-namespace
            :kinds="kind === 'Deployment' ? [kind,'ReplicaSet', 'Pod'] : [kind, 'Pod']"
            :cluster-id="clusterId"
            :namespace="namespace"
            :name="kindsNames"
            v-if="!loading">
          </EventQueryTableVue>
        </bcs-tab-panel>
        <bcs-tab-panel name="label" :label="$t('标签')">
          <bk-table :data="labels">
            <bk-table-column label="Key" prop="key"></bk-table-column>
            <bk-table-column label="Value" prop="value"></bk-table-column>
          </bk-table>
        </bcs-tab-panel>
        <bcs-tab-panel name="annotations" :label="$t('注解')">
          <bk-table :data="annotations">
            <bk-table-column label="Key" prop="key"></bk-table-column>
            <bk-table-column label="Value" prop="value"></bk-table-column>
          </bk-table>
        </bcs-tab-panel>
        <bcs-tab-panel name="selector" :label="$t('选择器')" v-if="['deployments', 'statefulsets'].includes(category)">
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
            renderLineHighlight: false,
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
    <bcs-dialog v-model="podRescheduleShow" :title="$t(`确认重新调度以下Pod`)" @confirm="handleConfirmReschedule">
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
import { defineComponent, computed, ref, onMounted, onBeforeUnmount, watch, toRefs } from 'vue';
import { bkOverflowTips } from 'bk-magic-vue';
import StatusIcon from '@/components/status-icon';
import Metric from '@/components/metric.vue';
import useDetail from './use-detail';
import detailBasicList from './detail-basic';
import CodeEditor from '@/components/monaco-editor/new-editor.vue';
import fullScreen from '@/directives/full-screen';
import useInterval from '@/composables/use-interval';
import BcsLog from '@/components/bcs-log/log-dialog.vue';
import usePage from '@/composables/use-page';
import { timeZoneTransForm } from '@/common/util';
import useSearch from '@/composables/use-search';
import EventQueryTableVue from '@/views/project-manage/event-query/event-query-table.vue';
import useTableSort from '@/composables/use-table-sort';
import $store from '@/store';
import $bkMessage from '@/common/bkmagic';
import $i18n from '@/i18n/i18n-setup';

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
    EventQueryTableVue,
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
    const updateStrategyMap = ref({
      RollingUpdate: $i18n.t('滚动升级'),
      InplaceUpdate: $i18n.t('原地升级'),
      OnDelete: $i18n.t('手动删除'),
      Recreate: $i18n.t('重新创建'),
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
      webAnnotations,
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
    })));
    // 排序
    const { handleSortChange, sortTableData: pods } = useTableSort(allPodsData, item => ({
      createTime: handleGetExtData(item.metadata?.uid, 'createTime'),
    }));
    // pods过滤
    const keys = ref(['metadata.name', 'images', 'podIPv6', 'status.hostIP', 'status.podIP', 'spec.nodeName']);
    const { searchValue, tableDataMatchSearch } = useSearch(pods, keys);
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
      } else {
        await handleReschedulePod();
      }
      await handleGetWorkloadPods();
      podLoading.value = false;
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
        message: $i18n.t('调度成功'),
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
          message: $i18n.t('调度成功'),
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
      webAnnotations,
      additionalColumns,
      basicInfoList,
      activePanel,
      params,
      pods,
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
      timeZoneTransForm,
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
    };
  },
});
</script>
<style lang="postcss" scoped>
@import './workload-detail.css';
</style>
