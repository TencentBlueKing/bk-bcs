<template>
  <div>
    <bk-form-item :label="$t('cluster.create.label.hostResource')" class="mb-[15px]">
      <bk-radio-group :disabled="disabled" v-model="type">
        <!-- todo 暂时不支持申请节点资源模式 -->
        <bk-radio disabled value="newNodes">
          <span v-bk-tooltips="$t('generic.msg.info.development')">{{ $t('cluster.create.label.applyResource') }}</span>
        </bk-radio>
        <bk-radio :disabled="disabled" value="existNodes">
          {{ $t('cluster.create.label.useExitHost') }}
        </bk-radio>
      </bk-radio-group>
    </bk-form-item>
    <!-- <bcs-divider></bcs-divider> -->
    <bk-form-item>
      <template v-if="type === 'newNodes'">
      <!-- 申请资源 -->
      </template>
      <template v-else-if="type === 'existNodes'">
        <bk-button
          theme="primary"
          outline
          :disabled="disabled"
          icon="plus"
          @click="showIpSelector = true">
          {{ $t('cluster.create.button.addHost') }}
        </bk-button>
        <bk-table :data="ipList" :key="tableKey" class="mt-[16px]" v-bkloading="{ isLoading }">
          <bk-table-column type="index" :label="$t('cluster.create.label.index')" width="60"></bk-table-column>
          <bk-table-column :label="$t('cluster.create.label.instanceID')">
            <template #default="{ row }">
              {{ row.nodeID || '--' }}
            </template>
          </bk-table-column>
          <bk-table-column :label="$t('generic.ipSelector.label.innerIp')" prop="ip"></bk-table-column>
          <bcs-table-column :label="$t('generic.ipSelector.label.agentStatus')" width="100">
            <template #default="{ row }">
              <StatusIcon :status="String(row.agent_alive)" :status-color-map="statusColorMap">
                {{row.agent_alive ? $t('generic.status.ready') : $t('generic.status.error')}}
              </StatusIcon>
            </template>
          </bcs-table-column>
          <bk-table-column :label="$t('cluster.labels.region')">
            <template #default="{ row }">
              {{ getRegionName(row.region) || '--' }}
              <span
                class="text-[#ea3636] text-[16px]"
                v-bk-tooltips="$t('cluster.create.validate.regionDiff', {
                  region1: getRegionName(row.region) || '--',
                  region2: getRegionName(region) || '--'
                })"
                v-if="region !== row.region">
                <i class="bk-icon icon-exclamation-circle-shape"></i>
              </span>
            </template>
          </bk-table-column>
          <bk-table-column label="VPC">
            <template #default="{ row }">
              {{ row.vpc || '--' }}
              <span
                class="text-[#ea3636] text-[16px]"
                v-bk-tooltips="$t('cluster.create.validate.vpcDiff', {
                  vpc1: row.vpc || '--',
                  vpc2: vpc.vpcID || '--'
                })"
                v-if="row.vpc !== vpc.vpcID">
                <i class="bk-icon icon-exclamation-circle-shape"></i>
              </span>
            </template>
          </bk-table-column>
          <bk-table-column :label="$t('cluster.ca.nodePool.create.az.title')">
            <template #default="{ row }">
              {{ row.zoneName || '--' }}
            </template>
          </bk-table-column>
          <!-- <bk-table-column :label="$t('generic.ipSelector.label.idc')" prop="idc_name"></bk-table-column> -->
          <bk-table-column
            :label="$t('generic.ipSelector.label.serverModel')" prop="instanceType"></bk-table-column>
          <bk-table-column :label="$t('generic.label.action')" width="100">
            <template #default="{ row }">
              <bk-button
                text
                :disabled="disabled"
                @click="handleDeleteIp(row)">{{ $t('cluster.create.button.remove') }}</bk-button>
            </template>
          </bk-table-column>
        </bk-table>
        <IpSelector
          :show-dialog="showIpSelector"
          :ip-list="list"
          :disabled-ip-list="disabledIpList"
          :cloud-id="cloudId"
          :region="region"
          :vpc="vpc"
          @confirm="handleChooseServer"
          @cancel="showIpSelector = false" />
      </template>
    </bk-form-item>
  </div>
</template>
<script lang="ts">
import { defineComponent, PropType, ref } from 'vue';

import IpSelector from '@/components/ip-selector/ip-selector.vue';
import StatusIcon from '@/components/status-icon';

export default defineComponent({
  name: 'ApplyHost',
  components: { IpSelector, StatusIcon },
  model: {
    prop: 'list',
    event: 'change',
  },
  props: {
    list: {
      type: Array,
      default: () => [],
    },
    disabled: {
      type: Boolean,
      default: false,
    },
    // 集群地域
    region: {
      type: String,
      default: '',
    },
    cloudId: {
      type: String,
      default: '',
    },
    disabledIpList: {
      type: Array as PropType<Array<string|{ip: string, tips: string}>>,
      default: () => [],
    },
    regionList: {
      type: Array as PropType<Array<any>>,
      default: () => [],
    },
    // 集群VPC
    vpc: {
      type: Object,
      default: () => ({}),
    },
  },
  setup(props, ctx) {
    const tableKey = ref('');
    const statusColorMap = ref({
      0: 'red',
      1: 'green',
    });
    const type = ref<'newNodes'|'existNodes'>('existNodes');

    const showIpSelector = ref(false);
    const isLoading = ref(false);
    const ipList = ref<any[]>(props.list);

    const handleChooseServer = async (data) => {
      showIpSelector.value = false;
      ipList.value = data;
      ctx.emit('change', ipList.value);
    };

    const handleDeleteIp = (row) => {
      const index = ipList.value.findIndex(item => item?.cloudArea?.id === row?.cloudArea?.id && item.ip === row.ip);
      index > -1 && ipList.value.splice(index, 1);
      tableKey.value = `${Math.random() * 10}`;
      ctx.emit('change', ipList.value);
    };

    const getRegionName = region => props.regionList.find(item => item.region === region)?.regionName;

    return {
      statusColorMap,
      tableKey,
      ipList,
      isLoading,
      type,
      showIpSelector,
      handleChooseServer,
      handleDeleteIp,
      getRegionName,
    };
  },
});
</script>
