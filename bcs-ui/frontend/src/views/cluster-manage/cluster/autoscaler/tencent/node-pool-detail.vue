<!-- eslint-disable max-len -->
<template>
  <BcsContent>
    <template #header>
      <HeaderNav :list="navList"></HeaderNav>
    </template>
    <div class="node-pool-detail" v-bkloading="{ isLoading: loading }">
      <div class="mb15 panel-header">
        <span class="title">{{$t('节点池详情')}}</span>
        <bk-button theme="primary" @click="handleEditPool">{{ $t('编辑') }}</bk-button>
      </div>
      <bk-form class="content-wrapper" v-if="nodePoolData">
        <bk-form-item :label="$t('节点池名称')">
          {{`${nodePoolData.nodeGroupID} (${nodePoolData.name}) `}}
        </bk-form-item>
        <bk-form-item :label="$t('节点池状态')">
          <LoadingIcon v-if="['CREATING', 'DELETING', 'UPDATING'].includes(nodePoolData.status)">
            {{ statusTextMap[nodePoolData.status] }}
          </LoadingIcon>
          <StatusIcon status="unknown" v-else-if="!nodePoolData.enableAutoscale && nodePoolData.status === 'RUNNING'">
            {{$t('已关闭')}}
          </StatusIcon>
          <StatusIcon
            :status="nodePoolData.status"
            :status-color-map="statusColorMap"
            v-else>
            {{ statusTextMap[nodePoolData.status] }}
          </StatusIcon>
        </bk-form-item>
        <!-- <bk-form-item :label="$t('是否启用自动扩缩容')">
                    {{nodePoolData.enableAutoscale ? $t('是') : $t('否')}}
                </bk-form-item> -->
        <bk-form-item :label="$t('节点数量范围')">
          {{`${nodePoolData.autoScaling.minSize} ~ ${nodePoolData.autoScaling.maxSize}`}}
        </bk-form-item>
        <bk-form-item :label="$t('是否开启调度')">
          {{nodePoolData.nodeTemplate.unSchedulable ? $t('否') : $t('是')}}
        </bk-form-item>
        <bk-form-item label="Labels">
          <bk-button
            text
            size="small"
            style="padding: 0;"
            @click="showLabels = true">{{$t('查看')}}</bk-button>
        </bk-form-item>
        <bk-form-item label="Taints">
          <bk-button
            text
            size="small"
            style="padding: 0;"
            @click="showTaints = true">{{$t('查看')}}</bk-button>
        </bk-form-item>
        <bk-form-item :label="$t('扩缩容模式')">
          {{scalingModeMap[nodePoolData.autoScaling.scalingMode]}}
        </bk-form-item>
        <bk-form-item :label="$t('实例创建策略')">
          {{multiZoneSubnetPolicyMap[nodePoolData.autoScaling.multiZoneSubnetPolicy]}}
        </bk-form-item>
        <bk-form-item :label="$t('重试策略')">
          {{retryPolicyMap[nodePoolData.autoScaling.retryPolicy]}}
        </bk-form-item>
        <bk-form-item :label="$t('操作系统')">
          {{nodePoolData.launchTemplate.imageInfo.imageName}}
        </bk-form-item>
        <bk-form-item :label="$t('云区域')">
          {{cloud.bk_cloud_name || '--'}}
        </bk-form-item>
        <bk-form-item :label="$t('运行时组件')">
          {{`${nodePoolData.nodeTemplate.runtime
            ? `${nodePoolData.nodeTemplate.runtime.containerRuntime} ${nodePoolData.nodeTemplate.runtime.runtimeVersion}`
            : '--'}`
          }}
        </bk-form-item>
        <bk-form-item :label="$t('容器目录')">
          {{nodePoolData.nodeTemplate.dockerGraphPath || '--'}}
        </bk-form-item>
        <bk-form-item :label="$t('机型')">
          {{nodePoolData.launchTemplate.instanceType}}
        </bk-form-item>
        <bk-form-item label="CPU">
          {{`${nodePoolData.launchTemplate.CPU}${$t('核')}`}}
        </bk-form-item>
        <bk-form-item label="内存">
          {{nodePoolData.launchTemplate.Mem}}G
        </bk-form-item>
        <bk-form-item :label="$t('支持子网')">
          {{nodePoolData.autoScaling.subnetIDs.join(', ') || '--'}}
        </bk-form-item>
        <bk-form-item :label="$t('安全组')">
          {{nodePoolData.launchTemplate.securityGroupIDs.join(', ') || '--'}}
        </bk-form-item>
        <bk-form-item :label="$t('自定义数据')">
          <bk-button
            text
            size="small"
            style="padding: 0;"
            v-if="nodePoolData.nodeTemplate.userScript"
            @click="showUserScript = true">{{$t('查看')}}</bk-button>
          <span v-else>--</span>
        </bk-form-item>
      </bk-form>
    </div>
    <!-- 标签 -->
    <bcs-dialog
      theme="primary"
      v-model="showLabels"
      :show-footer="false"
      title="Labels"
      header-position="left"
      width="600">
      <bcs-table
        :data="labels"
        :outer-border="false"
        :header-border="false"
        :header-cell-style="{ background: '#fff' }"
      >
        <bcs-table-column :label="$t('键')" prop="key"></bcs-table-column>
        <bcs-table-column :label="$t('值')" prop="value"></bcs-table-column>
      </bcs-table>
    </bcs-dialog>
    <!-- 污点 -->
    <bcs-dialog
      theme="primary"
      v-model="showTaints"
      :show-footer="false"
      title="Taints"
      header-position="left"
      width="600">
      <bcs-table
        :data="taints"
        :outer-border="false"
        :header-border="false"
        :header-cell-style="{ background: '#fff' }"
      >
        <bcs-table-column :label="$t('键')" prop="key"></bcs-table-column>
        <bcs-table-column :label="$t('值')" prop="value"></bcs-table-column>
        <bcs-table-column label="Effect" prop="effect"></bcs-table-column>
      </bcs-table>
    </bcs-dialog>
    <!-- 自定义数据 -->
    <bcs-dialog
      theme="primary"
      v-model="showUserScript"
      :show-footer="false"
      :title="$t('自定义数据')"
      header-position="left"
      width="600">
      <div class="user-script-content">
        {{nodePoolData && nodePoolData.nodeTemplate.userScript}}
      </div>
    </bcs-dialog>
  </BcsContent>
</template>
<script lang="ts">
import { defineComponent, computed, ref, onMounted } from 'vue';
import BcsContent from '@/views/cluster-manage/components/bcs-content.vue';
import HeaderNav from '@/views/cluster-manage/components/header-nav.vue';
import { useClusterList } from '@/views/cluster-manage/cluster/use-cluster';
import $i18n from '@/i18n/i18n-setup';
import $store from '@/store/index';
import $router from '@/router';
import StatusIcon from '@/components/status-icon';
import LoadingIcon from '@/components/loading-icon.vue';
import useInterval from '@/composables/use-interval';
import { nodemanCloudList } from '@/api/base';

export default defineComponent({
  components: {
    BcsContent,
    HeaderNav,
    StatusIcon,
    LoadingIcon,
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
    const showLabels = ref(false);
    const showTaints = ref(false);
    const showUserScript = ref(false);
    const nodePoolData = ref<any>(null);
    const loading = ref(false);
    const statusTextMap = { // 节点池状态
      RUNNING: $i18n.t('正常'),
      CREATING: $i18n.t('创建中'),
      DELETING: $i18n.t('删除中'),
      UPDATING: $i18n.t('更新中'),
      DELETED: $i18n.t('已删除'),
      'CREATE-FAILURE': $i18n.t('创建失败'),
      'UPDATE-FAILURE': $i18n.t('更新失败'),
    };
    const statusColorMap = {
      RUNNING: 'green',
      DELETED: 'gray',
      'CREATE-FAILURE': 'red',
      'UPDATE-FAILURE': 'red',
    };
    const scalingModeMap = {
      CLASSIC_SCALING: $i18n.t('释放模式'),
      WAKE_UP_STOPPED_SCALING: $i18n.t('关机模式'),
    };
    const multiZoneSubnetPolicyMap = {
      EQUALITY: $i18n.t('多可用区（子网）打散'),
      PRIORITY: $i18n.t('首先可用区（子网）优先'),
    };
    const retryPolicyMap = {
      IMMEDIATE_RETRY: $i18n.t('快速重试'),
      INCREMENTAL_INTERVALS: $i18n.t('间隔递增重试'),
      NO_RETRY: $i18n.t('不重试'),
    };
    const navList = computed(() => [
      {
        title: clusterList.value.find(item => item.clusterID === props.clusterId)?.clusterName,
        link: {
          name: 'clusterDetail',
        },
      },
      {
        title: 'Cluster Autoscaler',
        link: {
          name: 'clusterDetail',
          query: {
            active: 'autoscaler',
          },
        },
      },
      {
        title: $i18n.t('查看节点池详情'),
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
    const cloudList = ref<any[]>([]);
    const cloud = computed(() => cloudList.value.find(item => item.bk_cloud_id === nodePoolData.value.bkCloudID) || {});
    const handleGetNodeGroupDetail = async () => {
      loading.value = true;
      await getNodeGroupDetail();
      cloudList.value = await nodemanCloudList().catch(() => []);
      if (['CREATING', 'DELETING', 'UPDATING'].includes(nodePoolData.value.status)) {
        start();
      }
      loading.value = false;
    };
    const { start, stop } = useInterval(getNodeGroupDetail);

    // 编辑节点池详情
    const handleEditPool = () => {
      $router.push({
        name: 'editNodePool',
        params: {
          clusterId: props.clusterId,
          nodeGroupID: props.nodeGroupID,
        },
      });
    };

    const labels = computed(() => Object.keys(nodePoolData.value?.nodeTemplate?.labels || {}).map(key => ({
      key,
      value: nodePoolData.value.nodeTemplate.labels[key],
    })));
    const taints = computed(() => nodePoolData.value?.nodeTemplate?.taints || []);

    onMounted(() => {
      handleGetNodeGroupDetail();
    });
    return {
      cloud,
      loading,
      showLabels,
      showTaints,
      showUserScript,
      labels,
      taints,
      navList,
      statusTextMap,
      statusColorMap,
      nodePoolData,
      handleEditPool,
      scalingModeMap,
      multiZoneSubnetPolicyMap,
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
.user-script-content {
    min-height: 200px;
    background: #F5F7FA;
    border-radius: 2px;
    padding: 8px 16px;
}
</style>
