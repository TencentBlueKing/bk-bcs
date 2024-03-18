<template>
  <bk-table
    :data="sortClusterList"
    size="medium"
    :row-class-name="rowClassName"
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
          :pending="['INITIALIZATION', 'DELETING'].includes(row.status)"
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
      :filter-method="filterMethod">
      <template #default="{ row }">
        <template v-if="row.clusterType === 'virtual'">vCluster</template>
        <template v-else>
          {{ row.manageType === 'INDEPENDENT_CLUSTER' ? $t('bcs.cluster.selfDeployed') : $t('bcs.cluster.managed') }}
        </template>
      </template>
    </bk-table-column>
    <bk-table-column :label="$t('cluster.labels.nodeCounts')">
      <template #default="{ row }">
        <template v-if="perms[row.clusterID] && perms[row.clusterID].cluster_manage && row.clusterType !== 'virtual'">
          <LoadingIcon v-if="!clusterNodesMap[row.clusterID]">{{ $t('generic.status.loading') }}...</LoadingIcon>
          <div
            :class=" row.status === 'RUNNING' ? 'cursor-pointer' : 'cursor-not-allowed'"
            v-else
            @click.stop="handleGotoClusterNode(row)">
            <bk-button text :disabled="row.status !== 'RUNNING'">
              {{ clusterNodesMap[row.clusterID].length }}
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
    <bk-table-column :label="$t('generic.label.action')" prop="action" width="200">
      <template #default="{ row }">
        <!-- 进行中 -->
        <template v-if="['INITIALIZATION', 'DELETING'].includes(row.status)">
          <bk-button
            text
            @click.stop="handleShowClusterLog(row)">
            {{$t('generic.button.log')}}
          </bk-button>
        </template>
        <!-- 失败状态 -->
        <template v-else-if="['CREATE-FAILURE', 'DELETE-FAILURE'].includes(row.status)">
          <bk-button
            class="mr-[10px]"
            text
            @click.stop="handleRetry(row)">{{ $t('cluster.ca.nodePool.records.action.retry') }}</bk-button>
          <bk-button
            class="mr-[10px]"
            text
            @click.stop="handleShowClusterLog(row)">{{$t('generic.button.log')}}</bk-button>
          <PopoverSelector offset="0, 10">
            <span class="bcs-icon-more-btn"><i class="bcs-icon bcs-icon-more"></i></span>
            <template #content>
              <ul class="bg-[#fff]">
                <li
                  class="bcs-dropdown-item"
                  @click.stop="handleGotoClusterDetail(row, 'info')">{{ $t('generic.title.basicInfo1') }}</li>
                <li
                  class="bcs-dropdown-item"
                  v-if="!(row.clusterCategory === 'importer' && row.importCategory === 'kubeConfig')"
                  @click.stop="handleGotoClusterDetail(row, 'network')">
                  {{ $t('cluster.detail.title.network') }}
                </li>
                <li
                  class="bcs-dropdown-item"
                  v-if="!(row.clusterCategory === 'importer' && row.importCategory === 'kubeConfig')"
                  @click.stop="handleGotoClusterDetail(row, 'master')">
                  {{ $t('cluster.detail.title.master') }}
                </li>
                <li
                  class="bcs-dropdown-item"
                  v-authority="{
                    clickable: perms[row.clusterID]
                      && perms[row.clusterID].cluster_delete,
                    actionId: 'cluster_delete',
                    resourceName: row.clusterName,
                    disablePerms: true,
                    permCtx: {
                      project_id: curProject.projectID,
                      cluster_id: row.clusterID
                    }
                  }"
                  key="deleteCluster"
                  @click.stop="handleDeleteCluster(row)">
                  {{ $t('generic.button.delete') }}
                </li>
              </ul>
            </template>
          </PopoverSelector>
        </template>
        <!-- 正常状态 -->
        <template v-else-if="normalStatusList.includes(row.status)">
          <bk-button text class="mr10" @click.stop="handleGotoClusterOverview(row)">
            {{ $t('cluster.button.overview') }}
          </bk-button>
          <bk-button
            text
            class="mr10"
            v-if="row.clusterType === 'virtual'"
            @click.stop="handleGotoDashborad(row)">{{ $t('cluster.button.dashboard') }}</bk-button>
          <bk-button
            text
            class="mr10"
            v-authority="{
              clickable: perms[row.clusterID]
                && perms[row.clusterID].cluster_manage,
              actionId: 'cluster_manage',
              resourceName: row.clusterName,
              disablePerms: true,
              permCtx: {
                project_id: curProject.projectID,
                cluster_id: row.clusterID
              }
            }"
            key="nodeList"
            v-if="row.clusterType !== 'virtual'"
            @click.stop="handleGotoClusterNode(row)">
            {{ $t('cluster.detail.title.nodeList') }}
          </bk-button>
          <PopoverSelector offset="0, 10">
            <span class="bcs-icon-more-btn"><i class="bcs-icon bcs-icon-more"></i></span>
            <template #content>
              <ul class="bg-[#fff]">
                <li
                  class="bcs-dropdown-item"
                  @click.stop="handleGotoClusterDetail(row, 'info')">{{ $t('generic.title.basicInfo1') }}</li>
                <li
                  class="bcs-dropdown-item"
                  v-if="row.clusterType === 'virtual'"
                  @click.stop="handleGotoClusterDetail(row, 'quota')">
                  {{ $t('cluster.detail.title.quota') }}
                </li>
                <template v-else>
                  <li
                    class="bcs-dropdown-item"
                    v-if="!(row.clusterCategory === 'importer' && row.importCategory === 'kubeConfig')"
                    @click.stop="handleGotoClusterDetail(row, 'network')">
                    {{ $t('cluster.detail.title.network') }}
                  </li>
                  <li
                    class="bcs-dropdown-item"
                    v-if="!(row.clusterCategory === 'importer' && row.importCategory === 'kubeConfig')"
                    @click.stop="handleGotoClusterDetail(row, 'master')">
                    {{ $t('cluster.detail.title.master') }}
                  </li>
                  <li
                    class="bcs-dropdown-item"
                    v-if="clusterExtraInfo
                      && clusterExtraInfo[row.clusterID]
                      && clusterExtraInfo[row.clusterID].autoScale"
                    v-authority="{
                      clickable: perms[row.clusterID]
                        && perms[row.clusterID].cluster_manage,
                      actionId: 'cluster_manage',
                      resourceName: row.clusterName,
                      disablePerms: true,
                      permCtx: {
                        project_id: curProject.projectID,
                        cluster_id: row.clusterID
                      }
                    }"
                    key="ca"
                    @click.stop="handleGotoClusterCA(row)">
                    {{ $t('cluster.detail.title.autoScaler') }}
                  </li>
                </template>
                <li class="bcs-dropdown-item" @click.stop="handleGotoToken">KubeConfig</li>
                <li class="bcs-dropdown-item" @click.stop="handleGotoWebConsole(row)">WebConsole</li>
                <li
                  class="bcs-dropdown-item"
                  v-authority="{
                    clickable: perms[row.clusterID]
                      && perms[row.clusterID].cluster_delete,
                    actionId: 'cluster_delete',
                    resourceName: row.clusterName,
                    disablePerms: true,
                    permCtx: {
                      project_id: curProject.projectID,
                      cluster_id: row.clusterID
                    }
                  }"
                  key="deletevCluster"
                  v-if="row.clusterType === 'virtual' || row.clusterCategory === 'importer'"
                  @click.stop="handleDeleteCluster(row)">
                  {{ $t('generic.button.delete') }}
                </li>
                <li
                  :class="[
                    'bcs-dropdown-item',
                    { disabled: clusterNodesMap[row.clusterID] && clusterNodesMap[row.clusterID].length > 0 }
                  ]"
                  v-authority="{
                    clickable: perms[row.clusterID]
                      && perms[row.clusterID].cluster_delete,
                    actionId: 'cluster_delete',
                    resourceName: row.clusterName,
                    disablePerms: true,
                    permCtx: {
                      project_id: curProject.projectID,
                      cluster_id: row.clusterID
                    }
                  }"
                  key="deleteCluster"
                  v-bk-tooltips="{
                    content: $t('cluster.validate.exitNodes'),
                    disabled: !clusterNodesMap[row.clusterID] || clusterNodesMap[row.clusterID].length === 0,
                    placement: 'right'
                  }"
                  v-else
                  @click.stop="handleDeleteCluster(row)">
                  {{ $t('generic.button.delete') }}
                </li>
              </ul>
            </template>
          </PopoverSelector>
        </template>
        <bk-button
          text
          v-else-if="deletable(row) && row.status !== 'DELETING'"
          @click.stop="handleDeleteCluster(row)">
          {{ $t('generic.button.delete') }}
        </bk-button>
      </template>
    </bk-table-column>
    <template #empty>
      <BcsEmptyTableStatus :type="searchValue ? 'search-empty' : 'empty'" @clear="handleClearSearch" />
    </template>
  </bk-table>
</template>
<script lang="ts">
import { computed, defineComponent, PropType, ref, toRefs } from 'vue';

import { CLUSTER_ENV } from '@/common/constant';
import LoadingIcon from '@/components/loading-icon.vue';
import PopoverSelector from '@/components/popover-selector.vue';
import StatusIcon from '@/components/status-icon';
import { ICluster, useProject } from '@/composables/use-app';
import $i18n from '@/i18n/i18n-setup';
import $router from '@/router';
import RingCell from '@/views/cluster-manage/components/ring-cell.vue';

export default defineComponent({
  name: 'ClusterList',
  components: { StatusIcon, RingCell, PopoverSelector, LoadingIcon },
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
      type: Object,
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
  },
  setup(props, ctx) {
    const { curProject } = useProject();
    const { clusterExtraInfo, clusterList, highlightClusterId } = toRefs(props);
    const sortClusterList = computed(() => clusterList.value.sort((cur) => {
      if (cur.clusterID === highlightClusterId.value) return -1;

      return 0;
    }));
    const statusTextMap = ref({
      INITIALIZATION: $i18n.t('generic.status.initializing'),
      DELETING: $i18n.t('generic.status.deleting'),
      'CREATE-FAILURE': $i18n.t('generic.status.createFailed'),
      'DELETE-FAILURE': $i18n.t('generic.status.deleteFailed'),
      'IMPORT-FAILURE': $i18n.t('cluster.status.importFailed'),
      'CONNECT-FAILURE': $i18n.t('cluster.status.connectFailed'),
      RUNNING: $i18n.t('generic.status.ready'),
    });
    const statusList = computed(() => Object.keys(statusTextMap.value).map(status => ({
      text: statusTextMap.value[status],
      value: status,
    })));
    const statusColorMap = ref({
      'CREATE-FAILURE': 'red',
      'DELETE-FAILURE': 'red',
      'IMPORT-FAILURE': 'red',
      'CONNECT-FAILURE': 'red',
      RUNNING: 'green',
    });
    // 异常状态
    const failedStatusList = ['CREATE-FAILURE', 'DELETE-FAILURE', 'IMPORT-FAILURE'];
    // 支持详情页展示的状态
    const supportDetailStatusList = ['CREATE-FAILURE', 'DELETE-FAILURE', 'CONNECT-FAILURE', 'RUNNING'];
    // 正常状态
    const normalStatusList = ['CONNECT-FAILURE', 'RUNNING'];
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
    const clusterTypeFilterList = computed(() => {
      const typeMap =  clusterList.value.reduce<Record<string, string>>((pre, item) => {
        if (item.clusterType === 'virtual') {
          pre.virtual = 'vCluster';
        } else if (item.manageType === 'INDEPENDENT_CLUSTER') {
          pre.INDEPENDENT_CLUSTER = $i18n.t('bcs.cluster.selfDeployed');
        } else if (item.manageType === 'MANAGED_CLUSTER') {
          pre.MANAGED_CLUSTER = $i18n.t('bcs.cluster.managed');
        }
        return pre;
      }, {});

      return Object.keys(typeMap).map(key => ({
        value: key,
        text: typeMap[key],
      }));
    });
    const filterMethod = (value, row, column) => {
      const { property } = column;
      // 出来集群类型搜索逻辑
      if (property === 'manageType') {
        return value === 'virtual' ? row.clusterType === 'virtual' : row[property] === value;
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
    const deletable = cluster => !!clusterExtraInfo.value[cluster.clusterID]?.canDeleted;
    // 清空搜索条件
    const handleClearSearch = () => {
      ctx.emit('clear');
    };
    // 资源视图
    const handleGotoDashborad = (row) => {
      $router.push({
        name: 'dashboardHome',
        params: {
          clusterId: row.clusterID,
        },
      });
    };

    // 行点击事件
    const handleRowClick = (row, e, col) => {
      if (col.property === 'action' || !supportDetailStatusList.includes(row.status)) return;
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
      if (supportDetailStatusList.includes(row.status)) {
        classes += ' cluster-row-clickable';
      }
      if (row.clusterID === props.highlightClusterId) {
        classes += ' highlight-active-row';
      }
      return classes;
    };

    return {
      normalStatusList,
      failedStatusList,
      sortClusterList,
      supportDetailStatusList,
      CLUSTER_ENV,
      clusterEnvFilters,
      curProject,
      clusterTypeFilterList,
      statusTextMap,
      statusColorMap,
      statusList,
      filterMethod,
      handleGotoClusterOverview,
      handleGotoClusterDetail,
      handleGotoClusterNode,
      handleGotoClusterCA,
      handleDeleteCluster,
      handleShowClusterLog,
      handleRetry,
      deletable,
      handleClearSearch,
      handleGotoToken,
      handleGotoWebConsole,
      handleGotoDashborad,
      handleChangeActiveRow,
      rowClassName,
      handleRowClick,
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
