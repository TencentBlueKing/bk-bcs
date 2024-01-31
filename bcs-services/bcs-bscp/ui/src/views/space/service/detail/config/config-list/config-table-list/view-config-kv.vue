<template>
  <bk-sideslider width="640" quick-close :title="t('查看配置项')" :is-show="props.show" @closed="close">
    <div class="view-wrap">
      <bk-form label-width="100" form-type="vertical">
        <bk-form-item :label="t('配置项名称')">{{ props.config.key }}</bk-form-item>
        <bk-form-item :label="t('配置项类型')">{{ props.config.kv_type }}</bk-form-item>
        <bk-form-item :label="t('配置项值')">
          <span v-if="props.config.kv_type === 'string' || props.config.kv_type === 'number'">
            {{ props.config.value }}
          </span>
          <div v-else class="editor-wrap">
            <kvConfigContentEditor :content="props.config.value" :editable="false" :languages="props.config.kv_type" />
          </div>
        </bk-form-item>
      </bk-form>
    </div>
    <section class="action-btns">
      <bk-button @click="close">{{ t('关闭') }}</bk-button>
    </section>
  </bk-sideslider>
</template>
<script setup lang="ts">
  import { ref, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { IConfigKvItem } from '../../../../../../../../types/config';
  import kvConfigContentEditor from '../../components/kv-config-content-editor.vue';

  const { t } = useI18n();
  const props = defineProps<{
    config: IConfigKvItem;
    show: boolean;
  }>();

  const emits = defineEmits(['update:show', 'confirm']);

  const configForm = ref<IConfigKvItem>();
  const isFormChange = ref(false);

  watch(
    () => props.show,
    (val) => {
      if (val) {
        isFormChange.value = false;
        configForm.value = props.config;
      }
    },
  );

  const close = () => {
    emits('update:show', false);
  };
</script>
<style lang="scss" scoped>
  .view-wrap {
    padding: 20px 24px;
    height: calc(100vh - 101px);
    font-size: 12px;
    overflow: auto;
    :deep(.bk-form-item) {
      margin-bottom: 24px;
      .bk-form-label,
      .bk-form-content {
        font-size: 12px;
      }
      .bk-form-label {
        line-height: 26px;
        color: #979ba5;
      }
      .bk-form-content {
        line-height: normal;
        color: #323339;
      }
    }
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
