<script setup lang="ts">
  import { ref, defineProps, onMounted, onBeforeUnmount } from 'vue'
  import * as monaco from 'monaco-editor'

  defineProps({
    height: {
      type: Number,
      default: 400
    }
  })

  const codeEditorRef = ref()
  let editor: monaco.editor.IStandaloneCodeEditor
  const val = ref('')


  onMounted(() => {
    if (!editor) {
        editor = monaco.editor.create(codeEditorRef.value as HTMLElement, {
          value: val.value,
          theme: 'vs-dark',
          automaticLayout: true
        })
      }
      editor.onDidChangeModelContent((val:any) => {
          val.value = editor.getValue();
      })
  })

  onBeforeUnmount(() => {
    editor.dispose()
  })
</script>
<template>
  <section class="code-editor-wrapper" :style="`height: ${height}px`" ref="codeEditorRef"></section>
</template>
<style lang="scss" scoped>
  .code-editor-wrapper {
    .monaco-editor {
      width: 100%;
    }
  }
</style>