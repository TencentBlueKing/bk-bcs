import * as monaco from 'monaco-editor';
import { ref } from 'vue';

import BcsEditorTheme from './theme.json';

// 初始化内置主题
monaco.editor.defineTheme('bcs-theme-yaml', BcsEditorTheme);

export interface IConfig {
  language: string
  theme: string
  readonly: boolean
  diffEditor: boolean
  onDidUpdateDiff: (event) => void
  onDidChangeModelContent: (value: IDiffValue | string, event) => void
}

export interface IDiffValue {
  modified: string
  original: string
}

export type IOptions = monaco.editor.IDiffEditorOptions | monaco.editor.IEditorOptions;

// monaco通用编辑器
export default function useEditor(config?: Partial<IConfig>) {
  const {
    language,
    readonly,
    diffEditor,
    theme,
    onDidUpdateDiff,
    onDidChangeModelContent,
  } = Object.assign({
    theme: 'bcs-theme-yaml',
    language: 'yaml',
    readonly: false,
    diffEditor: false,
  }, config || {});

  const editor = ref<monaco.editor.IStandaloneCodeEditor | monaco.editor.IStandaloneDiffEditor | null>(null);
  const navi = ref<monaco.editor.IDiffNavigator|null>(null); // diff 模式 navigator
  const modifiedModel = ref<monaco.editor.ITextModel|null>(null);
  const originalModel = ref<monaco.editor.ITextModel|null>(null);
  const editorErr = ref<any>('');
  const diffStat = ref({
    insert: 0,
    delete: 0,
    changesCount: 0,
  }); // diff统计

  // 初始化编辑器
  const initMonaco = (el: HTMLElement, options: IOptions = {}) => {
    if (!el) {
      console.warn('editor el is null');
      return;
    };

    destroyMonaco();
    const opt = {
      language,
      theme,
      minimap: {
        enabled: false,
      },
      readOnly: readonly,
      automaticLayout: true,
      scrollbar: {
        alwaysConsumeMouseWheel: false,
      },
      contextmenu: !readonly,
      ...options,
    };
    if (diffEditor) {
      // diff模式
      editor.value = monaco.editor.createDiffEditor(el, opt);
      navi.value = monaco.editor.createDiffNavigator(editor.value, {
        followsCaret: true, // resets the navigator state when the user selects something in the editor
        ignoreCharChanges: true, // jump from line to line
      });
      editor.value?.onDidUpdateDiff((event) => {
        diffStat.value = setDiffStat();
        onDidUpdateDiff?.(event);
      });
    } else {
      // 编辑器
      editor.value = monaco.editor.create(el, opt);
      editor.value?.onDidChangeModelContent((event) => {
        try {
          editorErr.value = '';
          onDidChangeModelContent?.(getValue(), event);
        } catch (err: any) {
          editorErr.value = err?.message || String(err);
        }
      });
    }
    const { modified, original } = setModel();
    modifiedModel.value = modified;
    originalModel.value = original;
    return editor.value;
  };

  // 销毁编辑器
  const destroyMonaco = () => {
    editor.value?.dispose();
    editor.value = null;
  };

  // 设置模式
  const setModel = (modified?: string, original?: string) => {
    const modifiedModel = monaco.editor.createModel(modified || '', language);
    let originalModel: monaco.editor.ITextModel|null = null;
    if (diffEditor) {
      // diff模式
      originalModel = monaco.editor.createModel(original || '', language);
      (editor.value as monaco.editor.IStandaloneDiffEditor)?.setModel({
        original: originalModel,
        modified: modifiedModel,
      });
    } else {
      // 编辑模式
      (editor.value as monaco.editor.IStandaloneCodeEditor)?.setModel(modifiedModel);
    }

    return {
      modified: modifiedModel,
      original: originalModel,
    };
  };
  // 获取模式
  const getModel = () => ({
    modified: modifiedModel.value,
    original: originalModel.value,
  });

  // 设置编辑器值
  const setValue = (value: string, original?: string) => {
    try {
      modifiedModel.value?.setValue(value || '');
      originalModel.value?.setValue(original || '');
    } catch (err: any) {
      editorErr.value = err?.message || String(err);
    }
  };

  // 获取编辑器值
  const getValue = (): IDiffValue | string => {
    const modified = modifiedModel.value?.getValue() || '';
    if (diffEditor) {
      return {
        modified,
        original: originalModel.value?.getValue() || '',
      };
    }
    return modified;
  };

  // 更新配置
  const updateOptions = (options: IOptions) => {
    editor.value?.updateOptions(options);
  };

  // 设置语言
  const setModelLanguage = (lang: string) => {
    const { original, modified } = getModel();
    original && monaco.editor.setModelLanguage(original, lang);
    modified && monaco.editor.setModelLanguage(modified, lang);
  };

  // 设置diff模式下统计信息
  const setDiffStat = () => {
    const diffStat = {
      insert: 0,
      delete: 0,
      changesCount: 0,
    };
    const changes = (editor.value as monaco.editor.IStandaloneDiffEditor)?.getLineChanges() || [];
    diffStat.changesCount = changes.length;
    changes.forEach((item) => {
      if ((item.originalEndLineNumber >= item.originalStartLineNumber) && item.originalEndLineNumber > 0) {
        diffStat.delete += item.originalEndLineNumber - item.originalStartLineNumber + 1;
      }

      if ((item.modifiedEndLineNumber >= item.modifiedStartLineNumber) && item.modifiedEndLineNumber > 0) {
        diffStat.insert += item.modifiedEndLineNumber - item.modifiedStartLineNumber + 1;
      }
    });
    return diffStat;
  };

  // 设置位置
  const setPosition = (offset: number) => {
    const { modified } = getModel() || {};
    const pos = modified?.getPositionAt(offset);
    if (!pos) return;

    editor.value?.revealPositionNearTop(pos);
  };

  // 容器大小变化时重新调整编辑器布局
  const layout = () => {
    editor.value?.layout();
  };

  return {
    editorErr,
    diffStat,
    editor,
    navi,
    modifiedModel,
    originalModel,
    initMonaco,
    setModel,
    getModel,
    destroyMonaco,
    setValue,
    getValue,
    updateOptions,
    setModelLanguage,
    layout,
    setDiffStat,
    setPosition,
  };
}
