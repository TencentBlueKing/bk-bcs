<template>
  <bk-dialog
    :is-show="isShow"
    :title="title"
    :theme="'primary'"
    quick-close
    ext-cls="delete-confirm-dialog"
    @closed="handleClose"
  >
    <slot></slot>
    <template #footer>
      <bk-button theme="primary" @click="emits('confirm')" style="margin-right: 8px">{{ confirmText }}</bk-button>
      <bk-button @click="handleClose">取消</bk-button>
    </template>
  </bk-dialog>
</template>

<script lang="ts" setup>
withDefaults(
  defineProps<{
    isShow: boolean;
    title: string;
    confirmText?: string;
  }>(),
  {
    confirmText: '删除',
  },
);

const handleClose = () => {
  emits('close');
  emits('update:isShow', false);
};
const emits = defineEmits(['update:isShow', 'confirm', 'close']);
</script>

<style scoped lang="scss">
.delete-confirm-dialog {
  :deep(.bk-modal-footer) {
    height: auto;
    background-color: #fff;
    border: none;
    padding-bottom: 24px !important;
    .bk-button {
      width: 80px;
    }
  }
}
</style>
