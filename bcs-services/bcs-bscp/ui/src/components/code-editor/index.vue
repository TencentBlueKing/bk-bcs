<script setup lang="ts">
  import { ref, watch, onMounted, onBeforeUnmount } from 'vue'
  import * as monaco from 'monaco-editor'
  import editorWorker from 'monaco-editor/esm/vs/editor/editor.worker?worker'
  import jsonWorker from 'monaco-editor/esm/vs/language/json/json.worker?worker'
  import cssWorker from 'monaco-editor/esm/vs/language/css/css.worker?worker'
  import htmlWorker from 'monaco-editor/esm/vs/language/html/html.worker?worker'
  import tsWorker from 'monaco-editor/esm/vs/language/typescript/ts.worker?worker'

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
    editable?: boolean;
    language?: string;
  }>(), {
    editable: true,
    language: ''
  })

  const emit = defineEmits(['update:modelValue', 'change'])

  const codeEditorRef = ref()
  let editor: monaco.editor.IStandaloneCodeEditor
  const localVal = ref(props.modelValue)

  watch(() => props.modelValue, (val) => {
    if (val !== localVal.value) {
      editor.setValue(val)
    }
  })

  watch(() => props.language, (val) => {
    monaco.editor.setModelLanguage(<monaco.editor.ITextModel>editor.getModel(), val)
  })

  watch(() => props.editable, val => {
    editor.updateOptions({ readOnly: !val })
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
    editor.onDidChangeModelContent((val:any) => {
      localVal.value = editor.getValue();
      emit('update:modelValue', localVal.value)
      emit('change', localVal.value)
    })
  })

  onBeforeUnmount(() => {
    editor.dispose()
  })
</script>
<template>
  <section class="code-editor-wrapper" ref="codeEditorRef"></section>
</template>
<style lang="scss" scoped>
  .code-editor-wrapper {
    height: 100%;
    .monaco-editor {
      width: 100%;
    }
  }
</style>