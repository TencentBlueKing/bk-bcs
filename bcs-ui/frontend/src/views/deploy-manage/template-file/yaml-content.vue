<template>
  <div
    ref="contentRef"
    :class="[
      'flex flex-col overflow-hidden',
      'shadow-[0_2px_4px_0_rgba(0,0,0,0.16)]',
      'bg-[#2E2E2E] h-full rounded-t-sm',
    ]">
    <!-- 工具栏 -->
    <div
      :class="[
        'flex items-center justify-between pl-[24px] pr-[16px] h-[40px]',
        'border-b-[1px] border-solid border-[#000]'
      ]">
      <div class="text-[#C4C6CC] text-[14px]">
        <span></span>
        <bcs-button text @click="updateToHelm" v-if="!isEdit && !upgrade && renderMode !== 'Helm'">
          <span class="flex items-center">
            <i class="bcs-icon bcs-icon-shengji"></i>
            <span class="ml-[5px]">{{ $t('templateFile.button.upgrade') }}</span>
          </span>
        </bcs-button>
        <bcs-tag
          ext-cls="!bg-[#295039] text-[#48c169]"
          type="filled"
          v-show="renderMode === 'Helm' || upgrade">
          {{ $t('templateFile.tag.helm') }}</bcs-tag>
      </div>
      <span class="flex items-center text-[12px] gap-[20px] text-[#979BA5]">
        <!-- <i class="bk-icon icon-upload-cloud text-[14px] hover:text-[#699df4] cursor-pointer"></i> -->
        <AiAssistant ref="assistantRef" />
        <i
          :class="[
            'hover:text-[#699df4] cursor-pointer',
            isFullscreen ? 'bcs-icon bcs-icon-zoom-out' : 'bcs-icon bcs-icon-enlarge'
          ]"
          @click="switchFullScreen">
        </i>
      </span>
    </div>
    <bk-alert type="info" :show-icon="false" v-show="upgrade && renderMode !== 'Helm'">
      <div slot="title">
        <i class="bk-icon icon-info"></i>
        <span>{{ $t('templateFile.tips.helmdesc') }}</span>
        <bcs-button text @click="rollback" class="text-[12px]">{{ $t('templateFile.button.rollback') }}</bcs-button>
      </div>
    </bk-alert>
    <!-- 代码编辑器 -->
    <bcs-resize-layout
      placement="bottom"
      :border="false"
      :auto-minimize="true"
      :initial-divide="editorErr.message ? 100 : 0"
      :max="300"
      :min="100"
      :disabled="!editorErr.message"
      class="!h-0 flex-1 file-editor">
      <template #aside>
        <EditorStatus
          :message="editorErr.message"
          v-show="!!editorErr.message" />
      </template>
      <template #main>
        <CodeEditor
          :readonly="isEdit"
          :options="opt"
          multi-document
          :no-validate="renderMode === 'Helm' || upgrade"
          v-bkloading="{ isLoading: converting }"
          ref="codeEditorRef"
          v-model="content"
          @error="handleEditorErr" />
      </template>
    </bcs-resize-layout>
  </div>
</template>
<script setup lang="ts">
import { throttle } from 'lodash';
import * as monaco from 'monaco-editor';
import { computed, ref, watch } from 'vue';

import { TemplateSetService  } from '@/api/modules/new-cluster-resource';
import AiAssistant from '@/components/ai-assistant.vue';
import CodeEditor from '@/components/monaco-editor/new-editor.vue';
import useFullScreen from '@/composables/use-fullscreen';
import EditorStatus from '@/views/resource-view/resource-update/editor-status.vue';

const props = defineProps({
  isEdit: {
    type: Boolean,
    default: false,
  },
  value: {
    type: String,
    default: '',
  },
  version: {
    type: String,
    default: '',
  },
  renderMode: {
    type: String,
    default: 'Simple',
  },
  upgrade: {
    type: Boolean,
    default: false,
  },
  contentOrigin: {
    type: String,
    default: '',
  },
});

const emits = defineEmits(['updateUpgrade', 'setContentOrigin']);

const assistantRef = ref<InstanceType<typeof AiAssistant>>();
const codeEditorRef = ref<InstanceType<typeof CodeEditor>>();
const content = ref('');
const opt = computed<monaco.editor.IStandaloneEditorConstructionOptions>(() => {
  if (props.isEdit) return {
    roundedSelection: false,
    scrollBeyondLastLine: false,
    renderLineHighlight: 'none',
    minimap: { enabled: true },
  };

  return { minimap: { enabled: true } };
});

// yaml异常
const editorErr = ref({
  type: '',
  message: '',
});
function handleEditorErr(err: string) { // 捕获编辑器错误提示
  if (props.renderMode === 'Helm' || props.upgrade) return;
  editorErr.value.type = 'content'; // 编辑内容错误
  editorErr.value.message = err;
};

// 获取数据
const getData = () => content.value;

// 校验数据
const validate = async () => !editorErr.value.message;

// 全屏
const { contentRef, isFullscreen, switchFullScreen } = useFullScreen();
// 调用AI
const explainK8sIssue = throttle(() => {
  assistantRef.value?.handleSendMsg(editorErr.value.message);
  assistantRef.value?.showAITips();
}, 300);

// 升级为helm
const converting = ref(false);
async function updateToHelm() {
  if (!content.value) return;
  converting.value = true;
  const result = await TemplateSetService.changeToHelm({
    content: content.value,
  }).finally(() => converting.value = false);
  emits('setContentOrigin', content.value);
  content.value = result?.content;
  setValue(content.value);
  emits('updateUpgrade', true);
}
// 撤销本次升级
function rollback() {
  content.value = props.contentOrigin;
  setValue(content.value);
  emits('updateUpgrade', false);
}

function setPosition(offset) {
  codeEditorRef.value?.setPosition(offset);
}

function setValue(value) {
  codeEditorRef.value?.setValue(value, '');
}

watch(() => props.value, () => {
  if (!props.value) return;
  content.value = props.value;
  setValue(props.value);
}, { immediate: true });

watch(() => editorErr.value.message, () => {
  if (!editorErr.value.message) return;

  explainK8sIssue();
});

watch([
  () => props.upgrade,
  () => props.renderMode,
], () => {
  if (props.upgrade || props.renderMode === 'Helm') {
    // 清空错误提示
    editorErr.value.message = '';
  }
});

defineExpose({
  getData,
  validate,
  setPosition,
  setValue,
});
</script>
<style scoped lang="postcss">
/deep/ .dark-form {
  .bk-label {
    color: #B3B3B3;
    font-size: 12px;
  }
  .bk-form-input {
    background-color: #2E2E2E;
    border: 1px solid #575757;
    color: #B3B3B3;
    &:focus {
      background-color: unset !important;
    }
  }
  .bk-select {
    border: 1px solid #575757;
    color: #B3B3B3;
  }
}

/deep/ .file-editor .bk-resize-layout-aside {
  border-color: #292929;
}
</style>
