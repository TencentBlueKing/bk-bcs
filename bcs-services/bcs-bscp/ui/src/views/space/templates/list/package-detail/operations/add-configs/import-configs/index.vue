<template>
  <bk-dialog
    :is-show="props.show"
    :title="t('批量上传配置文件')"
    :theme="'primary'"
    width="960"
    height="720"
    ext-cls="import-file-dialog"
    :before-close="handleBeforeClose"
    :quick-close="false"
    @closed="handleClose">
    <div v-if="currentStep === 'upload'">
      <div :class="['select-wrap', { 'en-select-wrap': locale === 'en' }]">
        <div class="import-type-select">
          <div class="label">{{ t('导入方式') }}</div>
          <bk-radio-group v-model="importType">
            <bk-radio-button label="localFile">{{ t('导入本地文件') }}</bk-radio-button>
            <bk-radio-button label="otherSpace" :disabled="true">{{ t('从其他空间导入') }}</bk-radio-button>
          </bk-radio-group>
        </div>
        <div v-if="importType === 'localFile'">
          <ImportFromLocalFile
            :space-id="spaceId"
            :current-template-space="currentTemplateSpace"
            :is-template="true"
            @change="handleUploadFile"
            @delete="handleDeleteFile"
            @uploading="uploadFileLoading = $event"
            @decompressing="decompressing = $event"
            @file-processing="fileProcessing = $event" />
        </div>
      </div>
      <bk-loading
        :loading="decompressing || fileProcessing"
        :title="loadingText"
        class="config-table-loading"
        mode="spin"
        theme="primary"
        size="small"
        :opacity="0.7">
        <div v-if="importConfigList.length" class="content">
          <bk-alert
            v-if="isExceedMaxFileCount"
            style="margin-top: 4px"
            theme="error"
            :title="
              $t('配置文件数量超过最大上传限制 ({n} 个文件)', { n: spaceFeatureFlags.RESOURCE_LIMIT.TmplSetTmplCnt })
            " />
          <div class="head">
            <div class="tips">
              {{ t('共将导入') }} <span style="color: #3a84ff">{{ importConfigList.length }}</span>
              {{ t('个配置项，其中') }} <span style="color: #ffa519">{{ existConfigList.length }}</span>
              {{ t('个已存在,导入后将') }}
              <span style="color: #ffa519">{{ t('覆盖原配置') }}</span>
            </div>
          </div>
          <ConfigTable
            v-if="nonExistConfigList.length"
            :table-data="nonExistConfigList"
            :is-exsit-table="false"
            :expand="expandNonExistTable"
            @change-expand="expandNonExistTable = !expandNonExistTable"
            @change="handleTableChange($event, true)" />
          <ConfigTable
            v-if="existConfigList.length"
            :expand="expandExistTable"
            :table-data="existConfigList"
            :is-exsit-table="true"
            @change-expand="expandExistTable = !expandExistTable"
            @change="handleTableChange($event, false)" />
        </div>
      </bk-loading>
    </div>
    <SelectPackage
      v-else
      ref="selectedPkgsRef"
      :config-id-list="importConfigIdList"
      @toggle-btn-disabled="importBtnDisabled = $event" />
    <template #footer>
      <div v-if="currentStep === 'upload'">
        <bk-button
          theme="primary"
          style="margin-right: 8px"
          :disabled="nextBtnDisabled"
          @click="currentStep = 'import'">
          {{ t('下一步') }}
        </bk-button>
        <bk-button @click="emits('update:show', false)">{{ t('取消') }}</bk-button>
      </div>
      <div v-else>
        <bk-button style="margin-right: 8px" @click="currentStep = 'upload'">
          {{ t('上一步') }}
        </bk-button>
        <bk-button
          theme="primary"
          style="margin-right: 8px"
          :loading="pending"
          :disabled="importBtnDisabled"
          @click="handleImport">
          {{ t('导入') }}
        </bk-button>
        <bk-button @click="emits('update:show', false)">{{ t('取消') }}</bk-button>
      </div>
    </template>
  </bk-dialog>
</template>
<script lang="ts" setup>
  import { ref, watch, computed } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { storeToRefs } from 'pinia';
  import useGlobalStore from '../../../../../../../../store/global';
  import useTemplateStore from '../../../../../../../../store/template';
  import useModalCloseConfirmation from '../../../../../../../../utils/hooks/use-modal-close-confirmation';
  import { IConfigImportItem } from '../../../../../../../../../types/config';
  import { importTemplateBatchAdd } from '../../../../../../../../api/template';
  import ConfigTable from './config-table.vue';
  import SelectPackage from './select-package.vue';
  import Message from 'bkui-vue/lib/message';
  import ImportFromLocalFile from '../../../../../../service/detail/config/config-list/config-table-list/create-config/import-file/import-from-local-file.vue';

  const { t, locale } = useI18n();
  const props = defineProps<{
    show: boolean;
  }>();

  const templateStore = useTemplateStore();

  const emits = defineEmits(['update:show', 'added']);
  const { spaceId, spaceFeatureFlags } = storeToRefs(useGlobalStore());
  const { currentTemplateSpace } = storeToRefs(useTemplateStore());
  const isShow = ref(false);
  const isFormChange = ref(false);
  const pending = ref(false);
  const existConfigList = ref<IConfigImportItem[]>([]);
  const nonExistConfigList = ref<IConfigImportItem[]>([]);
  const expandNonExistTable = ref(true);
  const expandExistTable = ref(true);
  const isSelectPkgDialogShow = ref(false);
  const importType = ref('localFile');
  const uploadFileLoading = ref(false);
  const decompressing = ref(false); // 后台压缩包解压
  const fileProcessing = ref(false); // 后台文件处理
  const currentStep = ref('upload');
  const selectedPkgsRef = ref();
  const importBtnDisabled = ref(true);

  watch(
    () => props.show,
    (val) => {
      clearData();
      isShow.value = val;
      isFormChange.value = false;
      currentStep.value = 'upload';
    },
  );


  const importConfigList = computed(() => [...existConfigList.value, ...nonExistConfigList.value]);

  const nextBtnDisabled = computed(() => {
    return (
      uploadFileLoading.value ||
      decompressing.value ||
      importConfigList.value.length === 0 ||
      isExceedMaxFileCount.value
    );
  });

  const importConfigIdList = computed(() => {
    return importConfigList.value.map((item) => {
      return {
        name: item.name,
        id: item.id,
      };
    });
  });

  const loadingText = computed(() => {
    if (decompressing.value) {
      return t('压缩包正在解压，请稍后...');
    }
    if (fileProcessing.value) {
      return t('后台正在处理上传数据，请稍后...');
    }
    return '';
  });

  const isExceedMaxFileCount = computed(
    () => importConfigList.value.length > spaceFeatureFlags.value.RESOURCE_LIMIT.TmplSetTmplCnt,
  );

  const handleBeforeClose = async () => {
    if (isFormChange.value) {
      const result = await useModalCloseConfirmation();
      return result;
    }
    return true;
  };

  const handleClose = () => {
    clearData();
    emits('update:show', false);
  };

  const handleImport = async () => {
    const pkgIds = selectedPkgsRef.value.selectedPkgs;
    pending.value = true;
    try {
      const res = await importTemplateBatchAdd(
        spaceId.value,
        currentTemplateSpace.value,
        [...existConfigList.value, ...nonExistConfigList.value],
        pkgIds[0] === 0 ? [] : pkgIds,
      );
      templateStore.$patch((state) => {
        state.topIds = res.ids;
      });
      isSelectPkgDialogShow.value = false;
      handleClose();
      setTimeout(() => {
        emits('added');
        Message({
          theme: 'success',
          message: t('导入配置文件成功'),
        });
      }, 300);
    } catch (e) {
      console.log(e);
    } finally {
      pending.value = false;
    }
  };

  const handleTableChange = (data: IConfigImportItem[], isNonExistData: boolean) => {
    if (isNonExistData) {
      nonExistConfigList.value = data;
    } else {
      existConfigList.value = data;
    }
    isFormChange.value = true;
  };

  const clearData = () => {
    nonExistConfigList.value = [];
    existConfigList.value = [];
  };

  // 删除文件处理表格数据
  const handleDeleteFile = (fileName: string) => {
    existConfigList.value = existConfigList.value.filter((item) => item.file_name !== fileName);
    nonExistConfigList.value = nonExistConfigList.value.filter((item) => item.file_name !== fileName);
  };

  // 上传文件获取表格数据
  const handleUploadFile = (exist: IConfigImportItem[], nonExist: IConfigImportItem[]) => {
    existConfigList.value = [...existConfigList.value, ...exist];
    nonExistConfigList.value = [...nonExistConfigList.value, ...nonExist];
    isFormChange.value = true;
  };
</script>

<style scoped lang="scss">
  .select-wrap {
    .import-type-select {
      display: flex;
    }
    .label {
      padding-top: 8px;
      width: 72px;
      font-size: 12px;
      color: #63656e;
      margin-right: 22px;
      text-align: right;
    }
    :deep(.wrap) {
      display: flex;
      margin-top: 24px;
      .label {
        @extend .label;
      }
    }
    &.en-select-wrap {
      .label {
        width: 100px !important;
      }
      :deep(.wrap) {
        .label {
          @extend .label;
        }
      }
      :deep(.upload-file-list) {
        margin-left: 120px;
      }
    }
  }

  .content {
    margin-top: 24px;
    border-top: 1px solid #dcdee5;
    .head {
      display: flex;
      align-items: center;
      margin: 16px 0;
      font-size: 12px;
      color: #63656e;
      .bk-checkbox {
        margin-left: 0 !important;
        font-size: 12px;
      }
    }
  }
  .config-table-loading {
    min-height: 80px;
    :deep(.bk-loading-primary) {
      top: 60px;
      align-items: center;
    }
  }
</style>

<style lang="scss">
  .import-file-dialog {
    .bk-modal-content {
      height: calc(100% - 50px) !important;
      overflow: auto;
    }
  }
</style>
