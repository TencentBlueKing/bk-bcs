<script lang="ts" setup>
  import { ref, watch } from 'vue'
  import { storeToRefs } from 'pinia'
  import { Message } from 'bkui-vue';
  import useModalCloseConfirmation from '../../../utils/hooks/use-modal-close-confirmation'
  import { useGlobalStore } from '../../../store/global'
  import { updateVariable} from '../../../api/variable';
  import { IVariableEditParams } from '../../../../types/variable';
  import EditingForm from './editing-form.vue';

  const { spaceId } = storeToRefs(useGlobalStore())

  const props = defineProps<{
    show: boolean;
    id: number;
    data: IVariableEditParams;
  }>()

  const emits = defineEmits(['update:show', 'edited'])

  const isShow = ref(false)
  const isFormChanged = ref(false)
  const formRef = ref()
  const pending = ref(false)
  const variableConfig = ref<IVariableEditParams>({
    name: '',
    type: '',
    default_val: '',
    memo: ''
  })

  watch(() => props.show, val => {
    isShow.value = val
    if (val) {
      isFormChanged.value = false
      variableConfig.value = { ...props.data, name: props.data.name.replace('bk_bscp_', '') }
    }
  })

  const handleFormChange = (val: IVariableEditParams) => {
    isFormChanged.value = true
    variableConfig.value = { ...val }
  }

  const handleEditSubmit = async() => {
    await formRef.value.validate()
    try {
      pending.value = true
      const { default_val, memo } = variableConfig.value
      await updateVariable(spaceId.value, props.id, { default_val, memo })
      close()
      emits('edited')
      Message({
        theme: 'success',
        message: '编辑变量成功'
      })
    } catch (e) {
      console.log(e)
    } finally {
      pending.value = false
    }
  }

  const handleBeforeClose = async() => {
    if (isFormChanged.value) {
      const result = await useModalCloseConfirmation()
      return result
    }
    return true
  }

  const close = () => {
    emits('update:show', false)
  }
</script>
<template>
  <bk-sideslider
    title="编辑变量"
    :width="640"
    :is-show="isShow"
    :before-close="handleBeforeClose"
    @closed="close">
    <div class="variable-form">
      <EditingForm ref="formRef" type="edit" :value="variableConfig" @change="handleFormChange" />
    </div>
    <div class="action-btns">
      <bk-button theme="primary" :loading="pending" @click="handleEditSubmit">创建</bk-button>
      <bk-button @click="close">取消</bk-button>
    </div>
  </bk-sideslider>
</template>
<style lang="scss" scoped>
  .variable-form {
    padding: 20px 40px;
    height: calc(100vh - 101px);
    overflow: auto;
  }
  .action-btns {
    border-top: 1px solid #dcdee5;
    padding: 8px 24px;
    .bk-button {
      margin-right: 8px;
      min-width: 88px;
    }
  }
</style>
