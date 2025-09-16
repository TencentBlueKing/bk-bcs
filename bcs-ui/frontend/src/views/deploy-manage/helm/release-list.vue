<template>
  <BcsContent :title="$t('deploy.helm.releases')" hide-back>
    <Row>
      <template #right>
        <ClusterSelect v-model="clusterID" cluster-type="all" @change="handleClusterChange"></ClusterSelect>
        <NamespaceSelect
          :cluster-id="clusterID"
          class="w-[250px] ml-[5px]"
          :clearable="true"
          v-model="ns"
          @change="handleResetList">
        </NamespaceSelect>
        <bcs-input
          right-icon="bk-icon icon-search"
          class="w-[360px] ml-[5px]"
          :placeholder="$t('generic.placeholder.searchName')"
          clearable
          v-model="searchName">
        </bcs-input>
      </template>
    </Row>
    <bcs-table
      class="mt-[20px]"
      :pagination="pagination"
      :data="releaseList"
      v-bkloading="{ isLoading: loading }"
      @page-change="pageChange"
      @page-limit-change="pageSizeChange">
      <bcs-table-column :label="$t('generic.label.name')" min-width="100">
        <template #default="{ row }">
          <bcs-button
            text
            :disabled="!row.repo"
            v-authority="{
              clickable: webAnnotationsPerms[row.iamNamespaceID]
                && webAnnotationsPerms[row.iamNamespaceID].namespace_scoped_view,
              actionId: 'namespace_scoped_view',
              resourceName: row.namespace,
              disablePerms: true,
              permCtx: {
                project_id: projectID,
                cluster_id: clusterID,
                name: row.namespace
              }
            }"
            @click="handleUpdate(row)">
            <span
              class="bcs-ellipsis"
              v-bk-tooltips="{
                content: $t('deploy.helm.nonPlatformReleaseUpdate'),
                disabled: row.repo
              }">
              {{row.name}}
            </span>
          </bcs-button>
        </template>
      </bcs-table-column>
      <bcs-table-column :label="$t('generic.label.status')" prop="status" width="150">
        <template #default="{ row }">
          <StatusIcon
            :status="row.status"
            :status-color-map="releaseStatusColorMap"
            :status-text-map="statusTextMap"
            :pending="[
              'uninstalling',
              'pending-install',
              'pending-upgrade',
              'pending-rollback'
            ].includes(row.status)">
            <span
              v-bk-tooltips="{
                content: row.message,
                disabled: !row.message ||
                  ![
                    'failed',
                    'failed-install',
                    'failed-upgrade',
                    'failed-rollback',
                    'failed-uninstall'
                  ].includes(row.status),
                theme: 'bcs-tippy'
              }"
              :class="row.message && [
                'failed',
                'failed-install',
                'failed-upgrade',
                'failed-rollback',
                'failed-uninstall'
              ].includes(row.status) ? 'border-dashed border-0 border-b' : ''">
              {{statusTextMap[row.status]}}
            </span>
          </StatusIcon>
        </template>
      </bcs-table-column>
      <bcs-table-column label="Chart" prop="chart" min-width="200" show-overflow-tooltip>
        <template #default="{ row }">
          <span>{{`${row.chart}:${row.chartVersion}`}}</span>
        </template>
      </bcs-table-column>
      <bcs-table-column :label="$t('generic.label.cluster1')" show-overflow-tooltip>
        <template #default>
          <span>{{clusterName}}</span>
        </template>
      </bcs-table-column>
      <bcs-table-column :label="$t('k8s.namespace')" prop="namespace" show-overflow-tooltip></bcs-table-column>
      <bcs-table-column :label="$t('generic.label.updator')" prop="updateBy" width="130">
        <template #default="{ row }">
          <bk-user-display-name :user-id="row.updateBy"></bk-user-display-name>
        </template>
      </bcs-table-column>
      <bcs-table-column :label="$t('cluster.labels.updatedAt')" prop="updateTime" width="200"></bcs-table-column>
      <bcs-table-column :label="$t('generic.label.action')" width="220" fixed="right">
        <template #default="{ row }">
          <bcs-button
            text
            v-authority="{
              clickable: webAnnotationsPerms[row.iamNamespaceID]
                && webAnnotationsPerms[row.iamNamespaceID].namespace_scoped_view,
              actionId: 'namespace_scoped_view',
              resourceName: row.namespace,
              disablePerms: true,
              permCtx: {
                project_id: projectID,
                cluster_id: clusterID,
                name: row.namespace
              }
            }"
            @click="handleViewStatus(row)">{{ $t('generic.label.status') }}</bcs-button>
          <bcs-button
            text
            class="ml-[10px]"
            v-authority="{
              clickable: webAnnotationsPerms[row.iamNamespaceID]
                && webAnnotationsPerms[row.iamNamespaceID].namespace_scoped_view,
              actionId: 'namespace_scoped_view',
              resourceName: row.namespace,
              disablePerms: true,
              permCtx: {
                project_id: projectID,
                cluster_id: clusterID,
                name: row.namespace
              }
            }"
            @click="handleViewHistory(row)">{{ $t('deploy.helm.history') }}</bcs-button>
          <bk-popover
            placement="bottom"
            theme="light dropdown"
            :arrow="false"
            class="ml-[5px]"
            trigger="click">
            <span class="bcs-icon-more-btn">
              <i class="bcs-icon bcs-icon-more"></i>
            </span>
            <template #content>
              <ul>
                <li
                  :class="['bcs-dropdown-item', { disabled: !row.repo }]"
                  v-authority="{
                    clickable: webAnnotationsPerms[row.iamNamespaceID]
                      && webAnnotationsPerms[row.iamNamespaceID].namespace_scoped_update,
                    actionId: 'namespace_scoped_update',
                    resourceName: row.namespace,
                    disablePerms: true,
                    permCtx: {
                      project_id: projectID,
                      cluster_id: clusterID,
                      name: row.namespace
                    }
                  }"
                  v-bk-tooltips="{
                    content: $t('deploy.helm.disableAction'),
                    disabled: row.repo,
                    placement: 'left'
                  }"
                  @click="handleUpdate(row)">
                  {{$t('generic.button.update')}}
                </li>
                <li
                  v-if="row.revision !== 0 "
                  class="bcs-dropdown-item"
                  v-authority="{
                    clickable: webAnnotationsPerms[row.iamNamespaceID]
                      && webAnnotationsPerms[row.iamNamespaceID].namespace_scoped_view,
                    actionId: 'namespace_scoped_view',
                    resourceName: row.namespace,
                    disablePerms: true,
                    permCtx: {
                      project_id: projectID,
                      cluster_id: clusterID,
                      name: row.namespace
                    }
                  }"
                  @click="handleShowManifest(row)">Manifest</li>
                <li
                  :class="['bcs-dropdown-item', { disabled: !row.repo }]"
                  v-authority="{
                    clickable: webAnnotationsPerms[row.iamNamespaceID]
                      && webAnnotationsPerms[row.iamNamespaceID].namespace_scoped_update,
                    actionId: 'namespace_scoped_update',
                    resourceName: row.namespace,
                    disablePerms: true,
                    permCtx: {
                      project_id: projectID,
                      cluster_id: clusterID,
                      name: row.namespace
                    }
                  }"
                  v-bk-tooltips="{
                    content: $t('deploy.helm.disableAction'),
                    disabled: row.repo,
                    placement: 'left'
                  }"
                  @click="handleShowRollback(row)">
                  {{$t('deploy.helm.rollback')}}
                </li>
                <li
                  :class="['bcs-dropdown-item', { disabled: !row.repo && row.namespace === 'kube-system' }]"
                  v-authority="{
                    clickable: webAnnotationsPerms[row.iamNamespaceID]
                      && webAnnotationsPerms[row.iamNamespaceID].namespace_scoped_delete,
                    actionId: 'namespace_scoped_delete',
                    resourceName: row.namespace,
                    disablePerms: true,
                    permCtx: {
                      project_id: projectID,
                      cluster_id: clusterID,
                      name: row.namespace
                    }
                  }"
                  v-bk-tooltips="{
                    content: $t('deploy.helm.disableNS'),
                    disabled: row.repo || row.namespace !== 'kube-system',
                    placement: 'left'
                  }"
                  @click="handleDelete(row)">
                  {{$t('generic.button.delete')}}
                </li>
              </ul>
            </template>
          </bk-popover>
        </template>
      </bcs-table-column>
      <template #empty>
        <BcsEmptyTableStatus :type="searchName ? 'search-empty' : 'empty'" @clear="handleClearSearchData" />
      </template>
    </bcs-table>
    <!-- release状态 -->
    <bcs-sideslider
      :is-show.sync="showStatus"
      quick-close
      :width="800">
      <template #header>
        <Row class="pr-[30px]">
          <template #left>{{curRow.name}}</template>
          <template #right>
            <bcs-input
              class="w-[300px]"
              right-icon="bk-icon icon-search"
              :placeholder="$t('generic.placeholder.searchName')"
              clearable
              v-model="resourceName">
            </bcs-input>
          </template>
        </Row>
      </template>
      <template #content>
        <div class="bcs-sideslider-content" v-bkloading="{ isLoading: statusLoading }">
          <bcs-table :data="filterStatusData">
            <bcs-table-column :label="$t('generic.label.name')" prop="metadata.name">
              <template #default="{ row }">
                <span
                  v-bk-tooltips="{
                    disabled: !!row.metadata.uid && categoryMap[row.kind],
                    content: !categoryMap[row.kind] ? $t('deploy.helm.disableLink') : $t('deploy.helm.notFound')
                  }">
                  <bcs-button
                    text
                    :disabled="!row.metadata.uid || !categoryMap[row.kind]"
                    @click="handleGotoResourceDetail(row)">
                    {{row.metadata.name}}
                  </bcs-button>
                </span>
              </template>
            </bcs-table-column>
            <bcs-table-column :label="$t('generic.label.kind')" prop="kind">
              <template #default="{ row }">
                <bcs-tag theme="info">{{row.kind}}</bcs-tag>
              </template>
            </bcs-table-column>
            <bcs-table-column label="Pods" width="100">
              <template #default="{ row }">
                {{ row.status ? `${row.status.readyReplicas || '-'} / ${row.status.replicas || '-'}` : '- / -' }}
              </template>
            </bcs-table-column>
            <template #empty>
              <BcsEmptyTableStatus :type="resourceName ? 'search-empty' : 'empty'" @clear="resourceName = ''" />
            </template>
          </bcs-table>
        </div>
      </template>
    </bcs-sideslider>
    <!-- 回滚 -->
    <bcs-dialog
      header-position="left"
      :title="$t('deploy.helm.rollbackName', { name: curRow.name })"
      :width="1000"
      v-model="showRollback">
      <div class="mb-[5px]">{{ $t('deploy.helm.rollbackToVer') }}</div>
      <bcs-select class="mb-[10px]" :clearable="false" :loading="historyLoading" v-model="revision">
        <bcs-option
          v-for="item in historyData"
          :key="item.revision"
          :id="item.revision"
          :name="item.revision">
          <span>{{
            $t('deploy.helm.releaseInfo',
               {
                 version: item.revision,
                 time: item.updateTime,
                 chartVersion: item.chartVersion
               })
          }}
          </span>
        </bcs-option>
      </bcs-select>
      <div class="flex mb-[5px]">
        <span class="flex-1">{{$t('generic.title.curVersion')}}</span>
        <span class="flex-1">{{$t('deploy.helm.rollbackVer')}}</span>
      </div>
      <CodeEditor
        v-bkloading="{ isLoading: diffLoading }"
        diff-editor
        height="460px"
        full-screen
        readonly
        :value="diffData.newContent"
        :original="diffData.oldContent" />
      <template #footer>
        <bcs-button
          theme="primary"
          :disabled="!revision"
          :loading="confirmRollbackLoading"
          @click="handleConfirmRollback">{{$t('generic.button.confirm')}}</bcs-button>
        <bcs-button
          :loading="confirmRollbackLoading
          " @click="showRollback = false">{{$t('generic.button.cancel')}}</bcs-button>
      </template>
    </bcs-dialog>
    <!-- 更新记录 -->
    <bcs-sideslider
      :width="950"
      quick-close
      :is-show.sync="showReleaseHistory">
      <template #header>
        <div class="flex items-center">
          <span class="text-[16px] text-[#313238]">{{ $t('deploy.helm.releaseHistory') }}</span>
          <bcs-divider direction="vertical"></bcs-divider>
          <span class="text-[#979BA5] text-[12px]">{{ curRow ? curRow.name : '--' }}</span>
        </div>
      </template>
      <template #content>
        <div class="px-[24px] py-[20px]">
          <bcs-table
            :data="curPageHistoryData"
            :pagination="historyPageConfig"
            v-bkloading="{ isLoading: historyLoading }"
            max-height="600"
            @row-mouse-enter="handleMouseEnter"
            @row-mouse-leave="handleMouseLeave"
            @page-change="historyPageChange"
            @page-limit-change="historyPageSizeChange">
            <bcs-table-column label="Revision" prop="revision" width="90"></bcs-table-column>
            <bcs-table-column
              :label="$t('cluster.labels.updatedAt')"
              prop="updateTime"
              width="180"
              show-overflow-tooltip>
            </bcs-table-column>
            <bcs-table-column
              :label="$t('generic.label.status')"
              prop="status"
              show-overflow-tooltip></bcs-table-column>
            <bcs-table-column label="Chart" prop="chart" show-overflow-tooltip>
              <template #default="{ row }">
                {{`${row.chart}:${row.chartVersion}`}}
              </template>
            </bcs-table-column>
            <bcs-table-column label="App Version" width="120" prop="appVersion"></bcs-table-column>
            <bcs-table-column label="Values" width="90">
              <template #default="{ row }">
                <bcs-button
                  text
                  @click="handleShowValuesDetail(row)"
                  v-if="row.values">{{$t('generic.button.view')}}</bcs-button>
                <span v-else>--</span>
                <span
                  class="bcs-icon-btn ml5"
                  v-if="activeRevision === row.revision"
                  v-bk-tooltips="$t('deploy.helm.copyValues')"
                  @click="handleCopyValues(row)">
                  <i class="bcs-icon bcs-icon-copy"></i>
                </span>
              </template>
            </bcs-table-column>
            <bcs-table-column label="Manifest" width="90">
              <template #default="{ row }">
                <bcs-button
                  text
                  @click="handleShowManifest(row)"
                  v-if="row.values">{{$t('generic.button.view')}}</bcs-button>
              </template>
            </bcs-table-column>
            <bcs-table-column
              :label="$t('deploy.helm.upgradeDesc')"
              prop="description"
              min-width="160"
              show-overflow-tooltip>
            </bcs-table-column>
          </bcs-table>
        </div>
      </template>
    </bcs-sideslider>
    <!-- values内容 -->
    <bcs-sideslider
      :width="950"
      quick-close
      :is-show.sync="showValuesDialog">
      <template #header>
        <div class="flex items-center">
          <span class="text-[16px] text-[#313238]">Values</span>
          <bcs-divider direction="vertical"></bcs-divider>
          <span class="text-[#979BA5] text-[12px]">{{ `Revision ${curRevision}` }}</span>
        </div>
      </template>
      <template #content>
        <div class="max-h-[calc(100vh-60px)] overflow-hidden px-[24px] py-[20px]">
          <CodeEditor
            class="min-h-[calc(100vh-100px)]"
            v-model="curValues"
            v-full-screen="{
              tools: ['fullscreen', 'copy'],
              content: curValues
            }"
            readonly />
        </div>
      </template>
    </bcs-sideslider>
    <!-- Manifest文件列表 -->
    <bcs-sideslider
      :width="950"
      quick-close
      :is-show.sync="showManifest">
      <template #header>
        <div class="flex items-center">
          <span class="text-[16px] text-[#313238]">{{ $t('deploy.helm.title.manifestFile') }}</span>
          <bcs-divider direction="vertical"></bcs-divider>
          <span class="text-[#979BA5] text-[12px]">
            {{ showReleaseHistory ? `Revision ${curRevision}` : curRow.name }}
          </span>
        </div>
      </template>
      <template #content>
        <div class="max-h-[calc(100vh-60px)] overflow-hidden px-[24px] py-[20px]">
          <ChartFileTree class="h-[calc(100vh-100px)]" :contents="manifest" />
        </div>
      </template>
    </bcs-sideslider>
  </BcsContent>
</template>
<script lang="ts">
import { computed, defineComponent, onMounted, ref, watch } from 'vue';

import ChartFileTree from './chart-file-tree.vue';
import useHelm from './use-helm';

import $bkMessage from '@/common/bkmagic';
import { copyText } from '@/common/util';
import $bkInfo from '@/components/bk-magic-2.0/bk-info';
import ClusterSelect from '@/components/cluster-selector/cluster-select.vue';
import BcsContent from '@/components/layout/Content.vue';
import Row from '@/components/layout/Row.vue';
import CodeEditor from '@/components/monaco-editor/new-editor.vue';
import NamespaceSelect from '@/components/namespace-selector/namespace-select.vue';
import StatusIcon from '@/components/status-icon';
import { useCluster, useProject } from '@/composables/use-app';
import useDebouncedRef from '@/composables/use-debounce';
import useInterval from '@/composables/use-interval';
import usePageConf from '@/composables/use-page';
import fullScreen from '@/directives/full-screen';
import $i18n from '@/i18n/i18n-setup';
import $router from '@/router';
import $store from '@/store';

export default defineComponent({
  name: 'ReleaseList',
  directives: {
    'full-screen': fullScreen,
  },
  components: {
    BcsContent,
    Row,
    ClusterSelect,
    NamespaceSelect,
    CodeEditor,
    StatusIcon,
    ChartFileTree,
  },
  props: {
    namespace: {
      type: String,
      default: '',
    },
    name: {
      type: String,
      default: '',
    },
  },
  setup(props) {
    const {
      handleGetReleasesList,
      handleGetReleaseStatus,
      handleGetReleaseHistory,
      handleDeleteRelease,
      handleRollbackRelease,
      handlePreviewRelease,
      handleGetManifest,
    } = useHelm();

    const { clusterList } = useCluster();
    const { projectID } = useProject();
    const { currentRoute } = $router;
    const clusterID = ref(currentRoute.query.clusterId || $store.getters.curClusterId);
    const clusterName = computed(() => clusterList.value.find(item => item.clusterID === clusterID.value)?.clusterName);
    const ns = ref<string>(props.namespace || $store.state.curNamespace);
    const searchName = useDebouncedRef<string>(props.name, 360);

    // release 列表
    const loading = ref(false);
    const releaseList = ref<any[]>([]);
    const webAnnotationsPerms = ref({});
    const releaseStatusColorMap = ref({
      unknown: 'red',
      deployed: 'green',
      uninstalled: 'gray',
      superseded: 'gray',
      failed: 'red',
      'failed-install': 'red',
      'failed-upgrade': 'red',
      'failed-rollback': 'red',
      'failed-uninstall': 'red',
    });
    const statusTextMap = ref({
      unknown: $i18n.t('generic.status.error'),
      deployed: $i18n.t('generic.status.ready'),
      uninstalled: $i18n.t('generic.status.deleted'),
      superseded: $i18n.t('deploy.helm.invalidate'),
      failed: $i18n.t('generic.status.failed'),
      uninstalling: $i18n.t('generic.status.deleting'),
      'pending-install': $i18n.t('deploy.helm.pending'),
      'pending-upgrade': $i18n.t('generic.status.updating'),
      'pending-rollback': $i18n.t('deploy.helm.pendingRollback'),
      'failed-install': $i18n.t('deploy.helm.failed'),
      'failed-upgrade': $i18n.t('generic.status.updateFailed'),
      'failed-rollback': $i18n.t('deploy.helm.rollbackFailed'),
      'failed-uninstall': $i18n.t('generic.status.deleteFailed'),
    });
    const pagination = ref({
      count: 0,
      current: 1,
      limit: 10,
    });
    const { start, stop } = useInterval(() => handleGetList(false));
    watch(releaseList, () => {
      if (releaseList.value.some(item => [
        'uninstalling',
        'pending-install',
        'pending-upgrade',
        'pending-rollback',
      ].includes(item.status))) {
        start();
      } else {
        stop();
      }
    });
    // 搜索
    watch(searchName, () => {
      handleResetList();
    });
    const handleResetList = () => {
      pagination.value.current = 1;
      handleGetList();
    };
    const handleClusterChange = async (val) => {
      ns.value = '';
      if (val) {
        await $router.replace({
          query: {
            ...currentRoute.query,
            clusterId: val,
          },
        });
      }
      handleResetList();
    };

    const handleGetList = async (defaultLoading = true) => {
      if (!clusterID.value) return;
      loading.value = defaultLoading;
      const res = await handleGetReleasesList({
        $clusterId: clusterID.value,
        namespace: ns.value,
        name: searchName.value,
        page: pagination.value.current,
        size: pagination.value.limit,
      });
      releaseList.value = res?.data?.data || [];
      webAnnotationsPerms.value = res?.web_annotations?.perms || {};
      pagination.value.count = res.data.total;
      loading.value = false;
    };
    const pageChange = (page: number) => {
      pagination.value.current = page;
      handleGetList();
    };
    const pageSizeChange = (size: number) => {
      pagination.value.current = 1;
      pagination.value.limit = size;
      handleGetList();
    };

    const curRow = ref<Record<string, any>>({});
    // 查看release状态
    const showStatus = ref(false);
    const resourceName = ref('');
    const statusData = ref<any[]>([]);
    const statusLoading = ref(false);
    const filterStatusData = computed(() => statusData.value
      .filter(item => item.metadata.name?.includes(resourceName.value)));
    const handleViewStatus = async (row) => {
      statusLoading.value = true;
      curRow.value = row;
      showStatus.value = true;
      statusData.value = await handleGetReleaseStatus({
        $clusterId: clusterID.value,
        $namespaceId: row.namespace,
        $releaseName: row.name,
      });
      statusLoading.value = false;
    };
    // 支持跳转的资源（后续可扩展）
    const categoryMap = {
      CronJob: 'cronjobs',
      Deployment: 'deployments',
      DaemonSet: 'daemonsets',
      Job: 'jobs',
      Pod: 'pods',
      StatefulSet: 'statefulsets',
      GameDeployment: 'custom_objects',
      GameStatefulSet: 'custom_objects',
    };
    const crdMap = {
      GameDeployment: 'gamedeployments.tkex.tencent.com',
      GameStatefulSet: 'gamestatefulsets.tkex.tencent.com',
    };
    const handleGotoResourceDetail = (row) => {
      const { href } = $router.resolve({
        name: 'dashboardWorkloadDetail',
        params: {
          category: categoryMap[row.kind],
          name: row.metadata.name,
          namespace: row.metadata.namespace,
          clusterId: clusterID.value,
        },
        query: {
          kind: row.kind,
          crd: crdMap[row.kind],
        },
      });
      window.open(href);
    };

    // 更新记录
    const showReleaseHistory = ref(false);
    const historyLoading = ref(false);
    const historyData = ref<any[]>([]);
    const {
      pagination: historyPageConfig,
      pageChange: historyPageChange,
      pageSizeChange: historyPageSizeChange,
      curPageData: curPageHistoryData,
    } = usePageConf(historyData);
    const handleViewHistory = (row) => {
      curRow.value = row;
      showReleaseHistory.value = true;
      handleGetHistoryList();
    };
    const handleGetHistoryList = async (filter?: string) => {
      historyLoading.value = true;
      const { namespace, name } = curRow.value;
      historyData.value = await handleGetReleaseHistory(clusterID.value, namespace, name, filter);
      historyLoading.value = false;
    };
    const activeRevision = ref('');
    const handleMouseEnter = (index, event, row) => {
      activeRevision.value = row.revision;
    };
    const handleMouseLeave = () => {
      activeRevision.value = '';
    };
    const showValuesDialog = ref(false);
    const curValues = ref('');
    const curRevision = ref('');
    const handleShowValuesDetail = (row) => {
      showValuesDialog.value = true;
      curValues.value = row.values;
      curRevision.value = row.revision;
    };
    const handleCopyValues = (row) => {
      copyText(row.values);
      $bkMessage({
        theme: 'success',
        message: $i18n.t('generic.msg.success.copy'),
      });
    };

    // 更新
    const handleUpdate = (row) => {
      if (!row.repo) return;
      $router.push({
        name: 'updateRelease',
        params: {
          repoName: row.repo,
          cluster: clusterID.value,
          chartName: row.chart,
          releaseName: row.name,
          namespace: row.namespace,
        },
      });
    };

    // manifest
    const manifest = ref<Record<string, {
      content: string
      name: string
      path: string
    }>>({});
    const manifestLoading = ref(false);
    const showManifest = ref(false);
    const handleShowManifest = async (row) => {
      curRow.value = row;
      curRevision.value = row.revision;
      showManifest.value = true;

      manifestLoading.value = true;
      const { name, revision, namespace } = row;
      manifest.value = await handleGetManifest({
        $clusterId: clusterID.value,
        $releaseName: name,
        $namespaceId: namespace,
        $revision: revision,
      });
      manifestLoading.value = false;
    };

    // 回滚
    const revision = ref();
    const showRollback = ref(false);
    const diffData = ref<Record<string, any>>({});
    const diffLoading = ref(false);
    const confirmRollbackLoading = ref(false);
    watch(revision, async () => {
      if (!revision.value) return;
      const { namespace, name } = curRow.value || {};
      diffLoading.value = true;
      diffData.value = await handlePreviewRelease({
        $clusterId: clusterID.value,
        $namespaceId: namespace,
        $releaseName: name,
        revision: revision.value,
      });
      diffLoading.value = false;
    });
    const handleShowRollback = (row) => {
      if (!row.repo) return;
      revision.value = '';
      diffData.value = {};
      curRow.value = row;
      showRollback.value = true;
      handleGetHistoryList('superseded');
    };
    const handleConfirmRollback = async () => {
      confirmRollbackLoading.value = true;
      const result = await handleRollbackRelease({
        $clusterId: clusterID.value,
        $namespaceId: curRow.value.namespace,
        $releaseName: curRow.value.name,
        revision: revision.value,
      });
      confirmRollbackLoading.value = false;
      if (result) {
        showRollback.value = false;
        handleGetList();
      }
    };

    // 删除
    const handleDelete = (row) => {
      // 不允许删除kube system下面的系统内置chart
      if (!row.repo && row.namespace === 'kube-system') return;
      $bkInfo({
        type: 'warning',
        clsName: 'custom-info-confirm',
        title: $i18n.t('generic.title.confirmDelete'),
        subTitle: $i18n.t('generic.title.confirmDelete1', { name: row.name }),
        defaultInfo: true,
        confirmFn: async () => {
          const { namespace, name } = row;
          const result =  await handleDeleteRelease(clusterID.value, namespace, name);
          if (result) {
            pagination.value.current = 1;
            handleGetList();
          }
        },
      });
    };

    const handleClearSearchData = () => {
      searchName.value = '';
      handleResetList();
    };

    onMounted(() => {
      handleGetList();
    });

    return {
      projectID,
      categoryMap,
      curRow,
      clusterID,
      clusterName,
      ns,
      searchName,
      pagination,
      loading,
      releaseStatusColorMap,
      statusTextMap,
      webAnnotationsPerms,
      releaseList,
      showStatus,
      resourceName,
      filterStatusData,
      statusLoading,
      showReleaseHistory,
      historyLoading,
      historyPageConfig,
      historyData,
      diffData,
      diffLoading,
      confirmRollbackLoading,
      curPageHistoryData,
      historyPageChange,
      historyPageSizeChange,
      activeRevision,
      handleMouseEnter,
      handleMouseLeave,
      handleShowValuesDetail,
      handleCopyValues,
      revision,
      showValuesDialog,
      curValues,
      pageChange,
      pageSizeChange,
      handleViewStatus,
      handleViewHistory,
      handleUpdate,
      showRollback,
      handleShowRollback,
      handleConfirmRollback,
      handleDelete,
      handleGotoResourceDetail,
      handleClusterChange,
      handleResetList,
      handleClearSearchData,
      handleShowManifest,
      showManifest,
      manifest,
      curRevision,
    };
  },
});
</script>
