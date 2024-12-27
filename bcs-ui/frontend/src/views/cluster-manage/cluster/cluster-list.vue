<template>
  <bk-table
    :data="sortClusterList"
    size="medium"
    :row-class-name="rowClassName"
    height="h-[calc(100%-48px)]"
    @row-click="handleRowClick">
    <bk-table-column :label="$t('cluster.labels.nameAndId')" :min-width="160" :show-overflow-tooltip="false">
      <template #default="{ row }">
        <bk-button
          :disabled="!supportDetailStatusList.includes(row.status)"
          text
          @click.stop="handleChangeActiveRow(row)">
          <div class="flex items-center">
            <span class="bcs-ellipsis flex-1" v-bk-overflow-tips>{{ row.name }}</span>
            <span
              :class="[
                'flex items-center justify-center px-[2px] rounded-sm',
                'text-[#fff] text-[12px] bg-[#2DCB56]'
              ]"
              style="transform: scale(0.75);"
              v-if="highlightClusterId === row.clusterID">
              new
            </span>
          </div>
        </bk-button>
        <div :class="supportDetailStatusList.includes(row.status) ? 'text-[#979BA5]' : 'text-[#dcdee5]'">
          {{ row.clusterID }}
        </div>
      </template>
    </bk-table-column>
    <bk-table-column
      :label="$t('cluster.labels.provider')"
      prop="provider"
      :filters="providerNameList"
      :filter-method="filterMethod">
      <template #default="{ row }">
        <div class="flex items-center">
          <svg class="size-[20px] mr-[10px] flex-shrink-0" v-if="providerNameMap[row.provider]">
            <use :xlink:href="providerNameMap[row.provider]?.className"></use>
          </svg>
          <span class="text-[12px] bcs-ellipsis">{{ providerNameMap[row.provider]?.label || '--' }}</span>
        </div>
      </template>
    </bk-table-column>
    <bk-table-column
      :label="$t('cluster.labels.status')"
      prop="status"
      :filters="statusList"
      :filter-method="filterMethod">
      <template #default="{ row }">
        <StatusIcon
          :status-color-map="statusColorMap"
          :status-text-map="statusTextMap"
          :status="row.status"
          :message="failedStatusList.includes(row.status) ? row.message : ''"
          :pending="loadingStatusList.includes(row.status)"
          @click.native.stop="
            failedStatusList.includes(row.status)
              ? handleShowClusterLog(row)
              : handleChangeActiveRow(row)" />
      </template>
    </bk-table-column>
    <bk-table-column
      :label="$t('cluster.labels.env')"
      prop="environment"
      :filters="clusterEnvFilters"
      :filter-method="filterMethod">
      <template #default="{ row }">
        {{ CLUSTER_ENV[row.environment] }}
      </template>
    </bk-table-column>
    <bk-table-column
      :label="$t('cluster.labels.clusterType')"
      prop="manageType"
      :filters="clusterTypeFilterList"
      :filter-method="filterMethod"
      :key="clusterTypeKey">
      <template #default="{ row }">
        {{ getClusterTypeName(row) }}
      </template>
    </bk-table-column>
    <bk-table-column :label="$t('cluster.labels.nodeCounts')">
      <template #default="{ row }">
        <template
          v-if="perms[row.clusterID] && perms[row.clusterID].cluster_manage && row.clusterType !== 'virtual'">
          <LoadingIcon
            v-if="clusterNodesMap[row.clusterID] === undefined">
            {{ $t('generic.status.loading') }}...
          </LoadingIcon>
          <div
            :class=" row.status === 'RUNNING' ? 'cursor-pointer' : 'cursor-not-allowed'"
            v-else
            @click.stop="handleGotoClusterNode(row)">
            <bk-button text :disabled="row.status !== 'RUNNING'">
              {{ clusterNodesMap[row.clusterID] }}
            </bk-button>
          </div>
        </template>
        <span v-else>--</span>
      </template>
    </bk-table-column>
    <bk-table-column :label="$t('cluster.labels.metric')" min-width="200">
      <template #default="{ row }">
        <div class="flex items-center" v-if="overview[row.clusterID]">
          <RingCell
            :percent="overview[row.clusterID]['cpu_usage'] && overview[row.clusterID]['cpu_usage']['percent']"
            fill-color="#3762B8"
            class="!mr-[10px] !ml-[-2px]" />
          <RingCell
            :percent="overview[row.clusterID]['memory_usage'] && overview[row.clusterID]['memory_usage']['percent']"
            fill-color="#61B2C2"
            class="!mr-[10px]" />
          <RingCell
            :percent="overview[row.clusterID]['disk_usage'] && overview[row.clusterID]['disk_usage']['percent']"
            fill-color="#B5E0AB"
            v-if="row.clusterType !== 'virtual'" />
        </div>
        <span v-else>--</span>
      </template>
    </bk-table-column>
    <!-- 220: 兼容英文状态时大小 -->
    <bk-table-column :label="$t('generic.label.action')" prop="action" width="220">
      <template #default="{ row }">
        <slot name="action" :row="row">
          <tableCellAction
            :row="row"
            :perms="perms"
            :cluster-extra-info="clusterExtraInfo"
            :cluster-nodes-map="clusterNodesMap"
            @show-cluster-log="handleShowClusterLog"
            @retry="handleRetry"
            @goto-cluster-detail="handleGotoClusterDetail"
            @delete-cluster="handleDeleteCluster"
            @goto-dashboard="handleGotoDashborad"
            @goto-cluster-node="handleGotoClusterNode"
            @goto-cluster-overview="handleGotoClusterOverview"
            @goto-cluster-ca="handleGotoClusterCA"
            @goto-token="handleGotoToken"
            @goto-web-console="handleGotoWebConsole" />
        </slot>
      </template>
    </bk-table-column>
    <template #empty>
      <BcsEmptyTableStatus :type="searchValue ? 'search-empty' : 'empty'" @clear="handleClearSearch" />
    </template>
  </bk-table>
</template>
<script lang="ts">
import { computed, defineComponent, PropType, toRefs } from 'vue';

import tableCellAction from './table-cell-action.vue';

import { CLUSTER_ENV } from '@/common/constant';
import LoadingIcon from '@/components/loading-icon.vue';
import StatusIcon from '@/components/status-icon';
import { ICluster, useProject } from '@/composables/use-app';
import $i18n from '@/i18n/i18n-setup';
import $router from '@/router';
import $store from '@/store';
import { getClusterTypeName } from '@/views/cluster-manage/cluster/use-cluster';
import RingCell from '@/views/cluster-manage/components/ring-cell.vue';
import useCloud from '@/views/cluster-manage/use-cloud';

export default defineComponent({
  name: 'ClusterList',
  components: { StatusIcon, RingCell, tableCellAction, LoadingIcon },
  props: {
    clusterList: {
      type: Array as PropType<ICluster[]>,
      default: () => [],
    },
    overview: {
      type: Object,
      default: () => ({}),
    },
    perms: {
      type: Object,
      default: () => ({}),
    },
    clusterExtraInfo: {
      type: Object,
      default: () => ({}),
    },
    searchValue: {
      type: String,
      default: '',
    },
    clusterNodesMap: {
      type: Object as PropType<Record<string, number>>,
      default: () => ({}),
    },
    activeClusterId: {
      type: String,
      default: '',
    },
    highlightClusterId: {
      type: String,
      default: '',
    },
    // 集群状态文案
    statusTextMap: {
      type: Object as PropType<Record<string, string>>,
      default: () => ({
        INITIALIZATION: $i18n.t('generic.status.initializing'),
        DELETING: $i18n.t('generic.status.deleting'),
        'CREATE-FAILURE': $i18n.t('generic.status.createFailed'),
        'DELETE-FAILURE': $i18n.t('generic.status.deleteFailed'),
        'IMPORT-FAILURE': $i18n.t('cluster.status.importFailed'),
        'CONNECT-FAILURE': $i18n.t('cluster.status.connectFailed'),
        RUNNING: $i18n.t('generic.status.ready'),
      }),
    },
    // 集群状态icon颜色
    statusColorMap: {
      type: Object as PropType<Record<string, string>>,
      default: () => ({
        'CREATE-FAILURE': 'red',
        'DELETE-FAILURE': 'red',
        'IMPORT-FAILURE': 'red',
        'CONNECT-FAILURE': 'red',
        RUNNING: 'green',
      }),
    },
    // 异常状态
    failedStatusList: {
      type: Array as PropType<string[]>,
      default: () => ['CREATE-FAILURE', 'DELETE-FAILURE', 'IMPORT-FAILURE'],
    },
    // 支持详情页展示的状态
    supportDetailStatusList: {
      type: Array as PropType<string[]>,
      default: () => ['CREATE-FAILURE', 'DELETE-FAILURE', 'CONNECT-FAILURE', 'RUNNING'],
    },
    // 加载中的状态
    loadingStatusList: {
      type: Array as PropType<string[]>,
      default: () => ['INITIALIZATION', 'DELETING'],
    },
  },
  setup(props, ctx) {
    const { curProject } = useProject();
    const {
      clusterList,
      highlightClusterId,
      statusTextMap,
      supportDetailStatusList } = toRefs(props);
    const sortClusterList = computed(() => clusterList.value.sort((cur) => {
      if (cur.clusterID === highlightClusterId.value) return -1;

      return 0;
    }));
    const statusList = computed(() => Object.keys(statusTextMap.value).map(status => ({
      text: statusTextMap.value[status],
      value: status,
    })));
    const { providerNameMap } = useCloud();
    const providerNameList = computed(() => Object.keys(providerNameMap).map(name => ({
      text: providerNameMap[name].label,
      value: name,
    })));
    const clusterEnvFilters = computed(() => clusterList.value
      .reduce<Array<{value: string, text: string}>>((pre, cluster) => {
      if (pre.find(item => item.value === cluster.environment)) return pre;
      pre.push({
        text: CLUSTER_ENV[cluster.environment],
        value: cluster.environment,
      });
      return pre;
    }, []));
    // 集群类型列表
    const hideSharedCluster = computed(() => $store.state.hideSharedCluster);
    const clusterTypeFilterList = computed(() => {
      const typeMap = clusterList.value.reduce<Record<string, string>>((pre, item) => {
        const tempPre = pre;
        if (item.is_shared && !hideSharedCluster.value) {
          tempPre.isShared = $i18n.t('bcs.cluster.share');
        } else if (item.clusterType === 'virtual') {
          tempPre.virtual = 'vCluster';
        } else if (item.clusterType === 'federation') {
          tempPre.federation = $i18n.t('bcs.cluster.federation');
        } else if (item.manageType === 'INDEPENDENT_CLUSTER') {
          tempPre.INDEPENDENT_CLUSTER = $i18n.t('bcs.cluster.selfDeployed');
        } else if (item.manageType === 'MANAGED_CLUSTER') {
          tempPre.MANAGED_CLUSTER = $i18n.t('bcs.cluster.managed');
        }
        return tempPre;
      }, {});

      return Object.keys(typeMap).map(key => ({
        value: key,
        text: typeMap[key],
      }));
    });
    const clusterTypeKey = computed(() => clusterTypeFilterList.value.map(item => item.value).join(','));
    const filterMethod = (value, row, column) => {
      const { property } = column;
      if (property === 'manageType') {
        // 筛选集群类型搜索逻辑
        if (value === 'isShared') {
          return row.is_shared;
        }
        if (value === 'virtual') {
          return row.clusterType === 'virtual';
        }
        if (value === 'federation') {
          return row.clusterType === 'federation';
        }
        return row[property] === value && !row.is_shared && !['virtual', 'federation'].includes(row.clusterType);
      }
      return row[property] === value;
    };
    // 跳转集群概览界面
    const handleGotoClusterOverview = (cluster) => {
      ctx.emit('overview', cluster);
    };
    // 跳转集群详情页
    const handleGotoClusterDetail = (cluster, active = 'info') => {
      ctx.emit('detail', { cluster, active });
    };
    // 跳转集群节点页
    const handleGotoClusterNode = (cluster) => {
      ctx.emit('node', cluster);
    };
    // 跳转集群自动扩缩容
    const handleGotoClusterCA = (cluster) => {
      ctx.emit('autoscaler', cluster);
    };
    // 集群删除
    const handleDeleteCluster = (cluster) => {
      ctx.emit('delete', cluster);
    };
    // 集群日志
    const handleShowClusterLog = (cluster) => {
      ctx.emit('log', cluster);
    };
    // 集群失败重试
    const handleRetry = (cluster) => {
      ctx.emit('retry', cluster);
    };
    // 集群kubeconfig
    const handleGotoToken = () => {
      ctx.emit('kubeconfig');
    };
    // 集群控制台
    const handleGotoWebConsole = (cluster) => {
      ctx.emit('webconsole', cluster);
    };
    // 清空搜索条件
    const handleClearSearch = () => {
      ctx.emit('clear');
    };
    // 资源视图
    const handleGotoDashborad = (row) => {
      // 跳到集群视图
      $router.push({
        name: 'dashboardWorkloadDeployments',
        params: {
          clusterId: row.clusterID,
        },
      });
    };

    // 行点击事件
    const handleRowClick = (row, e, col) => {
      if (col.property === 'action' || !supportDetailStatusList.value.includes(row.status)) return;
      handleChangeActiveRow(row);
    };

    // 当前active行
    const handleChangeActiveRow = (row) => {
      ctx.emit('active-row', row.clusterID);
    };
    const rowClassName = ({ row }) => {
      let classes = 'cluster-list-row';
      if (row.clusterID === props.activeClusterId) {
        classes += ' active-row';
      }
      if (supportDetailStatusList.value.includes(row.status)) {
        classes += ' cluster-row-clickable';
      }
      if (row.clusterID === props.highlightClusterId) {
        classes += ' highlight-active-row';
      }
      return classes;
    };

    return {
      clusterTypeKey,
      sortClusterList,
      CLUSTER_ENV,
      clusterEnvFilters,
      curProject,
      clusterTypeFilterList,
      statusList,
      providerNameList,
      providerNameMap,
      filterMethod,
      handleGotoClusterOverview,
      handleGotoClusterDetail,
      handleGotoClusterNode,
      handleGotoClusterCA,
      handleDeleteCluster,
      handleShowClusterLog,
      handleRetry,
      handleClearSearch,
      handleGotoToken,
      handleGotoWebConsole,
      handleGotoDashborad,
      handleChangeActiveRow,
      rowClassName,
      handleRowClick,
      getClusterTypeName,
    };
  },
});
</script>
<style lang="postcss" scoped>
>>> .active-row {
  background-color: #E1ECFF;
}
>>> .highlight-active-row {
  background-color: #F2FFF4;
}
>>> .cluster-row-clickable {
  cursor: pointer;
}
>>> .bk-table-body-wrapper {
  max-height: calc(100vh - 176px);
  overflow-y: auto;
}
</style>
