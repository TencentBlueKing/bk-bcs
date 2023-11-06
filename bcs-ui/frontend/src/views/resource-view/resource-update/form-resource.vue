<template>
  <div class="biz-content form-resource" v-bkloading="{ isLoading }">
    <bcs-popconfirm
      class="switch-button-pop"
      :title="$t('dashboard.workload.editor.yamlMode.confirm')"
      :content="$t('dashboard.workload.editor.yamlMode.warnning')"
      width="280"
      trigger="click"
      @confirm="handleSwitchMode">
      <FixedButton position="unset" :title="$t('dashboard.workload.editor.actions.switchToYAMLMode')" />
    </bcs-popconfirm>
    <Header :title="title" />
    <div class="form-resource-content" ref="editorWrapperRef">
      <BKForm
        v-model="schemaFormData"
        ref="bkuiFormRef"
        :schema="formSchema.schema"
        :layout="formSchema.layout"
        :rules="formSchema.rules"
        :context="context"
        :http-adapter="{
          request
        }"
        form-type="vertical"
        v-show="!showDiff">
      </BKForm>
      <div class="code-diff" v-bkloading="{ isLoading: diffLoading }" v-if="showDiff">
        <div class="top-operate">
          <span class="title">
            {{ resourceName }}
            <span class="insert ml15">+{{ diffStat.insert }}</span>
            <span class="delete ml15">-{{ diffStat.delete }}</span>
          </span>
        </div>
        <CodeEditor
          :value="detail"
          :original="original"
          :height="height"
          :options="{
            renderLineHighlight: 'none'
          }"
          diff-editor
          readonly
          @diff-stat="handleDiffStatChange">
        </CodeEditor>
      </div>
    </div>

    <div class="bcs-fixed-footer">
      <template v-if="isEdit">
        <bk-button
          class="min-w-[88px]"
          theme="primary"
          v-if="!showDiff"
          @click="handleShowDiff">
          {{$t('generic.button.next')}}
        </bk-button>
        <span v-bk-tooltips.top="{ disabled: !disableUpdate, content: $t('dashboard.workload.editor.tips.contentUnchanged') }" v-else>
          <bk-button
            class="min-w-[88px]"
            theme="primary"
            :loading="loading"
            :disabled="disableUpdate"
            @click="handleSaveFormData">
            {{$t('generic.button.update')}}
          </bk-button>
        </span>
        <bk-button
          class="min-w-[88px] ml15"
          @click="handleToggleDiff">{{showDiff ? $t('dashboard.workload.editor.continueEditing') : $t('dashboard.workload.editor.showDifference')}}</bk-button>
        <bk-button class="min-w-[88px] ml15" @click="handleCancel">{{$t('generic.button.cancel')}}</bk-button>
      </template>
      <template v-else>
        <bk-button
          class="min-w-[88px]"
          theme="primary"
          :loading="loading"
          @click="handleSaveFormData">
          {{$t('generic.button.create')}}
        </bk-button>
        <bk-button class="min-w-[88px] ml15" @click="handlePreview">{{$t('generic.title.preview')}}</bk-button>
        <bk-button class="min-w-[88px] ml15" @click="handleCancel">{{$t('generic.button.cancel')}}</bk-button>
      </template>
    </div>

    <bcs-sideslider
      :is-show.sync="showSideslider"
      :title="resourceName"
      quick-close
      :width="800">
      <template #content>
        <CodeEditor
          v-full-screen="{
            tools: ['fullscreen', 'copy'],
            content: previewData
          }"
          v-bkloading="{ isLoading: previewLoading }"
          width="100%"
          height="100%"
          readonly
          :options="{
            roundedSelection: false,
            scrollBeyondLastLine: false,
            renderLineHighlight: false,
          }"
          :value="previewData">
        </CodeEditor>
      </template>
    </bcs-sideslider>
  </div>
</template>
<script>
import yamljs from 'js-yaml';

import createForm from '@blueking/bkui-form';

import FixedButton from './fixed-button.vue';

import '@blueking/bkui-form/dist/bkui-form.css';
import { CR_API_URL } from '@/api/base';
import request from '@/api/request';
import Header from '@/components/layout/Header.vue';
import CodeEditor from '@/components/monaco-editor/new-editor.vue';
import fullScreen from '@/directives/full-screen';
import $router from '@/router';
import $store from '@/store';

const BKForm = createForm({
  namespace: 'bcs',
  baseWidgets: {
    radio: 'bk-radio',
    'radio-group': 'bk-radio-group',
  },
});
export default {
  components: {
    BKForm,
    FixedButton,
    CodeEditor,
    Header,
  },
  directives: {
    'full-screen': fullScreen,
  },
  props: {
    // 命名空间（更新的时候需要--crd类型编辑是可能没有，创建的时候为空）
    namespace: {
      type: String,
      default: '',
    },
    // 父分类，eg: workloads、networks（注意复数）
    type: {
      type: String,
      default: '',
      required: true,
    },
    // 子分类，eg: deployments、ingresses
    category: {
      type: String,
      default: '',
    },
    // 名称（更新的时候需要，创建的时候为空）
    name: {
      type: String,
      default: '',
    },
    kind: {
      type: String,
      default: '',
    },
    // type 为crd时，必传
    crd: {
      type: String,
      default: '',
    },
    clusterId: {
      type: String,
      default: '',
      required: true,
    },
    // yaml模式回退到表单模式还原的数据
    formData: {
      type: Object,
      default: () => ({}),
    },
    formUpdate: {
      type: [Boolean, String],
      default: false,
    },
  },
  data() {
    return {
      schemaFormData: this.formData,
      formSchema: {},
      isLoading: false,
      loading: false,
      diffStat: {},
      detail: {},
      original: {},
      showDiff: false,
      diffLoading: false,
      height: 600,
      showSideslider: false,
      previewLoading: false,
      previewData: '',
    };
  },
  computed: {
    curProject() {
      return this.$store.state.curProject;
    },
    context() {
      return Object.assign({
        clusterID: this.clusterId,
        projectID: this.curProject.project_id,
        baseUrl: CR_API_URL,
      }, this.formSchema.context || {});
    },
    isEdit() {
      return !!this.name;
    },
    title() {
      const prefix = this.isEdit ? this.$t('generic.button.update') : this.$t('generic.button.create');
      return `${prefix} ${this.kind}`;
    },
    resourceName() {
      return this.detail?.metadata?.name || '';
    },
    disableUpdate() {
      return !Object.keys(this.diffStat).some(key => this.diffStat[key]);
    },
  },
  watch: {
    async showDiff(show) {
      if (show) {
        this.diffLoading = true;
        const detail = await this.handleGetManifestByFormData(this.schemaFormData);
        this.detail = {
          apiVersion: detail.apiVersion,
          kind: detail.kind,
          metadata: detail.metadata,
          ...detail,
        };
        this.diffLoading = false;
      }
    },
  },
  async created() {
    this.isLoading = true;
    if (this.formData && !!Object.keys(this.formData).length) {
      // 从yaml模式切换到表单模式，初始化原始数据
      const original = await this.handleGetManifestByFormData(this.formData);
      this.original = {
        apiVersion: original.apiVersion,
        kind: original.kind,
        metadata: original.metadata,
        ...original,
      };
    }
    await Promise.all([
      this.handleGetFormSchemaData(),
      this.handleGetDetail(),
    ]);
    this.isLoading = false;
  },
  mounted() {
    this.handleSetHeight();
  },
  methods: {
    handleSetHeight() {
      const bounding = this.$refs.editorWrapperRef?.getBoundingClientRect();
      this.height = bounding ? bounding.height - 80 : 600;
    },
    handleShowDiff() {
      const valid = this.$refs.bkuiFormRef?.validateForm();
      if (!valid) return;

      this.showDiff = true;
    },
    handleToggleDiff() {
      if (!this.showDiff) {
        this.handleShowDiff();
      } else {
        this.showDiff = false;
      }
    },
    handleDiffStatChange(stat) {
      this.diffStat = stat;
    },
    async handleGetManifestByFormData(formData) {
      const data = await this.$store.dispatch('dashboard/renderManifestPreview', {
        kind: this.kind,
        formData,
        $clusterId: this.clusterId,
      });
      return data;
    },
    async handleGetDetail() {
      if (!this.isEdit || (this.formData && Object.keys(this.formData).length)) return;

      let res = null;
      if (this.type === 'crd') {
        res = await this.$store.dispatch('dashboard/retrieveCustomResourceDetail', {
          $crd: this.crd,
          $category: this.category,
          $clusterId: this.clusterId,
          $name: this.name,
          namespace: this.namespace,
          format: 'formData',
        });
      } else {
        res = await this.$store.dispatch('dashboard/getResourceDetail', {
          $namespaceId: this.namespace,
          $category: this.category,
          $clusterId: this.clusterId,
          $name: this.name,
          $type: this.type,
          format: 'formData',
        });
      }
      this.schemaFormData = res.data.formData;
      const original = await this.handleGetManifestByFormData(res.data.formData);
      this.original = {
        apiVersion: original.apiVersion,
        kind: original.kind,
        metadata: original.metadata,
        ...original,
      };
    },
    async handleGetFormSchemaData() {
      this.formSchema = await this.$store.dispatch('dashboard/getFormSchema', {
        kind: this.kind,
        namespace: this.namespace,
        action: this.isEdit ? 'update' : 'create',
        $clusterId: this.clusterId,
      });
    },
    async request(url, config) {
      const requestMethods = request(config.method || 'get', url);
      const data = await requestMethods(config.params);
      return data?.selectItems || [];
    },
    handleCancel() { // 取消
      this.$bkInfo({
        type: 'warning',
        clsName: 'custom-info-confirm',
        title: this.$t('generic.msg.info.exitEdit.text'),
        subTitle: this.$t('generic.msg.info.exitEdit.subTitle'),
        defaultInfo: true,
        confirmFn: () => {
          this.$router.push({ name: this.$store.state.curSideMenu?.route });
        },
      });
    },
    // 切换Yaml模式
    async handleSwitchMode() {
      let params = {};
      if (this.isEdit) {
        params = {
          name: this.name,
        };
      } else {
        params = {
          defaultShowExample: true,
        };
      }
      this.$router.push({
        name: 'dashboardResourceUpdate',
        params: {
          ...params,
          formData: this.schemaFormData,
          defaultOriginal: this.original,
          namespace: this.namespace,
        },
        query: {
          type: this.type,
          category: this.category,
          kind: this.kind,
          crd: this.crd,
          formUpdate: this.formUpdate,
        },
      });
    },
    // 保存数据
    async handleSaveFormData() {
      this.loading = true;
      if (this.isEdit) {
        await this.handleUpdateFormResource();
      } else {
        await this.handleCreateFormResource();
      }
      this.loading = false;
    },
    // 更新表单资源
    async handleUpdateFormResource() {
      this.$bkInfo({
        type: 'warning',
        clsName: 'custom-info-confirm',
        title: this.$t('dashboard.workload.editor.dialog.confirmResourceUpdate'),
        subTitle: this.$t('dashboard.workload.editor.msg.replaceConflictWarning'),
        defaultInfo: true,
        confirmFn: async () => {
          let result = false;
          if (this.type === 'crd') {
            result = await this.$store.dispatch('dashboard/customResourceUpdate', {
              $crd: this.crd,
              $category: this.category,
              $clusterId: this.clusterId,
              $name: this.name,
              format: 'formData',
              rawData: this.schemaFormData,
              namespace: this.namespace,
            }).catch((err) => {
              console.log(err);
              return false;
            });
          } else {
            result = await this.$store.dispatch('dashboard/resourceUpdate', {
              $namespaceId: this.namespace,
              $type: this.type,
              $category: this.category,
              $clusterId: this.clusterId,
              $name: this.name,
              format: 'formData',
              rawData: this.schemaFormData,
            }).catch((err) => {
              console.log(err);
              return false;
            });
          }

          if (result) {
            this.$bkMessage({
              theme: 'success',
              message: this.$t('generic.msg.success.update'),
            });
            // 跳转回列表页
            $router.push({ name: $store.state.curSideMenu?.route });
          }
        },
      });
    },
    // 创建表单资源
    async handleCreateFormResource() {
      const valid = this.$refs.bkuiFormRef?.validateForm();
      if (!valid) return;
      let result = false;
      if (this.type === 'crd') {
        result = await this.$store.dispatch('dashboard/customResourceCreate', {
          $crd: this.crd,
          $category: this.category,
          $clusterId: this.clusterId,
          format: 'formData',
          rawData: this.schemaFormData,
        }).catch((err) => {
          console.error(err);
          return false;
        });
      } else {
        result = await this.$store.dispatch('dashboard/resourceCreate', {
          $type: this.type,
          $category: this.category,
          $clusterId: this.clusterId,
          format: 'formData',
          rawData: this.schemaFormData,
        }).catch((err) => {
          console.error(err);
          return false;
        });
      }

      if (result) {
        this.$bkMessage({
          theme: 'success',
          message: this.$t('generic.msg.success.create'),
        });
        this.$store.commit('updateCurNamespace', this.schemaFormData.metadata?.namespace);
        // 跳转回列表页
        $router.push({ name: $store.state.curSideMenu?.route });
      }
    },
    // 表单预览
    async handlePreview() {
      this.previewLoading = true;
      const detail = await this.handleGetManifestByFormData(this.schemaFormData);
      // 特殊处理-> apiVersion、kind、metadata强制排序在前三位
      this.detail = {
        apiVersion: detail.apiVersion,
        kind: detail.kind,
        metadata: detail.metadata,
        ...detail,
      };
      this.previewData = yamljs.dump(this.detail);
      this.showSideslider = true;
      this.previewLoading = false;
    },
  },
};
</script>
<style lang="postcss" scoped>
.form-resource {
    padding-bottom: 0;
    height: 100%;
    /deep/ .bk-form-radio {
        padding-left: 2px;
    }
    /deep/ .bk-sideslider .bk-sideslider-content {
        height: 100%;
    }
    .switch-button-pop {
        position: absolute;
        right: 16px;
        top: 72px;
        z-index: 1;
    }
    .icon-back {
        font-size: 16px;
        font-weight: bold;
        color: #3A84FF;
        margin-left: 20px;
        cursor: pointer;
    }
    .dashboard-top-title {
        display: inline-block;
        height: 60px;
        line-height: 60px;
        font-size: 16px;
        margin-left: 0px;
    }
    .form-resource-content {
        padding: 20px;
        max-height: calc(100vh - 162px);
        height: 100%;
        overflow: auto;
    }
    .code-diff {
        width: 100%;
        position: relative;
        .top-operate {
            display: flex;
            align-items: center;
            justify-content: space-between;
            background: #2e2e2e;
            height: 40px;
            padding: 0 10px 0 16px;
            color: #c4c6cc;
            .title {
                font-size: 14px;
            }
        }
        .status-wrapper.diff {
            height: 20%;
        }
        .insert {
            color: #5e8a48;
        }
        .delete {
            color: #e66565;
        }
    }
}
</style>
