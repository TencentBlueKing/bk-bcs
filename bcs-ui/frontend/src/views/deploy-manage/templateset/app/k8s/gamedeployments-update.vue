<template>
  <bk-dialog
    :is-show.sync="isVisible"
    :width="1050"
    :title="$t('更新')"
    :close-icon="true"
    :ext-cls="'biz-rolling-update-dialog'"
    :quick-close="false"
    @confirm="confirm"
    @cancel="hide">
    <template slot="content">
      <div class="gamestatefulset-update-wrapper" v-bkloading="{ isLoading: isLoading, opacity: 1 }">
        <bk-radio class="mb5" value="1" name="method" v-model="method" checked>{{$t('原地升级')}}</bk-radio>
        <monaco-editor
          v-if="!isLoading"
          ref="jsonEditor"
          class="editor"
          theme="monokai"
          language="json"
          :style="{ 'height': '455px', 'width': '100%' }"
          v-model="editorContent"
          :options="editorOptions"
          @mounted="handleEditorMount">
        </monaco-editor>
        <div class="tip">
          <i class="bcs-icon bcs-icon-alarm-insufficient" style="font-size: 16px;"></i>
          <span>
            {{$t('此操作相当于kubectl patch gamestatefulset {name} -n {namespace} --type=json -p={文本中的内容}',
                 { name: renderItem.name, namespace: renderItem.namespace })}}</span>
        </div>
      </div>
    </template>
    <div slot="footer">
      <div class="bk-dialog-outer">
        <template v-if="isUpdating">
          <bk-button type="primary" disabled>
            {{$t('更新中...')}}
          </bk-button>
          <bk-button type="button" class="bk-dialog-btn bk-dialog-btn-cancel disabled">
            {{$t('取消')}}
          </bk-button>
        </template>
        <template v-else>
          <bk-button
            type="primary" class="bk-dialog-btn bk-dialog-btn-confirm bk-btn-primary"
            @click="confirm">
            {{$t('更新')}}
          </bk-button>
          <bk-button type="button" @click="hide">
            {{$t('取消')}}
          </bk-button>
        </template>
      </div>
    </div>
  </bk-dialog>
</template>

<script>
import MonacoEditor from '@/components/monaco-editor/editor.vue';

import { catchErrorHandler } from '@/common/util';

export default {
  components: {
    MonacoEditor,
  },
  props: {
    isShow: {
      type: Boolean,
      default: false,
    },
    item: {
      type: Object,
      default: () => ({}),
    },
  },
  data() {
    return {
      bkMessageInstance: null,
      method: '1',
      isVisible: false,
      isLoading: true,
      renderItem: {},
      editorContent: '',
      editorOptions: {
        fontSize: 14,
      },
      editorInstance: null,
      isEditorFocus: false,
      isUpdating: false,
    };
  },
  computed: {
    projectId() {
      return this.$route.params.projectId;
    },
    isEn() {
      return this.$store.state.isEn;
    },
    showPlaceholder() {
      return !this.isEditorFocus && !this.editorContent.trim();
    },
  },
  watch: {
    isShow: {
      async handler(newVal) {
        this.isVisible = newVal;
        if (!this.isVisible) {
          return;
        }
        this.renderItem = Object.assign({}, this.item || {});
        setTimeout(() => {
          this.isLoading = false;
        }, 300);
      },
      immediate: true,
    },
  },
  destroyed() {
    this.bkMessageInstance?.close();
  },
  methods: {
    handleEditorMount(editorInstance) {
      this.editorContent = `[
    {
        "op": "replace",
        "path": "/spec/template/spec/containers/0/image",
        "value": "example.com:8443/paas/public/mesos/bcs-loadbalance:v1.1.0"
    }
]
`;
      this.editorInstance = editorInstance;
      editorInstance.onDidBlurEditorText(() => {
        this.isEditorFocus = false;
      });

      editorInstance.onDidFocusEditorText(() => {
        this.isEditorFocus = true;
      });

      this.$nextTick(() => {
        this.editorInstance.focus();
        this.editorInstance.setPosition({ lineNumber: 8, column: 1 });
      });
    },

    async confirm() {
      const editorContent = this.editorContent.trim();
      if (!editorContent) {
        this.bkMessageInstance?.close();
        this.bkMessageInstance = this.$bkMessage({
          theme: 'error',
          message: this.$t('请填写内容'),
        });
        return;
      }

      let editorContentObj = null;
      try {
        editorContentObj = JSON.parse(this.editorContent);
      } catch (e) {
        console.warn(e);
        this.bkMessageInstance?.close();
        this.bkMessageInstance = this.$bkMessage({
          theme: 'error',
          message: this.$t('请填写正确的JSON格式字符串'),
        });
        return;
      }
      try {
        this.isUpdating = true;
        await this.$store.dispatch('app/updateGameStatefulsetInfo', {
          projectId: this.projectId,
          clusterId: this.renderItem.cluster_id,
          gamestatefulsets: 'gamedeployments.tkex.tencent.com',
          name: this.renderItem.name,
          data: {
            namespace: this.renderItem.namespace,
            body: editorContentObj,
            patch_type: 'application/json-patch+json',
          },
        });
        this.bkMessageInstance?.close();
        this.bkMessageInstance = this.$bkMessage({
          theme: 'success',
          message: this.$t('更新成功'),
        });
        this.$emit('update-success');
      } catch (e) {
        catchErrorHandler(e, this);
      } finally {
        this.isUpdating = false;
      }
    },

    hide() {
      this.isLoading = true;
      this.editorContent = '';
      this.isEditorFocus = false;
      this.isUpdating = false;
      this.$emit('hide-update');
    },
  },
};
</script>

<style lang="postcss">
    .biz-rolling-update-dialog {
        .gamestatefulset-update-wrapper {
            min-height: 455px;
            position: relative;
            .editor-placeholder {
                position: absolute;
                color: #d4d4d4;
                z-index: 1;
                top: 35px;
                left: 67px;
                font-size: 14px;
                pre {
                    margin: 0;
                    padding: 0;
                }
            }
            .tip {
                padding: 5px;
                margin-top: 10px;
                background-color: #F2F2F2;
                font-size: 14px;
            }
        }
    }
</style>
