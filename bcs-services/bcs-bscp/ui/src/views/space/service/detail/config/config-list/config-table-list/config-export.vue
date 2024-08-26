<template>
  <bk-button
    v-if="isFileType"
    :key="appData.id"
    style="margin-left: 8px"
    :disabled="allExistConfigCount === 0 || conflictFileCount > 0"
    @click="handleExportFile">
    {{ $t('打包下载') }}
  </bk-button>
  <bk-popover
    v-else
    ref="buttonRef"
    :arrow="false"
    placement="bottom-start"
    theme="light export-config-button-popover"
    trigger="click"
    @after-show="isPopoverOpen = true"
    @after-hidden="isPopoverOpen = false">
    <bk-button :disabled="allExistConfigCount === 0">{{ $t('导出至') }}</bk-button>
    <template #content>
      <div class="export-config-operations">
        <div v-for="item in exportItem" :key="item.value" class="operation-item" @click="handleExportKv(item.value)">
          <span :class="['bk-bscp-icon', `icon-${item.value}`]" />
          <span class="text"> {{ item.text }}</span>
        </div>
      </div>
    </template>
  </bk-popover>
</template>

<script lang="ts" setup>
  import { computed, ref } from 'vue';
  import { getExportKvFile } from '../../../../../../../api/config';
  import { storeToRefs } from 'pinia';
  import useServiceStore from '../../../../../../../store/service';
  import useConfigStore from '../../../../../../../store/config';
  const serviceStore = useServiceStore();
  const configStore = useConfigStore();
  const { appData, isFileType } = storeToRefs(serviceStore);
  const { allExistConfigCount, conflictFileCount } = storeToRefs(configStore);

  const props = defineProps<{
    bkBizId: string;
    appId: number;
    versionId: number;
    versionName: string;
  }>();

  const buttonRef = ref();
  const isPopoverOpen = ref(false);

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

  const handleExportKv = async (type: string) => {
    const res = await getExportKvFile(props.bkBizId, props.appId, props.versionId, type);
    let content: any;
    let mimeType: string;
    let extension: string;
    const prefix = props.versionId ? `${appData.value.spec.name}_${props.versionName}` : `${appData.value.spec.name}`;
    if (type === 'json') {
      content = JSON.stringify(res, null, 2);
      mimeType = 'application/json';
      extension = 'json';
    } else {
      content = res;
      mimeType = 'text/yaml';
      extension = 'yaml';
    }
    buttonRef.value.hide();
    downloadFile(content, mimeType, `${prefix}.${extension}`);
  };

  const downloadFile = (content: any, mimeType: string, fileName: string) => {
    const blob = new Blob([content], { type: mimeType });
    const downloadUrl = URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.href = downloadUrl;
    link.download = fileName;
    link.click();
    URL.revokeObjectURL(downloadUrl);
  };

  // 导出文件型服务配置
  const handleExportFile = async () => {
    const link = document.createElement('a');
    link.href = `${(window as any).BK_BCS_BSCP_API}/api/v1/config/biz/${props.bkBizId}/apps/${props.appId}/releases/${
      props.versionId
    }/config_item/export`;
    link.click();
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
