<template>
  <bk-form class="bcs-small-form" :label-width="160" v-bkloading="{ isLoading }">
    <DescList :title="$t('generic.label.basicConfig')">
      <bk-form-item :label="$t('cluster.labels.region')">
        <LoadingIcon v-if="regionLoading">{{ $t('generic.status.loading') }}...</LoadingIcon>
        <span v-else>{{ region }}</span>
      </bk-form-item>
      <bk-form-item :label="$t('cluster.create.label.privateNet.text')">
        <LoadingIcon v-if="vpcLoading">{{ $t('generic.status.loading') }}...</LoadingIcon>
        <span v-else>{{ vpc }}</span>
      </bk-form-item>
      <bk-form-item :label="$t('cluster.labels.networkType')">
        {{ networkType }}
      </bk-form-item>
      <bk-form-item label="IPVS" v-if="clusterData.provider === 'tencentCloud'">
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
          <bk-table
            :data="subnets"
            :span-method="objectSpanMethod"
            col-border
            v-bkloading="{ isLoading: subnetLoading }"
            class="network-table">
            <bk-table-column :label="$t('tke.label.zone')" prop="zoneName"></bk-table-column>
            <bk-table-column :label="$t('tke.label.subnetID')" prop="subnetID"></bk-table-column>
            <bk-table-column :label="$t('tke.label.subnetName')" prop="subnetName"></bk-table-column>
            <bk-table-column label="CIDR" prop="cidrRange"></bk-table-column>
            <bk-table-column :label="$t('tke.label.ipNum')" prop="counts">
              <template #default="{ row }">
                <span>{{ getIpNumber(row) }}</span>
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
            <span>({{ clusterIPv4CIDR }})</span>
          </bk-form-item>
          <bk-form-item
            :label="$t('cluster.create.label.networkSetting.maxServiceNum.text')">
            <span>{{ maxServiceNum }}</span>
            <span>({{ serviceIPv4CIDR }})</span>
          </bk-form-item>
          <!-- 单节点Pod数量上限 -->
          <bk-form-item
            :label="$t('cluster.create.label.networkSetting.maxNodePodNum.text')">
            {{ maxNodePodNum }}
          </bk-form-item>
        </template>
        <!-- Microsoft 云 -->
        <template v-else-if="clusterData.provider === 'azureCloud'">
          <bk-form-item
            :label="$t('cluster.labels.maxPodNum')">
            <span>{{ cidrStep }}</span>
            <span>({{ clusterIPv4CIDR }})</span>
          </bk-form-item>
          <bk-form-item
            :label="$t('cluster.create.label.networkSetting.maxServiceNum.text')">
            <span>{{ maxServiceNum }}</span>
            <span>({{ serviceIPv4CIDR }})</span>
          </bk-form-item>
          <!-- 单节点Pod数量上限 -->
          <bk-form-item
            :label="$t('cluster.create.label.networkSetting.maxNodePodNum.text')">
            {{ maxNodePodNum }}
          </bk-form-item>
        </template>
        <!-- 腾讯公有云 -->
        <template v-else-if="clusterData.provider === 'tencentPublicCloud'">
          <bk-form-item :label="$t('tke.label.containerNet')">
            {{ `${clusterIPv4CIDR}(${clusterIPv4CIDRCount})` }}
          </bk-form-item>
          <bk-form-item label="Service IP">
            {{ maxServiceNum }}
          </bk-form-item>
          <bk-form-item label="Pod IP" :desc="$t('tke.tips.podIpCalcFormula')">
            {{ clusterIPv4CIDRCount - maxServiceNum}}
          </bk-form-item>
          <!-- 单节点Pod数量上限 -->
          <bk-form-item
            :label="$t('cluster.create.label.networkSetting.maxNodePodNum.text')">
            {{ maxNodePodNum }}
          </bk-form-item>
          <bk-form-item
            :label="$t('tke.label.clusterAvailableNodes')"
            :desc="$t('tke.tips.clusterAvailableNodes')">
            {{ maxNodePodNum ? (clusterIPv4CIDRCount - maxServiceNum) / maxNodePodNum : '--' }}
          </bk-form-item>
        </template>
        <!-- 腾讯云 -->
        <template v-else>
          <bk-form-item
            :label="$t('cluster.labels.cidr')">
            {{ clusterIPv4CIDR }}
          </bk-form-item>
          <bk-form-item
            :label="$t('cluster.create.label.networkSetting.cidrStep.text')">
            {{ clusterIPv4CIDRCount }}
          </bk-form-item>
          <bk-form-item
            :label="$t('cluster.create.label.networkSetting.maxServiceNum.text')">
            {{ maxServiceNum }}
          </bk-form-item>
          <!-- 单节点Pod数量上限 -->
          <bk-form-item
            :label="$t('cluster.create.label.networkSetting.maxNodePodNum.text')">
            {{ maxNodePodNum }}
          </bk-form-item>
        </template>
      </template>
    </DescList>
    <!-- 添加子网 -->
    <bk-dialog
      :is-show.sync="showSubnets"
      :title="$t('tke.title.addSubnets')"
      :width="480"
      @cancel="showSubnets = false">
      <bk-form form-type="vertical">
        <bk-form-item label="Pod IP" required>
          <VpcCni
            :subnets="newSubnets"
            :region="clusterData.region"
            :cloud-account-i-d="clusterData.cloudAccountID"
            :cloud-i-d="clusterData.provider"
            @change="handleSetSubnets" />
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

import { ISubnet } from '../add/tencent/types';
import VpcCni from '../add/tencent/vpc-cni.vue';
import { useClusterInfo } from '../cluster/use-cluster';

import { addSubnets, cloudSubnets } from '@/api/modules/cluster-manager';
import { countIPsInCIDR } from '@/common/util';
import DescList from '@/components/desc-list.vue';
import LoadingIcon from '@/components/loading-icon.vue';
import useCloud from '@/views/cluster-manage/use-cloud';

export default defineComponent({
  name: 'ClusterNetwork',
  components: { DescList, VpcCni, LoadingIcon },
  props: {
    clusterId: {
      type: String,
      default: '',
      required: true,
    },
  },
  setup(props) {
    const { clusterData, isLoading, getClusterDetail } = useClusterInfo();

    // 容器网段
    const clusterContainerCIDR = computed(() => {
      const { multiClusterCIDR = [], clusterIPv4CIDR = '' } = clusterData.value.networkSettings || {};
      return [...multiClusterCIDR, clusterIPv4CIDR];
    });
    const clusterIPv4CIDR = computed(() => clusterContainerCIDR.value.filter(cidr => !!cidr).join(', ') || '--');
    // 容器网段 IP数量
    const clusterIPv4CIDRCount = computed(() => clusterContainerCIDR.value.reduce((count, CIDR) => {
      count += countIPsInCIDR(CIDR) || 0;
      return count;
    }, 0));

    // service网段
    const serviceIPv4CIDR = computed(() => clusterData.value?.networkSettings?.serviceIPv4CIDR || '--');
    // service网段 IP数量
    const maxServiceNum = computed(() => clusterData.value?.networkSettings?.maxServiceNum || 0);

    // 自动扩容的IP步长
    const cidrStep = computed(() => clusterData.value?.networkSettings?.cidrStep);

    // Pod IP数量
    const maxNodePodNum = computed(() => clusterData.value?.networkSettings?.maxNodePodNum || 0);

    const IPVS = computed(() => clusterData.value?.clusterAdvanceSettings?.IPVS);

    // 网络插件类型
    const netTypeMap = {
      GR: 'Global Router',
    };
    const networkType = computed(() => {
      const type = clusterData.value?.clusterAdvanceSettings?.networkType;
      return type
        ? `${netTypeMap[type] || type}(${clusterData.value.networkType})`
        : '--';
    });

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
        resourceGroupName: clusterData.value.extraInfo?.nodeResourceGroup,
      }).catch(() => []);
      firstRowspan.value = {};
      // 排序子网，将相同子网放在一起
      subnets.value = subnets.value.sort((a, b) => {
        if (a.zone < b.zone) {
          return -1;
        }
        if (a.zone > b.zone) {
          return 1;
        }
        return 0;
      });
      subnetLoading.value = false;
    };

    // 可用区展示
    const zoneCounts = computed<Record<string, ISubnet[]>>(() => {
      const counts = {};
      subnets.value.forEach((item) => {
        if (!counts[item.zone]) {
          counts[item.zone] = [item];
        } else {
          counts[item.zone].push(item);
        }
      });
      return counts;
    });

    // 获取IP数量
    const getIpNumber = (row: ISubnet) => {
      const zones = zoneCounts.value[row.zone];
      const availableIPAddressCounts = zones.reduce((pre, item) => {
        pre += Number(item.availableIPAddressCount);
        return pre;
      }, 0);
      const cidrs = zones.reduce((pre, item) => {
        pre += Number(countIPsInCIDR(item.cidrRange));
        return pre;
      }, 0);
      return `${availableIPAddressCounts}/${cidrs}`;
    };

    // 添加子网
    const showSubnets = ref(false);
    const newSubnets = ref([{
      ipCnt: 256,
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
        await getClusterDetail(props.clusterId, true);
        await handleGetSubnets();
        showSubnets.value = false;
      }
      pending.value = false;
    };

    // 合并单元格
    const firstRowspan = ref({});
    const objectSpanMethod = ({ row, column }) => {
      if (column.property === 'zoneName' || column.property === 'counts') {
        if (zoneCounts.value[row.zone]?.length === 1) return {
          rowspan: 1,
          colspan: 1,
        };
        // 合并相同可用区
        if (zoneCounts.value[row.zone]?.length > 1 && !firstRowspan.value[`${column.property}-${row.zone}`]) {
          firstRowspan.value[`${column.property}-${row.zone}`] = true;
          return {
            rowspan: zoneCounts.value[row.zone]?.length,
            colspan: 1,
          };
        }
        return {
          rowspan: 0,
          colspan: 1,
        };
      }
    };

    const {
      regionLoading,
      regionList,
      handleGetRegionList,
      vpcList,
      vpcLoading,
      handleGetVPCList,
    } = useCloud();

    const region = computed(() => {
      const data = regionList.value.find(item => item.region === clusterData.value.region);
      return data ? `${data.regionName}(${data.region})` : clusterData.value.region;
    });
    const vpc = computed(() => {
      const data = vpcList.value.find(item => item.vpcId === clusterData.value.vpcID);
      return data ? `${data.name}(${data.vpcId})` : (clusterData.value.vpcID || '--');
    });

    onMounted(async () => {
      await getClusterDetail(props.clusterId, true);
      await Promise.all([
        handleGetSubnets(),
        handleGetRegionList({
          cloudAccountID: clusterData.value.cloudAccountID,
          cloudID: clusterData.value.provider,
        }),
        handleGetVPCList({
          region: clusterData.value.region,
          cloudAccountID: clusterData.value.cloudAccountID,
          cloudID: clusterData.value.provider,
          resourceGroupName: clusterData.value.extraInfo?.nodeResourceGroup,
        }),
      ]);
    });

    return {
      pending,
      isSubnetsValidate,
      showSubnets,
      newSubnets,
      subnetLoading,
      subnets,
      isLoading,
      clusterData,
      clusterIPv4CIDR,
      clusterIPv4CIDRCount,
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
      region,
      regionLoading,
      regionList,
      handleGetRegionList,
      vpc,
      vpcList,
      vpcLoading,
      handleGetVPCList,
      objectSpanMethod,
      getIpNumber,
    };
  },
});
</script>
<style lang="postcss" scoped>
@import './form.css';

>>> .bk-table .cell {
  padding-left: 15px !important;
}
</style>
