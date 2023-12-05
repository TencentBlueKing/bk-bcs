<template>
  <bk-form ref="formRef" form-type="vertical" :model="localVal" :rules="rules">
    <bk-form-item label="配置项名称" property="key" :required="true">
      <bk-input v-model="localVal.key" :disabled="editable || view" @change="change" />
    </bk-form-item>
    <bk-form-item label="数据类型" property="kv_type" :required="true">
      <bk-radio-group v-model="localVal.kv_type">
        <bk-radio
          v-for="kvType in CONFIG_KV_TYPE"
          :key="kvType.id"
          :label="kvType.id"
          :disabled="appData.spec.data_type !== 'any' || editable || view"
          >{{ kvType.name }}</bk-radio
        >
      </bk-radio-group>
    </bk-form-item>
    <bk-form-item label="配置项值" property="value" :required="true">
      <bk-input
        v-if="localVal.kv_type === 'string' || localVal.kv_type === 'number'"
        v-model.trim="localVal!.value"
        @change="change"
        :disabled="view"
      />
      <KvConfigContentEditor
        v-else
        :languages="localVal.kv_type"
        :content="localVal.value"
        :editable="!view"
        :variables="props.variables"
        @change="handleStringContentChange"
      />
    </bk-form-item>
  </bk-form>
</template>

<script lang="ts" setup>
import { ref, onMounted } from 'vue';
import { CONFIG_KV_TYPE } from '../../../../../../../constants/config';
import KvConfigContentEditor from '../../components/kv-config-content-editor.vue';
import { IConfigKvEditParams } from '../../../../../../../../types/config';
import { IVariableEditParams } from '../../../../../../../../types/variable';
import useServiceStore from '../../../../../../../store/service';
import { storeToRefs } from 'pinia';

const serviceStore = useServiceStore();
const { appData } = storeToRefs(serviceStore);

const props = withDefaults(
  defineProps<{
    config: IConfigKvEditParams;
    editable: boolean;
    view: boolean;
    variables?: IVariableEditParams[];
    bkBizId: string;
    id: number; // 服务ID或者模板空间ID
    isTpl?: boolean; // 是否未模板配置文件，非模板配置文件和模板配置文件的上传、下载接口参数有差异
  }>(),
  {
    editable: false,
    view: false,
  },
);

const formRef = ref();
const localVal = ref({
  ...props.config,
});

const rules = {
  value: [
    {
      validator: (value: string) => {
        if (localVal.value.kv_type === 'number') {
          return /^-?\d+(\.\d+)?$/.test(value);
        }
        return true;
      },
      message: '配置项值不为数字',
    },
  ],
};

// 编辑文件任意类型默认选中string
onMounted(() => {
  if (props.editable) {
    localVal.value.kv_type = appData.value.spec.data_type! === 'any' ? 'string' : appData.value.spec.data_type!;
  }
});

const validate = async () => {
  await formRef.value.validate();
  return true;
};


const emits = defineEmits(['change']);


const handleStringContentChange = (val: string) => {
  localVal.value!.value = val;
  change();
};

const change = () => {
  emits('change', localVal.value);
};

defineExpose({ validate });
</script>

<style scoped lang="scss"></style>
