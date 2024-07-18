<template>
  <div>
    <CodeEditor
      ref="codeEditorRef"
      :model-value="props.codeVal"
      :language="language"
      :editable="false"
      line-numbers="off"
      :minimap="false"
      :vertical-scrollbar-size="0"
      :horizon-scrollbar-size="0"
      render-line-highlight="none"
      :render-indent-guides="false"
      :variables="props.variables"
      :folding="false"
      :always-consume-mouse-wheel="false"
      :contextmenu="false"
      @update:model-value="emits('change', $event)" />
  </div>
</template>
<script lang="ts" setup>
  import { onBeforeUnmount, ref } from 'vue';
  import CodeEditor from '../../../../../components/code-editor/index.vue';
  import { IVariableEditParams } from '../../../../../../types/variable';

  const props = defineProps<{
    language: string;
    codeVal: string;
    variables?: IVariableEditParams[];
  }>();

  const emits = defineEmits(['change']);

  const codeEditorRef = ref();

  onBeforeUnmount(() => {
    codeEditorRef.value.destroy();
  });

  const scrollTo = () => {
    codeEditorRef.value.scrollToTop();
  };

  defineExpose({
    scrollTo,
  });
</script>

<style scoped lang="scss">
  :deep(.monaco-editor) {
    background-color: unset;
    // 取消默认背景色
    &.monaco-editor-background {
      background-color: unset;
    }
    // 行号占位背景，取消了也会占一定的空间
    .margin {
      background-color: #f5f7fa;
    }
    // 代码主区域背景色
    .monaco-editor-background {
      background-color: #f5f7fa;
    }
    // 取消选中时的边框
    // .view-overlays .current-line {
    //   border: none;
    // }
    // 取消滚动时上方一滩阴影
    .scroll-decoration {
      box-shadow: none;
    }
    // key的颜色
    .mtk23 {
      color: #37b4a0;
    }
    // value的颜色, 增加优先级
    .mtk5,
    .mtk6 {
      color: #d2734b;
    }
    // 注释颜色
    .mtk1,
    .mtk7 {
      color: #63656e;
    }
    // 取消链接下划线
    .detected-link,
    .detected-link-active {
      text-decoration: none;
    }
    // 高亮代码部分背景色
    .view-line .template-variable-item {
      color: #63656e;
      border: none;
      background-color: #ffd695;
      font-weight: 700;
    }
  }
</style>
