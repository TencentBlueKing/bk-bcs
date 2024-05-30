<template>
  <Teleport :disabled="!isOpenFullScreen" to="body">
    <div :class="['config-content-editor', { fullscreen: isOpenFullScreen }]">
      <div class="editor-title">
        <div class="tips">
          <InfoLine class="info-icon" />
          {{ t('仅支持大小不超过') }} 5M
        </div>
        <div class="btns">
          <i
            class="bk-bscp-icon icon-separator"
            v-bk-tooltips="{
              content: t('分隔符'),
              placement: 'top',
              distance: 20,
            }"
            @click="separatorShow = !separatorShow" />
          <Search
            v-bk-tooltips="{
              content: t('搜索'),
              placement: 'top',
              distance: 20,
            }"
            @click="codeEditorRef.openSearch()" />
          <FilliscreenLine
            v-if="!isOpenFullScreen"
            v-bk-tooltips="{
              content: t('全屏'),
              placement: 'top',
              distance: 20,
            }"
            @click="handleOpenFullScreen" />
          <UnfullScreen
            v-else
            v-bk-tooltips="{
              content: t('退出全屏'),
              placement: 'bottom',
              distance: 20,
            }"
            @click="handleCloseFullScreen" />
        </div>
      </div>
      <div class="editor-content">
        <CodeEditor
          ref="codeEditorRef"
          v-model="variables"
          :error-line="errorLine"
          :placeholder="editorPlaceholder"
          @paste="handlePaste"
          @enter="separatorShow = true"
          @validate="handleValidateEditor" />
        <div class="separator" v-show="separatorShow">
          <SeparatorSelect @closed="separatorShow = false" @confirm="handleSelectSeparator" />
        </div>
      </div>
    </div>
  </Teleport>
</template>
<script setup lang="ts">
  import { ref, onBeforeUnmount, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import BkMessage from 'bkui-vue/lib/message';
  import { InfoLine, FilliscreenLine, UnfullScreen, Search } from 'bkui-vue/lib/icon';
  import { batchImportTemplateVariables } from '../../../api/variable';
  import CodeEditor from '../../../components/code-editor/index.vue';
  import SeparatorSelect from './separator-select.vue';
  import useGlobalStore from '../../../store/global';
  import { storeToRefs } from 'pinia';

  interface errorLineItem {
    lineNumber: number;
    errorInfo: string;
  }

  const { t } = useI18n();
  const emits = defineEmits(['trigger']);

  const isOpenFullScreen = ref(false);
  const codeEditorRef = ref();
  const separatorShow = ref(false);
  const variables = ref('');
  const separator = ref(' ');
  const shouldValidate = ref(false);
  const errorLine = ref<errorLineItem[]>([]);
  const editorPlaceholder = ref([
    t('示例：'),
    t('变量名 变量类型 变量值 变量描述（可选）'),
    'bk_bscp_nginx_ip string 1.1.1.1',
    t(' bk_bscp_nginx_port number 8080 nginx端口'),
  ]);

  watch(
    () => variables.value,
    (val) => {
      handleValidateEditor();
      if (!val) emits('trigger', false);
    },
  );

  watch(
    () => errorLine.value,
    (val) => {
      shouldValidate.value = val.length > 0;
    },
  );

  onBeforeUnmount(() => {
    codeEditorRef.value.destroy();
  });
  const { spaceId } = storeToRefs(useGlobalStore());
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

  // 校验编辑器内容
  const handleValidateEditor = () => {
    const variablesArray = variables.value.split('\n').map((item) => item.trim());
    errorLine.value = [];
    let hasSeparatorError = false;
    variablesArray.forEach((item, index) => {
      if (item === '') return;
      const regex = separator.value === ' ' ? /\s+/ : separator.value;
      const variablesContent = item.split(regex).map((item) => item.trim());
      const key = variablesContent[0];
      const type = variablesContent[1];
      const value = variablesContent[2];
      if (variablesContent.length < 3) {
        errorLine.value.push({
          errorInfo: t('请检查是否已正确使用分隔符'),
          lineNumber: index + 1,
        });
        hasSeparatorError = true;
      } else if (!key.startsWith('bk_bscp_') && !key.startsWith('BK_BSCP_')) {
        errorLine.value.push({
          errorInfo: t('变量必须以bk_bscp_或BK_BSCP_开头'),
          lineNumber: index + 1,
        });
      } else if (type !== 'string' && type !== 'number') {
        errorLine.value.push({
          errorInfo: t('类型必须为 string 或者 number'),
          lineNumber: index + 1,
        });
      } else if (type === 'number' && !/^\d+(\.\d+)?$/.test(value)) {
        errorLine.value.push({
          errorInfo: t('类型为number 值不为number'),
          lineNumber: index + 1,
        });
      }
    });
    emits('trigger', variables.value && errorLine.value.length === 0);
    return hasSeparatorError;
  };
  // 导入变量
  const handleImport = async () => {
    handleValidateEditor();
    shouldValidate.value = true;
    if (errorLine.value.length > 0) return Promise.reject();
    const params = {
      separator: separator.value === ' ' ? 'white-space' : separator.value,
      variables: variables.value,
    };
    await batchImportTemplateVariables(spaceId.value, params);
    BkMessage({
      theme: 'success',
      message: t('导入变量成功'),
    });
  };

  const handleSelectSeparator = (selectSeparator: string) => {
    separator.value = selectSeparator;
    handleValidateEditor();
  };

  const handlePaste = () => {
    if (handleValidateEditor()) separatorShow.value = true;
  };

  defineExpose({
    handleImport,
  });
</script>
<style lang="scss" scoped>
  .config-content-editor {
    height: 640px;
    padding-top: 10px;
    &.fullscreen {
      position: fixed;
      top: 0;
      left: 0;
      width: 100vw;
      height: 100vh;
      z-index: 5000;
    }
    .editor-title {
      display: flex;
      align-items: center;
      justify-content: space-between;
      padding: 0 16px;
      height: 40px;
      color: #979ba5;
      background: #2e2e2e;
      border-radius: 2px 2px 0 0;
      .tips {
        display: flex;
        align-items: center;
        font-size: 12px;
        .info-icon {
          margin-right: 4px;
        }
      }
      .btns {
        display: flex;
        justify-content: space-between;
        width: 80px;
        height: 16px;
        align-items: center;
        & > span,
        & > i {
          cursor: pointer;
          &:hover {
            color: #3a84ff;
          }
        }
      }
    }
    .editor-content {
      position: relative;
      height: calc(100% - 130px);
      .separator {
        position: absolute;
        right: 0;
        top: 0;
      }
    }
  }
</style>
