<template>
  <div class="biz-content">
    <div
      class="px-[24px] py-[16px] h-full overflow-x-hidden"
      v-bkloading="{ isLoading, color: '#fafbfd' }"
      ref="contentRef">
      <template v-if="filterSharedClusterList.length">
        <div class="flex">
          <div class="flex items-center place-content-between mb-[16px] flex-1">
            <div class="flex items-center">
              <span
                v-bk-tooltips="{
                  disabled: isMaintainer,
                  content: $t('bcs.msg.notDevOps')
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
                  {{$t('cluster.button.addCluster')}}
                </bk-button>
              </span>
              <ApplyHost :title="$t('cluster.button.applyMaster')" class="ml10" v-if="$INTERNAL" />
              <bk-button
                class="ml-[10px]"
                v-if="flagsMap['NODETEMPLATE']"
                @click="goNodeTemplate">{{ $t('nav.nodeTemplate') }}</bk-button>
            </div>
            <bk-input
              right-icon="bk-icon icon-search"
              class="flex-1 ml-[10px] max-w-[360px]"
              :placeholder="$t('cluster.placeholder.searchCluster')"
              v-model.trim="searchValue"
              clearable>
            </bk-input>
          </div>
          <!-- flex左右布局空div -->
          <div :style="{ width: activeClusterID ? detailWidth : 0 }"></div>
        </div>
        <ListMode
          :cluster-list="clusterList"
          :overview="clusterOverviewMap"
          :perms="webAnnotations.perms"
          :search-value="searchValue"
          :cluster-extra-info="clusterExtraInfo"
          :cluster-nodes-map="clusterNodesMap"
          :active-cluster-id="activeClusterID"
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
          @active-row="handleChangeDetail" />
      </template>
      <ClusterGuide v-else-if="!isLoading" />
      <ClusterDetail
        :max-width="detailPanelMaxWidth"
        :key="activeClusterID"
        :cluster-id="activeClusterID"
        :active="activeTabName"
        v-if="activeClusterID"
        ref="clusterDetailRef"
        @width-change="handleDetailWidthChange" />
    </div>
    <!-- 集群日志 -->
    <bcs-sideslider
      :is-show.sync="showLogDialog"
      :title="curOperateCluster && curOperateCluster.cluster_id"
      :width="960"
      quick-close
      @hidden="handleCloseLog">
      <template #content>
        <TaskList v-bkloading="{ isLoading: logLoading }" class="px-[24px] py-[20px]" :data="taskData"></TaskList>
        <div class="bg-[#FAFBFD] h-[48px] flex items-center px-[24px] log-footer-border-top">
          <bcs-button
            class="w-[88px]"
            theme="primary"
            v-if="['CREATE-FAILURE', 'DELETE-FAILURE'].includes(curOperateCluster.status)"
            @click="handleRetry(curOperateCluster)">
            {{ $t('cluster.ca.nodePool.records.action.retry') }}
          </bcs-button>
          <bcs-button class="w-[88px]" @click="showLogDialog = false">{{ $t('generic.button.cancel') }}</bcs-button>
        </div>
      </template>
    </bcs-sideslider>
    <!-- 集群删除确认弹窗 -->
    <ConfirmDialog
      v-model="showConfirmDialog"
      :title="$t('cluster.button.delete.title')"
      :sub-title="$t('generic.subTitle.deleteConfirm')"
      :tips="deleteClusterTips"
      :ok-text="$t('generic.button.delete')"
      :cancel-text="$t('generic.button.close')"
      :confirm="confirmDeleteCluster" />
    <!-- 编辑项目集群信息 -->
    <ProjectConfig v-model="isProjectConfDialogShow" />
    <bcs-dialog
      v-model="showConnectCluster"
      :show-footer="false"
      width="588">
      <ConnectCluster
        :cluster="curRow"
        @confirm="handleRetryTask"
        @cancel="showConnectCluster = false" />
    </bcs-dialog>
  </div>
</template>

<script lang="ts">
/* eslint-disable camelcase */
import { throttle } from 'lodash';
import { computed, defineComponent, onMounted, ref, set, watch } from 'vue';

import ApplyHost from '../components/apply-host.vue';

import ListMode from './cluster-list.vue';
import ConnectCluster from './connect-cluster.vue';
import ClusterDetail from './detail.vue';
import { useClusterList, useClusterOperate, useClusterOverview, useTask, useVCluster } from './use-cluster';

import $bkMessage from '@/common/bkmagic';
import $bkInfo from '@/components/bk-magic-2.0/bk-info';
import ConfirmDialog from '@/components/comfirm-dialog.vue';
// import Header from '@/components/layout/Header.vue';
import { ICluster, useAppData, useCluster, useProject } from '@/composables/use-app';
import useSearch from '@/composables/use-search';
import $i18n from '@/i18n/i18n-setup';
import $router from '@/router';
import $store from '@/store';
import ClusterGuide from '@/views/app/cluster-guide.vue';
import TaskList from '@/views/cluster-manage/components/task-list.vue';
import useNode from '@/views/cluster-manage/node-list/use-node';
import ProjectConfig from '@/views/project-manage/project/project-config.vue';

export default defineComponent({
  name: 'ClusterOverview',
  components: {
    ApplyHost,
    ProjectConfig,
    ConfirmDialog,
    TaskList,
    // Header,
    ClusterGuide,
    ListMode,
    ClusterDetail,
    ConnectCluster,
  },
  props: {
    clusterId: {
      type: String,
      default: '',
    },
    active: {
      type: String,
      default: '',
    },
  },
  setup(props) {
    const { flagsMap } = useAppData();
    const { curProject, isMaintainer } = useProject();

    // 集群状态
    const statusTextMap = {
      INITIALIZATION: $i18n.t('generic.status.initializing'),
      DELETING: $i18n.t('generic.status.deleting'),
      'CREATE-FAILURE': $i18n.t('generic.status.createFailed'),
      'DELETE-FAILURE': $i18n.t('generic.status.deleteFailed'),
      'IMPORT-FAILURE': $i18n.t('cluster.status.importFailed'),
    };

    // 集群列表
    const {
      clusterList: clusterData,
      getClusterList,
      clusterExtraInfo,
      webAnnotations,
    } = useClusterList();
    const filterSharedClusterList = computed<ICluster[]>(() => clusterData.value.filter(item => !item.is_shared));
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
      handleChangeDetail(cluster.clusterID, 'overview');
    };
    // 跳转集群信息界面
    const goClusterDetail = async ({ cluster, active }) => {
      handleChangeDetail(cluster.clusterID, active);
    };
    // 跳转添加节点界面
    const goNodeInfo = async (cluster) => {
      handleChangeDetail(cluster.clusterID, 'node');
    };
    // 跳转扩缩容界面
    const goClusterAutoScaler = (cluster) => {
      handleChangeDetail(cluster.clusterID, 'autoscaler');
    };
    // kubeconfig
    const goClusterToken = () => {
      const { href } = $router.resolve({ name: 'token' });
      window.open(href);
    };
    // webconsole
    const { handleGotoConsole } = useCluster();
    const { deleteCluster, retryClusterTask } = useClusterOperate();
    const curOperateCluster = ref<any>(null);
    // 集群删除
    const showConfirmDialog = ref(false);
    const deleteClusterTips = computed(() => {
      if (curOperateCluster.value?.clusterType === 'virtual') {
        return [
          $i18n.t('cluster.button.delete.article1', { clusterName: curOperateCluster.value?.clusterID }),
          $i18n.t('cluster.button.delete.article2'),
          $i18n.t('cluster.button.delete.article3'),
        ];
      }
      return curOperateCluster.value?.clusterCategory === 'importer'
        ? [
          $i18n.t('cluster.button.delete.article1', { clusterName: curOperateCluster.value?.clusterID }),
          $i18n.t('cluster.button.delete.article4'),
          $i18n.t('cluster.button.delete.article5'),
          $i18n.t('cluster.button.delete.article6'),
        ]
        : [
          $i18n.t('cluster.button.delete.article1', { clusterName: curOperateCluster.value?.clusterID }),
          $i18n.t('cluster.button.delete.article7'),
          $i18n.t('cluster.button.delete.article3'),
          $i18n.t('cluster.button.delete.article8'),
        ];
    });
    const { handleDeleteVCluster } = useVCluster();
    const user = computed(() => $store.state.user);
    const confirmDeleteCluster = async () => {
      let result = false;
      if (curOperateCluster.value.clusterType === 'virtual') {
        result = await handleDeleteVCluster({
          operator: user.value.username,
          onlyDeleteInfo: false,
          $clusterId: curOperateCluster.value.clusterID,
        });
      } else {
        result = await deleteCluster(curOperateCluster.value);
      }
      if (result) {
        await handleGetClusterList();
        $bkMessage({
          theme: 'success',
          message: $i18n.t('generic.msg.success.deliveryTask'),
        });
      }
    };
    const handleDeleteCluster = (cluster) => {
      if (
        cluster.clusterType !== 'virtual'
        && cluster.clusterCategory !== 'importer'
        && clusterNodesMap.value[cluster.clusterID]?.length > 0
        && cluster.status === 'RUNNING'
      ) return;

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
      initialzing: $i18n.t('generic.status.initializing'),
      running: $i18n.t('generic.status.running'),
      success: $i18n.t('generic.status.success'),
      failure: $i18n.t('generic.status.failed'),
      timeout: $i18n.t('generic.status.timeout'),
      notstarted: $i18n.t('generic.status.todo'),
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
    const showConnectCluster = ref(false);
    const curRow = ref<ICluster>();
    const handleRetry = async (cluster) => {
      isLoading.value = true;
      // 判断是否是外网连接失败
      const { latestTask } = await taskList(cluster);
      isLoading.value = false;
      const steps = latestTask?.steps || {};
      const connectClusterFailure = Object.keys(steps)
        .some(step => steps?.[step]?.params?.connectCluster);

      if (connectClusterFailure) {
        curRow.value = cluster;
        showConnectCluster.value = true;
      } else {
        retryTask(cluster);
      }
    };
    // 重试任务
    const handleRetryTask = async (cluster) => {
      isLoading.value = true;
      const result = await retryClusterTask(cluster);
      if (result) {
        await handleGetClusterList();
        $bkMessage({
          theme: 'success',
          message: $i18n.t('generic.msg.success.deliveryTask'),
        });
      }
      isLoading.value = false;
    };
    const retryTask = (cluster) => {
      showLogDialog.value = false;
      if (['CREATE-FAILURE', 'DELETE-FAILURE'].includes(cluster.status)) {
        // 创建重试
        $bkInfo({
          type: 'warning',
          title: cluster.status === 'CREATE-FAILURE' ? $i18n.t('cluster.title.retryCreate') :  $i18n.t('cluster.title.confirmDelete'),
          clsName: 'custom-info-confirm default-info',
          subTitle: cluster.clusterName,
          confirmFn: async () => {
            await handleRetryTask(cluster);
          },
        });
      } else {
        $bkMessage({
          theme: 'error',
          message: $i18n.t('generic.status.unknown1'),
        });
      }
    };

    // 集群节点数
    const { getNodeList } = useNode();
    const clusterNodesMap = ref({});
    const handleGetClusterNodes = async () => {
      clusterNodesMap.value = {};
      clusterList.value
        .filter(cluster => webAnnotations.value.perms[cluster.clusterID]?.cluster_manage && cluster.clusterType !== 'virtual')
        .forEach((item) => {
          getNodeList(item.clusterID).then((data) => {
            set(clusterNodesMap.value, item.clusterID, data);
          });
        });
    };
    const throttleClusterNodesFunc = throttle(handleGetClusterNodes, 300);

    // 当前详情tag
    const activeTabName = computed<string>(() => props.active || 'overview');
    // 当前active 集群id
    const activeClusterID = ref(props.clusterId);
    watch(clusterList, () => {
      const activeCluster = clusterList.value.find(item => item.clusterID === activeClusterID.value);
      if (['INITIALIZATION', 'DELETING'].includes(activeCluster?.status)) {
        handleChangeDetail('');
      }
    });
    // 切换详情页
    const clusterDetailRef = ref();
    const handleChangeDetail = async (clusterID: string, active = activeTabName.value) => {
      document.body?.click?.();// 关闭popover
      if (activeClusterID.value === clusterID && activeTabName.value === active) {
        clusterDetailRef.value?.showDetailPanel();
        return;
      };

      await $router.replace({ query: { clusterId: clusterID, active } });
      activeClusterID.value = clusterID;
      clusterDetailRef.value?.showDetailPanel();
    };

    // 详情面板最大宽度
    const detailPanelMaxWidth = ref(1000);
    const minTableWidth = ref(280);
    const contentRef = ref();

    // 详情宽度
    const detailWidth = ref<string|number>('70%');
    const handleDetailWidthChange = (width: string|number) => {
      detailWidth.value = width;
      if (width === 0) {
        throttleClusterNodesFunc();
      }
    };

    // 节点模板
    const goNodeTemplate = () => {
      $router.push({ name: 'nodeTemplate' });
    };

    onMounted(async () => {
      detailPanelMaxWidth.value = contentRef.value.clientWidth - minTableWidth.value;
      await handleGetClusterList();
      await handleGetClusterNodes();
    });

    return {
      curRow,
      showConnectCluster,
      flagsMap,
      activeTabName,
      activeClusterID,
      isMaintainer,
      clusterNodesMap,
      searchValue,
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
      statusTextMap,
      handleGotoConsole,
      contentRef,
      detailPanelMaxWidth,
      detailWidth,
      handleChangeDetail,
      handleDetailWidthChange,
      goNodeTemplate,
      clusterDetailRef,
      handleRetryTask,
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
.log-footer-border-top {
  border-top: 1px solid #DCDEE5;
}
</style>
