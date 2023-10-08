<template>
  <bk-sideslider
    title="导入配置项"
    :width="960"
    :is-show="isShow"
    :before-close="handleBeforeClose"
    @closed="close">
    <div class="slider-content-container">
      <bk-form form-type="vertical">
        <bk-form-item label="上传配置包" required property="package">
          <bk-upload
            class="config-uploader"
            url=""
            theme="button"
            tip="支持扩展名：.zip  .tar  .gz"
            :size="100"
            :multiple="false"
            :files="fileList"
            :custom-request="handleFileUpload">
          </bk-upload>
        </bk-form-item>
      </bk-form>
    </div>
    <div class="action-btns">
      <bk-button theme="primary" :loading="pending" :disabled="fileList.length === 0">去导入</bk-button>
      <bk-button @click="close">取消</bk-button>
    </div>
  </bk-sideslider>
</template>
<script lang="ts" setup>
import { ref, watch } from 'vue';
import useModalCloseConfirmation from '../../../../../../../../utils/hooks/use-modal-close-confirmation';

const props = defineProps<{
    show: boolean;
  }>();

const emits = defineEmits(['update:show']);

const isShow = ref(false);
const fileList = ref<File[]>([]);
const isFormChange = ref(false);
const pending = ref(false);

watch(() => props.show, (val) => {
  isShow.value = val;
  isFormChange.value = false;
});

const handleFileUpload = () => {};

const handleBeforeClose = async () => {
  if (isFormChange.value) {
    const result = await useModalCloseConfirmation();
    return result;
  }
  return true;
};

const close = () => {
  emits('update:show', false);
};

</script>
<style lang="scss" scoped>
  .slider-content-container {
    padding: 20px 40px;
    height: calc(100vh - 101px);
  }
  .action-btns {
    border-top: 1px solid #dcdee5;
    padding: 8px 24px;
    .bk-button {
      margin-right: 8px;
      min-width: 88px;
    }
  }
</style>
