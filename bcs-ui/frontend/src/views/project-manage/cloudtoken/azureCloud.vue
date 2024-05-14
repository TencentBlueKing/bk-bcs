<template>
  <div class="p-5">
    <div class="flex items-center justify-between">
      <bk-button
        theme="primary"
        v-authority="{
          actionId: 'cloud_account_create',
          resourceName: curProject.project_name,
          permCtx: {
            resource_type: 'project',
            project_id: curProject.project_id,
            operator: user.username
          }
        }"
        @click="handleShowCreateDialog">{{$t('azureCloud.button.create')}}</bk-button>
      <bk-input
        class="w-[400px]"
        :placeholder="$t('azureCloud.placeholder.search')"
        right-icon="bk-icon icon-search"
        clearable
        v-model="searchValue">
      </bk-input>
    </div>
    <bcs-table
      class="mt20"
      :data="curPageData"
      :pagination="pagination"
      v-bkloading="{ isLoading: loading }"
      @page-change="pageChange"
      @page-limit-change="pageSizeChange">
      <bcs-table-column label="ID" width="220" show-overflow-tooltip>
        <template #default="{ row }">
          {{ row.account.accountID }}
        </template>
      </bcs-table-column>
      <bcs-table-column :label="$t('generic.label.name')" show-overflow-tooltip>
        <template #default="{ row }">
          {{ row.account.accountName }}
        </template>
      </bcs-table-column>
      <bcs-table-column :label="$t('cluster.create.label.desc')" show-overflow-tooltip>
        <template #default="{ row }">
          {{ row.account.desc || '--' }}
        </template>
      </bcs-table-column>
      <bcs-table-column label="SubscriptionID" show-overflow-tooltip>
        <template #default="{ row }">
          {{ row.account.account.subscriptionID }}
        </template>
      </bcs-table-column>
      <bcs-table-column label="TenantID">
        <template #default="{ row }">
          {{ row.account.account.tenantID }}
        </template>
      </bcs-table-column>
      <bcs-table-column label="ClientID" show-overflow-tooltip>
        <template #default="{ row }">
          {{ row.account.account.clientID }}
        </template>
      </bcs-table-column>
      <bcs-table-column label="ClientSecret">
        <template #default="{ row }">
          {{ row.account.account.clientSecret }}
        </template>
      </bcs-table-column>
      <bcs-table-column :label="$t('azureCloud.label.cluster')" show-overflow-tooltip>
        <template #default="{ row }">
          {{ row.clusters.join(',') || '--' }}
        </template>
      </bcs-table-column>
      <bcs-table-column :label="$t('cluster.labels.createdAt')" show-overflow-tooltip>
        <template #default="{ row }">
          {{ row.account.updateTime }}
        </template>
      </bcs-table-column>
      <bcs-table-column :label="$t('generic.label.action')" width="100">
        <template #default="{ row }">
          <bk-button
            text
            v-authority="{
              clickable: webAnnotations.perms[row.account.accountID]
                && webAnnotations.perms[row.account.accountID].cloud_account_manage,
              actionId: 'cloud_account_manage',
              resourceName: row.account.accountName,
              disablePerms: true,
              permCtx: {
                project_id: row.account.projectID,
                account_id: row.account.accountID,
                operator: user.username
              }
            }"
            @click="handleDeleteAccount(row)">{{$t('generic.button.delete')}}</bk-button>
        </template>
      </bcs-table-column>
      <template #empty>
        <BcsEmptyTableStatus :type="searchValue ? 'search-empty' : 'empty'" @clear="searchValue = ''" />
      </template>
    </bcs-table>
    <bcs-dialog
      v-model="showDialog"
      theme="primary"
      :mask-close="false"
      header-position="left"
      width="600"
      :title="$t('azureCloud.label.create')">
      <bk-form :label-width="130" :model="account" :rules="formRules" ref="formRef">
        <bk-form-item :label="$t('generic.label.name')" property="accountName" error-display-type="normal" required>
          <bk-input :maxlength="64" v-model="account.accountName"></bk-input>
        </bk-form-item>
        <bk-form-item :label="$t('cluster.create.label.desc')">
          <bk-input :maxlength="256" type="textarea" v-model="account.desc"></bk-input>
        </bk-form-item>
        <bk-form-item label="SubscriptionID" property="account.subscriptionID" error-display-type="normal" required>
          <bk-input :maxlength="64" v-model="account.account.subscriptionID"></bk-input>
        </bk-form-item>
        <bk-form-item label="TenantID" property="account.tenantID" error-display-type="normal" required>
          <bk-input :maxlength="64" v-model="account.account.tenantID"></bk-input>
        </bk-form-item>
        <bk-form-item label="ClientID" property="account.clientID" error-display-type="normal" required>
          <bk-input :maxlength="64" v-model="account.account.clientID"></bk-input>
        </bk-form-item>
        <bk-form-item label="ClientSecret" property="account.clientSecret" error-display-type="normal" required>
          <bk-input :maxlength="64" v-model="account.account.clientSecret"></bk-input>
        </bk-form-item>
      </bk-form>
      <template #footer>
        <div>
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
              :loading="validating"
              :outline="isValidate"
              @click="handleValidate">
              {{ $t('generic.button.validate') }}
            </bk-button>
          </bcs-badge>
          <bk-button
            theme="primary"
            :loading="createLoading"
            :disabled="!isValidate"
            @click="handleCreateAccount">
            {{ $t('generic.button.confirm') }}
          </bk-button>
          <bk-button @click="showDialog = false">{{ $t('generic.button.cancel') }}</bk-button>
        </div>
      </template>
    </bcs-dialog>
  </div>
</template>
<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue';

import $bkMessage from '@/common/bkmagic';
import { NAME_REGEX, SECRET_REGEX } from '@/common/constant';
import $bkInfo from '@/components/bk-magic-2.0/bk-info';
import usePage from '@/composables/use-page';
import useTableSearch from '@/composables/use-search';
import $i18n from '@/i18n/i18n-setup';
import $store from '@/store';
import useCloud from '@/views/cluster-manage/use-cloud';

const cloudID = 'azureCloud';
const curProject = computed(() => $store.state.curProject);
const user = computed(() => $store.state.user);

const loading = ref(false);
const createLoading = ref(false);
const showDialog = ref(false);
const data = ref([]);
const formRef = ref<any>();
const account = ref({
  accountName: '',
  desc: '',
  account: {
    subscriptionID: '',
    tenantID: '',
    clientID: '',
    clientSecret: '',
  },
  enable: true,
  creator: user.value.username,
  projectID: curProject.value.project_id,
});
watch(showDialog, () => {
  if (!showDialog.value) {
    // 重置数据
    account.value = {
      accountName: '',
      desc: '',
      account: {
        subscriptionID: '',
        tenantID: '',
        clientID: '',
        clientSecret: '',
      },
      enable: true,
      creator: user.value.username,
      projectID: curProject.value.project_id,
    };
  }
});
const formRules = ref({
  accountName: [
    {
      required: true,
      message: $i18n.t('generic.validate.required'),
      trigger: 'blur',
    },
    {
      message: $i18n.t('azureCloud.tips.nameRegex'),
      trigger: 'blur',
      validator(val) {
        return new RegExp(NAME_REGEX, 'g').test(val);
      },
    },
  ],
  'account.subscriptionID': [
    {
      required: true,
      message: $i18n.t('generic.validate.required'),
      trigger: 'blur',
    },
    {
      message: $i18n.t('azureCloud.tips.nameRegex'),
      trigger: 'blur',
      validator(val) {
        return new RegExp(NAME_REGEX, 'g').test(val);
      },
    },
  ],
  'account.tenantID': [
    {
      required: true,
      message: $i18n.t('generic.validate.required'),
      trigger: 'blur',
    },
    {
      message: $i18n.t('azureCloud.tips.nameRegex'),
      trigger: 'blur',
      validator(val) {
        return new RegExp(NAME_REGEX, 'g').test(val);
      },
    },
  ],
  'account.clientID': [
    {
      required: true,
      message: $i18n.t('generic.validate.required'),
      trigger: 'blur',
    },
    {
      message: $i18n.t('azureCloud.tips.nameRegex'),
      trigger: 'blur',
      validator(val) {
        return new RegExp(NAME_REGEX, 'g').test(val);
      },
    },
  ],

  'account.clientSecret': [
    {
      required: true,
      message: $i18n.t('generic.validate.required'),
      trigger: 'blur',
    },
    {
      message: $i18n.t('azureCloud.tips.secretRegex'),
      trigger: 'blur',
      validator(val) {
        return new RegExp(SECRET_REGEX, 'g').test(val);
      },
    },
  ],
});
const webAnnotations = ref({ perms: {} });
const keys = ref(['account.accountName', 'account.account.subscriptionID', 'account.account.tenantID',
  'account.account.clientID', 'clusters']); // 模糊搜索字段
const { tableDataMatchSearch, searchValue } = useTableSearch(data, keys);
const { pageChange, pageSizeChange, curPageData, pagination } = usePage(tableDataMatchSearch);

const handleGetCloud = async () => {
  loading.value = true;
  const res = await $store.dispatch('clustermanager/cloudAccounts', {
    $cloudId: cloudID,
    projectID: curProject.value.project_id,
    operator: user.value.username,
  });
  data.value = res.data;
  webAnnotations.value = res.web_annotations || { perms: {} };
  loading.value = false;
};
const handleDeleteAccount = (row) => {
  $bkInfo({
    type: 'warning',
    clsName: 'custom-info-confirm',
    subTitle: row.account.accountID,
    title: $i18n.t('azureCloud.button.delete'),
    defaultInfo: true,
    confirmFn: async () => {
      loading.value = true;
      const result = await $store.dispatch('clustermanager/deleteCloudAccounts', {
        $cloudId: cloudID,
        $accountID: row.account.accountID,
      });
      if (result) {
        $bkMessage({
          theme: 'success',
          message: $i18n.t('generic.msg.success.delete'),
        });
        await handleGetCloud();
      }
      loading.value = false;
    },
  });
};
// 校验云凭证
const { validateCloudAccounts } = useCloud();
const isValidate = ref(false);
const validateErrMsg = ref('');
const validating = ref(false);
watch(() => account.value.account, () => {
  isValidate.value = false;
}, { deep: true });
const handleValidate = async () => {
  const valid = await formRef.value?.validate().catch(() => false);
  if (!valid) return;

  validating.value = true;
  const errMsg = await validateCloudAccounts({
    $cloudId: cloudID,
    account: account.value.account,
  });
  validateErrMsg.value = errMsg ?? '';
  isValidate.value = !errMsg;
  validating.value = false;
};
// 创建云凭证
const handleCreateAccount = async () => {
  const valid = await formRef.value?.validate();
  if (!valid) return;

  createLoading.value = true;
  const result = await $store.dispatch('clustermanager/createCloudAccounts', {
    $cloudId: cloudID,
    ...account.value,
  });
  createLoading.value = false;
  if (!result) return;

  showDialog.value = false;
  $bkMessage({
    theme: 'success',
    message: $i18n.t('generic.msg.success.create'),
  });
  handleGetCloud();
};
const handleShowCreateDialog = () => {
  showDialog.value = true;
};
onMounted(() => {
  handleGetCloud();
});
</script>
<style lang="postcss" scoped>
/deep/ .bk-form-content .form-error-tip {
    text-align: left;
}
</style>
