<template>
  <bk-dialog
    :is-show="props.show"
    :title="t('批量导入')"
    :theme="'primary'"
    width="960"
    height="720"
    ext-cls="variable-import-dialog"
    :esc-close="false"
    :before-close="handleBeforeClose"
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
      <div class="wrap">
        <div class="label">{{ $t('选择版本') }}</div>
        <bk-select
          v-model="selectVerisonId"
          :loading="versionListLoading"
          style="width: 374px"
          filterable
          auto-focus
          @select="handleSelectVersion">
          <bk-option v-for="item in versionList" :id="item.id" :key="item.id" :name="item.spec.name" />
        </bk-select>
      </div>
    </div>
    <div v-else-if="importType === 'otherService'">
      <ImportFormOtherService :bk-biz-id="props.bkBizId" :app-id="props.appId" />
    </div>
    <div v-if="importType !== 'configTemplate' && importConfigList.length" class="content">
      <div class="head">
        <bk-checkbox style="margin-left: 24px" v-model="isClearDraft"> {{ $t('导入前清空草稿区') }} </bk-checkbox>
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
        :expand="!expandNonExistTable"
        :table-data="existConfigList"
        :is-exsit-table="true"
        v-if="existConfigList.length"
        @change-expand="expandNonExistTable = !expandNonExistTable"
        @change="handleTableChange($event, false)" />
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
  import { getConfigVersionList, importFromHistoryVersion } from '../../../../../../../../api/config';
  import { IConfigVersion, IConfigImportItem } from '../../../../../../../../../types/config';
  import createSamplePkg from '../../../../../../../../utils/sample-file-pkg';
  import ImportFromTemplate from './import/import-from-templates.vue';
  import ImportFromLocalFile from './import/import-from-local-file.vue';
  import ImportFormOtherService from './import/import-form-other-service.vue';
  import ConfigTable from '../../../../../../templates/list/package-detail/operations/add-configs/import-configs/config-table.vue';
  import useModalCloseConfirmation from '../../../../../../../../utils/hooks/use-modal-close-confirmation';

  const { t } = useI18n();
  const props = defineProps<{
    show: boolean;
    bkBizId: string;
    appId: number;
  }>();
  const emits = defineEmits(['update:show', 'confirm']);

  const isTableChange = ref(false);
  const importType = ref('localFile');
  const loading = ref(false);
  const downloadHref = ref('');
  const importFromTemplateRef = ref();
  const versionListLoading = ref(false);
  const selectVerisonId = ref();
  const versionList = ref<IConfigVersion[]>([]);
  const existConfigList = ref<IConfigImportItem[]>([]);
  const nonExistConfigList = ref<IConfigImportItem[]>([]);
  const isClearDraft = ref(false);
  const expandNonExistTable = ref(true);

  const btnDisabled = computed(() => {
    if (importType.value === 'configTemplate' && importFromTemplateRef.value) {
      return importFromTemplateRef.value.isImportBtnDisabled;
    }
    return false;
  });

  const importConfigList = computed(() => [...existConfigList.value, ...nonExistConfigList.value]);

  watch(
    () => props.show,
    () => {
      importType.value = 'localFile';
      isTableChange.value = false;
      nonExistConfigList.value = [];
      existConfigList.value = [];
      getVersionList();
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

  const handleSelectVersion = async (id: number) => {
    try {
      const params = { other_app_id: props.appId, release_id: id };
      const res = await importFromHistoryVersion(props.bkBizId, props.appId, params);
      existConfigList.value = res.data.exist;
      nonExistConfigList.value = res.data.non_exist;
      console.log(res);
    } catch (e) {
      console.error(e);
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

  const handleBeforeClose = async () => {
    if (isTableChange.value) {
      const result = await useModalCloseConfirmation();
      return result;
    }
    return true;
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

  .content {
    margin-top: 24px;
    border-top: 1px solid #dcdee5;
    overflow: auto;
    max-height: 490px;
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
      .tips {
        padding-left: 16px;
        border-left: 1px solid #dcdee5;
        margin-left: 16px;
      }
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
