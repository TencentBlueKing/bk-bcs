<template>
  <bk-dialog
    :is-show="props.show"
    :title="t('批量导入')"
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
        <bk-radio-button label="configTemplate">{{ t('从配置模板导入') }}</bk-radio-button>
        <bk-radio-button label="historyVersion">{{ t('从历史版本导入') }}</bk-radio-button>
        <bk-radio-button label="otherService">{{ t('从其他服务导入') }}</bk-radio-button>
      </bk-radio-group>
    </div>
    <div v-if="importType === 'localFile'">
      <ImportFromLocalFile
        :bk-biz-id="props.bkBizId"
        :app-id="props.appId"
        :is-template="false"
        @change="handleUploadFile"
        @delete="handleDeleteFile"
        @uploading="uploadFileLoading = $event"
        @decompressing="decompressing = $event" />
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
          :clearable="false"
          @select="handleSelectVersion(appId, $event)">
          <bk-option v-for="item in versionList" :id="item.id" :key="item.id" :name="item.spec.name" />
        </bk-select>
      </div>
    </div>
    <div v-else-if="importType === 'otherService'">
      <ImportFormOtherService
        :bk-biz-id="props.bkBizId"
        :app-id="props.appId"
        @select-version="handleSelectVersion"
        @clear="handleClearTable" />
    </div>
    <bk-loading
      :loading="decompressing || tableLoading"
      :title="decompressing ? t('压缩包正在解压，请稍后') : ''"
      class="config-table-loading"
      mode="spin"
      theme="primary"
      size="small"
      :opacity="0.7">
      <div
        v-if="importType !== 'configTemplate' && importConfigList.length + importTemplateConfigList.length > 0"
        class="content">
        <div class="head">
          <bk-checkbox style="margin-left: 24px" v-model="isClearDraft"> {{ $t('导入前清空草稿区') }} </bk-checkbox>
          <div v-if="!isClearDraft" class="tips">
            {{ t('共将导入') }} <span style="color: #3a84ff">{{ importConfigList.length }}</span> {{ t('个配置项') }},
            <span style="color: #3a84ff">{{ importTemplateConfigList.length }}</span> {{ t('个模板套餐') }},
            {{ t('其中') }}
            <span style="color: #ffa519">{{ existConfigList.length }}</span>
            {{ t('个配置项') }}, <span style="color: #ffa519">{{ existTemplateConfigList.length }}</span>
            {{ t('个模板套餐已存在，导入后将') }}
            <span style="color: #ffa519">{{ t('覆盖原配置') }}</span>
          </div>
          <div v-else class="tips">
            {{ t('将') }} <span style="color: #ffa519">{{ t('清空') }}</span> {{ t('现有草稿区,并导入') }}
            <span style="color: #3a84ff">{{ importConfigList.length }}</span>
            {{ t('个配置项') }}, <span style="color: #3a84ff">{{ importTemplateConfigList.length }}</span>
            {{ t('个模板套餐') }}
          </div>
        </div>
        <ConfigTable
          v-if="nonExistConfigList.length"
          :table-data="nonExistConfigList"
          :is-exsit-table="false"
          @change="handleConfigTableChange($event, true)" />
        <TemplateConfigTable
          v-if="nonExistTemplateConfigList.length"
          :table-data="nonExistTemplateConfigList"
          :is-exsit-table="false"
          @change="handleTemplateTableChange($event, true)" />
        <ConfigTable
          v-if="existConfigList.length"
          :table-data="existConfigList"
          :is-exsit-table="true"
          @change="handleConfigTableChange($event, false)" />
        <TemplateConfigTable
          v-if="existTemplateConfigList.length"
          :table-data="existTemplateConfigList"
          :is-exsit-table="true"
          @change="handleTemplateTableChange($event, true)" />
      </div>
    </bk-loading>

    <template #footer>
      <bk-button
        theme="primary"
        style="margin-right: 8px"
        :disabled="!confirmBtnDisabled || loading"
        :loading="loading"
        @click="handleConfirm">
        {{ t('导入') }}
      </bk-button>
      <bk-button :loading="closeLoading" :disabled="closeLoading" @click="handleClose">{{ t('取消') }}</bk-button>
    </template>
  </bk-dialog>
</template>

<script lang="ts" setup>
  import { ref, watch, computed, onMounted } from 'vue';
  import { useI18n } from 'vue-i18n';
  import {
    getConfigVersionList,
    importFromHistoryVersion,
    batchAddConfigList,
  } from '../../../../../../../../../api/config';
  import { IConfigVersion, IConfigImportItem } from '../../../../../../../../../../types/config';
  import { Message } from 'bkui-vue';
  import createSamplePkg from '../../../../../../../../../utils/sample-file-pkg';
  import ImportFromTemplate from './import-from-templates.vue';
  import ImportFromLocalFile from './import-from-local-file.vue';
  import ImportFormOtherService from './import-form-other-service.vue';
  import ConfigTable from '../../../../../../../templates/list/package-detail/operations/add-configs/import-configs/config-table.vue';
  import useModalCloseConfirmation from '../../../../../../../../../utils/hooks/use-modal-close-confirmation';
  import useServiceStore from '../../../../../../../../../store/service';
  import { ImportTemplateConfigItem } from '../../../../../../../../../../types/template';
  import TemplateConfigTable from './template-config-table.vue';

  const { t } = useI18n();

  const serviceStore = useServiceStore();

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
  const tableLoading = ref(false);
  const existConfigList = ref<IConfigImportItem[]>([]);
  const nonExistConfigList = ref<IConfigImportItem[]>([]);
  const existTemplateConfigList = ref<ImportTemplateConfigItem[]>([]);
  const nonExistTemplateConfigList = ref<ImportTemplateConfigItem[]>([]);
  const isClearDraft = ref(false);
  const uploadFileLoading = ref(false);
  const decompressing = ref(false);
  const closeLoading = ref(false);

  const confirmBtnDisabled = computed(() => {
    if (importType.value === 'configTemplate' && importFromTemplateRef.value) {
      return importFromTemplateRef.value.isImportBtnDisabled;
    }
    if (importType.value === 'localFile') {
      return (
        !uploadFileLoading.value &&
        !decompressing.value &&
        importConfigList.value.length + importTemplateConfigList.value.length > 0
      );
    }
    return importConfigList.value.length + importTemplateConfigList.value.length > 0;
  });

  const importConfigList = computed(() => [...nonExistConfigList.value, ...existConfigList.value]);
  const importTemplateConfigList = computed(() => [
    ...nonExistTemplateConfigList.value,
    ...existTemplateConfigList.value,
  ]);

  watch(
    () => props.show,
    (val) => {
      if (val) {
        importType.value = 'localFile';
        isTableChange.value = false;
        handleClearTable();
        selectVerisonId.value = undefined;
        getVersionList();
        decompressing.value = false;
        closeLoading.value = false;
      }
    },
  );

  watch(
    () => importType.value,
    () => {
      handleClearTable();
      selectVerisonId.value = undefined;
    },
  );

  onMounted(async () => {
    downloadHref.value = (await createSamplePkg()) as string;
  });

  const handleConfirm = async () => {
    loading.value = true;
    try {
      if (importType.value === 'configTemplate') {
        await importFromTemplateRef.value.handleImportConfirm();
      } else {
        let allVariables: {
          default_val: string;
          memo: string;
          name: string;
          type: string;
        }[] = [];
        const allConfigList: any[] = [];
        const allTemplateConfigList: any[] = [];
        importTemplateConfigList.value.forEach((templateConfig) => {
          const { template_set_id, template_revisions, template_space_id } = templateConfig;
          template_revisions.forEach((revision) => {
            allVariables = [...allVariables, ...revision.variables];
          });
          allTemplateConfigList.push({
            template_space_id,
            template_binding: {
              template_set_id,
              template_revisions: template_revisions.map((revision) => {
                const { template_id, template_revision_id, is_latest } = revision;
                return {
                  template_id,
                  template_revision_id,
                  is_latest,
                };
              }),
            },
          });
        });
        importConfigList.value.forEach((config) => {
          const { variables, ...rest } = config;
          if (variables) {
            allVariables = [...allVariables, ...config.variables];
          }
          allConfigList.push({
            ...rest,
          });
        });
        const query = {
          bindings: allTemplateConfigList,
          items: allConfigList,
          replace_all: isClearDraft.value,
          variables: allVariables,
        };
        const res = await batchAddConfigList(props.bkBizId, props.appId, query);
        serviceStore.$patch((state) => {
          state.topIds = res.ids;
        });
      }
      emits('update:show', false);
      setTimeout(() => {
        emits('confirm');
        Message({
          theme: 'success',
          message: t('配置文件导入成功'),
        });
      }, 300);
    } catch (error) {
      console.error(error);
    }
    loading.value = false;
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

  const handleSelectVersion = async (other_app_id: number, release_id: number) => {
    tableLoading.value = true;
    try {
      handleClearTable();
      const params = { other_app_id, release_id };
      const res = await importFromHistoryVersion(props.bkBizId, props.appId, params);
      res.data.non_template_configs.forEach((item: any) => {
        const config = {
          ...item,
          ...item.config_item_spec,
          ...item.config_item_spec.permission,
          sign: item.signature,
        };
        delete config.config_item_spec;
        delete config.permission;
        delete config.signature;
        if (item.is_exist) {
          existConfigList.value.push(config);
        } else {
          nonExistConfigList.value.push(config);
        }
      });
      res.data.template_configs.forEach((item: any) => {
        if (item.is_exist) {
          existTemplateConfigList.value.push(item);
        } else {
          nonExistTemplateConfigList.value.push(item);
        }
      });
    } catch (e) {
      console.error(e);
    } finally {
      tableLoading.value = false;
    }
  };

  const handleConfigTableChange = (data: IConfigImportItem[], isNonExistData: boolean) => {
    if (isNonExistData) {
      nonExistConfigList.value = data;
    } else {
      existConfigList.value = data;
    }
    isTableChange.value = true;
  };

  const handleTemplateTableChange = (deleteIndex: number, isNonExistData: boolean) => {
    if (isNonExistData) {
      nonExistTemplateConfigList.value.splice(deleteIndex, 1);
    } else {
      existTemplateConfigList.value.splice(deleteIndex, 1);
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

  // 上传文件获取表格数据
  const handleUploadFile = (exist: IConfigImportItem[], nonExist: IConfigImportItem[]) => {
    existConfigList.value = [...existConfigList.value, ...exist];
    nonExistConfigList.value = [...nonExistConfigList.value, ...nonExist];
  };

  // 删除文件处理表格数据
  const handleDeleteFile = (fileName: string) => {
    existConfigList.value = existConfigList.value.filter((item) => item.file_name !== fileName);
    nonExistConfigList.value = nonExistConfigList.value.filter((item) => item.file_name !== fileName);
  };

  const handleClearTable = () => {
    existConfigList.value = [];
    nonExistConfigList.value = [];
    existTemplateConfigList.value = [];
    nonExistTemplateConfigList.value = [];
  };

  const handleClose = () => {
    closeLoading.value = true;
    emits('update:show', false);
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

  :deep(.other-service-wrap) {
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
      .tips {
        padding-left: 16px;
        border-left: 1px solid #dcdee5;
        margin-left: 16px;
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
