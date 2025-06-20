<!-- eslint-disable vue/multi-word-component-names -->
<template>
  <BcsContent :title="curApp.name">
    <div>
      <div>
        <div class="biz-crd-header">
          <div class="left">
            <svg style="display: none;">
              <title>{{$t('deploy.templateset.icon')}}</title>
              <symbol id="biz-set-icon" viewBox="0 0 32 32">
                <path d="M6 3v3h-3v23h23v-3h3v-23h-23zM24 24v3h-19v-19h19v16zM27 24h-1v-18h-18v-1h19v19z"></path>
                <path d="M13.688 18.313h-6v6h6v-6z"></path>
                <path d="M21.313 10.688h-6v13.625h6v-13.625z"></path>
                <path d="M13.688 10.688h-6v6h6v-6z"></path>
              </symbol>
            </svg>
            <div class="info">
              <svg class="logo" style="cursor: pointer;">
                <use xlink:href="#biz-set-icon"></use>
              </svg>
              <div class="desc" :title="curApp.description">
                <span>{{$t('plugin.tools.intro')}}：</span>
                {{curApp.description || '--'}}
              </div>
            </div>
          </div>

          <div class="right">
            <div class="bk-collapse-item bk-collapse-item-active">
              <div class="biz-item-header" style="cursor: default;">
                {{$t('deploy.helm.args')}}
              </div>
              <div class="bk-collapse-item-content f12" style="padding: 15px;">
                <div class="config-box" style="min-width: 580px;">
                  <div class="inner">
                    <div class="inner-item">
                      <label class="title">{{$t('generic.label.name')}}</label>
                      <bkbcs-input :value="curApp.releaseName" :disabled="true" style="width: 250px;" />
                    </div>

                    <div class="inner-item">
                      <label class="title">{{$t('generic.label.version')}}</label>
                      <div>
                        <bcs-select v-model="curApp.version" style="width: 250px;">
                          <bcs-option
                            v-for="(opt, index) in chartVersionsList"
                            :key="index"
                            :name="opt.version"
                            :id="opt.version"
                          ></bcs-option>
                        </bcs-select>
                      </div>
                    </div>
                  </div>
                  <div class="inner">
                    <div class="inner-item">
                      <label class="title">{{$t('generic.label.cluster1')}}</label>
                      <bkbcs-input :value="curApp.cluster_id" :disabled="true" style="width: 250px;" />
                    </div>

                    <div class="inner-item">
                      <label class="title">{{$t('k8s.namespace')}}</label>
                      <div>
                        <bkbcs-input v-model="curApp.namespace" style="width: 250px;" disabled></bkbcs-input>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>

        <div class="action-box">
          <div class="title mb10">
            Values {{$t('plugin.tools.content')}}
          </div>
        </div>

        <div class="editor-box">
          <monaco-editor
            ref="yamlEditor"
            class="editor"
            theme="vs-dark"
            language="yaml"
            :style="{ height: `${editorHeight}px`, width: '100%' }"
            v-model="editorOptions.content"
            :diff-editor="editorOptions.isDiff"
            :options="editorOptions"
            :original="editorOptions.originContent">
          </monaco-editor>
        </div>

        <div class="create-wrapper">
          <bk-button type="primary" :title="$t('generic.button.update')" @click="handleUpdate">
            {{$t('generic.button.update')}}
          </bk-button>
          <bk-button @click="handleShowPreview">
            {{ $t('generic.title.preview') }}
          </bk-button>
          <bk-button type="default" :title="$t('generic.button.cancel')" @click="goBack">
            {{$t('generic.button.cancel')}}
          </bk-button>
        </div>
      </div>
    </div>

    <bk-dialog
      :width="1100"
      :title="updateConfirmDialog.title"
      :close-icon="!updateInstanceLoading"
      :is-show.sync="updateConfirmDialog.isShow"
      @cancel="hideConfirmDialog">
      <template slot="content">
        <p
          class="biz-tip mb5 tl" style="color: #666;"
          v-if="yamlDiffEditorOptions.isDiff">{{$t('plugin.tools.confirmUpgradeTips')}}</p>
        <div class="difference-code">
          <div class="editor-header" v-if="yamlDiffEditorOptions.isDiff">
            <div>{{ $t('generic.title.curVersion') }}</div>
            <div>{{ $t('deploy.helm.upgrade') }}</div>
          </div>

          <div
            :class="['diff-editor-box', { 'editor-fullscreen': yamlDiffEditorOptions.fullScreen }]"
            style="position: relative;">
            <monaco-editor
              ref="yamlEditor"
              class="editor"
              theme="vs-dark"
              language="yaml"
              :style="{ height: `${diffEditorHeight}px`, width: '100%' }"
              v-model="curAppDifference.content"
              :diff-editor="yamlDiffEditorOptions.isDiff"
              :key="differenceKey"
              :options="yamlDiffEditorOptions"
              :original="curAppDifference.originContent">
            </monaco-editor>
          </div>
        </div>
      </template>
      <div slot="footer">
        <div class="bk-dialog-outer mt10">
          <template>
            <bk-button
              :class="['bk-button bk-dialog-btn-confirm bk-primary', { 'is-disabled': updateInstanceLoading }]"
              @click="updateCrdController">
              {{updateInstanceLoading ? $t('generic.status.updating') : $t('generic.button.confirm')}}
            </bk-button>
            <bk-button
              :class="['bk-button bk-dialog-btn-cancel bk-default']"
              @click="hideConfirmDialog">
              {{$t('generic.button.cancel')}}
            </bk-button>
          </template>
        </div>
      </div>
    </bk-dialog>
    <!-- 预览 -->
    <bcs-sideslider
      quick-close
      :is-show.sync="showPreview"
      :width="1000"
      :title="$t('generic.title.preview')">
      <template #content>
        <ChartFileTree
          :contents="previewData.newContents"
          v-bkloading="{ isLoading: previewLoading }"
          class="bcs-sideslider-content"
          style="height: calc(100vh - 100px)" />
      </template>
    </bcs-sideslider>
  </BcsContent>
</template>

<script>
import { addonsDetail, addonsPreview, updateOns } from '@/api/modules/helm';
import { catchErrorHandler } from '@/common/util';
import BcsContent from '@/components/layout/Content.vue';
import MonacoEditor from '@/components/monaco-editor/editor.vue';
import ChartFileTree from '@/views/deploy-manage/helm/chart-file-tree.vue';
import useHelm from '@/views/deploy-manage/helm/use-helm';

export default {
  components: {
    MonacoEditor,
    BcsContent,
    ChartFileTree,
  },
  data() {
    return {
      editorOptions: {
        readOnly: false,
        fontSize: 14,
        fullScreen: false,
        content: '',
        originContent: '',
        isDiff: false,
      },
      curAppDifference: {
        content: '',
        originContent: '',
      },
      updateInstanceLoading: false,
      differenceKey: 0,
      yamlDiffEditorOptions: {
        readOnly: true,
        fontSize: 14,
        fullScreen: false,
        isDiff: false,
        ignoreTrimWhitespace: false,
      },
      updateConfirmDialog: {
        title: this.$t('plugin.tools.confirmUpgrade'),
        isShow: false,
        width: 1060,
        height: 350,
        lang: 'yaml',
        closeIcon: true,
        readOnly: true,
        fullScreen: false,
        values: [],
        editors: [],
      },
      diffEditorHeight: 350,
      editorHeight: 500,
      curApp: {
        namespace: '',
      },
      namespaceList: [],
      chartVersionsList: [],
      showPreview: false,
      previewData: {},
      previewLoading: false,
    };
  },
  computed: {
    curProject() {
      return this.$store.state.curProject;
    },
    projectId() {
      return this.$route.params.projectId;
    },
  },

  created() {
    this.curClusterId = this.$route.params.clusterId;
    this.curCrdId = this.$route.params.id;
    this.chartName = this.$route.params.chartName;

    this.getCommonCrdInstanceDetail();
    this.fetchChartVersionsList();
  },
  methods: {
    async getCommonCrdInstanceDetail() {
      try {
        const clusterId = this.curClusterId;
        const crdId = this.curCrdId;
        const item = await addonsDetail({
          $clusterId: clusterId,
          $name: crdId,
        });
        const data = {
          ...item,
          cluster_id: clusterId,
          values: item.currentValues,
        };
        this.curApp = data;
        this.editorOptions.content = data.values || '';
        this.editorOptions.originContent = data.values;
      } catch (e) {
        catchErrorHandler(e, this);
      }
    },
    async fetchChartVersionsList() {
      const { handleGetRepoChartVersions } = useHelm();
      const res = await handleGetRepoChartVersions('public-repo', this.chartName);
      this.chartVersionsList = res.data;
    },

    goBack() {
      this.$router.back();
    },

    async handleUpdate() {
      await this.getPreviewData();
      // this.curAppDifference.content = this.editorOptions.content;
      // this.curAppDifference.originContent = this.editorOptions.originContent;
      this.curAppDifference.content = this.previewData.newContent;
      this.curAppDifference.originContent = this.previewData.oldContent;
      if (this.curAppDifference.content === this.curAppDifference.originContent) {
        this.curAppDifference.content = this.$t('plugin.tools.noChages');
        this.yamlDiffEditorOptions.isDiff = false;
      } else {
        this.yamlDiffEditorOptions.isDiff = true;
      }
      this.updateConfirmDialog.isShow = true;
      setTimeout(() => {
        // eslint-disable-next-line no-plusplus
        this.differenceKey++;
      }, 0);
    },

    checkData() {
      if (!this.editorOptions.content) {
        this.$bkMessage({
          theme: 'error',
          message: this.$t('plugin.tools.input.values'),
        });
        return false;
      }
      return true;
    },

    hideConfirmDialog() {
      if (this.updateInstanceLoading) {
        return false;
      }
      this.updateConfirmDialog.isShow = false;
    },

    async updateCrdController() {
      if (this.updateInstanceLoading) {
        return false;
      }

      this.updateInstanceLoading = true;
      try {
        const clusterId = this.curClusterId;
        const result = await updateOns({
          $clusterId: clusterId,
          $name: this.curApp.name,
          values: this.editorOptions.content || '',
          version: this.curApp.version,
        }).then(() => true)
          .catch(() => false);
        if (result) {
          this.$bkMessage({
            theme: 'success',
            message: this.$t('plugin.tools.submit'),
          });
          this.goBack();
        }
      } catch (e) {
        catchErrorHandler(e, this);
      } finally {
        this.updateInstanceLoading = false;
      }
    },

    async handleShowPreview() {
      this.showPreview = true;
      this.previewLoading = true;
      await this.getPreviewData();
      this.previewLoading = false;
    },

    async getPreviewData() {
      this.previewData = await addonsPreview({
        $clusterId: this.curClusterId,
        $name: this.curApp.name,
        values: this.editorOptions.content || '',
        version: this.curApp.version,
      }).catch(() => ({
        newContent: '',
        newContents: {},
        oldContent: '',
        oldContents: {},
      }));
    },
  },
};
</script>

<style scoped>
    @import './detail.css';
</style>
