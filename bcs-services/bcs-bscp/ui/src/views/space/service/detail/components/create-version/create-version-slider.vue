<template>
  <bk-sideslider
    :title="t('生成版本')"
    :is-show="props.show"
    :width="640"
    :before-close="handleBeforeClose"
    @closed="close">
    <bk-loading :loading="pending"  :title="t('版本生成中')">
      <div class="slider-form-content">
      <div class="version-basic-form">
        <div class="section-title">{{t('版本信息')}}</div>
        <bk-form class="form-wrapper" form-type="vertical" ref="formRef" :rules="rules" :model="formData">
          <bk-form-item :label="t('版本名称')" property="name" :required="true">
            <bk-input v-model="formData.name" :placeholder="t('请输入')"  @change="formChange" />
          </bk-form-item>
          <bk-form-item :label="t('版本描述')" property="memo">
            <bk-input
              v-model="formData.memo"
              type="textarea"
              :placeholder="t('请输入')"
              :maxlength="200"
              @change="formChange"
              :resize="true" />
          </bk-form-item>
          <bk-checkbox v-model="isPublish" :true-label="true" :false-label="false" @change="formChange">
            <span style="font-size: 12px;">{{ t('同时上线版本') }}</span>
          </bk-checkbox>
        </bk-form>
      </div>
      <div class="variable-form" v-if="isFileType">
        <div v-bkloading="{ loading }" class="section-title">{{ t('服务变量赋值') }}</div>
        <ResetDefaultValue class="reset-default-btn" :list="initialVariables" @reset="handleResetDefault" />
        <VariablesTable ref="tableRef" :list="variableList" :editable="true" @change="handleVariablesChange" />
      </div>
    </div>
    </bk-loading>
    <div class="action-btns">
      <bk-button theme="primary" @click="confirm">{{ t('确定') }}</bk-button>
      <bk-button @click="close">{{ t('取消') }}</bk-button>
    </div>
  </bk-sideslider>
</template>
<script lang="ts" setup>
import { ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { assign } from 'lodash';
import { GET_UNNAMED_VERSION_DATA } from '../../../../../../constants/config';
import { createVersion } from '../../../../../../api/config';
import useModalCloseConfirmation from '../../../../../../utils/hooks/use-modal-close-confirmation';
import { IVariableEditParams } from '../../../../../../../types/variable';
import { getUnReleasedAppVariables } from '../../../../../../api/variable';
import useServiceStore from '../../../../../../store/service';
import { storeToRefs } from 'pinia';
import VariablesTable from '../../config/config-list/config-table-list/variables/variables-table.vue';
import ResetDefaultValue from '../../config/config-list/config-table-list/variables/reset-default-value.vue';

const props = defineProps<{
  show: boolean;
  bkBizId: string;
  appId: number;
  isDiffSliderShow: boolean;
}>();

const emits = defineEmits(['update:show', 'created', 'open-diff']);
const { t } = useI18n();

const serviceStore = useServiceStore();
const { isFileType } = storeToRefs(serviceStore);

const formData = ref({
  name: '',
  memo: '',
});
const isPublish = ref(false);
const loading = ref(false);
const initialVariables = ref<IVariableEditParams[]>([]);
const variableList = ref<IVariableEditParams[]>([]);
const formRef = ref();
const tableRef = ref();
const isFormChange = ref(false);
const pending = ref(false);
const rules = {
  name: [
    {
      validator: (value: string) => {
        if (value.length > 0) {
          return /^[\u4e00-\u9fa5a-zA-Z0-9][\u4e00-\u9fa5a-zA-Z0-9_\-.]*[\u4e00-\u9fa5a-zA-Z0-9]?$/.test(value);
        }
        return true;
      },
      message: t('仅允许使用中文、英文、数字、下划线、中划线、点，且必须以中文、英文、数字开头和结尾'),
    },
  ],
  memo: [
    {
      validator: (value: string) => value.length <= 200,
      message: t('最大长度200个字符'),
    },
  ],
};

watch(
  () => props.show,
  (val) => {
    if (val) {
      formData.value = {
        name: '',
        memo: '',
      };
      isPublish.value = false;
      getVariableList();
    }
  },
);

const getVariableList = async () => {
  loading.value = true;
  const res = await getUnReleasedAppVariables(props.bkBizId, props.appId);
  initialVariables.value = res.details.slice();
  variableList.value = res.details.slice();
  loading.value = false;
};

const formChange = () => {
  isFormChange.value = true;
};

const handleVariablesChange = (variables: IVariableEditParams[]) => {
  formChange();
  variableList.value = variables;
};

// const handleOpenDiff = async () => {
//   await formRef.value.validate();
//   if (!tableRef.value.validate()) {
//     return;
//   }
//   emits('open-diff', variableList.value);
// };

const confirm = async () => {
  if (!formRef.value.validate() || (isFileType.value && !tableRef.value.validate())) return;
  try {
    await formRef.value.validate();
    pending.value = true;
    const params = {
      name: formData.value.name,
      memo: formData.value.memo,
      variables: variableList.value,
    };
    const res = await createVersion(props.bkBizId, props.appId, params);
    // 创建接口未返回完整的版本详情数据，在前端拼接最新版本数据，加载完版本列表后再更新
    const newVersionData = assign({}, GET_UNNAMED_VERSION_DATA(), {
      id: res.data.id,
      spec: { name: formData.value.name, memo: formData.value.memo },
    });
    emits('created', newVersionData, isPublish.value);
  } catch (e) {
    Promise.reject(e);
  } finally {
    pending.value = false;
  }
};

const handleResetDefault = (list: IVariableEditParams[]) => {
  variableList.value = list;
};

const handleBeforeClose = async () => {
  if (props.isDiffSliderShow) {
    return false;
  }
  if (isFormChange.value) {
    const result = await useModalCloseConfirmation();
    return result;
  }
  return true;
};

const close = () => {
  emits('update:show', false);
};

defineExpose({
  confirm,
});
</script>
<style lang="scss" scoped>
.slider-form-content {
  position: relative;
  padding: 16px 24px;
  height: calc(100vh - 101px);
  overflow: auto;
}
.version-basic-form {
  padding-bottom: 24px;
  border-bottom: 1px solid #dcdee5;
}
.variable-form {
  position: relative;
  margin-top: 8px;
  .reset-default-btn {
    position: absolute;
    top: 4px;
    right: 10px;
  }
}
.section-title {
  margin: 0 0 16px;
  font-size: 14px;
  font-weight: 700;
  color: #63656e;
}
.action-btns {
  border-top: 1px solid #dcdee5;
  padding: 8px 24px;
  background: #fafbfd;
  .bk-button {
    margin-right: 8px;
    min-width: 88px;
  }
}
.form-wrapper {
  &:deep(.bk-form-label){
    font-size: 12px !important;
  }
}
</style>
