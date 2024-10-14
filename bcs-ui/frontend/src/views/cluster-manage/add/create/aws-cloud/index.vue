<!-- eslint-disable max-len -->
<template>
  <BcsContent
    :padding="0"
    class="add-aws-cluster"
    :title="$t('cluster.button.addCluster')"
    :desc="$t('generic.title.createPublicCluster')">
    <div class="h-full pt-[8px] bg-[#f0f1f5]">
      <bcs-tab
        :label-height="42"
        :validate-active="false"
        :active.sync="activeTabName"
        type="card-tab"
        class="h-full">
        <!-- 基本信息 -->
        <bcs-tab-panel :name="steps[0].name">
          <template #label>
            <StepTabLabel
              :title="$t('generic.title.basicInfo1')"
              :step-num="1"
              :active="activeTabName === steps[0].name"
              :is-error="steps[0].isErr" />
          </template>
          <BasicInfo
            :cloud-i-d="cloudID"
            ref="basicRef"
            @next="nextStep"
            @cancel="handleCancel" />
        </bcs-tab-panel>
        <!-- 网络配置 -->
        <bcs-tab-panel :name="steps[1].name" :disabled="steps[1].disabled">
          <template #label>
            <StepTabLabel
              :title="$t('cluster.detail.title.network')"
              :step-num="2"
              :active="activeTabName === steps[1].name"
              :disabled="steps[1].disabled"
              :is-error="steps[1].isErr" />
          </template>
          <Network
            :region="clusterData.region"
            :cloud-account-i-d="clusterData.cloudAccountID"
            :cloud-i-d="cloudID"
            ref="netRef"
            @pre="preStep"
            @next="nextStep"
            @cancel="handleCancel" />
        </bcs-tab-panel>
        <!-- Master配置 -->
        <bcs-tab-panel :name="steps[2].name" :disabled="steps[2].disabled">
          <template #label>
            <StepTabLabel
              :title="$t('cluster.detail.title.controlConfig')"
              :step-num="3"
              :active="activeTabName === steps[2].name"
              :disabled="steps[2].disabled"
              :is-error="steps[2].isErr" />
          </template>
          <Master
            :region="clusterData.region"
            :cloud-account-i-d="clusterData.cloudAccountID"
            :cloud-i-d="cloudID"
            :provider="clusterData.provider"
            :vpc-i-d="clusterData.vpcID"
            :nodes="clusterData.nodes"
            ref="masterRef"
            @pre="preStep"
            @next="nextStep"
            @cancel="handleCancel"
            @instances-change="handleMasterInstanceChange" />
        </bcs-tab-panel>
        <!-- 添加节点池 -->
        <bcs-tab-panel :name="steps[3].name" :disabled="steps[3].disabled">
          <template #label>
            <StepTabLabel
              :title="$t('cluster.create.aws.nodePool')"
              :step-num="4"
              :active="activeTabName === steps[3].name"
              :disabled="steps[3].disabled"
              :is-error="steps[3].isErr" />
          </template>
          <AddNodePools
            :region="clusterData.region"
            :cloud-account-i-d="clusterData.cloudAccountID"
            :cloud-i-d="cloudID"
            :environment="clusterData.environment"
            :provider="clusterData.provider"
            :vpc-i-d="clusterData.vpcID"
            :master="clusterData.master"
            :master-login="clusterData.nodeSettings ? clusterData.nodeSettings.masterLogin : undefined"
            :manage-type="clusterData.manageType"
            :auto-generate-master-nodes="clusterData.autoGenerateMasterNodes"
            :bk-cloud-i-d="clusterData?.clusterBasicSettings?.area?.bkCloudID"
            ref="nodesRef"
            @pre="preStep"
            @next="nextStep"
            @cancel="handleCancel"
            @confirm="handleShowConfirmDialog"
            @instances-change="handleNodesInstanceChange" />
        </bcs-tab-panel>
      </bcs-tab>
    </div>
  </BcsContent>
</template>
<script lang="ts" setup>
import { merge, mergeWith } from 'lodash';
import { computed, provide, ref, watch } from 'vue';

import { ClusterDataInjectKey, DeepPartial, IClusterData, IInstanceItem  } from '../../../types/types';

import AddNodePools from './add-node-pools.vue';
import BasicInfo from './basic.vue';
import Master from './master.vue';
import Network from './network.vue';

import { createCluster } from '@/api/modules/cluster-manager';
import $bkMessage from '@/common/bkmagic';
import $bkInfo from '@/components/bk-magic-2.0/bk-info';
import BcsContent from '@/components/layout/Content.vue';
import { useProject } from '@/composables/use-app';
import $i18n from '@/i18n/i18n-setup';
import $router from '@/router';
import $store from '@/store';
import StepTabLabel from '@/views/cluster-manage/add/components/step-tab-label.vue';

const cloudID: CloudID = 'awsCloud';

const steps = ref([
  { name: 'basicInfo', disabled: false, isErr: false },
  { name: 'network',  disabled: true, isErr: false },
  { name: 'master',  disabled: true, isErr: false },
  { name: 'nodes', disabled: true, isErr: false },
]);

const activeTabName = ref<typeof steps.value[number]['name']>('basicInfo');

const clusterData = ref<DeepPartial<IClusterData>>({});

provide(ClusterDataInjectKey, clusterData);

// tab切换
watch(activeTabName, () => {
  const tabItem = steps.value.find(item => item.name === activeTabName.value);
  if (tabItem) {
    tabItem.isErr = false;
  }
});

// 机型
const masterInstances = ref<IInstanceItem[]>([]);
const handleMasterInstanceChange = (data) => {
  masterInstances.value = data;
};
const nodeInstances = ref<IInstanceItem[]>([]);
const handleNodesInstanceChange = (data) => {
  nodeInstances.value = data;
};
// 上一步
const preStep = async () => {
  const index = steps.value.findIndex(step => activeTabName.value === step.name);
  if (index > -1 && index - 1 >= 0) {
    activeTabName.value = steps.value[index - 1]?.name;
  }
};
// 下一步
const nextStep = async (data = {}) => {
  clusterData.value = mergeWith({}, clusterData.value, data, (objValue, srcValue) => {
    if (Array.isArray(objValue)) {
      return srcValue;
    }
  });
  const index = steps.value.findIndex(step => activeTabName.value === step.name);
  if (index > -1 && index + 1 < steps.value.length) {
    steps.value[index + 1].disabled = false;
    activeTabName.value = steps.value[index + 1]?.name;
  }
};
const handleCancel = () => {
  $router.back();
};
// 校验表单
const basicRef = ref();
const netRef = ref();
const masterRef = ref();
const nodesRef = ref();
const validate = async () => {
  const list = [
    basicRef.value.validate().then((result) => {
      steps.value[0].isErr = !result;
      return result;
    }),
    netRef.value.validate().then((result) => {
      steps.value[1].isErr = !result;
      return result;
    }),
    masterRef.value.validate().then((result) => {
      steps.value[2].isErr = !result;
      return result;
    }),
    nodesRef.value.validate().then((result) => {
      steps.value[3].isErr = !result;
      return result;
    }),
  ];
  const data = await Promise.all(list);
  return data.every(result => !!result);
};
// 创建集群
const handleShowConfirmDialog = async () => {
  const result = await validate();
  if (!result) return;

  $bkInfo({
    type: 'warning',
    clsName: 'custom-info-confirm',
    title: $i18n.t('cluster.create.button.confirmCreateCluster.text'),
    defaultInfo: true,
    confirmFn: async () => {
      await handleCreateCluster();
    },
  });
};
const { curProject } = useProject();
const user = computed(() => $store.state.user);
const handleCreateCluster = async () => {
  const params = merge(
    {
      projectID: curProject.value.projectID,
      businessID: String(curProject.value.businessID),
      engineType: 'k8s',
      isExclusive: true,
      clusterType: 'single',
      creator: user.value.username,
    },
    {
      instances: [
        ...masterInstances.value,
        ...nodeInstances.value,
      ],
    },
    clusterData.value,
  );
  const result = await createCluster(params).catch(() => false);
  if (result) {
    $bkMessage({
      theme: 'success',
      message: $i18n.t('generic.msg.success.deliveryTask'),
    });
    $router.push({
      name: 'clusterMain',
      params: {
        highlightClusterId: result?.clusterID,
      },
    });
  }
};

</script>
<style lang="postcss" scoped>
>>> .bk-tab-header {
  padding: 0 8px;
}
>>> .bk-tab-section {
  overflow: auto;
  height: calc(100% - 80px);
}

>>> .k8s-form .bk-form-content {
  max-width: 600px;
  padding-right: 24px;
}
>>> .prefix {
  display: inline-block;
  height: 32px;
  line-height: 32px;
  background: #F0F1F5;
  border: 1px solid #C4C6CC;
  border-radius: 2px 0 0 2px;
  padding: 0 8px;
  font-size: 12px;
  &.disabled {
    border-color: #dcdee5;
  }
}

.add-aws-cluster {
  >>> .bk-form-item+.bk-form-item {
    margin-top: 24px;
  }
}
</style>
