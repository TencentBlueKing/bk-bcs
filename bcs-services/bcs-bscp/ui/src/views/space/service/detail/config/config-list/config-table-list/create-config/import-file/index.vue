<template>
  <bk-dialog
    :is-show="props.show"
    :title="t('批量导入')"
    :theme="'primary'"
    width="960"
    height="720"
    ext-cls="import-file-dialog"
    :before-close="handleBeforeClose"
    :quick-close="false"
    @closed="handleClose">
    <div :class="['select-wrap', { 'en-select-wrap': locale === 'en' }]">
      <div class="import-type-select">
        <div :class="['label', { 'en-label': locale === 'en' }]">{{ t('导入方式') }}</div>
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
          @decompressing="decompressing = $event"
          @file-processing="fileProcessing = $event" />
      </div>
      <div v-else-if="importType === 'configTemplate'">
        <ImportFromTemplate
          ref="importFromTemplateRef"
          :bk-biz-id="props.bkBizId"
          :app-id="props.appId"
          @toggle-disabled="templateImportBtnDisabled = $event"
          @close="handleClose" />
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
    </div>
    <bk-loading
      v-if="importType !== 'configTemplate' && allConfigList.length + allTemplateConfigList.length > 0"
      :loading="decompressing || fileProcessing || tableLoading"
      :title="loadingText"
      class="config-table-loading"
      mode="spin"
      theme="primary"
      size="small"
      :opacity="0.7">
      <div class="content">
        <bk-alert
          v-if="isExceedMaxFileCount"
          style="margin-top: 4px"
          theme="error"
          :title="
            $t('配置文件数量超过最大上传限制 ({n} 个文件)', { n: spaceFeatureFlags.RESOURCE_LIMIT.AppConfigCnt })
          " />
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
          <bk-select
            v-if="importType === 'historyVersion' || importType === 'otherService'"
            ref="configSelectRef"
            class="config-select"
            v-model="selectedConfigIds"
            selected-style="checkbox"
            :popover-options="{ theme: 'light bk-select-popover config-selector-popover', placement: 'bottom-end' }"
            collapse-tags
            filterable
            multiple
            show-select-all
            @toggle="handleToggleConfigSelectShow"
            @blur="handleCloseConfigSelect">
            <template #trigger>
              <div class="select-btn">{{ $t('选择配置文件') }}</div>
            </template>
            <bk-option-group :label="$t('配置文件')" collapsible>
              <bk-option v-for="(item, index) in allConfigList" :id="item.id" :key="index" :label="fileAP(item)">
              </bk-option>
            </bk-option-group>
            <bk-option-group :label="$t('模板套餐')" collapsible>
              <bk-option
                v-for="(item, index) in allTemplateConfigList"
                :id="`${item.template_space_id} - ${item.template_set_id}`"
                :key="index"
                :label="`${item.template_space_name} - ${item.template_set_name}`">
              </bk-option>
            </bk-option-group>
            <template #extension>
              <div class="config-select-btns">
                <bk-button theme="primary" @click="handleConfirmSelect">{{ $t('确定') }}</bk-button>
                <bk-button @click="handleCloseConfigSelect">{{ $t('取消') }}</bk-button>
              </div>
            </template>
          </bk-select>
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
          @change="handleTemplateTableChange($event, false)" />
      </div>
    </bk-loading>
    <template #footer>
      <bk-button
        theme="primary"
        style="margin-right: 8px"
        :disabled="confirmBtnDisabled || loading"
        :loading="loading"
        @click="handleConfirm">
        {{ t('导入') }}
      </bk-button>
      <bk-button :loading="closeLoading" :disabled="closeLoading" @click="handleClose">{{ t('取消') }}</bk-button>
    </template>
  </bk-dialog>
</template>

<script lang="ts" setup>
  import { ref, watch, computed } from 'vue';
  import { useI18n } from 'vue-i18n';
  import {
    getConfigVersionList,
    importFromHistoryVersion,
    batchAddConfigList,
  } from '../../../../../../../../../api/config';
  import { IConfigVersion, IConfigImportItem } from '../../../../../../../../../../types/config';
  import { Message } from 'bkui-vue';
  import ImportFromTemplate from './import-from-templates.vue';
  import ImportFromLocalFile from './import-from-local-file.vue';
  import ImportFormOtherService from './import-form-other-service.vue';
  import ConfigTable from '../../../../../../../templates/list/package-detail/operations/add-configs/import-configs/config-table.vue';
  import useModalCloseConfirmation from '../../../../../../../../../utils/hooks/use-modal-close-confirmation';
  import useServiceStore from '../../../../../../../../../store/service';
  import useGlobalStore from '../../../../../../../../../store/global';
  import { ImportTemplateConfigItem } from '../../../../../../../../../../types/template';
  import TemplateConfigTable from './template-config-table.vue';
  import { cloneDeep } from 'lodash';
  import { storeToRefs } from 'pinia';

  const { t, locale } = useI18n();

  const serviceStore = useServiceStore();

  const { spaceFeatureFlags } = storeToRefs(useGlobalStore());

  const props = defineProps<{
    show: boolean;
    bkBizId: string;
    appId: number;
  }>();
  const emits = defineEmits(['update:show', 'confirm']);

  const isFormChange = ref(false);
  const importType = ref('localFile');
  const loading = ref(false);
  const importFromTemplateRef = ref();
  const versionListLoading = ref(false);
  const selectVerisonId = ref();
  const versionList = ref<IConfigVersion[]>([]);
  const tableLoading = ref(false);
  const existConfigList = ref<IConfigImportItem[]>([]);
  const nonExistConfigList = ref<IConfigImportItem[]>([]);
  const existTemplateConfigList = ref<ImportTemplateConfigItem[]>([]);
  const nonExistTemplateConfigList = ref<ImportTemplateConfigItem[]>([]);
  const allConfigList = ref<IConfigImportItem[]>([]);
  const allTemplateConfigList = ref<ImportTemplateConfigItem[]>([]);
  const isClearDraft = ref(false);
  const uploadFileLoading = ref(false);
  const decompressing = ref(false); // 后台压缩包解压
  const fileProcessing = ref(false); // 后台文件处理
  const closeLoading = ref(false);
  const selectedConfigIds = ref<(string | number)[]>([]);
  const configSelectRef = ref();
  const lastSelectedConfigIds = ref<(string | number)[]>([]); // 上一次选中导入的配置项
  const templateImportBtnDisabled = ref(true); // 从配置模板导入按钮是否禁用

  const confirmBtnDisabled = computed(() => {
    if (importType.value === 'configTemplate' && importFromTemplateRef.value) {
      return templateImportBtnDisabled.value;
    }
    if (importType.value === 'localFile') {
      return (
        uploadFileLoading.value ||
        decompressing.value ||
        importConfigList.value.length + importTemplateConfigList.value.length === 0 ||
        isExceedMaxFileCount.value
      );
    }
    return importConfigList.value.length + importTemplateConfigList.value.length === 0 || hasError.value;
  });

  const importConfigList = computed(() => [...nonExistConfigList.value, ...existConfigList.value]);
  const importTemplateConfigList = computed(() => [
    ...nonExistTemplateConfigList.value,
    ...existTemplateConfigList.value,
  ]);

  const hasError = computed(() => {
    return importTemplateConfigList.value.some(
      (config) => !config.template_space_exist || !config.template_set_exist || config.template_set_is_empty,
    );
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
    () => importConfigList.value.length > spaceFeatureFlags.value.RESOURCE_LIMIT.AppConfigCnt,
  );

  watch(
    () => props.show,
    (val) => {
      if (val) {
        importType.value = 'localFile';
        isFormChange.value = false;
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

  // 配置文件绝对路径
  const fileAP = (config: IConfigImportItem) => {
    const { path, name } = config;
    if (path.endsWith('/')) {
      return `${path}${name}`;
    }
    return `${path}/${name}`;
  };

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
    isFormChange.value = true;
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
        allConfigList.value.push(config);
        selectedConfigIds.value.push(item.id);
      });
      res.data.template_configs.forEach((item: ImportTemplateConfigItem) => {
        if (item.is_exist) {
          existTemplateConfigList.value.push(item);
        } else {
          nonExistTemplateConfigList.value.push(item);
        }
        allTemplateConfigList.value.push(item);
        selectedConfigIds.value.push(`${item.template_space_id} - ${item.template_set_id}`);
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
    selectedConfigIds.value = selectedConfigIds.value.filter((id) => {
      if (typeof id === 'number') {
        return importConfigList.value.some((config) => config.id === id);
      }
      return true;
    });
    isFormChange.value = true;
  };

  const handleTemplateTableChange = (deleteId: string, isNonExistData: boolean) => {
    if (isNonExistData) {
      const index = nonExistTemplateConfigList.value.findIndex(
        (config) => `${config.template_space_id} - ${config.template_set_id}` === deleteId,
      );
      nonExistTemplateConfigList.value.splice(index, 1);
    } else {
      const index = existTemplateConfigList.value.findIndex(
        (config) => `${config.template_space_id} - ${config.template_set_id}` === deleteId,
      );
      existTemplateConfigList.value.splice(index, 1);
    }
    selectedConfigIds.value = selectedConfigIds.value.filter((id) => id !== deleteId);
    isFormChange.value = true;
  };

  // 上传文件获取表格数据
  const handleUploadFile = (exist: IConfigImportItem[], nonExist: IConfigImportItem[]) => {
    isFormChange.value = true;
    existConfigList.value = [...existConfigList.value, ...exist];
    nonExistConfigList.value = [...nonExistConfigList.value, ...nonExist];
    allConfigList.value = [...allConfigList.value, ...exist, ...nonExist];
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
    allConfigList.value = [];
    allTemplateConfigList.value = [];
    selectedConfigIds.value = [];
  };

  const handleBeforeClose = async () => {
    if (isFormChange.value) {
      const result = await useModalCloseConfirmation();
      return result;
    }
    return true;
  };

  const handleClose = async () => {
    closeLoading.value = true;
    handleClearTable();
    emits('update:show', false);
  };

  const handleCloseConfigSelect = () => {
    configSelectRef.value.hidePopover();
    selectedConfigIds.value = cloneDeep(lastSelectedConfigIds.value);
  };

  const handleToggleConfigSelectShow = (isShow: boolean) => {
    if (isShow) {
      lastSelectedConfigIds.value = cloneDeep(selectedConfigIds.value);
    }
  };

  const handleConfirmSelect = () => {
    selectedConfigIds.value.forEach((id) => {
      // 配置文件被删除后重新添加
      if (typeof id === 'number') {
        // 非模板配置文件
        const findConfig = importConfigList.value.find((config) => config.id === id);
        if (!findConfig) {
          const config = allConfigList.value.find((config) => config.id === id);
          if (config?.is_exist) {
            existConfigList.value.push(config);
          } else {
            nonExistConfigList.value.push(config!);
          }
        }
      } else {
        // 模板配置文件
        const findConfig = importTemplateConfigList.value.find(
          (config) => `${config.template_space_id} - ${config.template_set_id}` === id,
        );
        if (!findConfig) {
          const config = allTemplateConfigList.value.find(
            (config) => `${config.template_space_id} - ${config.template_set_id}` === id,
          );
          if (config?.is_exist) {
            existTemplateConfigList.value.push(config);
          } else {
            nonExistTemplateConfigList.value.push(config!);
          }
        }
      }
    });

    // 删除已选配置文件
    importConfigList.value.forEach((config) => {
      if (!selectedConfigIds.value.includes(config.id)) {
        if (config.is_exist) {
          const index = existConfigList.value.findIndex((item) => item.id === config.id);
          if (index !== -1) {
            existConfigList.value.splice(index, 1);
          }
        } else {
          const index = nonExistConfigList.value.findIndex((item) => item.id === config.id);
          if (index !== -1) {
            nonExistConfigList.value.splice(index, 1);
          }
        }
      }
    });

    importTemplateConfigList.value.forEach((config) => {
      if (!selectedConfigIds.value.includes(`${config.template_space_id} - ${config.template_set_id}`)) {
        if (config.is_exist) {
          const index = existTemplateConfigList.value.findIndex(
            (item) =>
              `${item.template_space_id} - ${item.template_set_id}` ===
              `${config.template_space_id} - ${config.template_set_id}`,
          );
          if (index !== -1) {
            existTemplateConfigList.value.splice(index, 1);
          }
        } else {
          const index = nonExistTemplateConfigList.value.findIndex(
            (item) =>
              `${item.template_space_id} - ${item.template_set_id}` ===
              `${config.template_space_id} - ${config.template_set_id}`,
          );
          if (index !== -1) {
            nonExistTemplateConfigList.value.splice(index, 1);
          }
        }
      }
    });
    configSelectRef.value.hidePopover();
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
      position: relative;
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
      .config-select {
        position: absolute;
        right: 0;
        top: 50%;
        transform: translateY(-50%);
        .select-btn {
          min-width: 102px;
          height: 32px;
          background: #ffffff;
          border: 1px solid #c4c6cc;
          border-radius: 2px;
          font-size: 14px;
          color: #63656e;
          line-height: 32px;
          text-align: center;
          cursor: pointer;
        }
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

  .config-select-btns {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 0 16px;
    justify-content: flex-end;
    width: 100%;
    height: 100%;
    background: #fafbfd;
  }
</style>

<style lang="scss">
  .import-file-dialog {
    .bk-modal-content {
      height: calc(100% - 50px) !important;
      overflow: auto;
    }
  }

  .config-selector-popover {
    width: 238px !important;
    .bk-select-option {
      padding: 0 12px !important;
    }
  }
</style>
