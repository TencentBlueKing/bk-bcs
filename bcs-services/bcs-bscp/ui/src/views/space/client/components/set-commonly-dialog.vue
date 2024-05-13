<template>
  <bk-dialog
    :is-show="isShow"
    :title="isCreate ? $t('设为常用') : $t('重命名')"
    ext-cls="set-commonly-dialog"
    theme="primary"
    confirm-text="保存"
    @closed="handleClose"
    @confirm="handleConfirm">
    <bk-form ref="formRef" :model="formData">
      <bk-form-item :label="$t('名称')" property="name" label-width="80" required>
        <bk-input v-model="formData.name"></bk-input>
      </bk-form-item>
    </bk-form>
  </bk-dialog>
</template>

<script lang="ts" setup>
  import { ref, watch } from 'vue';
  const props = defineProps<{
    isShow: boolean;
    isCreate: boolean;
    name?: string;
  }>();
  const emits = defineEmits(['close', 'update', 'create']);

  const formData = ref({
    name: '',
  });
  const formRef = ref();

  watch(
    () => props.isShow,
    (val) => {
      if (val) {
        formData.value.name = props.isCreate ? '' : props.name || '';
      }
    },
  );

  const handleConfirm = async () => {
    const isValid = await formRef.value.validate();
    if (!isValid) return;
    props.isCreate ? emits('create', formData.value.name) : emits('update', formData.value.name);
  };

  const handleClose = () => {
    emits('close');
  };
</script>

<style lang="scss">
  .set-commonly-dialog {
    .bk-modal-body .bk-modal-content {
      min-height: 100px !important;
      display: flex;
      align-items: center;
      .bk-input--text {
        width: 274px;
      }
    }
  }
</style>
