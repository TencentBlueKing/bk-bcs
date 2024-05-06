<template>
  <div class="flex items-center justify-center py-[24px]">
    <div class="shadow bg-[#fff] px-[24px] py-[16px] pb-[32px] min-w-[800px]">
      <div class="text-[14px] font-bold mb-[24px]">{{ $t('importAzureCloud.title') }}</div>
      <bk-form class="max-w-[640px]" :model="formData" :rules="formRules" ref="formRef">
        <bk-form-item
          :label="$t('importAzureCloud.label.clusterName')"
          property="clusterName"
          error-display-type="normal"
          required>
          <bcs-input v-model="formData.clusterName"></bcs-input>
        </bk-form-item>
        <bk-form-item
          :label="$t('importAzureCloud.label.token')"
          property="accountID" error-display-type="normal"
          required>
          <div class="flex">
            <bcs-select
              :loading="accountLoading"
              class="flex-1"
              :clearable="false"
              searchable
              v-model="formData.accountID">
              <bcs-option
                v-for="item in accountList"
                :key="item.account.accountID"
                :id="item.account.accountID"
                :name="item.account.accountName">
              </bcs-option>
              <template slot="extension">
                <SelectExtension
                  :link-text="$t('cluster.create.button.createCloudToken')"
                  :link="CloudTokenLink"
                  @refresh="handleGetAccountsList" />
              </template>
            </bcs-select>
          </div>
        </bk-form-item>
        <bk-form-item
          :label="$t('importAzureCloud.label.resourceGroups')" property="resourceGroupName"
          error-display-type="normal"
          required>
          <bcs-select
            :loading="resourceLoading"
            v-model="formData.resourceGroupName"
            searchable>
            <bcs-option
              v-for="data in resourceGroups"
              :key="data.name"
              :id="data.name"
              :name="data.name">
            </bcs-option>
          </bcs-select>
        </bk-form-item>
        <bk-form-item
          :label="$t('importAzureCloud.label.region')"
          property="region"
          error-display-type="normal"
          required>
          <bcs-select
            :loading="regionLoading"
            v-model="formData.region"
            searchable>
            <bcs-option-group v-for="item in regionList" :key="item.id" :name="item.name">
              <bcs-option
                v-for="data in item.children"
                :key="data.region"
                :id="data.region"
                :name="data.regionName">
              </bcs-option>
            </bcs-option-group>
          </bcs-select>
        </bk-form-item>
        <bk-form-item
          :label="$t('importAzureCloud.label.clusterID')"
          property="cloudID"
          error-display-type="normal"
          required>
          <bcs-select v-model="formData.cloudID" :loading="clusterLoading">
            <bcs-option-group v-for="item in clusterList" :key="item.id" :name="item.name">
              <bcs-option
                v-for="data in item.children"
                :key="data.clusterID"
                :id="data.clusterName"
                :name="data.clusterName">
              </bcs-option>
            </bcs-option-group>
          </bcs-select>
        </bk-form-item>
        <bk-form-item
          :label="$t('importAzureCloud.label.nodemanArea')"
          property="bkCloudID"
          error-display-type="normal"
          required>
          <bk-select
            searchable
            :clearable="false"
            :loading="nodemanCloudLoading"
            v-model="formData.bkCloudID">
            <bk-option
              v-for="item in nodemanCloudList"
              :key="item.bk_cloud_id"
              :id="item.bk_cloud_id"
              :name="item.bk_cloud_name">
            </bk-option>
          </bk-select>
        </bk-form-item>
        <bk-form-item :label="$t('importAzureCloud.label.desc')" property="description" error-display-type="normal">
          <bcs-input :maxlength="100" type="textarea" v-model="formData.description"></bcs-input>
        </bk-form-item>
        <bk-form-item>
          <bcs-badge
            :theme="isValidate ? 'success' : 'danger'"
            class="badge-icon"
            :icon="isValidate ? 'bk-icon icon-check-1' : 'bk-icon icon-close'"
            :visible="isValidate || !!validateErrMsg"
            v-bk-tooltips="{
              content: validateErrMsg,
              disabled: !validateErrMsg,
              theme: 'light'
            }">
            <bk-button
              theme="primary"
              :loading="testing"
              :outline="isValidate"
              @click="testClusterImport">
              {{ $t('importAzureCloud.button.test') }}
            </bk-button>
          </bcs-badge>
          <bk-button
            :loading="importLoading"
            :disabled="!isValidate"
            theme="primary"
            class="min-w-[88px] ml-[10px]"
            @click="handleImport">
            {{ $t('importAzureCloud.button.import') }}
          </bk-button>
          <bk-button class="ml-[10px]" @click="handleCancel">{{$t('generic.button.cancel')}}</bk-button>
        </bk-form-item>
      </bk-form>
    </div>
  </div>
</template>
<script setup lang="ts">
import { computed, onBeforeMount, ref, watch } from 'vue';

import { nodemanCloud } from '@/api/modules/cluster-manager';
import $bkMessage from '@/common/bkmagic';
import $i18n from '@/i18n/i18n-setup';
import $router from '@/router';
import $store from '@/store';
// 管控区域跳转链接
import SelectExtension from '@/views/cluster-manage/add/common/select-extension.vue';
import { INodeManCloud } from '@/views/cluster-manage/add/tencent/types';
import useCloud, { ICloudAccount } from '@/views/cluster-manage/use-cloud';

interface IGroup<T = any> {
  id: string
  name: string
  children: Array<T>
}
const { cloudAccounts, clusterConnect } = useCloud();
const formRef = ref<any>();
const formData = ref({
  clusterName: '',
  description: '',
  resourceGroupName: '',
  region: '',
  cloudID: '',
  bkCloudID: '',
  accountID: '',
});
const formRules = ref({
  clusterName: [
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
  resourceGroupName: [
    {
      required: true,
      message: $i18n.t('generic.validate.required'),
      trigger: 'blur',
    },
  ],
  cloudID: [
    {
      required: true,
      message: $i18n.t('generic.validate.required'),
      trigger: 'blur',
    },
  ],
  bkCloudID: [
    {
      required: true,
      message: $i18n.t('generic.validate.required'),
      trigger: 'blur',
    },
  ],
  accountID: [
    {
      required: true,
      message: $i18n.t('generic.validate.required'),
      trigger: 'blur',
    },
  ],
});

watch(() => formData.value.accountID, () => {
  formData.value.region = '';
  formData.value.cloudID = '';
});

watch(() => formData.value.region, () => {
  formData.value.cloudID = '';
});

watch(() => formData.value.cloudID, () => {
  isValidate.value = false;
  validateErrMsg.value = '';
});

// 云凭证
const accountLoading = ref(false);
const accountList = ref<ICloudAccount[]>([]);
const handleGetAccountsList = async () => {
  accountLoading.value  = true;
  const { data } = await cloudAccounts('azureCloud');
  accountList.value = data;
  accountLoading.value = false;
};
const CloudTokenLink = computed(() => {
  const { href } = $router.resolve({ name: 'azureCloud' });
  return href;
});
// 根据云凭证获取资源组
watch(() => formData.value.accountID, () => {
  getResourceGroups();
});
const resourceGroups = ref<Array<{
  name: string
  region: string
  provisioningState: string
}>>([]);
const resourceLoading = ref(false);
// 获取资源组
const getResourceGroups = async () => {
  if (!formData.value.accountID) return;
  resourceLoading.value = true;
  const data = await $store.dispatch('clustermanager/cloudResourceGroupByAccount', {
    $cloudId: 'azureCloud',
    accountID: formData.value.accountID,
  });
  resourceGroups.value = data;
  resourceLoading.value = false;
};
// 区域列表
watch(() => formData.value.accountID, () => {
  getRegionList();
});
const locationMap = {
  asia: $i18n.t('regions.asia'),
  australia: $i18n.t('regions.australia'),
  europe: $i18n.t('regions.europe'),
  northamerica: $i18n.t('regions.northamerica'),
  southamerica: $i18n.t('regions.southamerica'),
  us: $i18n.t('regions.us'),
};
const regionList = ref<Array<IGroup<{
  region: string
  regionName: string
}>>>([]);
const regionLoading = ref(false);
const getRegionList = async () => {
  if (!formData.value.accountID) return;
  regionLoading.value = true;
  const data = await $store.dispatch('clustermanager/cloudRegionByAccount', {
    $cloudId: 'azureCloud',
    accountID: formData.value.accountID,
  });
  regionList.value = data.reduce((pre, item) => {
    const location = item.region.split('-')?.[0];
    if (location && locationMap[location]) {
      const group = pre.find(item => item.id === location);
      if (group) {
        group.children.push(item);
      } else {
        pre.push({
          id: location,
          name: locationMap[location],
          children: [item],
        });
      }
    } else {
      const group = pre.find(item => item.id === 'other');
      group.children.push(item);
    }
    return pre;
  }, [{
    id: 'other',
    name: $i18n.t('regions.other'),
    children: [],
  }]);
  regionLoading.value = false;
};
// 集群列表
watch(() => [
  formData.value.accountID,
  formData.value.resourceGroupName,
  formData.value.region,
], () => {
  handleGetClusterList();
});
const clusterLoading = ref(false);
const clusterList = ref<Array<IGroup>>([]);
const currentCluster = computed(() => {
  const list = clusterList.value.reduce<any[]>((pre, item) => {
    pre.push(...item.children);
    return pre;
  }, []);
  return list.find(item => item.clusterID === formData.value.cloudID);
});
const levelMap = {
  zones: $i18n.t('importAzureCloud.clusterRegion.zones'),
  regions: $i18n.t('importAzureCloud.clusterRegion.regions'),
  other: $i18n.t('importAzureCloud.clusterRegion.other'),
};
const handleGetClusterList = async () => {
  if (!formData.value.region || !formData.value.accountID || !formData.value.resourceGroupName) return;
  clusterLoading.value = true;
  const data = await $store.dispatch('clustermanager/cloudClusterList', {
    region: formData.value.region,
    $cloudId: 'azureCloud',
    accountID: formData.value.accountID,
    resourceGroupName: formData.value.resourceGroupName,
  });
  clusterList.value = data.reduce((pre, item) => {
    const group = pre.find(data => data.id === item.clusterLevel);
    if (group) {
      group.children.push(item);
    } else {
      pre.push({
        id: item.clusterLevel,
        name: levelMap[item.clusterLevel] || levelMap.other,
        children: [item],
      });
    }
    return pre;
  }, []);
  clusterLoading.value = false;
};

// 管控区域
const nodemanCloudList = ref<Array<INodeManCloud>>([]);
const nodemanCloudLoading = ref(false);
const handleGetNodeManCloud = async () => {
  nodemanCloudLoading.value = true;
  nodemanCloudList.value = await nodemanCloud().catch(() => []);
  nodemanCloudLoading.value = false;
};

// 导入测试
const isValidate = ref(false);
const validateErrMsg = ref('');
const testing = ref(false);
const testClusterImport = async () => {
  if (!formData.value.accountID || !currentCluster.value || !formData.value.resourceGroupName) return;
  testing.value = true;
  validateErrMsg.value = await clusterConnect({
    $cloudId: 'azureCloud',
    $clusterID: currentCluster.value.clusterID,
    isExtranet: true,
    accountID: formData.value.accountID,
    resourceGroupName: formData.value.resourceGroupName,
    // resourceGroups: currentCluster.value.resourceGroups,
    region: currentCluster.value.location,
  });
  isValidate.value = !validateErrMsg.value;
  testing.value = false;
};

// 集群导入
const curProject = computed(() => $store.state.curProject);
const user = computed(() => $store.state.user);
const importLoading = ref(false);
const handleImport = async () => {
  const validate = await formRef.value.validate().catch(() => false);
  if (!validate) return;

  importLoading.value = true;
  const params = {
    clusterName: formData.value.clusterName,
    description: formData.value.description,
    provider: 'azureCloud',
    resourceGroupID: currentCluster.value.resourceGroups,
    resourceGroupName: formData.value.resourceGroupName,
    region: currentCluster.value.location,
    projectID: curProject.value.projectID,
    businessID: String(curProject.value.businessID),
    environment: 'prod',
    engineType: 'k8s',
    isExclusive: true,
    creator: user.value.username,
    cloudMode: {
      cloudID: currentCluster.value.clusterID,
      resourceGroup: formData.value.resourceGroupName,
    },
    clusterCategory: 'importer',
    networkType: 'overlay',
    area: {
      bkCloudID: formData.value.bkCloudID,
    },
    accountID: formData.value.accountID,
  };
  const result = await $store.dispatch('clustermanager/importCluster', params);
  importLoading.value = false;
  if (result) {
    $bkMessage({
      theme: 'success',
      message: $i18n.t('generic.msg.success.import'),
    });
    $router.push({ name: 'clusterMain' });
  }
};
// 取消
const handleCancel = () => {
  $router.back();
};

onBeforeMount(() => {
  handleGetAccountsList();
  handleGetNodeManCloud();
});
</script>
<style lang="postcss" scoped>
.refresh {
  font-size: 12px;
  color: #3a84ff;
  cursor: pointer;
  position: absolute;
  right: -20px
}

>>> .badge-icon {
  .bk-badge.bk-danger {
    background-color: #ff5656 !important;
  }
  .bk-badge.bk-success {
    background-color: #2dcb56 !important;
  }
  .bk-icon {
    color: #fff;
  }
  .bk-badge {
    min-width: 14px !important;
    height: 14px !important;
  }
}
</style>
