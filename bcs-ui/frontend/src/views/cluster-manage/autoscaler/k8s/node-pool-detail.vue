<!-- eslint-disable max-len -->
<template>
  <BcsContent>
    <template #header>
      <HeaderNav :list="navList"></HeaderNav>
    </template>
    <div class="node-pool-detail" v-bkloading="{ isLoading: loading }">
      <div class="mb15 panel-header">
        <span class="title">{{$t('generic.title.basicInfo')}}</span>
        <bk-button theme="primary" @click="handleEditPool">{{ $t('generic.button.edit') }}</bk-button>
      </div>
      <template v-if="nodePoolData">
        <bk-form class="content-wrapper">
          <bk-form-item :label="$t('cluster.ca.nodePool.label.name')">
            {{`${nodePoolData.nodeGroupID} (${nodePoolData.name}) `}}
          </bk-form-item>
          <bk-form-item :label="$t('cluster.ca.nodePool.label.status')">
            <LoadingIcon v-if="['CREATING', 'DELETING', 'UPDATING'].includes(nodePoolData.status)">
              {{ statusTextMap[nodePoolData.status] }}
            </LoadingIcon>
            <StatusIcon status="unknown" v-else-if="!nodePoolData.enableAutoscale && nodePoolData.status === 'RUNNING'">
              {{$t('cluster.ca.status.off')}}
            </StatusIcon>
            <StatusIcon
              :status="nodePoolData.status"
              :status-color-map="statusColorMap"
              v-else>
              {{ statusTextMap[nodePoolData.status] }}
            </StatusIcon>
          </bk-form-item>
          <bk-form-item :label="$t('cluster.ca.nodePool.label.nodeQuota')">
            {{ nodePoolData.autoScaling.maxSize }}
          </bk-form-item>
          <bk-form-item
            :label="$t('cluster.ca.nodePool.create.enableAutoscale.title1')"
            :desc="$t('cluster.ca.nodePool.create.enableAutoscale.tips')">
            {{nodePoolData.nodeTemplate.unSchedulable ? $t('units.boolean.false') : $t('units.boolean.true')}}
          </bk-form-item>
          <bk-form-item :label="$t('k8s.label')">
            <bk-button
              text
              size="small"
              style="padding: 0;"
              v-if="labels.length"
              @click="showLabels = true">{{$t('generic.button.view')}}</bk-button>
            <span v-else>--</span>
          </bk-form-item>
          <bk-form-item :label="$t('k8s.taint')">
            <bk-button
              text
              size="small"
              style="padding: 0;"
              v-if="taints.length"
              @click="showTaints = true">{{$t('generic.button.view')}}</bk-button>
            <span v-else>--</span>
          </bk-form-item>
          <bk-form-item :label="$t('k8s.annotation')">
            <bk-button
              text
              size="small"
              style="padding: 0;"
              v-if="annotations.length"
              @click="showAnnotations = true">{{$t('generic.button.view')}}</bk-button>
            <span v-else>--</span>
          </bk-form-item>
          <bk-form-item
            :label="$t('cluster.ca.nodePool.create.scalingMode.title')"
            :desc="$t('cluster.ca.nodePool.create.scalingMode.desc')">
            {{scalingModeMap[nodePoolData.autoScaling.scalingMode]}}
          </bk-form-item>
          <bk-form-item
            :label="$t('cluster.ca.nodePool.create.retryPolicy.title')"
            :desc="$t('cluster.ca.nodePool.create.retryPolicy.desc')">
            {{retryPolicyMap[nodePoolData.autoScaling.retryPolicy]}}
          </bk-form-item>
          <bk-form-item :label="$t('cluster.ca.nodePool.label.system')">
            {{clusterOS || '--'}}
          </bk-form-item>
          <bk-form-item :label="$t('cluster.ca.nodePool.create.containerRuntime.title')">
            {{`${clusterData.clusterAdvanceSettings
              ? `${clusterData.clusterAdvanceSettings.containerRuntime} ${clusterData.clusterAdvanceSettings.runtimeVersion}`
              : '--'}`
            }}
          </bk-form-item>
          <bk-form-item :label="$t('dashboard.workload.container.dataDir')">
            {{nodePoolData.nodeTemplate.dockerGraphPath || '--'}}
          </bk-form-item>
          <bk-form-item :label="$t('generic.ipSelector.label.serverModel')">
            {{nodePoolData.launchTemplate.instanceType}}
          </bk-form-item>
          <bk-form-item label="CPU">
            {{`${nodePoolData.launchTemplate.CPU}${$t('units.suffix.cores')}`}}
          </bk-form-item>
          <bk-form-item label="MEM">
            {{nodePoolData.launchTemplate.Mem}}G
          </bk-form-item>
        </bk-form>
        <bcs-tab class="mt20">
          <bcs-tab-panel :label="$t('cluster.ca.nodePool.create.scaleInitConfig.userScript')" name="scaleOutPostAction">
            <UserAction
              :script="nodePoolData.nodeTemplate.userScript"
              :addons="nodePoolData.nodeTemplate.scaleOutExtraAddons"
              actions-key="postActions"
              key="scaleOutPostAction" />
          </bcs-tab-panel>
          <bcs-tab-panel :label="$t('cluster.ca.nodePool.create.scaleInitConfig.scaleInPreScript')" name="scaleInPreAction">
            <UserAction
              :script="nodePoolData.nodeTemplate.scaleInPreScript"
              :addons="nodePoolData.nodeTemplate.scaleInExtraAddons"
              actions-key="preActions"
              key="scaleInPreAction" />
          </bcs-tab-panel>
        </bcs-tab>
      </template>
    </div>
    <!-- 标签 -->
    <bcs-dialog
      theme="primary"
      v-model="showLabels"
      :show-footer="false"
      :title="$t('k8s.label')"
      header-position="left"
      width="600">
      <bcs-table
        :data="labels"
        :outer-border="false"
        :header-border="false"
        :header-cell-style="{ background: '#fff' }"
      >
        <bcs-table-column :label="$t('generic.label.key')" prop="key"></bcs-table-column>
        <bcs-table-column :label="$t('generic.label.value')" prop="value"></bcs-table-column>
      </bcs-table>
    </bcs-dialog>
    <!-- 污点 -->
    <bcs-dialog
      theme="primary"
      v-model="showTaints"
      :show-footer="false"
      :title="$t('k8s.taint')"
      header-position="left"
      width="600">
      <bcs-table
        :data="taints"
        :outer-border="false"
        :header-border="false"
        :header-cell-style="{ background: '#fff' }"
      >
        <bcs-table-column :label="$t('generic.label.key')" prop="key"></bcs-table-column>
        <bcs-table-column :label="$t('generic.label.value')" prop="value"></bcs-table-column>
        <bcs-table-column label="Effect" prop="effect"></bcs-table-column>
      </bcs-table>
    </bcs-dialog>
    <!-- 注解 -->
    <bcs-dialog
      theme="primary"
      v-model="showAnnotations"
      :show-footer="false"
      :title="$t('k8s.annotation')"
      header-position="left"
      width="600">
      <bcs-table
        :data="annotations"
        :outer-border="false"
        :header-border="false"
        :header-cell-style="{ background: '#fff' }"
      >
        <bcs-table-column :label="$t('generic.label.key')" prop="key"></bcs-table-column>
        <bcs-table-column :label="$t('generic.label.value')" prop="value"></bcs-table-column>
      </bcs-table>
    </bcs-dialog>
  </BcsContent>
</template>
<script lang="ts">
import { computed, defineComponent, onMounted, ref } from 'vue';

import UserAction from '../components/user-action.vue';

import BcsContent from '@/components/layout/Content.vue';
import LoadingIcon from '@/components/loading-icon.vue';
import StatusIcon from '@/components/status-icon';
import useChainingRef from '@/composables/use-chaining';
import useInterval from '@/composables/use-interval';
import $i18n from '@/i18n/i18n-setup';
import $router from '@/router';
import $store from '@/store/index';
import { useClusterInfo, useClusterList } from '@/views/cluster-manage/cluster/use-cluster';
import HeaderNav from '@/views/cluster-manage/components/header-nav.vue';

export default defineComponent({
  components: {
    BcsContent,
    HeaderNav,
    StatusIcon,
    LoadingIcon,
    UserAction,
  },
  props: {
    clusterId: {
      type: String,
      default: '',
    },
    nodeGroupID: {
      type: String,
      default: '',
      required: true,
    },
  },
  setup(props) {
    const { clusterList } = useClusterList();
    const showDataDisks = ref(false);
    const showLabels = ref(false);
    const showTaints = ref(false);
    const nodePoolData = useChainingRef<any>({}, [
      'autoScaling',
      'nodeTemplate.extraArgs',
      'launchTemplate.systemDisk',
      {
        path: 'subnetIDs',
        type: 'array',
      },
    ]);
    const showAnnotations = ref(false);
    const loading = ref(false);
    const statusTextMap = { // 节点规格状态
      RUNNING: $i18n.t('generic.status.ready'),
      CREATING: $i18n.t('generic.status.creating'),
      DELETING: $i18n.t('generic.status.deleting'),
      UPDATING: $i18n.t('generic.status.updating'),
      DELETED: $i18n.t('generic.status.deleted'),
      'CREATE-FAILURE': $i18n.t('generic.status.createFailed'),
      'UPDATE-FAILURE': $i18n.t('generic.status.updateFailed'),
    };
    const statusColorMap = {
      RUNNING: 'green',
      DELETED: 'gray',
      'CREATE-FAILURE': 'red',
      'UPDATE-FAILURE': 'red',
    };
    const scalingModeMap = {
      CLASSIC_SCALING: $i18n.t('cluster.ca.nodePool.create.scalingMode.classic_scaling'),
      WAKE_UP_STOPPED_SCALING: $i18n.t('cluster.ca.nodePool.create.scalingMode.wake_up_stopped_scaling'),
    };
    const retryPolicyMap = {
      IMMEDIATE_RETRY: $i18n.t('cluster.ca.nodePool.create.retryPolicy.immediate_retry'),
      INCREMENTAL_INTERVALS: $i18n.t('cluster.ca.nodePool.create.retryPolicy.incremental_intervals'),
      NO_RETRY: $i18n.t('cluster.ca.nodePool.create.retryPolicy.no_retry'),
    };
    const diskTypeMap = {
      CLOUD_PREMIUM: $i18n.t('cluster.ca.nodePool.create.instanceTypeConfig.diskType.premium'),
      CLOUD_SSD: $i18n.t('cluster.ca.nodePool.create.instanceTypeConfig.diskType.ssd'),
      CLOUD_HSSD: $i18n.t('cluster.ca.nodePool.create.instanceTypeConfig.diskType.hssd'),
    };
    const navList = computed(() => [
      {
        title: clusterList.value.find(item => item.clusterID === props.clusterId)?.clusterName,
        link: {
          name: 'clusterMain',
        },
      },
      {
        title: 'Cluster Autoscaler',
        link: {
          name: 'clusterMain',
          query: {
            active: 'autoscaler',
            clusterId: props.clusterId,
          },
        },
      },
      {
        title: $i18n.t('cluster.ca.nodePool.label.detail'),
        link: null,
      },
    ]);

    // 获取详情
    const getNodeGroupDetail = async () => {
      nodePoolData.value = await $store.dispatch('clustermanager/nodeGroupDetail', {
        $nodeGroupID: props.nodeGroupID,
      });
      if (!['CREATING', 'DELETING', 'UPDATING'].includes(nodePoolData.value.status)) {
        stop();
      }
    };
    const handleGetNodeGroupDetail = async () => {
      await getNodeGroupDetail();
      if (['CREATING', 'DELETING', 'UPDATING'].includes(nodePoolData.value.status)) {
        start();
      }
    };
    const { start, stop } = useInterval(getNodeGroupDetail);

    // 编辑节点规格详情
    const handleEditPool = () => {
      $router.push({
        name: 'editNodePool',
        params: {
          clusterId: props.clusterId,
          nodeGroupID: props.nodeGroupID,
        },
        query: {
          provider: nodePoolData.value?.extraInfo?.resourcePoolType || 'yunti',
        },
      });
    };

    const labels = computed(() => Object.keys(nodePoolData.value?.nodeTemplate?.labels || {}).map(key => ({
      key,
      value: nodePoolData.value.nodeTemplate.labels[key],
    })));
    const annotations = computed(() => Object.keys(nodePoolData.value?.nodeTemplate?.annotations || {}).map(key => ({
      key,
      value: nodePoolData.value.nodeTemplate.annotations[key],
    })));
    const taints = computed(() => nodePoolData.value?.nodeTemplate?.taints || []);
    const dataDisks = computed(() => nodePoolData.value?.nodeTemplate?.dataDisks || []);

    // 集群详情
    const { clusterOS, clusterData, getClusterDetail } = useClusterInfo();

    onMounted(async () => {
      loading.value = true;
      await handleGetNodeGroupDetail();
      await getClusterDetail(props.clusterId, true);
      loading.value = false;
    });
    return {
      dataDisks,
      showDataDisks,
      diskTypeMap,
      clusterData,
      clusterOS,
      loading,
      showLabels,
      showTaints,
      showAnnotations,
      labels,
      taints,
      annotations,
      navList,
      statusTextMap,
      statusColorMap,
      nodePoolData,
      handleEditPool,
      scalingModeMap,
      retryPolicyMap,
    };
  },
});
</script>
<style lang="postcss" scoped>
.node-pool-detail {
    background: #fff;
    border: 1px solid #DDE4EB;
    border-radius: 2px;
    padding: 16px 32px 58px 32px;
    .panel-header {
        display: flex;
        align-items: center;
        justify-content: space-between;
        .title {
            font-size: 14px;
            font-weight: Bold;
            line-height: 22px;
        }
    }
    >>> .content-wrapper {
        display: flex;
        flex-wrap: wrap;
        .bk-form-item {
            width: 50%;
            margin-top: 0px;
        }
        .bk-label {
            font-size: 12px;
            color: #979BA5;
            text-align: left;
        }
        .bk-form-content {
            font-size: 12px;
            color: #313238;
            display: flex;
            align-items: center;
        }
    }
}
</style>
