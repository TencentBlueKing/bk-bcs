<template>
  <bk-form ref="formRef" form-type="vertical" :model="localVal" :rules="rules">
    <bk-form-item label="变量名称" property="name" :required="!isEditMode">
      <bk-input v-model.trim="localVal.name" :disabled="isEditMode" @change="change">
        <template #prefix>
          <bk-select
            v-model="localPrefix"
            class="prefix-selector"
            :clearable="false"
            :disabled="isEditMode"
            @change="change"
          >
            <bk-option id="bk_bscp_" name="bk_bscp_"></bk-option>
            <bk-option id="BK_BSCP_" name="BK_BSCP_"></bk-option>
          </bk-select>
        </template>
      </bk-input>
    </bk-form-item>
    <bk-form-item label="类型" property="type" :required="!isEditMode">
      <bk-select v-model="localVal.type" :clearable="false" :disabled="isEditMode" @change="change">
        <bk-option id="string" label="string"></bk-option>
        <bk-option id="number" label="number"></bk-option>
      </bk-select>
    </bk-form-item>
    <bk-form-item label="默认值" property="default_val" required>
      <bk-input v-model="localVal.default_val" @change="change" />
    </bk-form-item>
    <bk-form-item label="描述" property="memo">
      <bk-input v-model="localVal.memo" type="textarea" :maxlength="100" :rows="5" @change="change" :resize="true" />
    </bk-form-item>
  </bk-form>
</template>
<script lang="ts" setup>
import { ref, computed, watch } from 'vue';
import { IVariableEditParams } from '../../../../types/variable';

const props = defineProps<{
  type: string;
  prefix: string;
  value: IVariableEditParams;
}>();

const emits = defineEmits(['change']);

const localVal = ref({
  name: '',
  type: 'string',
  default_val: '',
  memo: '',
});
const localPrefix = ref(props.prefix);
const formRef = ref();
const rules = {
  name: [
    {
      required: true,
      message: '变量名称不能为空',
      trigger: 'blur',
    },
    {
      validator: (value: string) => value.length <= 128,
      message: '最大长度128个字符',
    },
    {
      validator: (value: string) => {
        if (value.length > 0) {
          return /^[a-zA-Z_]\w*$/.test(value);
        }
        return true;
      },
      message: '仅允许使用中文、英文、数字、下划线，且不能以数字开头',
      trigger: 'blur',
    },
  ],
  memo: [
    {
      validator: (value: string) => value.length <= 256,
      message: '最大长度256个字符',
    },
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

const isEditMode = computed(() => props.type === 'edit');

watch(
  () => props.value,
  (val) => {
    localVal.value = { ...val };
  },
  { immediate: true },
);

watch(
  () => props.prefix,
  (val) => {
    localPrefix.value = val;
  },
);

const change = () => {
  emits('change', { ...localVal.value }, localPrefix.value);
};

const validate = () => formRef.value.validate();

defineExpose({
  validate,
});
</script>
<style lang="scss" scoped>
.prefix-selector {
  width: 100px;
  border-right: 1px solid #c4c6cc;
  :deep(.bk-input) {
    height: 28px;
    border: none;
    .bk-input--text {
      background: #f5f7fa;
    }
  }
}
</style>
