<template>
  <bk-form class="bcs-small-form" :label-width="160" v-bkloading="{ isLoading }">
    <DescList :title="$t('generic.label.basicConfig')">
      <bk-form-item :label="$t('cluster.labels.region')">
        {{ clusterData.region || '--' }}
      </bk-form-item>
      <bk-form-item :label="$t('cluster.create.label.privateNet.text')">
        {{ clusterData.vpcID || '--' }}
      </bk-form-item>
      <bk-form-item :label="$t('cluster.labels.networkType')">
        {{ networkType }}
      </bk-form-item>
      <bk-form-item label="IPVS" v-if="clusterData.provider !== 'bluekingCloud'">
        {{ IPVS }}
      </bk-form-item>
    </DescList>
    <DescList class="mt-[16px]" :title="$t('k8s.containerNetwork')">
      <!-- VPC-CNI模式 -->
      <template
        v-if="clusterData.clusterAdvanceSettings && clusterData.clusterAdvanceSettings.networkType === 'VPC-CNI'">
        <bk-form-item
          :label="$t('cluster.create.label.networkSetting.maxServiceNum.text')">
          {{ maxServiceNum }}
        </bk-form-item>
        <bk-form-item
          :label="$t('tke.label.serviceCidr')">
          {{ clusterData.networkSettings &&clusterData.networkSettings.serviceIPv4CIDR
            ? clusterData.networkSettings.serviceIPv4CIDR : '--' }}
        </bk-form-item>
        <bk-form-item
          :label="$t('tke.label.staticIpMode')">
          {{ clusterData.networkSettings && clusterData.networkSettings.isStaticIpMode
            ? $t('generic.status.support')
            : $t('generic.status.unSupport') }}
        </bk-form-item>
        <div class="px-[30px]">
          <bk-table :data="subnets" v-bkloading="{ isLoading: subnetLoading }">
            <bk-table-column :label="$t('tke.label.zone')" prop="zoneName"></bk-table-column>
            <bk-table-column :label="$t('tke.label.subnetID')" prop="subnetID"></bk-table-column>
            <bk-table-column :label="$t('tke.label.subnetName')" prop="subnetName"></bk-table-column>
            <bk-table-column label="CIDR" prop="cidrRange"></bk-table-column>
            <bk-table-column :label="$t('tke.label.ipNum')">
              <template #default="{ row }">
                <span>{{ `${row.availableIPAddressCount}/${countIPsInCIDR(row.cidrRange)}` }}</span>
              </template>
            </bk-table-column>
          </bk-table>
          <div
            :class="[
              'h-[36px] bg-[#FAFBFD] bcs-border mt-[-1px]',
              'flex items-center justify-center cursor-pointer'
            ]"
            @click="handleShowSubnetsDialog">
            <i class="bk-icon icon-plus-circle-shape mr5 text-[#979BA5] text-[14px]"></i>
            <span>{{ $t('tke.button.addSubnets') }}</span>
          </div>
        </div>
      </template>
      <!-- 其他 -->
      <template v-else>
        <!-- Google 云 -->
        <template v-if="clusterData.provider === 'gcpCloud'">
          <bk-form-item
            :label="$t('cluster.labels.maxPodNum')">
            <span>{{ cidrStep }}</span>
            <span>({{ cidr }})</span>
          </bk-form-item>
          <bk-form-item
            :label="$t('cluster.create.label.networkSetting.maxServiceNum.text')">
            <span>{{ maxServiceNum }}</span>
            <span>({{ serviceIPv4CIDR }})</span>
          </bk-form-item>
        </template>
        <!-- 腾讯公有云 -->
        <template v-else-if="clusterData.provider === 'tencentPublicCloud'">
          <bk-form-item :label="$t('tke.label.serviceIpNum')">
            {{ maxServiceNum }}
          </bk-form-item>
          <bk-form-item :label="$t('tke.label.serviceCidr')">
            {{ serviceIPv4CIDR }}
          </bk-form-item>
          <bk-form-item :label="$t('tke.label.podIpNum')">
            {{ cidr }}
          </bk-form-item>
          <bk-form-item :label="$t('tke.label.podCidr')">
            {{ cidrStep }}
          </bk-form-item>
        </template>
        <!-- 腾讯云 -->
        <template v-else>
          <bk-form-item
            :label="$t('cluster.labels.cidr')">
            {{ cidr }}
          </bk-form-item>
          <bk-form-item
            :label="$t('cluster.create.label.networkSetting.cidrStep.text')">
            {{ cidrStep }}
          </bk-form-item>
          <bk-form-item
            :label="$t('cluster.create.label.networkSetting.maxServiceNum.text')">
            {{ maxServiceNum }}
          </bk-form-item>
        </template>
        <!-- 单节点Pod数量上限 -->
        <bk-form-item
          :label="$t('cluster.create.label.networkSetting.maxNodePodNum.text')">
          {{ maxNodePodNum }}
        </bk-form-item>
      </template>
    </DescList>
    <bk-dialog
      :is-show.sync="showSubnets"
      :title="$t('tke.title.addSubnets')"
      :width="480"
      @cancel="showSubnets = false">
      <bk-form form-type="vertical">
        <bk-form-item label="Pod IP" required>
          <VpcCni :zone-list="zoneList" :subnets="newSubnets" @change="handleSetSubnets" />
        </bk-form-item>
      </bk-form>
      <template #footer>
        <div>
          <bk-button
            :disabled="!isSubnetsValidate"
            :loading="pending"
            theme="primary"
            @click="handleAddSubnets">{{ $t('generic.button.confirm') }}</bk-button>
          <bk-button @click="showSubnets = false">{{ $t('generic.button.cancel') }}</bk-button>
        </div>
      </template>
    </bk-dialog>
  </bk-form>
</template>
<script lang="ts">
import { computed, defineComponent, onMounted, ref } from 'vue';

import { ISubnet, IZoneItem } from '../add/tencent/types';
import VpcCni from '../add/tencent/vpc-cni.vue';
import { useClusterInfo } from '../cluster/use-cluster';

import { addSubnets, cloudSubnets,  cloudsZones } from '@/api/modules/cluster-manager';
import { countIPsInCIDR } from '@/common/util';
import DescList from '@/components/desc-list.vue';

export default defineComponent({
  name: 'ClusterNetwork',
  components: { DescList, VpcCni },
  props: {
    clusterId: {
      type: String,
      default: '',
      required: true,
    },
  },
  setup(props) {
    const { clusterData, isLoading, getClusterDetail } = useClusterInfo();
    const cidr = computed(() => {
      const { multiClusterCIDR = [], clusterIPv4CIDR = '' } = clusterData.value.networkSettings || {};
      return [...multiClusterCIDR, clusterIPv4CIDR].filter(cidr => !!cidr).join(', ') || '--';
    });
    const IPVS = computed(() => clusterData.value?.clusterAdvanceSettings?.IPVS);
    const maxServiceNum = computed(() => clusterData.value?.networkSettings?.maxServiceNum || '--');
    const serviceIPv4CIDR = computed(() => clusterData.value?.networkSettings?.serviceIPv4CIDR || '--');
    const maxNodePodNum = computed(() => clusterData.value?.networkSettings?.maxNodePodNum || '--');
    const cidrStep = computed(() => clusterData.value?.networkSettings?.cidrStep);
    const networkType = computed(() => (clusterData.value?.clusterAdvanceSettings?.networkType
      ? `${clusterData.value?.clusterAdvanceSettings?.networkType}(${clusterData.value.networkType})`
      : '--'));

    // 可用区
    const zoneList = ref<Array<IZoneItem>>([]);
    const zoneLoading = ref(false);
    const handleGetZoneList = async () => {
      const { provider, cloudAccountID, region } = clusterData.value;
      if (!provider || !cloudAccountID || !region) return;
      zoneLoading.value = true;
      const data = await cloudsZones({
        $cloudId: provider,
        accountID: cloudAccountID,
        region,
      }).catch(() => []);
      zoneList.value = data.filter(item => item.zoneState === 'AVAILABLE');
      zoneLoading.value = false;
    };

    // 子网
    const subnets = ref<Array<ISubnet>>([]);
    const subnetLoading = ref(false);
    const handleGetSubnets = async () => {
      const { provider, cloudAccountID, region, vpcID } = clusterData.value;
      if (!provider || !cloudAccountID || !region || !vpcID) return;

      subnetLoading.value = true;
      subnets.value = await cloudSubnets({
        $cloudId: provider,
        region,
        accountID: cloudAccountID,
        vpcID,
        subnetID: clusterData.value.networkSettings?.eniSubnetIDs?.join(','),
      }).catch(() => []);
      subnetLoading.value = false;
    };

    // 添加子网
    const showSubnets = ref(false);
    const newSubnets = ref([{
      ipCnt: 32,
      zone: '',
    }]);
    const isSubnetsValidate = computed(() => newSubnets.value.every(item => item.ipCnt && item.zone));
    const handleSetSubnets = (data) => {
      newSubnets.value = data;
    };
    const handleShowSubnetsDialog = () => {
      showSubnets.value = true;
    };
    const pending = ref(false);
    const handleAddSubnets = async () => {
      pending.value = true;
      const result = await addSubnets({
        $clusterId: props.clusterId,
        subnet: {
          new: newSubnets.value,
        },
      }).then(() => true)
        .catch(() => false);
      if (result) {
        await getClusterDetail(props.clusterId);
        await handleGetSubnets();
        showSubnets.value = false;
      }
      pending.value = false;
    };

    onMounted(async () => {
      await getClusterDetail(props.clusterId);
      await Promise.all([
        handleGetSubnets(),
        handleGetZoneList(),
      ]);
    });

    return {
      pending,
      zoneList,
      isSubnetsValidate,
      showSubnets,
      newSubnets,
      subnetLoading,
      subnets,
      isLoading,
      clusterData,
      cidr,
      IPVS,
      cidrStep,
      maxServiceNum,
      maxNodePodNum,
      serviceIPv4CIDR,
      networkType,
      countIPsInCIDR,
      handleShowSubnetsDialog,
      handleSetSubnets,
      handleAddSubnets,
    };
  },
});
</script>
<style lang="postcss" scoped>
@import './form.css';
</style>
