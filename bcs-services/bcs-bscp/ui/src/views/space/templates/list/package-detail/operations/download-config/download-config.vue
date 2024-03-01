<template>
  <div class="download-config" @click="handleDownload">
    <slot>
      <bk-button text :theme="props.theme" :loading="pending">{{ props.text || t('下载') }}</bk-button>
    </slot>
  </div>
</template>
<script lang="ts" setup>
  import { ref } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { getTemplateVersionsNameByIds, downloadTemplateContent } from '../../../../../../../api/template';
  import { fileDownload } from '../../../../../../../utils/file';

  const { t } = useI18n();

  const props = withDefaults(
    defineProps<{
      spaceId: string;
      templateSpaceId: number;
      templateId: number;
      theme: string;
      text?: string;
    }>(),
    {
      theme: 'primary',
    },
  );

  const pending = ref(false);

  const handleDownload = async () => {
    if (pending.value) return;
    try {
      pending.value = true;
      const res = await getTemplateVersionsNameByIds(props.spaceId, [props.templateId]);
      const { template_name, latest_signature } = res.details[0];
      const content = await downloadTemplateContent(props.spaceId, props.templateSpaceId, latest_signature);
      fileDownload(content, template_name);
    } catch (e) {
      console.error(e);
    } finally {
      pending.value = false;
    }
  };
</script>
