<template>
  <bk-dialog
    :is-show="isShow"
    :title="title"
    :theme="'primary'"
    quick-close
    ext-cls="delete-confirm-dialog"
    @closed="handleClose">
    <slot></slot>
    <template #footer>
      <bk-button
        theme="primary"
        @click="emits('confirm')"
        :disabled="pending"
        :loading="pending"
        style="margin-right: 8px">
        {{ confirmText || t('删除') }}
      </bk-button>
      <bk-button @click="handleClose">{{ t('取消') }}</bk-button>
    </template>
  </bk-dialog>
</template>

<script lang="ts" setup>
  import { useI18n } from 'vue-i18n';
  const { t } = useI18n();
  withDefaults(
    defineProps<{
      isShow: boolean;
      title: string;
      pending?: boolean;
      confirmText?: string;
    }>(),
    {
      pending: false,
    },
  );

  const handleClose = () => {
    emits('close');
    emits('update:isShow', false);
  };
  const emits = defineEmits(['update:isShow', 'confirm', 'close']);
</script>

<style lang="scss">
  .delete-confirm-dialog {
    .bk-modal-body {
      .bk-dialog-header .bk-dialog-title {
        white-space: normal;
      }
    }
  }
</style>
