<template>
  <bk-dialog
    :is-show="isShow"
    :title="title"
    theme="primary"
    confirm-text="保存"
    @closed="handleClose"
    @confirm="handleConfirm">
    <bk-form ref="formRef" :model="formData">
      <bk-form-item label="名称" property="name" label-width="80" required>
        <bk-input v-model="formData.name"></bk-input>
      </bk-form-item>
    </bk-form>
  </bk-dialog>
</template>

<script lang="ts" setup>
  import { ref } from 'vue';
  const props = defineProps<{
    isShow: boolean;
    title: string;
  }>();
  const emits = defineEmits(['close', 'update', 'create']);

  const formData = ref({
    name: '',
  });
  const formRef = ref();

  const handleConfirm = async () => {
    const isValid = await formRef.value.validate();
    if (!isValid) return;
    props.title === '设为常用' ? emits('create', formData.value.name) : emits('update', formData.value.name);
  };

  const handleClose = () => {
    emits('close');
    formData.value.name = '';
  };
</script>

<style scoped lang="scss"></style>
