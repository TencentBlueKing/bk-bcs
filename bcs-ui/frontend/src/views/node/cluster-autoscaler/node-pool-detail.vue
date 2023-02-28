<!-- eslint-disable max-len -->
<template>
  <BcsContent>
    <template #header>
      <HeaderNav :list="navList"></HeaderNav>
    </template>
    <div class="node-pool-detail" v-bkloading="{ isLoading: loading }">
      <div class="mb15 panel-header">
        <span class="title">{{$t('基础信息')}}</span>
        <bk-button theme="primary" @click="handleEditPool">{{ $t('编辑') }}</bk-button>
      </div>
      <template v-if="nodePoolData">
        <bk-form class="content-wrapper">
          <bk-form-item :label="$t('节点规格名称')">
            {{`${nodePoolData.nodeGroupID} (${nodePoolData.name}) `}}
          </bk-form-item>
          <bk-form-item :label="$t('节点规格状态')">
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
          <bk-form-item :label="$t('节点配额')">
            {{ nodePoolData.autoScaling.maxSize }}
          </bk-form-item>
          <bk-form-item
            :label="$t('是否开启调度')"
            :desc="$t('节点规格启用后Autoscaler组件将会根据扩容算法使用该节点规格资源，开启Autoscaler组件后必须要开启至少一个节点规格')">
            {{nodePoolData.nodeTemplate.unSchedulable ? $t('否') : $t('是')}}
          </bk-form-item>
          <bk-form-item :label="$t('标签')">
            <bk-button
              text
              size="small"
              style="padding: 0;"
              v-if="labels.length"
              @click="showLabels = true">{{$t('查看')}}</bk-button>
            <span v-else>--</span>
          </bk-form-item>
          <bk-form-item :label="$t('污点')">
            <bk-button
              text
              size="small"
              style="padding: 0;"
              v-if="taints.length"
              @click="showTaints = true">{{$t('查看')}}</bk-button>
            <span v-else>--</span>
          </bk-form-item>
          <bk-form-item
            :label="$t('扩缩容模式')"
            :desc="$t('释放模式：缩容时自动释放Cluster AutoScaler判断的空余节点， 扩容时自动创建新的CVM节点加入到伸缩组<br/>关机模式：扩容时优先对已关机的节点执行开机操作，节点数依旧不满足要求时再创建新的CVM节点')">
            {{scalingModeMap[nodePoolData.autoScaling.scalingMode]}}
          </bk-form-item>
          <bk-form-item
            :label="$t('实例创建策略')"
            :desc="$t('首选可用区（子网）优先：自动扩缩容会在您首选的可用区优先执行扩缩容，若首选可用区无法扩缩容，才会在其他可用区进行扩缩容<br/>多可用区（子网）打散 ：在节点规格指定的多可用区（即指定多个子网）之间尽最大努力均匀分配CVM实例，只有配置了多个子网时该策略才能生效')">
            {{multiZoneSubnetPolicyMap[nodePoolData.autoScaling.multiZoneSubnetPolicy]}}
          </bk-form-item>
          <bk-form-item
            :label="$t('重试策略')"
            :desc="$t('快速重试 ：立即重试，在较短时间内快速重试，连续失败超过一定次数（5次）后不再重试，<br/>间隔递增重试 ：间隔递增重试，随着连续失败次数的增加，重试间隔逐渐增大，重试间隔从秒级到1天不等，<br/>不重试：不进行重试，直到再次收到用户调用或者告警信息后才会重试')">
            {{retryPolicyMap[nodePoolData.autoScaling.retryPolicy]}}
          </bk-form-item>
          <bk-form-item :label="$t('镜像提供方')">
            {{ imageProvider || '--'}}
          </bk-form-item>
          <bk-form-item :label="$t('操作系统')">
            {{clusterOS || '--'}}
          </bk-form-item>
          <bk-form-item :label="$t('运行时组件')">
            {{`${clusterData.clusterAdvanceSettings
              ? `${clusterData.clusterAdvanceSettings.containerRuntime} ${clusterData.clusterAdvanceSettings.runtimeVersion}`
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
          <bk-form-item label="系统盘">
            {{`${diskTypeMap[nodePoolData.launchTemplate.systemDisk.diskType]} ${nodePoolData.launchTemplate.systemDisk.diskSize}G`}}
          </bk-form-item>
          <bk-form-item label="数据盘">
            <bk-button
              text
              size="small"
              style="padding: 0;"
              @click="showDataDisks = true">{{$t('查看')}}</bk-button>
          </bk-form-item>
          <bk-form-item
            :label="$t('支持子网')"
            :desc="$t('内部上云环境根据集群所在VPC由产品自动分配可用子网，尽可能的把集群内的节点分配在不同的可用区，避免集群节点集中在同一可用区')">
            {{nodePoolData.autoScaling.subnetIDs.join(', ') || '--'}}
          </bk-form-item>
          <bk-form-item :label="$t('安全组')">
            <LoadingIcon v-if="securityGroupLoading">{{ $t('加载中') }}...</LoadingIcon>
            <span v-else>{{ securityGroupNames.join(',') || '--'}}</span>
          </bk-form-item>
          <bk-form-item
            :label="$t('扩容后转移模块')"
            :desc="$t('扩容节点后节点转移到关联业务的CMDB模块')">
            {{nodePoolData.nodeTemplate.module.scaleOutModuleName || '--'}}
          </bk-form-item>
          <!-- <bk-form-item :label="$t('缩容后转移模块')">
            {{nodePoolData.nodeTemplate.module.scaleInModuleName || '--'}}
          </bk-form-item> -->
        </bk-form>
        <div class="mt20 mb10 panel-header">
          <span class="title">{{$t('Kubelet组件参数')}}</span>
        </div>
        <kubeletParams readonly v-model="nodePoolData.nodeTemplate.extraArgs.kubelet" />
        <bcs-tab class="mt20">
          <bcs-tab-panel :label="$t('扩容前置初始化')" name="scaleOutPreAction">
            <UserAction
              :script="nodePoolData.nodeTemplate.preStartUserScript"
              key="scaleOutPreAction" />
          </bcs-tab-panel>
          <bcs-tab-panel :label="$t('扩容后置初始化')" name="scaleOutPostAction">
            <UserAction
              :script="nodePoolData.nodeTemplate.userScript"
              :addons="nodePoolData.nodeTemplate.scaleOutExtraAddons"
              actions-key="postActions"
              key="scaleOutPostAction" />
          </bcs-tab-panel>
          <bcs-tab-panel :label="$t('节点回收前清理配置')" name="scaleInPreAction">
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
    <!-- 数据盘 -->
    <bcs-dialog
      theme="primary"
      v-model="showDataDisks"
      :show-footer="false"
      :title="$t('数据盘')"
      header-position="left"
      width="600">
      <bcs-table
        :data="dataDisks"
        :outer-border="false"
        :header-border="false"
        :header-cell-style="{ background: '#fff' }"
      >
        <bcs-table-column :label="$t('类型')" prop="diskType">
          <template #default="{ row }">
            {{ diskTypeMap[row.diskType] }}
          </template>
        </bcs-table-column>
        <bcs-table-column :label="$t('大小')" prop="diskSize">
          <template #default="{ row }">
            {{ `${row.diskSize} GB` }}
          </template>
        </bcs-table-column>
        <bcs-table-column :label="$t('文件系统')" prop="fileSystem"></bcs-table-column>
        <bcs-table-column :label="$t('挂载点')" prop="mountTarget"></bcs-table-column>
      </bcs-table>
    </bcs-dialog>
  </BcsContent>
</template>
<script lang="ts">
import { defineComponent, computed, ref, onMounted } from '@vue/composition-api';
import BcsContent from '../bcs-content.vue';
import HeaderNav from '../header-nav.vue';
import { useClusterList, useClusterInfo } from '@/views/cluster/use-cluster';
import $i18n from '@/i18n/i18n-setup';
import $store from '@/store/index';
import $router from '@/router/index';
import StatusIcon from '@/views/dashboard/common/status-icon';
import LoadingIcon from '@/components/loading-icon.vue';
import useInterval from '@/views/dashboard/common/use-interval';
import kubeletParams from './kubelet-params.vue';
import UserAction from './user-action.vue';

export default defineComponent({
  components: {
    BcsContent,
    HeaderNav,
    StatusIcon,
    LoadingIcon,
    kubeletParams,
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
  setup(props, ctx) {
    const { clusterList } = useClusterList(ctx);
    const showDataDisks = ref(false);
    const showLabels = ref(false);
    const showTaints = ref(false);
    const nodePoolData = ref<any>(null);
    const loading = ref(false);
    const statusTextMap = { // 节点规格状态
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
    const diskTypeMap = {
      CLOUD_PREMIUM: $i18n.t('高性能云硬盘'),
      CLOUD_SSD: $i18n.t('SSD云硬盘'),
      CLOUD_HSSD: $i18n.t('增强型SSD云硬盘'),
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
            active: 'AutoScaler',
          },
        },
      },
      {
        title: $i18n.t('查看节点规格详情'),
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
      });
    };

    const labels = computed(() => Object.keys(nodePoolData.value?.nodeTemplate?.labels || {}).map(key => ({
      key,
      value: nodePoolData.value.nodeTemplate.labels[key],
    })));
    const taints = computed(() => nodePoolData.value?.nodeTemplate?.taints || []);
    const dataDisks = computed(() => nodePoolData.value?.nodeTemplate?.dataDisks || []);

    // 集群详情
    const { clusterOS, clusterData, getClusterDetail } = useClusterInfo();
    const imageProvider = computed(() => {
      const imageProviderMap = {
        PUBLIC_IMAGE: $i18n.t('公共镜像'),
        PRIVATE_IMAGE: $i18n.t('自定义镜像'),
      };
      return imageProviderMap[clusterData.value.extraInfo?.IMAGE_PROVIDER];
    });

    // 安全组信息
    const securityGroupsList = ref<any[]>([]);
    const securityGroupNames = computed(() => securityGroupsList.value
      .filter(item => nodePoolData.value.launchTemplate.securityGroupIDs.includes(item.securityGroupID))
      .map(item => item.securityGroupName));

    const securityGroupLoading = ref(false);
    const handleGetCloudSecurityGroups = async () => {
      securityGroupLoading.value = true;
      securityGroupsList.value = await $store.dispatch('clustermanager/cloudSecurityGroups', {
        $cloudID: clusterData.value.provider,
        region: clusterData.value.region,
        accountID: clusterData.value.cloudAccountID,
      });
      securityGroupLoading.value = false;
    };

    onMounted(async () => {
      loading.value = true;
      await handleGetNodeGroupDetail();
      await getClusterDetail(props.clusterId, true);
      loading.value = false;
      handleGetCloudSecurityGroups();
    });
    return {
      securityGroupLoading,
      dataDisks,
      showDataDisks,
      diskTypeMap,
      securityGroupNames,
      imageProvider,
      clusterData,
      clusterOS,
      loading,
      showLabels,
      showTaints,
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
</style>
