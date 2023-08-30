<script lang="ts" setup>
  import { ref } from 'vue'
  import useModalCloseConfirmation from '../../../../../../../../utils/hooks/use-modal-close-confirmation'
  import VariablesTable from './variables-table.vue';
  import { IConfigVariableItem } from '../../../../../../../../../types/variable';
  import { getUnReleasedAppVariables, updateUnReleasedAppVariables } from '../../../../../../../../api/variable'

  const props = defineProps<{
    bkBizId: string;
    appId: number;
  }>()

  const isSliderShow = ref(false)
  const loading = ref(false)
  const variableList = ref<IConfigVariableItem[]>([])
  const isFormChange = ref(false)
  const pending = ref(false)

  const getVariableList = async() => {
    loading.value = true
    const res = await getUnReleasedAppVariables(props.bkBizId, props.appId)
    variableList.value = res.details
    loading.value = false
  }

  const handleOpenSlider = () => {
    isSliderShow.value = true
    getVariableList()
  }

  const handleVariablesChange = () => {
    isFormChange.value = true
  }

  const handleSubmit = () => {}

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
    <VariablesTable
      class="variables-table-content"
      :list="variableList"
      :editable="true"
      :show-cited="true"
      @change="handleVariablesChange" />
    <section class="action-btns">
      <bk-button theme="primary" :loading="pending" @click="handleSubmit">保存</bk-button>
      <bk-button @click="close">取消</bk-button>
    </section>
  </bk-sideslider>
</template>
<style lang="scss" scoped>
  .variables-table-content {
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
