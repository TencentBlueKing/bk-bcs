<script setup lang="ts">
  import { ref, watch, onMounted, onBeforeUnmount } from 'vue'
  import * as monaco from 'monaco-editor'

  const props = withDefaults(defineProps<{
    modelValue: string,
    editable?: boolean
  }>(), {
    editable: true
  })

  const emit = defineEmits(['update:modelValue'])

  const codeEditorRef = ref()
  let editor: monaco.editor.IStandaloneCodeEditor
  const localVal = ref(props.modelValue)

  watch(() => props.modelValue, (val) => {
    if (val !== localVal.value) {
      editor.setValue(val)
    }
  })

  onMounted(() => {
    if (!editor) {
      editor = monaco.editor.create(codeEditorRef.value as HTMLElement, {
        value: localVal.value,
        theme: 'vs-dark',
        automaticLayout: true,
        readOnly: !props.editable
      })
    }
    editor.onDidChangeModelContent((val:any) => {
      localVal.value = editor.getValue();
      emit('update:modelValue', localVal.value)
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