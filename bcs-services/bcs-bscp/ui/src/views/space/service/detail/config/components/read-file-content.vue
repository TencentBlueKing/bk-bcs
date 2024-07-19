<template>
  <span class="read-file-content">
    <Upload class="upload-icon" @click="handleSelectFile" />
    <input ref="fileInput" class="file-input" type="file" @change="handleReadContent" />
  </span>
</template>
<script setup lang="ts">
  import { ref } from 'vue';
  import { Upload } from 'bkui-vue/lib/icon';
  import { Message } from 'bkui-vue';
  import { useI18n } from 'vue-i18n';
  const { t } = useI18n();

  const emits = defineEmits(['completed']);

  const fileInput = ref();

  const handleSelectFile = () => {
    fileInput.value.click();
  };

  const handleReadContent = (event: Event) => {
    const input = event.target as HTMLInputElement;
    const file = (input.files as FileList)[0];
    const reader = new FileReader();
    reader.readAsText(file);
    input.value = '';
    reader.onload = () => {
      const content = reader.result;

      // 读取为 ArrayBuffer 成功
      const isTextFile = checkFileIsText(content as string);
      if (isTextFile) {
        // 是文本文件
        emits('completed', content);
      } else {
        // 不是文本文件，处理错误
        Message({
          theme: 'error',
          message: t('文件格式错误，请重新选择文本文件上传'),
        });
      }
    };
  };

  const checkFileIsText = (content: string) => {
    // eslint-disable-next-line
    return /^[\x00-\x7F]*$/.test(content.slice(0, 4096));
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
