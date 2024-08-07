<template>
  <bk-sideslider
    width="640"
    :title="t('编辑配置文件')"
    :is-show="props.show"
    :before-close="handleBeforeClose"
    @closed="close">
    <bk-loading :loading="configDetailLoading" class="config-loading-container">
      <ConfigForm
        v-if="!configDetailLoading"
        ref="formRef"
        class="config-form-wrapper"
        v-model:fileUploading="fileUploading"
        :config="configForm"
        :content="content"
        :is-edit="true"
        :bk-biz-id="props.bkBizId"
        :id="props.appId"
        :file-size-limit="spaceFeatureFlags.RESOURCE_LIMIT.maxFileSize"
        @change="handleChange" />
    </bk-loading>
    <section class="action-btns">
      <bk-button
        theme="primary"
        :loading="pending"
        :disabled="configDetailLoading || fileUploading"
        @click="handleSubmit">
        {{ t('保存') }}
      </bk-button>
      <bk-button @click="close">{{ t('取消') }}</bk-button>
    </section>
  </bk-sideslider>
</template>
<script setup lang="ts">
  import { ref, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { storeToRefs } from 'pinia';
  import Message from 'bkui-vue/lib/message';
  import ConfigForm from './config-form.vue';
  import {
    getConfigItemDetail,
    getReleasedConfigItemDetail,
    updateConfigContent,
    downloadConfigContent,
    updateServiceConfigItem,
  } from '../../../../../../../api/config';
  import { getConfigEditParams } from '../../../../../../../utils/config';
  import { IConfigEditParams, IFileConfigContentSummary } from '../../../../../../../../types/config';
  import useGlobalStore from '../../../../../../../store/global';
  import useConfigStore from '../../../../../../../store/config';
  import useModalCloseConfirmation from '../../../../../../../utils/hooks/use-modal-close-confirmation';

  const { t } = useI18n();
  const { spaceFeatureFlags } = storeToRefs(useGlobalStore());
  const { versionData } = storeToRefs(useConfigStore());

  const props = defineProps<{
    bkBizId: string;
    appId: number;
    configId: number;
    show: Boolean;
  }>();

  const emits = defineEmits(['update:show', 'confirm']);

  const configDetailLoading = ref(true);
  const configForm = ref<IConfigEditParams>(getConfigEditParams());
  const content = ref<string | IFileConfigContentSummary>('');
  const formRef = ref();
  const fileUploading = ref(false);
  const pending = ref(false);
  const isFormChange = ref(false);

  watch(
    () => props.show,
    (val) => {
      if (val) {
        isFormChange.value = false;
        getConfigDetail();
      }
    },
  );

  // 获取配置文件详情配置及配置内容
  const getConfigDetail = async () => {
    try {
      configDetailLoading.value = true;
      let detail;
      let signature;
      let byte_size;
      if (versionData.value.id) {
        detail = await getReleasedConfigItemDetail(props.bkBizId, props.appId, versionData.value.id, props.configId);
        const { origin_byte_size, origin_signature } = detail.config_item.commit_spec.content;
        byte_size = origin_byte_size;
        signature = origin_signature;
      } else {
        detail = await getConfigItemDetail(props.bkBizId, props.configId, props.appId);
        byte_size = detail.content.byte_size;
        signature = detail.content.signature;
      }
      const { name, memo, path, file_type, permission } = detail.config_item.spec;
      configForm.value = { id: props.configId, name, memo, file_type, path, ...permission };

      if (file_type === 'binary') {
        content.value = { name, signature, size: byte_size };
      } else {
        const configContent = await downloadConfigContent(props.bkBizId, props.appId, signature);
        content.value = String(configContent);
      }
    } catch (e) {
      console.error(e);
    } finally {
      configDetailLoading.value = false;
    }
  };

  const handleBeforeClose = async () => {
    if (isFormChange.value || fileUploading.value) {
      const result = await useModalCloseConfirmation();
      return result;
    }
    return true;
  };

  const handleChange = (data: IConfigEditParams, configContent: IFileConfigContentSummary | string) => {
    configForm.value = data;
    content.value = configContent;
    isFormChange.value = true;
  };

  const handleSubmit = async () => {
    const isValid = await formRef.value.validate();
    if (!isValid) return;

    try {
      pending.value = true;
      const sign = await formRef.value.getSignature();
      let size = 0;
      if (configForm.value.file_type === 'binary') {
        size = Number((content.value as IFileConfigContentSummary).size);
      } else {
        const stringContent = content.value as string;
        size = new Blob([stringContent]).size;
        await updateConfigContent(props.bkBizId, props.appId, stringContent, sign);
      }
      const params = { ...configForm.value, ...{ sign, byte_size: size } };
      await updateServiceConfigItem(props.configId, props.appId, props.bkBizId, params);
      emits('confirm');
      close();
      Message({
        theme: 'success',
        message: t('编辑配置文件成功'),
      });
    } catch (e) {
      console.log(e);
    } finally {
      pending.value = false;
    }
  };

  const close = () => {
    content.value = '';
    configForm.value = getConfigEditParams();
    emits('update:show', false);
  };
</script>
<style lang="scss" scoped>
  .config-loading-container {
    height: calc(100vh - 101px);
    overflow: auto;
    .config-form-wrapper {
      padding: 20px 40px;
      height: 100%;
    }
  }
  .action-btns {
    border-top: 1px solid #dcdee5;
    padding: 8px 24px;
    .bk-button {
      margin-right: 8px;
      min-width: 88px;
    }
  }
</style>
