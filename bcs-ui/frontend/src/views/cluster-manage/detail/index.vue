<template>
  <div class="modal" ref="modalRef">
    <span class="fold">
      <span
        :class="['fold-left', { active: panelStatus === 'hidden' }]"
        @click="toggleDetailPanel(true)">
        <i class="bk-icon icon-angle-left"></i>
      </span>
      <span
        :class="['fold-right', { active: panelStatus === 'expanded' }]"
        @click="toggleDetailPanel(false)">
        <i class="bk-icon icon-angle-right"></i>
      </span>
    </span>
    <div class="absolute left-[-24px] h-full w-[24px] bg-[#f5f7fa]"></div>
    <div class="resize-line" ref="resizeRef" @mousedown="onMousedownEvent"></div>
    <div class="h-[56px] flex items-center px-[24px] pt-[16px] pb-[8px] text-[12px]">
      <span class="font-bold text-[14px] leading-[22px]">{{ curCluster.clusterName || '--' }}</span>
      <span class="ml-[14px] text-[#979BA5] leading-[22px] select-all">{{ curCluster.clusterID }}</span>
      <bcs-divider direction="vertical"></bcs-divider>
      <div class="h-[22px] inline-flex items-center">
        <StatusIcon
          :status-color-map="{
            'CREATE-FAILURE': 'red',
            'DELETE-FAILURE': 'red',
            'IMPORT-FAILURE': 'red',
            'CONNECT-FAILURE': 'red',
            RUNNING: 'green'
          }"
          :status-text-map="{
            INITIALIZATION: $t('generic.status.initializing'),
            DELETING: $t('generic.status.deleting'),
            'CREATE-FAILURE': $t('generic.status.createFailed'),
            'DELETE-FAILURE': $t('generic.status.deleteFailed'),
            'IMPORT-FAILURE': $t('cluster.status.importFailed'),
            'CONNECT-FAILURE': $t('cluster.status.connectFailed'),
            RUNNING: $t('generic.status.ready')
          }"
          :status="curCluster.status"
          :pending="['INITIALIZATION', 'DELETING'].includes(curCluster.status)"
          v-if="curCluster.status" />
      </div>
    </div>
    <bcs-tab
      :label-height="42"
      :active.sync="activeTabName"
      :validate-active="false"
      type="card-tab"
      class="cluster-detail-tab"
      v-if="panelStatus !== 'hidden'">
      <!-- 渲染标签页 -->
      <bcs-tab-panel
        v-for="(tab, index) in realTabs"
        :key="`${index}-${tab.name}`"
        :name="tab.name"
        :label="$t(tab.label)"
        render-directive="if">
        <template #label>
          {{ $t(tab.label) }}
        </template>
        <component
          :is="tab.component"
          class="p-[20px]"
          v-bind="tab.componentConfig"
          v-on="tab.eventHandlers" />
      </bcs-tab-panel>
    </bcs-tab>
  </div>
</template>
<script lang="ts" setup>
import { computed, defineExpose, onMounted, ref, watch } from 'vue';

import AutoScaler from '../autoscaler/autoscaler.vue';
import Node from '../node-list/node.vue';

import StatusIcon from '@/components/status-icon';
import { ICluster, useCluster } from '@/composables/use-app';
import $router from '@/router';
import Info from '@/views/cluster-manage/detail/basic-info.vue';
import TaskRecord from '@/views/cluster-manage/detail/cluster-record.vue';
import Master from '@/views/cluster-manage/detail/master/index.vue';
import Network from '@/views/cluster-manage/detail/network/index.vue';
import Overview from '@/views/cluster-manage/detail/overview.vue';
import SubCluster from '@/views/cluster-manage/detail/sub-cluster/index.vue';
import VClusterQuota from '@/views/cluster-manage/detail/vcluster-quota.vue';
import Namespace from '@/views/cluster-manage/namespace/namespace.vue';

const props = defineProps({
  maxWidth: {
    type: Number,
    default: 1000,
  },
  active: {
    type: String,
    default: 'overview',
  },
  clusterId: {
    type: String,
    default: '',
    required: true,
  },
  namespace: {
    type: String,
    default: '',
  },
  perms: {
    type: Object,
    default: () => ({}),
  },
});

const emits = defineEmits(['width-change', 'active-row']);

watch(() => props.active, () => {
  activeTabName.value = props.active;
});

// 正常状态
const normalStatusList = ref(['CONNECT-FAILURE', 'RUNNING']);

const activeTabName = ref(props.active);
watch(activeTabName, () => {
  if (props.active === activeTabName.value) return;
  // 更新路由
  $router.replace({
    query: {
      ...$router.currentRoute.query,
      active: activeTabName.value,
    },
  });
});
const { clusterList } = useCluster();
const curCluster = computed<Partial<ICluster>>(() => clusterList.value
  .find(item => item.clusterID === props.clusterId) || {});
// kubeConfig导入集群
const isKubeConfigImportCluster = computed(() => curCluster.value.clusterCategory === 'importer'
      && curCluster.value.importCategory === 'kubeConfig');
// 控制面导入集群
const isKubeAgentImportCluster = computed(() => curCluster.value.clusterCategory === 'importer' && curCluster.value.importCategory === 'machine');
// // 云区域详情
// const cloudDetail = ref<Record<string, any>|null>(null);
// const handleGetCloudDetail = async () => {
//   cloudDetail.value = await $store.dispatch('clustermanager/cloudDetail', {
//     $cloudId: curCluster.value?.provider,
//   });
// };
const showAutoScaler = computed(() => !!curCluster.value?.autoScale);

const tabs = computed(() => [
  {
    // 集群总览
    name: 'overview', // 默认显示的tab
    label: 'cluster.detail.title.overview', // tab名称
    component: Overview, // tab对应的组件
    isShow: !curCluster.value?.is_shared
      && curCluster.value.clusterType !== 'federation'
      && normalStatusList.value.includes(curCluster.value.status || ''), // 是否显示该tab
    componentConfig: {
      clusterId: props.clusterId,
    }, // tab对应的组件配置
  },
  {
    // 命名空间
    name: 'namespace',
    label: 'k8s.namespace',
    component: Namespace,
    isShow: (curCluster.value?.is_shared
    || normalStatusList.value.includes(curCluster.value.status || '')),
    componentConfig: {
      clusterId: props.clusterId,
      namespace: props.namespace,
    },
  },
  {
    // 基本信息
    name: 'info',
    label: 'generic.title.basicInfo1',
    component: Info,
    isShow: true,
    componentConfig: {
      clusterId: props.clusterId,
    },
  },
  {
    // 网络配置
    name: 'network',
    label: 'cluster.detail.title.network',
    component: Network,
    isShow: (curCluster.value?.is_shared
      || (!['virtual', 'federation'].includes(curCluster.value.clusterType || '')
        && !isKubeConfigImportCluster.value
        && !isKubeAgentImportCluster.value)),
    componentConfig: {
      clusterId: props.clusterId,
    },
  },
  {
    // 控制面配置
    name: 'master',
    label: 'cluster.detail.title.controlConfig',
    component: Master,
    isShow: isKubeConfigImportCluster.value || (!curCluster.value?.is_shared
      && !['virtual', 'federation'].includes(curCluster.value.clusterType || '')),
    componentConfig: {
      clusterId: props.clusterId,
    },
  },
  {
    // 节点列表
    name: 'node',
    label: 'cluster.detail.title.nodeList',
    component: Node,
    isShow: (!curCluster.value?.is_shared
      && !['virtual', 'federation'].includes(curCluster.value.clusterType || '')
      && normalStatusList.value.includes(curCluster.value.status || '')),
    componentConfig: {
      clusterId: props.clusterId,
      fromCluster: true,
      hideClusterSelect: true,
    },
  },
  {
    // 自动扩缩容
    name: 'autoscaler',
    label: 'cluster.detail.title.autoScaler',
    component: AutoScaler,
    isShow: (!curCluster.value?.is_shared
      && !['virtual', 'federation'].includes(curCluster.value.clusterType || '')
      && normalStatusList.value.includes(curCluster.value.status || '')
      && showAutoScaler.value),
    componentConfig: {
      clusterId: props.clusterId,
    },
  },
  {
    // 操作记录
    name: 'taskRecord',
    label: 'cluster.title.opRecord',
    component: TaskRecord,
    isShow: true,
    componentConfig: {
      clusterId: props.clusterId,
    },
  },
  {
    // 配额
    name: 'quota',
    label: 'cluster.detail.title.quota',
    component: VClusterQuota,
    isShow: (!curCluster.value?.is_shared
      && curCluster.value.clusterType === 'virtual'
      && normalStatusList.value.includes(curCluster.value.status || '')),
    componentConfig: {
      clusterId: props.clusterId,
    },
  },
  {
    // 成员集群
    name: 'subCluster',
    label: 'cluster.detail.title.subCluster',
    component: SubCluster,
    isShow: (curCluster.value.clusterType === 'federation'),
    componentConfig: {
      clusterId: props.clusterId,
      perms: props.perms,
    },
    eventHandlers: {
      'active-row': clusterID => handleChangeActiveRow(clusterID),
    },
  },
]);

const realTabs = computed(() => tabs.value.filter(tab => tab.isShow));

watch([
  () => showAutoScaler.value,
  () => activeTabName.value,
  () => props.clusterId,
], () => {
  if (!realTabs.value.some(item => item.name === activeTabName.value)) {
    activeTabName.value = realTabs.value?.[0]?.name || '';
  }
}, { immediate: true });

// resize event
const modalRef = ref();
const resizeRef = ref();
const minWidth = ref(400);
const onMousedownEvent = (e: MouseEvent) => {
  // 颜色改变提醒
  resizeRef.value.style.borderLeft = '1px solid #3a84ff';
  const startX = e.clientX;
  const { clientWidth } = modalRef.value;
  // 鼠标拖动事件
  document.onmousemove =  (e) => {
    document.body.style.userSelect = 'none';
    const endX = e.clientX;
    const moveLen = endX - startX;
    const width = clientWidth - moveLen;
    if (width < minWidth.value && moveLen >= 0) {
      hideDetailPanel();
    } else if (width > props.maxWidth && moveLen <= 0) {
      expandDetailPanel();
    } else {
      setPanelWidth(`${(width / document.body.clientWidth) * 100}%`);
    }
  };
  // 鼠标松开事件
  document.onmouseup =  () => {
    document.body.style.userSelect = '';
    resizeRef.value.style.borderLeft = '';
    document.onmousemove = null;
    document.onmouseup = null;
    resizeRef.value?.releaseCapture?.();
  };
  resizeRef.value?.setCapture?.();
};

const panelStatus = ref<'expanded'|'show'|'hidden'>('show');// expanded: 全部展开, show: 普通状态 hide: 隐藏状态
const setPanelWidth = (width: string) => {
  modalRef.value.style.width = width;
  panelStatus.value = 'show';
  emits('width-change', width);
};
// 全部展开面板
const expandDetailPanel = () => {
  modalRef.value.style.width = '100%';
  panelStatus.value = 'expanded';
};
// 显示面板
const showDetailPanel = () => {
  setPanelWidth('70%');
};
const toggleDetailPanel = (show: boolean) => {
  if (modalRef.value.style.width === '100%' || modalRef.value.style.width === '0px') {
    showDetailPanel();
  } else {
    if (show) {
      expandDetailPanel();
    } else {
      hideDetailPanel();
    }
  }
};
// 隐藏面板
const hideDetailPanel = () => {
  modalRef.value.style.width = '0px';
  panelStatus.value = 'hidden';
  emits('width-change', 0);
};

// 当前active行
const handleChangeActiveRow = (clusterID) => {
  emits('active-row', clusterID);
};

onMounted(() => {
  setTimeout(() => {
    setPanelWidth(`${(modalRef.value?.clientWidth / document.body.clientWidth) * 100}%`);
  });
});

defineExpose({
  showDetailPanel,
});

</script>
<style lang="postcss" scoped>
.modal {
  background: #f0f1f5;
  border-radius: 2px;
  bottom: 0;
  position: absolute;
  right: 0px;
  top: 0px;
  height: 100%;
  z-index: 10;
  width: 70%;
  .resize-line {
    width: 8px;
    border-left: 1px solid #DCDEE5;
    height: 100%;
    cursor: col-resize;
    position: absolute;
    z-index: 2;
    &::after {
      content: "";
      position: absolute;
      width: 2px;
      height: 2px;
      color: #63656E;
      background: #63656E;
      box-shadow: 0 4px 0 0 #63656E,0 8px 0 0 #63656E,0 -4px 0 0 #63656E,0 -8px 0 0 #63656E;
      left: 4px;
      top: 50%;
      transform: translate3d(0, -50%, 0);
    }
  }
  .fold {
    position: absolute;
    top: 16px;
    left: -16px;
    display: flex;
    align-items: center;
    z-index: 3;
    color: #979BA5;
    &-left {
      display: flex;
      align-items: center;
      justify-content: center;
      width: 16px;
      height: 32px;
      background: #FAFBFD;
      border: 1px solid #DCDEE5;
      border-radius: 4px 0 0 4px;
      cursor: pointer;
      position: relative;
      right: -1px;
      &.active {
        background-color: #699DF4;
        border-color: #699DF4;
        color: #fff;
      }
    }
    &-right {
      display: flex;
      align-items: center;
      justify-content: center;
      width: 16px;
      height: 32px;
      background: #FAFBFD;
      border: 1px solid #DCDEE5;
      border-radius: 0 4px 4px 0;
      cursor: pointer;
      &.active {
        background-color: #699DF4;
        border-color: #699DF4;
        color: #fff;
      }
    }
  }
}

>>> .cluster-detail-tab {
  height: calc(100% - 56px);
  .bk-tab-header {
    padding-left: 24px;
  }
  .bk-tab-section {
    padding: 0;
    height: calc(100% - 42px);
    overflow-y: auto;
    overflow-x: hidden;

    .bk-tab-content {
      height: 100%;
    }
  }
}
</style>
