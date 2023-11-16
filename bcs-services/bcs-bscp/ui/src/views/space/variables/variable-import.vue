<template>
  <bk-dialog
    :is-show="show"
    :title="'批量导入'"
    :theme="'primary'"
    @closed="handleClose"
    @confirm="handleConfirm"
    confirm-text="导入"
    width="960"
    height="720"
    ext-cls="variable-import-dialog"
  >
    <bk-form>
      <bk-form-item label="变量内容" required>
        <VariableContentEditor ref="editorRef" />
      </bk-form-item>
    </bk-form>
  </bk-dialog>
</template>

<script lang="ts" setup>
import { ref, watch } from 'vue';
import VariableContentEditor from './variables-content-editor.vue';
const isShow = ref(false);
const props = defineProps<{
  show: boolean;
}>();
const editorRef = ref();
const emits = defineEmits(['update:show', 'edited']);
watch(
  () => props.show,
  (val) => {
    isShow.value = val;
  },
);
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
