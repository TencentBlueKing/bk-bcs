<template>
  <bk-dialog
    :title="t('新建空间')"
    ext-cls="create-template-space-dialog"
    :confirm-text="t('创建')"
    :cancel-text="t('取消')"
    :width="480"
    :is-show="isShow"
    :is-loading="pending"
    :esc-close="false"
    :quick-close="false"
    @confirm="handleCreate"
    @closed="handleClose"
  >
    <bk-form ref="formRef" form-type="vertical" :model="localVal">
      <bk-form-item :label="t('模板空间名称')" required property="name">
        <bk-input v-model="localVal.name" />
      </bk-form-item>
      <bk-form-item :label="t('模板空间描述')" property="memo">
        <bk-input v-model="localVal.memo" type="textarea" :rows="6" :maxlength="100" :resize="true" />
      </bk-form-item>
    </bk-form>
  </bk-dialog>
</template>
<script lang="ts" setup>
import { ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { storeToRefs } from 'pinia';
import Message from 'bkui-vue/lib/message';
import useGlobalStore from '../../../../../store/global';
import { createTemplateSpace } from '../../../../../api/template';

const { spaceId } = storeToRefs(useGlobalStore());
const { t } = useI18n();

const props = defineProps<{
  show: boolean;
}>();

const emits = defineEmits(['update:show', 'created']);

const isShow = ref(false);
const formRef = ref();
const localVal = ref({
  name: '',
  memo: '',
});
const pending = ref(false);

watch(
  () => props.show,
  (val) => {
    isShow.value = val;
    if (val) {
      localVal.value = { name: '', memo: '' };
    }
  },
);

const handleCreate = () => {
  formRef.value.validate().then(async () => {
    try {
      pending.value = true;
      const res = await createTemplateSpace(spaceId.value, localVal.value);
      handleClose();
      emits('created', res.id);
      Message({
        theme: 'success',
        message: t('创建空间成功'),
      });
    } catch (e) {
      console.error(e);
    } finally {
      pending.value = false;
    }
  });
};

const handleClose = () => {
  emits('update:show', false);
};
</script>

<style lang="scss" scoped>
.bk-input {
  height: 33px;
}
</style>
