<template>
  <bk-sideslider :title="t('编辑变量')" :width="640" :is-show="isShow" :before-close="handleBeforeClose" @closed="close">
    <div class="variable-form">
      <EditingForm ref="formRef" type="edit" :prefix="prefix" :value="variableConfig" @change="handleFormChange" />
    </div>
    <div class="action-btns">
      <bk-button theme="primary" :loading="pending" @click="handleEditSubmit">{{ t('保存') }}</bk-button>
      <bk-button @click="close">{{ t('取消') }}</bk-button>
    </div>
  </bk-sideslider>
</template>
<script lang="ts" setup>
import { ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { storeToRefs } from 'pinia';
import Message from 'bkui-vue/lib/message';
import useModalCloseConfirmation from '../../../utils/hooks/use-modal-close-confirmation';
import useGlobalStore from '../../../store/global';
import { updateVariable } from '../../../api/variable';
import { IVariableEditParams } from '../../../../types/variable';
import EditingForm from './editing-form.vue';

const { spaceId } = storeToRefs(useGlobalStore());
const { t } = useI18n();

const props = defineProps<{
  show: boolean;
  id: number;
  data: IVariableEditParams;
}>();

const emits = defineEmits(['update:show', 'edited']);

const isShow = ref(false);
const isFormChanged = ref(false);
const formRef = ref();
const prefix = ref();
const pending = ref(false);
const variableConfig = ref<IVariableEditParams>({
  name: '',
  type: '',
  default_val: '',
  memo: '',
});

watch(
  () => props.show,
  (val) => {
    isShow.value = val;
    if (val) {
      isFormChanged.value = false;
      const name = props.data.name.replace(/(^bk_bscp_)|(^BK_BSCP_)/, '');
      let currentPrefix = 'bk_bscp_';
      if (/^BK_BSCP_/.test(props.data.name)) {
        currentPrefix = 'BK_BSCP_';
      }
      prefix.value = currentPrefix;
      variableConfig.value = { ...props.data, name };
    }
  },
);

const handleFormChange = (val: IVariableEditParams, localPrefix: string) => {
  isFormChanged.value = true;
  prefix.value = localPrefix;
  variableConfig.value = { ...val };
};

const handleEditSubmit = async () => {
  await formRef.value.validate();
  try {
    pending.value = true;
    const { default_val, memo } = variableConfig.value;
    await updateVariable(spaceId.value, props.id, { default_val, memo });
    close();
    emits('edited');
    Message({
      theme: 'success',
      message: t('编辑变量成功'),
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
.variable-form {
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
