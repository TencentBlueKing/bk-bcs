<!-- eslint-disable max-len -->
<template>
  <BcsContent :title="$t('cluster.button.addCluster')">
    <bcs-tab :label-height="42" :active.sync="activeTabName">
      <!-- 基本信息 -->
      <bcs-tab-panel :name="steps[0].name">
        <template #label>
          <StepTabLabel :title="$t('generic.title.basicInfo1')" :step-num="1" :active="activeTabName === steps[0].name" />
        </template>
        <bk-form :ref="steps[0].formRef" :model="basicInfo" :rules="basicInfoRules">
          <!-- <bk-form-item :label="$t('cluster.labels.clusterType')">
            {{ type || '--' }}
          </bk-form-item> -->
          <bk-form-item :label="$t('cluster.labels.name')" property="clusterName" error-display-type="normal" required>
            <bk-input
              :maxlength="64"
              :placeholder="$t('cluster.create.validate.name')"
              class="max-w-[600px]"
              v-model.trim="basicInfo.clusterName">
            </bk-input>
          </bk-form-item>
          <bk-form-item :label="$t('cluster.labels.env')" property="environment" error-display-type="normal" required>
            <bk-radio-group v-model="basicInfo.environment">
              <bk-radio value="stag" v-if="runEnv === 'dev'">
                UAT
              </bk-radio>
              <bk-radio :disabled="runEnv === 'dev'" value="debug">
                {{ $t('cluster.env.debug') }}
              </bk-radio>
              <bk-radio :disabled="runEnv === 'dev'" value="prod">
                {{ $t('cluster.env.prod') }}
              </bk-radio>
            </bk-radio-group>
          </bk-form-item>
          <bk-form-item :label="$t('cluster.create.label.clusterVersion')" property="clusterBasicSettings.version" error-display-type="normal" required>
            <bcs-select
              class="max-w-[600px]"
              :loading="templateLoading"
              v-model="basicInfo.clusterBasicSettings.version"
              searchable
              :clearable="false">
              <bcs-option v-for="item in versionList" :key="item" :id="item" :name="item"></bcs-option>
            </bcs-select>
          </bk-form-item>
          <bk-form-item :label="$t('cluster.create.label.desc')">
            <bk-input maxlength="100" class="max-w-[600px]" v-model="basicInfo.description" type="textarea"></bk-input>
          </bk-form-item>
        </bk-form>
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
        <bk-form :ref="steps[1].formRef" :model="networkConfig" :rules="networkConfigRules">
          <bk-form-item :label="$t('cluster.create.label.region')" property="region" error-display-type="normal" required>
            <bcs-select
              class="max-w-[600px]"
              v-model="networkConfig.region"
              :loading="regionLoading"
              searchable
              :clearable="false">
              <bcs-option
                v-for="item in regionList"
                :key="item.region" :id="item.region" :name="item.regionName"></bcs-option>
            </bcs-select>
          </bk-form-item>
          <bk-form-item :label="$t('cluster.create.label.networkMode.text')" property="networkType" error-display-type="normal" required>
            <bk-radio-group v-model="networkConfig.networkType">
              <bk-radio value="overlay">
                {{ $t('cluster.create.label.networkMode.overlay.text') }}
                <span
                  class="ml5 text-[#C4C6CC]"
                  v-bk-tooltips="{
                    content: $t('cluster.create.label.networkMode.overlay.desc')
                  }">
                  <i class="bcs-icon bcs-icon-info-circle-shape"></i>
                </span>
              </bk-radio>
              <!-- todo 暂不支持vpc网络 -->
              <bk-radio disabled value="vpc-cni">
                {{ $t('cluster.create.label.networkMode.vpc-cni.text') }}
                <span
                  class="ml5 text-[#C4C6CC]"
                  v-bk-tooltips="{
                    content: $t('cluster.create.label.networkMode.vpc-cni.desc')
                  }">
                  <i class="bcs-icon bcs-icon-info-circle-shape"></i>
                </span>
              </bk-radio>
            </bk-radio-group>
          </bk-form-item>
          <bk-form-item :label="$t('cluster.create.label.vpc.text')" property="vpcID" error-display-type="normal" required>
            <bcs-select
              class="max-w-[600px]"
              v-model="networkConfig.vpcID"
              :loading="vpcLoading"
              searchable
              :clearable="false"
              :remote-method="vpcRemoteSearch">
              <!-- VPC可用容器网络IP数量最低限制, 并以businessID分组（有businessID为业务专属VPC，否则为公共VPC） -->
              <bcs-option-group
                v-for="(vpc, index) in filterVpcList"
                :name="vpc.name"
                :key="index">

                <bcs-option
                  v-for="item in vpc.children"
                  :key="item.vpcID"
                  :id="item.vpcID"
                  :name="item.vpcName"
                  :disabled="basicInfo.environment === 'prod'
                    ? item.availableIPNum < 4096
                    : item.availableIPNum < 2048
                  "
                  v-bk-tooltips="{
                    content: $t('cluster.create.label.vpc.deficiencyIpNumTips'),
                    disabled: basicInfo.environment === 'prod'
                      ? item.availableIPNum >= 4096
                      : item.availableIPNum >= 2048
                  }">
                  <div class="flex items-center place-content-between">
                    <span>
                      {{item.vpcName}}
                      <span class="vpc-id">{{`(${item.vpcID})`}}</span>
                    </span>
                    <span class="text-[#979ba5]">
                      {{ $t('cluster.create.label.vpc.availableIpNum', [item.availableIPNum]) }}
                    </span>
                  </div>
                </bcs-option>
              </bcs-option-group>
            </bcs-select>
            <div class="text-[12px] text-[#979BA5]" v-if="curVpc">
              <i18n path="cluster.create.label.vpc.availableIpNum2">
                <span place="num" class="text-[#313238]">{{ curVpc.availableIPNum }}</span>
              </i18n>
            </div>
          </bk-form-item>
          <bk-form-item
            :label="$t('cluster.create.label.networkSetting.text')"
            property="networkSettings"
            error-display-type="normal"
            required>
            <div class="max-w-[600px] bg-[#F5F7FA] rounded-sm p-[16px] text-[12px]">
              <!-- 网络配置 -->
              <div class="flex items-center">
                <div class="flex-1 mr-[16px]">
                  <span
                    class="bcs-border-tips"
                    v-bk-tooltips="$t('cluster.create.label.networkSetting.cidrStep.desc')">
                    {{ $t('cluster.create.label.networkSetting.cidrStep.text') }}
                  </span>
                  <bcs-select class="bg-[#fff]" v-model="networkConfig.networkSettings.cidrStep" :clearable="false">
                    <bcs-option
                      v-for="ip in cidrStepList"
                      :key="ip"
                      :id="ip"
                      :name="ip">
                    </bcs-option>
                  </bcs-select>
                </div>
                <div class="flex-1 mr-[16px]">
                  <span
                    class="bcs-border-tips"
                    v-bk-tooltips="$t('cluster.create.label.networkSetting.maxServiceNum.desc')">
                    {{ $t('cluster.create.label.networkSetting.maxServiceNum.text') }}
                  </span>
                  <bcs-select class="bg-[#fff]" v-model="networkConfig.networkSettings.maxServiceNum" :clearable="false">
                    <bcs-option
                      v-for="item in serviceIpNumList"
                      :key="item"
                      :id="item"
                      :name="item">
                    </bcs-option>
                  </bcs-select>
                </div>
                <div class="flex-1">
                  <span
                    class="bcs-border-tips"
                    v-bk-tooltips="$t('cluster.create.label.networkSetting.maxNodePodNum.desc')">
                    {{ $t('cluster.create.label.networkSetting.maxNodePodNum.text') }}
                  </span>
                  <bcs-select class="bg-[#fff]" v-model="networkConfig.networkSettings.maxNodePodNum" :clearable="false">
                    <bcs-option v-for="item in nodePodNumList" :key="item" :id="item" :name="item"></bcs-option>
                  </bcs-select>
                </div>
              </div>
              <!-- 配置说明 -->
              <div class="flex flex-col mt-[10px] text-[#979BA5]">
                <i18n
                  class="leading-[20px]"
                  path="cluster.create.label.networkSetting.article1">
                  <span place="count" class="text-[#313238]">{{ maxNodeCount }}</span>
                </i18n>
                <i18n
                  class="leading-[20px]"
                  path="cluster.create.label.networkSetting.article2">
                  <span place="count" class="text-[#313238]">{{ maxCapacityCount }}</span>
                </i18n>
                <div class="leading-[20px]">
                  {{$t('cluster.create.label.networkSetting.article3')}}
                </div>
              </div>
            </div>
          </bk-form-item>
        </bk-form>
      </bcs-tab-panel>
      <!-- Master配置 -->
      <bcs-tab-panel :name="steps[2].name" :disabled="steps[2].disabled">
        <template #label>
          <StepTabLabel
            :title="$t('cluster.detail.title.master')"
            :step-num="3"
            :active="activeTabName === steps[2].name"
            :disabled="steps[2].disabled" />
        </template>
        <bk-form :ref="steps[2].formRef" :model="masterConfig" :rules="masterConfigRules">
          <bk-form-item :label="$t('cluster.create.label.manageType.text')">
            <div class="bk-button-group">
              <bk-button
                :class="['min-w-[136px]', { 'is-selected': manageType === 'MANAGED_CLUSTER' }]"
                @click="handleChangeManageType('MANAGED_CLUSTER')">
                <div class="flex items-center">
                  <span class="flex text-[16px] text-[#f85356]">
                    <i class="bcs-icon bcs-icon-hot"></i>
                  </span>
                  <span class="ml-[8px]">{{ $t('bcs.cluster.managed') }}</span>
                </div>
              </bk-button>
              <bk-button
                :class="['min-w-[136px]', { 'is-selected': manageType === 'INDEPENDENT_CLUSTER' }]"
                @click="handleChangeManageType('INDEPENDENT_CLUSTER')">
                {{ $t('bcs.cluster.selfDeployed') }}
              </bk-button>
            </div>
            <div class="text-[12px]">
              <span
                v-if="manageType === 'MANAGED_CLUSTER'">
                {{ $t('cluster.create.label.manageType.managed.desc') }}
              </span>
              <span
                v-else-if="manageType === 'INDEPENDENT_CLUSTER'">
                {{ $t('cluster.create.label.manageType.independent.desc') }}
              </span>
            </div>
          </bk-form-item>
          <bk-form-item key="level" :label="$t('cluster.create.label.manageType.managed.clusterLevel.text')" v-if="manageType === 'MANAGED_CLUSTER'">
            <div class="bk-button-group">
              <bk-button
                :class="['min-w-[48px]', { 'is-selected': item.level === masterConfig.clusterBasicSettings.clusterLevel }]"
                v-for="item in clusterScale"
                :key="item.level"
                @click="handleChangeClusterScale(item.level)">
                {{ item.level }}
              </bk-button>
              <bk-checkbox disabled :value="true" class="ml-[24px]">
                <span class="flex items-center">
                  <span class="text-[12px]">{{ $t('cluster.create.label.manageType.managed.automatic.text') }}</span>
                  <span
                    class="ml5"
                    v-bk-tooltips="{ content: $t('cluster.create.label.manageType.managed.automatic.tips') }">
                    <i class="bcs-icon bcs-icon-question-circle"></i>
                  </span>
                </span>
              </bk-checkbox>
            </div>
            <div class="text-[12px] leading-[20px] mt-[4px]">
              <i18n path="cluster.create.label.manageType.managed.clusterLevel.desc">
                <span place="nodes" class="text-[#313238]">{{ curClusterScale.level.split('L')[1] }}</span>
                <span place="pods" class="text-[#313238]">{{ curClusterScale.scale.maxNodePodNum }}</span>
                <span place="service" class="text-[#313238]">{{ curClusterScale.scale.maxServiceNum }}</span>
                <span place="crd" class="text-[#313238]">{{ curClusterScale.scale.cidrStep }}</span>
              </i18n>
            </div>
          </bk-form-item>
          <template v-else-if="manageType === 'INDEPENDENT_CLUSTER'">
            <bk-form-item key="master" class="tips-offset" :label-width="0.1" property="master" error-display-type="normal">
              <ApplyHost
                :region="networkConfig.region"
                :cloud-id="basicInfo.provider"
                :disabled-ip-list="nodesConfig.nodes.map(item => ({
                  ip: item.bk_host_innerip,
                  tips: $t('cluster.create.validate.ipExitInNode')
                }))"
                :region-list="regionList"
                :vpc="curVpc"
                v-model="masterConfig.master"
                @change="validateMaster" />
            </bk-form-item>
          </template>
        </bk-form>
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
        <bk-alert type="info">
          <template #title>
            <template v-if="manageType === 'MANAGED_CLUSTER'">
              <div>{{ $t('cluster.create.msg.managedClusterInfo') }}</div>
            </template>
            <template v-else-if="manageType === 'INDEPENDENT_CLUSTER'">
              <div>{{ $t('cluster.create.msg.independentClusterInfo.text') }}</div>
              <bk-checkbox v-model="skipAddNodes" class="text-[12px] mt5">
                <span class="text-[12px]">{{ $t('cluster.create.msg.independentClusterInfo.skip') }}</span>
              </bk-checkbox>
            </template>
          </template>
        </bk-alert>
        <bk-form :ref="steps[3].formRef" :model="nodesConfig" :rules="nodesConfigRules" class="mt-[16px]">
          <bk-form-item :label-width="0.1" class="tips-offset" property="nodes" error-display-type="normal">
            <ApplyHost
              :region="networkConfig.region"
              :cloud-id="basicInfo.provider"
              :disabled="manageType === 'INDEPENDENT_CLUSTER' && skipAddNodes"
              :disabled-ip-list="masterConfig.master.map(item => ({
                ip: item.bk_host_innerip,
                tips: $t('cluster.create.validate.ipExitInMaster')
              }))"
              :region-list="regionList"
              :vpc="curVpc"
              v-model="nodesConfig.nodes"
              @change="validateNodes" />
          </bk-form-item>
          <bk-form-item :label="$t('cluster.create.label.initNodeTemplate')">
            <TemplateSelector
              class="max-w-[500px]"
              v-model="nodesConfig.nodeTemplateID"
              :disabled="manageType === 'INDEPENDENT_CLUSTER' && skipAddNodes"
              provider="tencentCloud" />
          </bk-form-item>
        </bk-form>
      </bcs-tab-panel>
    </bcs-tab>
    <div class="mt-[24px]">
      <bk-button v-if="activeTabName !== steps[0].name" @click="preStep">{{ $t('generic.button.pre') }}</bk-button>
      <bk-button
        theme="primary"
        class="ml10"
        v-if="activeTabName === steps[steps.length - 1].name"
        @click="handleShowConfirmDialog">
        {{ $t('cluster.create.button.createCluster') }}
      </bk-button>
      <bk-button theme="primary" class="ml10" v-else @click="nextStep">{{ $t('generic.button.next') }}</bk-button>
      <bk-button class="ml10" @click="handleCancel">{{ $t('generic.button.cancel') }}</bk-button>
    </div>
    <ConfirmDialog
      v-model="showConfirmDialog"
      :width="800"
      :title="$t('cluster.create.button.confirmCreateCluster.text')"
      :sub-title="$t('generic.subTitle.confirmConfig')"
      :tips="confirmTips"
      :ok-text="$t('cluster.create.button.confirmCreate')"
      :cancel-text="$t('cluster.create.button.cancel')"
      theme="primary"
      :confirm="handleCreateCluster" />
  </BcsContent>
</template>
<script lang="ts">
import { computed, defineComponent, getCurrentInstance, onMounted, ref, watch } from 'vue';

import ApplyHost from './common/apply-host.vue';
import clusterScaleData from './common/cluster-scale.json';
import StepTabLabel from './common/step-tab-label.vue';

import $bkMessage from '@/common/bkmagic';
import ConfirmDialog from '@/components/comfirm-dialog.vue';
import BcsContent from '@/components/layout/Content.vue';
import { useProject } from '@/composables/use-app';
import $i18n from '@/i18n/i18n-setup';
import $router from '@/router';
import $store from '@/store';
import TemplateSelector from '@/views/cluster-manage/components/template-selector.vue';

interface IScale {
  level: string
  scale: {
    maxNodePodNum: number
    maxServiceNum: number
    cidrStep: number
  }
}
export default defineComponent({
  name: 'AddCluster',
  components: { ConfirmDialog, BcsContent, TemplateSelector, ApplyHost, StepTabLabel },
  setup() {
    const runEnv = ref(window.RUN_ENV);
    const steps = ref([
      { name: 'basicInfo', formRef: 'basicInfoRef', disabled: false },
      { name: 'network', formRef: 'networkRef', disabled: true },
      { name: 'master', formRef: 'masterRef', disabled: true },
      { name: 'nodes', formRef: 'nodesRef', disabled: true },
    ]);
    const activeTabName = ref<typeof steps.value[number]['name']>('basicInfo');
    // 基本信息
    const basicInfo = ref({
      clusterName: '',
      environment: '',
      provider: '',
      clusterBasicSettings: {
        version: '',
        isAutoUpgradeClusterLevel: true,
      },
      description: '',
    });
    const basicInfoRules = ref({
      clusterName: [
        {
          required: true,
          message: $i18n.t('generic.validate.required'),
          trigger: 'blur',
        },
      ],
      environment: [
        {
          required: true,
          message: $i18n.t('generic.validate.required'),
          trigger: 'blur',
        },
      ],
      'clusterBasicSettings.version': [
        {
          required: true,
          message: $i18n.t('generic.validate.required'),
          trigger: 'blur',
        },
      ],
    });
    // 网络配置
    const networkConfig = ref({
      region: '',
      networkType: 'overlay',
      vpcID: '',
      networkSettings: {
        cidrStep: '',
        maxNodePodNum: '',
        maxServiceNum: '',
      },
    });
    const networkConfigRules = ref({
      region: [
        {
          required: true,
          message: $i18n.t('generic.validate.required'),
          trigger: 'blur',
        },
      ],
      networkType: [
        {
          required: true,
          message: $i18n.t('generic.validate.required'),
          trigger: 'blur',
        },
      ],
      vpcID: [
        {
          required: true,
          message: $i18n.t('generic.validate.required'),
          trigger: 'blur',
        },
      ],
      networkSettings: [
        {
          message: $i18n.t('generic.validate.required'),
          trigger: 'blur',
          validator: () => Object.keys(networkConfig.value.networkSettings)
            .every(key => !!networkConfig.value.networkSettings[key]),
        },
      ],
    });
    // master配置
    const masterConfig = ref<{
      clusterBasicSettings: {
        clusterLevel: string
      }
      master: any[]
    }>({
      clusterBasicSettings: {
        clusterLevel: 'L20',
      },
      master: [],
    });
    // 动态 i18n 问题，这里使用computed
    const masterConfigRules = computed(() => ({
      master: [{
        message: $i18n.t('cluster.create.validate.masterNum35'),
        trigger: 'custom',
        validator: () => masterConfig.value.master.length && [3, 5].includes(masterConfig.value.master.length),
      }, {
        message: '',
        trigger: 'custom',
        validator: () => masterConfig.value.master.every(item => item.vpc === networkConfig.value.vpcID),
      }],
    }));
    // 节点配置
    const nodesConfig = ref<{
      nodeTemplateID: string
      nodes: any[]
    }>({
      nodeTemplateID: '',
      nodes: [],
    });
    // 动态 i18n 问题，这里使用computed
    const nodesConfigRules = computed(() => ({
      nodes: [{
        message: $i18n.t('generic.validate.required'),
        trigger: 'custom',
        validator: () => manageType.value === 'INDEPENDENT_CLUSTER' || !!nodesConfig.value.nodes.length,
      }, {
        message: '',
        trigger: 'custom',
        validator: () => nodesConfig.value.nodes.every(item => item.vpc === networkConfig.value.vpcID)
          || (skipAddNodes.value && manageType.value === 'INDEPENDENT_CLUSTER'),
      }],
    }));

    const skipAddNodes = ref(false);
    // 集群规模
    const clusterScale = ref<IScale[]>(clusterScaleData.data);
    const curClusterScale = computed<IScale>(() => clusterScale.value
      .find(item => item.level === masterConfig.value.clusterBasicSettings.clusterLevel)
      || { level: 'L5', scale: { maxNodePodNum: 0, maxServiceNum: 0, cidrStep: 0 } });
    // 托管集群规格
    const handleChangeClusterScale = (scale) => {
      masterConfig.value.clusterBasicSettings.clusterLevel = scale;
    };
    // 集群模板
    const templateList = ref<any[]>([]);
    const templateLoading = ref(false);
    const handleGetTemplateList = async () => {
      templateLoading.value = true;
      templateList.value = await $store.dispatch('clustermanager/fetchCloudList');
      basicInfo.value.provider = templateList.value[0]?.cloudID || '';
      templateLoading.value = false;
    };
    // 版本列表
    const versionList = computed(() => {
      const cloud = templateList.value.find(item => item.cloudID === basicInfo.value.provider);
      return cloud?.clusterManagement.availableVersion || [];
    });
    // 区域列表
    watch(() => basicInfo.value.provider, () => {
      getRegionList();
    });
    const regionList = ref<any[]>([]);
    const regionLoading = ref(false);
    const getRegionList = async () => {
      regionLoading.value = true;
      regionList.value = await $store.dispatch('clustermanager/fetchCloudRegion', {
        $cloudId: basicInfo.value.provider,
      });
      regionLoading.value = false;
    };
    watch(() => basicInfo.value.environment, () => {
      networkConfig.value.networkSettings.cidrStep = '';
      networkConfig.value.vpcID = '';
    });
    // 网络配置信息
    // watch(() => networkConfig.value.vpcID, () => {
    //   // 重置网络配置
    //   set(networkConfig.value, 'networkSettings', {
    //     cidrStep: '',
    //     maxNodePodNum: '',
    //     maxServiceNum: '',
    //   });
    // });
    watch(() => [networkConfig.value.region, networkConfig.value.networkType], () => {
      // 区域和网络类型变更时重置vpcId
      networkConfig.value.vpcID = '';
      getVpcList();
    });
    // vpc列表
    const filterValue = ref('');
    const vpcList = ref<any[]>([]);
    const vpcLoading = ref(false);
    const filterVpcList = computed(() => vpcList.value
      // filter放到reduce
      .reduce((acc, cur: any) => { // 分两组
        // 匹配下拉搜索项
        if (cur.vpcName.includes(filterValue.value) || cur.vpcID.includes(filterValue.value)) {
          // 有businessID的是归属业务专属vpc，index为0
          const index = cur.businessID ? 0 : 1;
          acc[index].children.push(cur);
        }
        return acc;
      }, [ // vpc组：业务专属vpc，公共vpc
        { id: 1, name: $i18n.t('cluster.create.label.vpc.businessSpecificVPC'), children: [] },
        { id: 2, name: $i18n.t('cluster.create.label.vpc.publicVPC'), children: [] },
      ])
      .filter(item => item.children.length > 0), // 去除空组
    );
    const getVpcList = async () => {
      vpcLoading.value = true;
      const data = await $store.dispatch('clustermanager/fetchCloudVpc', {
        cloudID: basicInfo.value.provider,
        region: networkConfig.value.region,
        networkType: networkConfig.value.networkType,
        businessID: curProject.value.businessID,
      });
      vpcList.value = data.filter(item => item.available === 'true');
      vpcLoading.value = false;
    };
    const getIpNumRange = (minExponential, maxExponential) => {
      const list: number[] = [];
      for (let i = minExponential; i <= maxExponential; i++) {
        list.push(Math.pow(2, i));
      }
      return list;
    };
    // vpc搜索
    const vpcRemoteSearch = (v) => {
      filterValue.value = v;
    };
    // 当前选择VPC
    const curVpc = computed(() => vpcList.value.find(item => item.vpcID === networkConfig.value.vpcID));
    // IP数量列表
    const cidrStepList = computed(() => {
      const cloud = templateList.value.find(item => item.cloudID === basicInfo.value.provider);
      const cidrStep = cloud?.networkInfo?.cidrStep || [];
      // 测试环境不允许选择4096
      return basicInfo.value.environment === 'prod'
        ? cidrStep
        : cidrStep.filter(ip => ip !== 4096);
    });
    // service ip选择列表
    const serviceIpNumList = computed(() => {
      const ipNumber = Number(networkConfig.value.networkSettings.cidrStep);
      if (!ipNumber) return [];

      const minExponential = Math.log2(128);
      const maxExponential = Math.log2(ipNumber / 2);

      return getIpNumRange(minExponential, maxExponential);
    });
    // pod数量列表
    const nodePodNumList = computed(() => {
      const ipNumber = Number(networkConfig.value.networkSettings.cidrStep);
      const serviceNumber = Number(networkConfig.value.networkSettings.maxServiceNum);
      if (!ipNumber || !serviceNumber) return [];

      const minExponential = Math.log2(32);
      const maxExponential = Math.log2(Math.min(ipNumber - serviceNumber, 256));
      return getIpNumRange(minExponential, maxExponential);
    });
    // 集群最大节点数
    const maxNodeCount = computed(() => {
      const { cidrStep, maxServiceNum, maxNodePodNum } = networkConfig.value.networkSettings;
      if (cidrStep && maxServiceNum && maxNodePodNum) {
        return Math.floor((Number(cidrStep) - Number(maxServiceNum)) / Number(maxNodePodNum)) || 0;
      }
      return 0;
    });
    // 扩容后最大节点数
    const maxCapacityCount = computed(() => {
      const { cidrStep, maxServiceNum, maxNodePodNum } = networkConfig.value.networkSettings;
      if (cidrStep && maxServiceNum && maxNodePodNum) {
        return Math.floor((Number(cidrStep) * 5 - Number(maxServiceNum)) / Number(maxNodePodNum)) || 0;
      }
      return 0;
    });

    // master配置
    const manageType = ref<'INDEPENDENT_CLUSTER' | 'MANAGED_CLUSTER'>('MANAGED_CLUSTER');
    const handleChangeManageType = (type: 'INDEPENDENT_CLUSTER' | 'MANAGED_CLUSTER') => {
      manageType.value = type;
      masterConfig.value.master = [];
      validateMaster();
    };

    // 上一步
    const preStep = async () => {
      const index = steps.value.findIndex(step => activeTabName.value === step.name);
      if (index > -1 && index - 1 >= 0) {
        activeTabName.value = steps.value[index - 1]?.name;
      }
    };
    // 下一步
    const nextStep = async () => {
      const $refs = proxy?.$refs || {};
      const index = steps.value.findIndex(step => activeTabName.value === step.name);
      const validate = await ($refs[steps.value[index]?.formRef] as any)?.validate().catch(() => false);
      if (!validate) return;
      if (index > -1 && index + 1 < steps.value.length) {
        steps.value[index + 1].disabled = false;
        activeTabName.value = steps.value[index + 1]?.name;
      }
    };
    const handleCancel = () => {
      $router.back();
    };
    // 创建集群
    const showConfirmDialog = ref(false);
    const confirmTips = computed(() => {
      const maxNodePodNum = Number(networkConfig.value?.networkSettings?.maxNodePodNum || 0);
      const clusterTips = [
        $i18n.t('cluster.create.button.confirmCreateCluster.doc.article1', {
          num: maxNodePodNum - 3,
        }),
        $i18n.t('cluster.create.button.confirmCreateCluster.doc.article2'),
      ];
      const nodesTips = [
        $i18n.t('cluster.create.button.confirmCreateCluster.doc.article3', {
          ip: nodesConfig.value.nodes[0]?.bk_host_innerip,
          num: nodesConfig.value.nodes.length,
        }),
      ];
      if (skipAddNodes.value && manageType.value === 'INDEPENDENT_CLUSTER') {
        return clusterTips;
      }
      return nodesConfig.value.nodes.length ? clusterTips.concat(nodesTips) : clusterTips;
    });
    const { proxy } = getCurrentInstance() || { proxy: null };
    const handleShowConfirmDialog = async () => {
      const $refs = proxy?.$refs || {};
      const validateList = steps.value.map(step => ($refs[step.formRef] as any)?.validate().catch(() => {
        activeTabName.value = step.name;
        return false;
      }));
      const validateResults = await Promise.all(validateList);
      if (validateResults.some(result => !result)) return;

      showConfirmDialog.value = true;
    };
    const { curProject } = useProject();
    const user = computed(() => $store.state.user);
    const handleCreateCluster = async () => {
      const params: Record<string, any> = {
        projectID: curProject.value.projectID,
        businessID: String(curProject.value.businessID),
        engineType: 'k8s',
        isExclusive: true,
        clusterType: 'single',
        creator: user.value.username,
        manageType: manageType.value,
        ...basicInfo.value,
        ...networkConfig.value,
        master: [],
        nodes: [],
        nodeTemplateID: '',
      };
      // 独立集群
      if (manageType.value === 'INDEPENDENT_CLUSTER') {
        params.master = masterConfig.value.master.map(item => item.bk_host_innerip);
      }
      // 托管集群
      if (manageType.value === 'MANAGED_CLUSTER') {
        params.clusterBasicSettings.clusterLevel = masterConfig.value.clusterBasicSettings.clusterLevel;
      }
      // 添加节点
      if (!skipAddNodes.value || manageType.value === 'MANAGED_CLUSTER') {
        params.nodes = nodesConfig.value.nodes.map(item => item.bk_host_innerip);
        params.nodeTemplateID = nodesConfig.value.nodeTemplateID;
      }

      const result = await $store.dispatch('clustermanager/createCluster', params);
      if (result) {
        $bkMessage({
          theme: 'success',
          message: $i18n.t('generic.msg.success.deliveryTask'),
        });
        $router.push({ name: 'clusterMain' });
      }
    };

    // 校验master节点
    const validateMaster = async () => {
      const $refs = proxy?.$refs || {};
      return await ($refs.masterRef as any)?.validate().catch(() => false);
    };
    // 校验node节点
    const validateNodes = async () => {
      const $refs = proxy?.$refs || {};
      return await ($refs.nodesRef as any)?.validate().catch(() => false);
    };

    onMounted(() => {
      handleGetTemplateList();
    });

    return {
      steps,
      activeTabName,
      runEnv,
      basicInfo,
      networkConfig,
      masterConfig,
      nodesConfig,
      basicInfoRules,
      networkConfigRules,
      masterConfigRules,
      nodesConfigRules,
      skipAddNodes,
      clusterScale,
      curClusterScale,
      templateLoading,
      versionList,
      regionLoading,
      regionList,
      vpcLoading,
      vpcList,
      filterVpcList,
      curVpc,
      cidrStepList,
      serviceIpNumList,
      nodePodNumList,
      maxNodeCount,
      maxCapacityCount,
      manageType,
      handleChangeManageType,
      handleChangeClusterScale,
      preStep,
      nextStep,
      handleCancel,
      showConfirmDialog,
      confirmTips,
      handleShowConfirmDialog,
      handleCreateCluster,
      validateMaster,
      validateNodes,
      vpcRemoteSearch,
    };
  },
});
</script>
<style lang="postcss" scoped>
>>> .tips-offset .form-error-tip {
  margin-left: 150px;
}
</style>
