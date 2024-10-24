<template>
  <bk-form
    :model="basicInfo"
    :rules="basicInfoRules"
    ref="formRef"
    class="k8s-form grid grid-cols-2 grid-rows-1 gap-[16px]">
    <DescList size="middle" :title="$t('cluster.create.label.clusterInfo')">
      <bk-form-item :label="$t('cluster.create.label.kubernetesProvider')">
        {{ $t('provider.weiruanyun') }}
      </bk-form-item>
      <bk-form-item :label="$t('tke.label.account')" property="cloudAccountID" error-display-type="normal" required>
        <bk-select :loading="accountLoading" v-model="basicInfo.cloudAccountID" searchable :clearable="false">
          <bk-option
            v-for="item in accountList"
            :key="item.account.accountID"
            :id="item.account.accountID"
            :name="item.account.accountName">
          </bk-option>
          <template slot="extension">
            <SelectExtension
              :link-text="$t('tke.link.cloudToken')"
              @link="gotoTencentToken"
              @refresh="handleGetAccounts" />
          </template>
        </bk-select>
      </bk-form-item>
      <bk-form-item
        :label="$t('importAzureCloud.label.resourceGroups')"
        property="extraInfo.cloudProjectId"
        error-display-type="normal"
        required>
        <bk-select
          searchable
          :clearable="false"
          :loading="projectLoading"
          v-model="basicInfo.extraInfo.cloudProjectId">
          <bk-option
            v-for="item in projectList"
            :key="item.projectID"
            :id="item.projectID"
            :name="item.projectID !== '0' ? `${item.projectName}(${item.projectID})` : item.projectName">
          </bk-option>
          <template slot="extension">
            <SelectExtension
              :link-text="$t('tke.link.cloudProject')"
              @link="gotoTencentProject"
              @refresh="handleGetProjects" />
          </template>
        </bk-select>
      </bk-form-item>
      <bk-form-item :label="$t('cluster.labels.name')" property="clusterName" error-display-type="normal" required>
        <bk-input
          :maxlength="64"
          :placeholder="$t('cluster.create.validate.name')"
          v-model.trim="basicInfo.clusterName">
        </bk-input>
      </bk-form-item>
      <bk-form-item
        :label="$t('cluster.create.label.clusterVersion')"
        property="clusterBasicSettings.version"
        error-display-type="normal"
        required>
        <bk-select
          :loading="templateLoading"
          v-model="basicInfo.clusterBasicSettings.version"
          searchable
          :clearable="false"
          class="max-w-[calc(50%-46px)]">
          <bk-option v-for="item in versionList" :key="item" :id="item" :name="item"></bk-option>
        </bk-select>
      </bk-form-item>
      <bk-form-item
        :label="$t('k8s.label')"
        property="labels"
        error-display-type="normal">
        <KeyValue
          v-model="basicInfo.labels"
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
        <bk-input maxlength="100" v-model="basicInfo.description" type="textarea"></bk-input>
      </bk-form-item>
    </DescList>
    <div>
      <DescList size="middle" :title="$t('cluster.labels.env')">
        <bk-form-item :label="$t('cluster.create.label.region')" property="region" error-display-type="normal" required>
          <Region
            :cloud-account-i-d="basicInfo.cloudAccountID"
            :cloud-i-d="cloudID"
            init-data
            v-model="basicInfo.region" />
        </bk-form-item>
        <bk-form-item
          :label="$t('tke.label.nodemanArea')"
          property="clusterBasicSettings.area.bkCloudID"
          error-display-type="normal"
          required>
          <bk-select
            searchable
            :clearable="false"
            :loading="nodemanCloudLoading"
            v-model="basicInfo.clusterBasicSettings.area.bkCloudID">
            <bk-option
              v-for="item in nodemanCloudList"
              :key="item.bk_cloud_id"
              :id="item.bk_cloud_id"
              :name="item.bk_cloud_name">
            </bk-option>
            <template slot="extension">
              <SelectExtension
                :link-text="$t('tke.link.nodeman')"
                :link="`${PROJECT_CONFIG.nodemanHost}/#/cloud-manager`"
                @refresh="handleGetNodeManCloud" />
            </template>
          </bk-select>
        </bk-form-item>
        <bk-form-item :label="$t('cluster.labels.env')" property="environment" error-display-type="normal" required>
          <bk-radio-group v-model="basicInfo.environment">
            <bk-radio value="debug">
              {{ $t('cluster.env.debug') }}
            </bk-radio>
            <bk-radio value="prod">
              {{ $t('cluster.env.prod') }}
            </bk-radio>
          </bk-radio-group>
        </bk-form-item>
      </DescList>
      <DescList class="mt-[24px]" size="middle" :title="$t('cluster.title.clusterConfig')">
        <bk-form-item :label="$t('cluster.create.azure.chargingModel')">
        </bk-form-item>
        <bk-form-item :label="$t('k8s.updateStrategy.text')">
        </bk-form-item>
      </DescList>
    </div>
    <div class="flex items-center h-[48px] bg-[#FAFBFD] px-[24px] fixed bottom-0 left-0 w-full bcs-border-top">
      <bk-button theme="primary" class="ml10" @click="nextStep">{{ $t('generic.button.next') }}</bk-button>
      <bk-button class="ml10" @click="handleCancel">{{ $t('generic.button.cancel') }}</bk-button>
    </div>
  </bk-form>
</template>
<script setup lang="ts">
import { computed, onMounted, PropType, ref, watch } from 'vue';

import { ICloudProject, INodeManCloud } from '../../../types/types';

import {
  cloudDetail,
  cloudProjects,
  cloudVersionModules,
  nodemanCloud,
} from '@/api/modules/cluster-manager';
import { LABEL_KEY_REGEXP } from '@/common/constant';
import DescList from '@/components/desc-list.vue';
import SelectExtension from '@/components/select-extension.vue';
import $i18n from '@/i18n/i18n-setup';
import $router from '@/router';
import Region from '@/views/cluster-manage/add/components/region.vue';
import KeyValue from '@/views/cluster-manage/components/key-value.vue';
import useCloud, { ICloudAccount } from '@/views/cluster-manage/use-cloud';;

const props = defineProps({
  cloudID: {
    type: String as PropType<CloudID>,
    default: '',
  },
});

const emits = defineEmits(['next', 'cancel']);

// 基本信息
const basicInfo = ref({
  cloudAccountID: '', // 云凭证
  clusterName: '',
  environment: '',
  provider: '',
  clusterBasicSettings: {
    version: '',
    OS: '',
    area: {  // 云区域
      bkCloudID: 0,
    },
  },
  description: '',
  region: '',
  labels: {},
  clusterAdvanceSettings: {
    containerRuntime: '', // 运行时
    runtimeVersion: '', // 运行时版本
  },
  extraInfo: {
    cloudProjectId: '0',
  },
});
watch(() => basicInfo.value.cloudAccountID, () => {
  basicInfo.value.extraInfo.cloudProjectId = '0';
  basicInfo.value.region = '';
});
watch(() => basicInfo.value.region, () => {
  basicInfo.value.clusterBasicSettings.OS = '';
});
const basicInfoRules = ref({
  cloudAccountID: [
    {
      required: true,
      message: $i18n.t('generic.validate.required'),
      trigger: 'blur',
    },
  ],
  'extraInfo.cloudProjectId': [
    {
      required: true,
      message: $i18n.t('generic.validate.required'),
      trigger: 'blur',
    },
  ],
  clusterName: [
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
      required: true,
      message: $i18n.t('generic.validate.required'),
      trigger: 'blur',
    },
  ],
  'clusterBasicSettings.area.bkCloudID': [
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

// 集群模板
const templateDetail = ref<Partial<ICloudTemplateDetail>>({});
const templateLoading = ref(false);
const handleGetTemplateList = async () => {
  templateLoading.value = true;
  templateDetail.value = await cloudDetail({
    $cloudId: props.cloudID,
  }).catch(() => []);
  basicInfo.value.provider = templateDetail.value?.cloudID || '';
  // 初始化默认集群版本
  basicInfo.value.clusterBasicSettings.version = templateDetail.value?.clusterManagement?.availableVersion?.[0] || '';
  templateLoading.value = false;
};

// 云凭证
const { cloudAccounts } = useCloud();
const accountLoading = ref(false);
const accountList = ref<ICloudAccount[]>([]);
const handleGetAccounts = async () => {
  accountLoading.value = true;
  const { data } = await cloudAccounts(props.cloudID);
  accountList.value = data;
  // 初始化第一个云凭证
  basicInfo.value.cloudAccountID = accountList.value?.[0]?.account?.accountID || '';
  accountLoading.value = false;
};
const gotoTencentToken = () => {
  const { href } = $router.resolve({ name: 'tencentPublicCloud' });
  window.open(href);
};

// 云项目
const projectLoading = ref(false);
const projectList = ref<Array<ICloudProject>>([]);
const handleGetProjects = async () => {
  if (!basicInfo.value.cloudAccountID) return;

  projectLoading.value = true;
  projectList.value = await cloudProjects({
    accountID: basicInfo.value.cloudAccountID,
    $cloudId: props.cloudID,
  }).catch(() => []);
  projectList.value.unshift({
    projectID: '0',
    projectName: $i18n.t('tke.label.defaultProject'),
  });
  projectLoading.value = false;
};

// 版本列表
const versionList = computed(() => templateDetail.value?.clusterManagement?.availableVersion || []);


// 管控区域
const nodemanCloudList = ref<Array<INodeManCloud>>([]);
const nodemanCloudLoading = ref(false);
const handleGetNodeManCloud = async () => {
  nodemanCloudLoading.value = true;
  nodemanCloudList.value = await nodemanCloud().catch(() => []);
  nodemanCloudLoading.value = false;
};

const moduleLoading = ref(false);
const runtimeModuleParams = ref<IRuntimeModuleParams[]>([]);
const getRuntimeModuleParams = async () => {
  if (!basicInfo.value.clusterBasicSettings.version) return;
  moduleLoading.value = true;
  const data = await cloudVersionModules({
    $cloudId: props.cloudID,
    $version: basicInfo.value.clusterBasicSettings.version,
    $module: 'runtime',
  });
  runtimeModuleParams.value = data.filter(item => item.enable);
  // 初始化默认运行时
  basicInfo.value.clusterAdvanceSettings.containerRuntime = runtimeModuleParams.value?.[0]?.flagName || '';
  basicInfo.value.clusterAdvanceSettings.runtimeVersion = runtimeModuleParams.value?.[0]?.defaultValue || '';
  moduleLoading.value = false;
};
const gotoTencentProject = () => {
  window.open('https://console.cloud.tencent.com/project');
};

watch(() => basicInfo.value.cloudAccountID, () => {
  handleGetProjects();
});

// 运行时组件参数
watch(() => basicInfo.value.clusterBasicSettings.version, () => {
  getRuntimeModuleParams();
});

onMounted(() => {
  handleGetTemplateList();
  handleGetAccounts();
  handleGetNodeManCloud();
});

// 校验
const formRef = ref();
const validate = async () => {
  const result = await formRef.value?.validate().catch(() => false);
  return result;
};

// 下一步
const nextStep = async () => {
  const result = await validate();
  result && emits('next', basicInfo.value);
};
// 取消
const handleCancel = () => {
  emits('cancel');
};

defineExpose({
  validate,
});
</script>
<style scoped lang="postcss">
/deep/ .bk-option-group-name {
  border-bottom: 0 !important;
}
</style>
