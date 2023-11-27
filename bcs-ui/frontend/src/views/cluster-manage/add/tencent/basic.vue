<template>
  <bk-form
    :model="basicInfo"
    :rules="basicInfoRules"
    ref="formRef"
    class="k8s-form grid grid-cols-2 grid-rows-1 gap-[16px]">
    <DescList size="middle" :title="$t('cluster.create.label.clusterInfo')">
      <bk-form-item :label="$t('cluster.create.label.kubernetesProvider')">
        {{ $t('tke.label.tencentCloud') }}
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
        :label="$t('tke.label.project.text')"
        :desc="$t('tke.label.project.desc')"
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
            :name="item.projectID !== 0 ? `${item.projectName}(${item.projectID})` : item.projectName">
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
          <bk-select
            v-model="basicInfo.region"
            searchable
            :clearable="false"
            :loading="regionLoading">
            <bk-option
              v-for="item in regionList"
              :key="item.region"
              :id="item.region"
              :name="item.regionName">
            </bk-option>
          </bk-select>
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
          :label="$t('cluster.create.label.system')"
          :desc="systemDesc"
          property="clusterBasicSettings.OS"
          error-display-type="normal"
          required>
          <bcs-select searchable :clearable="false" :loading="osLoading" v-model="imageID" @change="handleImageChange">
            <bcs-option-group
              v-for="group in imageListByGroup"
              :key="group.provider"
              :name="group.name"
              :is-collapse="collapseList.includes(group.provider)"
              :class="[
                'mt-[8px]'
              ]">
              <template #group-name>
                <CollapseTitle
                  :title="`${group.name} (${group.children.length})`"
                  :collapse="collapseList.includes(group.provider)"
                  @click="handleToggleCollapse(group.provider)" />
              </template>
              <bcs-option
                v-for="item in group.children"
                :key="item.imageID"
                :id="item.imageID"
                :name="item.alias">
                <div
                  class="flex items-center justify-between"
                  v-bk-tooltips="{
                    content: item.clusters ? item.clusters.join(',') : '--',
                    disabled: !item.clusters.length
                  }">
                  <span class="flex items-center">
                    {{ `${item.alias}(${item.imageID})` }}
                    <bcs-tag
                      theme="info"
                      radius="45px"
                      v-if="item.alias === 'TencentOS Server 3.1 (TK4)' && item.provider === 'PUBLIC_IMAGE'">
                      {{ $t('推荐') }}
                    </bcs-tag>
                  </span>
                  <span v-if="item.clusters.length" class="text-[#979BA5]">
                    {{ $t('tke.tips.imageUsedInCluster') }}
                  </span>
                </div>
              </bcs-option>
            </bcs-option-group>
          </bcs-select>
          <div id="systemDesc">
            <i18n path="tke.image.desc.public" tag="div">
              <bk-link
                theme="primary"
                href="https://cloud.tencent.com/document/product/213/4941"
                target="_blank">
                <span class="!text-[12px]">{{ $t('tke.image.link.imageType') }}</span>
              </bk-link>
            </i18n>
            <i18n path="tke.image.desc.market" tag="div">
              <bk-link
                theme="primary"
                href="https://cloud.tencent.com/document/product/457/61448#.E6.93.8D.E4.BD.9C.E6.AD.A5.E9.AA.A4"
                target="_blank">
                <span class="!text-[12px]">{{ $t('tke.image.link.qGPU') }}</span>
              </bk-link>
            </i18n>
            <i18n path="tke.image.desc.private" tag="div">
              <bk-link
                theme="primary"
                href="https://cloud.tencent.com/document/product/457/39563"
                target="_blank">
                <span class="!text-[12px]">{{ $t('tke.image.link.privateImage') }}</span>
              </bk-link>
            </i18n>
          </div>
        </bk-form-item>
        <bk-form-item
          :label="$t('cluster.create.label.containerRuntime')"
          property="clusterBasicSettings.containerRuntime"
          error-display-type="normal"
          required>
          <bk-radio-group
            v-model="basicInfo.clusterAdvanceSettings.containerRuntime"
            @change="handleRuntimeChange">
            <bk-radio
              value="containerd"
              :disabled="!runtimeModuleParamsMap['containerd']"
            >
              <span v-bk-tooltips="{
                content: $t('当前集群版本不支持'),
                disabled: runtimeModuleParamsMap['containerd']
              }">
                containerd
              </span>
            </bk-radio>
            <bk-radio
              value="docker"
              :disabled="!runtimeModuleParamsMap['docker']"
            >
              <span v-bk-tooltips="{
                content: $t('当前集群版本不支持'),
                disabled: runtimeModuleParamsMap['docker']
              }">
                docker
              </span>
            </bk-radio>
          </bk-radio-group>
        </bk-form-item>
        <bk-form-item
          :label="$t('cluster.create.label.runtimeVersion')"
          property="clusterBasicSettings.runtimeVersion"
          error-display-type="normal"
          required>
          <bk-select
            searchable
            :clearable="false"
            v-model="basicInfo.clusterAdvanceSettings.runtimeVersion"
            class="max-w-[50%]">
            <bk-option v-for="item in runtimeVersionList" :key="item" :id="item" :name="item"></bk-option>
          </bk-select>
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

import KeyValue from '../../components/key-value.vue';

import SelectExtension from './select-extension.vue';
import { ICloudProject, ICloudRegion, IImageGroup, IImageItem, INodeManCloud } from './types';

import {
  cloudDetail,
  cloudOsImage,
  cloudProjects,
  cloudRegionByAccount,
  cloudVersionModules,
  nodemanCloud,
} from '@/api/modules/cluster-manager';
import { LABEL_KEY_REGEXP } from '@/common/constant';
import CollapseTitle from '@/components/cluster-selector/collapse-title.vue';
import DescList from '@/components/desc-list.vue';
import $i18n from '@/i18n/i18n-setup';
import $router from '@/router';
import useCloud, { CloudID, ICloudAccount } from '@/views/project-manage/cloudtoken/use-cloud';

const props = defineProps({
  cloudId: {
    type: String as PropType<CloudID>,
    default: '',
  },
});

const emits = defineEmits(['next', 'cancel', 'set-image-group', 'set-region-list']);

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
    cloudProjectId: 0,
  },
});
watch(() => basicInfo.value.cloudAccountID, () => {
  basicInfo.value.extraInfo.cloudProjectId = 0;
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
    $cloudId: props.cloudId,
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
  const { data } = await cloudAccounts(props.cloudId);
  accountList.value = data;
  // 初始化第一个云凭证
  basicInfo.value.cloudAccountID = accountList.value?.[0]?.account?.accountID || '';
  accountLoading.value = false;
};
const gotoTencentToken = () => {
  const { href } = $router.resolve({ name: 'tencentCloud' });
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
    $cloudId: props.cloudId,
  }).catch(() => []);
  projectList.value.unshift({
    projectID: 0,
    projectName: $i18n.t('默认项目'),
  });
  projectLoading.value = false;
};
const gotoTencentProject = () => {
  window.open('https://console.cloud.tencent.com/project');
};

// 版本列表
const versionList = computed(() => templateDetail.value?.clusterManagement?.availableVersion || []);

// 区域列表
const regionLoading = ref(false);
const regionList = ref<Array<ICloudRegion>>([]);
const handleGetRegionList = async () => {
  if (!basicInfo.value.cloudAccountID) return;

  regionLoading.value = true;
  const data = await cloudRegionByAccount({
    $cloudId: props.cloudId,
    accountID: basicInfo.value.cloudAccountID,
  }).catch(() => []);
  regionList.value = data.filter(item => item.regionState === 'AVAILABLE');
  emits('set-region-list', regionList.value);
  regionLoading.value = false;
};

// 管控区域
const nodemanCloudList = ref<Array<INodeManCloud>>([]);
const nodemanCloudLoading = ref(false);
const handleGetNodeManCloud = async () => {
  nodemanCloudLoading.value = true;
  nodemanCloudList.value = await nodemanCloud().catch(() => []);
  nodemanCloudLoading.value = false;
};

// 运行时组件变更
const handleRuntimeChange = (flagName: string) => {
  basicInfo.value.clusterAdvanceSettings.runtimeVersion = runtimeModuleParamsMap.value[flagName]?.defaultValue;
};

const moduleLoading = ref(false);
const runtimeModuleParams = ref<IRuntimeModuleParams[]>([]);
const runtimeModuleParamsMap = computed<Record<string, IRuntimeModuleParams>>(() => runtimeModuleParams.value
  .reduce((pre, item) => {
    pre[item.flagName] = item;
    return pre;
  }, {}));
const runtimeVersionList = computed(() => {
  // 运行时版本
  const params = runtimeModuleParams.value
    .find(item => item.flagName === basicInfo.value.clusterAdvanceSettings.containerRuntime);
  return params?.flagValueList;
});
const getRuntimeModuleParams = async () => {
  if (!basicInfo.value.clusterBasicSettings.version) return;
  moduleLoading.value = true;
  const data = await cloudVersionModules({
    $cloudId: props.cloudId,
    $version: basicInfo.value.clusterBasicSettings.version,
    $module: 'runtime',
  });
  runtimeModuleParams.value = data.filter(item => item.enable);
  // 初始化默认运行时
  basicInfo.value.clusterAdvanceSettings.containerRuntime = runtimeModuleParams.value?.[0]?.flagName || '';
  basicInfo.value.clusterAdvanceSettings.runtimeVersion = runtimeModuleParams.value?.[0]?.defaultValue || '';
  moduleLoading.value = false;
};

// 操作系统
const imageID = ref('');
const handleImageChange = (imageID: string) => {
  const imageItem = imageList.value.find(item => item.imageID === imageID);
  if (!imageItem) return;

  if (imageItem.provider === 'PRIVATE_IMAGE') {
    basicInfo.value.clusterBasicSettings.OS = imageItem.imageID;
  } else {
    basicInfo.value.clusterBasicSettings.OS = imageItem.osName;
  }
};
const systemDesc = ref({
  allowHTML: true,
  content: '#systemDesc',
});
const collapseList = ref<string[]>([]);
const handleToggleCollapse = (provider: string) => {
  const index = collapseList.value.findIndex(item => item === provider);
  if (index > -1) {
    collapseList.value.splice(index, 1);
  } else {
    collapseList.value.push(provider);
  }
};
const imageList = ref<Array<IImageItem>>([]);
const providerMap = {
  PUBLIC_IMAGE: $i18n.t('tke.label.publicImage'),
  MARKET_IMAGE: $i18n.t('tke.label.marketImage'),
  PRIVATE_IMAGE: $i18n.t('tke.label.privateImage'),
};
const imageListByGroup = computed<Record<string, IImageGroup>>(() => imageList.value
  .reduce((pre, item) => {
    if (!pre[item.provider]) {
      pre[item.provider] = {
        name: providerMap[item.provider],
        provider: item.provider,
        children: [item],
      };
    } else {
      pre[item.provider].children.push(item);
    }
    return pre;
  }, {}));
const osLoading = ref(false);
const handleGetOsList = async () => {
  if (!basicInfo.value.region || !basicInfo.value.cloudAccountID) return;
  osLoading.value = true;
  imageList.value = await cloudOsImage({
    $cloudId: props.cloudId,
    accountID: basicInfo.value.cloudAccountID,
    region: basicInfo.value.region,
    provider: 'ALL',
  }).catch(() => []);
  // 设置默认镜像
  imageID.value = imageList.value
    .find(item => item.alias === 'TencentOS Server 3.1 (TK4)' && item.provider === 'PUBLIC_IMAGE')?.imageID || '';
  basicInfo.value.clusterBasicSettings.OS = imageID.value;
  emits('set-image-group', imageListByGroup.value);
  osLoading.value = false;
};

watch(() => basicInfo.value.cloudAccountID, () => {
  handleGetProjects();
  handleGetRegionList();
});

// 运行时组件参数
watch(() => basicInfo.value.clusterBasicSettings.version, () => {
  getRuntimeModuleParams();
});

watch([
  () => basicInfo.value.region,
  () => basicInfo.value.cloudAccountID,
], () => {
  handleGetOsList();
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
</script>
<style scoped lang="postcss">
/deep/ .bk-option-group-name {
  border-bottom: 0 !important;
}
</style>
