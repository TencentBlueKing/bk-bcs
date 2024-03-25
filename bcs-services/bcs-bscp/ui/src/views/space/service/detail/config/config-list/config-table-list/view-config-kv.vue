<template>
  <bk-sideslider
    ref="sideSliderRef"
    width="640"
    quick-close
    :title="t('查看配置项')"
    :is-show="props.show"
    @closed="close"
    @shown="setEditorHeight">
    <div class="view-wrap">
      <bk-tab v-model:active="activeTab" type="card-grid" ext-cls="view-config-tab">
        <bk-tab-panel name="content" :label="t('配置项信息')">
          <bk-form label-width="100" form-type="vertical">
            <bk-form-item :label="t('配置项名称')">{{ props.config.spec.key }}</bk-form-item>
            <bk-form-item :label="t('配置项类型')">{{ props.config.spec.kv_type }}</bk-form-item>
            <bk-form-item :label="t('配置项值')">
              <span v-if="props.config.spec.kv_type === 'string' || props.config.spec.kv_type === 'number'">
                {{ props.config.spec.value }}
              </span>
              <div v-else class="editor-wrap">
                <kvConfigContentEditor
                  :content="props.config.spec.value"
                  :editable="false"
                  :height="editorHeight"
                  :languages="props.config.spec.kv_type" />
              </div>
            </bk-form-item>
          </bk-form>
        </bk-tab-panel>
        <bk-tab-panel name="meta" :label="t('元数据')">
          <ConfigContentEditor
            language="json"
            :content="JSON.stringify(metaData, null, 2)"
            :editable="false"
            :show-tips="false" />
        </bk-tab-panel>
      </bk-tab>
    </div>
    <section class="action-btns">
      <bk-button @click="close">{{ t('关闭') }}</bk-button>
    </section>
  </bk-sideslider>
</template>
<script setup lang="ts">
  import { ref, computed, watch, nextTick } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { IConfigKvType } from '../../../../../../../../types/config';
  import kvConfigContentEditor from '../../components/kv-config-content-editor.vue';
  import ConfigContentEditor from '../../components/config-content-editor.vue';

  const { t } = useI18n();
  const props = defineProps<{
    config: IConfigKvType;
    show: boolean;
  }>();

  const emits = defineEmits(['update:show', 'confirm']);

  const activeTab = ref('content');
  const isFormChange = ref(false);
  const sideSliderRef = ref();
  const editorHeight = ref(0);

  const metaData = computed(() => {
    const { content_spec, revision, spec } = props.config;
    const { byte_size, signature } = content_spec;
    const { create_at, creator, reviser, update_at } = revision;
    const { key, kv_type } = spec;
    return { key, kv_type, byte_size, signature, create_at, creator, reviser, update_at };
  });

  watch(
    () => props.show,
    (val) => {
      if (val) {
        isFormChange.value = false;
        activeTab.value = 'content';
      }
    },
  );

  const setEditorHeight = () => {
    nextTick(() => {
      const el = sideSliderRef.value.$el.querySelector('.view-wrap');
      editorHeight.value = el.offsetHeight > 310 ? el.offsetHeight - 310 : 300;
    });
  };

  const close = () => {
    emits('update:show', false);
  };
</script>
<style lang="scss" scoped>
  .view-wrap {
    height: calc(100vh - 101px);
    font-size: 12px;
    overflow: hidden;
    .view-config-tab {
      height: 100%;
      overflow: auto;
      :deep(.bk-tab-header) {
        padding: 8px 24px 0;
        font-size: 14px;
        background: #eaebf0;
      }
      :deep(.bk-tab-content) {
        padding: 24px 40px;
        box-shadow: none;
      }
    }
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
