<template>
  <bk-sideslider :title="t('被引用')" :width="640" :is-show="isShow" :before-close="handleBeforeClose" @closed="close">
    <div class="slider-content-container"></div>
    <div class="action-btns">
      <bk-button @click="close">{{ t('关闭') }}</bk-button>
    </div>
  </bk-sideslider>
</template>
<script lang="ts" setup>
  import { ref, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import useModalCloseConfirmation from '../../../../../../../utils/hooks/use-modal-close-confirmation';

  const { t } = useI18n();
  const props = defineProps<{
    show: boolean;
  }>();

  const emits = defineEmits(['update:show']);

  const isShow = ref(false);
  const isFormChange = ref(false);
  // const pending = ref(false);

  watch(
    () => props.show,
    (val) => {
      isShow.value = val;
      isFormChange.value = false;
    },
  );

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
  .slider-content-container {
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
