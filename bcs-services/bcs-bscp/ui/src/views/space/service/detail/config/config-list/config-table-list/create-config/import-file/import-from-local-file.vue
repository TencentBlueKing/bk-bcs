<template>
  <div class="wrap">
    <div class="label">{{ $t('上传配置项') }}</div>
    <div>
      <bk-checkbox v-model="isDecompression"> {{ $t('压缩包自动解压') }} </bk-checkbox>
      <div class="tips">{{ uploadTips }}</div>
    </div>
  </div>
  <div class="upload-file-list">
    <bk-upload
      class="config-uploader"
      theme="button"
      :size="100000"
      :multiple="true"
      :custom-request="handleFileUpload">
      <template #trigger>
        <bk-button class="upload-button">
          <Upload fill="#979BA5" />
          <span class="text">{{ $t('上传文件') }}</span>
        </bk-button>
      </template>
    </bk-upload>
    <div v-if="fileList.length > 0">
      <div :class="['open-btn', { 'is-open': isOpenFileList }]" @click="isOpenFileList = !isOpenFileList">
        <angle-double-up-line class="icon" />
        {{ isOpenFileList ? $t('收起上传列表') : $t('展开上传列表') }}
      </div>
      <div v-show="isOpenFileList" class="file-list">
        <div v-for="fileItem in fileList" :key="fileItem.file.name" class="file-item">
          <div class="file-wrapper">
            <div class="status-icon-area">
              <Done v-if="fileItem.status === 'success'" class="success-icon" />
              <Error v-if="fileItem.status === 'fail'" class="error-icon" />
            </div>
            <TextFill class="file-icon" />
            <div class="file-content">
              <div class="name">{{ fileItem.file.name }}</div>
              <div v-if="fileItem.status !== 'success'" class="progress">
                <bk-progress
                  :percent="fileItem.progress"
                  :theme="fileItem.status === 'fail' ? 'danger' : 'primary'"
                  size="small" />
              </div>
            </div>
            <Del class="del-icon" @click="handleDeleteFile(fileItem.file.name)" />
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script lang="ts" setup>
  import { ref, computed, watch } from 'vue';
  import { Upload, AngleDoubleUpLine, Done, TextFill, Error, Del } from 'bkui-vue/lib/icon';
  import { importNonTemplateConfigFile } from '../../../../../../../../../api/config';
  import { useI18n } from 'vue-i18n';
  import { IConfigImportItem } from '../../../../../../../../../../types/config';

  interface IUploadFileList {
    file: File;
    status: string;
    progress: number;
  }

  const { t } = useI18n();

  const props = defineProps<{
    bkBizId: string;
    appId: number;
  }>();

  const emits = defineEmits(['change', 'delete', 'uploading']);

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
    },
    { deep: true },
  );

  const handleFileUpload = async (option: { file: File }) => {
    loading.value = true;
    try {
      if (fileList.value.find((fileItem) => fileItem.file.name === option.file.name)) {
        handleDeleteFile(option.file.name);
      }
      fileList.value?.push({
        file: option.file,
        status: 'uploading',
        progress: 0,
      });
      const res = await importNonTemplateConfigFile(
        props.bkBizId,
        props.appId,
        option.file,
        isDecompression.value,
        (progress: any) => {
          fileList.value.find((fileItem) => fileItem.file === option.file)!.progress = progress;
        },
      );
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
    } catch (e) {
      console.error(e);
      fileList.value.find((fileItem) => fileItem.file === option.file)!.status = 'fail';
    } finally {
      loading.value = false;
    }
  };

  const handleDeleteFile = (fileName: string) => {
    fileList.value = fileList.value.filter((fileItem) => fileItem.file.name !== fileName);
    emits('delete', fileName);
  };
</script>

<style scoped lang="scss">
  .tips {
    color: #979ba5;
    height: 20px;
    font-size: 12px;
    line-height: 20px;
    margin-bottom: 8px;
  }
  :deep(.config-uploader) {
    .bk-upload-list {
      display: none;
    }
  }
  .upload-file-list {
    margin-left: 94px;
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
          max-width: 360px;
          margin-right: 4px;
          color: #63656e;
          white-space: nowrap;
          text-overflow: ellipsis;
          overflow: hidden;
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
</style>
