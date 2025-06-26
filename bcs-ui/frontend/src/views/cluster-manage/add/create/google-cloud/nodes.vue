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
          class="w-[600px]"
          height="200px"
          :ref="el => tableRef = el"
          v-bkloading="{ isLoading: loading }"
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
            prop="name">
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
            prop="nodeTemplate.nodeOS">
          </bcs-table-column>
          <bcs-table-column
            :label="$t('generic.label.action')"
            prop="action"
            width="100">
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
        :label="$t('cluster.ca.nodePool.create.loginType.text')"
        property="nodeSettings.workerLogin"
        error-display-type="normal"
        required>
        <bk-radio-group value="ssh">
          <span class="inline-block">
            <bk-radio value="ssh">{{$t('cluster.ca.nodePool.create.loginType.ssh.text')}}</bk-radio>
          </span>
        </bk-radio-group>
        <div class="bg-[#F5F7FA] p-[16px] mt-[4px] max-w-[600px]">
          <bk-form-item
            :label-width="100"
            :label="$t('googleCloud.label.loginUser')"
            property="nodeConfig.launchTemplate.initLoginUsername"
            error-display-type="normal">
            <bcs-input v-model="nodeConfig.nodeSettings.initLoginUsername"></bcs-input>
          </bk-form-item>
          <bk-form-item
            :label="$t('cluster.ca.nodePool.create.loginType.ssh.label.publicKey.text')"
            :label-width="100"
            :desc="$t('cluster.ca.nodePool.create.loginType.ssh.label.publicKey.desc')"
            property="nodeConfig.launchTemplate.keyPair.keyPublic"
            error-display-type="normal">
            <bcs-input
              type="textarea"
              :rows="4"
              v-model="nodeConfig.nodeSettings.keyPublic">
            </bcs-input>
          </bk-form-item>
          <bk-form-item
            :label="$t('cluster.ca.nodePool.create.loginType.ssh.label.privateKey.text')"
            :label-width="100"
            :desc="$t('cluster.ca.nodePool.create.loginType.ssh.label.privateKey.desc')"
            property="nodePoolConfig.launchTemplate.keyPair.keySecret"
            error-display-type="normal">
            <bcs-input
              type="textarea"
              :rows="4"
              :placeholder="$t('generic.placeholder.input')"
              v-model="nodeConfig.nodeSettings.keySecret" />
          </bk-form-item>
        </div>
      </bk-form-item>
      <bk-form-item
        :label="$t('tke.label.nodeModule.text')"
        :desc="$t('tke.label.nodeModule.desc')"
        property="clusterBasicSettings.module.workerModuleID"
        error-display-type="normal"
        required>
        <TopoSelector
          :placeholder="$t('generic.placeholder.select')"
          v-model="nodeConfig.clusterBasicSettings.module.workerModuleID"
          class="max-w-[600px]" />
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

import NodePoolConfig from './node-pool-config.vue';
import NodePoolInfo from './node-pool-info.vue';

import { recommendNodeGroupConf } from '@/api/modules/cluster-manager';
import { mergeDeep } from '@/common/util';
import { useFocusOnErrorField } from '@/composables/use-focus-on-error-field';
import $i18n from '@/i18n/i18n-setup';
import $store from '@/store/index';
import TopoSelector from '@/views/cluster-manage/autoscaler/components/topo-select-tree.vue';

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
  manageType: {
    type: String as PropType<'MANAGED_CLUSTER'|'INDEPENDENT_CLUSTER'>,
    default: 'MANAGED_CLUSTER',
  },
  bkCloudID: {
    type: Number,
    default: 0,
  },
});

const emits = defineEmits(['next', 'cancel', 'pre', 'confirm', 'instances-change']);
// node配置
const nodeConfig = ref<any>({
  nodeGroups: [],
  clusterBasicSettings: {
    module: {
      workerModuleID: '',
    },
  },
  nodeSettings: {
    initLoginUsername: '',
    keyPublic: '',
    keySecret: '',
  },
});

// 动态 i18n 问题，这里使用computed
const nodeConfigRules = computed(() => ({
  'nodeSettings.workerLogin': [
    {
      trigger: 'custom',
      message: $i18n.t('generic.validate.required'),
      validator() {
        return nodeConfig.value.nodeSettings?.initLoginUsername
          && nodeConfig.value.nodeSettings?.keySecret
          && nodeConfig.value.nodeSettings?.keyPublic;
      },
    },
  ],
  nodePoolGroups: [
    {
      trigger: 'custom',
      message: $i18n.t('cluster.create.aws.validate.subnet'),
      validator() {
        // const list = nodePoolTableData.value;
        // if (list.length > 0) {
        //   return list.every(item => item?.autoScaling?.subnetIDs && item?.autoScaling?.subnetIDs.length > 0);
        // }
        return true;
      },
    },
  ],
  'clusterBasicSettings.module.workerModuleID': [
    {
      required: true,
      message: $i18n.t('generic.validate.required'),
      trigger: 'custom',
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

// 登录方式
const validateLogin = (trigger = '') => {
  formRef.value?.$refs?.loginTypeRef?.validate(trigger);
};
const handleLoginValueChange = (value) => {
  nodeConfig.value.nodeSettings.workerLogin = value;
};

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
      diskSize: '50',
      diskType: '',
    },
    keyPair: {
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
  if (!props.cloudID || !props.region || !props.cloudAccountID) return;
  loading.value = true;
  const result = await recommendNodeGroupConf({
    $cloudId: props.cloudID,
    region: props.region,
    accountID: props.cloudAccountID,
  }).catch(() => []);
  nodePoolTableData.value = result.map((item, index) => setNodeInfo(item, index));
  loading.value = false;
}
// 磁盘类型
const diskEnum = ref([
  {
    id: 'pd-balanced',
    name: $i18n.t('cluster.ca.nodePool.create.instanceTypeConfig.diskType.balanced'),
  },
  {
    id: 'pd-ssd',
    name: $i18n.t('cluster.ca.nodePool.create.instanceTypeConfig.diskType.ssd'),
  },
  {
    id: 'pd-standard',
    name: $i18n.t('cluster.ca.nodePool.create.instanceTypeConfig.diskType.standard'),
  },
]);
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
      subnetIDs: [], // 支持子网
      maxSize: data?.scalingProfile?.maxSize,
      scalingMode: data?.scalingProfile?.scalingMode || 'Delete', // 扩缩容模式, 释放模式改为Delete
      desiredSize: data?.scalingProfile?.desiredSize || 0,
      zones: data?.zones || [],
    },
    launchTemplate: {
      CPU: data?.hardwareProfile?.CPU || '',
      Mem: data?.hardwareProfile?.Mem || '',
      instanceType: data?.instanceProfile?.instanceType, // 机型信息
      InstanceChargeType: data?.instanceProfile?.instanceChargeType, // 实例计费方式
      systemDisk: {
        diskType: diskEnum.value.some(item => item.id === data?.hardwareProfile?.systemDisk?.diskType)
          ? data?.hardwareProfile?.systemDisk?.diskType
          : 'pd-standard', // 系统盘类型
        diskSize: data?.hardwareProfile?.systemDisk?.diskSize || '100', // 系统盘大小
      },
      internetAccess: {
        publicIPAssigned: data?.networkProfile?.publicIPAssigned, // 分配免费公网IP
      },
      // 密钥信息
      keyPair: {
        keySecret: '',
      },
      initLoginUsername: '', // 用户名
      dataDisks: data?.hardwareProfile?.dataDisks || [], // 数据盘
      // 默认值
      isSecurityService: true,
      isMonitorService: true,
      imageInfo: {
        imageName: 'UBUNTU_CONTAINERD',
        imageType: 'PUBLIC_IMAGE',
      },
    },
    nodeTemplate: {
      dataDisks: data?.hardwareProfile?.dataDisks || [],
      nodeOS: data?.instanceProfile?.nodeOS || '', // 节点操作系统,
      dockerGraphPath: '/data/bcs/service/docker',
    },
  };
}

// 添加、编辑节点池
const showNodeConfig = ref(false);

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

// 校验表单
const formRef = ref();
const validate = async () => {
  const result = await formRef.value?.validate().catch(() => false);
  return result;
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
    nodeConfig.value.nodeGroups.forEach((item) => {
      !item?.launchTemplate && (item.launchTemplate = {});
      !item?.launchTemplate?.keyPair && (item.launchTemplate.keyPair = {});
      item.launchTemplate.initLoginUsername = nodeConfig.value.nodeSettings?.initLoginUsername;
      item.launchTemplate.keyPair.keySecret = nodeConfig.value.nodeSettings?.keySecret;
      item.launchTemplate.keyPair.keyPublic = nodeConfig.value.nodeSettings?.keyPublic;
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
  getNodeGroupData();
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
