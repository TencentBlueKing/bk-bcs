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

  const props = defineProps<{
    current: string,
    base: string
  }>()

  const textDiffRef = ref()
  let diffEditor: monaco.editor.IStandaloneDiffEditor

  watch(() => [props.base, props.current], () => {
    createDiffEditor()
  })

  onMounted(() => {
    createDiffEditor()
  })

  const createDiffEditor = () => {
    if (diffEditor) {
      diffEditor.dispose()
    }
    const originalModel = monaco.editor.createModel(props.base)
    const modifiedModel = monaco.editor.createModel(props.current)
    diffEditor = monaco.editor.createDiffEditor(textDiffRef.value as HTMLElement, { 
      theme: 'vs-dark',
      automaticLayout: true
    })
    diffEditor.setModel({
      original: originalModel,
      modified: modifiedModel
    })
  }

  onBeforeUnmount(() => {
    diffEditor.dispose()
  })

</script>
<template>
  <section ref="textDiffRef" class="text-diff-wrapper"></section>
</template>
<style lang="scss" scoped>
  .text-diff-wrapper {
    height: 100%;
  }
  .bk-code-diff {
    height: 100%;
  }
  :deep(.d2h-file-wrapper) {
    border: none;
  }
</style>