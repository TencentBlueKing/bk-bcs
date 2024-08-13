<template>
  <bk-sideslider
    width="960"
    :title="t('新增配置项')"
    :is-show="props.show"
    :before-close="handleBeforeClose"
    @closed="close">
    <ConfigForm
      ref="formRef"
      class="config-form-wrapper"
      :config="configForm"
      :content="content"
      :bk-biz-id="props.bkBizId"
      :id="props.appId"
      @change="handleFormChange" />
    <section class="action-btns">
      <bk-button theme="primary" @click="handleSubmit">{{ t('保存') }}</bk-button>
      <bk-button @click="close">{{ t('取消') }}</bk-button>
    </section>
  </bk-sideslider>
</template>
<script lang="ts" setup>
  import { ref, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import Message from 'bkui-vue/lib/message';
  import { IConfigKvEditParams } from '../../../../../../../../../types/config';
  import { createKv } from '../../../../../../../../api/config';
  import useModalCloseConfirmation from '../../../../../../../../utils/hooks/use-modal-close-confirmation';
  import ConfigForm from '../config-form-kv.vue';
  import useServiceStore from '../../../../../../../../store/service';

  const serviceStore = useServiceStore();

  const props = defineProps<{
    show: boolean;
    bkBizId: string;
    appId: number;
  }>();

  const { t } = useI18n();

  const emits = defineEmits(['update:show', 'confirm']);
  const content = ref('');
  const formRef = ref();
  const isFormChange = ref(false);
  const configForm = ref({
    key: '',
    kv_type: '',
    value: '',
    memo: '',
  });
  watch(
    () => props.show,
    (val) => {
      if (val) {
        configForm.value = {
          key: '',
          kv_type: '',
          value: '',
          memo: '',
        };
        content.value = '';
        isFormChange.value = false;
      }
    },
  );

  const handleFormChange = (data: IConfigKvEditParams, configContent: string) => {
    configForm.value = data;
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
    if (configForm.value.kv_type === 'number') {
      configForm.value.value = configForm.value.value.replace(/^0+(?=\d|$)/, '');
    }
    try {
      const res = await createKv(props.bkBizId, props.appId, { ...configForm.value });
      serviceStore.$patch((state) => {
        state.topIds = [res.data.id];
      });
      emits('confirm');
      close();
      Message({
        theme: 'success',
        message: '新建配置项成功',
      });
    } catch (e) {
      console.log(e);
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
