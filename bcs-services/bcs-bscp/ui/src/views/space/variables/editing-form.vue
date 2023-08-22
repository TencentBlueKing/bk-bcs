<script lang="ts" setup>
  import { ref, computed, watch } from 'vue'
  import { IVariableEditParams } from '../../../../types/variable';

  const props = defineProps<{
    type: string;
    value: IVariableEditParams;
  }>()

  const emits = defineEmits(['change'])

  const localVal = ref({
    name: '',
    type: 'string',
    default_val: '',
    memo: ''
  })
  const formRef = ref()
  const rules = {
    name: [
      {
        validator: (value: string) => value.length <= 128,
        message: '最大长度128个字符'
      },
      {
        validator: (value: string) => {
          if (value.length > 0) {
            return /^[\u4e00-\u9fa5a-zA-Z0-9][\u4e00-\u9fa5a-zA-Z0-9_\-]*[\u4e00-\u9fa5a-zA-Z0-9]?$/.test(value)
          }
          return true
        },
        message: '仅允许使用中文、英文、数字、下划线、中划线，且必须以中文、英文、数字开头和结尾'
      }
    ],
    memo: [
      {
        validator: (value: string) => value.length <= 256,
        message: '最大长度256个字符'
      }
    ]
  }

  const isEditMode = computed(() => {
    return props.type === 'edit'
  })

  watch(() => props.value, val => {
    localVal.value = { ...val }
  }, { immediate: true })

  const change = () => {
    emits('change', { ...localVal.value })
  }

  const validate = () => {
    return formRef.value.validate()
  }

  defineExpose({
    validate
  })

</script>
<template>
  <bk-form ref="formRef" form-type="vertical" :model="localVal" :rules="rules">
    <bk-form-item label="变量名称" property="name" :required="!isEditMode">
      <bk-input v-model.trim="localVal.name" prefix="bk_bscp_" :disabled="isEditMode" @change="change" />
    </bk-form-item>
    <bk-form-item label="类型" property="type" :required="!isEditMode">
      <bk-select v-model="localVal.type" :clearable="false" :disabled="isEditMode" @change="change">
        <bk-option id="string" label="string"></bk-option>
        <bk-option id="number" label="number"></bk-option>
        <bk-option id="bool" label="bool"></bk-option>
      </bk-select>
    </bk-form-item>
    <bk-form-item label="默认值" property="default_val" required>
      <bk-input v-model="localVal.default_val" @change="change" />
    </bk-form-item>
    <bk-form-item label="描述" property="memo">
      <bk-input v-model="localVal.memo" type="textarea" :maxlength="100" :rows="5" @change="change" />
    </bk-form-item>
  </bk-form>
</template>
