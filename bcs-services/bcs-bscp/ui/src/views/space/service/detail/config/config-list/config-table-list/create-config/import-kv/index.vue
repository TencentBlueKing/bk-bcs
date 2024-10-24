<template>
  <bk-dialog
    :is-show="props.show"
    :title="t('批量导入')"
    :theme="'primary'"
    width="1200"
    height="720"
    ext-cls="import-kv-dialog"
    :before-close="handleBeforeClose"
    :quick-close="false"
    @closed="handleClose">
    <div :class="['select-wrap', { 'en-select-wrap': locale === 'en' }]">
      <div class="import-type-select">
        <div class="label">{{ t('导入方式') }}</div>
        <bk-radio-group v-model="importType">
          <bk-radio-button label="text">{{ t('文本格式导入') }}</bk-radio-button>
          <bk-radio-button label="historyVersion">{{ t('从历史版本导入') }}</bk-radio-button>
          <bk-radio-button label="otherService">{{ t('从其他服务导入') }}</bk-radio-button>
        </bk-radio-group>
      </div>
      <div v-if="importType === 'text'">
        <TextImport ref="textImport" :bk-biz-id="bkBizId" :app-id="appId" />
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
      <div v-else>
        <ImportFormOtherService
          :bk-biz-id="bkBizId"
          :app-id="appId"
          @select-version="handleSelectVersion"
          @clear="handleClearTable" />
      </div>
    </div>
    <div v-if="importType !== 'text' && allConfigList.length" class="content">
      <bk-loading :loading="tableLoading">
        <div class="head">
          <bk-checkbox style="margin-left: 24px" v-model="isClearDraft"> {{ $t('导入前清空草稿区') }} </bk-checkbox>
          <div v-if="!isClearDraft" class="tips">
            {{ t('共将导入') }} <span style="color: #3a84ff">{{ importConfigList.length }}</span>
            {{ t('个配置项，其中') }} <span style="color: #ffa519">{{ existConfigList.length }}</span>
            {{ t('个已存在,导入后将') }}
            <span style="color: #ffa519">{{ t('覆盖原配置') }}</span>
          </div>
          <div v-else class="tips">
            {{ t('将') }} <span style="color: #ffa519">{{ t('清空') }}</span> {{ t('现有草稿区,并导入') }}
            <span style="color: #3a84ff">{{ importConfigList.length }}</span>
            {{ t('个配置项') }}
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
              <div class="select-btn">{{ $t('选择配置项') }}</div>
            </template>
            <bk-option v-for="(item, index) in allConfigList" :id="item.key" :key="index" :label="item.key" />
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
      </bk-loading>
    </div>
    <template #footer>
      <bk-button
        theme="primary"
        style="margin-right: 8px"
        :disabled="confirmBtnDisabled || loading"
        :loading="loading"
        @click="handleConfirm">
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
    getConfigVersionList,
    importKvFromHistoryVersion,
    importKvFormText,
  } from '../../../../../../../../../api/config';
  import { IConfigVersion, IConfigKvItem } from '../../../../../../../../../../types/config';
  import { Message } from 'bkui-vue';
  import TextImport from './text-import.vue';
  import ImportFormOtherService from '../import-file/import-form-other-service.vue';
  import useModalCloseConfirmation from '../../../../../../../../../utils/hooks/use-modal-close-confirmation';
  import ConfigTable from './kv-config-table.vue';
  import { cloneDeep } from 'lodash';
  import useServiceStore from '../../../../../../../../../store/service';

  const serviceStore = useServiceStore();

  const { t, locale } = useI18n();
  const props = defineProps<{
    show: boolean;
    bkBizId: string;
    appId: number;
  }>();
  const emits = defineEmits(['update:show', 'confirm']);

  const isFormChange = ref(false);
  const importType = ref('text');
  const loading = ref(false);
  const selectVerisonId = ref();
  const versionListLoading = ref(false);
  const versionList = ref<IConfigVersion[]>([]);
  const tableLoading = ref(false);
  const existConfigList = ref<IConfigKvItem[]>([]);
  const nonExistConfigList = ref<IConfigKvItem[]>([]);
  const isClearDraft = ref(false);
  const expandNonExistTable = ref(true);
  const expandExistTable = ref(true);
  const textImport = ref();
  const selectedConfigIds = ref<string[]>([]);
  const allConfigList = ref<IConfigKvItem[]>([]);
  const configSelectRef = ref();
  const lastSelectedConfigIds = ref<string[]>([]); // 上一次选中导入的配置项

  watch(
    () => props.show,
    (val) => {
      if (val) {
        importType.value = 'text';
        isFormChange.value = false;
        nonExistConfigList.value = [];
        existConfigList.value = [];
        allConfigList.value = [];
        selectedConfigIds.value = [];
        getVersionList();
      }
    },
  );

  watch(
    () => importType.value,
    () => {
      nonExistConfigList.value = [];
      existConfigList.value = [];
      selectVerisonId.value = undefined;
      allConfigList.value = [];
      selectedConfigIds.value = [];
    },
  );

  onMounted(() => {
    getVersionList();
  });

  const confirmBtnDisabled = computed(() => {
    if (textImport.value && importType.value === 'text') return textImport.value.hasError();
    return !importConfigList.value.length;
  });

  const importConfigList = computed(() => [...existConfigList.value, ...nonExistConfigList.value]);

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
      const params = { other_app_id, release_id };
      const res = await importKvFromHistoryVersion(props.bkBizId, props.appId, params);
      existConfigList.value = res.data.exist;
      nonExistConfigList.value = res.data.non_exist;
      existConfigList.value = existConfigList.value.map((item) => ({ ...item, is_exist: true }));
      nonExistConfigList.value = nonExistConfigList.value.map((item) => ({ ...item, is_exist: false }));
      allConfigList.value = [...existConfigList.value, ...nonExistConfigList.value];
      selectedConfigIds.value = allConfigList.value.map((item) => item.key);
    } catch (e) {
      console.error(e);
    } finally {
      tableLoading.value = false;
    }
  };

  const handleBeforeClose = async () => {
    if (isFormChange.value) {
      const result = await useModalCloseConfirmation();
      return result;
    }
    return true;
  };

  const handleClose = () => {
    emits('update:show', false);
  };

  const handleConfirm = async () => {
    loading.value = true;
    try {
      if (importType.value === 'text') {
        await textImport.value.handleImport();
      } else {
        const res = await importKvFormText(props.bkBizId, props.appId, importConfigList.value, isClearDraft.value);
        serviceStore.$patch((state) => {
          state.topIds = res.data.ids;
        });
      }
      emits('update:show', false);
      setTimeout(() => {
        emits('confirm');
        Message({
          theme: 'success',
          message: t('配置项导入成功'),
        });
      }, 300);
    } catch (error) {
      console.error(error);
    } finally {
      loading.value = false;
    }
  };

  const handleTableChange = (data: IConfigKvItem[], isNonExistData: boolean) => {
    if (isNonExistData) {
      nonExistConfigList.value = data;
    } else {
      existConfigList.value = data;
    }
    selectedConfigIds.value = selectedConfigIds.value.filter((key) => {
      return importConfigList.value.some((config) => config.key === key);
    });
    isFormChange.value = true;
  };

  const handleClearTable = () => {
    nonExistConfigList.value = [];
    existConfigList.value = [];
  };

  const handleConfirmSelect = () => {
    // 配置项添加
    selectedConfigIds.value.forEach((key) => {
      const findConfig = importConfigList.value.find((item) => item.key === key);
      if (!findConfig) {
        const addConfig = allConfigList.value.find((item) => item.key === key);
        console.log(addConfig);
        if (addConfig?.is_exist) {
          existConfigList.value.push(addConfig!);
        } else {
          nonExistConfigList.value.push(addConfig!);
        }
      }
    });

    // 配置项删除
    importConfigList.value.forEach((config) => {
      if (!selectedConfigIds.value.includes(config.key)) {
        if (config.is_exist) {
          existConfigList.value = existConfigList.value.filter((item) => item.key !== config.key);
        } else {
          nonExistConfigList.value = nonExistConfigList.value.filter((item) => item.key !== config.key);
        }
      }
    });

    configSelectRef.value.hidePopover();
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
</script>

<style scoped lang="scss">
  .select-wrap {
    .import-type-select {
      display: flex;
    }
    .label {
      width: 70px;
      height: 32px;
      line-height: 32px;
      font-size: 12px;
      color: #63656e;
      text-align: right;
      margin-right: 22px;
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
    &.en-select-wrap {
      .label {
        width: 100px !important;
      }
      :deep(.wrap) {
        .label {
          @extend .label;
        }
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
  }

  .content {
    margin-top: 24px;
    border-top: 1px solid #dcdee5;
    overflow: auto;
    max-height: 490px;
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
  .import-kv-dialog {
    .bk-modal-content {
      height: calc(100% - 50px) !important;
      overflow: hidden !important;
    }
  }
  .config-selector-popover {
    width: 238px !important;
    .bk-select-option {
      padding: 0 12px !important;
    }
  }
</style>
