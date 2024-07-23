<template>
  <Teleport :disabled="!isOpenFullScreen" to="body">
    <div :class="['script-editor', { fullscreen: isOpenFullScreen, 'is-show-var': isShowVariable }]">
      <div class="editor-header">
        <div class="head-title">
          <slot name="header"></slot>
        </div>
        <div class="actions">
          <span
            v-if="!props.isPreview"
            v-bk-tooltips="{
              content: t('内置变量'),
              placement: 'top',
              distance: 20,
            }"
            :class="['bk-bscp-icon', 'icon-variable', { 'show-var': isShowVariable }]"
            @click="emits('update:isShowVariable', !isShowVariable)"></span>
          <ReadFileContent
            v-if="props.uploadIcon"
            v-bk-tooltips="{
              content: t('上传'),
              placement: 'top',
              distance: 20,
            }"
            class="action-icon"
            @completed="handleFileReadComplete" />
          <FilliscreenLine
            v-if="!isOpenFullScreen"
            class="action-icon"
            v-bk-tooltips="{
              content: t('全屏'),
              placement: 'top',
              distance: 20,
            }"
            @click="handleOpenFullScreen" />
          <UnfullScreen
            v-else
            class="action-icon"
            v-bk-tooltips="{
              content: t('退出全屏'),
              placement: 'bottom',
              distance: 20,
            }"
            @click="handleCloseFullScreen" />
        </div>
      </div>
      <div class="content-wrapper">
        <slot name="preContent" :fullscreen="isOpenFullScreen"></slot>
        <CodeEditor
          :model-value="props.modelValue"
          :lf-eol="true"
          :editable="props.editable"
          :language="props.language"
          @change="emits('update:modelValue', $event)" />
        <slot name="sufContent" :fullscreen="isOpenFullScreen" :is-show-variable="isShowVariable"></slot>
      </div>
    </div>
  </Teleport>
</template>
<script setup lang="ts">
  import { ref } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { FilliscreenLine, UnfullScreen } from 'bkui-vue/lib/icon';
  import BkMessage from 'bkui-vue/lib/message';
  import ReadFileContent from '../../service/detail/config/components/read-file-content.vue';
  import CodeEditor from '../../../../components/code-editor/index.vue';

  const { t } = useI18n();

  const props = withDefaults(
    defineProps<{
      modelValue: string;
      isShowVariable?: boolean;
      language?: string;
      editable?: boolean;
      uploadIcon?: boolean;
      isPreview?: boolean; // 是否是脚本预览
    }>(),
    {
      editable: true,
      uploadIcon: true,
      isPreview: false,
    },
  );

  const emits = defineEmits(['update:modelValue', 'update:isShowVariable']);

  const isOpenFullScreen = ref(false);

  const handleFileReadComplete = (content: string) => {
    emits('update:modelValue', content);
  };
  // 打开全屏
  const handleOpenFullScreen = () => {
    isOpenFullScreen.value = true;
    window.addEventListener('keydown', handleEscClose, { once: true });
    BkMessage({
      theme: 'primary',
      message: t('按 Esc 即可退出全屏模式'),
    });
  };

  const handleCloseFullScreen = () => {
    isOpenFullScreen.value = false;
    window.removeEventListener('keydown', handleEscClose);
  };

  // Esc按键事件处理
  const handleEscClose = (event: KeyboardEvent) => {
    if (event.code === 'Escape') {
      isOpenFullScreen.value = false;
    }
  };
</script>
<style lang="scss" scoped>
  .script-editor {
    &.fullscreen {
      position: fixed;
      top: 0;
      left: 0;
      width: 100vw;
      height: 100vh;
      z-index: 5000;
      .content-wrapper {
        height: calc(100vh - 40px);
      }
      &.is-show-var {
        :deep(.code-editor-wrapper) {
          width: calc(100vw - 272px);
        }
        :deep(.var-wrap) {
          position: absolute;
          right: 0;
          top: 40px;
          .content {
            height: calc(100% - 40px);
            .example {
              height: 100%;
            }
            .bk-textarea {
              height: 100%;
            }
          }
        }
      }
    }
  }
  .editor-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding-right: 14px;
    background: #2e2e2e;
    .head-title {
      flex: 1;
    }
    .actions {
      display: flex;
      align-items: center;
      .action-icon {
        color: #979ba5;
        cursor: pointer;
        &:hover {
          color: #3a84ff;
        }
      }
    }
  }
  .content-wrapper {
    height: 600px;
  }
  .pre-content {
    height: 100%;
  }
  .bk-bscp-icon {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    font-size: 12px;
    width: 24px;
    height: 24px;
    margin-right: 10px;
    color: #979ba5;
    cursor: pointer;
  }
  .show-var {
    background: #000000;
    border-radius: 2px;
    color: #bdbfc5;
  }
</style>
