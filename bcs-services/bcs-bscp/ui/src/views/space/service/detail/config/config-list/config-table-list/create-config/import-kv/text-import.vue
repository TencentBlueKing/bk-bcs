<template>
  <div class="wrap">
    <div class="label">{{ $t('文本格式') }}</div>
    <bk-radio-group v-model="selectFormat">
      <bk-radio label="text">{{ $t('简单文本') }}</bk-radio>
      <bk-radio label="json">JSON</bk-radio>
      <bk-radio label="yaml">YAML</bk-radio>
    </bk-radio-group>
    <div class="tips">{{ tips }}</div>
  </div>
  <div :class="['content-wrapper', { 'show-example': isShowFormateExample }]">
    <KvContentEditor
      v-model="isShowFormateExample"
      :bk-biz-id="props.bkBizId"
      :app-id="props.appId"
      :format="selectFormat">
      <template #sufContent>
        <FormatExample v-if="isShowFormateExample" :format="selectFormat" />
      </template>
    </KvContentEditor>
  </div>
</template>

<script lang="ts" setup>
  import { ref, computed } from 'vue';
  import { useI18n } from 'vue-i18n';
  import KvContentEditor from '../../../../components/kv-import-editor.vue';
  import FormatExample from './format-example.vue';

  const { t } = useI18n();

  const props = defineProps<{
    bkBizId: string;
    appId: number;
  }>();

  const selectFormat = ref('text');
  const isShowFormateExample = ref(true);

  const tips = computed(() => {
    if (selectFormat.value === 'text') {
      return t('每行表示一个配置项，包含配置项名称、数据类型和配置项值，默认通过空格分隔');
    }
    if (selectFormat.value === 'json') {
      return t(
        '以 JSON 格式导入键值 (KV) 配置项，配置项名称作为 JSON 对象的 Key，而配置项的数据类型和值组成一个嵌套对象，作为对应 Key 的 Value',
      );
    }
    return t(
      '以 YAML 格式导入键值 (KV) 配置项，配置项名称作为 YAML 对象的 Key，而配置项的数据类型和值分别作为嵌套对象的子键，形成对应键的值',
    );
  });
</script>

<style scoped lang="scss">
  :deep(.bk-radio-label) {
    font-size: 12px;
  }
  .tips {
    flex-basis: 100%;
    margin-left: 70px;
    font-size: 12px;
    color: #979ba5;
  }
  .content-wrapper {
    width: 100%;
    margin-top: 24px;
    &.show-example {
      :deep(.config-content-editor) {
        height: 484px;
        .code-editor-wrapper {
          width: calc(100% - 520px);
        }
      }
    }
  }
  :deep(.editor-content) {
    display: flex;
    .code-editor-wrapper {
      width: 100%;
    }
  }
</style>
