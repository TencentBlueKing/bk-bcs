<template>
  <bk-form form-type="vertical" ref="formRef" :model="localData" :rules="rules">
    <bk-form-item :label="t('服务名称')" property="name" required>
      <bk-input
        placeholder="需以英文、数字和下划线组成，不超过128字符"
        v-model="localData.name"
        :disabled="editable"
        @change="handleChange"
      />
    </bk-form-item>
    <bk-form-item :label="t('服务别名')" property="alias" required>
      <bk-input v-model="localData.alias" placeholder="需以英文、数字和下划线组成，不超过128字符" @change="handleChange" />
    </bk-form-item>
    <bk-form-item :label="t('服务描述')" property="memo">
      <bk-input
        v-model="localData.memo"
        placeholder="请输入"
        type="textarea"
        :autosize="true"
        :resize="false"
        @change="handleChange"
      />
    </bk-form-item>
    <bk-form-item :label="t('数据格式')" description="@todo 表单说明需要产品提供">
      <bk-radio-group v-model="localData.config_type"  :disabled="editable" @change="handleConfigTypeChange">
        <bk-radio label="file">{{ t('文件型') }}</bk-radio>
        <bk-radio label="kv">{{ t('键值型') }}</bk-radio>
      </bk-radio-group>
    </bk-form-item>
    <!-- @todo 补充编辑场景下，类型切换时的校验逻辑 -->
    <bk-form-item
      v-if="localData.config_type === 'kv'"
      :label="t('数据类型')"
      property="kv_type"
      description="@todo 表单key需要后台确定，表单说明需要产品提供"
    >
      <bk-radio-group v-model="localData.data_type" @change="handleChange">
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

const { t } = useI18n();

const emits = defineEmits(['change']);

const props = defineProps<{
  formData: IServiceEditForm;
  editable?: boolean;
}>();

const rules = {
  name: [
    {
      validator: (value: string) => value.length >= 2,
      message: '最小长度2个字符',
    },
    {
      validator: (value: string) => value.length <= 128,
      message: '最大长度128个字符',
    },
    {
      validator: (value: string) => /^[a-zA-Z0-9][a-zA-Z0-9_-]*[a-zA-Z0-9]?$/.test(value),
      message: '服务名称由英文、数字、下划线、中划线组成且以英文、数字开头和结尾',
    },
  ],
  alias: [
    {
      validator: (value: string) => value.length >= 2,
      message: '最小长度2个字符',
    },
    {
      validator: (value: string) => value.length <= 128,
      message: '最大长度128个字符',
    },
    {
      validator: (value: string) => /^[a-zA-Z0-9][a-zA-Z0-9_-]*[a-zA-Z0-9]?$/.test(value),
      message: '服务名称由英文、数字、下划线、中划线组成且以英文、数字开头和结尾',
    },
  ],
  memo: [
    {
      validator: (value: string) => {
        if (!value) return true;
        return /^[\u4e00-\u9fa5a-zA-Z0-9][\u4e00-\u9fa5a-zA-Z0-9_\-()\s]*[\u4e00-\u9fa5a-zA-Z0-9]$/.test(value);
      },
      message: '无效备注，只允许包含中文、英文、数字、下划线()、连字符(-)、空格，且必须以中文、英文、数字开头和结尾',
      trigger: 'change',
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
