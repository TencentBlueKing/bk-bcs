<template>
  <bk-form ref="formRef" form-type="vertical" :model="localVal" :rules="rules">
    <div class="form-row">
      <bk-form-item :label="t('配置项名称')" property="key" :required="true">
        <bk-input
          class="name-input"
          v-model="localVal.key"
          :disabled="props.editMode"
          @input="change"
          :placeholder="t('请输入')" />
      </bk-form-item>
      <bk-form-item :label="t('数据类型')" property="kv_type" :required="true" :description="typeDescription">
        <bk-select v-model="localVal.kv_type" class="type-select" :disabled="selectDisabled">
          <bk-option v-for="kvType in CONFIG_KV_TYPE" :key="kvType.id" :id="kvType.id" :name="kvType.name" />
        </bk-select>
      </bk-form-item>
    </div>
    <bk-form-item :label="t('配置项描述')" property="memo">
      <bk-input v-model="localVal.memo" type="textarea" :maxlength="200" :placeholder="t('请输入')" @input="change" />
    </bk-form-item>
    <bk-form-item :label="t('配置项值')" property="value" :required="true">
      <bk-input
        v-if="localVal.kv_type === 'string' || localVal.kv_type === 'number'"
        v-model.trim="localVal!.value"
        class="value-input"
        @input="change" />
      <KvConfigContentEditor
        v-else
        ref="KvCodeEditorRef"
        :languages="localVal.kv_type"
        :content="localVal.value"
        :height="editorHeight"
        @change="handleStringContentChange" />
    </bk-form-item>
  </bk-form>
</template>

<script lang="ts" setup>
  import { ref, onMounted, computed, nextTick } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { CONFIG_KV_TYPE } from '../../../../../../../constants/config';
  import KvConfigContentEditor from '../../components/kv-config-content-editor.vue';
  import { IConfigKvEditParams } from '../../../../../../../../types/config';
  import useServiceStore from '../../../../../../../store/service';
  import { storeToRefs } from 'pinia';

  const serviceStore = useServiceStore();
  const { appData } = storeToRefs(serviceStore);
  const { t } = useI18n();

  const props = withDefaults(
    defineProps<{
      config: IConfigKvEditParams;
      editMode?: boolean;
      bkBizId: string;
      id: number; // 服务ID或者模板空间ID
    }>(),
    {
      editMode: false,
    },
  );

  const KvCodeEditorRef = ref();
  const formRef = ref();
  const localVal = ref({
    ...props.config,
  });
  const editorHeight = ref(0);

  const typeDescription = computed(() => {
    if (appData.value.spec.data_type !== 'any' && !props.editMode) {
      return t('已限制该服务下所有配置项数据类型为{n}，如需其他数据类型，请调整服务属性下的数据类型', {
        n: appData.value.spec.data_type,
      });
    }
    return '';
  });

  const selectDisabled = computed(() => appData.value.spec.data_type !== 'any' || props.editMode);

  const rules = {
    key: [
      {
        validator: (value: string) => value.length <= 128,
        message: t('最大长度128个字符'),
      },
      {
        validator: (value: string) =>
          /^[\p{Script=Han}\p{L}\p{N}]([\p{Script=Han}\p{L}\p{N}_-]*[\p{Script=Han}\p{L}\p{N}])?$/u.test(value),
        message: t('只允许包含中文、英文、数字、下划线 (_)、连字符 (-)，并且必须以中文、英文、数字开头和结尾'),
      },
    ],
    memo: [
      {
        validator: (value: string) => value.length <= 200,
        message: t('最大长度200个字符'),
      },
    ],
    value: [
      {
        validator: (value: string) => {
          if (localVal.value.kv_type === 'number') {
            return /^-?\d+(\.\d+)?$/.test(value);
          }
          return true;
        },
        message: t('配置项值不为数字'),
      },
    ],
  };

  onMounted(() => {
    if (!props.editMode) {
      localVal.value.kv_type = appData.value.spec.data_type! === 'any' ? 'string' : appData.value.spec.data_type!;
    }
    nextTick(() => {
      const editorMinHeight = 300; // 编辑器最小高度
      const remainingHeight = formRef.value.$el.offsetHeight - 355; // 容器高度减去其他元素已占用高度
      editorHeight.value = remainingHeight > editorMinHeight ? remainingHeight : editorMinHeight;
    });
  });

  const validate = async () => {
    await formRef.value.validate();
    switch (localVal.value.kv_type) {
      case 'json':
      case 'xml':
      case 'yaml':
        return KvCodeEditorRef.value.validate();
    }
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
<style lang="scss" scoped>
  :deep(.bk-form-item:last-child) {
    margin-bottom: 0;
  }
  .form-row {
    display: flex;
    justify-content: space-between;
    .name-input,
    .type-select {
      width: 428px;
    }
  }
  .value-input {
    width: 428px;
  }
</style>
