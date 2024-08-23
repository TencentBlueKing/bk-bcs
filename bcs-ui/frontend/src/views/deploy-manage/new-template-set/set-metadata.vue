<template>
  <bcs-dialog
    :title="$t('templateSet.title.modifyMetadata')"
    header-position="left"
    :value="value"
    width="640"
    @cancel="cancel">
    <bk-form
      form-type="vertical"
      :model="formData"
      :rules="rules"
      :key="formKey"
      ref="metaDataFormRef">
      <bk-form-item property="name" error-display-type="normal" :label="$t('templateSet.label.name')" required>
        <bcs-input clearable maxlength="64" v-model="formData.name" />
      </bk-form-item>
      <bk-form-item :label="$t('templateSet.label.appVersion')">
        <bcs-input v-model="formData.version" />
      </bk-form-item>
      <bk-form-item :label="$t('templateSet.label.desc')">
        <bcs-input maxlength="256" v-model="formData.description" />
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

import $i18n from '@/i18n/i18n-setup';

export type FormValue = Pick<CreateTemplateSetReq, 'name'|'description'|'version'>;

interface Props {
  value?: Boolean
  data?: FormValue
}

type Emits = (e: 'confirm'|'cancel', v: FormValue) => void;

const props = withDefaults(defineProps<Props>(), {
  value: () => false,
  data: () => ({
    name: '',
    description: '',
    version: '',
  }),
});

const emits = defineEmits<Emits>();

const formKey = ref(0);
const metaDataFormRef = ref();
const formData = ref<FormValue>(cloneDeep(props.data));
const rules = ref({
  name: [
    {
      message: $i18n.t('generic.validate.required'),
      trigger: 'blur',
      required: true,
    },
  ],
});

function cancel() {
  emits('cancel', formData.value);
}

async function confirm() {
  const result = await metaDataFormRef.value?.validate().catch(() => false);
  if (!result) return;

  emits('confirm', formData.value);
};

watch(() => props.value, () => {
  if (!props.value) {
    // 重置校验状态
    formKey.value = new Date().getTime();
    return;
  };
  formData.value = cloneDeep(props.data);
}, { immediate: true });
</script>
