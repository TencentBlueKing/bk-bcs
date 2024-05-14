<template>
  <bk-sideslider
    :title="t('克隆模板套餐')"
    :width="640"
    :is-show="isShow"
    :before-close="handleBeforeClose"
    @closed="close">
    <div class="package-form">
      <PackageForm ref="formRef" :space-id="spaceId" :data="data" @change="handleChange" />
    </div>
    <div class="action-btns">
      <bk-button theme="primary" :loading="pending" @click="handleSave">{{ t('创建') }}</bk-button>
      <bk-button @click="close">{{ t('取消') }}</bk-button>
    </div>
  </bk-sideslider>
</template>
<script lang="ts" setup>
  import { ref, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { storeToRefs } from 'pinia';
  import Message from 'bkui-vue/lib/message';
  import useGlobalStore from '../../../../../store/global';
  import { ITemplatePackageEditParams, ITemplatePackageItem } from '../../../../../../types/template';
  import { createTemplatePackage } from '../../../../../api/template';
  import useModalCloseConfirmation from '../../../../../utils/hooks/use-modal-close-confirmation';
  import PackageForm from './package-form.vue';

  const { spaceId } = storeToRefs(useGlobalStore());
  const { t } = useI18n();

  const props = defineProps<{
    show: boolean;
    templateSpaceId: number;
    pkg: ITemplatePackageItem;
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
      const { name, memo, public: isPublic, bound_apps, template_ids } = props.pkg.spec;
      data.value = { name, memo, public: isPublic, bound_apps, template_ids };
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
        await createTemplatePackage(spaceId.value, props.templateSpaceId, data.value);
        close();
        emits('created');
        Message({
          theme: 'success',
          message: t('克隆成功'),
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
