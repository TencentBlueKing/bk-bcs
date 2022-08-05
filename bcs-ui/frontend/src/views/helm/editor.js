/*
* Tencent is pleased to support the open source community by making
* 蓝鲸智云PaaS平台社区版 (BlueKing PaaS Community Edition) available.
*
* Copyright (C) 2021 THL A29 Limited, a Tencent company.  All rights reserved.
*
* 蓝鲸智云PaaS平台社区版 (BlueKing PaaS Community Edition) is licensed under the MIT License.
*
* License for 蓝鲸智云PaaS平台社区版 (BlueKing PaaS Community Edition):
*
* ---------------------------------------------------
* Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated
* documentation files (the "Software"), to deal in the Software without restriction, including without limitation
* the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and
* to permit persons to whom the Software is furnished to do so, subject to the following conditions:
*
* The above copyright notice and this permission notice shall be included in all copies or substantial portions of
* the Software.
*
* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO
* THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
* AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF
* CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS
* IN THE SOFTWARE.
*/
import assign from 'nano-assign';
import monokaiTheme from './theme.json';
import * as monaco from 'monaco-editor';

self.MonacoEnvironment = {
  getWorkerUrl() {
    return `${window.DEVOPS_BCS_HOST}${window.STATIC_URL}${window.VERSION_STATIC_URL}/editor.worker.js`;
  },
};

export default {
  name: 'MonacoEditor',

  props: {
    original: {
      type: String,
      default: '',
    },
    value: {
      type: String,
      required: true,
    },
    theme: {
      type: String,
      default: 'vs',
    },
    language: {
      type: String,
      default: 'javascript',
    },
    options: {
      type: Object,
      default() {
        return {};
      },
    },
    amdRequire: {
      type: Function,
    },
    diffEditor: {
      type: Boolean,
      default: false,
    },
  },

  model: {
    event: 'change',
  },

  watch: {
    options: {
      deep: true,
      handler(options) {
        if (this.editor) {
          const editor = this.getModifiedEditor();
          editor.updateOptions(options);
        }
      },
    },

    value(newValue) {
      if (this.editor) {
        const editor = this.getModifiedEditor();
        if (newValue !== editor.getValue()) {
          editor.setValue(newValue);
        }
      }
    },

    language(newVal) {
      if (this.editor) {
        const editor = this.getModifiedEditor();
        this.monaco.editor.setModelLanguage(editor.getModel(), newVal);
      }
    },

    theme(newVal) {
      if (this.editor) {
        this.monaco.editor.setTheme(newVal);
      }
    },
  },

  mounted() {
    this.monaco = monaco;
    this.monaco.editor.defineTheme('monokai', monokaiTheme);
    this.initMonaco(monaco);
    // if (this.amdRequire) {
    //     this.amdRequire(['vs/editor/editor.main'], () => {
    //         this.monaco = window.monaco
    //         this.initMonaco(window.monaco)
    //     })
    // } else {
    //     // ESM format so it can't be resolved by commonjs `require` in eslint
    //     // eslint-disable-next-line import/no-unresolved
    //     const monaco = require('monaco-editor')
    //     this.monaco = monaco
    //     this.initMonaco(monaco)
    // }
  },

  beforeDestroy() {
    // eslint-disable-next-line @typescript-eslint/prefer-optional-chain
    this.editor && this.editor.dispose();
  },

  methods: {
    initMonaco(monaco) {
      this.$emit('editorWillMount', this.monaco);

      const options = assign(
        {
          value: this.value,
          autoIndent: true,
          theme: this.theme,
          language: this.language,
        },
        this.options,
      );

      if (this.diffEditor) {
        this.editor = monaco.editor.createDiffEditor(this.$el, options);
        const originalModel = monaco.editor.createModel(
          this.original,
          this.language,
        );
        const modifiedModel = monaco.editor.createModel(
          this.value,
          this.language,
        );
        this.editor.setModel({
          original: originalModel,
          modified: modifiedModel,
        });
      } else {
        this.editor = monaco.editor.create(this.$el, options);
      }

      // @event `change`
      const editor = this.getModifiedEditor();
      editor.onDidChangeModelContent((event) => {
        const value = editor.getValue();
        if (this.value !== value) {
          this.$emit('change', value, event);
          this.$emit('input', value, event);
        }
      });

      this.$emit('mounted', this.editor, this.monaco.editor);
    },

    /** @deprecated */
    getMonaco() {
      return this.editor;
    },

    getEditor() {
      return this.editor;
    },

    getModifiedEditor() {
      return this.diffEditor ? this.editor.getModifiedEditor() : this.editor;
    },

    focus() {
      this.editor.focus();
    },
  },

  render(h) {
    return h('div');
  },
};
