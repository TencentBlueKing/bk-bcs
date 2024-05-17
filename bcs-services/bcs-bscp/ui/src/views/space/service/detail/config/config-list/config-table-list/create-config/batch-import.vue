<template>
  <bk-dialog
    :is-show="props.show"
    :title="t('批量导入')"
    :theme="'primary'"
    width="960"
    height="720"
    ext-cls="variable-import-dialog"
    :esc-close="false"
    @closed="emits('update:show', false)">
    <div class="import-type-select">
      <div class="label">{{ t('导入方式') }}</div>
      <bk-radio-group v-model="importType">
        <bk-radio-button label="localFile">{{ t('导入本地文件') }}</bk-radio-button>
        <bk-radio-button label="configTemplate">{{ t('从配置模板导入') }}</bk-radio-button>
        <bk-radio-button label="historyVersion">{{ t('从历史版本导入') }}</bk-radio-button>
        <bk-radio-button label="otherService">{{ t('从其他服务导入') }}</bk-radio-button>
      </bk-radio-group>
    </div>
    <div v-if="importType === 'localFile'">
      <ImportFromLocalFile :bk-biz-id="props.bkBizId" :app-id="props.appId" />
    </div>
    <div v-else-if="importType === 'configTemplate'">
      <ImportFromTemplate ref="importFromTemplateRef" :bk-biz-id="props.bkBizId" :app-id="props.appId" />
    </div>
    <div v-else-if="importType === 'historyVersion'">
      <bk-select />
    </div>
    <div v-else-if="importType === 'otherService'">
      <div :label-width="100" :label="t('选择服务')">
        <bk-select />
      </div>
      <div :label-width="100" :label="t('选择版本')">
        <bk-select />
      </div>
    </div>
    <template #footer>
      <bk-button
        theme="primary"
        style="margin-right: 8px"
        :disabled="!btnDisabled"
        :loading="loading"
        @click="handleConfirm">
        {{ t('导入') }}
      </bk-button>
      <bk-button @click="emits('update:show', false)">{{ t('取消') }}</bk-button>
    </template>
  </bk-dialog>
</template>

<script lang="ts" setup>
  import { ref, watch, computed, onMounted } from 'vue';
  import { useI18n } from 'vue-i18n';
  import createSamplePkg from '../../../../../../../../utils/sample-file-pkg';
  import ImportFromTemplate from './import/import-from-templates.vue';
  import ImportFromLocalFile from './import/import-from-local-file.vue';

  const { t } = useI18n();
  const props = defineProps<{
    show: boolean;
    bkBizId: string;
    appId: number;
  }>();
  const emits = defineEmits(['update:show', 'confirm']);

  const isFormChange = ref(false);
  const importType = ref('localFile');
  const loading = ref(false);
  const downloadHref = ref('');
  const importFromTemplateRef = ref();

  const btnDisabled = computed(() => {
    if (importType.value === 'configTemplate' && importFromTemplateRef.value) {
      return importFromTemplateRef.value.isImportBtnDisabled;
    }
    return false;
  });

  watch(
    () => props.show,
    () => {
      importType.value = 'localFile';
      isFormChange.value = false;
    },
  );

  onMounted(async () => {
    downloadHref.value = (await createSamplePkg()) as string;
  });

  const handleConfirm = async () => {
    loading.value = true;
    if (importType.value === 'configTemplate') {
      await importFromTemplateRef.value.handleImportConfirm();
    }
    loading.value = false;
    emits('update:show', false);
    emits('confirm');
  };
</script>

<style scoped lang="scss">
  .import-type-select {
    display: flex;
  }
  .label {
    width: 94px;
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

  .bk-upload {
    display: flex;
    align-items: center;
    :deep(.bk-upload__tip) {
      color: #979ba5;
      margin-left: 12px;
    }
  }
  .upload-button {
    .text {
      margin-left: 5px;
      color: #63656e;
    }
  }
  :deep(.bk-form) {
    .div .bk-form-label {
      font-size: 12px;
    }
  }
  .import-other-service {
    display: flex;
    .bk-select {
      width: 362px;
    }
  }
</style>

<style lang="scss">
  .variable-import-dialog {
    .bk-modal-content {
      overflow: hidden !important;
    }
  }
</style>
