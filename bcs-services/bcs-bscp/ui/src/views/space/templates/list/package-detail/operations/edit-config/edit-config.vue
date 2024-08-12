<template>
  <bk-sideslider
    :title="t('编辑配置文件')"
    :width="640"
    :is-show="props.show"
    :quick-close="!isSelectPkgDialogShow"
    :before-close="handleBeforeClose"
    @closed="close">
    <bk-loading :loading="configDetailLoading" class="config-loading-container">
      <ConfigForm
        v-if="!configDetailLoading"
        ref="formRef"
        v-model:fileUploading="fileUploading"
        class="config-form-wrapper"
        :config="configForm"
        :content="content"
        :is-edit="true"
        :is-tpl="true"
        :bk-biz-id="spaceId"
        :id="currentTemplateSpace"
        :file-size-limit="spaceFeatureFlags.RESOURCE_LIMIT.maxFileSize"
        @change="handleFormChange" />
    </bk-loading>
    <div class="action-btns">
      <bk-button :loading="submitPending" theme="primary" @click="handleCreateConfirm">{{ t('保存') }}</bk-button>
      <bk-button @click="close">{{ t('取消') }}</bk-button>
    </div>
  </bk-sideslider>
</template>
<script lang="ts" setup>
  import { ref, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { storeToRefs } from 'pinia';
  import Message from 'bkui-vue/lib/message';
  import useGlobalStore from '../../../../../../../store/global';
  import useTemplateStore from '../../../../../../../store/template';
  import {
    updateTemplateConfig,
    getTemplateConfigMeta,
    downloadTemplateContent,
    updateTemplateContent,
  } from '../../../../../../../api/template';
  import { IConfigEditParams, IFileConfigContentSummary } from '../../../../../../../../types/config';
  import { getConfigEditParams } from '../../../../../../../utils/config';
  import { ITemplateVersionEditingData } from '../../../../../../../../types/template';
  import useModalCloseConfirmation from '../../../../../../../utils/hooks/use-modal-close-confirmation';
  import ConfigForm from '../../../../../service/detail/config/config-list/config-table-list/config-form.vue';

  const { spaceId, spaceFeatureFlags } = storeToRefs(useGlobalStore());
  const { currentTemplateSpace } = storeToRefs(useTemplateStore());
  const { t } = useI18n();

  const props = defineProps<{
    id: number;
    spaceId: string;
    show: Boolean;
    memo: string;
  }>();

  const emits = defineEmits(['update:show', 'added', 'edited']);

  const configForm = ref<IConfigEditParams>(getConfigEditParams());
  const fileUploading = ref(false);
  const content = ref<IFileConfigContentSummary | string>('');
  const formRef = ref();
  const submitPending = ref(false);
  const isSelectPkgDialogShow = ref(false);
  const isFormChanged = ref(false);
  const configDetailLoading = ref(false);

  watch(
    () => props.show,
    (val) => {
      if (val) {
        isFormChanged.value = false;
        getConfigDetail();
      }
    },
  );

  const handleFormChange = (data: IConfigEditParams, configContent: IFileConfigContentSummary | string) => {
    configForm.value = data;
    content.value = configContent;
    isFormChanged.value = true;
  };

  // 获取配置文件详情配置及配置内容
  const getConfigDetail = async () => {
    try {
      configDetailLoading.value = true;
      const res = await getTemplateConfigMeta(props.spaceId, props.id);
      const { name, path, file_type, user, user_group, privilege, signature, byte_size, template_revision_id } =
        res.data.detail;
      configForm.value = {
        ...configForm.value,
        id: props.id,
        name,
        memo: props.memo,
        file_type,
        path,
        user,
        user_group,
        privilege,
        template_revision_id,
      };
      if (file_type === 'binary') {
        content.value = { name, signature, size: byte_size };
      } else {
        const configContent = await downloadTemplateContent(props.spaceId, currentTemplateSpace.value, signature);
        content.value = String(configContent);
      }
    } catch (e) {
      console.error(e);
    } finally {
      configDetailLoading.value = false;
    }
  };

  const handleCreateConfirm = async () => {
    const isValid = await formRef.value.validate();
    if (!isValid) return;
    try {
      submitPending.value = true;
      const sign = await formRef.value.getSignature();
      let size = 0;
      if (configForm.value.file_type === 'binary') {
        size = Number((content.value as IFileConfigContentSummary).size);
      } else {
        const stringContent = content.value as string;
        size = new Blob([stringContent]).size;
        await updateTemplateContent(props.spaceId, currentTemplateSpace.value, stringContent, sign);
      }
      const { memo, file_type, file_mode, user, user_group, privilege, revision_name, template_revision_id } =
        configForm.value;
      const formData = {
        revision_memo: memo,
        file_type,
        file_mode,
        user,
        user_group,
        privilege,
        sign,
        byte_size: size,
        revision_name,
        template_revision_id,
      };
      await updateTemplateConfig(
        props.spaceId,
        currentTemplateSpace.value,
        props.id,
        formData as ITemplateVersionEditingData,
      );
      Message({
        theme: 'success',
        message: t('编辑配置文件成功'),
      });
      close();
      emits('edited');
    } catch (e) {
      console.log(e);
    } finally {
      submitPending.value = false;
    }
  };

  const handleBeforeClose = async () => {
    if (isFormChanged.value || fileUploading.value) {
      const result = await useModalCloseConfirmation();
      return result;
    }
    return true;
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
