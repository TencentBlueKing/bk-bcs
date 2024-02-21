<template>
  <bk-popover :arrow="false" placement="bottom-start" theme="light export-config-button-popover">
    <bk-button>{{ $t('导出至') }}</bk-button>
    <template #content>
      <div class="export-config-operations">
        <div v-for="item in exportItem" :key="item.value" class="operation-item">
          <span :class="['bk-bscp-icon', `icon-${item.value}`]" />
          <span class="text" @click="handleExport(item.value)"> {{ item.text }}</span>
        </div>
      </div>
    </template>
  </bk-popover>
</template>

<script lang="ts" setup>
  import { computed } from 'vue';
  import jsyaml from 'js-yaml';
  import { getExportKvFile } from '../../../../../../../api/config';
  const props = defineProps<{
    bkBizId: string;
    appId: number;
    verisionId: number;
  }>();

  const exportItem = computed(() => [
    {
      text: 'JSON',
      value: 'json',
    },
    {
      text: 'YAML',
      value: 'yaml',
    },
  ]);

  const handleExport = async (type: string) => {
    const res = await getExportKvFile(props.bkBizId, props.appId, props.verisionId, type);
    let content;
    let mimeType;
    let extension;
    if (type === 'json') {
      content = JSON.stringify(res, null, 2);
      mimeType = 'application/json';
      extension = 'json';
    } else if (type === 'yaml') {
      content = jsyaml.dump(res);
      mimeType = 'text/yaml';
      extension = 'yaml';
    }
    downloadFile(content, mimeType!, `data.${extension}`);
  };

  const downloadFile = (content: string, mimeType: string, fileName: string) => {
    const blob = new Blob([content], { type: mimeType });
    const downloadUrl = URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.href = downloadUrl;
    link.download = fileName;
    link.click();
    URL.revokeObjectURL(downloadUrl);
  };
</script>

<style lang="scss">
  .export-config-button-popover.bk-popover.bk-pop2-content {
    padding: 4px 0;
    border: 1px solid #dcdee5;
    box-shadow: 0 2px 6px 0 #0000001a;
    .export-config-operations {
      .operation-item {
        padding: 0 12px;
        min-width: 100px;
        height: 32px;
        line-height: 32px;
        color: #63656e;
        font-size: 14px;
        align-items: center;
        cursor: pointer;
        &:hover {
          background: #f5f7fa;
        }
        .text {
          margin-left: 4px;
          font-size: 12px;
        }
      }
    }
  }
</style>
