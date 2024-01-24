<template>
  <Teleport :disabled="!isOpenFullScreen" to="body">
    <div :class="['config-content-editor', { fullscreen: isOpenFullScreen }]">
      <div class="editor-title">
        <div class="tips">
          <InfoLine class="info-icon" />
          {{ t('仅支持大小不超过') }}2M
        </div>
        <div class="btns">
          <i
            class="bk-bscp-icon icon-separator"
            v-bk-tooltips="{
              content: t('分隔符'),
              placement: 'top',
              distance: 20,
            }"
            @click="separatorShow = !separatorShow"
          />
          <Search
            v-bk-tooltips="{
              content: t('搜索'),
              placement: 'top',
              distance: 20,
            }"
            @click="codeEditorRef.openSearch()"
          />
          <FilliscreenLine
            v-if="!isOpenFullScreen"
            v-bk-tooltips="{
              content: t('全屏'),
              placement: 'top',
              distance: 20,
            }"
            @click="handleOpenFullScreen"
          />
          <UnfullScreen
            v-else
            v-bk-tooltips="{
              content: t('退出全屏'),
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
          v-model="kvsContent"
          @enter="separatorShow = true"
          :error-line="errorLine"
          :placeholder="editorPlaceholder"
        />
        <div class="separator" v-show="separatorShow">
          <SeparatorSelect @closed="separatorShow = false" @confirm="handleSelectSeparator" />
        </div>
      </div>
    </div>
  </Teleport>
</template>
<script setup lang="ts">
import { ref, onBeforeUnmount, watch } from 'vue';
import { useRoute } from 'vue-router';
import { useI18n } from 'vue-i18n';
import BkMessage from 'bkui-vue/lib/message';
import { InfoLine, FilliscreenLine, UnfullScreen, Search } from 'bkui-vue/lib/icon';
import CodeEditor from '../../../../../../components/code-editor/index.vue';
import SeparatorSelect from '../../../../variables/separator-select.vue';
import { IConfigKvItem } from '../../../../../../../types/config';
import { batchUpsertKv } from '../../../../../../api/config';

interface errorLineItem {
  lineNumber: number;
  errorInfo: string;
}

const { t } = useI18n();
const emits = defineEmits(['trigger']);

const route = useRoute();
const isOpenFullScreen = ref(false);
const codeEditorRef = ref();
const separatorShow = ref(false);
const kvsContent = ref('');
const kvs = ref<IConfigKvItem[]>([]);
const separator = ref(' ');
const errorLine = ref<errorLineItem[]>([]);
const editorPlaceholder = ref([t('格式：'), t('key 类型 value'), 'name string nginx', ' port number 8080']);
const bkBizId = ref(String(route.params.spaceId));
const appId = ref(Number(route.params.appId));

watch(
  () => kvsContent.value,
  (val) => {
    handleValidateEditor();
    if (!val) emits('trigger', false);
  },
);

onBeforeUnmount(() => {
  codeEditorRef.value.destroy();
});
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

// 校验编辑器内容 处理上传kv格式
const handleValidateEditor = () => {
  const kvArray = kvsContent.value.split('\n');
  errorLine.value = [];
  kvs.value = [];
  kvArray.forEach((item, index) => {
    if (item === '') return;
    const kvContent = item.split(separator.value);
    const key = kvContent[0];
    const kv_type = kvContent[1] ? kvContent[1].toLowerCase() : '';
    const value = kvContent[2];
    kvs.value.push({
      key,
      kv_type,
      value,
    });
    if (kvContent.length !== 3) {
      errorLine.value.push({
        errorInfo: t('请检查是否已正确使用分隔符'),
        lineNumber: index + 1,
      });
    } else if (kv_type !== 'string' && kv_type !== 'number') {
      errorLine.value.push({
        errorInfo: t('类型必须为 string 或者 number'),
        lineNumber: index + 1,
      });
    } else if (kv_type === 'number' && !/^\d+(\.\d+)?$/.test(value)) {
      errorLine.value.push({
        errorInfo: t('类型为number 值不为number'),
        lineNumber: index + 1,
      });
    } else if (value === '') {
      errorLine.value.push({
        errorInfo: t('value不能为空'),
        lineNumber: index + 1,
      });
    }
  });
  emits('trigger', kvsContent.value && errorLine.value.length === 0);
};

// 导入kv
const handleImport = async () => {
  await batchUpsertKv(bkBizId.value, appId.value, kvs.value);
};

const handleSelectSeparator = (selectSeparator: string) => {
  separator.value = selectSeparator;
  handleValidateEditor();
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
