<!-- eslint-disable max-len -->
<template>
  <BcsContent :title="$t('添加集群')">
    <bcs-tab :label-height="42" :active.sync="activeTabName">
      <!-- 基本信息 -->
      <bcs-tab-panel :name="steps[0].name">
        <template #label>
          <StepTabLabel :title="$t('基本信息')" :step-num="1" :active="activeTabName === steps[0].name" />
        </template>
        <bk-form :ref="steps[0].formRef" :model="basicInfo" :rules="basicInfoRules">
          <!-- <bk-form-item :label="$t('集群类型')">
            {{ type || '--' }}
          </bk-form-item> -->
          <bk-form-item :label="$t('集群名称')" property="clusterName" error-display-type="normal" required>
            <bk-input
              :maxlength="64"
              :placeholder="$t('仅支持中文、英文、数字和字符{0}, 长短0~64字符', ['-_[]()【】（）'])"
              class="max-w-[600px]"
              v-model.trim="basicInfo.clusterName">
            </bk-input>
          </bk-form-item>
          <bk-form-item :label="$t('集群环境')" property="environment" error-display-type="normal" required>
            <bk-radio-group v-model="basicInfo.environment">
              <bk-radio value="stag" v-if="runEnv === 'dev'">
                UAT
              </bk-radio>
              <bk-radio :disabled="runEnv === 'dev'" value="debug">
                {{ $t('测试环境') }}
              </bk-radio>
              <bk-radio :disabled="runEnv === 'dev'" value="prod">
                {{ $t('正式环境') }}
              </bk-radio>
            </bk-radio-group>
          </bk-form-item>
          <bk-form-item :label="$t('集群版本')" property="clusterBasicSettings.version" error-display-type="normal" required>
            <bcs-select
              class="max-w-[600px]"
              :loading="templateLoading"
              v-model="basicInfo.clusterBasicSettings.version"
              searchable
              :clearable="false">
              <bcs-option v-for="item in versionList" :key="item" :id="item" :name="item"></bcs-option>
            </bcs-select>
          </bk-form-item>
          <bk-form-item :label="$t('描述')">
            <bk-input maxlength="100" class="max-w-[600px]" v-model="basicInfo.description" type="textarea"></bk-input>
          </bk-form-item>
        </bk-form>
      </bcs-tab-panel>
      <!-- 网络配置 -->
      <bcs-tab-panel :name="steps[1].name" :disabled="steps[1].disabled">
        <template #label>
          <StepTabLabel
            :title="$t('网络配置')"
            :step-num="2"
            :active="activeTabName === steps[1].name"
            :disabled="steps[1].disabled" />
        </template>
        <bk-form :ref="steps[1].formRef" :model="networkConfig" :rules="networkConfigRules">
          <bk-form-item :label="$t('所属区域')" property="region" error-display-type="normal" required>
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
          <bk-form-item :label="$t('网络模式')" property="networkType" error-display-type="normal" required>
            <bk-radio-group v-model="networkConfig.networkType">
              <bk-radio value="overlay">
                {{ $t('全局路由模式') }}
                <span
                  class="ml5 text-[#C4C6CC]"
                  v-bk-tooltips="{
                    content: $t('全局路由网络模式是TKE基于底层私有网络（VPC）的全局路由能力，实现了容器网络和VPC互访的路由策略，类似云原生的Overlay网络模式')
                  }">
                  <i class="bcs-icon bcs-icon-info-circle-shape"></i>
                </span>
              </bk-radio>
              <!-- todo 暂不支持vpc网络 -->
              <bk-radio disabled value="vpc-cni">
                {{ $t('全局路由与 VPC-CNI 混合模式') }}
                <span
                  class="ml5 text-[#C4C6CC]"
                  v-bk-tooltips="{
                    content: $t('集群同时支持全局路由模式与VPC-CNI模式，VPC-CNI 网络模式是 TKE 基于 CNI 和 VPC 弹性网卡实现的容器网络能力，VPC-CNI网络模式类似云原生的Underlay网络模式，如需使用请联系蓝鲸容器助手协助')
                  }">
                  <i class="bcs-icon bcs-icon-info-circle-shape"></i>
                </span>
              </bk-radio>
            </bk-radio-group>
          </bk-form-item>
          <bk-form-item :label="$t('所属VPC')" property="vpcID" error-display-type="normal" required>
            <bcs-select
              class="max-w-[600px]"
              v-model="networkConfig.vpcID"
              :loading="vpcLoading"
              searchable
              :clearable="false"
              :remote-method="vpcRemoteSearch">
              <!-- VPC可用容器网络IP数量最低限制 -->
              <bcs-option
                v-for="item in filterVpcList"
                :key="item.vpcID"
                :id="item.vpcID"
                :name="item.vpcName"
                :disabled="basicInfo.environment === 'prod'
                  ? item.availableIPNum < 4096
                  : item.availableIPNum < 2048
                "
                v-bk-tooltips="{
                  content: $t('可用容器网络IP数量不足'),
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
                    {{ $t('可用IP数量: {0}', [item.availableIPNum]) }}
                  </span>
                </div>
              </bcs-option>
            </bcs-select>
            <div class="text-[12px] text-[#979BA5]" v-if="curVpc">
              <i18n path="可用容器网络 IP {num} 个">
                <span place="num" class="text-[#313238]">{{ curVpc.availableIPNum }}</span>
              </i18n>
            </div>
          </bk-form-item>
          <bk-form-item
            :label="$t('全局路由网络分配')"
            property="networkSettings"
            error-display-type="normal"
            required>
            <div class="max-w-[600px] bg-[#F5F7FA] rounded-sm p-[16px] text-[12px]">
              <!-- 网络配置 -->
              <div class="flex items-center">
                <div class="flex-1 mr-[16px]">
                  <span
                    class="bcs-border-tips"
                    v-bk-tooltips="$t('集群内总的全局路由容器网络可用IP数量，IP数量 = 集群内Service数量上限 + 单节点Pod数量上限 * 节点数量')">
                    {{ $t('IP数量') }}
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
                    v-bk-tooltips="$t('集群Service可用IP数量上限，分配后将无法调整，请谨慎评估后再填写')">
                    {{ $t('集群内Service数量上限') }}
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
                    v-bk-tooltips="$t('单节点Pod数量上限一旦分配后将无法调整，请谨慎评估后再填写')">
                    {{ $t('单节点Pod数量上限') }}
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
                  path="容器网络资源有限，请合理分配，当前容器网络配置下，集群最多可以添加 {count} 个节点">
                  <span place="count" class="text-[#313238]">{{ maxNodeCount }}</span>
                </i18n>
                <i18n
                  class="leading-[20px]"
                  path="当容器网络资源超额使用时，会触发容器网络自动扩容，扩容后集群最多可以添加 {count} 个节点">
                  <span place="count" class="text-[#313238]">{{ maxCapacityCount }}</span>
                </i18n>
                <div class="leading-[20px]">
                  {{$t('集群可添加节点数（包含Master节点与Node节点）= (IP数量 - Service的数量) / 单节点Pod数量上限')}}
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
            :title="$t('Master配置')"
            :step-num="3"
            :active="activeTabName === steps[2].name"
            :disabled="steps[2].disabled" />
        </template>
        <bk-form :ref="steps[2].formRef" :model="masterConfig" :rules="masterConfigRules">
          <bk-form-item :label="$t('集群模式')">
            <div class="bk-button-group">
              <bk-button
                :class="['min-w-[136px]', { 'is-selected': manageType === 'MANAGED_CLUSTER' }]"
                @click="handleChangeManageType('MANAGED_CLUSTER')">
                <div class="flex items-center">
                  <span class="flex text-[16px] text-[#f85356]">
                    <i class="bcs-icon bcs-icon-hot"></i>
                  </span>
                  <span class="ml-[8px]">{{ $t('托管集群') }}</span>
                </div>
              </bk-button>
              <bk-button
                :class="['min-w-[136px]', { 'is-selected': manageType === 'INDEPENDENT_CLUSTER' }]"
                @click="handleChangeManageType('INDEPENDENT_CLUSTER')">
                {{ $t('独立集群') }}
              </bk-button>
            </div>
            <div class="text-[12px]">
              <span
                v-if="manageType === 'MANAGED_CLUSTER'">
                {{ $t('Kubernetes 集群的 Master 和 Etcd 会由 TKE 团队集中管理和维护，集群管理员不需要关心集群 Master 的管理和维护。') }}
              </span>
              <span
                v-else-if="manageType === 'INDEPENDENT_CLUSTER'">
                {{ $t('使用申请或已存在 CVM 资源作为 Master 节点所需资源，仅支持节点数量为 3 台与 5 台。') }}
              </span>
            </div>
          </bk-form-item>
          <bk-form-item key="level" :label="$t('集群规格')" v-if="manageType === 'MANAGED_CLUSTER'">
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
                  <span class="text-[12px]">{{ $t('自动升配') }}</span>
                  <span
                    class="ml5"
                    v-bk-tooltips="{ content: $t('已开启自动升配功能，当集群资源超过当前规格设定阈值时，会自动升级到下一个等级。') }">
                    <i class="bcs-icon bcs-icon-question-circle"></i>
                  </span>
                </span>
              </bk-checkbox>
            </div>
            <div class="text-[12px] leading-[20px] mt-[4px]">
              <i18n path="当前集群规格最多管理 {nodes} 个节点，{pods} 个 Pod，{service} 个 ConfigMap，{crd} 个 CRD">
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
                  tips: $t('当前IP已经被添加为节点')
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
            :title="$t('添加节点')"
            :step-num="4"
            :active="activeTabName === steps[3].name"
            :disabled="steps[3].disabled" />
        </template>
        <bk-alert type="info">
          <template #title>
            <template v-if="manageType === 'MANAGED_CLUSTER'">
              <div>{{ $t('创建托管集群必须添加至少一个节点，以运行必要的服务') }}</div>
            </template>
            <template v-else-if="manageType === 'INDEPENDENT_CLUSTER'">
              <div>{{ $t('独立集群可选择在创建集群后再按需添加节点') }}</div>
              <bk-checkbox v-model="skipAddNodes" class="text-[12px] mt5">
                <span class="text-[12px]">{{ $t('跳过，后续添加') }}</span>
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
                tips: $t('当前IP已经被添加为Master')
              }))"
              :region-list="regionList"
              :vpc="curVpc"
              v-model="nodesConfig.nodes"
              @change="validateNodes" />
          </bk-form-item>
          <bk-form-item :label="$t('节点初始化模板')">
            <TemplateSelector
              v-model="nodesConfig.nodeTemplateID"
              :disabled="manageType === 'INDEPENDENT_CLUSTER' && skipAddNodes"
              is-tke-cluster />
          </bk-form-item>
        </bk-form>
      </bcs-tab-panel>
    </bcs-tab>
    <div class="mt-[24px]">
      <bk-button v-if="activeTabName !== steps[0].name" @click="preStep">{{ $t('上一步') }}</bk-button>
      <bk-button
        theme="primary"
        class="ml10"
        v-if="activeTabName === steps[steps.length - 1].name"
        @click="handleShowConfirmDialog">
        {{ $t('创建集群') }}
      </bk-button>
      <bk-button theme="primary" class="ml10" v-else @click="nextStep">{{ $t('下一步') }}</bk-button>
      <bk-button class="ml10" @click="handleCancel">{{ $t('取消') }}</bk-button>
    </div>
    <ConfirmDialog
      v-model="showConfirmDialog"
      :width="800"
      :title="$t('确定创建集群')"
      :sub-title="$t('请确认以下配置:')"
      :tips="confirmTips"
      :ok-text="$t('确定创建')"
      :cancel-text="$t('我再想想')"
      theme="primary"
      :confirm="handleCreateCluster" />
  </BcsContent>
</template>
<script lang="ts">
import { defineComponent, ref, computed, onMounted, watch, getCurrentInstance } from 'vue';
import ConfirmDialog from '@/components/comfirm-dialog.vue';
import BcsContent from '@/components/layout/Content.vue';
import TemplateSelector from '@/views/cluster-manage/components/template-selector.vue';
import ApplyHost from './apply-host.vue';
import StepTabLabel from './step-tab-label.vue';
import $i18n from '@/i18n/i18n-setup';
import $store from '@/store';
import $router from '@/router';
import { useProject } from '@/composables/use-app';
import clusterScaleData from './cluster-scale.json';
import $bkMessage from '@/common/bkmagic';

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
          message: $i18n.t('必填项'),
          trigger: 'blur',
        },
        {
          validator(value) {
            return /^[\u4e00-\u9fa50-9a-zA-Z-_[\]()【】（）]+$/g.test(value);
          },
          message: $i18n.t('仅支持中文、英文、数字和字符{0}', ['-_[]()【】（）']),
          trigger: 'blur',
        },
      ],
      environment: [
        {
          required: true,
          message: $i18n.t('必填项'),
          trigger: 'blur',
        },
      ],
      'clusterBasicSettings.version': [
        {
          required: true,
          message: $i18n.t('必填项'),
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
          message: $i18n.t('必填项'),
          trigger: 'blur',
        },
      ],
      networkType: [
        {
          required: true,
          message: $i18n.t('必填项'),
          trigger: 'blur',
        },
      ],
      vpcID: [
        {
          required: true,
          message: $i18n.t('必填项'),
          trigger: 'blur',
        },
      ],
      networkSettings: [
        {
          message: $i18n.t('必填项'),
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
        message: $i18n.t('仅支持 3 台与 5 台'),
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
        message: $i18n.t('必填项'),
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
      .filter(item => item.vpcName.includes(filterValue.value) || item.vpcID.includes(filterValue.value)));
    const getVpcList = async () => {
      vpcLoading.value = true;
      const data = await $store.dispatch('clustermanager/fetchCloudVpc', {
        cloudID: basicInfo.value.provider,
        region: networkConfig.value.region,
        networkType: networkConfig.value.networkType,
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
      const maxExponential = Math.log2(Math.min(ipNumber - serviceNumber, 128));
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
        $i18n.t('该集群创建后单个节点最大允许创建 {num} 个pod（TKE内部需占用3个IP），创建后不允许调整，请慎重确认', {
          num: maxNodePodNum - 3,
        }),
        $i18n.t('为了保证集群环境标准化，创建集群会格式化数据盘/dev/vdb，盘内数据将被清除，请确认该数据盘内没有放置业务数据'),
      ];
      const nodesTips = [
        $i18n.t('为了保证集群环境标准化，此操作会对 {ip} 等 {num} 个IP进行操作系统重装初始化和安装容器服务相关组件等操作', {
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
          message: $i18n.t('任务下发成功'),
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
