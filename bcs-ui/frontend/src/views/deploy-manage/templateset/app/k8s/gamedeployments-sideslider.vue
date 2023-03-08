<template>
  <bk-sideslider
    :is-show.sync="isVisible"
    :title="title"
    :width="width"
    :quick-close="true"
    class="gamestatefulset-sideslider"
    @hidden="hide">
    <div slot="content">
      <div class="wrapper" v-bkloading="{ isLoading: isLoading, opacity: 0.8 }">
        <monaco-editor
          v-if="!isLoading"
          ref="yamlEditor"
          class="editor"
          theme="monokai"
          language="yaml"
          :style="{ 'height': editorHeight, 'width': '100%' }"
          v-model="editorContent"
          :options="yamlEditorOptions"
          @mounted="handleEditorMount">
        </monaco-editor>
      </div>
    </div>
  </bk-sideslider>
</template>

<script>
import yamljs from 'js-yaml';

import MonacoEditor from '@/components/monaco-editor/editor.vue';

export default {
  name: 'GamestatefulsetSideslider',
  components: {
    MonacoEditor,
  },
  props: {
    isShow: {
      type: Boolean,
      default: false,
    },
    name: {
      type: String,
    },
    clusterId: {
      type: String,
    },
    namespaceName: {
      type: String,
    },
    crd: {
      type: String,
    },
  },
  data() {
    return {
      bkMessageInstance: null,
      title: '',
      width: 740,
      isVisible: false,
      isLoading: false,
      editorInstance: null,
      editorContent: '',
      yamlEditorOptions: {
        readOnly: true,
        fontSize: 14,
      },
      editorHeight: `${window.innerHeight - 60}px`,
    };
  },
  computed: {
    projectId() {
      return this.$route.params.projectId;
    },
    projectCode() {
      return this.$route.params.projectCode;
    },
    isEn() {
      return this.$store.state.isEn;
    },
  },
  watch: {
    isShow: {
      async handler(newVal) {
        this.isVisible = newVal;
        if (this.isVisible) {
          this.title = this.name;
          this.isLoading = true;
          await this.fetchData();
        }
      },
      immediate: true,
    },
  },
  mounted() {
  },
  destroyed() {
    this.bkMessageInstance?.close();
  },
  methods: {
    /**
             * 获取 yaml 数据
             */
    async fetchData() {
      try {
        const res = await this.$store.dispatch('app/getGameStatefulsetInfo', {
          projectId: this.projectId,
          clusterId: this.clusterId,
          gamestatefulsets: this.crd || 'gamedeployments.tkex.tencent.com',
          name: this.name,
          data: {
            namespace: this.namespaceName,
          },
        });
        const data = res.data || {};
        this.editorContent = yamljs.dump(data || {});
      } catch (e) {
        console.error(e);
      } finally {
        this.isLoading = false;
      }
    },

    handleEditorMount(editorInstance) {
      this.editorInstance = editorInstance;
    },

    /**
             * 隐藏 sideslider
             */
    hide() {
      this.$emit('hide-sideslider', false);
    },
  },
};
</script>

<style scoped>
    @import '../gamestatefulset-sideslider.css';
</style>
