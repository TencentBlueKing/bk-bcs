<template>
  <div class="workload-detail">
    <div class="workload-detail-info" v-bkloading="{ isLoading }">
      <div class="workload-info-basic">
        <div class="basic-left">
          <span class="name mr20">{{ detail && detail.containerName }}</span>
          <div class="basic-wrapper">
            <div class="basic-item">
              <span class="label">{{ $t('dashboard.workload.container.hostIP') }}</span>
              <span class="value">{{ detail && detail.hostIP || '--' }}</span>
            </div>
            <div class="basic-item">
              <span class="label">{{ $t('dashboard.workload.container.containerIP') }}</span>
              <span class="value">{{ detail && detail.containerIP || '--' }}</span>
            </div>
          </div>
        </div>
      </div>
      <div class="workload-main-info">
        <div class="info-item">
          <span class="label">{{ $t('cluster.labels.hostName') }}</span>
          <span class="value" v-bk-overflow-tips>{{ detail && detail.hostName }}</span>
        </div>
        <div class="info-item">
          <span class="label">{{ $t('dashboard.workload.container.containerID') }}</span>
          <span class="value" v-bk-overflow-tips>{{ detail && detail.containerID }}</span>
        </div>
        <div class="info-item">
          <span class="label">{{ $t('k8s.image') }}</span>
          <span class="value" v-bk-overflow-tips>{{ detail && detail.image }}</span>
        </div>
        <div class="info-item">
          <span class="label">{{ $t('cluster.create.label.networkMode.text') }}</span>
          <span class="value">{{ detail && detail.networkMode }}</span>
        </div>
      </div>
      <div class="workload-main-info">
        <div class="info-item">
          <span class="label">{{ $t('container.label.startedAt') }}</span>
          <span class="value" v-bk-overflow-tips>{{ detail ? timeFormat(detail.startedAt) : '--' }}</span>
        </div>
        <div class="info-item">
          <span class="label">{{ $t('container.label.restartCnt') }}</span>
          <span class="value" v-bk-overflow-tips>{{ detail ? detail.restartCnt : '--' }}</span>
        </div>
        <div class="info-item">
          <span class="label">{{ $t('container.label.lastState') }}</span>
          <span class="value">
            <bk-popover :disabled="!lastState.status">
              <span :class="{ 'bcs-border-tips inline-flex': lastState.status }">{{ lastState.status || '--' }}</span>
              <template #content>
                <div>
                  <span>{{ $t('container.label.reason') }}:</span>
                  {{ lastState.reason || '--' }}
                </div>
                <div>
                  <span>{{ $t('container.label.exitCode') }}:</span>
                  {{ lastState.exitCode || '--' }}
                </div>
                <div>
                  <span>{{ $t('container.label.startedAt') }}:</span>
                  {{ lastState.startedAt || '--' }}
                </div>
                <div>
                  <span>{{ $t('container.label.finishedAt') }}:</span>
                  {{ lastState.finishedAt || '--' }}
                </div>
              </template>
            </bk-popover>
          </span>
        </div>
      </div>
    </div>
    <div class="workload-detail-body">
      <div class="workload-metric">
        <Metric
          :title="$t('metrics.cpuUsage')"
          :metric="['cpu_usage', 'cpu_limit']"
          :params="params"
          category="containers"
          :colors="['#30d878', '#ff5656']"
          :series="[{ }, { areaStyle: null }]">
        </Metric>
        <Metric
          :title="$t('metrics.memUsage1')"
          :metric="['memory_used', 'memory_limit']"
          :params="params"
          unit="byte"
          category="containers"
          :colors="['#3a84ff', '#ff5656']"
          :desc="$t('dashboard.workload.tips.containerMemoryWorkingSetBytesOom')"
          :series="[{ }, { areaStyle: null }]">
        </Metric>
        <Metric
          :title="$t('metrics.diskIOUsage1')"
          :metric="['disk_read_total', 'disk_write_total']"
          :params="params"
          category="containers"
          unit="byte"
          :colors="['#853cff', '#30d878']"
          :suffix="[$t('metrics.pod.disk.read'), $t('metrics.pod.disk.write')]">
        </Metric>
      </div>
      <bcs-tab class="workload-tab" :active.sync="activePanel" type="card" :label-height="42">
        <bcs-tab-panel name="ports" :label="$t('dashboard.network.portmapping')">
          <bk-table :data="ports">
            <bk-table-column label="Name" prop="name">
              <template #default="{ row }">
                {{ row.name || '--' }}
              </template>
            </bk-table-column>
            <bk-table-column label="Host Port" prop="hostPort">
              <template #default="{ row }">
                {{ row.hostPort || '--' }}
              </template>
            </bk-table-column>
            <bk-table-column label="Container Port" prop="containerPort"></bk-table-column>
            <bk-table-column label="Protocol" prop="protocol"></bk-table-column>
          </bk-table>
        </bcs-tab-panel>
        <bcs-tab-panel name="command" :label="$t('dashboard.workload.container.command')">
          <bk-table :data="command">
            <bk-table-column label="Command" prop="command">
              <template #default="{ row }">
                {{ row.command || '--' }}
              </template>
            </bk-table-column>
            <bk-table-column label="Args" prop="args">
              <template #default="{ row }">
                {{ row.args || '--' }}
              </template>
            </bk-table-column>
          </bk-table>
        </bcs-tab-panel>
        <bcs-tab-panel name="volumes" :label="$t('k8s.volume')">
          <bk-table :data="volumes">
            <bk-table-column label="Host Path" prop="name"></bk-table-column>
            <bk-table-column label="Mount Path" prop="mountPath"></bk-table-column>
            <bk-table-column label="ReadOnly" prop="readonly">
              <template #default="{ row }">
                {{ row.readonly }}
              </template>
            </bk-table-column>
          </bk-table>
        </bcs-tab-panel>
        <bcs-tab-panel name="env" :label="$t('dashboard.workload.container.env')">
          <bk-table :data="envs" v-bkloading="{ isLoading: envsTableLoading }">
            <bk-table-column label="Key" prop="name"></bk-table-column>
            <bk-table-column label="Value" prop="value"></bk-table-column>
          </bk-table>
        </bcs-tab-panel>
        <bcs-tab-panel name="label" :label="$t('k8s.label')">
          <bk-table :data="labels">
            <bk-table-column label="Key" prop="key"></bk-table-column>
            <bk-table-column label="Value" prop="val"></bk-table-column>
          </bk-table>
        </bcs-tab-panel>
        <bcs-tab-panel name="resources" :label="$t('dashboard.workload.container.limits')">
          <bk-table :data="resources">
            <bk-table-column label="Cpu">
              <template #default="{ row }">
                {{ `requests: ${
                  row.requests ? row.requests.cpu : '--'} | limits: ${row.limits ? row.limits.cpu : '--'}` }}
              </template>
            </bk-table-column>
            <bk-table-column label="Memory">
              <template #default="{ row }">
                {{ `requests: ${
                  row.requests ? row.requests.memory : '--'} | limits: ${row.limits ? row.limits.memory : '--'}` }}
              </template>
            </bk-table-column>
          </bk-table>
        </bcs-tab-panel>
      </bcs-tab>
    </div>
  </div>
</template>
<script lang="ts">
/* eslint-disable camelcase */
import { bkOverflowTips } from 'bk-magic-vue';
import { computed, defineComponent, onMounted, ref, toRefs } from 'vue';

import { timeFormat } from '@/common/util';
import Metric from '@/components/metric.vue';
import $store from '@/store';

export default defineComponent({
  name: 'ContainerDetail',
  components: {
    Metric,
  },
  directives: {
    bkOverflowTips,
  },
  props: {
    namespace: {
      type: String,
      default: '',
      required: true,
    },
    // pod名
    pod: {
      type: String,
      default: '',
      required: true,
    },
    // 容器名
    name: {
      type: String,
      default: '',
      required: true,
    },
    // 容器ID
    // id: {
    //   type: String,
    //   default: '',
    //   required: true,
    // },
    clusterId: {
      type: String,
      default: '',
      required: true,
    },
  },
  setup(props) {
    const { name, namespace, pod, clusterId } = toRefs(props);

    // 详情loading
    const isLoading = ref(false);
    // 环境变量表格loading
    const envsTableLoading = ref(false);
    // 详情数据
    const detail = ref<Record<string, any>|null>(null);
    const activePanel = ref('ports');

    // 图表指标参数
    const params = computed(() => ({
      $namespaceId: namespace.value,
      $containerId: name.value,
      $podId: pod.value,
      $clusterId: clusterId.value,
    }));

    // 最后一次重启状态
    const lastState = computed(() => {
      const lastStatus = Object.keys(detail.value?.lastState || {})?.[0] || '';
      const state = detail.value?.lastState?.[lastStatus] || {};
      return {
        ...state,
        status: lastStatus,
      };
    });
    // 端口映射
    const ports = computed(() => detail.value?.ports || []);
    // 命令
    const command = computed(() => [detail.value?.command || {}]);
    // 挂载卷
    const volumes = computed(() => detail.value?.volumes || []);
    // 标签数据
    const labels = computed(() => detail.value?.labels || []);
    // 资源限额
    const resources = computed(() => [detail.value?.resources || { requests: {}, limits: {} }]);
    // 环境变量
    const envs = ref([]);

    // 容器详情
    const handleGetDetail = async () => {
      isLoading.value = true;
      detail.value = await $store.dispatch('dashboard/retrieveContainerDetail', {
        $namespaceId: namespace.value,
        $category: 'pods',
        $name: pod.value,
        $containerName: name.value,
        $clusterId: clusterId.value,
      });
      isLoading.value = false;
      return detail.value;
    };

    // 容器环境变量
    const handleGetContainerEnv = async () => {
      envsTableLoading.value = true;
      envs.value = await $store.dispatch('dashboard/fetchContainerEnvInfo', {
        $namespaceId: namespace.value,
        $podId: pod.value,
        $containerName: name.value,
        $clusterId: clusterId.value,
      });
      envsTableLoading.value = false;
      return envs.value;
    };

    onMounted(() => {
      handleGetDetail();
      handleGetContainerEnv();
    });

    return {
      envsTableLoading,
      params,
      isLoading,
      detail,
      activePanel,
      ports,
      command,
      resources,
      volumes,
      labels,
      envs,
      lastState,
      timeFormat,
      handleGetDetail,
      handleGetContainerEnv,
    };
  },
});
</script>
<style lang="postcss" scoped>
@import './container-detail.css';
</style>
