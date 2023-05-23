<template>
  <div class="biz-content">
    <Header>
      {{$t('集群列表')}}
      <span class="ml-[5px] text-[12px] text-[#979ba5]">
        {{`( ${$t('业务')}: ${curProject.businessName} )`}}
      </span>
      <span class="bk-text-button bk-default f12 ml5" @click="handleShowProjectConf">
        <i class="bcs-icon bcs-icon-edit"></i>
      </span>
    </Header>
    <div class="p-[20px] bcs-content-wrapper" v-bkloading="{ isLoading, color: '#fafbfd' }">
      <template v-if="filterSharedClusterList.length">
        <Row class="mb-[16px]">
          <template #left>
            <span
              v-bk-tooltips="{
                disabled: isMaintainer,
                content: $t('您不是当前项目绑定业务的运维人员，请联系业务运维人员')
              }">
              <bk-button
                theme="primary"
                icon="plus"
                v-authority="{
                  actionId: 'cluster_create',
                  resourceName: curProject.project_name,
                  permCtx: {
                    resource_type: 'project',
                    project_id: curProject.project_id
                  }
                }"
                :disabled="!isMaintainer"
                @click="goCreateCluster">
                {{$t('添加集群')}}
              </bk-button>
            </span>
            <ApplyHost :title="$t('申请Master节点')" class="ml10" v-if="$INTERNAL" />
          </template>
          <template #right>
            <bk-input
              right-icon="bk-icon icon-search"
              class="w-[360px]"
              :placeholder="$t('集群名称/集群ID')"
              v-model="searchValue"
              clearable>
            </bk-input>
            <div class="flex ml-[8px]">
              <bcs-icon
                :class="['bcs-icon-btn bcs-icon bcs-icon-lie', { active: activeType === 'list' }]"
                type=""
                v-bk-tooltips="$t('列表视图')"
                @click="handleChangeType('list')" />
              <bcs-icon
                :class="['bcs-icon-btn bcs-icon bcs-icon-kuai ml-[-1px]', { active: activeType === 'card' }]"
                type=""
                v-bk-tooltips="$t('卡片视图')"
                @click="handleChangeType('card')" />
            </div>
          </template>
        </Row>
        <CardMode
          :cluster-list="clusterList"
          :overview="clusterOverviewMap"
          :perms="webAnnotations.perms"
          :search-value="searchValue"
          :cluster-extra-info="clusterExtraInfo"
          :status-text-map="statusTextMap"
          :cluster-nodes-map="clusterNodesMap"
          @overview="goOverview"
          @detail="goClusterDetail"
          @node="goNodeInfo"
          @autoscaler="goClusterAutoScaler"
          @delete="handleDeleteCluster"
          @log="handleShowLog"
          @retry="handleRetry"
          @create="goCreateCluster"
          @kubeconfig="goClusterToken"
          @webconsole="handleGotoConsole"
          v-if="activeType === 'card'" />
        <ListMode
          :cluster-list="clusterList"
          :overview="clusterOverviewMap"
          :perms="webAnnotations.perms"
          :search-value="searchValue"
          :cluster-extra-info="clusterExtraInfo"
          :cluster-nodes-map="clusterNodesMap"
          @overview="goOverview"
          @detail="goClusterDetail"
          @node="goNodeInfo"
          @autoscaler="goClusterAutoScaler"
          @delete="handleDeleteCluster"
          @log="handleShowLog"
          @retry="handleRetry"
          @create="goCreateCluster"
          @clear="searchValue = ''"
          @kubeconfig="goClusterToken"
          @webconsole="handleGotoConsole"
          v-else />
      </template>
      <ClusterGuide v-else-if="!isLoading" />
    </div>
    <!-- 集群日志 -->
    <bcs-sideslider
      :is-show.sync="showLogDialog"
      :title="curOperateCluster && curOperateCluster.cluster_id"
      :width="860"
      quick-close
      @hidden="handleCloseLog">
      <template #content>
        <TaskList v-bkloading="{ isLoading: logLoading }" class="px-[30px] py-[20px]" :data="taskData"></TaskList>
      </template>
    </bcs-sideslider>
    <!-- 集群删除确认弹窗 -->
    <ConfirmDialog
      v-model="showConfirmDialog"
      :title="$t('确定删除集群？')"
      :sub-title="$t('此操作无法撤回，请确认：')"
      :tips="deleteClusterTips"
      :ok-text="$t('删除')"
      :cancel-text="$t('关闭')"
      :confirm="confirmDeleteCluster" />
    <!-- 编辑项目集群信息 -->
    <ProjectConfig v-model="isProjectConfDialogShow" />
  </div>
</template>

<script lang="ts">
/* eslint-disable camelcase */
import { defineComponent, ref, computed, onMounted, set } from 'vue';
import { useClusterList, useClusterOverview, useClusterOperate, useTask } from './use-cluster';
import { useCluster, useProject } from '@/composables/use-app';
import useSearch from '@/composables/use-search';
import ApplyHost from '../components/apply-host.vue';
import ProjectConfig from '@/views/project-manage/project/project-config.vue';
import ConfirmDialog from '@/components/comfirm-dialog.vue';
import TaskList from '@/views/cluster-manage/components/task-list.vue';
import Header from '@/components/layout/Header.vue';
import ClusterGuide from './cluster-guide.vue';
import Row from '@/components/layout/Row.vue';
import ListMode from './cluster-list.vue';
import CardMode from './cluster-card.vue';
import useNode from '@/views/cluster-manage/node-list/use-node';
import $store from '@/store';
import $router from '@/router';
import $i18n from '@/i18n/i18n-setup';
import $bkMessage from '@/common/bkmagic';
import $bkInfo from '@/components/bk-magic-2.0/bk-info';

export default defineComponent({
  name: 'ClusterOverview',
  components: {
    ApplyHost,
    ProjectConfig,
    ConfirmDialog,
    TaskList,
    Header,
    ClusterGuide,
    Row,
    ListMode,
    CardMode,
  },
  setup() {
    const { curProject, isMaintainer } = useProject();

    // 集群状态
    const statusTextMap = {
      INITIALIZATION: $i18n.t('初始化中'),
      DELETING: $i18n.t('删除中'),
      'CREATE-FAILURE': $i18n.t('创建失败'),
      'DELETE-FAILURE': $i18n.t('删除失败'),
      'IMPORT-FAILURE': $i18n.t('导入失败'),
    };
    // 切换展示模式
    const activeType = computed<'card'|'list'>(() => $store.state.clusterViewType as any);
    const handleChangeType = (type) => {
      $store.commit('updateClusterViewType', type);
    };

    // 集群列表
    const { clusterList: clusterData, getClusterList, clusterExtraInfo, webAnnotations } = useClusterList();
    const filterSharedClusterList = computed(() => clusterData.value.filter(item => !item.is_shared));
    const keys = ref(['name', 'clusterID']);
    const { searchValue, tableDataMatchSearch: clusterList } = useSearch(filterSharedClusterList, keys);
    const isLoading = ref(false);
    const handleGetClusterList = async () => {
      isLoading.value = true;
      await getClusterList();
      isLoading.value = false;
    };
    // 集群指标
    const { getClusterOverview, clusterOverviewMap } = useClusterOverview(clusterList);

    // 集群信息编辑
    const isProjectConfDialogShow = ref(false);
    const handleShowProjectConf = () => {
      isProjectConfDialogShow.value = true;
    };
    // 跳转创建集群界面
    const goCreateCluster = async () => {
      $router.push({ name: 'clusterCreate' });
    };
    // 跳转集群预览界面
    const goOverview = async (cluster) => {
      if (cluster.status !== 'RUNNING') return;
      $router.push({
        name: 'clusterOverview',
        params: {
          clusterId: cluster.cluster_id,
        },
      });
    };
    // 跳转集群信息界面
    const goClusterDetail = async ({ cluster, active }) => {
      $router.push({
        name: 'clusterDetail',
        params: {
          clusterId: cluster.cluster_id,
        },
        query: {
          active,
        },
      });
    };
    // 跳转添加节点界面
    const goNodeInfo = async (cluster) => {
      $router.push({
        name: 'clusterNode',
        params: {
          clusterId: cluster.cluster_id,
        },
      });
    };
    // 跳转扩缩容界面
    const goClusterAutoScaler = (cluster) => {
      $router.push({
        name: 'clusterDetail',
        params: {
          clusterId: cluster.cluster_id,
        },
        query: {
          active: 'autoscaler',
        },
      });
    };
    // kubeconfig
    const goClusterToken = () => {
      $router.push({ name: 'token' });
    };
    // webconsole
    const { handleGotoConsole } = useCluster();
    const { deleteCluster, retryClusterTask } = useClusterOperate();
    const curOperateCluster = ref<any>(null);
    // 集群删除
    const showConfirmDialog = ref(false);
    const deleteClusterTips = computed(() => (curOperateCluster.value?.clusterCategory === 'importer'
      ? [
        $i18n.t('您正在尝试删除集群 {clusterName}，此操作不可逆，请谨慎操作', { clusterName: curOperateCluster.value?.clusterID }),
        $i18n.t('删除导入集群不会对集群有任何操作'),
        $i18n.t('删除导入集群后蓝鲸容器管理平台会丧失对该集群的管理能力'),
        $i18n.t('删除导入集群不会清理集群上的任何资源，删除集群后请视情况手工清理'),
      ]
      : [
        $i18n.t('您正在尝试删除集群 {clusterName}，此操作不可逆，请谨慎操作', { clusterName: curOperateCluster.value?.clusterID }),
        $i18n.t('请确认已清理该集群下的所有应用与节点'),
        $i18n.t('集群删除时会清理集群上的工作负载、服务、路由等集群上的所有资源'),
        $i18n.t('集群删除后服务器如不再使用请尽快回收，避免产生不必要的成本'),
      ]));
    const confirmDeleteCluster = async () => {
      const result = await deleteCluster(curOperateCluster.value);
      if (result) {
        await handleGetClusterList();
        $bkMessage({
          theme: 'success',
          message: $i18n.t('任务下发成功'),
        });
      }
    };
    const handleDeleteCluster = (cluster) => {
      if (clusterNodesMap.value[cluster.clusterID]?.length > 0 && cluster.status === 'RUNNING') return;

      curOperateCluster.value = cluster;
      setTimeout(() => {
        showConfirmDialog.value = true;
      }, 0);
    };
    // 集群日志
    const logLoading = ref(false);
    const { taskList } = useTask();
    const showLogDialog = ref(false);
    const latestTask = ref<any>(null);
    const taskTimer = ref<any>(null);
    const statusColorMap = ref({
      initialzing: 'blue',
      running: 'blue',
      success: 'green',
      failure: 'red',
      timeout: 'red',
      notstarted: 'blue',
    });
    const taskStatusTextMap = ref({
      initialzing: $i18n.t('初始化中'),
      running: $i18n.t('运行中'),
      success: $i18n.t('成功'),
      failure: $i18n.t('失败'),
      timeout: $i18n.t('超时'),
      notstarted: $i18n.t('未执行'),
    });
    const taskData = computed(() => {
      const steps = latestTask.value?.stepSequence || [];
      return steps.map(step => latestTask.value?.steps[step]);
    });
    const fetchLogData = async (cluster) => {
      const res = await taskList(cluster);
      latestTask.value = res.latestTask;
      if (['RUNNING', 'INITIALZING'].includes(latestTask.value?.status)) {
        taskTimer.value = setTimeout(() => {
          fetchLogData(cluster);
        }, 5000);
      } else {
        clearTimeout(taskTimer.value);
        taskTimer.value = null;
      }
    };
    const handleShowLog = async (cluster) => {
      logLoading.value = true;
      showLogDialog.value = true;
      curOperateCluster.value = cluster;
      await fetchLogData(cluster);
      logLoading.value = false;
    };
    const handleCloseLog = () => {
      curOperateCluster.value = null;
      clearTimeout(taskTimer.value);
    };
    // 失败重试
    const handleRetry = async (cluster) => {
      isLoading.value = true;
      if (['CREATE-FAILURE', 'DELETE-FAILURE'].includes(cluster.status)) {
        // 创建重试
        $bkInfo({
          type: 'warning',
          title: cluster.status === 'CREATE-FAILURE' ? $i18n.t('确认重新创建集群') :  $i18n.t('确认删除集群'),
          clsName: 'custom-info-confirm default-info',
          subTitle: cluster.clusterName,
          confirmFn: async () => {
            isLoading.value = true;
            const result = await retryClusterTask(cluster);
            if (result) {
              await handleGetClusterList();
              $bkMessage({
                theme: 'success',
                message: $i18n.t('任务下发成功'),
              });
            }
            isLoading.value = false;
          },
        });
      } else {
        $bkMessage({
          theme: 'error',
          message: $i18n.t('未知状态'),
        });
      }
      isLoading.value = false;
    };

    // 集群节点数
    const { getNodeList } = useNode();
    const clusterNodesMap = ref({});
    const handleGetClusterNodes = async () => {
      clusterList.value
        .filter(cluster => webAnnotations.value.perms[cluster.clusterID].cluster_manage)
        .forEach((item) => {
          getNodeList(item.clusterID).then((data) => {
            set(clusterNodesMap.value, item.clusterID, data);
          });
        });
    };

    onMounted(async () => {
      await handleGetClusterList();
      await handleGetClusterNodes();
    });

    return {
      isMaintainer,
      clusterNodesMap,
      searchValue,
      activeType,
      isLoading,
      filterSharedClusterList,
      clusterList,
      curProject,
      clusterOverviewMap,
      isProjectConfDialogShow,
      curOperateCluster,
      handleShowProjectConf,
      getClusterOverview,
      goCreateCluster,
      goOverview,
      goClusterDetail,
      goClusterAutoScaler,
      showConfirmDialog,
      deleteClusterTips,
      confirmDeleteCluster,
      handleDeleteCluster,
      showLogDialog,
      latestTask,
      taskData,
      statusColorMap,
      taskStatusTextMap,
      logLoading,
      handleShowLog,
      handleCloseLog,
      handleRetry,
      goNodeInfo,
      goClusterToken,
      clusterExtraInfo,
      webAnnotations,
      handleChangeType,
      statusTextMap,
      handleGotoConsole,
    };
  },
});
</script>
<style lang="postcss" scoped>
.bcs-icon-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  border: 1px solid #C4C6CC;
  background-color: #fff;
  border-radius: 0 2px 2px 0;
  &.active {
    border-color: #3A84FF;
    background-color: #E1ECFF;
    z-index: 2;
  }
}
</style>
