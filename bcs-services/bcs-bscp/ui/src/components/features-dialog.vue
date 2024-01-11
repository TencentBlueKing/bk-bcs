<template>
  <bk-dialog
  :is-show="props.isShow"
  :theme="'primary'"
  :show-mask="true"
  quick-close
  :show-footer="false"
  dialog-type="show"
  :title="t('产品功能特性')"
  @closed="emits('update:isShow', false)"
  header-align="center">
    <div v-html="props.detail" class="markdown-theme-style detail-content"></div>
  </bk-dialog>
</template>

<script lang="ts" setup>
import { useI18n } from 'vue-i18n';
const { t } = useI18n();
const props = withDefaults(
  defineProps<{
      detail: string
      isShow: boolean
    }>(),
  {},
);
const emits = defineEmits(['update:isShow']);
</script>

<style scoped lang="scss">
.detail-content {
  padding: 40px;
  padding-top: 0;
}
</style>
