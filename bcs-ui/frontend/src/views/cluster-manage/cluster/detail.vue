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
    <div class="resize-line" ref="resizeRef" @mousedown="onMousedownEvent"></div>
    <div class="h-[56px] flex items-center px-[24px] pt-[16px] pb-[8px] text-[12px]">
      <span class="font-bold text-[14px] leading-[22px]">{{ curCluster.clusterName || '--' }}</span>
      <span class="ml-[14px] text-[#979BA5] leading-[22px]">{{ curCluster.clusterID }}</span>
      <bcs-divider direction="vertical"></bcs-divider>
      <div class="h-[22px] inline-flex items-center">
        <StatusIcon
          :status-color-map="{
            'CREATE-FAILURE': 'red',
            'DELETE-FAILURE': 'red',
            'IMPORT-FAILURE': 'red',
            RUNNING: 'green'
          }"
          :status-text-map="{
            INITIALIZATION: $t('generic.status.initializing'),
            DELETING: $t('generic.status.deleting'),
            'CREATE-FAILURE': $t('generic.status.createFailed'),
            'DELETE-FAILURE': $t('generic.status.deleteFailed'),
            'IMPORT-FAILURE': $t('cluster.status.importFailed'),
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
      <bcs-tab-panel name="overview" :label="$t('cluster.detail.title.overview')" render-directive="if">
        <Overview class="px-[20px]" :cluster-id="clusterId" />
      </bcs-tab-panel>
      <bcs-tab-panel name="info" :label="$t('generic.title.basicInfo1')" render-directive="if">
        <Info class="p-[20px]" :cluster-id="clusterId" />
      </bcs-tab-panel>
      <bcs-tab-panel
        name="quota"
        :label="$t('cluster.detail.title.quota')"
        render-directive="if"
        v-if="curCluster.clusterType === 'virtual'">
        <VClusterQuota class="p-[20px]" :cluster-id="clusterId" />
      </bcs-tab-panel>
      <template v-else>
        <bcs-tab-panel name="network" :label="$t('cluster.detail.title.network')" render-directive="if">
          <Network class="p-[20px]" :cluster-id="clusterId" />
        </bcs-tab-panel>
        <bcs-tab-panel name="master" :label="$t('cluster.detail.title.controlConfig')" render-directive="if">
          <Master class="p-[20px]" :cluster-id="clusterId" />
        </bcs-tab-panel>
        <bcs-tab-panel name="node" :label="$t('cluster.detail.title.nodeList')" render-directive="if">
          <Node
            class="p-[20px] max-h-[calc(100vh-188px)]"
            :cluster-id="clusterId"
            hide-cluster-select
            from-cluster />
        </bcs-tab-panel>
        <bcs-tab-panel
          name="autoscaler"
          :label="$t('cluster.detail.title.autoScaler')"
          render-directive="if"
          ref="autoScalerTabRef"
          v-if="showAutoScaler">
          <template #label>
            {{ $t('cluster.detail.title.autoScaler') }}
            <bk-tag theme="danger">NEW</bk-tag>
          </template>
          <AutoScaler :cluster-id="clusterId" />
        </bcs-tab-panel>
      </template>
    </bcs-tab>
  </div>
</template>
<script lang="ts" setup>
import { computed, defineExpose, onMounted, ref, watch } from 'vue';

import AutoScaler from '../autoscaler/autoscaler.vue';
import Node from '../node-list/node.vue';

import StatusIcon from '@/components/status-icon';
import { useCluster } from '@/composables/use-app';
import $router from '@/router';
import Info from '@/views/cluster-manage/detail/basic-info.vue';
import Master from '@/views/cluster-manage/detail/master.vue';
import Network from '@/views/cluster-manage/detail/network.vue';
import Overview from '@/views/cluster-manage/detail/overview.vue';
import VClusterQuota from '@/views/cluster-manage/detail/vcluster-quota.vue';

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
});

const emits = defineEmits(['width-change']);

watch(() => props.active, () => {
  activeTabName.value = props.active;
});

const activeTabName = ref(props.active);
watch(activeTabName, () => {
  if ($router.currentRoute?.query?.active === activeTabName.value) return;
  // 更新路由
  $router.replace({
    query: {
      ...$router.currentRoute.query,
      active: activeTabName.value,
    },
  });
});
const { clusterList } = useCluster();
const curCluster = computed(() => clusterList.value.find(item => item.clusterID === props.clusterId) || {
  provider: '',
  clusterName: '',
  clusterID: '',
  clusterType: '',
});
// // 云区域详情
// const cloudDetail = ref<Record<string, any>|null>(null);
// const handleGetCloudDetail = async () => {
//   cloudDetail.value = await $store.dispatch('clustermanager/cloudDetail', {
//     $cloudId: curCluster.value?.provider,
//   });
// };
const showAutoScaler = computed(() => !!curCluster.value.autoScale);

watch(
  [
    () => showAutoScaler.value,
    () => activeTabName.value,
    () => props.clusterId,
  ],
  () => {
    setTimeout(() => {
      /**
       * - 当前集群不支持autoscaler需要跳转回overview tab
       * - 托管集群只有overview、basicInfo和quota三个tab详情
       */
      if (
        (!showAutoScaler.value && activeTabName.value === 'autoscaler')
        || (curCluster.value.clusterType === 'virtual' && !['overview', 'info', 'quota'].includes(activeTabName.value))
        || (curCluster.value.clusterType !== 'virtual' && activeTabName.value === 'quota')
      ) {
        activeTabName.value = 'overview';
      }
    });
  },
  { immediate: true },
);

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
      setPanelWidth(`${width}px`);
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
  const defaultWidth = document.body.clientWidth * 0.7;
  setPanelWidth(`${defaultWidth}px`);
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

// onBeforeMount(() => {
//   handleGetCloudDetail();
// });

onMounted(() => {
  setPanelWidth(`${modalRef.value?.clientWidth}px`);
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
  .bk-tab-header {
    padding-left: 24px;
  }
  .bk-tab-section {
    padding: 0;
  }
  .bk-tab-content {
    height: calc(100vh - 150px);
    overflow-y: auto;
    overflow-x: hidden;
  }
}
</style>
