<template>
  <section class="code-editor-wrapper" ref="codeEditorRef"></section>
</template>
<script setup lang="ts">
  import { ref, watch, onMounted, computed } from 'vue';
  import * as monaco from 'monaco-editor';

  const props = defineProps<{
    modelValue: string;
    isVisible: boolean;
  }>();

  const emit = defineEmits(['update:modelValue']);

  const codeEditorRef = ref();
  let editor: monaco.editor.IStandaloneCodeEditor;
  const localVal = ref(props.modelValue);

  watch(
    () => props.modelValue,
    (val) => {
      localVal.value = val;
      if (props.isVisible) {
        editor.setValue(val);
      } else {
        editor.setValue(maskValue.value);
      }
    },
  );

  watch(
    () => props.isVisible,
    (val) => {
      if (val) {
        editor.setValue(localVal.value);
      } else {
        editor.setValue(maskValue.value);
      }
    },
  );

  const maskValue = computed(() => {
    return localVal.value.replace(/[^\r\n]/g, '*');
  });

  onMounted(() => {
    if (!editor) {
      editor = monaco.editor.create(codeEditorRef.value as HTMLElement, {
        value: maskValue.value,
        theme: 'custom-theme',
        automaticLayout: true,
        language: 'plaintext',
        readOnly: false,
      });
    }

    editor.onDidChangeModelContent(() => {
      const value = editor.getValue();
      const { lineNumber, column } = editor.getPosition() as any;
      const valueList = value.split('\n'); // 将所有内容按换行符切割
      console.log(valueList);
      if (value.length > localVal.value.length) {
        // 添加的内容长度
        const addLength = value.length - localVal.value.length;
        // 添加内容的位置
        const addLine = lineNumber - 1;
        const addColumn = column - 1 - addLength;
        // 添加内容
        const addContent = valueList[addLine].slice(addColumn, addColumn + addLength);
        console.log('add', addLength, addLine, addColumn, addContent);
        localVal.value = localVal.value.slice(0, addLine) + addContent + localVal.value.slice(addLine);
        emit('update:modelValue', localVal.value);
      } else {
        console.log('del');
      }

      // const newOffset = editor.getOffsetForColumn(newPosition);
      // emit('update:modelValue', localVal.value);
    });
  });
  // const handleValueChange = (value: string, event: any) => {
  //   if (value.length > secretValue.value.length) {
  //     // 添加的内容长度
  //     const addLength = value.length - secretValue.value.length;
  //     // 添加索引
  //     const addIndex = event.target.selectionStart - addLength;
  //     // 添加内容
  //     const addContent = value.slice(addIndex, event.target.selectionStart);
  //     secretValue.value = secretValue.value.slice(0, addIndex) + addContent + secretValue.value.slice(addIndex);
  //   } else {
  //     // 删除的内容长度
  //     const deleteLength = secretValue.value.length - value.length;
  //     // 删除索引
  //     const deleteIndex = event.target.selectionStart;
  //     secretValue.value = secretValue.value.slice(0, deleteIndex) + secretValue.value.slice(deleteIndex + deleteLength);
  //   }
  //   validateSecretValue();
  // };

  // 打开搜索框
  const openSearch = () => {
    const findAction = editor.getAction('actions.find');
    findAction.run();
  };

  const destroy = () => {
    if (editor) {
      editor.dispose();
    }
  };

  defineExpose({
    destroy,
    openSearch,
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
