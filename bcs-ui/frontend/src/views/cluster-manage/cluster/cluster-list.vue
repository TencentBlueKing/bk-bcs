<template>
  <bk-table :data="clusterList" size="medium" :row-class-name="rowClassName" @row-click="handleRowClick">
    <bk-table-column :label="$t('cluster.labels.nameAndId')" :min-width="160" :show-overflow-tooltip="false">
      <template #default="{ row }">
        <bk-button
          :disabled="row.status !== 'RUNNING'"
          text
          @click.stop="handleChangeActiveRow(row)">
          <span class="bcs-ellipsis" v-bk-overflow-tips>{{ row.name }}</span>
        </bk-button>
        <div :class="row.status === 'RUNNING' ? 'text-[#979BA5]' : 'text-[#dcdee5]'">
          {{ row.clusterID }}
        </div>
      </template>
    </bk-table-column>
    <bk-table-column
      :label="$t('cluster.labels.status')"
      prop="status"
      :filters="[{
        text: $t('generic.status.ready'),
        value: 'RUNNING'
      },{
        text: $t('generic.status.initializing'),
        value: 'INITIALIZATION'
      },{
        text: $t('generic.status.deleting'),
        value: 'DELETING'
      },{
        text: $t('generic.status.createFailed'),
        value: 'CREATE-FAILURE'
      },{
        text: $t('generic.status.deleteFailed'),
        value: 'DELETE-FAILURE'
      },{
        text: $t('cluster.status.importFailed'),
        value: 'IMPORT-FAILURE'
      }]"
      :filter-method="filterMethod">
      <template #default="{ row }">
        <StatusIcon
          :status-color-map="{
            'CREATE-FAILURE': 'red',
            'DELETE-FAILURE': 'red',
            'IMPORT-FAILURE': 'red',
            RUNNING: 'green'
          }"
          :status-text-map="{
            INITIALIZATION: $t('generic.status.initializing'),
            DELETING: $t('generic.status.deleting'),
            'CREATE-FAILURE': $t('generic.status.createFailed'),
            'DELETE-FAILURE': $t('generic.status.deleteFailed'),
            'IMPORT-FAILURE': $t('cluster.status.importFailed'),
            RUNNING: $t('generic.status.ready')
          }"
          :status="row.status"
          :pending="['INITIALIZATION', 'DELETING'].includes(row.status)" />
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
        <template v-else-if="row.status === 'RUNNING'">
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
                    @click.stop="handleGotoClusterDetail(row, 'network')">{{ $t('cluster.detail.title.network') }}</li>
                  <li
                    class="bcs-dropdown-item"
                    @click.stop="handleGotoClusterDetail(row, 'master')">{{ $t('cluster.detail.title.master') }}</li>
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
import { computed, defineComponent, PropType, toRefs } from 'vue';

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
  },
  setup(props, ctx) {
    const { curProject } = useProject();
    const { clusterExtraInfo, clusterList } = toRefs(props);
    const clusterEnvFilters = computed(() => props.clusterList
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
      if (cluster.status !== 'RUNNING') return;
      ctx.emit('overview', cluster);
    };
    // 跳转集群详情页
    const handleGotoClusterDetail = (cluster, active = 'info') => {
      ctx.emit('detail', { cluster, active });
    };
    // 跳转集群节点页
    const handleGotoClusterNode = (cluster) => {
      if (cluster.status !== 'RUNNING') return;
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
      if (col.property === 'action' || row.status !== 'RUNNING') return;
      handleChangeActiveRow(row);
    };

    // 当前active行
    const handleChangeActiveRow = (row) => {
      ctx.emit('active-row', row.clusterID);
    };
    const rowClassName = ({ row }) => (row.clusterID === props.activeClusterId ? 'cluster-list-row active-row' : 'cluster-list-row');

    return {
      CLUSTER_ENV,
      clusterEnvFilters,
      curProject,
      clusterTypeFilterList,
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
>>> .cluster-list-row {
  cursor: pointer;
}
</style>
