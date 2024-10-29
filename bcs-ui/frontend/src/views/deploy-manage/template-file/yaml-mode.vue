<template>
  <div class="overflow-hidden" ref="contentRef">
    <bcs-resize-layout
      collapsible
      disabled
      :border="false"
      ref="yamlLayoutRef"
      initial-divide="230px"
      class="h-full"
      @collapse-change="handleCollapseChange">
      <div
        slot="aside"
        class="bg-[#fff] h-full overflow-y-auto overflow-x-hidden">
        <left-nav
          :list="yamlToJson"
          :active-index="activeContentIndex"
          @cellClick="({ item }) => handleAnchor(item)">
          <template #default="{ item }">
            <span class="bcs-ellipsis" v-bk-overflow-tips>{{ item?.name }}</span>
          </template>
        </left-nav>
      </div>
      <div
        slot="main"
        :class="[
          'flex flex-col',
          'shadow-[0_2px_4px_0_rgba(0,0,0,0.16)]',
          'bg-[#2E2E2E] h-full rounded-t-sm',
          isCollapse ? '' : 'ml-[16px]'
        ]">
        <!-- 工具栏 -->
        <div
          :class="[
            'flex items-center justify-between pl-[24px] pr-[16px] h-[40px]',
            'border-b-[1px] border-solid border-[#000]'
          ]">
          <span class="text-[#C4C6CC] text-[14px]"></span>
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
              ref="codeEditorRef"
              v-model="content"
              @error="handleEditorErr" />
          </template>
        </bcs-resize-layout>
      </div>
    </bcs-resize-layout>
  </div>
</template>
<script setup lang="ts">
import yamljs from 'js-yaml';
import { throttle } from 'lodash';
import * as monaco from 'monaco-editor';
import { computed, ref, watch } from 'vue';

import leftNav from './left-nav.vue';

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
});

const assistantRef = ref<InstanceType<typeof AiAssistant>>();
const codeEditorRef = ref<InstanceType<typeof CodeEditor>>();
const content = ref('');
const activeContentIndex = ref(0);
const yamlToJson = computed(() => {
  let offset = 0;
  return content.value.split('---')
    .reduce<Array<{ name: string; offset: number }>>((pre, doc) => {
    const name = yamljs.load(doc)?.metadata?.name;
    if (name) {
      pre.push({
        name,
        offset,
      });
      offset += doc.length;
    }
    return pre;
  }, []);
});
const opt = computed<monaco.editor.IStandaloneEditorConstructionOptions>(() => {
  if (props.isEdit) return {
    roundedSelection: false,
    scrollBeyondLastLine: false,
    renderLineHighlight: 'none',
    minimap: { enabled: true },
  };

  return { minimap: { enabled: true } };
});

const yamlLayoutRef = ref();
const watchOnce = watch(yamlToJson, () => {
  // 只有一项数据时折叠起来
  if (yamlToJson.value && yamlToJson.value.length < 2) {
    yamlLayoutRef.value?.setCollapse(true);
  }
  watchOnce();
});


// yaml异常
const editorErr = ref({
  type: '',
  message: '',
});
function handleEditorErr(err: string) { // 捕获编辑器错误提示
  editorErr.value.type = 'content'; // 编辑内容错误
  editorErr.value.message = err;
};

// 跳转到对应的yaml
const handleAnchor = (item: typeof yamlToJson.value[number]) => {
  const index = yamlToJson.value.findIndex(d => d === item);
  codeEditorRef.value?.setPosition(item.offset);
  activeContentIndex.value = index;
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

const isCollapse = ref(false);
const handleCollapseChange = (value: boolean) => {
  isCollapse.value = value;
};

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
