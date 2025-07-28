<template>
  <div :class="['bcs-md-preview markdown-body', theme]" v-html="filterXSS"></div>
</template>

<script>
import MarkdownIt from 'markdown-it';
import { computed, defineComponent, onMounted, ref, toRefs, watch } from 'vue';
import xss from 'xss';

import hljs from './md-highlight.js';

export default defineComponent({
  name: 'BCSMd',
  props: {
    theme: {
      type: String,
      default: 'light',
    },
    code: {
      type: String,
      default: '',
    },
    linkTarget: {
      type: String,
      default: '_blank',
    },
  },
  setup(props) {
    const { code, linkTarget } = toRefs(props);
    const html = ref(null);
    const md = new MarkdownIt({
      highlight(str, lang) {
        if (lang && hljs.getLanguage(lang)) {
          try {
            return `<pre class="hljs"><code>${hljs.highlight(lang, str).value}</code></pre>`;
          } catch {}
        }
        return `<pre class="bcs-default-md-hljs"><code>${md.utils.escapeHtml(str)}</code></pre>`;
      },
    });
    const render = (value) => {
      html.value = md.render(value);
    };
    const filterXSS = computed(() => xss(html.value));
    watch(code, () => {
      render(code.value);
    });
    // create render rules
    const defaultRender = md.renderer.rules.link_open || function (tokens, idx, options, env, self) {
      return self.renderToken(tokens, idx, options);
    };
    md.renderer.rules.link_open = function (tokens, idx, options, env, self) {
      const aIndex = tokens[idx].attrIndex('target');

      if (aIndex < 0) {
        tokens[idx].attrPush(['target', linkTarget.value]);
      } else {
        tokens[idx].attrs[aIndex][1] = linkTarget.value;
      }
      return defaultRender(tokens, idx, options, env, self);
    };
    onMounted(() => {
      render(code.value);
    });
    return {
      filterXSS,
    };
  },
});
</script>

<style lang="postcss">
@import './github-md-base.css';
@import './github-md-theme.css';
.bcs-default-md-hljs {
    color: #fff;
}
</style>
