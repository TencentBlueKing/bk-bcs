<template>
  <div :class="height" class="flex items-center overflow-hidden" ref="contentRef">
    <div
      :class="[
        'flex flex-col',
        'shadow-[0_2px_4px_0_rgba(0,0,0,0.16)]',
        'flex-1 w-0 bg-[#2E2E2E] h-full rounded-t-sm'
      ]">
      <!-- 工具栏 -->
      <div
        :class="[
          'flex items-center justify-between pl-[24px] pr-[16px] h-[40px]',
          'border-b-[1px] border-solid border-[#000]'
        ]">
        <span class="text-[#C4C6CC] text-[14px]"></span>
        <span class="flex items-center text-[12px] gap-[20px] text-[#979BA5]">
          <AiAssistant ref="assistantRef" preset="KubernetesProfessor" />
          <i
            :class="[
              'hover:text-[#699df4] cursor-pointer',
              isFullscreen ? 'bcs-icon bcs-icon-zoom-out' : 'bcs-icon bcs-icon-enlarge'
            ]"
            @click="switchFullScreen">
          </i>
        </span>
      </div>
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
            :multi-document="multiDocument"
            ref="codeEditorRef"
            v-model="content"
            @error="handleEditorErr" />
        </template>
      </bcs-resize-layout>
    </div>
  </div>
</template>
<script setup lang="ts">
import { throttle } from 'lodash';
import { computed, ref, watch } from 'vue';

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
  multiDocument: {
    type: Boolean,
    default: true,
  },
  height: {
    type: String,
    default: '',
  },
});

const assistantRef = ref<InstanceType<typeof AiAssistant>>();
const codeEditorRef = ref<InstanceType<typeof CodeEditor>>();
const content = ref('');

// yaml异常
const editorErr = ref({
  type: '',
  message: '',
});
function handleEditorErr(err: string) { // 捕获编辑器错误提示
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

watch(() => props.value, () => {
  if (!props.value) return;
  content.value = props.value;
  codeEditorRef.value?.setValue(props.value, '');
}, { immediate: true });

watch(() => editorErr.value.message, () => {
  if (!editorErr.value.message) return;

  explainK8sIssue();
});

defineExpose({
  getData,
  validate,
  content,
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
