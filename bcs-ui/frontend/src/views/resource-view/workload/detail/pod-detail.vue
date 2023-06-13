<template>
  <div class="workload-detail bcs-content-wrapper">
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
          <bk-button theme="primary" @click="handleShowYamlPanel">{{ $t('查看YAML配置') }}</bk-button>
          <template v-if="!hiddenOperate">
            <bk-button
              theme="primary"
              @click="handleUpdateResource">{{$t('更新')}}</bk-button>
            <bk-button
              theme="danger"
              @click="handleDeleteResource">{{$t('删除')}}</bk-button>
          </template>
        </div>
      </div>
      <div class="workload-main-info">
        <div class="info-item">
          <span class="label">{{ $t('命名空间') }}</span>
          <span class="value" v-bk-overflow-tips>{{ metadata.namespace }}</span>
        </div>
        <div class="info-item">
          <span class="label">{{ $t('镜像') }}</span>
          <span
            class="value"
            v-bk-overflow-tips="getImagesTips(manifestExt.images)">
            {{ manifestExt.images && manifestExt.images.join(', ') }}
          </span>
        </div>
        <div class="info-item">
          <span class="label">{{ $t('节点') }}</span>
          <span class="value" v-bk-overflow-tips>{{ spec.nodeName }}</span>
        </div>
        <div class="info-item">
          <span class="label">UID</span>
          <span class="value" v-bk-overflow-tips>{{ metadata.uid }}</span>
        </div>
        <div class="info-item">
          <span class="label">{{ $t('创建时间') }}</span>
          <span class="value">{{ timeZoneTransForm(manifestExt.createTime) }}</span>
        </div>
        <div class="info-item">
          <span class="label">{{ $t('存在时间') }}</span>
          <span class="value">{{ manifestExt.age }}</span>
        </div>
      </div>
    </div>
    <div class="workload-detail-body">
      <div class="workload-metric">
        <Metric :title="$t('CPU使用率')" metric="cpu_usage" :params="params" category="pods" colors="#30d878"></Metric>
        <Metric
          :title="$t('内存使用量')"
          metric="memory_used"
          :params="params"
          unit="byte"
          category="pods"
          colors="#3a84ff"
          :desc="$t('container_memory_working_set_bytes，limit限制时oom判断依据')">
        </Metric>
        <Metric
          :title="$t('网络')"
          :metric="['network_receive', 'network_transmit']"
          :params="params"
          category="pods"
          unit="byte"
          :colors="['#853cff', '#30d878']"
          :suffix="[$t('入流量'), $t('出流量')]">
        </Metric>
      </div>
      <bcs-tab class="workload-tab" :active.sync="activePanel" type="card" :label-height="42">
        <bcs-tab-panel name="container" :label="$t('容器')" v-bkloading="{ isLoading: containerLoading }">
          <bk-table :data="container">
            <bk-table-column :label="$t('容器名称')" prop="name">
              <template #default="{ row }">
                <bk-button class="bcs-button-ellipsis" text @click="gotoContainerDetail(row)">{{ row.name }}</bk-button>
              </template>
            </bk-table-column>
            <bk-table-column :label="$t('状态')" width="200" prop="status">
              <template #default="{ row }">
                <StatusIcon :status="row.status"></StatusIcon>
              </template>
            </bk-table-column>
            <bk-table-column :label="$t('镜像')" prop="image"></bk-table-column>
            <bk-table-column :label="$t('操作')" width="200" :resizable="false" :show-overflow-tooltip="false">
              <template #default="{ row }">
                <bk-button text @click="handleShowTerminal(row)">WebConsole</bk-button>
                <bk-popover
                  placement="bottom"
                  theme="light dropdown"
                  :arrow="false"
                  trigger="click"
                  v-if="row.containerID && !isSharedCluster">
                  <bk-button style="cursor: default;" text class="ml10">{{ $t('日志检索') }}</bk-button>
                  <div slot="content">
                    <!-- 内部版 -->
                    <ul v-if="$INTERNAL">
                      <a
                        :href="logLinks[row.containerID] && logLinks[row.containerID].std_log_link"
                        target="_blank" class="dropdown-item">
                        {{ $t('标准输出检索') }}
                      </a>
                      <a
                        :href="logLinks[row.containerID] && logLinks[row.containerID].file_log_link"
                        target="_blank" class="dropdown-item">
                        {{ $t('文件日志检索') }}
                      </a>
                    </ul>
                    <ul v-else>
                      <a
                        :href="logLinks[row.containerID] && logLinks[row.containerID].std_log_link"
                        target="_blank" class="dropdown-item">
                        {{ $t('标准日志') }}
                      </a>
                      <a
                        :href="logLinks[row.containerID] && logLinks[row.containerID].file_log_link"
                        target="_blank" class="dropdown-item">
                        {{ $t('文件路径日志') }}
                      </a>
                    </ul>
                  </div>
                </bk-popover>
              </template>
            </bk-table-column>
          </bk-table>
        </bcs-tab-panel>
        <bcs-tab-panel name="event" :label="$t('事件')">
          <EventQueryTableVue
            class="min-h-[360px]"
            hide-cluster-and-namespace
            :kinds="['Pod']"
            :cluster-id="clusterId"
            :namespace="namespace"
            :name="name">
          </EventQueryTableVue>
        </bcs-tab-panel>
        <bcs-tab-panel name="conditions" :label="$t('状态（Conditions）')">
          <bk-table :data="conditions">
            <bk-table-column :label="$t('类别')" prop="type"></bk-table-column>
            <bk-table-column :label="$t('状态')" prop="status">
              <template #default="{ row }">
                <StatusIcon :status="row.status"></StatusIcon>
              </template>
            </bk-table-column>
            <bk-table-column :label="$t('最后迁移时间')" prop="lastTransitionTime">
              <template #default="{ row }">
                {{ formatTime(row.lastTransitionTime, 'yyyy-MM-dd hh:mm:ss') }}
              </template>
            </bk-table-column>
            <bk-table-column :label="$t('原因')">
              <template #default="{ row }">
                {{ row.reason || '--' }}
              </template>
            </bk-table-column>
            <bk-table-column :label="$t('消息')">
              <template #default="{ row }">
                {{ row.message || '--' }}
              </template>
            </bk-table-column>
          </bk-table>
        </bcs-tab-panel>
        <bcs-tab-panel name="storage" :label="$t('存储')" v-bkloading="{ isLoading: storageLoading }">
          <div class="storage storage-pvcs">
            <div class="title">PersistentVolumeClaims</div>
            <bk-table :data="storageTableData.pvcs">
              <bk-table-column :label="$t('名称')" prop="metadata.name" sortable :resizable="false"></bk-table-column>
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
              <bk-table-column label="Access Modes">
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
              <bk-table-column :label="$t('名称')" prop="metadata.name" sortable :resizable="false"></bk-table-column>
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
              <bk-table-column :label="$t('名称')" prop="metadata.name" sortable :resizable="false"></bk-table-column>
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
        <bcs-tab-panel name="label" :label="$t('标签')">
          <bk-table :data="labels">
            <bk-table-column label="Key" prop="key"></bk-table-column>
            <bk-table-column label="Value" prop="value"></bk-table-column>
          </bk-table>
        </bcs-tab-panel>
        <bcs-tab-panel name="annotations" :label="$t('注解')">
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
import { computed, defineComponent, onMounted, ref, toRefs, toRef, reactive } from 'vue';
import { bkOverflowTips } from 'bk-magic-vue';
import StatusIcon from '@/components/status-icon';
import Metric from '@/components/metric.vue';
import useDetail from './use-detail';
import { formatTime, timeZoneTransForm } from '@/common/util';
import CodeEditor from '@/components/monaco-editor/new-editor.vue';
import fullScreen from '@/directives/full-screen';
import EventQueryTableVue from '@/views/project-manage/event-query/event-query-table.vue';
import $store from '@/store';
import $router from '@/router';
import { useConfig } from '@/composables/use-app';

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
    EventQueryTableVue,
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
    const $route = computed(() => toRef(reactive($router), 'currentRoute').value);
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
    const curProject = computed(() => $store.state.curProject);
    const handleGetContainer = async () => {
      containerLoading.value = true;
      container.value = await $store.dispatch('dashboard/listContainers', {
        $podId: name.value,
        $namespaceId: namespace.value,
        $clusterId: clusterId.value,
      });
      const containerIDs = container.value.map(item => item.containerID).filter(id => !!id);
      if (containerIDs.length) {
        logLinks.value = _INTERNAL_.value
          ? await $store.dispatch('dashboard/logLinks', {
            container_ids: containerIDs.join(','),
          })
          : await $store.dispatch('crdcontroller/getLogLinks', {
            container_ids: containerIDs.join(','),
            bk_biz_id: curProject.value?.businessID,
          });
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
    const projectId = computed(() => $route.value.params.projectId);
    const terminalWins = new Map();
    const handleShowTerminal = (row) => {
      const url = `${window.DEVOPS_BCS_API_URL}/web_console/projects/${projectId.value}/clusters/${clusterId.value}/?namespace=${props.namespace}&pod_name=${props.name}&container_name=${row.name}`;
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

    const isSharedCluster = computed(() => ($store.state.cluster.clusterList as any[])
      .find(item => item.clusterID === clusterId.value)?.is_shared);

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
      isSharedCluster,
      timeZoneTransForm,
      handleShowYamlPanel,
      handleGetStorage,
      gotoContainerDetail,
      handleGetExtData,
      formatTime,
      getImagesTips,
      handleUpdateResource,
      handleDeleteResource,
      handleShowTerminal,
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
