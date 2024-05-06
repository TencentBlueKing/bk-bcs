<template>
  <BcsContent :title="nodeName">
    <!-- 节点信息 -->
    <div class="flex border bcs-border rounded-sm text-[14px] bg-[#fafbfd]">
      <div
        v-for="(item, index) in infoList"
        :key="item.prop"
        :class="[
          'flex flex-col justify-center flex-1 h-[75px] px-[20px]',
          index < (infoList.length - 1) ? 'bcs-border-right' : ''
        ]">
        <div class="bcs-ellipsis font-bold">{{ item.label }}:</div>
        <div class="bcs-ellipsis mt-[8px]" v-bk-overflow-tips>
          {{ (item.format ? item.format(nodeInfo[item.prop]) : nodeInfo[item.prop]) || '--' }}
        </div>
      </div>
    </div>
    <!-- 节点指标 -->
    <div class="grid grid-cols-2 grid-rows-2 bcs-border mt-[20px] bg-[#fff]">
      <Metric
        class="bcs-border-right bcs-border-bottom"
        :title="$t('cluster.nodeList.metric.cpuUsageAndRequestUsage')"
        :metric="['cpu_usage', 'cpu_request_usage']"
        :params="{ $nodeIP: nodeName, $clusterId: clusterId }"
        :colors="['#3a84ff', '#30d878']"
        category="nodes">
      </Metric>
      <Metric
        class="bcs-border-bottom"
        :title="$t('cluster.nodeList.metric.memUsageAndRequestUsage')"
        :metric="['memory_usage', 'memory_request_usage']"
        :params="{ $nodeIP: nodeName, $clusterId: clusterId }"
        :colors="['#853cff', '#3ede78']"
        category="nodes">
      </Metric>
      <Metric
        class="bcs-border-right"
        :title="$t('k8s.networking')"
        :metric="['network_receive', 'network_transmit']"
        :params="{ $nodeIP: nodeName, $clusterId: clusterId }"
        :colors="['#3ede78', '#853cff']"
        unit="byte"
        category="nodes">
      </Metric>
      <Metric
        :title="$t('cluster.nodeList.metric.diskIOAndUsage')"
        :metric="['disk_usage', 'diskio_usage']"
        :params="{ $nodeIP: nodeName, $clusterId: clusterId }"
        :colors="['#853cff', '#30d878']"
        category="nodes">
      </Metric>
    </div>
    <!-- Pods & 事件 -->
    <bcs-tab class="mt20" type="card" :label-height="42">
      <bcs-tab-panel name="pod" label="Pods">
        <Row class="mb-[20px]">
          <template #right>
            <span class="bcs-form-prefix bg-[#f5f7fa]">{{$t('k8s.namespace')}}</span>
            <bcs-select
              class="w-[200px]"
              v-model="namespaceValue"
              :loading="namespaceLoading"
              searchable
              clearable
              :placeholder="' '">
              <bcs-option key="all" id="" :name="$t('cluster.nodeList.detail.label.selectAllNs')"></bcs-option>
              <bcs-option
                v-for="option in namespaceList"
                :key="option.name"
                :id="option.name"
                :name="option.name">
              </bcs-option>
            </bcs-select>
            <bk-input
              class="ml5 w-[350px]"
              clearable
              v-model="searchValue"
              right-icon="bk-icon icon-search"
              :placeholder="$t('generic.placeholder.searchPod')">
            </bk-input>
          </template>
        </Row>
        <bk-table
          :data="curPodsData"
          :pagination="pagination"
          v-bkloading="{ isLoading: podLoading }"
          @page-change="pageChange"
          @page-limit-change="pageSizeChange"
          @sort-change="handleSortChange">
          <bk-table-column :label="$t('generic.label.name')" min-width="130" sortable fixed="left">
            <template #default="{ row }">
              <bk-button
                class="bcs-button-ellipsis"
                text
                v-authority="{
                  clickable: podsWebAnnotations.perms.items[row.uid].detailBtn.clickable,
                  actionId: 'namespace_scoped_view',
                  resourceName: row.namespace,
                  disablePerms: true,
                  permCtx: {
                    project_id: projectID,
                    cluster_id: clusterId,
                    name: row.namespace
                  }
                }"
                @click="gotoPodDetail(row)">
                {{ row.name }}
              </bk-button>
            </template>
          </bk-table-column>
          <bk-table-column :label="$t('k8s.namespace')" min-width="100" sortable>
            <template #default="{ row }">
              <span>{{ row.namespace }}</span>
            </template>
          </bk-table-column>
          <bk-table-column :label="$t('k8s.image')" min-width="200" :show-overflow-tooltip="false">
            <template #default="{ row }">
              <span v-bk-tooltips.top="(row.images || []).join('<br />')">
                {{ (row.images || []).join(', ') }}
              </span>
            </template>
          </bk-table-column>
          <bk-table-column label="Status" width="140">
            <template #default="{ row }">
              <StatusIcon :status="row.status"></StatusIcon>
            </template>
          </bk-table-column>
          <bk-table-column label="Ready" width="100">
            <template #default="{ row }">
              {{row.readyCnt}}/{{row.totalCnt}}
            </template>
          </bk-table-column>
          <bk-table-column label="Restarts" width="100">
            <template #default="{ row }">{{row.restartCnt}}</template>
          </bk-table-column>
          <bk-table-column label="Host IP" min-width="140">
            <template #default="{ row }">{{row.hostIP || '--'}}</template>
          </bk-table-column>
          <bk-table-column label="Pod IPv4" width="140">
            <template #default="{ row }">{{row.podIPv4 || '--'}}</template>
          </bk-table-column>
          <bk-table-column label="Pod IPv6" min-width="200">
            <template #default="{ row }">{{row.podIPv6 || '--'}}</template>
          </bk-table-column>
          <bk-table-column label="Node">
            <template #default="{ row }">{{row.node || '--'}}</template>
          </bk-table-column>
          <bk-table-column label="Age" sortable prop="createTime">
            <template #default="{ row }">
              <span>{{row.age || '--'}}</span>
            </template>
          </bk-table-column>
          <bk-table-column :label="$t('generic.label.editMode.text')" width="100">
            <template #default="{ row }">
              <span>{{row.editModel === 'form' ? $t('generic.label.editMode.form') : 'YAML'}}</span>
            </template>
          </bk-table-column>
          <bk-table-column :label="$t('generic.label.action')" width="180" fixed="right">
            <template #default="{ row }">
              <bk-button
                text
                v-authority="{
                  clickable: podsWebAnnotations.perms.items[row.uid].detailBtn.clickable,
                  actionId: 'namespace_scoped_view',
                  resourceName: row.namespace,
                  disablePerms: true,
                  permCtx: {
                    project_id: projectID,
                    cluster_id: clusterId,
                    name: row.namespace
                  }
                }"
                @click="handleShowLog(row)">
                {{ $t('generic.button.log1') }}
              </bk-button>
              <bk-button
                class="ml10"
                text
                v-authority="{
                  clickable: podsWebAnnotations.perms.items[row.uid].updateBtn.clickable,
                  actionId: 'namespace_scoped_update',
                  resourceName: row.namespace,
                  disablePerms: true,
                  permCtx: {
                    project_id: projectID,
                    cluster_id: clusterId,
                    name: row.namespace
                  }
                }"
                @click="handleUpdateResource(row)">
                {{ $t('generic.button.update') }}
              </bk-button>
              <bk-button
                class="ml10"
                text
                v-authority="{
                  clickable: podsWebAnnotations.perms.items[row.uid].deleteBtn.clickable,
                  actionId: 'namespace_scoped_delete',
                  resourceName: row.namespace,
                  disablePerms: true,
                  permCtx: {
                    project_id: projectID,
                    cluster_id: clusterId,
                    name: row.namespace
                  }
                }"
                @click="handleDeleteResource(row)">
                {{ $t('generic.button.delete') }}
              </bk-button>
            </template>
          </bk-table-column>
          <template #empty>
            <BcsEmptyTableStatus
              :type="(namespaceValue || searchValue) ? 'search-empty' : 'empty'"
              @clear="handleClearSearchData" />
          </template>
        </bk-table>
      </bcs-tab-panel>
      <bcs-tab-panel name="event" :label="$t('generic.label.event')">
        <EventQueryTable
          class="min-h-[360px]"
          hide-cluster-and-namespace
          kinds="Node"
          :cluster-id="clusterId"
          :name="nodeName" />
      </bcs-tab-panel>
    </bcs-tab>
    <BcsLog
      v-model="showLog"
      :cluster-id="clusterId"
      :namespace="currentRow.namespace"
      :name="currentRow.name">
    </BcsLog>
  </BcsContent>
</template>
<script lang="ts">
import { computed, defineComponent, onMounted, ref, toRefs } from 'vue';

import { fetchNodePodsData } from '@/api/modules/cluster-resource';
import { clusterNodeInfo } from '@/api/modules/monitor';
import $bkMessage from '@/common/bkmagic';
import { formatBytes } from '@/common/util';
import BcsLog from '@/components/bcs-log/log-dialog.vue';
import $bkInfo from '@/components/bk-magic-2.0/bk-info';
import BcsContent from '@/components/layout/Content.vue';
import Row from '@/components/layout/Row.vue';
import Metric from '@/components/metric.vue';
import StatusIcon from '@/components/status-icon';
import { useProject } from '@/composables/use-app';
import usePage from '@/composables/use-page';
import useSearch from '@/composables/use-search';
import useTableSort from '@/composables/use-table-sort';
import $i18n from '@/i18n/i18n-setup';
import $router from '@/router/index';
import $store from '@/store';
import { useSelectItemsNamespace } from '@/views/cluster-manage/namespace/use-namespace';
import EventQueryTable from '@/views/project-manage/event-query/event-query-table.vue';

export default defineComponent({
  name: 'NodeOverview',
  components: { BcsContent, Metric, EventQueryTable, Row, BcsLog, StatusIcon },
  props: {
    nodeName: {
      type: String,
      default: '',
      required: true,
    },
    clusterId: {
      type: String,
      default: '',
      required: true,
    },
  },
  setup(props) {
    const { nodeName, clusterId } = toRefs(props);
    const { projectCode, projectID } = useProject();
    // 节点详情
    const infoList = ref([
      {
        label: 'IP',
        prop: 'ip',
      },
      {
        label: 'CPU',
        prop: 'cpu_count',
      },
      {
        label: $i18n.t('generic.label.mem'),
        prop: 'memory',
        format: formatBytes,
      },
      {
        label: $i18n.t('generic.label.storage'),
        prop: 'disk',
        format: formatBytes,
      },
      {
        label: $i18n.t('cluster.nodeList.detail.label.kernel'),
        prop: 'release',
      },
      {
        label: $i18n.t('cluster.nodeList.detail.label.runtime'),
        prop: 'container_runtime_version',
      },
      {
        label: $i18n.t('cluster.ca.nodePool.label.system'),
        prop: 'sysname',
      },
    ]);
    const nodeInfo = ref({});
    const handleGetNodeInfo = async () => {
      nodeInfo.value = await clusterNodeInfo({
        $projectCode: projectCode.value,
        $clusterId: clusterId.value,
        $nodeIP: nodeName.value,
      }).catch(() => ({}));
    };

    // Pods数据
    const allPodsData = ref<any[]>([]);
    const podsWebAnnotations = ref<any>({});
    const podLoading = ref(false);
    const handleGetPodsData = async () => {
      podLoading.value = true;
      const res = await fetchNodePodsData({
        $projectId: projectID.value,
        $clusterId: clusterId.value,
        $nodename: nodeName.value,
      }, { needRes: true }).catch(() => ({ data: [], webAnnotations: {} }));
      podLoading.value = false;
      allPodsData.value = res.data;
      podsWebAnnotations.value = res.webAnnotations;
    };

    // 命名空间
    const namespaceValue = ref('');
    const { namespaceLoading, namespaceList, getNamespaceData } = useSelectItemsNamespace();

    // 排序
    const { handleSortChange, sortTableData: podsData } = useTableSort(allPodsData);
    // 搜索
    const keys = ref(['name', 'hostIP', 'podIP', 'podIPv4', 'podIPv6']);
    const { searchValue, tableDataMatchSearch } = useSearch(podsData, keys);
    const curSearchTableData = computed(() => tableDataMatchSearch.value
      .filter(item => item.namespace.includes(namespaceValue.value)));
    const {
      pagination,
      curPageData: curPodsData,
      pageChange,
      pageSizeChange,
    } = usePage(curSearchTableData);

    // 跳转Pods详情
    const gotoPodDetail = (row) => {
      $router.push({
        name: 'nodePodDetail',
        params: {
          category: 'pods',
          name: row.name,
          namespace: row.namespace,
          clusterId: clusterId.value,
          nodeId: nodeName.value,
          nodeName: nodeName.value,
          from: 'nodePods',
        },
        query: {
          kind: 'Pod',
        },
      });
    };
    // 更新资源
    const handleUpdateResource = (row) => {
      const { name, namespace, editMode } = row || {};
      if (editMode === 'yaml') {
        $router.push({
          name: 'dashboardResourceUpdate',
          params: {
            namespace,
            name,
            clusterId: clusterId.value,
          },
          query: {
            type: 'workloads',
            category: 'pods',
            kind: 'Pod',
          },
        });
      } else {
        $router.push({
          name: 'dashboardFormResourceUpdate',
          params: {
            namespace,
            name,
            clusterId: clusterId.value,
          },
          query: {
            type: 'workloads',
            category: 'pods',
            kind: 'Pod',
          },
        });
      }
    };
    // 删除资源
    const handleDeleteResource = (row) => {
      const { name, namespace } = row || {};
      $bkInfo({
        type: 'warning',
        clsName: 'custom-info-confirm',
        title: $i18n.t('dashboard.title.confirmDelete'),
        subTitle: `Pod ${name}`,
        defaultInfo: true,
        confirmFn: async () => {
          let result = false;
          result = await $store.dispatch('dashboard/resourceDelete', {
            $namespaceId: namespace,
            $type: 'workloads',
            $category: 'pods',
            $name: name,
            $clusterId: clusterId.value,
          });
          result && $bkMessage({
            theme: 'success',
            message: $i18n.t('generic.msg.success.delete'),
          });
          handleGetPodsData();
        },
      });
    };
    // 显示日志
    const showLog = ref(false);
    const currentRow = ref<Record<string, any>>({});
    const handleShowLog = (row) => {
      currentRow.value = row;
      showLog.value = true;
    };

    // 清空搜索
    const handleClearSearchData = () => {
      namespaceValue.value = '';
      searchValue.value = '';
    };

    onMounted(() => {
      handleGetNodeInfo();
      getNamespaceData({ clusterId: clusterId.value });
      handleGetPodsData();
    });

    return {
      projectID,
      infoList,
      nodeInfo,
      namespaceValue,
      namespaceLoading,
      namespaceList,
      searchValue,
      podLoading,
      curPodsData,
      podsWebAnnotations,
      pagination,
      pageChange,
      pageSizeChange,
      showLog,
      currentRow,
      handleShowLog,
      gotoPodDetail,
      handleUpdateResource,
      handleDeleteResource,
      handleClearSearchData,
      handleSortChange,
    };
  },
});
</script>
