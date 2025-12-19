<!-- eslint-disable max-len -->
<template>
  <BcsContent :title="$t('cluster.button.addCluster')">
    <bcs-tab :label-height="42" :active.sync="activeTabName">
      <!-- 基本信息 -->
      <bcs-tab-panel :name="steps[0].name">
        <template #label>
          <StepTabLabel :title="$t('generic.title.basicInfo1')" :step-num="1" :active="activeTabName === steps[0].name" />
        </template>
        <bk-form
          class="k8s-form grid grid-cols-2 grid-rows-1 gap-[16px]"
          :ref="steps[0].formRef"
          :model="basicInfo"
          :rules="basicInfoRules">
          <!-- <bk-form-item :label="$t('cluster.labels.clusterType')">
            {{ type || '--' }}
          </bk-form-item> -->
          <DescList size="middle" :title="$t('cluster.create.label.clusterInfo')">
            <bk-form-item :label="$t('cluster.labels.name')" property="clusterName" error-display-type="normal" required>
              <bk-input
                :maxlength="64"
                :placeholder="$t('cluster.create.validate.name')"
                class="max-w-[600px]"
                v-model.trim="basicInfo.clusterName">
              </bk-input>
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
            <bk-form-item
              :label="$t('k8s.label')"
              property="labels"
              error-display-type="normal">
              <KeyValue
                v-model="basicInfo.labels"
                class="max-w-[600px]"
                :key-rules="[
                  {
                    message: $t('generic.validate.label'),
                    validator: LABEL_KEY_REGEXP,
                  }
                ]"
                :value-rules="[
                  {
                    message: $t('generic.validate.label'),
                    validator: LABEL_KEY_REGEXP,
                  }
                ]" />
            </bk-form-item>
            <bk-form-item :label="$t('cluster.create.label.desc')">
              <bk-input maxlength="100" class="max-w-[600px]" v-model="basicInfo.description" type="textarea"></bk-input>
            </bk-form-item>
          </DescList>
          <div>
            <DescList size="middle" :title="$t('cluster.labels.env')">
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
              <bk-form-item
                :label="$t('tke.label.nodemanArea')"
                property="clusterBasicSettings.area.bkCloudID"
                error-display-type="normal"
                required>
                <NodeManArea
                  class="max-w-[600px]"
                  disabled
                  v-model="basicInfo.clusterBasicSettings.area.bkCloudID" />
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
            </DescList>
            <DescList class="mt-[24px]" size="middle" :title="$t('cluster.title.clusterConfig')">
              <bk-form-item
                :label="$t('cluster.create.label.system')"
                property="clusterBasicSettings.OS"
                error-display-type="normal"
                required>
                <ImageList
                  class="max-w-[600px]"
                  v-model="imageID"
                  :region="networkConfig.region"
                  :cloud-i-d="basicInfo.provider"
                  init-data
                  @os-change="handleSetOS" />
              </bk-form-item>
              <bk-form-item
                :label="$t('cluster.create.label.containerRuntime')"
                property="clusterAdvanceSettings.containerRuntime"
                error-display-type="normal"
                required>
                <bk-radio-group
                  v-model="basicInfo.clusterAdvanceSettings.containerRuntime"
                  @change="handleRuntimeChange">
                  <bk-radio
                    value="containerd"
                    :disabled="!runtimeModuleParamsMap['containerd']"
                  >
                    <span
                      v-bk-tooltips="{
                        content: $t('tke.tips.notSupportInCurrentClusterVersion'),
                        disabled: runtimeModuleParamsMap['containerd']
                      }">
                      containerd
                    </span>
                  </bk-radio>
                  <bk-radio
                    value="docker"
                    :disabled="!runtimeModuleParamsMap['docker']"
                  >
                    <span
                      v-bk-tooltips="{
                        content: $t('tke.tips.notSupportInCurrentClusterVersion'),
                        disabled: runtimeModuleParamsMap['docker']
                      }">
                      docker
                    </span>
                  </bk-radio>
                </bk-radio-group>
              </bk-form-item>
              <bk-form-item
                :label="$t('cluster.create.label.runtimeVersion')"
                property="clusterAdvanceSettings.runtimeVersion"
                error-display-type="normal"
                required>
                <RuntimeVersions
                  class="max-w-[50%]"
                  :version="basicInfo.clusterBasicSettings.version"
                  :container-runtime="basicInfo.clusterAdvanceSettings.containerRuntime"
                  :cloud-id="basicInfo.provider"
                  init-data
                  v-model="basicInfo.clusterAdvanceSettings.runtimeVersion"
                  @data-change="handleVersions" />
              </bk-form-item>
            </DescList>
          </div>
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
              disabled
              :clearable="false">
              <bcs-option
                v-for="item in regionList"
                :key="item.region" :id="item.region" :name="item.regionName"></bcs-option>
            </bcs-select>
          </bk-form-item>
          <bk-form-item :label="$t('cluster.create.label.privateNet.text')" property="vpcID" error-display-type="normal" required>
            <div class="flex items-center w-full">
              <bcs-select
                class="flex-1 max-w-[600px]"
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
              <span
                :class="[
                  'inline-flex items-center',
                  'px-[16px] h-[24px] rounded-full bg-[#F0F1F5] text-[12px] ml-[8px]'
                ]"
                v-if="curVpc">
                {{ $t('tke.tips.totalIpNum', [curVpc.availableIPNum || '--']) }}
              </span>
            </div>
          </bk-form-item>
          <bk-form-item
            :label="$t('cluster.create.label.defaultNetPlugin')"
            property="networkType"
            error-display-type="normal"
            required>
            <bcs-select :clearable="false" searchable v-model="networkConfig.networkType" class="max-w-[600px]">
              <bcs-option id="overlay" :name="$t('cluster.create.label.networkMode.overlay.text')"></bcs-option>
            </bcs-select>
          </bk-form-item>
          <!-- VPC-CNI需要设置POD ID -->
          <bk-form-item label="VPC-CNI">
            <bcs-switcher v-model="networkConfig.networkSettings.enableVPCCni"></bcs-switcher>
          </bk-form-item>
          <template v-if="networkConfig.networkSettings.enableVPCCni">
            <bk-form-item
              label="Pod IP"
              :desc="$t('tke.tips.vpcCniPodIp')"
              property="podIP"
              error-display-type="normal"
              required>
              <VpcCni
                :subnets="networkConfig.networkSettings.subnetSource.new"
                cloud-i-d="tencentCloud"
                :region="networkConfig.region"
                :node-pod-num-list="podNumList"
                class="max-w-[600px]"
                @change="handleSetSubnetSourceNew" />
            </bk-form-item>
            <!-- <bk-form-item
              :label="$t('tke.label.staticIpMode')"
              :desc="{
                allowHTML: true,
                content: '#staticIpMode',
              }"
              key="staticIpMode">
              <bk-checkbox v-model="networkConfig.networkSettings.isStaticIpMode">
                {{ $t('tke.label.enableStaticIpMode') }}
              </bk-checkbox>
              <div id="staticIpMode">
                <div>{{ $t('tke.tips.staticIpMode.p1') }}</div>
                <i18n tag="div" path="tke.tips.staticIpMode.p2">
                  <bk-link theme="primary" target="_blank" href="https://cloud.tencent.com/document/product/457/34994">
                    <span class="text-[12px]">{{ $t('tke.link.staticIpMode') }}</span>
                  </bk-link>
                </i18n>
              </div>
            </bk-form-item> -->
          </template>
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
                  <template #count>
                    <span class="text-[#313238]">{{ maxNodeCount }}</span>
                  </template>
                </i18n>
                <i18n
                  class="leading-[20px]"
                  path="cluster.create.label.networkSetting.article2">
                  <template #count>
                    <span class="text-[#313238]">{{ maxCapacityCount }}</span>
                  </template>
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
            :title="$t('cluster.detail.title.controlConfig')"
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
                <template #nodes>
                  <span class="text-[#313238]">{{ curClusterScale.level.split('L')[1] }}</span>
                </template>
                <template #pods>
                  <span class="text-[#313238]">{{ curClusterScale.scale.maxNodePodNum }}</span>
                </template>
                <template #service>
                  <span class="text-[#313238]">{{ curClusterScale.scale.maxServiceNum }}</span>
                </template>
                <template #crd>
                  <span class="text-[#313238]">{{ curClusterScale.scale.cidrStep }}</span>
                </template>
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
          <bk-form-item
            class="tips-offset"
            error-display-type="normal"
            :label-width="0.1"
            :property="manageType === 'INDEPENDENT_CLUSTER' ? 'nodes' : ''">
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
      <bk-button
        theme="primary"
        :class="activeTabName !== steps[0].name ?'ml10':''"
        v-else
        @click="nextStep">{{ $t('generic.button.next') }}</bk-button>
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

import { cloudList } from '@/api/modules/cluster-manager';
import $bkMessage from '@/common/bkmagic';
import { LABEL_KEY_REGEXP } from '@/common/constant';
import ConfirmDialog from '@/components/comfirm-dialog.vue';
import DescList from '@/components/desc-list.vue';
import BcsContent from '@/components/layout/Content.vue';
import { useProject } from '@/composables/use-app';
import $i18n from '@/i18n/i18n-setup';
import $router from '@/router';
import $store from '@/store';
import ApplyHost from '@/views/cluster-manage/add/components/apply-host.vue';
import clusterScaleData from '@/views/cluster-manage/add/components/cluster-scale.json';
import NodeManArea from '@/views/cluster-manage/add/components/nodeman-area.vue';
import RuntimeVersions from '@/views/cluster-manage/add/components/runtime-versions.vue';
import StepTabLabel from '@/views/cluster-manage/add/components/step-tab-label.vue';
import KeyValue from '@/views/cluster-manage/components/key-value.vue';
import TemplateSelector from '@/views/cluster-manage/components/template-selector.vue';
import VpcCni from '@/views/cluster-manage/components/vpc-cni.vue';
import ImageList from '@/views/cluster-manage/node-list/tencent-image-list.vue';

interface IScale {
  level: string
  scale: {
    maxNodePodNum: number
    maxServiceNum: number
    cidrStep: number
  }
}
export default defineComponent({
  name: 'TencentCloud',
  components: {
    ConfirmDialog,
    BcsContent,
    TemplateSelector,
    ApplyHost,
    StepTabLabel,
    VpcCni,
    DescList,
    ImageList,
    NodeManArea,
    RuntimeVersions,
    KeyValue,
  },
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
      provider: 'tencentCloud',
      clusterBasicSettings: {
        version: '',
        isAutoUpgradeClusterLevel: true,
        area: {
          bkCloudID: 0,
        },
        OS: '',
      },
      labels: {},
      description: '',
      clusterAdvanceSettings: {
        containerRuntime: '', // 运行时
        runtimeVersion: '', // 运行时版本
      },
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
      region: [
        {
          message: $i18n.t('generic.validate.required'),
          trigger: 'custom',
          validator: () => !!networkConfig.value.region,
        },
      ],
      'clusterBasicSettings.OS': [
        {
          required: true,
          message: $i18n.t('generic.validate.required'),
          trigger: 'blur',
        },
      ],
      'clusterAdvanceSettings.containerRuntime': [
        {
          required: true,
          message: $i18n.t('generic.validate.required'),
          trigger: 'blur',
        },
      ],
      'clusterAdvanceSettings.runtimeVersion': [
        {
          required: true,
          message: $i18n.t('generic.validate.required'),
          trigger: 'blur',
        },
      ],
      labels: [
        {
          message: $i18n.t('generic.validate.label'),
          trigger: 'custom',
          validator: () => {
            const { labels } = basicInfo.value;
            const rule = new RegExp(LABEL_KEY_REGEXP);
            return Object.keys(labels).every(key => rule.test(key) && rule.test(labels[key]));
          },
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
        enableVPCCni: false,
        isStaticIpMode: true, // 暂不支持配置
        claimExpiredSeconds: 300,
        subnetSource: {
          new: [],
        },
        networkMode: 'tke-route-eni', // 共享网卡模式
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
          validator: () => networkConfig.value.networkSettings.cidrStep
          && networkConfig.value.networkSettings.maxNodePodNum
          && networkConfig.value.networkSettings.maxServiceNum,
        },
      ],
      // POD IP数量
      podIP: [
        {
          trigger: 'blur',
          message: $i18n.t('generic.validate.required'),
          validator() {
            if (networkConfig.value.networkSettings.enableVPCCni) {
              const subnetSource = networkConfig.value.networkSettings.subnetSource.new as Array<{
                ipCnt: number
                zone: string
              }>;
              return subnetSource.length && subnetSource.every(item => item.ipCnt && item.zone);
            }
            return true;
          },
        },
        {
          trigger: 'custom',
          message: $i18n.t('tke.validate.podIPsNeedLessThanVpcIPs'),
          validator() {
            const subnetSource = networkConfig.value.networkSettings.subnetSource.new as Array<{
              ipCnt: number
              zone: string
            }>;
            const counts = subnetSource.reduce((counts, item) => {
              counts += item.ipCnt;
              return counts;
            }, 0);

            return counts <= (curVpc.value.availableIPNum || 0);
          },
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
      master: [
        {
          message: $i18n.t('cluster.create.validate.masterNum35'),
          trigger: 'custom',
          validator: () => masterConfig.value.master.length && [3, 5].includes(masterConfig.value.master.length),
        },
      ],
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
      nodes: [
        {
          message: $i18n.t('generic.validate.required'),
          trigger: 'custom',
          validator: () => manageType.value === 'INDEPENDENT_CLUSTER' || !!nodesConfig.value.nodes.length,
        },
        {
          message: $i18n.t('generic.validate.required'),
          trigger: 'custom',
          validator: () => !!nodesConfig.value.nodes.length || (skipAddNodes.value && manageType.value === 'INDEPENDENT_CLUSTER'),
        },
        {
          message: $i18n.t('cluster.create.validate.maxNum50'),
          trigger: 'custom',
          validator: () => nodesConfig.value.nodes.length <= 50 || (skipAddNodes.value && manageType.value === 'INDEPENDENT_CLUSTER'),
        },
      ],
    }));

    const skipAddNodes = ref(false);
    watch(skipAddNodes, () => {
      const $refs = proxy?.$refs || {};
      const index = steps.value.findIndex(step => activeTabName.value === step.name);
      ($refs[steps.value[index]?.formRef] as any)?.validate().catch(() => false);
    });
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
    const businessID = computed(() => $store.state.curProject.businessID);
    const cloudInfo = computed(() => templateList.value.find(item => item.cloudID === basicInfo.value.provider));
    const podNumList = computed(() => cloudInfo.value?.networkInfo?.underlaySteps || []);
    const templateList = ref<any[]>([]);
    const templateLoading = ref(false);
    const handleGetTemplateList = async () => {
      templateLoading.value = true;
      templateList.value = await cloudList({
        businessID: businessID.value,
      }).catch(() => []);
      templateLoading.value = false;
    };
    // 版本列表
    const versionList = computed(() => cloudInfo.value?.clusterManagement?.availableVersion || []);
    const regionList = ref<any[]>([]);
    const regionLoading = ref(false);
    const getRegionList = async () => {
      regionLoading.value = true;
      regionList.value = await $store.dispatch('clustermanager/fetchCloudRegion', {
        $cloudId: basicInfo.value.provider,
      });
      regionLoading.value = false;
    };
    // 操作系统
    const imageID = ref('');
    const handleSetOS = (image: string) => {
      basicInfo.value.clusterBasicSettings.OS = image;
    };
    // 运行时组件变更
    const handleRuntimeChange = (flagName: string) => {
      basicInfo.value.clusterAdvanceSettings.runtimeVersion = runtimeModuleParamsMap.value[flagName]?.defaultValue;
    };
    watch(() => networkConfig.value.region, () => {
      basicInfo.value.clusterBasicSettings.OS = '';
      imageID.value = '';
    });
    // 运行时组件
    const runtimeModuleParamsMap = ref<Record<string, IRuntimeModuleParams>>({});
    function handleVersions(runtimeModuleParams, runtimeParamsMap) {
      runtimeModuleParamsMap.value = runtimeParamsMap;
      // 初始化默认运行时
      basicInfo.value.clusterAdvanceSettings.containerRuntime = runtimeModuleParams?.[0]?.flagName || '';
      basicInfo.value.clusterAdvanceSettings.runtimeVersion = runtimeModuleParams?.[0]?.defaultValue || '';
    }

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
    // vpc搜索
    const vpcRemoteSearch = (v) => {
      filterValue.value = v;
    };
    // 当前选择VPC
    const curVpc = computed(() => vpcList.value.find(item => item.vpcID === networkConfig.value.vpcID));

    // 设置vpc-cni子网
    const handleSetSubnetSourceNew = (data) => {
      networkConfig.value.networkSettings.subnetSource.new = data;
    };

    // IP数量列表
    const cidrStepList = computed(() => {
      const cidrStep = cloudInfo.value?.networkInfo?.cidrSteps
        ?.filter(item => item.env === basicInfo.value.environment)
        ?.map(item => item.step) || [];
      // 测试环境不允许选择4096
      return basicInfo.value.environment === 'prod'
        ? cidrStep
        : cidrStep.filter(ip => ip !== 4096);
    });
    // service ip选择列表
    const serviceIpNumList = computed(() => cloudInfo.value?.networkInfo?.serviceSteps || []);
    // pod数量列表
    const nodePodNumList = computed(() => cloudInfo.value?.networkInfo?.perNodePodNum || []);
    // 集群最大节点数
    const maxNodeCount = computed(() => {
      const { cidrStep, maxServiceNum, maxNodePodNum } = networkConfig.value.networkSettings;
      if (cidrStep && maxServiceNum && maxNodePodNum) {
        const result = Math.floor((Number(cidrStep) - Number(maxServiceNum)) / Number(maxNodePodNum));
        if (result > 0) {
          return result;
        }
      }
      return 0;
    });
    // 扩容后最大节点数
    const maxCapacityCount = computed(() => {
      const { cidrStep, maxServiceNum, maxNodePodNum } = networkConfig.value.networkSettings;
      if (cidrStep && maxServiceNum && maxNodePodNum) {
        const result = Math.floor((Number(cidrStep) * 5 - Number(maxServiceNum)) / Number(maxNodePodNum));
        if (result > 0) {
          return result;
        }
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
      getRegionList();
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
      podNumList,
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
      handleSetSubnetSourceNew,
      imageID,
      handleSetOS,
      handleRuntimeChange,
      runtimeModuleParamsMap,
      handleVersions,
      LABEL_KEY_REGEXP,
    };
  },
});
</script>
<style lang="postcss" scoped>
>>> .tips-offset .form-error-tip {
  margin-left: 150px;
}
</style>
