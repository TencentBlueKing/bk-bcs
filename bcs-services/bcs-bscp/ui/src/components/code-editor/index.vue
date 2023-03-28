<script setup lang="ts">
  import { ref, withDefaults, onMounted, onBeforeUnmount } from 'vue'
  import * as monaco from 'monaco-editor'

  const props = withDefaults(defineProps<{
    height?: number,
    modelValue: string,
    editable?: boolean
  }>(), {
    height: 400,
    editable: true
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
          automaticLayout: true,
          readOnly: !props.editable
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