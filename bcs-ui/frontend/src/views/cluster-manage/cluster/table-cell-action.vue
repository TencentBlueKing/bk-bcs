<template>
  <div>
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
              v-if="!(row.clusterCategory === 'importer' && row.importCategory === 'kubeConfig')
                && row.clusterType !== 'federation'"
              @click.stop="handleGotoClusterDetail(row, 'network')">
              {{ $t('cluster.detail.title.network') }}
            </li>
            <li
              class="bcs-dropdown-item"
              v-if="!(row.clusterCategory === 'importer' && row.importCategory === 'kubeConfig')
                && row.clusterType !== 'federation'"
              @click.stop="handleGotoClusterDetail(row, 'master')">
              {{ $t('cluster.detail.title.master') }}
            </li>
            <li
              class="bcs-dropdown-item"
              v-if="row.clusterType !== 'federation'"
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
      <bk-button text class="mr10" @click.stop="handleGotoClusterDetail(row, 'namespace')">
        {{ $t('k8s.namespace') }}
      </bk-button>
      <!-- 共享集群 -->
      <!-- 托管集群 -->
      <template v-if="row.is_shared || row.clusterType === 'virtual'">
        <bk-button
          text
          class="mr10"
          @click.stop="handleGotoDashborad(row)">
          {{ $t('cluster.button.dashboard') }}
        </bk-button>
      </template>
      <!-- 联邦代理集群 -->
      <template v-else-if="row.clusterType === 'federation'">
        <bk-button
          text
          class="mr10"
          @click.stop="handleGotoClusterDetail(row, 'subCluster')">
          {{ $t('cluster.detail.title.subCluster') }}
        </bk-button>
      </template>
      <!-- 独立集群 -->
      <template v-else>
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
          @click.stop="handleGotoClusterNode(row)">
          {{ $t('cluster.detail.title.nodeList') }}
        </bk-button>
      </template>
      <PopoverSelector offset="0, 10">
        <span class="bcs-icon-more-btn"><i class="bcs-icon bcs-icon-more"></i></span>
        <template #content>
          <ul class="bg-[#fff]">
            <li
              class="bcs-dropdown-item"
              v-if="row.clusterType !== 'federation' && !row.is_shared"
              @click.stop="handleGotoClusterOverview(row)">{{ $t('cluster.button.overview') }}</li>
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
                v-if="!(row.clusterCategory === 'importer' && row.importCategory === 'kubeConfig')
                  && row.clusterType !== 'federation'"
                @click.stop="handleGotoClusterDetail(row, 'network')">
                {{ $t('cluster.detail.title.network') }}
              </li>
              <li
                class="bcs-dropdown-item"
                v-if="!(row.clusterCategory === 'importer' && row.importCategory === 'kubeConfig')
                  && !row.is_shared && row.clusterType !== 'federation'"
                @click.stop="handleGotoClusterDetail(row, 'master')">
                {{ $t('cluster.detail.title.master') }}
              </li>
              <li
                class="bcs-dropdown-item"
                v-if="clusterExtraInfo
                  && clusterExtraInfo[row.clusterID]
                  && clusterExtraInfo[row.clusterID].autoScale
                  && !row.is_shared
                  && row.clusterType !== 'federation'"
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
            <template v-if="!row.is_shared">
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
                  { disabled: clusterNodesMap[row.clusterID] && clusterNodesMap[row.clusterID] > 0 }
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
                  disabled: !clusterNodesMap[row.clusterID] || clusterNodesMap[row.clusterID] === 0,
                  placement: 'right'
                }"
                v-else-if="row.clusterType !== 'federation'"
                @click.stop="handleDeleteCluster(row)">
                {{ $t('generic.button.delete') }}
              </li>
            </template>
          </ul>
        </template>
      </PopoverSelector>
    </template>
    <bk-button
      text
      v-else-if="deletable(row) && row.status !== 'DELETING' && row.clusterType !== 'federation'"
      @click.stop="handleDeleteCluster(row)">
      {{ $t('generic.button.delete') }}
    </bk-button>
    <span v-else>--</span>
  </div>
</template>
<script lang="ts">
import { defineComponent, PropType, toRefs } from 'vue';

import PopoverSelector from '@/components/popover-selector.vue';
import { ICluster, useProject } from '@/composables/use-app';

export default defineComponent({
  components: { PopoverSelector },
  props: {
    row: {
      type: Object as PropType<ICluster>,
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
    clusterNodesMap: {
      type: Object,
      default: () => ({}),
    },
  },
  setup(props, { emit }) {
    const { curProject } = useProject();
    const { clusterExtraInfo } = toRefs(props);
    // 正常状态
    const normalStatusList = ['CONNECT-FAILURE', 'RUNNING'];

    function handleShowClusterLog(row: ICluster) {
      emit('show-cluster-log', row);
    }
    function handleRetry(row: ICluster) {
      emit('retry', row);
    }
    function handleGotoClusterDetail(row: ICluster, active = 'info') {
      emit('goto-cluster-detail', row, active);
    }
    function handleDeleteCluster(row: ICluster) {
      emit('delete-cluster', row);
    }
    function handleGotoDashborad(row: ICluster) {
      emit('goto-dashboard', row);
    }
    function handleGotoClusterNode(row: ICluster) {
      emit('goto-cluster-node', row);
    }
    function handleGotoClusterOverview(row: ICluster) {
      emit('goto-cluster-overview', row);
    }
    function handleGotoClusterCA(row: ICluster) {
      emit('goto-cluster-ca', row);
    }
    function handleGotoToken() {
      emit('goto-token');
    }
    function handleGotoWebConsole(row: ICluster) {
      emit('goto-web-console', row);
    }
    function deletable(row: ICluster) {
      return !!clusterExtraInfo.value[row.clusterID]?.canDeleted;
    }

    return {
      curProject,
      normalStatusList,
      handleShowClusterLog,
      handleRetry,
      handleGotoClusterDetail,
      handleDeleteCluster,
      handleGotoDashborad,
      handleGotoClusterNode,
      handleGotoClusterOverview,
      handleGotoClusterCA,
      handleGotoToken,
      handleGotoWebConsole,
      deletable,
    };
  },
});
</script>
