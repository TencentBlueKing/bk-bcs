<template>
  <Teleport :disabled="!isOpenFullScreen" to="body">
    <div :class="['config-content-editor', { fullscreen: isOpenFullScreen }]">
      <div class="editor-title">
        <div class="tips">
          <InfoLine class="info-icon" />
          仅支持大小不超过 100M
        </div>
        <div class="btns">
          <Transfer
            v-bk-tooltips="{
              content: '分隔符',
              placement: 'top',
              distance: 20,
            }"
            @click="separatorShow = !separatorShow"
          />
          <Search
            v-bk-tooltips="{
              content: '搜索',
              placement: 'top',
              distance: 20,
            }"
            @click="handleSearch"
          />
          <FilliscreenLine
            v-if="!isOpenFullScreen"
            v-bk-tooltips="{
              content: '全屏',
              placement: 'top',
              distance: 20,
            }"
            @click="handleOpenFullScreen"
          />
          <UnfullScreen
            v-else
            v-bk-tooltips="{
              content: '退出全屏',
              placement: 'bottom',
              distance: 20,
            }"
            @click="handleCloseFullScreen"
          />
        </div>
      </div>
      <div class="editor-content">
        <CodeEditor
          ref="codeEditorRef"
          :model-value="variables"
          @update:model-value="variables = $event"
          @enter="separatorShow = true"
        />
        <div class="separator" v-show="separatorShow">
          <SeparatorSelect @closed="separatorShow = false" @confirm="separator = $event" />
        </div>
      </div>
    </div>
  </Teleport>
</template>
<script setup lang="ts">
import { ref, onBeforeUnmount } from 'vue';
import BkMessage from 'bkui-vue/lib/message';
import { InfoLine, FilliscreenLine, UnfullScreen, Search, Transfer } from 'bkui-vue/lib/icon';
// import { batchImportTemplateVariables } from '../../../api/variable';
import CodeEditor from '../../../components/code-editor/index.vue';
import SeparatorSelect from './separator-select.vue';

const isOpenFullScreen = ref(false);
const codeEditorRef = ref();
const separatorShow = ref(false);
const variables = ref('');
const separator = ref(' ');
onBeforeUnmount(() => {
  codeEditorRef.value.destroy();
});

// 打开全屏
const handleOpenFullScreen = () => {
  isOpenFullScreen.value = true;
  window.addEventListener('keydown', handleEscClose, { once: true });
  BkMessage({
    theme: 'primary',
    message: '按 Esc 即可退出全屏模式',
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

const handleSearch = () => {
  const variablesList = variables.value.split('\r\n');
};

// 导入之前 校验编辑器内容
// const handleValidateEditor = () => {};
</script>
<style lang="scss" scoped>
.config-content-editor {
  height: 640px;
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
      & > span {
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
