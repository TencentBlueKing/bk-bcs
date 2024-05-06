<template>
  <div ref="mancoEditor" :class="['biz-manco-editor', { 'full-screen': isFullScreen }]">
    <div class="build-code-fullscreen" :title="isFullScreen ? $t('generic.button.close') : $t('generic.button.fullScreen.text')" @click="setFullScreen()">
      <i class="bcs-icon bcs-icon-full-screen" v-if="!isFullScreen"></i>
      <i class="bcs-icon bcs-icon-close" v-else></i>
    </div>
  </div>
</template>
<script>
import * as monaco from 'monaco-editor';
import assign from 'nano-assign';

import monokaiTheme from './theme.json';

export default {
  name: 'MonacoEditor',
  model: {
    event: 'change',
  },
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
      default: 'yaml',
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

  data() {
    return {
      isFullScreen: false,
      defaultWidth: 0,
      defaultHeight: 0,
    };
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
    this.$nextTick(() => {
      const rect = this.$refs.mancoEditor?.getBoundingClientRect();
      this.defaultWidth = rect.width;
      this.defaultHeight = rect.height;
    });
  },

  beforeDestroy() {
    this.editor?.dispose();
  },

  methods: {
    initMonaco(monaco) {
      this.$emit('editorWillMount', this.monaco);

      const options = assign({
        value: this.value,
        autoIndent: true,
        theme: this.theme,
        language: this.language,
      }, this.options);

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

    setFullScreen() {
      this.isFullScreen = !this.isFullScreen;
      // eslint-disable-next-line @typescript-eslint/no-this-alias
      const self = this;
      if (this.isFullScreen) {
        this.$nextTick(() => {
          this.editor.layout({
            width: window.innerWidth,
            height: window.innerHeight,
          });
        });
      } else {
        setTimeout(() => {
          this.editor.layout({
            width: self.defaultWidth,
            height: self.defaultHeight,
          });
        }, 0);
      }
    },
  },
};
</script>

<style lang="postcss" scoped>
    .biz-manco-editor {
    position: relative;

    &.full-screen {
        position: fixed;
        left: 0;
        right: 0;
        top: 0;
        bottom: 0;
        margin: auto;
        z-index: 10000;
        height: auto !important;
        background: #272822;
    }

    .build-code-fullscreen {
        padding: 7px;
        cursor: pointer;
        position: absolute;
        right: 10px;
        color: #fafbfd;
        z-index: 10;
        font-size: 16px;
        i.icon-full-screen {
            font-weight: 700;
        }
    }
}
</style>
