<script lang="ts" setup>
  import { ref } from 'vue'
  import VariablesTable from './variables-table.vue';
  import { IConfigVariableItem } from '../../../../../../../../../types/variable';
  import { getReleasedAppVariables } from '../../../../../../../../api/variable'

  const props = defineProps<{
    bkBizId: string;
    appId: number;
    verisionId: number
  }>()

  const isSliderShow = ref(false)
  const loading = ref(false)
  const variableList = ref<IConfigVariableItem[]>([])

  const getVariableList = async() => {
    loading.value = true
    const res = await getReleasedAppVariables(props.bkBizId, props.appId, props.verisionId)
    variableList.value = res.details
    loading.value = false
  }

  const handleOpenSlider = () => {
    isSliderShow.value = true
    getVariableList()
  }

  const close = () => {
    isSliderShow.value = false
  }

</script>
<template>
  <bk-button @click="handleOpenSlider">查看变量</bk-button>
  <bk-sideslider
    width="960"
    title="查看变量"
    :is-show="isSliderShow"
    @closed="close">
    <VariablesTable
      class="variables-table-content"
      :list="variableList"
      :editable="false"
      :show-cited="true" />
    <section class="action-btns">
      <bk-button @click="close">关闭</bk-button>
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
