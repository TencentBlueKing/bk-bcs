<template>
  <div class="workload-detail">
    <div class="workload-detail-info" v-bkloading="{ isLoading }">
      <div class="workload-info-basic">
        <div class="basic-left">
          <span class="name mr20">{{ metadata.name }}</span>
          <div class="basic-wrapper">
            <StatusIcon class="basic-item" :status="manifestExt.status"></StatusIcon>
            <div class="basic-item">
              <span class="label">Ready</span>
              <span class="value">{{ manifestExt.readyCnt }} / {{ manifestExt.totalCnt }}</span>
            </div>
            <div class="basic-item">
              <span class="label">Host IP</span>
              <span class="value">{{ status.hostIP || '--' }}</span>
            </div>
            <div class="basic-item">
              <span class="label">Pod IPv4</span>
              <span class="value">{{ manifestExt.podIPv4 || '--' }}</span>
            </div>
            <div class="basic-item">
              <span class="label">Pod IPv6</span>
              <span class="value">{{ manifestExt.podIPv6 || '--' }}</span>
            </div>
          </div>
        </div>
        <div class="btns">
          <bk-button theme="primary" @click="handleShowYamlPanel">{{ $t('dashboard.workload.button.yaml') }}</bk-button>
          <template v-if="!hiddenOperate">
            <bk-button
              theme="primary"
              @click="handleUpdateResource">{{$t('generic.button.update')}}</bk-button>
            <bk-button
              theme="danger"
              @click="handleDeleteResource">{{$t('generic.button.delete')}}</bk-button>
          </template>
        </div>
      </div>
      <div class="workload-main-info">
        <div class="info-item">
          <span class="label">{{ $t('cluster.labels.name') }}</span>
          <span class="value">{{ clusterNameMap[clusterId] }}</span>
        </div>
        <div class="info-item">
          <span class="label">{{ $t('k8s.namespace') }}</span>
          <span class="value" v-bk-overflow-tips>{{ metadata.namespace }}</span>
        </div>
        <div class="info-item">
          <span class="label">{{ $t('k8s.image') }}</span>
          <span
            class="value"
            v-bk-overflow-tips="getImagesTips(manifestExt.images)">
            {{ manifestExt.images && manifestExt.images.join(', ') }}
          </span>
        </div>
        <div class="info-item">
          <span class="label">{{ $t('dashboard.workload.pods.node') }}</span>
          <span class="value" v-bk-overflow-tips>{{ spec.nodeName }}</span>
        </div>
        <div class="info-item">
          <span class="label">UID</span>
          <span class="value" v-bk-overflow-tips>{{ metadata.uid }}</span>
        </div>
        <div class="info-item">
          <span class="label">{{ $t('cluster.labels.createdAt') }}</span>
          <span class="value">{{ timeFormat(manifestExt.createTime) }}</span>
        </div>
        <div class="info-item">
          <span class="label">{{ $t('k8s.age') }}</span>
          <span class="value">{{ manifestExt.age }}</span>
        </div>
      </div>
    </div>
    <div class="workload-detail-body">
      <div class="workload-metric">
        <Metric
          :title="$t('metrics.cpuUsage')"
          metric="cpu_usage"
          :params="params"
          category="pods"
          colors="#30d878"
          unit="percent-number">
        </Metric>
        <Metric
          :title="$t('metrics.memUsage1')"
          metric="memory_used"
          :params="params"
          unit="byte"
          category="pods"
          colors="#3a84ff"
          :desc="$t('dashboard.workload.tips.containerMemoryWorkingSetBytesOom')">
        </Metric>
        <Metric
          :title="$t('k8s.networking')"
          :metric="['network_receive', 'network_transmit']"
          :params="params"
          category="pods"
          unit="byte"
          :colors="['#853cff', '#30d878']"
          :suffix="[$t('metrics.network.receive'), $t('metrics.network.transmit')]">
        </Metric>
      </div>
      <bcs-tab class="workload-tab" :active.sync="activePanel" type="card" :label-height="42">
        <bcs-tab-panel
          name="container"
          :label="$t('dashboard.workload.container.title')"
          v-bkloading="{ isLoading: containerLoading }"
          render-directive="if">
          <bk-table :data="container">
            <bk-table-column :label="$t('dashboard.workload.container.name')" prop="name">
              <template #default="{ row }">
                <bk-button class="bcs-button-ellipsis" text @click="gotoContainerDetail(row)">{{ row.name }}</bk-button>
              </template>
            </bk-table-column>
            <bk-table-column :label="$t('generic.label.status')" width="200" prop="status">
              <template #default="{ row }">
                <StatusIcon :status="row.status"></StatusIcon>
              </template>
            </bk-table-column>
            <bk-table-column :label="$t('container.label.startedAt')" prop="startedAt">
              <template #default="{ row }">
                {{ timeFormat(row.startedAt) }}
              </template>
            </bk-table-column>
            <bk-table-column :label="$t('container.label.restartCnt')" prop="restartCnt"></bk-table-column>
            <bk-table-column :label="$t('k8s.image')" prop="image"></bk-table-column>
            <bk-table-column
              :label="$t('generic.label.action')"
              width="200"
              :resizable="false"
              :show-overflow-tooltip="false">
              <template #default="{ row }">
                <bk-button text @click="handleShowTerminal(row)">WebConsole</bk-button>
                <bk-popover
                  placement="bottom"
                  theme="light dropdown"
                  :arrow="false"
                  trigger="click"
                  v-if="row.containerID">
                  <bk-button
                    style="cursor: default;"
                    text
                    class="ml10">{{ $t('dashboard.workload.pods.log') }}</bk-button>
                  <div slot="content">
                    <ul>
                      <a
                        :href="logLinks[row.containerID] && logLinks[row.containerID].std_log_url"
                        target="_blank" class="dropdown-item">
                        {{ $t('dashboard.workload.pods.stdoutLog') }}
                      </a>
                      <a
                        :href="logLinks[row.containerID] && logLinks[row.containerID].file_log_url"
                        target="_blank" class="dropdown-item">
                        {{ $t('dashboard.workload.pods.filelog') }}
                      </a>
                    </ul>
                  </div>
                </bk-popover>
              </template>
            </bk-table-column>
          </bk-table>
        </bcs-tab-panel>
        <bcs-tab-panel name="event" :label="$t('generic.label.event')" render-directive="if">
          <EventQueryTable
            class="min-h-[360px]"
            hide-cluster-and-namespace
            :kinds="['Pod']"
            :cluster-id="clusterId"
            :namespace="namespace"
            :name="name">
          </EventQueryTable>
        </bcs-tab-panel>
        <bcs-tab-panel name="conditions" :label="$t('k8s.conditions')" render-directive="if">
          <bk-table :data="conditions">
            <bk-table-column :label="$t('generic.label.type')" prop="type"></bk-table-column>
            <bk-table-column :label="$t('generic.label.status')" prop="status">
              <template #default="{ row }">
                <StatusIcon :status="row.status"></StatusIcon>
              </template>
            </bk-table-column>
            <bk-table-column :label="$t('k8s.lastTransitionTime')" prop="lastTransitionTime">
              <template #default="{ row }">
                {{ timeFormat(row.lastTransitionTime) }}
              </template>
            </bk-table-column>
            <bk-table-column :label="$t('dashboard.workload.label.reason')">
              <template #default="{ row }">
                {{ row.reason || '--' }}
              </template>
            </bk-table-column>
            <bk-table-column :label="$t('generic.label.message')">
              <template #default="{ row }">
                {{ row.message || '--' }}
              </template>
            </bk-table-column>
          </bk-table>
        </bcs-tab-panel>
        <bcs-tab-panel
          name="storage"
          :label="$t('generic.label.storage')"
          v-bkloading="{ isLoading: storageLoading }"
          render-directive="if">
          <div class="storage storage-pvcs">
            <div class="title">PersistentVolumeClaims</div>
            <bk-table :data="storageTableData.pvcs">
              <bk-table-column
                :label="$t('generic.label.name')"
                prop="metadata.name"
                sortable
                :resizable="false">
              </bk-table-column>
              <bk-table-column label="Status">
                <template #default="{ row }">
                  <span>{{ row.status.phase || '--' }}</span>
                </template>
              </bk-table-column>
              <bk-table-column label="Volume">
                <template #default="{ row }">
                  <span>{{ row.spec.volumeName || '--' }}</span>
                </template>
              </bk-table-column>
              <bk-table-column label="Capacity">
                <template #default="{ row }">
                  <span>{{ row.status.capacity ? row.status.capacity.storage : '--' }}</span>
                </template>
              </bk-table-column>
              <bk-table-column label="Access modes">
                <template #default="{ row }">
                  <span>{{ handleGetExtData(row.metadata.uid, 'pvcs','accessModes').join(', ') }}</span>
                </template>
              </bk-table-column>
              <bk-table-column label="StorageClass">
                <template #default="{ row }">
                  <span>{{ row.spec.storageClassName || '--' }}</span>
                </template>
              </bk-table-column>
              <bk-table-column label="VolumeMode">
                <template #default="{ row }">
                  <span>{{ row.spec.volumeMode || '--' }}</span>
                </template>
              </bk-table-column>
              <bk-table-column label="Age" :resizable="false" :show-overflow-tooltip="false">
                <template #default="{ row }">
                  <span v-bk-tooltips="{ content: handleGetExtData(row.metadata.uid, 'pvcs','createTime') }">
                    {{ handleGetExtData(row.metadata.uid, 'pvcs','age') }}
                  </span>
                </template>
              </bk-table-column>
            </bk-table>
          </div>
          <div class="storage storage-config">
            <div class="title">ConfigMaps</div>
            <bk-table :data="storageTableData.configmaps">
              <bk-table-column
                :label="$t('generic.label.name')"
                prop="metadata.name"
                sortable
                :resizable="false">
              </bk-table-column>
              <bk-table-column label="Data">
                <template #default="{ row }">
                  <span>{{ handleGetExtData(row.metadata.uid, 'configmaps','data').join(', ') || '--' }}</span>
                </template>
              </bk-table-column>
              <bk-table-column label="Age" :resizable="false" :show-overflow-tooltip="false">
                <template #default="{ row }">
                  <span v-bk-tooltips="{ content: handleGetExtData(row.metadata.uid, 'configmaps','createTime') }">
                    {{ handleGetExtData(row.metadata.uid, 'configmaps','age') }}
                  </span>
                </template>
              </bk-table-column>
            </bk-table>
          </div>
          <div class="storage storage-secrets">
            <div class="title">Secrets</div>
            <bk-table :data="storageTableData.secrets">
              <bk-table-column
                :label="$t('generic.label.name')"
                prop="metadata.name"
                sortable
                :resizable="false">
              </bk-table-column>
              <bk-table-column label="Type">
                <template #default="{ row }">
                  <span>{{ row.type || '--' }}</span>
                </template>
              </bk-table-column>
              <bk-table-column label="Data">
                <template #default="{ row }">
                  <span>{{ handleGetExtData(row.metadata.uid, 'secrets','data').join(', ') || '--' }}</span>
                </template>
              </bk-table-column>
              <bk-table-column label="Age" :resizable="false" :show-overflow-tooltip="false">
                <template #default="{ row }">
                  <span v-bk-tooltips="{ content: handleGetExtData(row.metadata.uid, 'secrets','createTime') }">
                    {{ handleGetExtData(row.metadata.uid, 'secrets','age') }}
                  </span>
                </template>
              </bk-table-column>
            </bk-table>
          </div>
        </bcs-tab-panel>
        <bcs-tab-panel name="label" :label="$t('k8s.label')" render-directive="if">
          <bk-table :data="labels">
            <bk-table-column label="Key" prop="key"></bk-table-column>
            <bk-table-column label="Value" prop="value"></bk-table-column>
          </bk-table>
        </bcs-tab-panel>
        <bcs-tab-panel name="annotations" :label="$t('k8s.annotation')" render-directive="if">
          <bk-table :data="annotations">
            <bk-table-column label="Key" prop="key"></bk-table-column>
            <bk-table-column label="Value" prop="value"></bk-table-column>
          </bk-table>
        </bcs-tab-panel>
      </bcs-tab>
    </div>
    <bcs-sideslider quick-close :title="metadata.name" :is-show.sync="showYamlPanel" :width="800">
      <template #content>
        <CodeEditor
          v-full-screen="{ tools: ['fullscreen', 'copy'], content: yaml }"
          width="100%"
          height="100%"
          readonly
          :options="{
            roundedSelection: false,
            scrollBeyondLastLine: false,
            renderLineHighlight: false,
          }"
          :value="yaml">
        </CodeEditor>
      </template>
    </bcs-sideslider>
  </div>
</template>
<script lang="ts">
/* eslint-disable camelcase */
import { bkOverflowTips } from 'bk-magic-vue';
import { computed, defineComponent, onMounted, ref, toRefs } from 'vue';

import useDetail from './use-detail';

import { logCollectorEntrypoints } from '@/api/modules/monitor';
import { timeFormat } from '@/common/util';
import Metric from '@/components/metric.vue';
import CodeEditor from '@/components/monaco-editor/new-editor.vue';
import StatusIcon from '@/components/status-icon';
import { useCluster, useConfig, useProject } from '@/composables/use-app';
import fullScreen from '@/directives/full-screen';
import $store from '@/store';
import EventQueryTable from '@/views/project-manage/event-query/event-query-table.vue';

export interface IDetail {
  manifest: any;
  manifestExt: any;
}

export interface IStorage {
  pvcs: IDetail | null;
  configmaps: IDetail | null;
  secrets: IDetail | null;
}

export default defineComponent({
  name: 'PodDetail',
  components: {
    StatusIcon,
    Metric,
    CodeEditor,
    EventQueryTable,
  },
  directives: {
    bkOverflowTips,
    'full-screen': fullScreen,
  },
  props: {
    namespace: {
      type: String,
      default: '',
      required: true,
    },
    // pod 名称
    name: {
      type: String,
      default: '',
      required: true,
    },
    clusterId: {
      type: String,
      default: '',
      required: true,
    },
    // 是否隐藏 更新 和 删除操作（兼容集群管理应用详情）
    hiddenOperate: {
      type: Boolean,
      default: false,
    },
  },
  setup(props, ctx) {
    const { clusterNameMap } = useCluster();
    const {
      isLoading,
      detail,
      activePanel,
      labels,
      annotations,
      metadata,
      manifestExt,
      yaml,
      showYamlPanel,
      handleGetDetail,
      handleShowYamlPanel,
      handleUpdateResource,
      handleDeleteResource,
    } = useDetail({
      ...props,
      category: 'pods',
      defaultActivePanel: 'container',
      type: 'workloads',
    });
    const { name, namespace, clusterId } = toRefs(props);
    const params = computed(() => ({
      $namespaceId: namespace.value,
      $clusterId: clusterId.value,
      pod_name_list: [name.value],
    }));

    // 容器
    const { _INTERNAL_ } = useConfig();
    const container = ref<any[]>([]);
    const containerLoading = ref(false);
    const logLinks = ref({});
    const handleGetContainer = async () => {
      containerLoading.value = true;
      container.value = await $store.dispatch('dashboard/listContainers', {
        $podId: name.value,
        $namespaceId: namespace.value,
        $clusterId: clusterId.value,
      });
      const containerIDs = container.value.map(item => item.containerID).filter(id => !!id);
      if (containerIDs.length) {
        logLinks.value = await logCollectorEntrypoints({
          container_ids: containerIDs,
          $clusterId: clusterId.value,
        }).catch(() => ({}));
      }
      containerLoading.value = false;
    };
    // 状态
    const conditions = computed(() => detail.value?.manifest.status?.conditions || []);
    // status 数据
    const status = computed(() => detail.value?.manifest?.status || {});
    // spec 数据
    const spec = computed(() => detail.value?.manifest?.spec || {});

    // 存储
    const storage = ref<IStorage>({
      pvcs: null,
      configmaps: null,
      secrets: null,
    });
    const storageTableData = computed(() => ({
      pvcs: storage.value.pvcs?.manifest.items || [],
      configmaps: storage.value.configmaps?.manifest.items || [],
      secrets: storage.value.secrets?.manifest.items || [],
    }));
    // 获取存储数据
    const storageLoading = ref(false);
    const handleGetStorage = async () => {
      storageLoading.value = true;
      const types = ['pvcs', 'configmaps', 'secrets'];
      const promises = types.map(type => $store.dispatch('dashboard/listStoragePods', {
        $podId: name.value,
        $type: type,
        $namespaceId: namespace.value,
        $clusterId: clusterId.value,
      }));
      const [pvcs = {}, configmaps = {}, secrets = {}] = await Promise.all(promises);
      storage.value = {
        pvcs,
        configmaps,
        secrets,
      };
      storageLoading.value = false;
    };
    // 获取存储manifestExt的字段
    const handleGetExtData = (uid, type, prop) => storage.value[type]?.manifestExt?.[uid]?.[prop] || '';

    // 跳转容器详情
    const gotoContainerDetail = (row) => {
      ctx.emit('container-detail', row);
    };

    // 获取镜像tips
    const getImagesTips = (images) => {
      if (!images) {
        return {
          content: '',
        };
      }
      return {
        allowHTML: true,
        maxWidth: 480,
        content: images.join('<br />'),
      };
    };

    // 容器操作
    // 1. 跳转WebConsole
    const { projectCode } = useProject();
    const terminalWins = new Map();
    const handleShowTerminal = (row) => {
      const url = `${window.BCS_API_HOST}/bcsapi/v4/webconsole/projects/${projectCode.value}/clusters/${clusterId.value}/?namespace=${props.namespace}&pod_name=${props.name}&container_name=${row.name}`;
      if (terminalWins.has(row.containerID)) {
        const win = terminalWins.get(row.containerID);
        if (!win.closed) {
          terminalWins.get(row.containerID).focus();
        } else {
          const win = window.open(url, '_blank');
          terminalWins.set(row.containerID, win);
        }
      } else {
        const win = window.open(url, '_blank');
        terminalWins.set(row.containerID, win);
      }
    };
    // 2. 日志检索
    const isDropdownShow = ref(false);

    onMounted(async () => {
      handleGetDetail();
      handleGetStorage();
      handleGetContainer();
    });

    return {
      params,
      container,
      conditions,
      storage,
      storageTableData,
      isLoading,
      detail,
      metadata,
      manifestExt,
      spec,
      status,
      activePanel,
      labels,
      annotations,
      storageLoading,
      containerLoading,
      yaml,
      showYamlPanel,
      isDropdownShow,
      logLinks,
      handleShowYamlPanel,
      handleGetStorage,
      gotoContainerDetail,
      handleGetExtData,
      timeFormat,
      getImagesTips,
      handleUpdateResource,
      handleDeleteResource,
      handleShowTerminal,
      clusterNameMap,
    };
  },
});
</script>
<style lang="postcss" scoped>
@import './detail-info.css';
.workload-detail {
    width: 100%;
    /deep/ .bk-sideslider .bk-sideslider-content {
        height: 100%;
    }
    &-info {
        @mixin detail-info 3;
    }
    &-body {
        background: #FAFBFD;
        padding: 0 24px;
        .workload-metric {
            display: flex;
            background: #fff;
            margin-top: 16px;
            height: 230px;
        }
        .workload-tab {
            margin-top: 16px;
        }
        .storage {
            margin-bottom: 24px;
            .title {
                font-size: 14px;
                color: #313238;
                margin-bottom: 8px;
            }
        }
    }
}
>>> .dropdown-item {
    display: block;
    height: 32px;
    line-height: 33px;
    padding: 0 16px;
    color: #63656e;
    font-size: 12px;
    text-decoration: none;
    white-space: nowrap;
    cursor: pointer;
    &:hover {
        background-color: #eaf3ff;
        color: #3a84ff;
    }
}
</style>
