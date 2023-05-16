<template>
  <div>
    <bk-form-item :label="$t('资源')" class="mb-[15px]">
      <bk-radio-group :disabled="disabled" v-model="type">
        <!-- todo 暂时不支持申请节点资源模式 -->
        <bk-radio disabled value="newNodes">
          <span v-bk-tooltips="$t('功能正在开发中')">{{ $t('申请节点资源') }}</span>
        </bk-radio>
        <bk-radio :disabled="disabled" value="existNodes">
          {{ $t('使用已有服务器') }}
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
          {{ $t('添加服务器') }}
        </bk-button>
        <bk-table :data="ipList" class="mt-[16px]" v-bkloading="{ isLoading }">
          <bk-table-column type="index" :label="$t('序列')" width="60"></bk-table-column>
          <bk-table-column :label="$t('实例ID')">
            <template #default="{ row }">
              {{ row.nodeID || '--' }}
            </template>
          </bk-table-column>
          <bk-table-column :label="$t('内网IP')" prop="bk_host_innerip"></bk-table-column>
          <bk-table-column :label="$t('地域')">
            <template #default="{ row }">
              {{ getRegionName(row.region) || '--' }}
              <span
                class="text-[#ea3636] text-[16px]"
                v-bk-tooltips="$t('节点地域需要与集群地域保持一致，节点地域：{region1}，集群地域：{region2}', {
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
                v-bk-tooltips="$t('节点VPC需要与集群VPC保存一致，节点VPC：{vpc1}，集群VPC：{vpc2}', {
                  vpc1: row.vpc || '--',
                  vpc2: vpc.vpcID || '--'
                })"
                v-if="row.vpc !== vpc.vpcID">
                <i class="bk-icon icon-exclamation-circle-shape"></i>
              </span>
            </template>
          </bk-table-column>
          <bk-table-column :label="$t('可用区')">
            <template #default="{ row }">
              {{ row.zoneName || '--' }}
            </template>
          </bk-table-column>
          <bk-table-column :label="$t('机房')" prop="idc_name"></bk-table-column>
          <bk-table-column :label="$t('机型')" prop="svr_device_class"></bk-table-column>
          <bk-table-column :label="$t('操作')" width="100">
            <template #default="{ row }">
              <bk-button text :disabled="disabled" @click="handleDeleteIp(row)">{{ $t('移除') }}</bk-button>
            </template>
          </bk-table-column>
        </bk-table>
        <IpSelector
          v-model="showIpSelector"
          :ip-list="list"
          :disabled-ip-list="disabledIpList"
          :cloud-id="cloudId"
          :region="region"
          :vpc="vpc"
          @confirm="handleChooseServer" />
      </template>
    </bk-form-item>
  </div>
</template>
<script lang="ts">
import { defineComponent, ref, PropType } from 'vue';
import IpSelector from '@/components/ip-selector/selector-dialog.vue';

export default defineComponent({
  name: 'ApplyHost',
  components: { IpSelector },
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
      const index = ipList.value.findIndex(item => item.bk_host_innerip === row.bk_host_innerip);
      index > -1 && ipList.value.splice(index, 1);
      ctx.emit('change', ipList.value);
    };

    const getRegionName = region => props.regionList.find(item => item.region === region)?.regionName;

    return {
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
