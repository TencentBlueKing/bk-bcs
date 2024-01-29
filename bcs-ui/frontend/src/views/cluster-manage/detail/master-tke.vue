<template>
  <bk-form class="bcs-small-form">
    <!-- 腾讯云 - 独立集群 -->
    <template v-if="curCluster.manageType === 'INDEPENDENT_CLUSTER'">
      <bk-form-item :label="$t('cluster.labels.clusterType')">
        <span class="text-[#313238]">
          {{ $t('bcs.cluster.selfDeployed') }}
        </span>
      </bk-form-item>
      <bk-form-item :label="$t('cluster.labels.masterInfo')">
        <div class="flex max-w-[800px]">
          <bk-table :data="masterData" v-bkloading="{ isLoading }">
            <bk-table-column :label="$t('cluster.labels.hostName')">
              <template #default="{ row }">
                {{ row.nodeName || '--' }}
              </template>
            </bk-table-column>
            <bk-table-column label="IPv4">
              <template #default="{ row }">
                {{ row.innerIP || '--' }}
              </template>
            </bk-table-column>
            <bk-table-column label="IPv6">
              <template #default="{ row }">
                {{ row.innerIPv6 || '--' }}
              </template>
            </bk-table-column>
            <template v-if="$INTERNAL">
              <bk-table-column :label="$t('generic.ipSelector.label.idc')" prop="idc"></bk-table-column>
              <bk-table-column :label="$t('cluster.labels.rack')" prop="rack"></bk-table-column>
              <bk-table-column
                :label="$t('generic.ipSelector.label.serverModel')"
                prop="deviceClass">
              </bk-table-column>
            </template>
          </bk-table>
        </div>
      </bk-form-item>
      <template v-if="['tencentPublicCloud', 'tencentCloud'].includes(curCluster.provider || '')">
        <bk-form-item
          :label="$t('tke.label.masterModule.text')"
          :desc="$t('tke.tips.transferMasterCMDBModule')"
          class="master-module-item">
          <template v-if="isEditModule">
            <div class="flex items-center">
              <TopoSelector
                :placeholder="$t('generic.placeholder.select')"
                :cluster-id="clusterId"
                v-model="curModuleID"
                class="w-[360px]"
                @change="handleWorkerModuleChange"
                @node-data-change="handleNodeChange" />
              <span
                class="text-[12px] text-[#3a84ff] ml-[8px] cursor-pointer"
                text
                @click="handleSaveWorkerModule">{{ $t('generic.button.save') }}</span>
              <span
                class="text-[12px] text-[#3a84ff] ml-[8px] cursor-pointer"
                text
                @click="isEditModule = false">{{ $t('generic.button.cancel') }}</span>
            </div>
          </template>
          <template v-else>
            <span>
              {{
                clusterData.clusterBasicSettings && clusterData.clusterBasicSettings.module
                  ? clusterData.clusterBasicSettings.module.masterModuleName || '--'
                  : '--'
              }}
            </span>
            <span
              class="hover:text-[#3a84ff] cursor-pointer ml-[8px]"
              @click="handleEditWorkerModule">
              <i class="bk-icon icon-edit-line"></i>
            </span>
          </template>
        </bk-form-item>
        <bk-form-item :label="$t('tke.label.apiServerCLB.text')" v-if="curCluster.provider === 'tencentPublicCloud'">
          <ClusterConnect
            :cluster-connect-setting="curCluster.clusterAdvanceSettings.clusterConnectSetting"
            :security-group-name="securityGroupName"
            :api-server="apiServer" />
        </bk-form-item>
      </template>
    </template>
    <!-- 腾讯云 - 托管集群 -->
    <template v-else>
      <bk-form-item :label="$t('cluster.labels.clusterType')">
        <span class="text-[#313238]">
          {{ $t('bcs.cluster.managed') }}
        </span>
        <span class="text-[#979BA5]">
          ({{ $t('cluster.create.label.manageType.managed.desc') }})
        </span>
      </bk-form-item>
      <bk-form-item :label="$t('cluster.create.label.manageType.managed.clusterLevel.text')">
        <span class="text-[#313238]">{{ clusterLevel }}</span>
        <span class="text-[#979BA5]">
          ({{
            $t('cluster.create.label.manageType.managed.clusterLevel.desc', {
              nodes: curClusterScale.level.split('L')[1],
              pods: curClusterScale.scale.maxNodePodNum,
              service: curClusterScale.scale.maxServiceNum,
              crd: curClusterScale.scale.cidrStep
            })
          }})
        </span>
        <bk-checkbox disabled :value="isAutoUpgradeClusterLevel" class="ml-[10px]">
          <span class="flex items-center">
            <span
              class="text-[12px] bcs-border-tips"
              v-bk-tooltips="{ content: $t('cluster.create.label.manageType.managed.automatic.tips') }">
              {{ $t('cluster.create.label.manageType.managed.automatic.text') }}
            </span>
          </span>
        </bk-checkbox>
      </bk-form-item>
      <bk-form-item :label="$t('tke.label.apiServerCLB.text')">
        <ClusterConnect
          :cluster-connect-setting="curCluster.clusterAdvanceSettings.clusterConnectSetting"
          :security-group-name="securityGroupName"
          :api-server="apiServer" />
      </bk-form-item>
    </template>
  </bk-form>
</template>
<script lang="ts">
import { computed, defineComponent, onBeforeMount, ref } from 'vue';

import clusterScaleData from '../add/common/cluster-scale.json';
import { useClusterInfo, useClusterList } from '../cluster/use-cluster';

import ClusterConnect from './cluster-connect.vue';

import { masterList, setClusterModule } from '@/api/modules/cluster-manager';
import $bkMessage from '@/common/bkmagic';
import { copyText } from '@/common/util';
import $bkInfo from '@/components/bk-magic-2.0/bk-info';
import { ICluster } from '@/composables/use-app';
import $i18n from '@/i18n/i18n-setup';
import $store from '@/store';
import TopoSelector from '@/views/cluster-manage/autoscaler/topo-select-tree.vue';
import useCloud from '@/views/cluster-manage/use-cloud';

export default defineComponent({
  name: 'TKEMaster',
  components: { ClusterConnect, TopoSelector },
  props: {
    clusterId: {
      type: String,
      default: '',
      required: true,
    },
  },
  setup(props) {
    const { clusterList } = useClusterList();
    const { clusterData, getClusterDetail } = useClusterInfo();// clusterData和curCluster一样，就是多了云上的数据信息
    const apiServer = computed(() => clusterData.value?.extraInfo?.apiServer || '');
    const curCluster = computed<Partial<ICluster>>(() => clusterList.value
      .find(item => item.clusterID === props.clusterId) || {});

    const isAutoUpgradeClusterLevel = computed(() => curCluster.value.clusterBasicSettings.isAutoUpgradeClusterLevel);

    // 修改master转移模块设置
    const isEditModule = ref(false);
    const curModuleID = ref();
    const curNodeModule = ref<Record<string, any>>({});
    const handleEditWorkerModule = () => {
      curModuleID.value = Number(clusterData.value.clusterBasicSettings?.module?.masterModuleID);
      isEditModule.value = true;
    };
    const handleWorkerModuleChange = (moduleID) => {
      curModuleID.value = moduleID;
    };
    const handleNodeChange = (node) => {
      curNodeModule.value = node;
    };
    const handleSaveWorkerModule = async () => {
      if (curModuleID.value === clusterData.value.clusterBasicSettings?.module?.masterModuleID) {
        isEditModule.value = false;
        return;
      };

      $bkInfo({
        type: 'warning',
        clsName: 'custom-info-confirm',
        title: $i18n.t('tke.title.confirmUpdateMasterCMDBModule'),
        subTitle: $i18n.t('tke.title.confirmUpdateMasterCMDBModuleSubTitle', [curNodeModule.value.path]),
        defaultInfo: true,
        confirmFn: async () => {
          const result = await setClusterModule({
            $clusterId: props.clusterId,
            module: {
              masterModuleID: curModuleID.value,
            },
            operator: $store.state.user?.username,
          }).then(() => true)
            .catch(() => false);
          if (result) {
            await getClusterDetail(props.clusterId, true);
            $bkMessage({
              theme: 'success',
              message: $i18n.t('generic.msg.success.modify'),
            });
            isEditModule.value = false;
          }
        },
      });
    };

    // 托管集群集群规格信息
    const clusterLevel = computed(() => curCluster.value?.clusterBasicSettings?.clusterLevel || '--');
    const clusterScale = ref(clusterScaleData.data);
    const curClusterScale = computed(() => clusterScale.value
      .find(item => item.level === clusterLevel.value)
      || { level: '', scale: { maxNodePodNum: 0, maxServiceNum: 0, cidrStep: 0 } });

    // 独立集群Master信息
    const isLoading = ref(false);
    const masterData = ref<any[]>([]);
    const handleGetMasterData = async () => {
      isLoading.value = true;
      masterData.value = await masterList({
        $clusterId: props.clusterId,
      }).catch(() => []);
      isLoading.value = false;
    };

    // 复制IP
    const handleCopyIPv4 = () => {
      copyText(masterData.value.map(item => item.innerIP).join('\n'));
      $bkMessage({
        theme: 'success',
        message: $i18n.t('generic.msg.success.copy'),
      });
    };

    const {
      securityGroups,
      handleGetSecurityGroups,
    } = useCloud();
    const securityGroupName = computed(() => {
      const id = curCluster.value.clusterAdvanceSettings?.clusterConnectSetting?.securityGroup;
      return securityGroups.value.find(item => item.securityGroupID === id)?.securityGroupName;
    });

    onBeforeMount(() => {
      if (Object.keys(curCluster.value.master || {}).length) {
        handleGetMasterData();
      }
      handleGetSecurityGroups({
        region: curCluster.value.region,
        cloudAccountID: curCluster.value.cloudAccountID,
        cloudID: curCluster.value.provider,
      });
      getClusterDetail(props.clusterId, true);// 获取云上信息
    });

    return {
      isLoading,
      curCluster,
      clusterData, // 全量数据
      apiServer,
      clusterLevel,
      curClusterScale,
      masterData,
      isAutoUpgradeClusterLevel,
      securityGroupName,
      handleCopyIPv4,
      isEditModule,
      curModuleID,
      handleEditWorkerModule,
      handleWorkerModuleChange,
      handleSaveWorkerModule,
      handleNodeChange,
    };
  },
});
</script>
<style lang="postcss" scoped>
@import './form.css';

>>> .master-module-item {
  .bk-label {
    line-height: 32px !important;
  }
  .bk-form-content {
    line-height: 32px !important;
  }
}
</style>
