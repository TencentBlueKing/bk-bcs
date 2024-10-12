<template>
  <bk-form ref="formRef" form-type="vertical" :model="localVal" :rules="rules">
    <bk-form-item :label="t('变量名称')" property="name" :required="!isEditMode">
      <bk-input v-model.trim="localVal.name" :placeholder="t('请输入')" :disabled="isEditMode" @input="change">
        <template #prefix>
          <bk-select
            v-model="localPrefix"
            class="prefix-selector"
            :clearable="false"
            :disabled="isEditMode"
            @change="change">
            <bk-option id="bk_bscp_" name="bk_bscp_"></bk-option>
            <bk-option id="BK_BSCP_" name="BK_BSCP_"></bk-option>
          </bk-select>
        </template>
      </bk-input>
    </bk-form-item>
    <bk-form-item :label="t('类型')" property="type" :required="!isEditMode">
      <bk-select v-model="localVal.type" :clearable="false" :disabled="isEditMode" @change="change">
        <bk-option id="string" label="string"></bk-option>
        <bk-option id="number" label="number"></bk-option>
      </bk-select>
    </bk-form-item>
    <bk-form-item :label="t('默认值')" property="default_val" :required="localVal.type === 'number'">
      <bk-input v-model="localVal.default_val" :placeholder="t('请输入')" @input="change" />
    </bk-form-item>
    <bk-form-item :label="t('描述')" property="memo">
      <bk-input
        v-model="localVal.memo"
        type="textarea"
        :placeholder="t('请输入')"
        :maxlength="200"
        :rows="5"
        :resize="true"
        @input="change" />
    </bk-form-item>
  </bk-form>
</template>
<script lang="ts" setup>
  import { ref, computed, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { IVariableEditParams } from '../../../../types/variable';

  const { t } = useI18n();
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
        message: t('变量名称不能为空'),
        trigger: 'blur',
      },
      {
        validator: (value: string) => value.length <= 128,
        message: t('最大长度128个字符'),
      },
      {
        validator: (value: string) => {
          if (value.length > 0) {
            return /^[a-zA-Z0-9_]+$/.test(localPrefix.value + value);
          }
          return true;
        },
        message: t('仅允许使用英文、数字、下划线'),
        trigger: 'blur',
      },
    ],
    memo: [
      {
        validator: (value: string) => value.length <= 200,
        message: t('最大长度200个字符'),
      },
    ],
    default_val: [
      {
        validator: (value: string) => {
          if (localVal.value.type === 'string') return true;
          return /^-?\d+(\.\d+)?$/.test(value);
        },
        message: t('无效默认值，类型为number值不为数字'),
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
