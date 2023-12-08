<!-- eslint-disable max-len -->
<template>
  <div>
    <Header :title="$t('cluster.button.addCluster')" :desc="$t('generic.title.createPublicCluster')" />
    <div class="pt-[8px] bg-[#f0f1f5]">
      <bcs-tab
        :label-height="42"
        :validate-active="false"
        :active.sync="activeTabName"
        type="card-tab">
        <!-- 基本信息 -->
        <bcs-tab-panel :name="steps[0].name">
          <template #label>
            <StepTabLabel :title="$t('generic.title.basicInfo1')" :step-num="1" :active="activeTabName === steps[0].name" />
          </template>
          <BasicInfo
            :cloud-id="cloudID"
            @set-image-group="setImageGroup"
            @set-region-list="setRegionList"
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
              :disabled="steps[1].disabled" />
          </template>
          <Network
            :region="clusterData.region"
            :cloud-account-i-d="clusterData.cloudAccountID"
            :cloud-i-d="cloudID"
            :region-list="regionList"
            @set-vpc-list="setVpcList"
            @set-zone-list="setZoneList"
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
              :disabled="steps[2].disabled" />
          </template>
          <Master
            :region="clusterData.region"
            :cloud-account-i-d="clusterData.cloudAccountID"
            :cloud-i-d="cloudID"
            :environment="clusterData.environment"
            :provider="clusterData.provider"
            :vpc-i-d="clusterData.vpcID"
            :nodes="clusterData.nodes"
            :region-list="regionList"
            :vpc-list="vpcList"
            :image-list-by-group="imageListByGroup"
            :zone-list="zoneList"
            :os="clusterData.clusterBasicSettings ? clusterData.clusterBasicSettings.OS : ''"
            @pre="preStep"
            @next="nextStep"
            @cancel="handleCancel"
            @instances-change="handleMasterInstanceChange" />
        </bcs-tab-panel>
        <!-- 添加节点 -->
        <bcs-tab-panel :name="steps[3].name" :disabled="steps[3].disabled">
          <template #label>
            <StepTabLabel
              :title="$t('cluster.nodeList.create.text')"
              :step-num="4"
              :active="activeTabName === steps[3].name"
              :disabled="steps[3].disabled" />
          </template>
          <Nodes
            :region="clusterData.region"
            :cloud-account-i-d="clusterData.cloudAccountID"
            :cloud-i-d="cloudID"
            :environment="clusterData.environment"
            :provider="clusterData.provider"
            :vpc-i-d="clusterData.vpcID"
            :master="clusterData.master"
            :region-list="regionList"
            :vpc-list="vpcList"
            :image-list-by-group="imageListByGroup"
            :zone-list="zoneList"
            :os="clusterData.clusterBasicSettings ? clusterData.clusterBasicSettings.OS : ''"
            :master-login="clusterData.nodeSettings ? clusterData.nodeSettings.masterLogin : undefined"
            :auto-generate-master-nodes="clusterData.autoGenerateMasterNodes"
            @pre="preStep"
            @next="nextStep"
            @cancel="handleCancel"
            @confirm="handleShowConfirmDialog"
            @instances-change="handleNodesInstanceChange" />
        </bcs-tab-panel>
      </bcs-tab>
    </div>
  </div>
</template>
<script lang="ts" setup>
import { merge } from 'lodash';
import { computed, provide, ref } from 'vue';

import StepTabLabel from '../common/step-tab-label.vue';

import BasicInfo from './basic.vue';
import Master from './master.vue';
import Network from './network.vue';
import Nodes from './nodes.vue';
import { ClusterDataInjectKey, DeepPartial, IClusterData, IInstanceItem  } from './types';

import { createCluster } from '@/api/modules/cluster-manager';
import $bkMessage from '@/common/bkmagic';
import $bkInfo from '@/components/bk-magic-2.0/bk-info';
import Header from '@/components/layout/Header.vue';
import { useProject } from '@/composables/use-app';
import $i18n from '@/i18n/i18n-setup';
import $router from '@/router';
import $store from '@/store';;

const cloudID = 'tencentPublicCloud';

const steps = ref([
  { name: 'basicInfo', disabled: false },
  { name: 'network',  disabled: true },
  { name: 'master',  disabled: true },
  { name: 'nodes', disabled: true },
]);

const activeTabName = ref<typeof steps.value[number]['name']>('basicInfo');

const clusterData = ref<DeepPartial<IClusterData>>({});

provide(ClusterDataInjectKey, clusterData);

const regionList = ref([]);
const setRegionList = (data) => {
  regionList.value = data;
};

const imageListByGroup = ref({});
const setImageGroup = (data) => {
  imageListByGroup.value = data;
};

const vpcList = ref([]);
const setVpcList = (data) => {
  vpcList.value = data;
};

const zoneList = ref([]);
const setZoneList = (data) => {
  zoneList.value = data;
};
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
  clusterData.value = merge({}, clusterData.value, data);
  const index = steps.value.findIndex(step => activeTabName.value === step.name);
  if (index > -1 && index + 1 < steps.value.length) {
    steps.value[index + 1].disabled = false;
    activeTabName.value = steps.value[index + 1]?.name;
  }
};
const handleCancel = () => {
  $router.back();
};
// 创建集群
const handleShowConfirmDialog = async () => {
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
  console.log(params);
  const result = await createCluster(params).then(() => true)
    .catch(() => false);
  if (result) {
    $bkMessage({
      theme: 'success',
      message: $i18n.t('generic.msg.success.deliveryTask'),
    });
    $router.push({ name: 'clusterMain' });
  }
};

</script>
<style lang="postcss" scoped>
>>> .bk-tab-header {
  padding: 0 8px;
}
>>> .bk-tab-content {
  height: calc(100vh - 224px);
  overflow: auto;
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
</style>
