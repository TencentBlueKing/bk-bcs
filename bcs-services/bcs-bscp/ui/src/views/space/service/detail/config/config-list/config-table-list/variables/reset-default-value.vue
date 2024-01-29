<template>
  <bk-button class="reset-default-value-btn" text theme="primary" @click="triggerReset">
    {{ t('恢复默认值') }}
    <Help class="help-icon" v-bk-tooltips="{ content: t('如果以下变量存在于全局变量中，其值将被重置为全局变量的默认值'), placement: 'top' }" />
  </bk-button>
</template>
<script lang="ts" setup>
import { ref, watch, onMounted } from 'vue';
import { useI18n } from 'vue-i18n';
import { Help } from 'bkui-vue/lib/icon';
import { cloneDeep } from 'lodash';
import { IVariableEditParams, IVariableItem } from '../../../../../../../../../types/variable';
import { getVariableList } from '../../../../../../../../api/variable';
import { ICommonQuery } from '../../../../../../../../../types/index';

const { t } = useI18n();
const props = defineProps<{
  list: IVariableEditParams[];
  bkBizId: string
}>();

const emits = defineEmits(['reset']);

const initialList = ref(cloneDeep(props.list));
const variableList = ref<IVariableItem[]>();

watch(
  () => props.list,
  () => {
    initialList.value = cloneDeep(props.list);
  },
);

onMounted(() => {
  getVariable();
});

const getVariable = async () => {
  const params: ICommonQuery = {
    start: 0,
    all: true,
  };
  const res = await getVariableList(props.bkBizId, params);
  variableList.value = res.details;
};

const triggerReset = () => {
  initialList.value.forEach((item) => {
    const variable = variableList.value?.find(variable => variable.spec.name === item.name);
    if (variable) item.default_val = variable.spec.default_val;
  });
  emits('reset', cloneDeep(initialList.value));
};
</script>
<style lang="scss" scoped>
.reset-default-value-btn {
  font-size: 12px;
  .help-icon {
    margin-left: 4px;
  }
}
</style>
