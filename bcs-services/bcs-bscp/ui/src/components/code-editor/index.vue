<script setup lang="ts">
  import { ref, watch, onMounted } from 'vue'
  import * as monaco from 'monaco-editor'
  import editorWorker from 'monaco-editor/esm/vs/editor/editor.worker?worker'
  import jsonWorker from 'monaco-editor/esm/vs/language/json/json.worker?worker'
  import cssWorker from 'monaco-editor/esm/vs/language/css/css.worker?worker'
  import htmlWorker from 'monaco-editor/esm/vs/language/html/html.worker?worker'
  import tsWorker from 'monaco-editor/esm/vs/language/typescript/ts.worker?worker'
  import { IVariableEditParams } from '../../../types/variable';
  import useEditorVariableReplace from '../../utils/hooks/use-editor-variable-replace';

  self.MonacoEnvironment = {
    getWorker(_, label) {
      if (label === 'json') {
        return new jsonWorker()
      }
      if (label === 'css' || label === 'scss' || label === 'less') {
        return new cssWorker()
      }
      if (label === 'html' || label === 'handlebars' || label === 'razor') {
        return new htmlWorker()
      }
      if (label === 'typescript' || label === 'javascript') {
        return new tsWorker()
      }
      return new editorWorker()
    }
  }

  const props = withDefaults(defineProps<{
    modelValue: string;
    lfEol?: boolean;
    variables?: IVariableEditParams[];
    editable?: boolean;
    language?: string;
  }>(), {
    variables: () => [],
    editable: true,
    language: ''
  })

  const emit = defineEmits(['update:modelValue', 'change'])

  const codeEditorRef = ref()
  let editor: monaco.editor.IStandaloneCodeEditor
  let editorHoverProvider:  monaco.IDisposable
  const localVal = ref(props.modelValue)

  watch(() => props.modelValue, (val) => {
    if (val !== localVal.value) {
      editor.setValue(val)
    }
  })

  watch(() => props.language, val => {
    monaco.editor.setModelLanguage(<monaco.editor.ITextModel>editor.getModel(), val)
  })

  watch(() => props.editable, val => {
    editor.updateOptions({ readOnly: !val })
  })

  watch(() => props.variables, val => {
    if (Array.isArray(val) && val.length > 0) {
      editorHoverProvider = useEditorVariableReplace(editor, val)
    }
  })

  onMounted(() => {
    if (!editor) {
      editor = monaco.editor.create(codeEditorRef.value as HTMLElement, {
        value: localVal.value,
        theme: 'vs-dark',
        automaticLayout: true,
        language: props.language,
        readOnly: !props.editable
      })
    }
    if (props.lfEol) {
      const model = <monaco.editor.ITextModel>editor.getModel()
      model.setEOL(monaco.editor.EndOfLineSequence.LF)
    }
    if (Array.isArray(props.variables) && props.variables.length > 0) {
      editorHoverProvider = useEditorVariableReplace(editor, props.variables)
    }
    editor.onDidChangeModelContent((val:any) => {
      localVal.value = editor.getValue();
      emit('update:modelValue', localVal.value)
      emit('change', localVal.value)
    })
  })

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
      editor.dispose()
    }
    if (editorHoverProvider) {
      editorHoverProvider.dispose()
    }
  }

  defineExpose({
    destroy
  })
</script>
<template>
  <section class="code-editor-wrapper" ref="codeEditorRef"></section>
</template>
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
