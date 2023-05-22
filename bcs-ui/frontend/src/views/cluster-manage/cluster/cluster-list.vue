<template>
  <bk-table :data="clusterList" size="medium">
    <bk-table-column :label="$t('集群名称/ID')" :min-width="160" :show-overflow-tooltip="false">
      <template #default="{ row }">
        <bk-button
          :disabled="row.status !== 'RUNNING'"
          text
          @click="handleGotoClusterDetail(row, 'overview')">
          <span class="bcs-ellipsis" v-bk-overflow-tips>{{ row.name }}</span>
        </bk-button>
        <div :class="row.status === 'RUNNING' ? 'text-[#979BA5]' : 'text-[#dcdee5]'">
          {{ row.clusterID }}
        </div>
      </template>
    </bk-table-column>
    <bk-table-column
      :label="$t('集群状态')"
      prop="status"
      :filters="[{
        text: $t('正常'),
        value: 'RUNNING'
      },{
        text: $t('初始化中'),
        value: 'INITIALIZATION'
      },{
        text: $t('删除中'),
        value: 'DELETING'
      },{
        text: $t('创建失败'),
        value: 'CREATE-FAILURE'
      },{
        text: $t('删除失败'),
        value: 'DELETE-FAILURE'
      },{
        text: $t('导入失败'),
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
            INITIALIZATION: $t('初始化中'),
            DELETING: $t('删除中'),
            'CREATE-FAILURE': $t('创建失败'),
            'DELETE-FAILURE': $t('删除失败'),
            'IMPORT-FAILURE': $t('导入失败'),
            RUNNING: $t('正常')
          }"
          :status="row.status"
          :pending="['INITIALIZATION', 'DELETING'].includes(row.status)" />
      </template>
    </bk-table-column>
    <bk-table-column
      :label="$t('集群环境')"
      prop="environment"
      :filters="clusterEnvFilters"
      :filter-method="filterMethod">
      <template #default="{ row }">
        {{ CLUSTER_ENV[row.environment] }}
      </template>
    </bk-table-column>
    <bk-table-column
      :label="$t('集群类型')"
      prop="manageType"
      :filters="[{
        text: $t('独立集群'),
        value: 'INDEPENDENT_CLUSTER'
      },{
        text: $t('托管集群'),
        value: 'MANAGED_CLUSTER'
      }]"
      :filter-method="filterMethod">
      <template #default="{ row }">
        {{ row.manageType === 'INDEPENDENT_CLUSTER' ? $t('独立集群') : $t('托管集群') }}
      </template>
    </bk-table-column>
    <bk-table-column :label="$t('集群节点数')">
      <template #default="{ row }">
        <template v-if="perms[row.clusterID] && perms[row.clusterID].cluster_manage">
          <LoadingIcon v-if="!clusterNodesMap[row.clusterID]">{{ $t('加载中') }}...</LoadingIcon>
          <div
            :class=" row.status === 'RUNNING' ? 'cursor-pointer' : 'cursor-not-allowed'"
            v-else
            @click="handleGotoClusterNode(row)">
            <bk-button text :disabled="row.status !== 'RUNNING'">
              {{ clusterNodesMap[row.clusterID].length }}
            </bk-button>
          </div>
        </template>
        <span v-else>--</span>
      </template>
    </bk-table-column>
    <bk-table-column :label="$t('集群资源(CPU/内存/磁盘)')" min-width="200">
      <template #default="{ row }">
        <template v-if="overview[row.clusterID]">
          <RingCell
            :percent="overview[row.clusterID]['cpu_usage'] && overview[row.clusterID]['cpu_usage']['percent']"
            fill-color="#3762B8"
            class="!mr-[10px]" />
          <RingCell
            :percent="overview[row.clusterID]['memory_usage'] && overview[row.clusterID]['memory_usage']['percent']"
            fill-color="#61B2C2"
            class="!mr-[10px]" />
          <RingCell
            :percent="overview[row.clusterID]['disk_usage'] && overview[row.clusterID]['disk_usage']['percent']"
            fill-color="#B5E0AB" />
        </template>
        <span v-else>--</span>
      </template>
    </bk-table-column>
    <bk-table-column :label="$t('操作')" width="180">
      <template #default="{ row }">
        <!-- 进行中 -->
        <template v-if="['INITIALIZATION', 'DELETING'].includes(row.status)">
          <bk-button
            text
            @click="handleShowClusterLog(row)">
            {{$t('查看日志')}}
          </bk-button>
        </template>
        <!-- 失败状态 -->
        <template v-else-if="['CREATE-FAILURE', 'DELETE-FAILURE'].includes(row.status)">
          <bk-button class="mr-[10px]" text @click="handleRetry(row)">{{ $t('重试') }}</bk-button>
          <bk-button class="mr-[10px]" text @click="handleShowClusterLog(row)">{{$t('查看日志')}}</bk-button>
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
                  @click="handleDeleteCluster(row)">
                  {{ $t('删除') }}
                </li>
              </ul>
            </template>
          </PopoverSelector>
        </template>
        <!-- 正常状态 -->
        <template v-else-if="row.status === 'RUNNING'">
          <bk-button text class="mr10" @click="handleGotoClusterOverview(row)">{{ $t('集群总览') }}</bk-button>
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
            @click="handleGotoClusterNode(row)">{{ $t('节点列表') }}</bk-button>
          <PopoverSelector offset="0, 10">
            <span class="bcs-icon-more-btn"><i class="bcs-icon bcs-icon-more"></i></span>
            <template #content>
              <ul class="bg-[#fff]">
                <li class="bcs-dropdown-item" @click="handleGotoClusterDetail(row, 'info')">{{ $t('基本信息') }}</li>
                <li class="bcs-dropdown-item" @click="handleGotoClusterDetail(row, 'network')">{{ $t('网络配置') }}</li>
                <li class="bcs-dropdown-item" @click="handleGotoClusterDetail(row, 'master')">{{ $t('Master配置') }}</li>
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
                  @click="handleGotoClusterCA(row)">
                  {{ $t('弹性扩缩容') }}
                </li>
                <li class="bcs-dropdown-item" @click="handleGotoToken">KubeConfig</li>
                <li class="bcs-dropdown-item" @click="handleGotoWebConsole(row)">WebConsole</li>
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
                    content: $t('集群下存在节点，无法删除'),
                    disabled: !clusterNodesMap[row.clusterID] || clusterNodesMap[row.clusterID].length === 0,
                    placement: 'right'
                  }"
                  @click="handleDeleteCluster(row)">
                  {{ $t('删除') }}
                </li>
              </ul>
            </template>
          </PopoverSelector>
        </template>
        <bk-button
          text
          v-else-if="deletable(row) && row.status !== 'DELETING'"
          @click="handleDeleteCluster(row)">
          {{ $t('删除') }}
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
import StatusIcon from '@/components/status-icon';
import RingCell from '@/views/cluster-manage/components/ring-cell.vue';
import PopoverSelector from '@/components/popover-selector.vue';
import LoadingIcon from '@/components/loading-icon.vue';
import { useProject } from '@/composables/use-app';
import { CLUSTER_ENV } from '@/common/constant';

export default defineComponent({
  name: 'ClusterList',
  components: { StatusIcon, RingCell, PopoverSelector, LoadingIcon },
  props: {
    clusterList: {
      type: Array as PropType<any[]>,
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
  },
  setup(props, ctx) {
    const { curProject } = useProject();
    const { clusterExtraInfo } = toRefs(props);
    const clusterEnvFilters = computed(() => props.clusterList.reduce((pre, cluster) => {
      if (pre.find(item => item.value === cluster.environment)) return pre;
      pre.push({
        text: CLUSTER_ENV[cluster.environment],
        value: cluster.environment,
      });
      return pre;
    }, []));
    const filterMethod = (value, row, column) => {
      const { property } = column;
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

    return {
      CLUSTER_ENV,
      clusterEnvFilters,
      curProject,
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
    };
  },
});
</script>
