<template>
  <bk-dialog
    :is-show="props.show"
    :width="440"
    :title="props.title"
    :loading="pending"
    :confirm-text="t('确定')"
    :cancel-text="t('取消')"
    @confirm="handleConfirm"
    @closed="handleCancel">
    <div class="confirm-content">
      <p>
        {{ t('版本名称') }}：<span class="name">{{ props.version.spec.name }}</span>
      </p>
      <p>{{ props.tips }}</p>
    </div>
  </bk-dialog>
</template>
<script lang="ts" setup>
  import { ref, watch } from 'vue';
  import { IConfigVersion } from '../../../../../../../types/config';
  import { useI18n } from 'vue-i18n';

  const { t } = useI18n();
  const emits = defineEmits(['confirm', 'cancel', 'update:show']);

  const props = defineProps<{
    show: boolean;
    title: string;
    tips: string;
    version: IConfigVersion;
    confirmFn: Function;
  }>();

  const pending = ref(false);

  watch(
    () => props.show,
    (val) => {
      if (val) {
        pending.value = false;
      }
    },
  );

  const handleConfirm = async () => {
    if (pending.value) {
      return;
    }
    pending.value = true;
    if (typeof props.confirmFn === 'function') {
      await props.confirmFn();
    }
    pending.value = false;
    emits('confirm');
  };

  const handleCancel = () => {
    emits('update:show', false);
    emits('cancel');
  };
</script>
<style lang="scss">
  .confirm-content {
    color: #63656e;
    .name {
      color: #313238;
    }
    > p {
      margin: 0 0 16px;
    }
  }
</style>
