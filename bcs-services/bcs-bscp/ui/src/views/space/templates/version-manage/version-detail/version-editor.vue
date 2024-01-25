<template>
  <div class="version-editor">
    <div class="header-wrapper">
      <div class="title">
        <div v-if="props.type === 'create'" class="create-version-title">{{ t('新建版本') }}</div>
        <div v-else class="version-view-title">
          {{ props.versionName }}
          <span class="cited-info">{{ t('被引用') }}：{{ boundCount }}</span>
        </div>
      </div>
    </div>
    <div :class="['template-config-content-wrapper', { 'view-mode': isViewMode }]">
      <div v-if="!isViewMode" class="config-form">
        <bk-form ref="formRef" form-type="vertical" :model="formData" :rules="rules">
          <bk-form-item :label="t('版本号')" property="revision_name">
            <bk-input v-model="formData.revision_name" :placeholder="t('请输入')"/>
          </bk-form-item>
          <bk-form-item :label="t('版本描述')" property="revision_memo">
            <bk-input
              v-model="formData.revision_memo"
              type="textarea"
              :placeholder="t('请输入')"
              :rows="4"
              :maxlength="200"
              :resize="true" />
          </bk-form-item>
          <bk-form-item :label="t('文件权限')" required>
            <PermissionInputPicker v-model="formData.privilege" />
          </bk-form-item>
          <bk-form-item :label="t('用户')" required>
            <bk-input v-model="formData.user" :placeholder="t('请输入')"/>
          </bk-form-item>
          <bk-form-item :label="t('用户组')" required>
            <bk-input v-model="formData.user_group" :placeholder="t('请输入')"/>
          </bk-form-item>
        </bk-form>
      </div>
      <div v-bkloading="{ loading: contentLoading }" class="config-content">
        <div v-if="props.data.file_type === 'binary'" class="file-uploader-wrapper">
          <bk-upload
            class="config-uploader"
            url=""
            theme="button"
            :tip="t('支持扩展名：.bin，文件大小100M以内')"
            :size="100"
            :disabled="isViewMode"
            :multiple="false"
            :files="fileList"
            :custom-request="handleFileUpload">
            <template #file="{ file }">
              <div class="file-wrapper">
                <Done class="done-icon" />
                <TextFill class="file-icon" />
                <div class="name" @click="handleDownloadFile">{{ file.name }}</div>
                ({{ file.size }})
              </div>
            </template>
          </bk-upload>
        </div>
        <CodeEditor v-else v-model="stringContent" :editable="!isViewMode" />
      </div>
    </div>
    <div class="action-btns">
      <bk-button class="submit-btn" theme="primary" @click="handleSubmitClick">{{ t('提交') }}</bk-button>
      <bk-button class="cancel-btn" @click="emits('close')">{{ t('取消') }}</bk-button>
    </div>
  </div>
  <CreateVersionConfirmDialog
    v-model:show="isConfirmDialogShow"
    :space-id="props.spaceId"
    :template-space-id="props.templateSpaceId"
    :template-id="props.templateId"
    :version-id="props.versionId"
    :pending="submitPending"
    @confirm="triggerCreate"/>
</template>
<script lang="ts" setup>
import { computed, onMounted, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import SHA256 from 'crypto-js/sha256';
import WordArray from 'crypto-js/lib-typedarrays';
import Message from 'bkui-vue/lib/message';
import { Done, TextFill } from 'bkui-vue/lib/icon';
import { ITemplateVersionEditingData } from '../../../../../../types/template';
import { IFileConfigContentSummary } from '../../../../../../types/config';
import { stringLengthInBytes } from '../../../../../utils/index';
import { transFileToObject, fileDownload } from '../../../../../utils/file';
import {
  updateTemplateContent,
  downloadTemplateContent,
  createTemplateVersion,
  getCountsByTemplateVersionIds,
} from '../../../../../api/template';
import CodeEditor from '../../../../../components/code-editor/index.vue';
import PermissionInputPicker from '../../../../../components/permission-input-picker.vue';
import CreateVersionConfirmDialog from './create-version-confirm-dialog.vue';

const { t } = useI18n();
const props = defineProps<{
  spaceId: string;
  templateSpaceId: number;
  templateId: number;
  versionId: number;
  versionName: string;
  templateName: string;
  type: string;
  data: ITemplateVersionEditingData;
}>();

const emits = defineEmits(['created', 'close']);

const rules = {
  revision_name: [
    {
      validator: (value: string) => value.length <= 128,
      message: t('最大长度128个字符'),
    },
    {
      validator: (value: string) => {
        if (value.length > 0) {
          return /^[\u4e00-\u9fa5a-zA-Z0-9][\u4e00-\u9fa5a-zA-Z0-9_-]*[\u4e00-\u9fa5a-zA-Z0-9]?$/.test(value);
        }
        return true;
      },
      message: t('仅允许使用中文、英文、数字、下划线、中划线，且必须以中文、英文、数字开头和结尾'),
    },
  ],
  revision_memo: [
    {
      validator: (value: string) => value.length <= 200,
      message: t('最大长度200个字符'),
    },
  ],
};

const formData = ref<ITemplateVersionEditingData>({
  revision_name: '',
  revision_memo: '',
  file_type: '',
  file_mode: '',
  user: '',
  user_group: '',
  privilege: '',
  sign: '',
  byte_size: 0,
});
const formRef = ref();
const stringContent = ref('');
const fileContent = ref<IFileConfigContentSummary | File>();
const isFileChanged = ref(false);
const contentLoading = ref(false);
const uploadPending = ref(false);
const boundCountLoading = ref(false);
const boundCount = ref(0);
const submitPending = ref(false);
const isConfirmDialogShow = ref(false);

const isViewMode = computed(() => props.type === 'view');

// 传入到bk-upload组件的文件对象
const fileList = computed(() => (fileContent.value ? [transFileToObject(fileContent.value as File)] : []));

watch(
  () => props.data,
  (val) => {
    formData.value = { ...val };
  },
  { immediate: true },
);

watch(
  () => props.versionId,
  (val) => {
    if (val) {
      getContent();
      getBoundCount();
    }
  },
);

onMounted(() => {
  if (props.versionId) {
    getContent();
    getBoundCount();
  }
});

const handleFileUpload = (option: { file: File }) => {
  isFileChanged.value = true;
  return new Promise((resolve) => {
    fileContent.value = option.file;
    uploadContent().then((res) => {
      resolve(res);
    });
  });
};

// 获取非文件类型配置文件内容，文件类型手动点击时再下载
const getContent = async () => {
  try {
    contentLoading.value = true;
    const { file_type, sign: signature, byte_size } = formData.value;
    if (file_type === 'binary') {
      fileContent.value = { name: props.templateName, signature, size: String(byte_size) };
    } else {
      const configContent = await downloadTemplateContent(props.spaceId, props.templateSpaceId, signature);
      stringContent.value = String(configContent);
    }
  } catch (e) {
    console.error(e);
  } finally {
    contentLoading.value = false;
  }
};

const getBoundCount = async () => {
  boundCountLoading.value = true;
  const res = await getCountsByTemplateVersionIds(props.spaceId, props.templateSpaceId, props.templateId, [
    props.versionId,
  ]);
  boundCount.value = res.details[0].bound_unnamed_app_count;
  boundCountLoading.value = false;
};

// 上传配置内容
const uploadContent = async () => {
  const signature = await getSignature();
  const data = formData.value.file_type === 'binary' ? fileContent.value : stringContent.value;
  uploadPending.value = true;
  // @ts-ignore
  return updateTemplateContent(props.spaceId, props.templateSpaceId, data, signature).then(() => {
    if (formData.value.file_type === 'binary') {
      formData.value.byte_size = Number((fileContent.value as IFileConfigContentSummary | File).size);
    } else {
      formData.value.byte_size = new Blob([stringContent.value]).size;
    }
    formData.value.sign = signature;
    uploadPending.value = false;
  });
};

// 生成文件或文本的sha256
const getSignature = async () => {
  if (props.data.file_type === 'binary') {
    if (isFileChanged.value) {
      return new Promise((resolve) => {
        const reader = new FileReader();
        // @ts-ignore
        reader.readAsArrayBuffer(fileContent.value);
        reader.onload = () => {
          const wordArray = WordArray.create(reader.result);
          resolve(SHA256(wordArray).toString());
        };
      });
    }
    return (fileContent.value as IFileConfigContentSummary).signature;
  }
  return SHA256(stringContent.value).toString();
};

// 下载已上传文件
const handleDownloadFile = async () => {
  const { signature, name } = fileContent.value as IFileConfigContentSummary;
  const res = await downloadTemplateContent(props.spaceId, props.templateSpaceId, signature);
  fileDownload(String(res), `${name}.bin`);
};

const validate = async () => {
  await formRef.value.validate();
  if (formData.value.file_type === 'binary') {
    if (fileList.value.length === 0) {
      Message({ theme: 'error', message: t('请上传文件') });
      return false;
    }
  } else if (formData.value.file_type === 'text') {
    if (stringLengthInBytes(stringContent.value) > 1024 * 1024 * 50) {
      Message({ theme: 'error', message: t('配置内容不能超过50M') });
      return false;
    }
  }
  return true;
};

const handleSubmitClick = async () => {
  const result = await validate();
  if (!result) return;
  isConfirmDialogShow.value = true;
};

const triggerCreate = async () => {
  try {
    submitPending.value = true;
    if (formData.value.file_type !== 'binary') {
      await uploadContent();
    }
    const res = await createTemplateVersion(props.spaceId, props.templateSpaceId, props.templateId, formData.value);
    isConfirmDialogShow.value = false;
    emits('created', res.id);
    Message({
      theme: 'success',
      message: t('创建版本成功'),
    });
  } catch (e) {
    console.log(e);
  } finally {
    submitPending.value = false;
  }
};
</script>
<style lang="scss" scoped>
.version-editor {
  height: 100%;
}
.header-wrapper {
  height: 40px;
  background: #242424;
  box-shadow: 0 2px 4px 0 #00000029;
}
.title {
  display: flex;
  align-items: center;
  padding: 0 24px;
  height: 100%;
  .create-version-title {
    line-height: 20px;
    font-size: 14px;
    color: #8a8f99;
  }
  .version-view-title {
    line-height: 20px;
    font-size: 14px;
    color: #c4c6cc;
    .cited-info {
      margin-left: 16px;
      padding-left: 16px;
      color: #63656e;
      border-left: 1px solid #63656e;
    }
  }
}
.template-config-content-wrapper {
  display: flex;
  align-items: flex-start;
  height: calc(100% - 86px);
  &.view-mode {
    height: calc(100% - 40px);
    .config-content {
      width: 100%;
    }
  }
  .config-form {
    padding: 24px;
    width: 260px;
    height: 100%;
    background: #2a2a2a;
    overflow: auto;
    :deep(.bk-form) {
      .bk-form-label {
        font-size: 12px;
        color: #979ba5;
      }
      .bk-input {
        border: 1px solid #63656e;
      }
      .bk-input--text {
        background: transparent;
        color: #c4c6cc;
        &::placeholder {
          color: #63656e;
        }
      }
      .bk-textarea {
        background: transparent;
        border: 1px solid #63656e;
        textarea {
          color: #c4c6cc;
          background: transparent;
          &::placeholder {
            color: #63656e;
          }
        }
      }
    }
  }
  .config-content {
    width: calc(100% - 260px);
    height: 100%;
  }
}
.permission-input-picker {
  :deep(.perm-panel-trigger) {
    background: #1e3250;
  }
}
.file-uploader-wrapper {
  padding: 24px;
  height: 100%;
}
.config-uploader {
  :deep(.bk-upload-list__item) {
    padding: 0;
    border: none;
  }
  :deep(.bk-upload-list--disabled .bk-upload-list__item) {
    pointer-events: inherit;
  }
  .file-wrapper {
    display: flex;
    align-items: center;
    color: #979ba5;
    font-size: 12px;
    .done-icon {
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
      &:hover {
        color: #3a84ff;
        text-decoration: underline;
      }
    }
  }
}
.action-btns {
  padding: 7px 24px;
  background: #2a2a2a;
  box-shadow: 0 -1px 0 0 #141414;
  .submit-btn {
    margin-right: 8px;
    min-width: 120px;
  }
  .cancel-btn {
    min-width: 88px;
    background: transparent;
    border-color: #979ba5;
    color: #979ba5;
  }
}
</style>
