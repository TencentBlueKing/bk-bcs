<template>
  <div class="wrap">
    <div class="label">{{ $t('上传配置项') }}</div>
    <bk-upload
      class="config-uploader"
      :tip="$t('支持上传多个文件、目录树结构，同时也支持将扩展名为 .zip / .tar / .tar.gz / .tgz 的压缩文件解压后导入')"
      theme="button"
      :size="100"
      :custom-request="handleFileUpload">
      <template #trigger>
        <bk-button class="upload-button">
          <Upload fill="#979BA5" />
          <span class="text">{{ $t('上传文件') }}</span>
        </bk-button>
      </template>
    </bk-upload>
    <bk-checkbox style="margin-left: 24px" v-model="isDecompression"> {{ $t('压缩包自动解压') }} </bk-checkbox>
  </div>
</template>

<script lang="ts" setup>
  import { ref } from 'vue';
  import { Upload } from 'bkui-vue/lib/icon';

  const props = defineProps<{
    bkBizId: string;
    appId: number;
  }>();

  const loading = ref(false);
  const isDecompression = ref(false);

  const handleFileUpload = async (option: { file: File }) => {
    loading.value = true;
    try {
      console.log(option, props.appId);
      // existConfigList.value = res.exist;
      // nonExistConfigList.value = res.non_exist;
      // nonExistConfigList.value.forEach((item: IConfigImportItem) => {
      //   item.privilege = '644';
      //   item.user = 'root';
      //   item.user_group = 'root';
      // });
      // if (nonExistConfigList.value.length === 0) expandNonExistTable.value = false;
      // isTableChange.value = false;
    } catch (e) {
      console.error(e);
    } finally {
      loading.value = false;
    }
  };
</script>

<style scoped lang="scss">
  .tips {
    margin-left: 100px;
  }
  :deep(.config-uploader) {
    position: relative;
    .bk-upload__tip {
      margin-left: 0;
      position: absolute;
      bottom: -28px;
      left: 0;
      width: 700px;
      height: 20px;
      font-size: 12px;
      color: #979ba5;
    }
    .bk-upload-list {
      display: none;
    }
  }
</style>
