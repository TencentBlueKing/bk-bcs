<template>
  <bk-button @click="handleOpenSlider">{{ t('查看变量') }}</bk-button>
  <bk-sideslider width="960" :title="t('查看变量')" :is-show="isSliderShow" @closed="close">
    <VariablesTable
      class="variables-table-content"
      :list="variableList"
      :cited-list="citedList"
      :editable="false"
      :show-cited="true" />
    <section class="action-btns">
      <bk-button @click="close">{{ t('关闭') }}</bk-button>
    </section>
  </bk-sideslider>
</template>
<script lang="ts" setup>
  import { ref } from 'vue';
  import { useI18n } from 'vue-i18n';
  import VariablesTable from './variables-table.vue';
  import { IVariableEditParams, IVariableCitedByConfigDetailItem } from '../../../../../../../../../types/variable';
  import { getReleasedAppVariables, getReleasedAppVariablesCitedDetail } from '../../../../../../../../api/variable';

  const { t } = useI18n();
  const props = defineProps<{
    bkBizId: string;
    appId: number;
    verisionId: number;
  }>();

  const isSliderShow = ref(false);
  const loading = ref(false);
  const variableList = ref<IVariableEditParams[]>([]);
  const citedList = ref<IVariableCitedByConfigDetailItem[]>([]);

  const getVariableList = async () => {
    loading.value = true;
    const [variableListRes, citedListRes] = await Promise.all([
      getReleasedAppVariables(props.bkBizId, props.appId, props.verisionId),
      getReleasedAppVariablesCitedDetail(props.bkBizId, props.appId, props.verisionId),
    ]);
    variableList.value = variableListRes.details;
    citedList.value = citedListRes.details;
    loading.value = false;
  };

  const handleOpenSlider = () => {
    isSliderShow.value = true;
    getVariableList();
  };

  const close = () => {
    isSliderShow.value = false;
    variableList.value = [];
  };
</script>
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
