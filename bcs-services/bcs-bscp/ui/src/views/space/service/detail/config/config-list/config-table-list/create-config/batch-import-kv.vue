<template>
  <bk-dialog
    :is-show="props.show"
    :title="t('批量导入')"
    :theme="'primary'"
    width="960"
    height="720"
    ext-cls="variable-import-dialog"
    :esc-close="false"
    @closed="handleClose"
  >
    <bk-form>
      <bk-form-item :label="t('导入方式')">
        <bk-radio-group v-model="importType">
          <bk-radio label="text">{{ t('文本导入') }}</bk-radio>
          <bk-radio label="file">{{ t('文件导入') }}</bk-radio>
        </bk-radio-group>
        <div class="tips" v-if="importType === 'text'">{{ t('只支持string、number类型,其他类型请使用文件导入') }}</div>
      </bk-form-item>
      <bk-form-item :label="t('配置文件内容')" required>
        <KvContentEditor v-if="importType === 'text'" ref="editorRef" @trigger="confirmBtnPerm = $event" />
        <bk-upload
          v-else
          class="file-uploader"
          :multiple="false"
          :before-upload="handleSelectFile"
          :accept="'.json,.yaml,.yml'"
        >
          <template #tip>
            <div class="upload-tips">
              <span>{{ t('支持 JSON、YAML 等类型文件，后台会自动检测文件类型，配置项格式请参照') }}</span>
              <span class="sample">{{ t('示例文件包') }}</span>
            </div>
          </template>
        </bk-upload>
        <div v-if="selectedFile" :class="['file-wrapper', { error: !isFileUploadSuccess }]">
          <div class="file-left">
            <Done v-if="isFileUploadSuccess" class="success-icon" />
            <TextFill class="file-icon" />
            <div class="name" :title="selectedFile.name">{{ selectedFile.name }}</div>
          </div>
          <div v-if="!isFileUploadSuccess" class="file-right">
            <span class="error-msg">{{ t('解析失败，配置项格式不正确') }}</span>
            <span class="del-icon" @click="selectedFile = undefined">
              <Del />
            </span>
          </div>
        </div>
      </bk-form-item>
    </bk-form>
    <template #footer>
      <bk-button theme="primary" style="margin-right: 8px" @click="handleConfirm">{{ t('导入') }}</bk-button>
      <bk-button @click="handleClose">{{ t('取消') }}</bk-button>
    </template>
  </bk-dialog>
</template>

<script lang="ts" setup>
import { ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { Done, TextFill, Del } from 'bkui-vue/lib/icon';
import KvContentEditor from '../../../components/kv-import-editor.vue';
import { batchImportKvFile } from '../../../../../../../../api/config';
import { Message } from 'bkui-vue';

const { t } = useI18n();
const props = defineProps<{
  show: boolean;
  bkBizId: string;
  appId: number;
}>();
const emits = defineEmits(['update:show', 'confirm']);

const editorRef = ref();
const isFormChange = ref(false);
const importType = ref('text');
const confirmBtnPerm = ref(false);
const selectedFile = ref<File>();
const isFileUploadSuccess = ref(true);
watch(
  () => props.show,
  () => {
    isFormChange.value = false;
  },
);

const handleClose = () => {
  emits('update:show', false);
};
const handleConfirm = async () => {
  if (importType.value === 'file') {
    try {
      await batchImportKvFile(props.bkBizId, props.appId, selectedFile.value);
      Message({
        theme: 'success',
        message: t('文件导入成功'),
      });
    } catch (error) {
      console.error(error);
      isFileUploadSuccess.value = false;
      return;
    }
  } else {
    await editorRef.value.handleImport();
  }
  emits('update:show', false);
  emits('confirm');
};

const handleSelectFile = (file: File) => {
  selectedFile.value = file;
  isFormChange.value = true;
  return false;
};
</script>

<style scoped lang="scss">
.tips {
  font-size: 12px;
  color: #979ba5;
}
.upload-tips {
  font-size: 12px;
  color: #63656e;
  .sample {
    margin-left: 2px;
    color: #3a84ff;
    cursor: pointer;
  }
}
.file-uploader {
  :deep(.bk-upload-list__item) {
    display: none;
  }
}
.file-wrapper {
  display: flex;
  width: 492px;
  justify-content: space-between;
  color: #979ba5;
  font-size: 12px;
  .file-left {
    display: flex;
    align-items: center;

    .success-icon {
      font-size: 20px;
      color: #2dcb56;
    }
    .file-icon {
      margin: 0 6px 0 0;
    }
    .name {
      max-width: 360px;
      margin-right: 4px;
      color: #63656e;
      white-space: nowrap;
      text-overflow: ellipsis;
      overflow: hidden;
      cursor: pointer;
    }
  }
  .file-right {
    display: flex;
    align-items: center;
    .error-msg {
      color: #ff5656;
      margin-right: 10px;
    }
    .del-icon {
      font-size: 14px;
      color: #939ba5;
      cursor: pointer;
      &:hover {
        color: #3a84ff;
      }
    }
    &:hover .del-icon {
      display: block;
    }
  }
}
.error {
  position: relative;
  background-color: #fff;
  border-bottom: 2px solid #f0f1f5;
  &::before {
    position: absolute;
    display: block;
    content: '';
    width: 156px;
    height: 2px;
    bottom: -2px;
    background-color: #ff5656;
  }
}
</style>

<style lang="scss">
.variable-import-dialog {
  .bk-modal-content {
    overflow: hidden !important;
  }
}
</style>
