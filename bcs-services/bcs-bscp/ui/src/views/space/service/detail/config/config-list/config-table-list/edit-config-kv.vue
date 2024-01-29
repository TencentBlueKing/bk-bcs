<template>
  <bk-sideslider
    width="640"
    :title="t('编辑配置项')"
    :is-show="props.show"
    :before-close="handleBeforeClose"
    @closed="close"
  >
    <div class="config-container">
      <ConfigForm
        ref="formRef"
        class="config-form-wrapper"
        :config="(configForm as IConfigKvItem)"
        :content="content"
        :editable="true"
        :bk-biz-id="props.bkBizId"
        :id="props.appId"
        @change="handleChange"
      />
    </div>

    <section class="action-btns">
        <bk-button theme="primary" :loading="pending" @click="handleSubmit"> {{ t('保存') }} </bk-button>
        <bk-button @click="close">{{ t('取消') }}</bk-button>
    </section>
  </bk-sideslider>
</template>
<script setup lang="ts">
import { ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import Message from 'bkui-vue/lib/message';
import ConfigForm from './config-form-kv.vue';
import { updateKv } from '../../../../../../../api/config';
import { IConfigKvItem } from '../../../../../../../../types/config';
import useModalCloseConfirmation from '../../../../../../../utils/hooks/use-modal-close-confirmation';

const { t } = useI18n();
const props = defineProps<{
  bkBizId: string;
  appId: number;
  config: IConfigKvItem;
  show: boolean;
}>();

const emits = defineEmits(['update:show', 'confirm']);

const configForm = ref<IConfigKvItem>();
const content = ref('');
const formRef = ref();
const pending = ref(false);
const isFormChange = ref(false);

watch(
  () => props.show,
  (val) => {
    if (val) {
      isFormChange.value = false;
      configForm.value = props.config;
    }
  },
);

const handleBeforeClose = async () => {
  if (isFormChange.value) {
    const result = await useModalCloseConfirmation();
    return result;
  }
  return true;
};

const handleChange = (data: IConfigKvItem) => {
  configForm.value = data;
  isFormChange.value = true;
};

const handleSubmit = async () => {
  const isValid = await formRef.value.validate();
  if (!isValid) return;
  if (configForm.value!.kv_type === 'number') configForm.value!.value = configForm.value!.value.replace(/^0+(?=\d|$)/, '');
  try {
    pending.value = true;
    await updateKv(props.bkBizId, props.appId, configForm.value!.key, configForm.value!.value);
    emits('confirm');
    close();
    Message({
      theme: 'success',
      message: t('编辑配置项成功'),
    });
  } catch (e) {
    console.error(e);
  } finally {
    pending.value = false;
  }
};

const close = () => {
  emits('update:show', false);
};
</script>
<style lang="scss" scoped>
.config-container {
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
