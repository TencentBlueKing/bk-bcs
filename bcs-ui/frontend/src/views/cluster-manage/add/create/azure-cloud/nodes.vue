<template>
  <div>
    <bk-form :model="nodeConfig" :rules="nodeConfigRules" ref="formRef">
      <bk-form-item
        :label="$t('cluster.create.aws.label.defaultNodePool')"
        required
        error-display-type="normal"
        property="nodePoolGroups"
      >
        <!-- <bcs-button
          theme="primary"
          outline
          icon="plus"
          @click="handleAddNode">
          {{$t('cluster.create.aws.nodePool')}}
        </bcs-button>
        <bcs-button
          theme="primary"
          outline
          icon="delete"
          @click="handleDeleteNode">
          {{$t('generic.button.delete')}}
        </bcs-button> -->
        <bcs-table
          class="w-[650px]"
          height="200px"
          :ref="el => tableRef = el"
          v-bkloading="{ isLoading: loading }"
          :row-class-name="getNodePoolRowClassName"
          :data="nodePoolTableData"
          @selection-change="handleSelectionChange">
          <!-- <bcs-table-column type="selection" width="60" :selectable="handleSelectable">
            <template #default="{ row }">
              <bcs-checkbox
                v-bk-tooltips="{
                  content: $t('cluster.create.aws.tips.subnetEmpty'),
                  disabled: handleSelectable(row)
                }"
                :checked="!!selectionList.find(item => item === row)"
                :disabled="!handleSelectable(row)"
                @change="handleRowSelectionChange(row)">
              </bcs-checkbox>
            </template>
          </bcs-table-column> -->
          <bcs-table-column
            :label="$t('cluster.ca.nodePool.label.name')"
            prop="name"
            width="150">
          </bcs-table-column>
          <bcs-table-column
            :label="$t('cluster.ca.nodePool.label.nodeQuota')"
            prop="autoScaling.maxSize"
            width="100">
          </bcs-table-column>
          <bcs-table-column
            :label="$t('cluster.ca.nodePool.label.nodeCounts')"
            prop="autoScaling.desiredSize"
            width="100">
          </bcs-table-column>
          <bcs-table-column
            :label="$t('generic.ipSelector.label.serverModel')"
            prop="nodeOS">
          </bcs-table-column>
          <bcs-table-column
            :label="$t('generic.label.action')"
            prop="action">
            <template #default="{ $index, row }">
              <bcs-button text @click="handleEdit(row, $index)">{{ $t('generic.button.edit') }}</bcs-button>
              <!-- <bcs-button text class="ml-[10px]" @click="handleDelete($index)">
                {{ $t('generic.button.delete') }}
              </bcs-button> -->
            </template>
          </bcs-table-column>
        </bcs-table>
      </bk-form-item>
      <!-- 用户名和密码 -->
      <bk-form-item
        :label="$t('tke.label.loginType.text')"
        :desc="$t('tke.label.loginType.desc')"
        required
        error-display-type="normal">
        <LoginTypeWithName
          class="max-w-[650px]"
          :ref="el => loginTypeRef = el"
          :value="nodeConfig.nodeSettings.workerLogin"
          @change="handleGetLoginMessage" />
      </bk-form-item>
      <bk-form-item
        :label="$t('cluster.create.aws.securityGroup.text')"
        :desc="{
          allowHTML: true,
          content: '#secPlugin1',
        }"
        error-display-type="normal"
        required
        property="networkSettings.securityGroupIDs">
        <SecurityGroups
          class="max-w-[650px]"
          multiple
          :region="region"
          :show-extension="false"
          :cloud-account-i-d="cloudAccountID"
          :cloud-i-d="cloudID"
          :resource-group-name="resourceGroupName"
          v-model="nodeConfig.nodeSettings.securityGroupIDs" />
        <div id="secPlugin1">
          <div>
            <i18n path="cluster.create.azure.securityGroup.vpcLink">
              <span
                class="text-[12px] text-[#699DF4] cursor-pointer"
                @click="openLink('https://portal.azure.com/#browse/Microsoft.Network%2FNetworkSecurityGroups')">
                {{ $t('cluster.create.azure.securityGroup.vpcMaster') }}
              </span>
            </i18n>
          </div>
        </div>
      </bk-form-item>
      <div class="flex items-center h-[48px] bg-[#FAFBFD] px-[24px] fixed bottom-0 left-0 w-full bcs-border-top">
        <bk-button class="min-w-[88px]" @click="preStep">{{ $t('generic.button.pre') }}</bk-button>
        <bk-button
          theme="primary"
          class="ml10 min-w-[88px]"
          @click="handleConfirm">{{ $t('generic.button.confirm') }}</bk-button>
        <bk-button class="ml10" @click="handleCancel">{{ $t('generic.button.cancel') }}</bk-button>
      </div>
    </bk-form>
    <bcs-sideslider
      :is-show.sync="showNodeConfig"
      :quick-close="false"
      :title="model === 'add' ? $t('cluster.create.aws.nodePool') : $t('cluster.create.aws.editNodePool')"
      :width="740">
      <template #content>
        <keep-alive>
          <component
            :is="stepComMap[curStep]"
            :save-loading="saveLoading"
            :region="region"
            :cloud-account-i-d="cloudAccountID"
            :cloud-i-d="cloudID"
            :default-values="nodeDefaultValues"
            :vpc-i-d="vpcID"
            :bk-cloud-i-d="bkCloudID"
            :resource-group-name="resourceGroupName"
            v-if="!isLoading"
            @next="handleNextStep"
            @pre="handlePreStep"
            @add="handleAddnode"
            @close="closeSlider"
          />
        </keep-alive>
      </template>
    </bcs-sideslider>
  </div>
</template>
<script setup lang="ts">
import { cloneDeep } from 'lodash';
import { computed, onMounted, PropType, ref, watch } from 'vue';

import LoginTypeWithName from '../../components/login-type-with-name.vue';

import NodePoolConfig from './node-pool-config.vue';
import NodePoolInfo from './node-pool-info.vue';

import { recommendNodeGroupConf } from '@/api/modules/cluster-manager';
import { mergeDeep } from '@/common/util';
import { useFocusOnErrorField } from '@/composables/use-focus-on-error-field';
import $i18n from '@/i18n/i18n-setup';
import $store from '@/store/index';
import SecurityGroups from '@/views/cluster-manage/add/components/security-groups.vue';

const props = defineProps({
  region: {
    type: String,
    default: '',
  },
  cloudAccountID: {
    type: String,
    default: '',
  },
  cloudID: {
    type: String,
    default: '',
  },
  environment: {
    type: String,
    default: '',
  },
  provider: {
    type: String,
    default: '',
  },
  vpcID: {
    type: String,
    default: '',
  },
  master: {
    type: Array as PropType<string[]>,
    default: () => [],
  },
  masterLogin: {
    type: Object,
    default: () => ({
      initLoginUsername: '',
      initLoginPassword: '',
      keyPair: {
        keyID: '',
        keySecret: '',
        keyPublic: '',
      },
    }),
  },
  manageType: {
    type: String as PropType<'MANAGED_CLUSTER'|'INDEPENDENT_CLUSTER'>,
    default: 'MANAGED_CLUSTER',
  },
  bkCloudID: {
    type: Number,
    default: 9,
  },
  resourceGroupName: {
    type: String,
    default: '',
  },
});

const emits = defineEmits(['next', 'cancel', 'pre', 'confirm', 'instances-change']);
// node配置
const nodeConfig = ref<any>({
  nodeGroups: [],
  nodeSettings: {
    workerLogin: {
      initLoginUsername: 'azureuser',
      initLoginPassword: '',
      confirmPassword: '',
      keyPair: {
        keySecret: '',
        keyPublic: '',
      },
    },
    securityGroupIDs: [],
  },
});

// 动态 i18n 问题，这里使用computed
const nodeConfigRules = computed(() => ({
  'networkSettings.securityGroupIDs': [
    {
      trigger: 'custom',
      message: $i18n.t('generic.validate.required'),
      validator() {
        return nodeConfig.value.nodeSettings.securityGroupIDs.length > 0;
      },
    },
  ],
  nodePoolGroups: [
    {
      trigger: 'custom',
      message: $i18n.t('cluster.create.aws.validate.subnet'),
      validator() {
        const list = nodePoolTableData.value;
        if (list.length > 0) {
          return list.every(item => item?.autoScaling?.subnetIDs && item?.autoScaling?.subnetIDs.length > 0);
        }
        return true;
      },
    },
  ],
}));

// 跳转链接
const openLink = (link: string) => {
  if (!link) return;

  window.open(link);
};

const curIndex = ref();
const model = ref<'add' | 'edit'>('add');
const nodePoolTableData = ref<any[]>([]);
const selectionList = ref<any[]>([]);
// 点击添加节点
function handleAddNode() {
  nodeDefaultValues.value = cloneDeep(defaultValues.value);
  model.value = 'add';
  showNodeConfig.value = true;
};
// 删除节点按钮
function handleDeleteNode() {
  if (selectionList.value.length === 0) return;
  nodePoolTableData.value = nodePoolTableData.value.reduce((pre, curItem) => {
    if (selectionList.value.findIndex(item => item.$id === curItem.$id) < 0) {
      pre.push(curItem);
    }
    return pre;
  }, []);
};
// 添加/编辑节点操作
function handleAddnode() {
  if (model.value === 'add') {
    nodePoolTableData.value = [...nodePoolTableData.value, nodePoolData.value];
  } else {
    nodePoolTableData.value.splice(curIndex.value, 1, nodePoolData.value);
  }
  nodePoolData.value = {};
}
// 删除节点操作
function handleDelete(index) {
  nodePoolTableData.value = nodePoolTableData.value.filter((_, i) => i !== index);
}
// 编辑节点
function handleEdit(row, index) {
  formRef.value?.clearError();
  loginTypeRef.value?.clearError();
  curIndex.value = index;
  nodeDefaultValues.value = cloneDeep(mergeDeep(nodeDefaultValues.value, row));
  model.value = 'edit';
  showNodeConfig.value = true;
}
// 选择节点
function handleSelectionChange(selection) {
  selectionList.value = selection;
}


// 选择单个节点
const tableRef = ref();
function handleRowSelectionChange(row) {
  const index = selectionList.value.findIndex(item => item.$id === row.$id);
  if (index > -1) {
    selectionList.value.splice(index, 1);
    // 更新表格原本选中状态
    tableRef.value?.toggleRowSelection(row, false);
  } else {
    selectionList.value.push(row);
    tableRef.value?.toggleRowSelection(row, true);
  }
};
// 关闭侧滑框
function closeSlider() {
  showNodeConfig.value = false;
  curStep.value = 1;
  nodeDefaultValues.value = cloneDeep(defaultValues.value);
}

const curStep = ref(1);
const stepComMap = {
  1: NodePoolConfig,
  2: NodePoolInfo,
};
const nodePoolData = ref<Record<string, any>>({});
const handleNextStep = (data) => {
  nodePoolData.value = mergeDeep(nodePoolData.value, data);
  if (curStep.value + 1 <= 2) {
    curStep.value = curStep.value + 1;
  }
};
const handlePreStep = () => {
  curStep.value = curStep.value - 1;
};

const isLoading = ref(false);
const saveLoading = ref(false);

// 节点池默认配置
const defaultValues = ref({
  autoScaling: {
    maxSize: 10,
    minSize: 0,
    multiZoneSubnetPolicy: 'PRIORITY',
    retryPolicy: 'IMMEDIATE_RETRY',
    scalingMode: 'CLASSIC_SCALING',
  },
  clusterID: '',
  creator: '',
  enableAutoscale: true,
  launchTemplate: {
    CPU: 4,
    Mem: 8,
    dataDisks: [],
    imageInfo: {
      imageID: 'img-eb30mz89',
    },
    internetAccess: {
      internetChargeType: 'TRAFFIC_POSTPAID_BY_HOUR',
      internetMaxBandwidth: '0',
      publicIPAssigned: false,
    },
    systemDisk: {
      diskSize: '100',
      diskType: 'DISK_TYPE_ZFS',
    },
    keyPair: {
      keyPublic: '',
      keySecret: '',
    },
  },
  name: '',
  nodeTemplate: {
    dockerGraphPath: '/data/bcs/service/docker',
    taints: [],
    unSchedulable: 0,
  },
  region: '',
});
const nodeDefaultValues = ref();
const loading = ref(false);
// 获取节点池默认数据
async function getNodeGroupData() {
  loading.value = true;
  const result = await recommendNodeGroupConf({
    $cloudId: props.cloudID,
    region: props.region,
    accountID: props.cloudAccountID,
  }).catch(() => []);
  nodePoolTableData.value = result.map((item, index) => setNodeInfo(item, index));
  loading.value = false;
}

const user = computed(() => $store.state.user);
function setNodeInfo(data, index) {
  return {
    $id: `${index}id_${Date.now()}`,
    status: 'CREATING',
    creator: user.value.username,
    name: data?.name || '', // 节点名称
    nodeOS: data?.instanceProfile?.nodeOS || '', // 节点操作系统
    autoScaling: {
      serviceRole: data?.serviceRoleName || '', // iam角色
      subnetIDs: data?.networkProfile?.subnetIDs || [], // 支持子网
      maxSize: data?.scalingProfile?.maxSize,
      desiredSize: data?.scalingProfile?.desiredSize || 0,
    },
    launchTemplate: {
      CPU: data?.hardwareProfile?.CPU || '',
      GPU: data?.hardwareProfile?.GPU || '',
      Mem: data?.hardwareProfile?.Mem || '',
      instanceType: data?.instanceProfile?.instanceType, // 机型信息
      InstanceChargeType: data?.instanceProfile?.instanceChargeType, // 实例计费方式
      systemDisk: {
        diskType: data?.hardwareProfile?.systemDisk?.diskType,
        diskSize: data?.hardwareProfile?.systemDisk?.diskSize, // 系统盘大小
      },
      internetAccess: {
        publicIPAssigned: data?.networkProfile?.publicIPAssigned, // 分配免费公网IP
      },
      // 密钥信息
      keyPair: {
        keyPublic: '',
        keySecret: '',
      },
      initLoginUsername: '', // 用户名
      securityGroupIDs: [], // 安全组
      dataDisks: data?.hardwareProfile?.dataDisks || [], // 数据盘
      // 默认值
      isSecurityService: true,
      isMonitorService: true,
    },
    nodeTemplate: {
      dataDisks: data?.hardwareProfile?.dataDisks || [],
    },
  };
}

// 添加、编辑节点池
const showNodeConfig = ref(false);

// 设置行背景颜色
function getNodePoolRowClassName({ row }) {
  return row?.autoScaling?.subnetIDs && row?.autoScaling?.subnetIDs.length > 0 ? '' : '!bg-red-50';
};
// 设置是否可选
function handleSelectable(row) {
  return row?.autoScaling?.subnetIDs && row?.autoScaling?.subnetIDs.length > 0;
};

watch([
  () => props.region,
  () => props.cloudAccountID,
], () => {
  getNodeGroupData();
});

watch(() => nodeConfig.value.nodeSettings.workerLogin, () => {
  defaultValues.value.launchTemplate.keyPair = nodeConfig.value.nodeSettings?.workerLogin?.keyPair;
  nodePoolTableData.value.forEach((node) => {
    !node?.launchTemplate && (node.launchTemplate = {});
    !node?.launchTemplate?.keyPair && (node.launchTemplate.keyPair = {});
    node.launchTemplate.keyPair.keyPublic = nodeConfig.value.nodeSettings?.workerLogin?.keyPair?.keyPublic;
    node.launchTemplate.keyPair.keySecret = nodeConfig.value.nodeSettings?.workerLogin?.keyPair?.keySecret;
  });
}, { deep: true });

// 校验表单
const formRef = ref();
const loginTypeRef = ref();
const validate = async () => {
  const result = await formRef.value?.validate().catch(() => false);
  const result1 = await loginTypeRef.value?.validate().catch(() => false);
  return result && result1;
};

function handleGetLoginMessage(data) {
  nodeConfig.value.nodeSettings.workerLogin = data;
};

// 上一步
const preStep = () => {
  emits('pre');
};
// 下一步
const { focusOnErrorField } = useFocusOnErrorField();
const handleConfirm = async () => {
  const result = await validate();
  if (result) {
    nodeConfig.value.nodeGroups = nodePoolTableData.value;
    nodeConfig.value.nodeGroups.forEach((item, index) => {
      item.autoScaling.vpcID = props.vpcID;
      item.bkCloudID = props.bkCloudID;
      item.launchTemplate.securityGroupIDs = [...nodeConfig.value.nodeSettings.securityGroupIDs];
      item.launchTemplate = mergeDeep(item.launchTemplate, nodeConfig.value.nodeSettings.workerLogin);
      item.nodeGroupType = index === 0 ? 'System' : 'User';
    });
    emits('next', {
      ...nodeConfig.value,
    });
    emits('confirm');
  } else {
    // 自动滚动到第一个错误的位置
    focusOnErrorField();
  }
};
// 取消
const handleCancel = () => {
  emits('cancel');
};

onMounted(async () => {
  isLoading.value = true;
  nodeDefaultValues.value = cloneDeep(defaultValues.value);
  isLoading.value = false;
});

defineExpose({
  validate,
});
</script>
<style lang="postcss" scoped>
/deep/ .bk-table-header, /deep/ .bk-table-body {
  width: 100% !important;
}
</style>
