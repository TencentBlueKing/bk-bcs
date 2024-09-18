<template>
  <div class="diff-wrapper">
    <div class="permission-diff" v-show="isShowPermissionDiff">
      <div class="left-header">{{ t('文件属性') }}</div>
      <section ref="permissionDiffRef" class="fill-diff-wrapper"></section>
    </div>
    <div v-show="props.isSecret && !props.secretVisible" class="un-view-diff-content">
      <div class="diff">{{ t('敏感数据不可见，无法查看实际内容') }}</div>
      <div class="diff">{{ t('敏感数据不可见，无法查看实际内容') }}</div>
    </div>
    <div v-show="props.secretVisible" class="text-diff">
      <div class="left-header" v-show="isShowPermissionDiff">{{ t('文件内容') }}</div>
      <section ref="textDiffRef" class="text-diff-wrapper"></section>
    </div>
    <div v-show="props.secretVisible" class="footer">
      <TextDiffLegend
        :permission-diff-number="permissionDiffNumber"
        :diff-editor="diffEditor"
        :permission-editor="permissionEditor" />
    </div>
  </div>
</template>
<script setup lang="ts">
  import { ref, watch, onMounted, onBeforeUnmount, computed } from 'vue';
  import { useI18n } from 'vue-i18n';
  import * as monaco from 'monaco-editor';
  import editorWorker from 'monaco-editor/esm/vs/editor/editor.worker.js?worker';
  import jsonWorker from 'monaco-editor/esm/vs/language/json/json.worker.js?worker';
  import cssWorker from 'monaco-editor/esm/vs/language/css/css.worker.js?worker';
  import htmlWorker from 'monaco-editor/esm/vs/language/html/html.worker.js?worker';
  import tsWorker from 'monaco-editor/esm/vs/language/typescript/ts.worker.js?worker';
  import { IVariableEditParams } from '../../../types/variable';
  import useDiffEditorVariableReplace from '../../utils/hooks/use-diff-editor-variable-replace';
  import TextDiffLegend from './text-diff-legend.vue';

  const { t } = useI18n();
  self.MonacoEnvironment = {
    getWorker(_, label) {
      if (label === 'json') {
        return new jsonWorker();
      }
      if (label === 'css' || label === 'scss' || label === 'less') {
        return new cssWorker();
      }
      if (label === 'html' || label === 'handlebars' || label === 'razor') {
        return new htmlWorker();
      }
      if (label === 'typescript' || label === 'javascript') {
        return new tsWorker();
      }
      return new editorWorker();
    },
  };

  const props = withDefaults(
    defineProps<{
      base: string;
      baseLanguage?: string;
      baseVariables?: IVariableEditParams[];
      basePermission?: string;
      current: string;
      currentLanguage?: string;
      currentVariables?: IVariableEditParams[];
      currentPermission?: string;
      isSecret?: boolean;
      secretVisible?: boolean;
      isCipherShow?: boolean;
    }>(),
    {
      baseLanguage: '',
      currentVariables: () => [],
      currentLanguage: '',
      baseVariables: () => [],
      isSecret: false,
      secretVisible: true,
    },
  );

  const textDiffRef = ref();
  const permissionDiffRef = ref();
  const permissionDiffNumber = ref(0);
  const cipherText = '********';
  let diffEditor: monaco.editor.IStandaloneDiffEditor;
  let diffEditorHoverProvider: monaco.IDisposable;
  let permissionEditor: monaco.editor.IStandaloneDiffEditor;

  const isShowPermissionDiff = computed(() => props.basePermission !== props.currentPermission);

  const baseShowContent = computed(() => {
    if (props.isSecret && props.isCipherShow) {
      return cipherText;
    }
    return props.base;
  });

  const currentShowContent = computed(() => {
    if (props.isSecret && props.isCipherShow) {
      return cipherText;
    }
    return props.current;
  });

  watch(
    () => [props.base, props.current],
    () => {
      updateModel();
      replaceDiffVariables();
    },
  );

  watch(
    () => [props.baseLanguage, props.currentLanguage],
    () => {
      updateModel();
      replaceDiffVariables();
    },
  );

  watch(
    () => [props.basePermission, props.currentPermission],
    () => {
      updateModel();
      replaceDiffVariables();
      getPermissionDiffNumber();
    },
  );

  watch(
    () => props.isCipherShow,
    () => {
      updateModel();
    },
  );

  onMounted(() => {
    createDiffEditor();
    replaceDiffVariables();
  });

  onBeforeUnmount(() => {
    diffEditor.dispose();
    permissionEditor.dispose();
    if (diffEditorHoverProvider) {
      diffEditorHoverProvider.dispose();
    }
  });

  const createDiffEditor = () => {
    if (diffEditor) {
      diffEditor.dispose();
      permissionEditor.dispose();
    }
    const originalModel = monaco.editor.createModel(baseShowContent.value, props.baseLanguage);
    const modifiedModel = monaco.editor.createModel(currentShowContent.value, props.currentLanguage);
    const originaPermissionModel = monaco.editor.createModel(props.basePermission as string, props.baseLanguage);
    const modifieFilldModel = monaco.editor.createModel(props.currentPermission as string, props.currentLanguage);

    diffEditor = monaco.editor.createDiffEditor(textDiffRef.value, {
      theme: 'vs-dark',
      automaticLayout: true,
      scrollBeyondLastLine: false,
      readOnly: true,
      unicodeHighlight: {
        ambiguousCharacters: false,
      },
    });
    diffEditor.setModel({
      original: originalModel,
      modified: modifiedModel,
    });
    permissionEditor = monaco.editor.createDiffEditor(permissionDiffRef.value, {
      theme: 'vs-dark',
      automaticLayout: true,
      readOnly: true,
      scrollBeyondLastLine: false,
      lineNumbers: () => '',
    });
    permissionEditor.setModel({
      original: originaPermissionModel,
      modified: modifieFilldModel,
    });
    const leftDiffEditor = diffEditor.getOriginalEditor();
    const rightDiffEditor = diffEditor.getModifiedEditor();
    const leftPermissionEditor = permissionEditor.getOriginalEditor();
    const rightPermissionEditor = permissionEditor.getModifiedEditor();
    leftDiffEditor.onDidChangeCursorPosition(() => {
      syncCursor(leftDiffEditor, rightDiffEditor);
    });
    rightDiffEditor.onDidChangeCursorPosition(() => {
      syncCursor(rightDiffEditor, leftDiffEditor);
    });
    leftPermissionEditor.onDidChangeCursorPosition(() => {
      syncCursor(leftPermissionEditor, rightPermissionEditor);
    });
    rightPermissionEditor.onDidChangeCursorPosition(() => {
      syncCursor(rightPermissionEditor, leftPermissionEditor);
    });
  };

  const updateModel = () => {
    const originalModel = monaco.editor.createModel(baseShowContent.value, props.baseLanguage);
    const modifiedModel = monaco.editor.createModel(currentShowContent.value, props.currentLanguage);
    const originaPermissionModel = monaco.editor.createModel(props.basePermission as string, props.baseLanguage);
    const modifiedPermissionModel = monaco.editor.createModel(props.currentPermission as string, props.currentLanguage);
    diffEditor.setModel({
      original: originalModel,
      modified: modifiedModel,
    });
    permissionEditor.setModel({
      original: originaPermissionModel,
      modified: modifiedPermissionModel,
    });
  };

  const replaceDiffVariables = () => {
    if (
      (props.currentVariables && props.currentVariables.length > 0) ||
      (props.baseVariables && props.baseVariables.length > 0)
    ) {
      diffEditorHoverProvider = useDiffEditorVariableReplace(diffEditor, props.currentVariables, props.baseVariables);
    }
  };

  // 实现两边的编辑器光标统一移动
  const syncCursor = (editorA: any, editorB: any) => {
    const positionA = editorA.getPosition();
    const positionB = editorB.getPosition();
    if (
      positionA &&
      positionB &&
      (positionA.lineNumber !== positionB.lineNumber || positionA.column !== positionB.column)
    ) {
      editorB.setPosition(positionA);
    }
  };

  // 获取文件属性差异个数
  const getPermissionDiffNumber = () => {
    const basePermissionList = props.basePermission?.split('\n');
    const currentPermissionList = props.currentPermission?.split('\n');
    let count = 0;
    basePermissionList?.forEach((item, index) => {
      if (item !== currentPermissionList![index]) {
        count += 1;
      }
    });
    permissionDiffNumber.value = count;
  };
</script>
<style lang="scss" scoped>
  .text-diff {
    flex: 1;
    overflow: hidden;
    border-top: 2px solid #2c2c2c;
    .text-diff-wrapper {
      height: 100%;
      :deep(.monaco-editor) {
        .template-variable-item {
          color: #3a84ff;
          border: 1px solid #1768ef;
          cursor: pointer;
        }
      }
    }
  }
  .permission-diff {
    height: 100px;
  }
  .fill-diff-wrapper {
    height: 80px;
    :deep(.monaco-editor) {
      .template-variable-item {
        color: #1768ef;
        border: 1px solid #1768ef;
        cursor: pointer;
      }
    }
  }
  :deep(.d2h-file-wrapper) {
    border: none;
  }

  .left-header {
    height: 20px;
    line-height: 20px;
    padding-left: 5px;
    color: #5a5a5b;
    background-color: #1e1e1e;
  }

  .footer {
    position: sticky;
    bottom: 0px;
    overflow: hidden;
    :deep(.navigator-wrap) {
      width: 100%;
    }
  }
  .diff-wrapper {
    display: flex;
    flex-direction: column;
    height: 100%;
  }
  .un-view-diff-content {
    display: flex;
    height: 100%;
    background: #1d1d1d;
    color: #979ba5;
    .diff {
      margin-top: 120px;
      text-align: center;
      flex: 1;
    }
  }
</style>
