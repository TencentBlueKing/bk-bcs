<template>
  <!-- 集群面板 -->
  <div class="flex flex-wrap">
    <bk-exception type="search-empty" v-if="searchValue && !clusterList.length"></bk-exception>
    <template v-else>
      <div
        class="cluster-card"
        v-for="cluster in clusterList"
        :key="cluster.clusterID">
        <!-- todo指标过载 -->
        <!-- <div v-bk-tooltips="{ content: $t('集群指标过载') }">!</div> -->
        <Row class="h-[75px] header-border-bottom">
          <template #left>
            <!-- 集群信息 -->
            <div class="ml-[24px]">
              <div
                :class="[
                  'bcs-ellipsis-word text-[16px] text-[#333C48]',
                  { 'hover:text-[#3a84ff] hover:cursor-pointer': cluster.status === 'RUNNING' }
                ]"
                v-bk-overflow-tips
                @click="handleGotoClusterDetail(cluster, 'overview')">
                {{ cluster.name || '--' }}
              </div>
              <div class="flex items-center">
                <span v-bk-overflow-tips class="bcs-ellipsis text-[12px] text-[#C3CDD7]">
                  {{cluster.clusterID || '--'}}
                </span>
                <bcs-tag theme="warning" v-if="['stag', 'debug'].includes(cluster.environment)">
                  {{cluster.environment === 'stag' ? 'UAT' : $t('测试')}}
                </bcs-tag>
                <bcs-tag theme="info" v-else-if="cluster.environment === 'prod'">
                  {{$t('正式')}}
                </bcs-tag>
              </div>
            </div>
          </template>
          <template #right>
            <!-- 集群操作菜单(正常状态) -->
            <PopoverSelector
              class="mr-[12px]"
              offset="0, 10"
              trigger="click"
              v-if="cluster.status === 'RUNNING'">
              <span class="bcs-icon-more-btn">
                <i class="text-[24px] bcs-icon bcs-icon-more"></i>
              </span>
              <ul slot="content" class="bg-[#fff]">
                <li class="bcs-dropdown-item" @click="handleGotoClusterOverview(cluster)">{{$t('集群总览')}}</li>
                <li class="bcs-dropdown-item" @click="handleGotoClusterDetail(cluster, 'info')">{{$t('基本信息')}}</li>
                <li class="bcs-dropdown-item" @click="handleGotoClusterDetail(cluster, 'network')">{{$t('网络配置')}}</li>
                <li
                  class="bcs-dropdown-item"
                  @click="handleGotoClusterDetail(cluster, 'master')">
                  {{$t('Master配置')}}
                </li>
                <li
                  class="bcs-dropdown-item"
                  v-authority="{
                    clickable: perms[cluster.clusterID]
                      && perms[cluster.clusterID].cluster_manage,
                    actionId: 'cluster_manage',
                    resourceName: cluster.clusterName,
                    disablePerms: true,
                    permCtx: {
                      project_id: curProject.projectID,
                      cluster_id: cluster.clusterID
                    }
                  }"
                  key="nodeList"
                  :disabled="cluster.clusterCategory === 'importer'"
                  @click="handleGotoClusterNode(cluster)">
                  {{$t('节点列表')}}
                </li>
                <li
                  class="bcs-dropdown-item"
                  key="ca"
                  v-authority="{
                    clickable: perms[cluster.clusterID]
                      && perms[cluster.clusterID].cluster_manage,
                    actionId: 'cluster_manage',
                    resourceName: cluster.clusterName,
                    disablePerms: true,
                    permCtx: {
                      project_id: curProject.projectID,
                      cluster_id: cluster.clusterID
                    }
                  }"
                  @click="handleGotoClusterCA(cluster)">
                  {{$t('弹性扩缩容')}}
                </li>
                <li class="bcs-dropdown-item" @click="handleGotoToken">KubeConfig</li>
                <li class="bcs-dropdown-item" @click="handleGotoWebConsole(cluster)">WebConsole</li>
                <li
                  :class="[
                    'bcs-dropdown-item',
                    { disabled: !clusterNodesMap[cluster.clusterID] || clusterNodesMap[cluster.clusterID].length > 0 }
                  ]"
                  key="deleteCluster"
                  v-authority="{
                    clickable: perms[cluster.clusterID]
                      && perms[cluster.clusterID].cluster_delete,
                    actionId: 'cluster_delete',
                    resourceName: cluster.clusterName,
                    disablePerms: true,
                    permCtx: {
                      project_id: curProject.projectID,
                      cluster_id: cluster.clusterID
                    }
                  }"
                  v-bk-tooltips="{
                    content: $t('集群下存在节点，无法删除'),
                    disabled: !clusterNodesMap[cluster.clusterID] || clusterNodesMap[cluster.clusterID].length === 0,
                    placement: 'right'
                  }"
                  @click="handleDeleteCluster(cluster)">
                  {{$t('删除')}}
                </li>
              </ul>
            </PopoverSelector>
            <PopoverSelector
              class="mr-[12px]"
              offset="0, 10"
              trigger="click"
              v-else-if="deletable(cluster) && cluster.status !== 'DELETING'">
              <span class="bcs-icon-more-btn">
                <i class="text-[24px] bcs-icon bcs-icon-more"></i>
              </span>
              <ul slot="content" class="bg-[#fff]">
                <li class="bcs-dropdown-item" @click="handleDeleteCluster(cluster)">
                  {{$t('删除')}}
                </li>
              </ul>
            </PopoverSelector>
          </template>
        </Row>
        <!-- 集群状态 -->
        <div class="flex flex-col items-center justify-center flex-1 text-[12px] p-[24px]">
          <!-- 进行中（移除中、初始化中） -->
          <div
            class="flex flex-col items-center"
            v-if="['INITIALIZATION', 'DELETING'].includes(cluster.status)">
            <LoadingCell class="bk-spin-loading-large" />
            <p class="mt-[6px]">{{ statusTextMap[cluster.status] }}</p>
            <bk-button
              text
              class="text-[12px] mt-[4px]"
              @click="handleShowClusterLog(cluster)">
              {{$t('查看日志')}}
            </bk-button>
          </div>
          <!-- 失败 -->
          <div
            class="flex flex-col items-center"
            v-else-if="['CREATE-FAILURE', 'DELETE-FAILURE'].includes(cluster.status)">
            <img class="w-[220px] h-[100px]" :src="maintain" />
            <p class="mt-[6px]">{{ statusTextMap[cluster.status] }}</p>
            <div class="mt-[4px]">
              <bk-button class="text-[12px] mr-[5px]" text @click="handleRetry(cluster)">{{ $t('重试') }}</bk-button>|
              <bk-button class="text-[12px]" text @click="handleShowClusterLog(cluster)">{{$t('查看日志')}}</bk-button>
            </div>
          </div>
          <!-- 导入失败 -->
          <div
            class="flex flex-col items-center"
            v-else-if="cluster.status === 'IMPORT-FAILURE'">
            <img class="w-[220px] h-[100px]" :src="maintain" />
            <p class="mt-[6px]">{{ statusTextMap[cluster.status] }}</p>
            <bk-button
              text
              class="text-[12px] mt-[4px]">
            <!-- 导入失败目前只能删了, 重新导入 -->
            </bk-button>
          </div>
          <!-- 正常 -->
          <template v-else-if="cluster.status === 'RUNNING'">
            <!-- 指标信息 -->
            <div
              v-for="item, index in clusterMetricList"
              :key="item.metric"
              :class="['h-[28px] w-full', { mb20: index < (clusterMetricList.length - 1) }]">
              <div class="flex place-content-between">
                <span>{{item.title}}</span>
                <span class="text-[#979BA5]">
                  {{
                    overview[cluster.clusterID] && overview[cluster.clusterID][item.metric]
                      ? overview[cluster.clusterID][item.metric].percent || 0
                      : 0
                  }}%
                </span>
              </div>
              <bcs-progress
                :class="['mt-[4px]', { loading: !overview[cluster.clusterID] }]"
                :color="item.color"
                :percent="
                  overview[cluster.clusterID] && overview[cluster.clusterID][item.metric]
                    ? (overview[cluster.clusterID][item.metric].percent || 0) / 100
                    : 0
                "
                :show-text="false">
              </bcs-progress>
            </div>
          </template>
          <!-- 未知状态 -->
          <div class="flex flex-col items-center" v-else>
            <img class="w-[220px] h-[100px]" :src="maintain" />
            <p class="mt-[6px]">{{ statusTextMap[cluster.status] || $t('未知集群状态') }}</p>
            <bk-button
              text
              class="text-[12px] mt-[4px]">
            <!-- 未知集群状态 -->
            </bk-button>
          </div>
        </div>
      </div>
    </template>
  </div>
</template>
<script lang="ts">
import { useProject } from '@/composables/use-app';
import { defineComponent, PropType, toRefs } from 'vue';
import Row from '@/components/layout/Row.vue';
import LoadingCell from '@/views/cluster-manage/components/loading-cell.vue';
import $i18n from '@/i18n/i18n-setup';
import PopoverSelector from '@/components/popover-selector.vue';
import maintain from '@/images/500.svg';

export default defineComponent({
  name: 'ClusterCard',
  components: { Row, LoadingCell, PopoverSelector },
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
    statusTextMap: {
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
    const { clusterExtraInfo } = toRefs(props);
    const { curProject, isMaintainer } = useProject();
    // 指标信息配置
    const clusterMetricList = [
      {
        metric: 'cpu_usage',
        title: $i18n.t('CPU使用率'),
        color: '#9dcaff',
      },
      {
        metric: 'memory_usage',
        title: $i18n.t('内存使用率'),
        color: '#97ebbb',
      },
      {
        metric: 'disk_usage',
        title: $i18n.t('磁盘使用率'),
        color: '#ffd97f',
      },
    ];

    // 跳转集群概览界面
    const handleGotoClusterOverview = (cluster) => {
      if (cluster.status !== 'RUNNING') return;
      ctx.emit('overview', cluster);
    };
    // 跳转集群详情页
    const handleGotoClusterDetail = (cluster, active = 'info') => {
      if (cluster.status !== 'RUNNING') return;
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
    // 创建集群
    const handleCreateCluster = () => {
      if (!isMaintainer.value) return;
      ctx.emit('create');
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

    return {
      isMaintainer,
      maintain,
      curProject,
      clusterMetricList,
      handleGotoClusterOverview,
      handleGotoClusterDetail,
      handleGotoClusterNode,
      handleGotoClusterCA,
      handleDeleteCluster,
      handleShowClusterLog,
      handleRetry,
      handleCreateCluster,
      deletable,
      handleGotoToken,
      handleGotoWebConsole,
    };
  },
});
</script>
<style lang="postcss" scoped>
.cluster-card {
  display: flex;
  flex-direction: column;
  width: 260px;
  height: 260px;
  margin-right: 16px;
  margin-bottom: 16px;
  background-color: #fff;
  border: 1px solid #DDE4EB;
  border-radius: 2px;
  .header-border-bottom {
    border-bottom: 1px solid #DDE4EB;
  }
  &.create-cluster:hover:not(.disabled) {
    border-color: #3a84ff;
    cursor: pointer;
    i, span {
      color: #3a84ff;
    }
  }
  &.disabled {
    cursor: not-allowed;
    color: #c4c6cc;
    border-color: #dcdee5;
  }
}
>>> .bk-progress {
  .progress-bar.bk-progress-normal {
    height: 6px;
  }
  &.loading .progress-bar.bk-progress-normal {
    background-color: #fafbfd;
    background-image: repeating-linear-gradient(90deg, #fafbfd, #fafbfd 20px, #f0f1f5 20px, #f0f1f5 60px);
    background-size: 250px 100%;
    animation: status-progress-loading 5s linear infinite;
  }
}

@keyframes status-progress-loading {
  0% {
      background-position: 0% 0%;
  }
  100% {
      background-position: 250px 0%;
  }
}
</style>
