<template>
  <div class="example-wrap">
    <form-option @option-data="getOptionData" ref="fileOptionRef" />
    <div class="preview-container">
      <span class="preview-label">示例预览</span>
      <bk-button @click="copyExample" theme="primary" class="copy-btn">复制示例</bk-button>
      <code-preview class="preview-component" :code-val="replaceVal" />
    </div>
  </div>
</template>

<script lang="ts" setup>
  import { computed, ref, inject, Ref, provide } from 'vue';
  import { copyToClipBoard } from '../../../../../../utils/index';
  import BkMessage from 'bkui-vue/lib/message';
  import FormOption from '../form-option.vue';
  import CodePreview from '../code-preview.vue';
  import { useI18n } from 'vue-i18n';
  import { useRoute } from 'vue-router';
  // import yamlString from '/src/assets/exampleData/example.yaml?raw';
  import yamlString from '/src/assets/exampleData/file-container.yaml?raw';
  const { t } = useI18n();
  const route = useRoute();
  const fileOptionRef = ref();
  const serviceName = inject<Ref<string>>('serviceName');
  // fileOption组件传递过来的数据汇总
  const optionData = ref({
    clientKey: '',
    privacyCredential: '',
    labelArr: [],
    tempContents: '',
  });
  const bkBizId = ref(String(route.params.spaceId));
  const codeVal = yamlString;
  const formError = ref<number>();
  provide('formError', formError);
  const getOptionData = (data: any) => {
    optionData.value = data;
  };
  // 修改后的预览数据
  const replaceVal = computed(() => {
    const labelArr = optionData.value.labelArr.length ? JSON.stringify(optionData.value.labelArr.join(', ')) : [];
    let updateString = codeVal.replace('动态替换bkBizId', bkBizId.value);
    updateString = updateString.replace('动态替换serviceName', serviceName!.value);
    updateString = updateString.replaceAll('动态替换labels', labelArr);
    updateString = updateString.replaceAll('动态替换clientKey', optionData.value.privacyCredential);
    updateString = updateString.replaceAll('动态替换目录路径', optionData.value.tempContents);
    return updateString;
  });
  // 复制示例
  const copyExample = async () => {
    try {
      await fileOptionRef.value.formRef.validate();
      // 复制示例使用未脱敏的密钥
      const reg = /'(.{1}|.{3})\*{3}(.{1}|.{3})'/g;
      const copyVal = replaceVal.value.replaceAll(reg, `'${optionData.value.clientKey}'`);
      copyToClipBoard(copyVal);
      BkMessage({
        theme: 'success',
        message: t('示例已复制'),
      });
    } catch (error) {
      // 通知密钥选择组件校验状态
      formError.value = new Date().getTime();
      console.log(error);
    }
  };
  // watch(
  //   () => props.optionData,
  //   (newV) => {
  //     console.log(newV, '123123');
  //   },
  //   { deep: true },
  // );
</script>

<style scoped lang="scss">
  .example-wrap {
    display: flex;
    flex-direction: column;
    height: 100%;
    .preview-component {
      margin-top: 16px;
      padding: 16px 0 0;
      height: calc(100% - 48px);
      background-color: #f5f7fa;
    }
  }
  .preview-container {
    margin-top: 32px;
    padding: 8px 0;
    flex: 1;
    border-top: 1px solid #dcdee5;
    overflow: hidden;
  }

  .preview-label {
    font-weight: 700;
    font-size: 14px;
    letter-spacing: 0;
    line-height: 22px;
    color: #63656e;
  }
  .copy-btn {
    margin-left: 16px;
  }
</style>
