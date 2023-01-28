/* eslint-disable @typescript-eslint/no-require-imports */
/* eslint-disable @typescript-eslint/no-unused-vars */
/* eslint-disable @typescript-eslint/prefer-optional-chain */

// 不要再引入， 不要再引入，不要再引入，统一用monaco-editor/new-editor组件
module.exports = {
  template: '<div :style="{height: calcSize(height), width: calcSize(width)}"></div>',
  props: {
    value: {
      type: String,
      default: '',
    },
    width: {
      type: [Number, String],
      default: 500,
    },
    height: {
      type: [Number, String],
      default: 300,
    },
    lang: {
      type: String,
      default: 'text',
    },
    theme: {
      type: String,
      default: 'monokai',
    },
    readOnly: {
      type: Boolean,
      default: false,
    },
    fullScreen: {
      type: Boolean,
      default: false,
    },
    hasError: {
      type: Boolean,
      default: false,
    },
    showGutter: {
      type: Boolean,
      default: true,
    },
  },
  data() {
    return {
      $ace: null,
    };
  },
  watch: {
    value(newVal) {
      if (this.$ace && this.$ace.setValue) {
        this.$ace.setValue(newVal, 1);
        // 设置光标在第一行
        setTimeout(() => {
          this.$ace.scrollToLine(1, true, true);
        }, 0);
      }
    },
    lang(newVal) {
      if (newVal) {
        require([`brace/mode/${newVal}`], (langModule) => {
          this.$ace.getSession().setMode(`ace/mode/${newVal}`);
        });
        // import(
        //     /* webpackChunkName: 'brace-[request]' */
        //     `brace/mode/${newVal}`
        // ).then(langModule => {
        //     this.$ace.getSession().setMode(`ace/mode/${newVal}`)
        // })

        // require(`brace/mode/${newVal}`)
        // this.$ace.getSession().setMode(`ace/mode/${newVal}`)
      }
    },
    fullScreen() {
      this.$el.classList.toggle('ace-full-screen');
      this.$ace.resize();
    },
  },
  methods: {
    calcSize(size) {
      const _size = size.toString();

      if (_size.match(/^\d*$/)) return `${size}px`;
      if (_size.match(/^[0-9]?%$/)) return _size;

      return '100%';
    },
    showSearchBox() {
      this.$ace && this.$ace.execCommand('find');
    },
  },
  mounted() {
    import(
      /* webpackChunkName: 'brace' */
      'brace').then((ace) => {
      this.$ace = ace.edit(this.$el);
      const {
        $ace,
        readOnly,
      } = this;

      let {
        lang,
        theme,
      } = this;
      const session = $ace.getSession();
      lang = lang || 'javascript';
      theme = theme || 'monokai';
      this.$ace.setFontSize(14);
      this.$ace.renderer.setShowGutter(this.showGutter);
      this.$emit('init', $ace);

      // require(`brace/mode/${lang}`)
      // require('brace/mode/javascript')
      // require('brace/mode/json')
      // require('brace/mode/yaml')
      // require(`brace/theme/${theme}`)
      require('brace/ext/searchbox');
      import(
        /* webpackChunkName: 'brace-[request]' */
        `brace/mode/${lang}`).then(() => {
        require(`brace/theme/${theme}`);
        session.setMode(`ace/mode/${lang}`); // 配置语言
        $ace.setTheme(`ace/theme/${theme}`); // 配置主题
        session.setUseWrapMode(true); // 自动换行
        $ace.setValue(this.value, 1); // 设置默认内容
        $ace.setReadOnly(readOnly); // 设置是否为只读模式
        $ace.setShowPrintMargin(false); // 不显示打印边距

        // 绑定输入事件回调
        $ace.on('change', ($editor, $fn) => {
          const content = $ace.getValue();

          this.$emit('update:hasError', !content);
          this.$emit('input', content, $editor, $fn);
        });

        $ace.on('blur', ($editor, $fn) => {
          const content = $ace.getValue();

          this.$emit('update:hasError', !content);
          this.$emit('blur', content, $editor, $fn);
        });

        session.on('changeAnnotation', (args, instance) => {
          const annotations = instance.$annotations;
          if (annotations && annotations.length) {
            this.$emit('change-annotation', annotations);
          }
        });
      });
    });
  },
};
