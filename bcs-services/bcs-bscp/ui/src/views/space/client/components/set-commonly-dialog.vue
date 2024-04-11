<template>
  <bk-dialog
    :is-show="isShow"
    :title="isCreate ? $t('设为常用') : $t('重命名')"
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
  }>();
  const emits = defineEmits(['close', 'update', 'create']);

  const formData = ref({
    name: '',
  });
  const formRef = ref();

  watch(
    () => props.isShow,
    (val) => {
      if (val) formData.value.name = '';
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

<style scoped lang="scss"></style>
