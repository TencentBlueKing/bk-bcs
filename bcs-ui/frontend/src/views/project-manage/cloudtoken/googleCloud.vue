<template>
  <div class="px-[20px]">
    <Row class="mt-[20px]">
      <template #left>
        <bcs-button theme="primary" icon="plus" @click="showCreate = true">
          {{ $t('googleCloud.button.create') }}
        </bcs-button>
      </template>
      <template #right>
        <bcs-input
          class="w-[400px]"
          :placeholder="$t('googleCloud.placeholder.search')"
          right-icon="bk-icon icon-search"
          clearable
          v-model="searchValue">
        </bcs-input>
      </template>
    </Row>
    <bcs-table
      :data="curPageData"
      :pagination="pagination"
      v-bkloading="{ isLoading: loading }"
      size="medium"
      class="mt-[20px]"
      :row-class-name="getRowClassName"
      @row-mouse-enter="handleRowEnter"
      @row-mouse-leave="handleRowLeave"
      @page-change="pageChange"
      @page-limit-change="pageSizeChange">
      <bcs-table-column
        :label="$t('googleCloud.label.projectID')"
        prop="account.accountID"
        show-overflow-tooltip
        width="200">
        <template #default="{ row }">
          <span>{{ JSON.parse(row.account.account.serviceAccountSecret).project_id || '--' }}</span>
        </template>
      </bcs-table-column>
      <bcs-table-column
        :label="$t('googleCloud.label.tokenName')"
        prop="account.accountName"
        show-overflow-tooltip>
      </bcs-table-column>
      <bcs-table-column :label="$t('googleCloud.label.desc')" min-width="180" show-overflow-tooltip>
        <template #default="{ row, $index }">
          <div class="flex items-start py-[6px]">
            <bcs-input
              type="textarea"
              :value="row.account.desc"
              v-if="editIndex === $index"
              @change="handleDescChange">
            </bcs-input>
            <pre class="m-[0px]" v-else>{{ row.account.desc || '--' }}</pre>
            <span
              class="text-[#3a84ff] cursor-pointer ml-[8px]"
              v-if="hoverIndex === $index && editIndex === -1" @click="editDesc($index)">
              <i class="bk-icon icon-edit-line"></i>
            </span>
            <span v-else-if="editIndex === $index" class="flex items-center ml-[8px] mt-[6px]">
              <i
                class="bcs-icon bcs-icon-check-1 text-[#2DCB56] font-bold cursor-pointer"
                @click="handleUpdateDesc(row)"></i>
              <i class="bcs-icon bcs-icon-close-5 text-[#979BA5] ml-[8px] cursor-pointer" @click="editIndex = -1"></i>
            </span>
          </div>
        </template>
      </bcs-table-column>
      <bcs-table-column :label="$t('googleCloud.label.tokenID')" show-overflow-tooltip>
        <template #default="{ row }">
          {{ getAccountSecretData(row, 'private_key_id') }}
        </template>
      </bcs-table-column>
      <bcs-table-column :label="$t('googleCloud.label.oAuth2ID')" show-overflow-tooltip>
        <template #default="{ row }">
          {{ getAccountSecretData(row, 'client_id') }}
        </template>
      </bcs-table-column>
      <bcs-table-column :label="$t('googleCloud.label.cluster')" show-overflow-tooltip>
        <template #default="{ row }">
          {{ row.clusters.join(',') || '--' }}
        </template>
      </bcs-table-column>
      <bcs-table-column
        :label="$t('googleCloud.label.createdAt')"
        prop="account.creatTime"
        show-overflow-tooltip>
      </bcs-table-column>
      <bcs-table-column
        :label="$t('googleCloud.label.createdBy')"
        prop="account.creator"
        show-overflow-tooltip>
        <template #default="{ row }">
          <bk-user-display-name v-if="row.account?.creator" :user-id="row.account?.creator">
          </bk-user-display-name>
          <span v-else>--</span>
        </template>
      </bcs-table-column>
      <bcs-table-column :label="$t('googleCloud.label.operate')" width="100" show-overflow-tooltip>
        <template #default="{ row }">
          <span
            v-bk-tooltips="{
              content: $t('importGoogleCloud.tips.disabledTips'),
              disabled: !row.clusters.length
            }">
            <bcs-button
              text
              :disabled="!!row.clusters.length"
              @click="handleDeleteToken(row)">
              {{ $t('generic.button.delete') }}
            </bcs-button>
          </span>
        </template>
      </bcs-table-column>
      <template #empty>
        <BcsEmptyTableStatus :type="searchValue ? 'search-empty' : 'empty'" @clear="searchValue = ''">
          <bcs-button
            theme="primary"
            text
            class="mt-[10px]"
            @click="showCreate = true">
            {{ $t('importGoogleCloud.button.create') }}
          </bcs-button>
        </BcsEmptyTableStatus>
      </template>
    </bcs-table>
    <!-- 创建 -->
    <bcs-sideslider
      :is-show.sync="showCreate"
      :width="640"
      :before-close="handleBeforeClose"
      :title="$t('googleCloud.button.create')"
      quick-close
      @hidden="showCreate = false">
      <div slot="content" class="px-[40px] pt-[24px] h-[calc(100vh-112px)] overflow-auto">
        <bk-form form-type="vertical" :model="formData" :rules="formRules" ref="formRef">
          <bk-form-item
            :label="$t('googleCloud.label.gcpFile')"
            error-display-type="normal"
            property="account.serviceAccountSecret"
            required>
            <div
              class="flex items-center justify-between h-[40px] bg-[#2E2E2E] shadow pl-[24px] pr-[16px]"
              ref="editorRef">
              <span class="text-[#979BA5] text-[14px] flex items-center">
                <i class="bk-icon icon-info-circle"></i>
                <span class="ml-[10px]">{{ $t('googleCloud.tips.onlySupportJsonFile') }}</span>
              </span>
              <span class="text-[#979BA5] flex items-center">
                <!-- 导入 -->
                <span class="text-[20px] relative">
                  <i class="bk-icon icon-upload-cloud"></i>
                  <input
                    class="absolute top-0 left-0 w-full h-full opacity-0 cursor-pointer z-10"
                    type="file"
                    ref="fileRef"
                    tabindex="-1"
                    accept=".json"
                    @change="handleInputFile">
                </span>
                <!-- 全屏 -->
                <!-- <span class="cursor-pointer ml-[16px]" @click="handleFullScreen">
                  <i class="bcs-icon bcs-icon-enlarge"></i>
                </span> -->
              </span>
            </div>
            <CodeEditor
              ref="codeEditorRef"
              lang="json"
              :height="360"
              :key="editorKey"
              v-model="formData.account.serviceAccountSecret" />
          </bk-form-item>
          <bk-form-item
            :label="$t('googleCloud.label.tokenName')"
            property="accountName"
            class="!mt-[24px]"
            error-display-type="normal"
            required>
            <bcs-input v-model="formData.accountName" />
          </bk-form-item>
          <bk-form-item
            :label="$t('googleCloud.label.desc')"
            property="desc"
            class="!mt-[24px]"
            error-display-type="normal">
            <bcs-input type="textarea" v-model="formData.desc" :rows="4" />
          </bk-form-item>
        </bk-form>
      </div>
      <div slot="footer" class="flex items-center h-[52px] bg-[#FAFBFD] px-[24px] w-full footer">
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
          <bcs-button
            theme="primary"
            class="min-w-[88px]"
            :loading="validating"
            :outline="isValidate"
            @click="handleValidateFile">{{ $t('googleCloud.button.validate') }}</bcs-button>
        </bcs-badge>

        <bcs-button
          theme="primary"
          class="min-w-[88px] ml-[10px]"
          :loading="creating"
          :disabled="!isValidate"
          @click="handleCreateToken">
          {{ $t('googleCloud.button.create1') }}
        </bcs-button>
        <bcs-button class="ml-[10px]" @click="showCreate = false">{{ $t('googleCloud.button.cancel') }}</bcs-button>
      </div>
    </bcs-sideslider>
  </div>
</template>
<script setup lang="ts">
import { onMounted, ref, watch } from 'vue';

import $bkMessage from '@/common/bkmagic';
// import { exitFullscreen, fullScreen } from '@/common/util';
import $bkInfo from '@/components/bk-magic-2.0/bk-info';
import Row from '@/components/layout/Row.vue';
import CodeEditor from '@/components/monaco-editor/new-editor.vue';
import usePage from '@/composables/use-page';
import useTableSearch from '@/composables/use-search';
import useSideslider from '@/composables/use-sideslider';
import $i18n from '@/i18n/i18n-setup';
import useCloud, { ICloudAccount } from '@/views/cluster-manage/use-cloud';

const {
  cloudAccounts,
  updateCloudAccounts,
  validateCloudAccounts,
  createCloudAccounts,
  deleteCloudAccounts,
} = useCloud();

const formRef = ref<any>();
const formData = ref({
  accountName: '',
  desc: '',
  account: {
    serviceAccountSecret: '',
  },
});

const formRules = ref({
  'account.serviceAccountSecret': [
    {
      required: true,
      message: $i18n.t('generic.validate.required'),
      trigger: 'blur',
    },
  ],
  accountName: [
    {
      required: true,
      message: $i18n.t('generic.validate.required'),
      trigger: 'blur',
    },
    {
      message: $i18n.t('projects.cloudToken.nameRegex'),
      trigger: 'blur',
      validator(val) {
        return /^[0-9a-zA-Z-]+$/g.test(val);
      },
    },
  ],
});

// 抽屉关闭校验
const { handleBeforeClose } = useSideslider(formData);

// 编辑描述字段
const editIndex = ref(-1);
const newDesc = ref('');
const hoverIndex = ref(-1);
const handleRowEnter = (index) => {
  hoverIndex.value = index;
};
const handleRowLeave = () => {
  hoverIndex.value = -1;
};
const editDesc = (index: number) => {
  editIndex.value = index;
};
const handleDescChange = (desc: string) => {
  newDesc.value = desc;
};
const handleUpdateDesc = async (row: ICloudAccount) => {
  loading.value = true;
  const result = await updateCloudAccounts({
    $cloudId: 'gcpCloud',
    $accountID: row.account.accountID,
    desc: newDesc.value,
  });
  loading.value = false;
  if (result) {
    editIndex.value = -1;
    handleGetAccountsList();
  }
};

// 云账号列表
const loading = ref(false);
const webAnnotations = ref({ perms: {} });
const tableData = ref<ICloudAccount[]>([]);
const keys = ref(['account.accountName', 'clusters']); // 模糊搜索字段
const { tableDataMatchSearch, searchValue } = useTableSearch(tableData, keys);
const { pageChange, pageSizeChange, curPageData, pagination } = usePage(tableDataMatchSearch);
const getAccountSecretData = (row: ICloudAccount, key: string) => {
  try {
    return JSON.parse(row.account.account.serviceAccountSecret || '')[key] || '--';
  } catch (_) {
    return '--';
  }
};
const handleGetAccountsList = async () => {
  loading.value  = true;
  const { data, web_annotations } = await cloudAccounts('gcpCloud');
  tableData.value = data;
  webAnnotations.value = web_annotations;
  loading.value = false;
};

// 删除Token
const handleDeleteToken = (row) => {
  $bkInfo({
    type: 'warning',
    clsName: 'custom-info-confirm',
    subTitle: row.account.accountName,
    title: $i18n.t('projects.cloudToken.delete'),
    defaultInfo: true,
    confirmFn: async () => {
      const result = await deleteCloudAccounts({
        $cloudId: 'gcpCloud',
        $accountID: row.account.accountID,
      });
      if (result) {
        $bkMessage({
          theme: 'success',
          message: $i18n.t('generic.msg.success.delete'),
        });
        handleGetAccountsList();
      }
    },
  });
};

// 导入文件
const codeEditorRef = ref<any>(null);
const editorKey = ref();
const fileRef = ref<any>(null);
const handleInputFile = (event) => {
  const [file] = event.target?.files || [];
  if (!file) return;

  const reader = new FileReader();
  reader.readAsText(file);
  reader.onload = () => {
    formData.value.account.serviceAccountSecret = String(reader.result);
    formData.value.accountName = JSON.parse(reader.result as any)?.client_email?.split('@')?.[0];
    fileRef.value && (fileRef.value.value = '');
    editorKey.value = Math.random();
  };
  reader.onerror = () => {
    fileRef.value && (fileRef.value.value = '');
  };
};
// 全屏
const editorRef = ref();
// const handleFullScreen = () => {
//   fullScreen(editorRef.value);
// };
// const handleExitFullScreen = () => {
//   exitFullscreen(editorRef.value);
// };

// 校验file
const validating = ref(false);
const isValidate = ref(false);
const validateErrMsg = ref('');
watch(() => formData.value.account.serviceAccountSecret, () => {
  isValidate.value = false;
});
const handleValidateFile = async () => {
  if (!formData.value.account.serviceAccountSecret) return;

  validating.value = true;
  validateErrMsg.value = await validateCloudAccounts({
    $cloudId: 'gcpCloud',
    account: {
      serviceAccountSecret: formData.value.account.serviceAccountSecret,
    },
  });
  validating.value = false;
  isValidate.value = !validateErrMsg.value;
};

const showCreate = ref(false);
watch(showCreate, () => {
  // 重置表单数据
  formData.value = {
    accountName: '',
    desc: '',
    account: {
      serviceAccountSecret: '',
    },
  };
  isValidate.value = false;
  validateErrMsg.value = '';
});
// 高亮当前创建行
const activeRow = ref<any>(null);
const getRowClassName = ({ row }) => (row.account.accountID === activeRow.value?.accountID ? 'high-row' : '');
// 创建Token
const creating = ref(false);
const handleCreateToken = async () => {
  const valid = await formRef.value?.validate().catch(() => false);
  if (!valid) return;
  creating.value = true;
  const data = await createCloudAccounts({
    ...formData.value,
    $cloudId: 'gcpCloud',
  });
  creating.value = false;
  if (data) {
    activeRow.value = data;
    $bkMessage({
      theme: 'success',
      message: $i18n.t('generic.msg.success.create'),
    });
    showCreate.value = false;
    handleGetAccountsList().then(() => {
      setTimeout(() => {
        activeRow.value = null;
      }, 2000);
    });
  }
};

onMounted(() => {
  handleGetAccountsList();
});

</script>
<style lang="postcss" scoped>
>>> .high-row {
  background-color: #F2FFF4;
}
.footer {
  border-top: 1px solid #dfe0e5;
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
