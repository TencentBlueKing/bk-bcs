<template>
  <BcsContent :padding="0">
    <div
      class="px-[24px] py-[16px] h-full overflow-x-hidden"
      v-bkloading="{ isLoading, color: '#fafbfd' }"
      ref="contentRef">
      <template v-if="clusterData.length">
        <!-- 撑满屏幕, 让右侧搜索框宽度跟detail详情一致 -->
        <div class="flex mx-[-24px]">
          <div class="flex items-center place-content-between mb-[16px] flex-1 pl-[24px]">
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
            <div class="flex items-center flex-1 ml-[10px] max-w-[480px]">
              <!-- 隐藏共享集群 -->
              <bcs-checkbox
                class="mr-[8px] whitespace-nowrap"
                :value="hideSharedCluster"
                @change="changeSharedClusterVisible">
                {{ $t('cluster.labels.hideSharedCluster') }}
              </bcs-checkbox>
              <bk-input
                right-icon="bk-icon icon-search"
                class="flex-1"
                :placeholder="$t('cluster.placeholder.searchCluster')"
                v-model.trim="searchValue"
                clearable>
              </bk-input>
            </div>
          </div>
          <!-- flex左右布局空div -->
          <div
            :style="{
              width: activeClusterID ? detailWidth : 0
            }"
            class="ml-[24px]">
          </div>
        </div>
        <ListMode
          :cluster-list="curClusterList"
          :overview="clusterOverviewMap"
          :perms="webAnnotations.perms"
          :search-value="searchValue"
          :cluster-extra-info="clusterExtraInfo"
          :cluster-nodes-map="clusterNodesMap"
          :active-cluster-id="activeClusterID"
          :highlight-cluster-id="highlightClusterId"
          class="h-[calc(100%-48px)]"
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
        :namespace="namespace"
        v-if="activeClusterID"
        ref="clusterDetailRef"
        @width-change="handleDetailWidthChange" />
    </div>
    <!-- 集群日志 -->
    <bcs-sideslider
      :is-show.sync="showLogDialog"
      :title="curOperateCluster ? `${curOperateCluster.clusterName} (${curOperateCluster.clusterID})` : '--'"
      :width="960"
      quick-close
      @hidden="handleCloseLog">
      <template #content>
        <TaskList
          v-bkloading="{ isLoading: logLoading }"
          class="px-[24px] py-[20px]"
          :data="taskData"
          @retry="handleRetry(curOperateCluster)"
          @skip="handleSkip">
        </TaskList>
        <div class="bg-[#FAFBFD] h-[48px] flex items-center px-[24px] log-footer-border-top">
          <!-- <bcs-button
            class="w-[88px]"
            theme="primary"
            v-if="['CREATE-FAILURE', 'DELETE-FAILURE'].includes(curOperateCluster.status)"
            @click="handleRetry(curOperateCluster)">
            {{ $t('cluster.ca.nodePool.records.action.retry') }}
          </bcs-button> -->
          <bcs-button class="w-[88px]" @click="showLogDialog = false">{{ $t('generic.button.close') }}</bcs-button>
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
    <!-- 修改集群安全组信息-->
    <bcs-dialog
      v-model="showConnectCluster"
      :show-footer="false"
      width="588"
      render-directive="if">
      <SetConnectInfo
        :cluster="curRow"
        @confirm="handleRetryTask"
        @cancel="showConnectCluster = false" />
    </bcs-dialog>
    <!-- 修改集群管控区域信息-->
    <bcs-dialog
      v-model="showInstallGseAgent"
      :show-footer="false"
      width="588"
      render-directive="if">
      <SetAgentArea
        :cluster="curRow"
        @confirm="handleRetryTask"
        @cancel="showInstallGseAgent = false" />
    </bcs-dialog>
  </BcsContent>
</template>

<script lang="ts">
/* eslint-disable camelcase */
import { throttle } from 'lodash';
import { computed, defineComponent, onMounted, ref, set, watch } from 'vue';

import ApplyHost from '../components/apply-host.vue';

import ListMode from './cluster-list.vue';
import ClusterDetail from './detail.vue';
import SetAgentArea from './set-agent-area.vue';
import SetConnectInfo from './set-connect-info.vue';
import { useClusterList, useClusterOperate, useClusterOverview, useTask, useVCluster } from './use-cluster';

import { clusterMeta } from '@/api/modules/cluster-manager';
import $bkMessage from '@/common/bkmagic';
import $bkInfo from '@/components/bk-magic-2.0/bk-info';
import ConfirmDialog from '@/components/comfirm-dialog.vue';
import BcsContent from '@/components/layout/Content.vue';
import { ICluster, useAppData, useCluster, useProject } from '@/composables/use-app';
import useSearch from '@/composables/use-search';
import $i18n from '@/i18n/i18n-setup';
import $router from '@/router';
import $store from '@/store';
import ClusterGuide from '@/views/app/cluster-guide.vue';
import TaskList from '@/views/cluster-manage/components/task-list.vue';
import ProjectConfig from '@/views/project-manage/project/project-config.vue';

export default defineComponent({
  name: 'ClusterMain',
  components: {
    ApplyHost,
    ProjectConfig,
    ConfirmDialog,
    TaskList,
    ClusterGuide,
    ListMode,
    ClusterDetail,
    SetConnectInfo,
    SetAgentArea,
    BcsContent,
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
    namespace: {
      type: String,
      default: '',
    },
    highlightClusterId: {
      type: String,
      default: '',
    },
  },
  setup(props) {
    const { flagsMap } = useAppData();
    const { curProject, isMaintainer } = useProject();

    const hideSharedCluster = computed(() => $store.state.hideSharedCluster);
    const changeSharedClusterVisible = (v) => {
      $store.commit('updateHideClusterStatus', v);
      handleGetClusterNodes();
    };
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
    const filterSharedClusterList = computed<ICluster[]>(() => clusterData.value
      .filter(item => (hideSharedCluster.value ? !item.is_shared : true)));
    const keys = ref(['name', 'clusterID']);
    const { searchValue, tableDataMatchSearch: curClusterList } = useSearch(filterSharedClusterList, keys);
    const isLoading = ref(false);
    const handleGetClusterList = async () => {
      isLoading.value = true;
      await getClusterList();
      isLoading.value = false;
    };
    // 集群指标
    const { getClusterOverview, clusterOverviewMap } = useClusterOverview(curClusterList);

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
          $i18n.t('cluster.button.delete.article1', { clusterName: `${curOperateCluster.value?.clusterName}(${curOperateCluster.value?.clusterID})` }),
          $i18n.t('cluster.button.delete.article2'),
          $i18n.t('cluster.button.delete.article3'),
        ];
      }
      return curOperateCluster.value?.clusterCategory === 'importer'
        ? [
          $i18n.t('cluster.button.delete.article1', { clusterName: `${curOperateCluster.value?.clusterName}(${curOperateCluster.value?.clusterID})` }),
          $i18n.t('cluster.button.delete.article4'),
          $i18n.t('cluster.button.delete.article5'),
          $i18n.t('cluster.button.delete.article6'),
        ]
        : [
          $i18n.t('cluster.button.delete.article1', { clusterName: `${curOperateCluster.value?.clusterName}(${curOperateCluster.value?.clusterID})` }),
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
    const handleDeleteCluster = (cluster: ICluster) => {
      if (
        cluster.clusterType !== 'virtual'
        && cluster.clusterCategory !== 'importer'
        && clusterNodesMap.value[cluster.clusterID] > 0
        && cluster.status === 'RUNNING'
      ) return;

      curOperateCluster.value = cluster;
      setTimeout(() => {
        showConfirmDialog.value = true;
      }, 0);
    };
    // 集群日志
    const logLoading = ref(false);
    const { taskList, skipTask } = useTask();
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
    const showInstallGseAgent = ref(false);
    const curRow = ref<ICluster>();
    const handleRetry = async (cluster) => {
      isLoading.value = true;
      const { latestTask } = await taskList(cluster);
      isLoading.value = false;
      const steps = latestTask?.steps || {};
      const connectClusterFailure = Object.keys(steps)
        .some(step => steps?.[step]?.params?.connectCluster === 'true');
      const installGseAgent = Object.keys(steps)
        .some(step => steps?.[step]?.params?.installGseAgent === 'true');

      if (connectClusterFailure) {
        // 判断是否是外网连接失败
        curRow.value = cluster;
        showConnectCluster.value = true;
      } else if (installGseAgent) {
        // 判断是否管控区域不正确
        curRow.value = cluster;
        showInstallGseAgent.value = true;
      } else {
        retryTask(cluster);
      }
    };
    // 重试任务
    const handleRetryTask = async (clusterID: string) => {
      isLoading.value = true;
      const result = await retryClusterTask(clusterID);
      if (result) {
        $bkMessage({
          theme: 'success',
          message: $i18n.t('generic.msg.success.deliveryTask'),
        });
        handleGetClusterList();
        if (showLogDialog.value) {
          fetchLogData(clusterData.value.find(item => item.clusterID === clusterID));
        }
      }
      isLoading.value = false;
    };
    const retryTask = (cluster: ICluster) => {
      // 创建重试
      $bkInfo({
        type: 'warning',
        title: $i18n.t('cluster.title.retryTask'),
        clsName: 'custom-info-confirm default-info',
        subTitle: cluster.clusterName,
        confirmFn: async () => {
          await handleRetryTask(cluster.clusterID);
        },
      });
    };

    // 跳过任务
    const handleSkip = (row) => {
      $bkInfo({
        type: 'warning',
        title: $i18n.t('cluster.title.skipTask'),
        clsName: 'custom-info-confirm default-info',
        subTitle: row.taskName || row.name,
        confirmFn: async () => {
          const result = await skipTask(latestTask.value.taskID);
          if (result) {
            $bkMessage({
              theme: 'success',
              message: $i18n.t('generic.msg.success.deliveryTask'),
            });
            handleGetClusterList();
            if (showLogDialog.value) {
              fetchLogData(curOperateCluster.value);
            }
          }
        },
      });
    };

    // 集群节点数
    const clusterNodesMap = ref<Record<string, number>>({});
    const handleGetClusterNodes = async () => {
      const clusterIDs = curClusterList.value.map(item => item.clusterID);
      if (!clusterIDs.length) return;

      clusterNodesMap.value = {};
      const data = await clusterMeta({
        clusters: clusterIDs,
      }).catch(() => []);
      data.map((item) => {
        set(clusterNodesMap.value, item.clusterId, item.clusterNodeNum);
      });
    };
    const throttleClusterNodesFunc = throttle(handleGetClusterNodes, 300);

    // 支持详情页展示的状态
    const supportDetailStatusList = ['CREATE-FAILURE', 'DELETE-FAILURE', 'CONNECT-FAILURE', 'RUNNING'];
    // 当前详情tag
    const activeTabName = computed<string>(() => props.active || 'overview');
    // 当前active 集群id
    const activeClusterID = ref(props.clusterId);
    watch(activeClusterID, () => {
      const isShared = clusterData.value.find(item => item.clusterID === activeClusterID.value)?.is_shared;
      if (!!isShared) {
        changeSharedClusterVisible(false);
      }
    }, { immediate: true });
    watch(curClusterList, () => {
      const activeCluster = curClusterList.value.find(item => item.clusterID === activeClusterID.value);
      if (activeClusterID.value && !supportDetailStatusList.includes(activeCluster?.status)) {
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
      if (!width || width === '0%' || width === '0px') {
        detailWidth.value = '0px';
        throttleClusterNodesFunc();
      } else {
        detailWidth.value = `${width}`;
      }
    };

    // 节点模板
    const goNodeTemplate = () => {
      $router.push({ name: 'nodeTemplate' });
    };

    // scroll cluster into view
    const handleScollActiveClusterIntoView = () => {
      setTimeout(() => {
        const activeDom = document.getElementsByClassName('active-row');
        activeDom[0]?.scrollIntoView();
      });
    };

    onMounted(async () => {
      setTimeout(() => {
        detailPanelMaxWidth.value = document.body.clientWidth - minTableWidth.value;
      });
      handleScollActiveClusterIntoView();
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
      clusterData,
      filterSharedClusterList,
      curClusterList,
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
      handleSkip,
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
      hideSharedCluster,
      changeSharedClusterVisible,
      showInstallGseAgent,
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
