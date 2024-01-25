<template>
  <bk-dialog
    :is-show="props.show"
    :title="t('批量导入')"
    :theme="'primary'"
    width="960"
    height="720"
    ext-cls="variable-import-dialog"
    :esc-close="false"
    @closed="handleClose">
    <bk-form>
      <bk-form-item :label="t('变量内容')" required>
        <VariableContentEditor ref="editorRef" @trigger="confirmBtnPerm = $event"/>
      </bk-form-item>
    </bk-form>
    <template #footer>
      <bk-button theme="primary" style="margin-right: 8px" :disabled="!confirmBtnPerm" @click="handleConfirm">
        {{ t('导入') }}
      </bk-button>
      <bk-button @click="handleClose">{{ t('取消') }}</bk-button>
    </template>
  </bk-dialog>
</template>

<script lang="ts" setup>
import { ref } from 'vue';
import { useI18n } from 'vue-i18n';
import VariableContentEditor from './variables-content-editor.vue';

const { t } = useI18n();

const props = defineProps<{
  show: boolean;
}>();
const editorRef = ref();
const confirmBtnPerm = ref(false);
const emits = defineEmits(['update:show', 'edited']);

const handleClose = () => {
  emits('update:show', false);
};
const handleConfirm = async () => {
  await editorRef.value.handleImport();
  emits('update:show', false);
  emits('edited');
};
</script>

<style scoped lang="scss"></style>

<style lang="scss">
.variable-import-dialog {
  .bk-modal-content {
    overflow: hidden !important;
  }
}
</style>
