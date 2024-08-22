<template>
  <section class="code-editor-wrapper" ref="codeEditorRef"></section>
</template>
<script setup lang="ts">
  import { ref, watch, onMounted, computed } from 'vue';
  import * as monaco from 'monaco-editor';
  import editorWorker from 'monaco-editor/esm/vs/editor/editor.worker.js?worker';

  let editor: monaco.editor.IStandaloneCodeEditor;

  const props = defineProps<{
    modelValue: string;
    isCipher: boolean;
    editable: boolean;
  }>();

  const emits = defineEmits(['update:modelValue']);

  const codeEditorRef = ref();
  const localVal = ref(props.modelValue);
  const isUpdating = ref(false);

  const cipherText = computed(() => (localVal.value ? '************' : ''));

  self.MonacoEnvironment = {
    getWorker() {
      return new editorWorker();
    },
  };

  watch(
    () => props.modelValue,
    (val) => {
      localVal.value = val;
      if (val !== localVal.value && !props.isCipher) {
        editor.setValue(val);
      } else {
        editor.setValue(cipherText.value);
      }
    },
  );

  watch(
    () => props.isCipher,
    (val) => {
      if (val) {
        editor.setValue(cipherText.value);
      } else {
        editor.setValue(localVal.value);
      }
    },
  );

  onMounted(() => {
    if (!editor) {
      editor = monaco.editor.create(codeEditorRef.value as HTMLElement, {
        value: props.isCipher ? cipherText.value : localVal.value,
        theme: 'vs-dark',
        automaticLayout: true,
        language: 'plaintext',
        readOnly: !props.editable,
      });
    }

    editor.onDidChangeModelContent(() => {
      if (props.isCipher) {
        //  防止事件递归触发
        if (isUpdating.value) {
          return;
        }
        const newVal = editor.getValue().replace(/\*/g, '');
        if (newVal) {
          localVal.value = newVal;
          emits('update:modelValue', newVal);
        }
        // 标记正在更新内容
        isUpdating.value = true;
        editor.setValue(cipherText.value);
        // 重置标记
        isUpdating.value = false;
      } else {
        localVal.value = editor.getValue();
        emits('update:modelValue', localVal.value);
      }
    });
  });

  const destroy = () => {
    if (editor) {
      editor.dispose();
    }
  };

  defineExpose({
    destroy,
  });
</script>
<style lang="scss" scoped>
  .code-editor-wrapper {
    height: 100%;
    :deep(.monaco-editor) {
      width: 100%;
      .template-variable-item {
        color: #1768ef;
        border: 1px solid #1768ef;
        cursor: pointer;
      }
    }
  }
  .placeholderBox {
    height: 100%;
    background-color: #1e1e1e;
    box-sizing: content-box;
    .placeholderLine {
      display: flex;
      height: 19px;
      line-height: 19px;
      .lineNumber {
        font-family: Consolas, 'Courier New', monospace;
        width: 64px;
        text-align: center;
        color: #979ba5;
        font-size: 14px;
      }
      .lineContent {
        color: #63656e;
      }
    }
  }
  .error-msg-container {
    display: flex;
    align-items: flex-start;
    padding: 8px 16px;
    background: #212121;
    border-left: 4px solid #b34747;
    max-height: 60px;
    overflow: auto;
    .error-icon {
      display: flex;
      align-items: center;
      color: #b34747;
      height: 20px;
      font-size: 12px;
    }
    .message {
      margin-left: 8px;
      color: #dcdee5;
      line-height: 20px;
      font-size: 12px;
      word-break: break-all;
    }
  }
</style>
