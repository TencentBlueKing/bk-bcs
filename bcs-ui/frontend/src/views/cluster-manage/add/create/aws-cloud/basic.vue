<template>
  <bk-form
    :model="basicInfo"
    :rules="basicInfoRules"
    ref="formRef"
    class="k8s-form grid grid-cols-2 grid-rows-1 gap-[16px]">
    <DescList size="middle" :title="$t('cluster.create.label.clusterInfo')">
      <bk-form-item :label="$t('cluster.create.label.kubernetesProvider')">
        {{ $t('provider.yamaxunyun') }}
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
      <bk-form-item :label="$t('cluster.create.label.role')" property="role" error-display-type="normal" required>
        <bk-select :loading="roleLoading" v-model="basicInfo.clusterIamRole" searchable :clearable="false">
          <bk-option
            v-for="item in roleList"
            :key="item.roleName"
            :id="item.roleName"
            :name="item.roleName">
          </bk-option>
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
          v-model="basicInfo.clusterBasicSettings.clusterTags"
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
        <bk-form-item
          :desc="{
            content: upgradeStrategyData.flagDesc,
          }"
          :label="$t('k8s.updateStrategy.text')">
          <div class="flex">
            <div
              @click="handleUpgrade('EXTENDED')"
              :class="[
                'flex-1 rounded-md border p-[8px] pt-[2px] cursor-pointer',
                upgradeStrategyData.defaultValue === 'EXTENDED' ? 'bg-[#E1F0FF] border-[#3A84FF]' : '',
              ]">
              <div class="flex items-center">
                <bk-radio :checked="upgradeStrategyData.defaultValue === 'EXTENDED'"></bk-radio>
                <span class="font-bold ml-[6px]">{{ labelEnum['EXTENDED'] }}</span>
              </div>
              <div class="text-[12px] leading-normal ml-[24px]">{{ upgradEnum['EXTENDED'] }}</div>
            </div>
            <div
              @click="handleUpgrade('STANDARD')"
              :class="[
                'flex-1 ml-[10px] rounded-md border p-[8px] pt-[2px] cursor-pointer',
                upgradeStrategyData.defaultValue === 'STANDARD' ? 'bg-[#E1F0FF] border-[#3A84FF]' : '',
              ]">
              <div class="flex items-center">
                <bk-radio :checked="upgradeStrategyData.defaultValue === 'STANDARD'"></bk-radio>
                <span class="font-bold ml-[6px]">{{ labelEnum['STANDARD'] }}</span>
              </div>
              <div class="text-[12px] leading-normal ml-[24px]">{{ upgradEnum['STANDARD'] }}</div>
            </div>
          </div>
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

import { INodeManCloud } from '../../../types/types';

import {
  cloudDetail,
  cloudVersionModules,
  nodemanCloud,
} from '@/api/modules/cluster-manager';
import { CLUSTER_NAME_REGEX, LABEL_KEY_REGEXP } from '@/common/constant';
import DescList from '@/components/desc-list.vue';
import SelectExtension from '@/components/select-extension.vue';
import $i18n from '@/i18n/i18n-setup';
import $router from '@/router';
import Region from '@/views/cluster-manage/add/components/region.vue';
import KeyValue from '@/views/cluster-manage/components/key-value.vue';
import useCloud, { ICloudAccount, ICloudRole } from '@/views/cluster-manage/use-cloud';

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
  clusterIamRole: '', // 集群服务角色
  clusterBasicSettings: {
    version: '',
    area: {  // 云区域
      bkCloudID: 0,
    },
    clusterTags: {},
    upgradePolicy: {
      supportType: 'EXTENDED',
    },
  },
  description: '',
  region: '',
  labels: {},
  clusterAdvanceSettings: {
    containerRuntime: '', // 运行时
    runtimeVersion: '', // 运行时版本
  },
});
watch(() => basicInfo.value.cloudAccountID, () => {
  basicInfo.value.region = '';
  handleGetRoles();
});
const basicInfoRules = ref({
  cloudAccountID: [
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
    {
      message: $i18n.t('generic.validate.clusterName'),
      validator: () => {
        const { clusterName } = basicInfo.value;
        const rule = new RegExp(CLUSTER_NAME_REGEX);
        return rule.test(clusterName);
      },
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
const { cloudAccounts, getCloudRolesList } = useCloud();
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
  const { href } = $router.resolve({ name: 'amazonCloud' });
  window.open(href);
};

// 集群服务角色
const roleLoading = ref(false);
const roleList = ref<ICloudRole[]>([]);
async function handleGetRoles() {
  roleLoading.value = true;
  const { data } = await getCloudRolesList(props.cloudID, basicInfo.value.cloudAccountID, 'cluster');
  roleList.value = data;
  // 初始化第一个角色
  basicInfo.value.clusterIamRole = roleList.value?.[0]?.roleName || '';
  roleLoading.value = false;
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
// 升级策略
async function getRuntimeModuleParams() {
  if (!basicInfo.value.clusterBasicSettings.version) return {};
  moduleLoading.value = true;
  const data = await cloudVersionModules({
    $cloudId: props.cloudID,
    $version: basicInfo.value.clusterBasicSettings.version,
    $module: 'runtime',
  }).catch(() => []);
  moduleLoading.value = false;
  return data.length > 0 ? data[0] : {};
};

// 升级策略数据
const upgradeStrategyData = ref<any>({
  defaultValue: '',
});
const labelEnum = {
  EXTENDED: $i18n.t('cluster.create.aws.runtime.extended.label'),
  STANDARD: $i18n.t('cluster.create.aws.runtime.standard.label'),
};
const upgradEnum = {
  EXTENDED: $i18n.t('cluster.create.aws.runtime.extended.text'),
  STANDARD: $i18n.t('cluster.create.aws.runtime.standard.text'),
};
watch(() => basicInfo.value.clusterBasicSettings.version, async () => {
  const result = await getRuntimeModuleParams();
  upgradeStrategyData.value = result;
  basicInfo.value.clusterBasicSettings.upgradePolicy.supportType = upgradeStrategyData.value.defaultValue || '';
}, { immediate: true });
function handleUpgrade(type) {
  basicInfo.value.clusterBasicSettings.upgradePolicy.supportType = type;
  upgradeStrategyData.value = {
    ...upgradeStrategyData.value,
    defaultValue: type,
  };
}

onMounted(async () => {
  handleGetTemplateList();
  await handleGetAccounts();
  handleGetNodeManCloud();
  handleGetRoles();
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
