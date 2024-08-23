<template>
  <div
    class="code-editor"
    :style="style"
    ref="editorRef"
    v-full-screen="{ tools }">
  </div>
</template>
<script lang="ts">
import yamljs from 'js-yaml';
import * as monaco from 'monaco-editor';
import { computed, defineComponent, onBeforeMount, onMounted, PropType, ref, toRefs, watch } from 'vue';

import useEditor, { IDiffValue } from './use-editor';

import { isObject } from '@/common/util';
import fullScreen from '@/directives/full-screen';

export default defineComponent({
  name: 'CodeEditor',
  directives: {
    'full-screen': fullScreen,
  },
  model: {
    prop: 'value',
    event: 'change',
  },
  props: {
    value: { type: [String, Object], default: () => ({}) },
    diffEditor: { type: Boolean, default: false }, // 是否使用diff模式
    width: { type: [String, Number], default: '100%' },
    height: { type: [String, Number], default: '100%' },
    original: { type: [String, Object], default: () => ({}) }, // 只有在diff模式下有效
    lang: { type: String, default: 'yaml' },
    theme: { type: String, default: 'vs-dark' },
    readonly: { type: Boolean, default: false },
    options: { type: Object as PropType<monaco.editor.IStandaloneEditorConstructionOptions>, default: () => ({}) },
    ignoreKeys: { type: [Array, String], default: () => '' },
    fullScreen: { type: Boolean, default: false },
    isModelValue: { type: Boolean, default: false }, // 是否开启在value变化时自动更新编辑器值（不能和v-model同时使用，不然会出现编辑器一直跳到第一行）
    multiDocument: { type: Boolean, default: false }, // 是否支持多个yaml编辑
  },
  setup(props, ctx) {
    const {
      value,
      diffEditor,
      width,
      height,
      original,
      lang,
      theme,
      options,
      readonly,
      ignoreKeys,
      fullScreen,
      isModelValue,
      multiDocument,
    } = toRefs(props);

    const tools = computed(() => {
      const data: string[] = [];
      if (fullScreen.value) {
        data.push('fullscreen');
      }
      return data;
    });
    const onContentChange = (data: IDiffValue | string = '') => {
      const emitValue = typeof data === 'string' ? data : data.modified;
      let resolveValue: string | Record<string, any> = '';
      switch (lang.value) {
        case 'yaml':{
          // 触发一次yaml格式校验（原始字符串需要校验）
          let yamlToJson = {};
          if (multiDocument.value) {
            yamlToJson = yamljs.loadAll(emitValue) || [];
          } else {
            yamlToJson = yamljs.load(emitValue) || {};
          }
          // 保留原始数据格式
          resolveValue = isObject(value.value) ? yamlToJson : emitValue;
          break;
        }
        default:
          resolveValue = emitValue;
      }
      ctx.emit('change', resolveValue, value.value);
    };
    const {
      initMonaco,
      destroyMonaco,
      updateOptions,
      setValue: handleSetValue,
      getValue,
      layout,
      getModel,
      setPosition,
      navi,
      diffStat,
      editorErr,
      editor,
    } = useEditor({
      language: lang.value,
      theme: theme.value,
      readonly: readonly.value,
      diffEditor: diffEditor.value,
      onDidChangeModelContent: onContentChange,
      onDidUpdateDiff: onContentChange,
    });

    // 转换代码为string类型
    const parseValue = (code: string | object): string => {
      if (typeof code === 'object') {
        switch (lang.value) {
          case 'yaml':
            return Object.keys(code).length ? yamljs.dump(handleIgnoreKeys(code)) : '';
          default:
            return JSON.stringify(code);
        }
      }
      return code;
    };
    // 过滤指定字段
    const handleIgnoreKeys = (data) => {
      const cloneData = JSON.parse(JSON.stringify(data));
      const keys: string[] = typeof ignoreKeys.value === 'string' ? [ignoreKeys.value] : ignoreKeys.value as string[];
      keys.forEach((key) => {
        const props = key.split('.');
        props.reduce((pre, prop, index) => {
          if (index === (props.length - 1) && pre) {
            delete pre[prop];
          }

          if (pre && (prop in pre)) {
            return pre[prop];
          }

          return '';
        }, cloneData);
      });
      return cloneData;
    };
    // 当前值
    const modifiedValue = computed<string>(() => parseValue(value.value));
    // 原始值
    const originalValue = computed<string>(() => parseValue(original.value));
    // 样式
    const style = computed(() => ({
      width: typeof width.value === 'number' || !isNaN(width.value as any) ? `${width.value}px` : width.value,
      height: typeof height.value === 'number' || !isNaN(height.value as any) ? `${height.value}px` : height.value,
    }));

    watch(options, (opt) => {
      updateOptions(opt);
    }, { deep: true });

    // 非只读模式时，建议手动调用setValue方法，watch在双向绑定时会让编辑器抖动(鼠标不断跳转到第一个位置)
    watch(
      [
        modifiedValue,
        originalValue,
      ],
      () => {
        (readonly.value || isModelValue.value) && setValue(modifiedValue.value, originalValue.value);
      },
    );

    watch([width, height], () => {
      layout();
    });

    watch(editorErr, (err) => {
      ctx.emit('error', err);
    });

    watch(diffStat, () => {
      ctx.emit('diff-stat', diffStat.value);
    });

    // 跳转下一个change
    const nextDiffChange = () => {
      navi.value?.next();
    };
    // 跳转上一个change
    const previousDiffChange = () => {
      navi?.value?.previous();
    };
    // 设置值
    const setValue = (value, original) => {
      handleSetValue(parseValue(value), parseValue(original));
    };
    // 更新值
    const update = () => {
      setTimeout(() => {
        setValue(modifiedValue.value, originalValue.value);
      });
    };

    const editorRef = ref<any>(null);
    onMounted(() => {
      initMonaco(editorRef.value, options.value);
      setValue(modifiedValue.value, originalValue.value);
      ctx.emit('init', editor.value);
    });

    onBeforeMount(() => {
      destroyMonaco();
    });

    return {
      editor,
      tools,
      diffStat,
      editorRef,
      editorErr,
      modifiedValue,
      originalValue,
      style,
      layout,
      getValue,
      setValue,
      getModel,
      update,
      nextDiffChange,
      previousDiffChange,
      setPosition,
    };
  },
});
</script>
<style scoped>
.code-editor {
    width: 100%;
    height: 100%;
    border-radius: 2px;
}
</style>
