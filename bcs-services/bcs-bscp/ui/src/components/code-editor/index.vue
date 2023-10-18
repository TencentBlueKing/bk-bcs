<template>
  <section class="code-editor-wrapper" ref="codeEditorRef"></section>
</template>
<script setup lang="ts">
import { ref, watch, onMounted } from 'vue';
import * as monaco from 'monaco-editor';
import editorWorker from 'monaco-editor/esm/vs/editor/editor.worker.js?worker';
import jsonWorker from 'monaco-editor/esm/vs/language/json/json.worker.js?worker';
import cssWorker from 'monaco-editor/esm/vs/language/css/css.worker.js?worker';
import htmlWorker from 'monaco-editor/esm/vs/language/html/html.worker.js?worker';
import tsWorker from 'monaco-editor/esm/vs/language/typescript/ts.worker.js?worker';
import { IVariableEditParams } from '../../../types/variable';
import useEditorVariableReplace from '../../utils/hooks/use-editor-variable-replace';

interface errorLineItem {
  lineNumber: number;
  errorInfo: string;
}

self.MonacoEnvironment = {
  getWorker(_, label) {
    if (label === 'json') {
      return new jsonWorker();
    }
    if (label === 'css' || label === 'scss' || label === 'less') {
      return new cssWorker();
    }
    if (label === 'html' || label === 'handlebars' || label === 'razor') {
      return new htmlWorker();
    }
    if (label === 'typescript' || label === 'javascript') {
      return new tsWorker();
    }
    return new editorWorker();
  },
};

const props = withDefaults(
  defineProps<{
    modelValue: string;
    lfEol?: boolean;
    variables?: IVariableEditParams[];
    editable?: boolean;
    language?: string;
    errorLine?: errorLineItem[];
  }>(),
  {
    variables: () => [],
    editable: true,
    language: '',
  }
);

const emit = defineEmits(['update:modelValue', 'change', 'enter']);

const codeEditorRef = ref();
let editor: monaco.editor.IStandaloneCodeEditor;
let editorHoverProvider: monaco.IDisposable;
const localVal = ref(props.modelValue);

watch(
  () => props.modelValue,
  (val) => {
    if (val !== localVal.value) {
      editor.setValue(val);
    }
  }
);

watch(
  () => props.language,
  (val) => {
    monaco.editor.setModelLanguage(editor.getModel() as monaco.editor.ITextModel, val);
  }
);

watch(
  () => props.editable,
  (val) => {
    editor.updateOptions({ readOnly: !val });
  }
);

watch(
  () => props.variables,
  (val) => {
    if (Array.isArray(val) && val.length > 0) {
      editorHoverProvider = useEditorVariableReplace(editor, val);
    }
  }
);

watch(
  () => props.errorLine,
  () => {
    setErrorLine();
  }
);

onMounted(() => {
  if (!editor) {
    editor = monaco.editor.create(codeEditorRef.value as HTMLElement, {
      value: localVal.value,
      theme: 'vs-dark',
      automaticLayout: true,
      language: props.language,
      readOnly: !props.editable,
    });
  }
  if (props.lfEol) {
    const model = editor.getModel() as monaco.editor.ITextModel;
    model.setEOL(monaco.editor.EndOfLineSequence.LF);
  }
  if (Array.isArray(props.variables) && props.variables.length > 0) {
    editorHoverProvider = useEditorVariableReplace(editor, props.variables);
  }
  editor.onDidChangeModelContent(() => {
    localVal.value = editor.getValue();
    emit('update:modelValue', localVal.value);
    emit('change', localVal.value);
  });
  // 监听第一次回车事件
  const listener = editor.onKeyDown((event) => {
    if (event.keyCode === monaco.KeyCode.Enter) {
      emit('enter');
      // 取消监听键盘事件
      listener.dispose();
    }
  });
});

// 添加错误行
const setErrorLine = () => {
  // 创建错误标记列表
  const markers = props.errorLine!.map(({ lineNumber, errorInfo }) => ({
    startLineNumber: lineNumber,
    endLineNumber: lineNumber,
    startColumn: 1,
    endColumn: 200,
    message: errorInfo,
    severity: monaco.MarkerSeverity.Error,
  }));
  // 设置编辑器模型的标记
  monaco.editor.setModelMarkers(editor.getModel() as monaco.editor.ITextModel, 'error', markers);
};

// @bug vue3的Teleport组件销毁时，子组件的onBeforeUnmount不会被执行，会出现内存泄漏，目前尚未被修复 https://github.com/vuejs/core/issues/6347
// onBeforeUnmount(() => {
//   if (editor) {
//     editor.dispose()
//   }
//   if (editorHoverProvider) {
//     editorHoverProvider.dispose()
//   }
// })
const destroy = () => {
  if (editor) {
    editor.dispose();
  }
  if (editorHoverProvider) {
    editorHoverProvider.dispose();
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
</style>
