<template>
  <bk-sideslider title="新建模板套餐" :width="640" :is-show="isShow" :before-close="handleBeforeClose" @closed="close">
    <div class="create-package-form">
      <PackageForm ref="formRef" :space-id="spaceId" :data="data" @change="handleChange" />
    </div>
    <div class="action-btns">
      <bk-button theme="primary" :loading="pending" @click="handleCreate">创建</bk-button>
      <bk-button @click="close">取消</bk-button>
    </div>
  </bk-sideslider>
</template>
<script lang="ts" setup>
import { ref, watch } from 'vue';
import { storeToRefs } from 'pinia';
import Message from 'bkui-vue/lib/message';
import useGlobalStore from '../../../../../store/global';
import { ITemplatePackageEditParams } from '../../../../../../types/template';
import { createTemplatePackage } from '../../../../../api/template';
import useModalCloseConfirmation from '../../../../../utils/hooks/use-modal-close-confirmation';
import PackageForm from './package-form.vue';

const { spaceId } = storeToRefs(useGlobalStore());

const props = defineProps<{
  show: boolean;
  templateSpaceId: number;
}>();

const emits = defineEmits(['update:show', 'created']);

const isShow = ref(false);
const formRef = ref();
const data = ref<ITemplatePackageEditParams>({
  name: '',
  memo: '',
  public: true,
  template_ids: [],
  bound_apps: [],
});
const isFormChange = ref(false);
const pending = ref(false);

watch(
  () => props.show,
  (val) => {
    isShow.value = val;
    isFormChange.value = false;
  },
);

const handleChange = (formData: ITemplatePackageEditParams) => {
  isFormChange.value = true;
  data.value = formData;
};

const handleCreate = () => {
  formRef.value.validate().then(async () => {
    try {
      pending.value = true;
      const res = await createTemplatePackage(spaceId.value, props.templateSpaceId, data.value);
      close();
      emits('created', res.id);
      Message({
        theme: 'success',
        message: '创建成功',
      });
    } catch (e) {
      console.error(e);
    } finally {
      pending.value = false;
    }
  });
};

const handleBeforeClose = async () => {
  if (isFormChange.value) {
    const result = await useModalCloseConfirmation();
    return result;
  }
  return true;
};

const close = () => {
  emits('update:show', false);
  data.value = { name: '', memo: '', public: true, template_ids: [], bound_apps: [] };
};
</script>
<style lang="scss" scoped>
.create-package-form {
  padding: 20px 40px;
  height: calc(100vh - 101px);
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
