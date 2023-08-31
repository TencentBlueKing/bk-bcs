<script lang="ts" setup>
  import { ref } from 'vue'
  import { Message } from 'bkui-vue';
  import useModalCloseConfirmation from '../../../../../../../../utils/hooks/use-modal-close-confirmation'
  import VariablesTable from './variables-table.vue';
  import { IVariableEditParams, IVariableCitedByConfigDetailItem } from '../../../../../../../../../types/variable';
  import { getUnReleasedAppVariables, getUnReleasedAppVariablesCitedDetail, updateUnReleasedAppVariables } from '../../../../../../../../api/variable'

  const props = defineProps<{
    bkBizId: string;
    appId: number;
  }>()

  const isSliderShow = ref(false)
  const loading = ref(false)
  const variableList = ref<IVariableEditParams[]>([])
  const citedList = ref<IVariableCitedByConfigDetailItem[]>([])
  const tableRef = ref()
  const isFormChange = ref(false)
  const pending = ref(false)

  const getVariableList = async() => {
    loading.value = true
    const [variableListRes, citedListRes] = await Promise.all([
      getUnReleasedAppVariables(props.bkBizId, props.appId),
      getUnReleasedAppVariablesCitedDetail(props.bkBizId, props.appId)
    ])
    variableList.value = variableListRes.details
    citedList.value = citedListRes.details
    loading.value = false
  }

  const handleOpenSlider = () => {
    isSliderShow.value = true
    getVariableList()
  }

  const handleVariablesChange = (variables: IVariableEditParams[]) => {
    isFormChange.value = true
    variableList.value = variables
  }

  const handleSubmit = async() => {
    await tableRef.value.validate()
    try {
      pending.value = true
      await updateUnReleasedAppVariables(props.bkBizId, props.appId, variableList.value)
      close()
      Message({
        theme: 'success',
        message: '设置变量成功'
      })
    } catch (e) {
      console.log(e)
    } finally {
      pending.value = false
    }
  }

  const handleBeforeClose = async () => {
    if (isFormChange.value) {
      const result = await useModalCloseConfirmation()
      return result
    }
    return true
  }

  const close = () => {
    isSliderShow.value = false
    isFormChange.value = false
  }

</script>
<template>
  <bk-button @click="handleOpenSlider">设置变量</bk-button>
  <bk-sideslider
    width="960"
    title="设置变量"
    :is-show="isSliderShow"
    :before-close="handleBeforeClose"
    @closed="close">
    <div v-bkloading="{ loading: loading }" class="variables-table-content">
      <VariablesTable
        ref="tableRef"
        :list="variableList"
        :cited-list="citedList"
        :editable="true"
        :show-cited="true"
        @change="handleVariablesChange" />
    </div>
    <section class="action-btns">
      <bk-button theme="primary" :loading="pending" @click="handleSubmit">保存</bk-button>
      <bk-button @click="close">取消</bk-button>
    </section>
  </bk-sideslider>
</template>
<style lang="scss" scoped>
  .variables-table-content {
    padding: 48px 24px;
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
