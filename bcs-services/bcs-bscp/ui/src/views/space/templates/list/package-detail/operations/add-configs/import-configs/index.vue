<template>
  <bk-dialog
    :is-show="props.show"
    :title="t('批量上传配置文件')"
    :theme="'primary'"
    width="960"
    height="720"
    ext-cls="import-file-dialog"
    :esc-close="false"
    :before-close="handleBeforeClose"
    @closed="emits('update:show', false)">
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
        @decompressing="decompressing = $event" />
    </div>
    <bk-loading
      :loading="decompressing"
      :title="t('压缩包正在解压，请稍后')"
      class="config-table-loading"
      mode="spin"
      theme="primary"
      size="small"
      :opacity="0.7">
      <div v-if="importConfigList.length" class="content">
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

    <template #footer>
      <bk-button
        theme="primary"
        style="margin-right: 8px"
        :disabled="!confirmBtnDisabled"
        @click="isSelectPkgDialogShow = true">
        {{ t('导入') }}
      </bk-button>
      <bk-button @click="emits('update:show', false)">{{ t('取消') }}</bk-button>
    </template>
  </bk-dialog>
  <SelectPackage v-model:show="isSelectPkgDialogShow" :pending="pending" @confirm="handleImport" />
</template>
<script lang="ts" setup>
  import { ref, watch, computed } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { storeToRefs } from 'pinia';
  import useGlobalStore from '../../../../../../../../store/global';
  import useTemplateStore from '../../../../../../../../store/template';
  import useModalCloseConfirmation from '../../../../../../../../utils/hooks/use-modal-close-confirmation';
  import { IConfigImportItem } from '../../../../../../../../../types/config';
  import { importTemplateBatchAdd, addTemplateToPackage } from '../../../../../../../../api/template';
  import ConfigTable from './config-table.vue';
  import SelectPackage from './select-package.vue';
  import Message from 'bkui-vue/lib/message';
  import ImportFromLocalFile from '../../../../../../service/detail/config/config-list/config-table-list/create-config/import-file/import-from-local-file.vue';

  const { t } = useI18n();
  const props = defineProps<{
    show: boolean;
  }>();

  const emits = defineEmits(['update:show', 'added']);
  const { spaceId } = storeToRefs(useGlobalStore());
  const { currentTemplateSpace, batchUploadIds } = storeToRefs(useTemplateStore());
  const isShow = ref(false);
  const isTableChange = ref(false);
  const pending = ref(false);
  const existConfigList = ref<IConfigImportItem[]>([]);
  const nonExistConfigList = ref<IConfigImportItem[]>([]);
  const expandNonExistTable = ref(true);
  const expandExistTable = ref(true);
  const isSelectPkgDialogShow = ref(false);
  const importType = ref('localFile');
  const uploadFileLoading = ref(false);
  const decompressing = ref(false);

  watch(
    () => props.show,
    (val) => {
      clearData();
      isShow.value = val;
      isTableChange.value = false;
    },
  );

  const importConfigList = computed(() => [...existConfigList.value, ...nonExistConfigList.value]);

  const confirmBtnDisabled = computed(() => {
    return !uploadFileLoading.value && !decompressing.value && importConfigList.value.length > 0;
  });

  const handleBeforeClose = async () => {
    if (isTableChange.value) {
      const result = await useModalCloseConfirmation();
      return result;
    }
    return true;
  };

  const close = () => {
    clearData();
    emits('update:show', false);
  };

  const handleImport = async (pkgIds: number[]) => {
    pending.value = true;
    try {
      const res = await importTemplateBatchAdd(spaceId.value, currentTemplateSpace.value, [
        ...existConfigList.value,
        ...nonExistConfigList.value,
      ]);
      // 选择未指定套餐时,不需要调用添加接口
      if (pkgIds.length > 1 || pkgIds[0] !== 0) {
        await addTemplateToPackage(spaceId.value, currentTemplateSpace.value, res.ids, pkgIds);
      }
      batchUploadIds.value = res.ids;
      isSelectPkgDialogShow.value = false;
      close();
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
    isTableChange.value = true;
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
  };
</script>

<style scoped lang="scss">
  .import-type-select {
    display: flex;
  }
  .label {
    width: 72px;
    height: 32px;
    line-height: 32px;
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
