<script setup lang="ts">
  import { ref, onMounted, onBeforeUnmount } from 'vue'
  import * as monaco from 'monaco-editor'

  const props = defineProps({
    height: {
      type: Number,
      default: 400
    },
    modelValue: String
  })

  const emit = defineEmits(['update:modelValue'])

  const codeEditorRef = ref()
  let editor: monaco.editor.IStandaloneCodeEditor
  const val = ref(props.modelValue)

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
          emit('update:modelValue', val.value)
      })
  })

  onBeforeUnmount(() => {
    editor.dispose()
  })
</script>
<template>
  <section class="code-editor-wrapper" :style="`height: ${props.height}px`" ref="codeEditorRef"></section>
</template>
<style lang="scss" scoped>
  .code-editor-wrapper {
    .monaco-editor {
      width: 100%;
    }
  }
</style>