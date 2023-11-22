<template>
    <bk-form form-type="vertical" ref="formRef" :model="localData" :rules="rules">
      <bk-form-item :label="t('服务名称')" property="name" required>
        <bk-input
          placeholder="请输入2~32字符，只允许英文、数字、下划线、中划线且必须以英文、数字开头和结尾"
          v-model="localData.name"
          @change="handleChange"/>
      </bk-form-item>
      <bk-form-item :label="t('服务别名')" required>
        <bk-input
          placeholder="@todo 需要确认校验规则以及字段key"
          @change="handleChange"/>
      </bk-form-item>
      <bk-form-item :label="t('服务描述')" property="memo">
        <bk-input
          v-model="localData.memo"
          placeholder="请输入"
          type="textarea"
          :autosize="true"
          :resize="false"
          @change="handleChange"/>
      </bk-form-item>
      <bk-form-item :label="t('数据格式')" description="@todo 表单说明需要产品提供">
        <bk-radio-group v-model="localData.config_type" @change="handleConfigTypeChange">
          <bk-radio label="file">{{ t('文件型') }}</bk-radio>
          <bk-radio :label="''">{{ t('键值型') }}</bk-radio>
        </bk-radio-group>
      </bk-form-item>
      <!-- @todo 补充编辑场景下，类型切换时的校验逻辑 -->
      <bk-form-item :label="t('数据类型')" property="kv_type" description="@todo 表单key需要后台确定，表单说明需要产品提供">
        <bk-radio-group @change="handleChange">
          <bk-radio :label="''">{{ t('任意类型') }}</bk-radio>
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

  const emits = defineEmits(['change'])

  const props = defineProps<{
    formData: IServiceEditForm
  }>()

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

  const localData = ref({ ...props.formData })
  const formRef = ref()

  watch(() => props.formData, val => {
    localData.value = { ...val };
  });

  const handleConfigTypeChange = (val: string) => {
    localData.value.kv_type = val === 'file' ? '' : '@todo' // @todo 待和后台确认kv类型的默认值，切换为键值类型时，清空kv类型
    handleChange()
  }

  const handleChange = () => {
    emits('change', localData.value)
  }

  const validate = () => {
    return formRef.value.validate()
  }

  defineExpose({
    validate
  })
</script>
