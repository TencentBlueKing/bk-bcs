<template>
  <bk-form ref="formRef" form-type="vertical" :model="localVal" :rules="rules">
    <bk-form-item label="配置文件名称" property="key" :required="true">
      <bk-input v-model="localVal.key" :disabled="!editable" @change="change" />
    </bk-form-item>
    <bk-form-item label="配置类型" property="kv_type" :required="true">
      <bk-radio-group v-model="localVal.kv_type">
        <bk-radio
          v-for="kvType in CONFIG_KV_TYPE"
          :key="kvType.id"
          :label="kvType.id"
          :disabled="appData.spec.data_type !== 'any'"
          >{{ kvType.name }}</bk-radio
        >
      </bk-radio-group>
    </bk-form-item>
    <bk-form-item label="配置值" property="value" :required="true">
      <bk-input
        v-if="localVal.kv_type === 'string' || localVal.kv_type === 'number'"
        v-model="localVal!.value"
        @change="change"
      />
      <KvConfigContentEditor
        v-else
        :languages="localVal.kv_type"
        :content="localVal.value"
        :editable="editable"
        :variables="props.variables"
        @change="handleStringContentChange"
      />
    </bk-form-item>
  </bk-form>
</template>

<script lang="ts" setup>
import { ref } from 'vue';
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
    variables?: IVariableEditParams[];
    bkBizId: string;
    id: number; // 服务ID或者模板空间ID
    isTpl?: boolean; // 是否未模板配置文件，非模板配置文件和模板配置文件的上传、下载接口参数有差异
  }>(),
  {
    editable: true,
  },
);

const formRef = ref();

const rules = {
  value: [
    {
      validator: (value: string) => {
        if (localVal.value.kv_type === 'number') {
          return /^\d+$/.test(value);
        }
        return true;
      },
      message: '配置值不为数字',
    },
  ],
};

const validate = async () => {
  await formRef.value.validate();
  return true;
};

const localVal = ref({
  ...props.config,
  kv_type: appData.value.spec.data_type! === 'any' ? 'string' : appData.value.spec.data_type!,
});


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
