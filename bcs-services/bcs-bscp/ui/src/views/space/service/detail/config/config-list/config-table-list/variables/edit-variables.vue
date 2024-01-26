<template>
  <bk-button
    v-cursor="{ active: !hasEditServicePerm }"
    :class="{ 'bk-button-with-no-perm': !hasEditServicePerm }"
    @click="handleOpenSlider"
    :disabled="variableList.length === 0"
  >
    {{ t('设置变量') }}
  </bk-button>
  <bk-sideslider width="960" :title="t('设置变量')" :is-show="isSliderShow" :before-close="handleBeforeClose" @closed="close">
    <bk-loading :loading="loading" class="variables-table-content">
      <ResetDefaultValue class="reset-default" :bk-biz-id="bkBizId" :list="initialVariables" @reset="handleResetDefault" />
      <VariablesTable
        ref="tableRef"
        :list="variableList"
        :cited-list="citedList"
        :editable="true"
        :show-cited="true"
        @change="handleVariablesChange"
      />
    </bk-loading>
    <section class="action-btns">
      <bk-button theme="primary" :loading="pending" @click="handleSubmit">{{t('保存')}}</bk-button>
      <bk-button @click="close">{{ t('取消') }}</bk-button>
    </section>
  </bk-sideslider>
</template>
<script lang="ts" setup>
import { ref, onMounted } from 'vue';
import { useI18n } from 'vue-i18n';
import Message from 'bkui-vue/lib/message';
import { storeToRefs } from 'pinia';
import useServiceStore from '../../../../../../../../store/service';
import useModalCloseConfirmation from '../../../../../../../../utils/hooks/use-modal-close-confirmation';
import { IVariableEditParams, IVariableCitedByConfigDetailItem } from '../../../../../../../../../types/variable';
import {
  getUnReleasedAppVariables,
  getUnReleasedAppVariablesCitedDetail,
  updateUnReleasedAppVariables,
} from '../../../../../../../../api/variable';
import VariablesTable from './variables-table.vue';
import ResetDefaultValue from './reset-default-value.vue';

const serviceStore = useServiceStore();
const { permCheckLoading, hasEditServicePerm } = storeToRefs(serviceStore);
const { checkPermBeforeOperate } = serviceStore;
const { t } = useI18n();

const props = defineProps<{
  bkBizId: string;
  appId: number;
}>();

const isSliderShow = ref(false);
const loading = ref(false);
const initialVariables = ref<IVariableEditParams[]>([]); // 变量列表默认配置，用来恢复默认值
const variableList = ref<IVariableEditParams[]>([]);
const citedList = ref<IVariableCitedByConfigDetailItem[]>([]);
const tableRef = ref();
const isFormChange = ref(false);
const pending = ref(false);

onMounted(() => {
  getVariableList();
});


const getVariableList = async () => {
  loading.value = true;
  const [variableListRes, citedListRes] = await Promise.all([
    getUnReleasedAppVariables(props.bkBizId, props.appId),
    getUnReleasedAppVariablesCitedDetail(props.bkBizId, props.appId),
  ]);
  initialVariables.value = variableListRes.details.slice();
  variableList.value = variableListRes.details.slice();
  citedList.value = citedListRes.details;
  loading.value = false;
};

const handleOpenSlider = () => {
  if (permCheckLoading.value || !checkPermBeforeOperate('update')) {
    return;
  }
  isSliderShow.value = true;
};

const handleVariablesChange = (variables: IVariableEditParams[]) => {
  isFormChange.value = true;
  variableList.value = variables;
};

const handleResetDefault = (list: IVariableEditParams[]) => {
  variableList.value = list;
};

const handleSubmit = async () => {
  if (!tableRef.value.validate()) {
    return;
  }
  try {
    pending.value = true;
    await updateUnReleasedAppVariables(props.bkBizId, props.appId, variableList.value);
    close();
    Message({
      theme: 'success',
      message: t('设置变量成功'),
    });
  } catch (e) {
    console.log(e);
  } finally {
    pending.value = false;
  }
};

const handleBeforeClose = async () => {
  if (isFormChange.value) {
    const result = await useModalCloseConfirmation();
    return result;
  }
  return true;
};

const close = () => {
  isSliderShow.value = false;
  isFormChange.value = false;
};

defineExpose({ getVariableList });
</script>
<style lang="scss" scoped>
.variables-table-content {
  position: relative;
  padding: 48px 24px;
  height: calc(100vh - 101px);
  overflow: auto;
}
.reset-default {
  position: absolute;
  top: 20px;
  right: 24px;
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
