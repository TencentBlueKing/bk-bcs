<template>
  <bk-sideslider
    :title="t('新建配置文件')"
    :width="640"
    :is-show="props.show"
    :quick-close="!isSelectPkgDialogShow"
    :before-close="handleBeforeClose"
    @closed="close">
    <div class="slider-content-container">
      <ConfigForm
        ref="formRef"
        v-model:fileUploading="fileUploading"
        :config="configForm"
        :content="content"
        :editable="true"
        :is-tpl="true"
        :bk-biz-id="spaceId"
        :id="currentTemplateSpace"
        @change="handleFormChange" />
    </div>
    <div class="action-btns">
      <bk-button theme="primary" @click="handleCreateClick">{{ t('去创建') }}</bk-button>
      <bk-button @click="close">{{ t('取消') }}</bk-button>
    </div>
  </bk-sideslider>
  <SelectPackage
    v-model:show="isSelectPkgDialogShow"
    :pending="pending"
    :config-form="configForm"
    @confirm="handleCreateConfirm" />
</template>
<script lang="ts" setup>
  import { ref, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { storeToRefs } from 'pinia';
  import Message from 'bkui-vue/lib/message';
  import useGlobalStore from '../../../../../../../../store/global';
  import useTemplateStore from '../../../../../../../../store/template';
  import { updateTemplateContent, createTemplate, addTemplateToPackage } from '../../../../../../../../api/template';
  import { IConfigEditParams, IFileConfigContentSummary } from '../../../../../../../../../types/config';
  import { getConfigEditParams } from '../../../../../../../../utils/config';
  import useModalCloseConfirmation from '../../../../../../../../utils/hooks/use-modal-close-confirmation';
  import ConfigForm from '../../../../../../service/detail/config/config-list/config-table-list/config-form.vue';
  import SelectPackage from './select-package.vue';

  const { spaceId } = storeToRefs(useGlobalStore());
  const { currentTemplateSpace } = storeToRefs(useTemplateStore());
  const { t } = useI18n();

  const props = defineProps<{
    show: boolean;
  }>();

  const emits = defineEmits(['update:show', 'added']);

  const configForm = ref<IConfigEditParams>(getConfigEditParams());
  const fileUploading = ref(false);
  const content = ref<IFileConfigContentSummary | string>('');
  const formRef = ref();
  const pending = ref(false);
  const isSelectPkgDialogShow = ref(false);
  const isFormChanged = ref(false);

  watch(
    () => props.show,
    (val) => {
      if (val) {
        content.value = '';
        configForm.value = getConfigEditParams();
        isFormChanged.value = false;
      }
    },
  );

  const handleFormChange = (data: IConfigEditParams, configContent: IFileConfigContentSummary | string) => {
    configForm.value = data;
    content.value = configContent;
    isFormChanged.value = true;
  };

  const handleCreateClick = async () => {
    const isValid = await formRef.value.validate();
    if (!isValid) return;
    isSelectPkgDialogShow.value = true;
  };

  const handleCreateConfirm = async (pkgIds: number[]) => {
    try {
      pending.value = true;
      const sign = await formRef.value.getSignature();
      let size = 0;
      if (configForm.value.file_type === 'binary') {
        size = Number((content.value as IFileConfigContentSummary).size);
      } else {
        const stringContent = content.value as string;
        size = new Blob([stringContent]).size;
        await updateTemplateContent(spaceId.value, currentTemplateSpace.value, stringContent, sign);
      }
      if (configForm.value.path?.endsWith('/') && configForm.value.path !== '/') {
        configForm.value.path = configForm.value.path.slice(0, -1);
      }
      const params = { ...configForm.value, ...{ sign, byte_size: size } };
      const res = await createTemplate(spaceId.value, currentTemplateSpace.value, params);
      // 选择未指定套餐时,不需要调用添加接口
      if (pkgIds.length > 1 || pkgIds[0] !== 0) {
        await addTemplateToPackage(spaceId.value, currentTemplateSpace.value, [res.data.id], pkgIds);
      }
      isSelectPkgDialogShow.value = false;
      emits('added');
      close();
      Message({
        theme: 'success',
        message: t('新建配置文件成功'),
      });
    } catch (e) {
      console.log(e);
    } finally {
      pending.value = false;
    }
  };

  const handleBeforeClose = async () => {
    if (isFormChanged.value) {
      const result = await useModalCloseConfirmation();
      return result;
    }
    return true;
  };

  const close = () => {
    emits('update:show', false);
  };
</script>
<style lang="scss" scoped>
  .slider-content-container {
    padding: 20px 40px;
    height: calc(100vh - 101px);
    overflow: auto;
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
