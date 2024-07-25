<template>
  <bk-form form-type="vertical" ref="formRef" :model="localData" :rules="rules">
    <bk-form-item :label="t('form_服务名称')" property="name" required>
      <bk-input
        v-model="localData.name"
        :placeholder="t('请输入2-32字符，只允许英文、数字、下划线、中划线且必须以英文、数字开头和结尾')"
        :disabled="editable"
        @input="handleChange"
        v-bk-tooltips="{
          content: t('请输入2-32字符，只允许英文、数字、下划线、中划线且必须以英文、数字开头和结尾'),
          disabled: locale === 'zh-cn',
        }" />
    </bk-form-item>
    <bk-form-item :label="t('form_服务别名')" property="alias" required>
      <bk-input
        v-model="localData.alias"
        :placeholder="t('请输入2-128字符，只允许中文、英文、数字、下划线、中划线且必须以中文、英文、数字开头和结尾')"
        @input="handleChange"
        v-bk-tooltips="{
          content: t('请输入2-128字符，只允许中文、英文、数字、下划线、中划线且必须以中文、英文、数字开头和结尾'),
          disabled: locale === 'zh-cn',
        }" />
    </bk-form-item>
    <bk-form-item :label="t('服务描述')" property="memo">
      <bk-input
        v-model="localData.memo"
        :placeholder="t('服务描述限制200字符')"
        type="textarea"
        :autosize="true"
        :resize="false"
        :maxlength="200"
        @input="handleChange" />
    </bk-form-item>
    <bk-form-item :label="t('数据格式')" :description="t('tips.config')">
      <bk-radio-group v-model="localData.config_type" :disabled="editable" @change="handleConfigTypeChange">
        <bk-radio label="file">{{ t('文件型') }}</bk-radio>
        <bk-radio label="kv">{{ t('键值型') }}</bk-radio>
      </bk-radio-group>
    </bk-form-item>
    <!-- @todo 补充编辑场景下，类型切换时的校验逻辑 -->
    <bk-form-item
      v-if="localData.config_type === 'kv'"
      :label="t('数据类型')"
      property="kv_type"
      :description="t('tips.type')">
      <bk-radio-group
        v-model="localData.data_type"
        :class="{ 'en-type-group': locale === 'en' }"
        @change="handleChange">
        <bk-radio label="any">{{ t('任意类型') }}</bk-radio>
        <bk-radio v-for="kvType in CONFIG_KV_TYPE" :key="kvType.id" :label="kvType.id">{{ kvType.name }}</bk-radio>
      </bk-radio-group>
    </bk-form-item>
  </bk-form>
</template>
<script setup lang="ts">
  import { ref, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { IServiceEditForm } from '../../../../../../types/service';
  import { CONFIG_KV_TYPE } from '../../../../../constants/config';

  const { t, locale } = useI18n();

  const emits = defineEmits(['change']);

  const props = defineProps<{
    formData: IServiceEditForm;
    editable?: boolean;
  }>();

  const rules = {
    name: [
      {
        validator: (value: string) => value.length >= 2,
        message: t('最小长度2个字符'),
      },
      {
        validator: (value: string) => value.length <= 32,
        message: t('最大长度32个字符'),
      },
      {
        validator: (value: string) => /^[a-zA-Z0-9](?:[a-zA-Z0-9_-]*[a-zA-Z0-9])?$/.test(value),
        message: t('服务名称由英文、数字、下划线、中划线组成且以英文、数字开头和结尾'),
      },
    ],
    alias: [
      {
        validator: (value: string) => value.length >= 2,
        message: t('最小长度2个字符'),
      },
      {
        validator: (value: string) => value.length <= 128,
        message: t('最大长度128个字符'),
      },
      {
        validator: (value: string) =>
          /^[a-zA-Z0-9\u4e00-\u9fa5][a-zA-Z0-9_\-\u4e00-\u9fa5]*[a-zA-Z0-9\u4e00-\u9fa5]$/.test(value),
        message: t('服务别名由中文、英文、数字、下划线、中划线且必须以中文、英文、数字开头和结尾'),
      },
    ],
  };

  const localData = ref({ ...props.formData });
  const formRef = ref();

  watch(
    () => props.formData,
    (val) => {
      localData.value = { ...val };
    },
  );

  const handleConfigTypeChange = (val: string) => {
    localData.value.data_type = val === 'file' ? '' : 'any';
    handleChange();
  };

  const handleChange = () => {
    emits('change', localData.value);
  };

  const validate = () => formRef.value.validate();

  defineExpose({
    validate,
  });
</script>

<style lang="scss" scoped>
  .en-type-group {
    :deep(.bk-radio ~ .bk-radio) {
      margin-left: 20px;
    }
  }
</style>
