<template>
  <bk-dialog
    :is-show="props.show"
    :title="t('批量导入')"
    :theme="'primary'"
    width="1200"
    height="720"
    ext-cls="import-kv-dialog"
    :esc-close="false"
    @closed="handleClose">
    <div class="import-type-select">
      <div class="label">{{ t('导入方式') }}</div>
      <bk-radio-group v-model="importType">
        <bk-radio-button label="text">{{ t('文本格式导入') }}</bk-radio-button>
        <bk-radio-button label="historyVersion">{{ t('从历史版本导入') }}</bk-radio-button>
        <bk-radio-button label="otherService">{{ t('从其他服务导入') }}</bk-radio-button>
      </bk-radio-group>
    </div>
    <div v-if="importType === 'text'">
      <TextImport :bk-biz-id="bkBizId" :app-id="appId" />
    </div>
    <div v-else-if="importType === 'historyVersion'">
      <div class="wrap">
        <div class="label">{{ $t('选择版本') }}</div>
        <bk-select
          v-model="selectVerisonId"
          :loading="versionListLoading"
          style="width: 374px"
          filterable
          auto-focus
          @select="handleSelectVersion(appId, $event)">
          <bk-option v-for="item in versionList" :id="item.id" :key="item.id" :name="item.spec.name" />
        </bk-select>
      </div>
    </div>
    <div v-else>
      <ImportFormOtherService :bk-biz-id="bkBizId" :app-id="appId" @select-version="handleSelectVersion" />
    </div>
    <template #footer>
      <bk-button theme="primary" style="margin-right: 8px" :disabled="!confirmBtnPerm" @click="handleConfirm">
        {{ t('导入') }}
      </bk-button>
      <bk-button @click="handleClose">{{ t('取消') }}</bk-button>
    </template>
  </bk-dialog>
</template>

<script lang="ts" setup>
  import { ref, watch, computed, onMounted } from 'vue';
  import { useI18n } from 'vue-i18n';
  import {
    batchImportKvFile,
    getConfigVersionList,
    importKvFromHistoryVersion,
  } from '../../../../../../../../api/config';
  import { IConfigVersion } from '../../../../../../../../../types/config';
  import { Message } from 'bkui-vue';
  import TextImport from './import-kv/text-import.vue';
  import ImportFormOtherService from './import/import-form-other-service.vue';

  const { t } = useI18n();
  const props = defineProps<{
    show: boolean;
    bkBizId: string;
    appId: number;
  }>();
  const emits = defineEmits(['update:show', 'confirm']);

  const editorRef = ref();
  const isFormChange = ref(false);
  const importType = ref('text');
  const textConfirmBtnPerm = ref(false);
  const selectedFile = ref<File>();
  const isFileUploadSuccess = ref(true);
  const loading = ref(false);
  const selectVerisonId = ref();
  const versionListLoading = ref(false);
  const versionList = ref<IConfigVersion[]>([]);
  const tableLoading = ref(false);

  watch(
    () => props.show,
    () => {
      isFormChange.value = false;
    },
  );

  onMounted(() => {
    getVersionList();
  });

  const confirmBtnPerm = computed(() => {
    if (importType.value === 'text') return textConfirmBtnPerm.value;
    return !!selectedFile.value && isFileUploadSuccess.value;
  });

  const getVersionList = async () => {
    try {
      versionListLoading.value = true;
      const params = {
        start: 0,
        all: true,
      };
      const res = await getConfigVersionList(props.bkBizId, props.appId, params);
      versionList.value = res.data.details;
    } catch (e) {
      console.error(e);
    } finally {
      versionListLoading.value = false;
    }
  };

  const handleSelectVersion = async (other_app_id: number, release_id: number) => {
    tableLoading.value = true;
    try {
      const params = { other_app_id, release_id };
      const res = await importKvFromHistoryVersion(props.bkBizId, props.appId, params);
      console.log(res);
      // existConfigList.value = res.data.exist;
      // nonExistConfigList.value = res.data.non_exist;
      // initExistConfigList.value = cloneDeep(res.data.exist);
      // initNonExistConfigList.value = cloneDeep(res.data.non_exist);
    } catch (e) {
      console.error(e);
    } finally {
      tableLoading.value = false;
    }
  };

  const handleClose = () => {
    selectedFile.value = undefined;
    isFileUploadSuccess.value = true;
    emits('update:show', false);
  };

  const handleConfirm = async () => {
    loading.value = true;
    if (importType.value === 'file') {
      try {
        await batchImportKvFile(props.bkBizId, props.appId, selectedFile.value);
        Message({
          theme: 'success',
          message: t('文件导入成功'),
        });
      } catch (error) {
        console.error(error);
        isFileUploadSuccess.value = false;
        return;
      } finally {
        loading.value = false;
      }
    } else {
      try {
        await editorRef.value.handleImport();
        Message({
          theme: 'success',
          message: t('文本导入成功'),
        });
      } catch (error) {
        console.error(error);
      } finally {
        loading.value = false;
      }
    }
    emits('update:show', false);
    emits('confirm');
  };
</script>

<style scoped lang="scss">
  .import-type-select {
    display: flex;
  }
  .label {
    width: 70px;
    height: 32px;
    line-height: 32px;
    font-size: 12px;
    color: #63656e;
  }
  .wrap {
    display: flex;
    margin-top: 24px;
  }
  :deep(.wrap) {
    display: flex;
    flex-wrap: wrap;
    margin-top: 24px;
    .label {
      @extend .label;
    }
  }
  :deep(.other-service-wrap) {
    display: flex;
    margin-top: 24px;
    gap: 24px;
    .label {
      @extend .label;
    }
  }
</style>

<style lang="scss">
  .import-kv-dialog {
    .bk-modal-content {
      height: 100% !important;
      overflow: hidden !important;
    }
  }
</style>
