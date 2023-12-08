<template>
  <bk-form form-type="vertical" ref="formRef" :model="localData" :rules="rules">
    <bk-form-item :label="t('服务名称')" property="name" required>
      <bk-input
        v-model="localData.name"
        placeholder="请输入2-32字符，只允许英文、数字、下划线、中划线且必须以英文、数字开头和结尾"
        :disabled="editable"
        @change="handleChange"
      />
    </bk-form-item>
    <bk-form-item :label="t('服务别名')" property="alias" required>
      <bk-input
        v-model="localData.alias"
        placeholder="请输入2-64字符，只允许中文、英文、数字、下划线、中划线且必须以中文、英文、数字开头和结尾"
        @change="handleChange"
      />
    </bk-form-item>
    <bk-form-item :label="t('服务描述')" property="memo">
      <bk-input
        v-model="localData.memo"
        placeholder="服务描述限制200字符"
        type="textarea"
        :autosize="true"
        :resize="false"
        :maxlength="200"
        @change="handleChange"
      />
    </bk-form-item>
    <bk-form-item :label="t('数据格式')" :description="tips.config">
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
      :description="tips.type"
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
      validator: (value: string) => value.length <= 32,
      message: '最大长度32个字符',
    },
    {
      validator: (value: string) => /^[a-zA-Z0-9\u4e00-\u9fa5][a-zA-Z0-9_\-\u4e00-\u9fa5]*[a-zA-Z0-9\u4e00-\u9fa5]$/.test(value),
      message: '服务名称由中文、英文、数字、下划线、中划线且必须以中文、英文、数字开头和结尾',
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
      message: '服务别名由英文、数字、下划线、中划线组成且以英文、数字开头和结尾',
    },
  ],
};

const tips = {
  config: `文件型：通常以文件的形式存储,通常具有良好的可读性和可维护性
           键值型：以键值对的形式存储，其中键（key）用于位置标识一个配置项，值（value）为该配置项的具体内容，kv型配置通常存储在数据库，使用SDK或API的方式读取`,
  type: `任意类型：可以创建以下任意类型的配置。否则只能创建单一类型的配置
         string：单行字符串
         number：数值，包含整数、浮点数、会校验数据类型
         text：多行字符串文本，不校验数据结构
         json、xml、yaml：不同格式的结构化数据，会校验数据结构`,
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
