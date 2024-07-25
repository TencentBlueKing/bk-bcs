<template>
  <bk-sideslider
    width="640"
    :title="t('新增配置文件')"
    :is-show="props.show"
    :before-close="handleBeforeClose"
    @closed="close">
    <ConfigForm
      ref="formRef"
      class="config-form-wrapper"
      v-model:fileUploading="fileUploading"
      :config="configForm"
      :content="content"
      :is-edit="false"
      :bk-biz-id="props.bkBizId"
      :id="props.appId"
      :file-size-limit="spaceFeatureFlags.RESOURCE_LIMIT.maxFileSize"
      @change="handleFormChange" />
    <section class="action-btns">
      <bk-button theme="primary" :loading="pending" :disabled="fileUploading" @click="handleSubmit">
        {{ t('保存') }}
      </bk-button>
      <bk-button @click="close">{{ t('取消') }}</bk-button>
    </section>
  </bk-sideslider>
</template>
<script lang="ts" setup>
  import { ref, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { storeToRefs } from 'pinia';
  import Message from 'bkui-vue/lib/message';
  import { IConfigEditParams, IFileConfigContentSummary } from '../../../../../../../../../types/config';
  import { createServiceConfigItem, updateConfigContent } from '../../../../../../../../api/config';
  import { getConfigEditParams } from '../../../../../../../../utils/config';
  import useModalCloseConfirmation from '../../../../../../../../utils/hooks/use-modal-close-confirmation';
  import ConfigForm from '../config-form.vue';
  import useServiceStore from '../../../../../../../../store/service';
  import useGlobalStore from '../../../../../../../../store/global';

  const { spaceFeatureFlags } = storeToRefs(useGlobalStore());

  const props = defineProps<{
    show: boolean;
    bkBizId: string;
    appId: number;
  }>();

  const serviceStore = useServiceStore();
  const { lastCreatePermission } = storeToRefs(serviceStore);
  const { t } = useI18n();

  const emits = defineEmits(['update:show', 'confirm']);
  const fileUploading = ref(false);
  const pending = ref(false);
  const content = ref<IFileConfigContentSummary | string>('');
  const formRef = ref();
  const isFormChange = ref(false);
  const configForm = ref<IConfigEditParams>(getConfigEditParams());
  watch(
    () => props.show,
    (val) => {
      if (val) {
        configForm.value = Object.assign(getConfigEditParams(), lastCreatePermission.value);
        content.value = '';
        isFormChange.value = false;
      }
    },
  );

  const handleFormChange = (data: IConfigEditParams, configContent: IFileConfigContentSummary | string) => {
    configForm.value = data;
    const { privilege, user, user_group } = data;
    serviceStore.$patch((state) => {
      state.lastCreatePermission = {
        privilege: privilege as string,
        user: user as string,
        user_group: user_group as string,
      };
    });
    content.value = configContent;
    isFormChange.value = true;
  };

  const handleBeforeClose = async () => {
    if (isFormChange.value) {
      const result = await useModalCloseConfirmation();
      return result;
    }
    return true;
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
        await updateConfigContent(props.bkBizId, props.appId, stringContent, sign, () => {});
      }
      const params = { ...configForm.value, ...{ sign, byte_size: size } };
      const res = await createServiceConfigItem(props.appId, props.bkBizId, params);
      serviceStore.$patch((state) => {
        state.topIds = [res.data.id];
      });
      emits('confirm');
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

  const close = () => {
    emits('update:show', false);
  };
</script>
<style lang="scss" scoped>
  .config-form-wrapper {
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
