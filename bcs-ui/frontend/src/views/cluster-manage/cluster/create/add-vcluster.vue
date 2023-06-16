<template>
  <BcsContent :title="$t('添加集群')">
    <bcs-tab :label-height="42" :active.sync="activeTabName">
      <!-- 基本信息 -->
      <bcs-tab-panel :name="steps[0].name">
        <template #label>
          <StepTabLabel :title="$t('基本信息')" :step-num="1" :active="activeTabName === steps[0].name" />
        </template>
        <bk-form :ref="steps[0].formRef" :model="basicInfo" :rules="basicInfoRules">
          <bk-form-item :label="$t('集群类型')">
            <span class="text-[12px]">vCluster</span>
          </bk-form-item>
          <bk-form-item :label="$t('集群名称')" property="clusterName" error-display-type="normal" required>
            <bk-input
              :maxlength="64"
              :placeholder="$t('请输入集群名称，不超过64个字符')"
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
          <bk-form-item :label="$t('所属区域')" property="region" error-display-type="normal" required>
            <bcs-select
              class="max-w-[600px]"
              v-model="basicInfo.region"
              :loading="regionLoading"
              searchable
              :clearable="false">
              <bcs-option
                v-for="item in regionList"
                :key="item.region"
                :id="item.region"
                :name="item.regionName" />
            </bcs-select>
          </bk-form-item>
          <bk-form-item
            :label="$t('集群版本')"
            property="clusterBasicSettings.version"
            error-display-type="normal"
            required>
            <bcs-select
              class="max-w-[600px]"
              :loading="loading"
              v-model="basicInfo.clusterBasicSettings.version"
              searchable
              :clearable="false">
              <bcs-option
                v-for="item in versionList"
                :key="item"
                :id="item"
                :name="item" />
            </bcs-select>
          </bk-form-item>
          <!-- 接口拉取 -->
          <bk-form-item
            :label="$t('Kubernetes提供商')"
            error-display-type="normal"
            property="kubernetes"
            required>
            <bcs-select class="max-w-[600px]" v-model="basicInfo.extraInfo.provider" :clearable="false">
              <bcs-option id="DevCloud" name="DevCloud"></bcs-option>
              <!-- <bcs-option id="IDC" :name="$t('IDC自研云')"></bcs-option> -->
            </bcs-select>
          </bk-form-item>
          <bk-form-item :label="$t('描述')">
            <bk-input maxlength="100" class="max-w-[600px]" v-model="basicInfo.description" type="textarea"></bk-input>
          </bk-form-item>
        </bk-form>
      </bcs-tab-panel>
      <!-- 配额管理 -->
      <bcs-tab-panel :name="steps[1].name" :disabled="steps[1].disabled">
        <template #label>
          <StepTabLabel
            :title="$t('配额管理')"
            :step-num="2"
            :active="activeTabName === steps[1].name"
            :disabled="steps[1].disabled" />
        </template>
        <bk-form :ref="steps[1].formRef" :model="quotaInfo" :rules="quotaInfoRules">
          <bk-form-item :label="$t('配额方案')" error-display-type="normal" property="ns.quota" required>
            <ClusterQuota v-model="quota" />
            <p
              class="text-[12px] text-[#ea3636]"
              v-if="quota.cpu < 40 || quota.mem < 40">
              {{ $t('CPU配额最少40核，内存配额最少40GiB') }}
            </p>
          </bk-form-item>
          <bk-form-item :label="$t('IP资源')" error-display-type="normal" property="networkSettings" required>
            <div class="flex items-center bg-[#F5F7FA] p-[16px] text-[12px] max-w-[600px]">
              <div class="flex-1">
                <div>{{ $t('Pod IP 数量') }}</div>
                <bcs-select class="bg-[#fff]" v-model="quotaInfo.networkSettings.maxNodePodNum">
                  <bcs-option v-for="item in [1024, 512, 256, 128, 64]" :id="item" :name="item" :key="item" />
                </bcs-select>
              </div>
              <div class="flex-1 ml-[16px]">
                <div>{{ $t('Service IP 数量') }}</div>
                <bcs-select class="bg-[#fff]" v-model="quotaInfo.networkSettings.maxServiceNum">
                  <bcs-option v-for="item in [256, 128, 64]" :id="item" :name="item" :key="item" />
                </bcs-select>
              </div>
            </div>
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
        @click="createVCluster">
        {{ $t('创建集群') }}
      </bk-button>
      <bk-button theme="primary" class="ml10" v-else @click="nextStep">{{ $t('下一步') }}</bk-button>
      <bk-button class="ml10" @click="handleCancel">{{ $t('取消') }}</bk-button>
    </div>
  </BcsContent>
</template>
<script lang="ts" setup>
import { ref, onBeforeMount, getCurrentInstance, computed } from 'vue';
import BcsContent from '@/components/layout/Content.vue';
import StepTabLabel from './step-tab-label.vue';
import ClusterQuota from './cluster-quota.vue';
import $store from '@/store';
import $router from '@/router';
import { useVCluster } from '@/views/cluster-manage/cluster/use-cluster';
import $i18n from '@/i18n/i18n-setup';
import $bkInfo from '@/components/bk-magic-2.0/bk-info';
import { useProject } from '@/composables/use-app';

const steps = ref([
  { name: 'basicInfo', formRef: 'basicInfoRef', disabled: false },
  { name: 'quotaInfo', formRef: 'networkRef', disabled: true },
]);
const activeTabName = ref<typeof steps.value[number]['name']>('basicInfo');

const runEnv = ref(window.RUN_ENV);
// 基本信息
const basicInfo = ref({
  clusterName: '',
  environment: '',
  provider: 'tencentCloud',
  clusterBasicSettings: {
    version: '',
  },
  region: '',
  description: '',
  extraInfo: {
    provider: 'DevCloud',
  },
});
const basicInfoRules = ref({
  clusterName: [
    {
      required: true,
      message: $i18n.t('必填项'),
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
  region: [
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
  // 当前字段暂时不传递
  kubernetes: [
    {
      required: true,
      message: $i18n.t('必填项'),
      trigger: 'blur',
    },
  ],
});

// 配额管理
const quota = ref({
  mem: 1,
  cpu: 1,
});
const quotaInfo = ref({
  ns: {
    quota: {
      cpuRequests: '',
      cpuLimits: '',
      memoryRequests: '',
      memoryLimits: '',
    },
  },
  networkSettings: {
    maxNodePodNum: '',
    maxServiceNum: '',
  },
});
const quotaInfoRules = ref({
  'ns.quota': [
    {
      message: $i18n.t('必填项'),
      trigger: 'blur',
      validator() {
        return Object.keys(quota.value).every(key => quota.value[key]);
      },
    },
  ],
  networkSettings: [
    {
      message: $i18n.t('必填项'),
      trigger: 'blur',
      validator() {
        return quotaInfo.value.networkSettings.maxNodePodNum && quotaInfo.value.networkSettings.maxServiceNum;
      },
    },
  ],
});

const { loading, sharedClusterList, getSharedclusters, handleCreateVCluster } = useVCluster();

// 区域列表
const regionData = ref<any[]>([]);
const regionList = computed(() => regionData.value.filter(item => sharedClusterList.value
  .some(cluster => cluster.region === item.region)));
const regionLoading = ref(false);
const getRegionList = async () => {
  regionLoading.value = true;
  regionData.value = await $store.dispatch('clustermanager/fetchCloudRegion', {
    $cloudId: basicInfo.value.provider,
  });
  regionLoading.value = false;
};
// const isRegionDisabled = (region) => {
//   const  { version } = basicInfo.value.clusterBasicSettings;
//   return sharedClusterList.value
//     .filter(item => !version || item.clusterBasicSettings.version === version)
//     .every(item => item.region !== region);
// };

// 版本列表
const versionList = computed(() => sharedClusterList.value.map(item => item.clusterBasicSettings?.version));
// const isVersionDisabled = (version) => {
//   const { region } = basicInfo.value;
//   return sharedClusterList.value
//     .filter(item => !region || item.region === region)
//     .every(item => item.clusterBasicSettings.version !== version);
// };

// 创建集群
const { curProject } = useProject();
const user = computed(() => $store.state.user);
const createVCluster = async () => {
  if (quota.value.cpu < 40 || quota.value.mem < 40) {
    activeTabName.value = 'quotaInfo';
    return;
  };
  const $refs = proxy?.$refs || {};
  const validateList = steps.value.map(step => ($refs[step.formRef] as any)?.validate().catch(() => {
    activeTabName.value = step.name;
    return false;
  }));
  const validateResults = await Promise.all(validateList);
  if (validateResults.some(result => !result)) return;

  $bkInfo({
    type: 'warning',
    title: $i18n.t('确认创建集群'),
    clsName: 'custom-info-confirm default-info',
    subTitle: basicInfo.value.clusterName,
    confirmFn: async () => {
      quotaInfo.value.ns.quota = {
        cpuRequests: String(quota.value.cpu),
        cpuLimits: String(quota.value.cpu),
        memoryRequests: `${String(quota.value.mem)}Gi`,
        memoryLimits: `${String(quota.value.mem)}Gi`,
      };
      const result = await handleCreateVCluster({
        projectID: curProject.value.projectID,
        businessID: curProject.value.businessID,
        projectCode: curProject.value.projectCode,
        engineType: 'k8s',
        isExclusive: true,
        clusterType: 'single',
        creator: user.value.username,
        clusterAdvanceSettings: {},
        ...basicInfo.value,
        ...quotaInfo.value,
      });
      result && $router.push({ name: 'clusterMain' });
    },
  });
};
// 上一步
const preStep = async () => {
  const index = steps.value.findIndex(step => activeTabName.value === step.name);
  if (index > -1 && index - 1 >= 0) {
    activeTabName.value = steps.value[index - 1]?.name;
  }
};
// 下一步
const { proxy } = getCurrentInstance() || { proxy: null };
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
// 取消
const handleCancel = () => {
  $router.back();
};

onBeforeMount(() => {
  getRegionList();
  getSharedclusters();
});
</script>
