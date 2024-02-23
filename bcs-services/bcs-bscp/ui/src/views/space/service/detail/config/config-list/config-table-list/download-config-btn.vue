<template>
  <bk-button text theme="primary" :loading="pending" :disabled="props.disabled" @click="handleDownloadConfig">
    {{ $t('下载') }}
  </bk-button>
</template>
<script lang="ts" setup>
  import { storeToRefs } from 'pinia';
  import { ref } from 'vue';
  import {
    getConfigItemDetail,
    getReleasedConfigItemDetail,
    downloadConfigContent,
  } from '../../../../../../../api/config';
  import {
    getTemplateVersionsDetailByIds,
    getTemplateVersionDetail,
    downloadTemplateContent,
  } from '../../../../../../../api/template';
  import useConfigStore from '../../../../../../../store/config';
  import { fileDownload } from '../../../../../../../utils/file';

  const { versionData } = storeToRefs(useConfigStore());

  const props = defineProps<{
    bkBizId: string;
    appId: number;
    id: number;
    type: string; // 取值为config/template，分别表示非模板套餐下配置文件和模板套餐下配置文件
    disabled: boolean;
  }>();

  const pending = ref(false);

  const handleDownloadConfig = async () => {
    let signature;
    let content;
    let fileName;
    let fileType;
    pending.value = true;
    if (props.type === 'config') {
      let res;
      if (versionData.value.id) {
        res = await getReleasedConfigItemDetail(props.bkBizId, props.appId, versionData.value.id, props.id);
        signature = res.config_item.commit_spec.content.origin_signature;
      } else {
        res = await getConfigItemDetail(props.bkBizId, props.id, props.appId);
        signature = res.content.signature;
      }
      fileName = res.config_item.spec.name;
      fileType = res.config_item.spec.file_type;
      content = await downloadConfigContent(props.bkBizId, props.appId, signature);
    } else {
      let templateSpaceId;
      if (versionData.value.id) {
        const res = await getTemplateVersionDetail(props.bkBizId, props.appId, versionData.value.id, props.id);
        signature = res.detail.origin_signature;
        fileName = res.detail.name;
        fileType = res.detail.file_type;
        templateSpaceId = res.detail.template_space_id;
      } else {
        const res = await getTemplateVersionsDetailByIds(props.bkBizId, [props.id]);
        signature = res.details[0].spec.content_spec.signature;
        fileName = res.details[0].spec.name;
        fileType = res.details[0].spec.file_type;
        templateSpaceId = res.details[0].attachment.template_space_id;
      }
      content = await downloadTemplateContent(props.bkBizId, templateSpaceId, signature);
    }
    if (fileType === 'binary') {
      fileName += '.bin';
    }
    fileDownload(content, fileName);

    pending.value = false;
  };
</script>
