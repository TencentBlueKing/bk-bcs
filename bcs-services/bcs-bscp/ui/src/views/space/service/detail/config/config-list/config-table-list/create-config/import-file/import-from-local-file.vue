<template>
  <div class="wrap">
    <div class="label">{{ $t('上传配置项') }}</div>
    <div>
      <bk-checkbox v-model="isDecompression" class="decompression"> {{ $t('压缩包自动解压') }} </bk-checkbox>
      <div class="tips">{{ uploadTips }}</div>
    </div>
  </div>
  <div class="upload-file-list">
    <bk-upload
      class="config-uploader"
      theme="button"
      :size="100000"
      :multiple="true"
      :limit="10"
      :custom-request="handleFileUpload"
      @exceed="handleExceed">
      <template #trigger>
        <bk-button class="upload-button">
          <Upload class="icon" />
          <span class="text">{{ $t('上传文件') }}</span>
        </bk-button>
      </template>
    </bk-upload>

    <div v-if="fileList.length > 0">
      <div :class="['open-btn', { 'is-open': isOpenFileList }]" @click="isOpenFileList = !isOpenFileList">
        <angle-double-up-line class="icon" />
        {{ isOpenFileList ? $t('收起上传列表') : $t('展开上传列表') }}
      </div>
      <RecycleScroller v-show="isOpenFileList" class="file-list" :items="fileList" :item-size="40" v-slot="{ item }">
        <div class="file-item">
          <div class="file-wrapper">
            <div class="status-icon-area">
              <Done v-if="item.status === 'success'" class="success-icon" />
              <Error v-if="item.status === 'fail'" class="error-icon" />
            </div>
            <TextFill class="file-icon" />
            <div class="file-content">
              <div class="name">
                <div v-overflow-title>{{ item.file.name }}</div>
              </div>
              <div v-if="item.status !== 'success' && item.status !== 'fail'" class="progress">
                <bk-progress
                  :percent="item.progress"
                  :theme="item.status === 'fail' ? 'danger' : 'primary'"
                  size="small" />
              </div>
              <div v-else-if="item.status === 'fail'" class="error-message">
                <div v-overflow-title>{{ item.errorMessage }}</div>
              </div>
            </div>
            <Del class="del-icon" @click="handleDeleteFile(item.file.name)" />
          </div>
        </div>
      </RecycleScroller>
    </div>
  </div>
</template>

<script lang="ts" setup>
  import { ref, computed, watch } from 'vue';
  import { Upload, AngleDoubleUpLine, Done, TextFill, Error, Del } from 'bkui-vue/lib/icon';
  import { importNonTemplateConfigFile } from '../../../../../../../../../api/config';
  import { useI18n } from 'vue-i18n';
  import { IConfigImportItem } from '../../../../../../../../../../types/config';
  import { importTemplateFile } from '../../../../../../../../../api/template';
  import { Message } from 'bkui-vue';
  import useGlobalStore from '../../../../../../../../../store/global';
  import { storeToRefs } from 'pinia';

  interface IUploadFileList {
    file: File;
    status: string;
    progress: number;
    errorMessage?: string;
    id: string;
  }

  const { t } = useI18n();
  const { spaceFeatureFlags } = storeToRefs(useGlobalStore());

  const props = defineProps<{
    isTemplate: boolean; // 是否是配置模板导入
    bkBizId?: string;
    appId?: number;
    spaceId?: string;
    currentTemplateSpace?: number;
  }>();

  const emits = defineEmits(['change', 'delete', 'uploading', 'decompressing', 'fileProcessing']);

  const loading = ref(false);
  const isDecompression = ref(true);
  const isOpenFileList = ref(true);
  const fileList = ref<IUploadFileList[]>([]);

  const uploadTips = computed(() => {
    if (isDecompression.value) {
      return t('支持上传多个文件、目录树结构，同时也支持将扩展名为 .zip / .tar / .tar.gz / .tgz 的压缩文件解压后导入');
    }
    return t(
      '支持上传多个文件、目录树结构，其中扩展名为 .zip / .tar / .tar.gz / .tgz 的压缩文件不会解压，将整体作为二进制配置项上传',
    );
  });

  watch(
    () => fileList.value,
    () => {
      if (fileList.value.some((fileItem) => fileItem.status === 'uploading')) {
        emits('uploading', true);
      } else {
        emits('uploading', false);
      }
      if (fileList.value.some((fileItem) => fileItem.status === 'decompressing')) {
        const decompressingFile = fileList.value.find((file) => file.status === 'decompressing');
        const isCompressionFile = handleCheckIsCompressedFile(decompressingFile!.file.name);
        if (isDecompression.value && isCompressionFile) {
          emits('decompressing', true);
        } else {
          emits('fileProcessing', true);
        }
      } else {
        emits('decompressing', false);
        emits('fileProcessing', false);
      }
    },
    { deep: true },
  );

  const handleFileUpload = async (option: { file: File }) => {
    const {
      RESOURCE_LIMIT: { MaxUploadContentLength, maxFileSize },
    } = spaceFeatureFlags.value;
    const fileSize = option.file.size / 1024 / 1024;
    const isCompressionFile = handleCheckIsCompressedFile(option.file.name);

    if (isDecompression.value) {
      if (isCompressionFile && fileSize > MaxUploadContentLength) {
        Message({
          theme: 'error',
          message: t('压缩包大小不能超过{n}GB', { n: MaxUploadContentLength / 1024 }),
        });
        return;
      }
      if (!isCompressionFile && fileSize > maxFileSize) {
        Message({
          theme: 'error',
          message: t('单文件大小不能超过{n}M', { n: maxFileSize }),
        });
        return;
      }
    }

    if (!isDecompression.value && fileSize > maxFileSize) {
      Message({
        theme: 'error',
        message: t('单文件大小不能超过{n}M', { n: maxFileSize }),
      });
      return;
    }

    loading.value = true;
    try {
      if (fileList.value.find((fileItem) => fileItem.file.name === option.file.name)) {
        handleDeleteFile(option.file.name);
      }
      fileList.value = [
        ...fileList.value,
        {
          file: option.file,
          status: 'uploading',
          progress: 0,
          id: `${option.file.name}-${Date.now()}`,
        },
      ];
      let res;
      if (props.isTemplate) {
        res = await importTemplateFile(
          props.spaceId!,
          props.currentTemplateSpace!,
          option.file,
          isDecompression.value,
          (progress: number) => {
            const fileItem = fileList.value.find((fileItem) => fileItem.file === option.file);
            if (progress === 100) fileItem!.status = 'decompressing';
            fileList.value.find((fileItem) => fileItem.file === option.file)!.progress = progress;
          },
        );
      } else {
        res = await importNonTemplateConfigFile(
          props.bkBizId!,
          props.appId!,
          option.file,
          isDecompression.value,
          (progress: number) => {
            const fileItem = fileList.value.find((fileItem) => fileItem.file === option.file);
            if (progress === 100) fileItem!.status = 'decompressing';
            fileItem!.progress = progress;
          },
        );
      }
      fileList.value.find((fileItem) => fileItem.file === option.file)!.status = 'success';
      res.non_exist.forEach((item: IConfigImportItem) => {
        item.privilege = '644';
        item.user = 'root';
        item.user_group = 'root';
        item.file_name = option.file.name;
      });
      res.exist.forEach((item: IConfigImportItem) => {
        item.file_name = option.file.name;
      });
      emits('change', res.exist, res.non_exist);
    } catch (e: any) {
      console.error(e);
      const file = fileList.value.find((fileItem) => fileItem.file === option.file);
      file!.status = 'fail';
      file!.errorMessage = e.response.data.error.message;
    } finally {
      loading.value = false;
    }
  };

  const handleDeleteFile = (fileName: string) => {
    fileList.value = fileList.value.filter((fileItem) => fileItem.file.name !== fileName);
    emits('delete', fileName);
  };

  // 判断是否是压缩包
  const handleCheckIsCompressedFile = (filename: string) => {
    const ext = filename.split('.').pop()!.toLowerCase();
    return ['zip', 'rar', 'tar', 'gz', 'tgz'].includes(ext);
  };

  const handleExceed = async (files: any) => {
    let index = 0;
    const uploadNext = async () => {
      if (index < files.length) {
        index += 1;
        const file = files[index];
        await handleFileUpload({ file });
        await uploadNext();
      }
    };
    const uploadPromises = [];
    for (let i = 0; i < 5; i++) {
      uploadPromises.push(uploadNext());
    }
    await Promise.allSettled(uploadPromises);
  };
</script>

<style scoped lang="scss">
  .tips {
    color: #979ba5;
    font-size: 12px;
    line-height: 20px;
    margin-bottom: 8px;
    width: 780px;
  }
  :deep(.config-uploader) {
    .bk-upload-list {
      display: none;
    }
  }
  .upload-file-list {
    margin-left: 100px;
    padding: 12px;
    width: 520px;
    background: #f5f7fa;
    border-radius: 2px;
    .open-btn {
      display: flex;
      align-items: center;
      gap: 6px;
      font-size: 12px;
      color: #3a84ff;
      cursor: pointer;
      margin-top: 8px;
      .icon {
        transform: rotate(180deg);
        transition: all 0.3s ease-in-out;
      }
      &.is-open {
        .icon {
          transform: rotate(0deg);
        }
      }
    }
    .file-list {
      margin-top: 10px;
      max-height: 300px;
    }
    .file-wrapper {
      display: flex;
      align-items: center;
      color: #979ba5;
      font-size: 12px;
      margin-bottom: 5px;
      height: 32px;
      &:hover {
        background-color: #f0f1f5;
        .del-icon {
          display: block !important;
        }
      }
      .status-icon-area {
        display: flex;
        width: 20px;
        height: 100%;
        align-items: center;
        justify-content: center;
        margin-right: 12px;
        .success-icon {
          font-size: 20px;
          color: #2dcb56;
        }
        .error-icon {
          font-size: 14px;
          color: #ea3636;
        }
      }
      .file-icon {
        margin: 0 6px 0 0;
        font-size: 16px;
      }
      .del-icon {
        display: none !important;
        font-size: 16px;
        cursor: pointer;
        &:hover {
          color: #3a84ff;
        }
      }
      .file-content {
        position: relative;
        width: 100%;
        height: 20px;
        .name {
          max-width: 200px;
          margin-right: 4px;
          color: #63656e;
          white-space: nowrap;
          text-overflow: ellipsis;
          overflow: hidden;
        }
        .error-message {
          position: absolute;
          color: #ff5656;
          right: 10px;
          top: 0;
          max-width: 200px;
          text-overflow: ellipsis;
          white-space: nowrap;
          overflow: hidden;
          cursor: pointer;
        }
        :deep(.bk-progress) {
          position: absolute;
          bottom: -6px;
          .progress-outer {
            position: relative;
            .progress-text {
              position: absolute;
              right: 8px;
              top: -22px;
              font-size: 12px !important;
              color: #63656e !important;
            }
            .progress-bar {
              height: 2px;
            }
          }
        }
      }
    }
  }

  .upload-button {
    .icon {
      font-size: 16px;
      color: #63656e;
    }
    .text {
      font-size: 12px;
      margin-left: 7px;
    }
  }

  .decompression {
    font-size: 12px !important;
  }
</style>
