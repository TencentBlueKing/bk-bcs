<template>
  <bcs-dialog
    :title="$t('templateFile.title.modifyMetadata')"
    header-position="left"
    :value="value"
    width="640"
    @cancel="cancel">
    <!-- 文件元信息-->
    <bk-form
      :key="formKey"
      :model="formData"
      :rules="rules"
      :label-width="300"
      form-type="vertical"
      ref="metaDataFormRef">
      <bk-form-item error-display-type="normal" property="name" :label="$t('templateFile.label.templateName')" required>
        <bcs-input v-model.trim="formData.name" maxlength="64" clearable />
      </bk-form-item>
      <bk-form-item :label="$t('templateFile.label.desc')">
        <bcs-input maxlength="256" v-model="formData.description" clearable />
      </bk-form-item>
    </bk-form>
    <template #footer>
      <bcs-button theme="primary" @click="confirm">{{ $t('generic.button.save') }}</bcs-button>
      <bcs-button @click="cancel">{{ $t('generic.button.cancel') }}</bcs-button>
    </template>
  </bcs-dialog>
</template>
<script setup lang="ts">
import { cloneDeep } from 'lodash';
import { ref, watch } from 'vue';

import { filterPlainText } from '@blueking/xss-filter';

import { TEMPLATE_FOLDER_REGEX } from '@/common/constant';
import $i18n from '@/i18n/i18n-setup';

type FormValue = Pick<ClusterResource.CreateTemplateMetadataReq, 'name'|'description'>;

interface Props {
  value?: Boolean // 是否显示详情
  data?: FormValue// 表单数据
}

type Emits = (e: 'confirm'|'cancel', v: FormValue) => void;

const props = withDefaults(defineProps<Props>(), {
  value: () => false,
  data: () => ({
    name: '',
    description: '',
  }),
});

const emits = defineEmits<Emits>();
// 表单数据
const formKey = ref(0);
const metaDataFormRef = ref();
const formData = ref<FormValue>(cloneDeep(props.data));
const regex = new RegExp(TEMPLATE_FOLDER_REGEX);
const rules = ref({
  name: [
    {
      message: $i18n.t('generic.validate.required'),
      trigger: 'blur',
      required: true,
    },
    // 验证路径格式：不能以 / 开头或结尾，不能包含 //，名称段不能为空
    {
      message: $i18n.t('templateFile.tips.invalidPath'),
      trigger: 'blur',
      validator: (value: string) => {
        if (!value) return true;
        return !value.startsWith('/') && !value.endsWith('/') && !value.includes('//') && !value.split('/').some(seg => !seg);
      },
    },
    // 验证文件夹名称：使用 TEMPLATE_FOLDER_REGEX
    {
      message: $i18n.t('templateFile.tips.invalidFolderName'),
      trigger: 'blur',
      validator: (value: string) => {
        if (!value) return true;
        const segments = value.split('/');
        // 验证所有文件夹部分（除了最后一段文件名）
        for (let i = 0; i < segments.length - 1; i++) {
          if (!regex.test(segments[i])) {
            return false;
          }
        }
        return true;
      },
    },
    // 验证文件名称：不能包含特殊字符 \ : * ? " < > |
    {
      message: $i18n.t('templateFile.tips.invalidFileName'),
      trigger: 'blur',
      validator: (value: string) => {
        if (!value) return true;
        const segments = value.split('/');
        const fileName = segments[segments.length - 1];
        // 文件名不能包含 \/ : * ? " < > | 等特殊字符
        return regex.test(fileName);
      },
    },
  ],
});
/**
 * 自动去除名称中每一段结尾的 . 和空格
 */
function handleResetName() {
  if (!formData.value.name) return;
  formData.value.name = formData.value.name
    .split('/')
    .map(seg => seg.trim().replace(/\.+$/, ''))
    .filter(seg => seg !== '')
    .join('/');
}

// 取消修改
function cancel() {
  emits('cancel', formData.value);
}
// 确定修改元信息
async function confirm() {
  const result = await metaDataFormRef.value?.validate().catch(() => false);
  if (!result) return;

  handleResetName();
  const data = cloneDeep(formData.value);
  const xssDesc = filterPlainText(data.description);
  if (data.description !== xssDesc) {
    console.warn('Intercepted by XSS');
  }
  data.description = xssDesc;
  emits('confirm', data);
};

watch(() => props.value, () => {
  if (!props.value) {
    // 重置校验状态（hack 组件库限制，为提供reset方法）
    formKey.value = new Date().getTime();
    return;
  };
  formData.value = cloneDeep(props.data);
}, { immediate: true });
</script>
