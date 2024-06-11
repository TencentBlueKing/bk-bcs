<template>
  <div>
    <CodeEditor
      ref="codeEditorRef"
      v-model="codeStrVal"
      :language="'yaml'"
      :editable="false"
      line-numbers="off"
      :minimap="false"
      :vertical-scrollbar-size="0"
      :horizon-scrollbar-size="0"
      render-line-highlight="none"
      :custom-class-name="testObj" />
  </div>
</template>
<script lang="ts" setup>
  import { onMounted, onBeforeUnmount, ref, watch } from 'vue';
  import CodeEditor from '../../../../../components/code-editor/index.vue';
  const codeEditorRef = ref();
  const props = defineProps<{
    codeVal: string;
  }>();
  const codeStrVal = ref('');
  onMounted(() => {
    codeStrVal.value = props.codeVal;
  });

  // 装饰器测试数据
  const testObj = {
    className: 'abca1',
    rangeArr: [
      {
        rowStart: 1,
        columnStart: 1,
        rowEnd: 10,
        columnEnd: 20,
      },
      {
        rowStart: 11,
        columnStart: 1,
        rowEnd: 20,
        columnEnd: 20,
      },
    ],
  };
  const scrollTo = () => {
    codeEditorRef.value.scrollToTop();
  };
  defineExpose({
    scrollTo,
  });
  watch(
    () => {
      return props.codeVal;
    },
    (newV: string) => {
      codeStrVal.value = newV;
    },
  );
  onBeforeUnmount(() => {
    codeEditorRef.value.destroy();
  });
</script>

<style scoped lang="scss">
  .abca1 {
    color: red;
  }
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
  }
</style>
