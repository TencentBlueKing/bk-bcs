<template>
  <div>
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
      <bk-table-column :label="$t('generic.ipSelector.label.innerIp')" prop="ip"></bk-table-column>
      <bk-table-column :label="$t('generic.ipSelector.label.agentStatus')" width="100">
        <template #default="{ row }">
          <StatusIcon :status="String(row.agent_alive)" :status-color-map="statusColorMap">
            {{row.agent_alive ? $t('generic.status.ready') : $t('generic.status.error')}}
          </StatusIcon>
        </template>
      </bk-table-column>
      <bk-table-column :label="$t('generic.label.action')" width="100">
        <template #default="{ row }">
          <bk-button
            text
            :disabled="disabled"
            @click="handleDeleteIp(row)">{{ $t('cluster.create.button.remove') }}</bk-button>
        </template>
      </bk-table-column>
      <template #empty>
        <BcsEmptyTableStatus type="empty" />
      </template>
    </bk-table>
    <IpSelector
      :show-dialog="showIpSelector"
      :ip-list="list"
      :disabled-ip-list="disabledIpList"
      :cloud-id="cloudId"
      :region="region"
      :vpc="vpc"
      :validate-vpc-and-region="validateVpcAndRegion"
      :account-i-d="accountID"
      :available-zone-list="availableZoneList"
      @confirm="handleChooseServer"
      @cancel="showIpSelector = false" />
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
    // 是否校验集群vpc和区域
    validateVpcAndRegion: {
      type: Boolean,
      default: true,
    },
    accountID: {
      type: String,
      default: '',
    },
    availableZoneList: {
      type: Array,
      default: () => [],
    },
  },
  setup(props, ctx) {
    const tableKey = ref('');
    const statusColorMap = ref({
      0: 'red',
      1: 'green',
    });

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
      showIpSelector,
      handleChooseServer,
      handleDeleteIp,
      getRegionName,
    };
  },
});
</script>
