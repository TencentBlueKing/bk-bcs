<template>
  <span class="read-file-content">
    <Upload class="upload-icon" @click="handleSelectFile" />
    <input ref="fileInput" class="file-input" type="file" @change="handleReadContent" />
  </span>
</template>
<script setup lang="ts">
  import { ref } from 'vue';
  import { Upload } from 'bkui-vue/lib/icon';

  const emits = defineEmits(['completed']);

  const fileInput = ref();

  const handleSelectFile = () => {
    fileInput.value.click();
  };

  const handleReadContent = (event: Event) => {
    const input = event.target as HTMLInputElement;
    const reader = new FileReader();
    reader.readAsText((input.files as FileList)[0]);
    input.value = '';
    reader.onload = () => {
      emits('completed', reader.result);
    };
  };
</script>
<style lang="scss" scoped>
  .read-file-content {
    display: inline-flex;
    align-items: center;
    margin-right: 10px;
  }
  .upload-icon {
    font-size: 14px;
  }
  .file-input {
    position: absolute;
    width: 0;
    height: 0;
    visibility: hidden;
  }
</style>
