<template>
  <bk-sideslider title="编辑模板套餐" :width="640" :is-show="isShow" :before-close="handleBeforeClose" @closed="close">
    <div class="package-form">
      <PackageForm ref="formRef" :space-id="spaceId" :data="data" :apps="apps" @change="handleChange" />
    </div>
    <div class="action-btns">
      <bk-button theme="primary" :loading="pending" @click="handleSave">保存</bk-button>
      <bk-button @click="close">取消</bk-button>
    </div>
  </bk-sideslider>
</template>
<script lang="ts" setup>
import { ref, watch } from 'vue';
import { storeToRefs } from 'pinia';
import Message from 'bkui-vue/lib/message';
import useGlobalStore from '../../../../../store/global';
import { ITemplatePackageEditParams, ITemplatePackageItem } from '../../../../../../types/template';
import { updateTemplatePackage } from '../../../../../api/template';
import useModalCloseConfirmation from '../../../../../utils/hooks/use-modal-close-confirmation';
import PackageForm from './package-form.vue';

const { spaceId } = storeToRefs(useGlobalStore());

const props = defineProps<{
  show: boolean;
  templateSpaceId: number;
  pkg: ITemplatePackageItem;
}>();

const emits = defineEmits(['update:show', 'edited']);

const isShow = ref(false);
const formRef = ref();
const data = ref<ITemplatePackageEditParams>({
  name: '',
  memo: '',
  public: true,
  bound_apps: [],
  template_ids: []
});
const apps = ref<number[]>([]);
const isFormChange = ref(false);
const pending = ref(false);

watch(
  () => props.show,
  (val) => {
    isShow.value = val;
    if (val) {
      isFormChange.value = false;
      const { name, memo, public: isPublic, bound_apps, template_ids } = props.pkg.spec;
      data.value = { name, memo, public: isPublic, bound_apps, template_ids };
      apps.value = bound_apps.slice();
    }
  },
);

const handleChange = (formData: ITemplatePackageEditParams) => {
  isFormChange.value = true;
  data.value = formData;
};

const handleSave = () => {
  formRef.value.validate().then(async () => {
    try {
      pending.value = true;
      if (data.value.public === true) {
        data.value.bound_apps = [];
      }

      await updateTemplatePackage(spaceId.value, props.templateSpaceId, props.pkg.id, data.value);
      close();
      emits('edited');
      Message({
        theme: 'success',
        message: '编辑成功',
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
};
</script>
<style lang="scss" scoped>
.package-form {
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
