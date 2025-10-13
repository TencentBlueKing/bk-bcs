<!-- eslint-disable max-len -->
<template>
  <BcsContent
    :title="isEdit ? releaseName : $t('deploy.helm.chartInstall')"
    v-bkloading="{ isLoading }"
    ref="contentRef">
    <!-- 基本信息 -->
    <div class="bcs-border min-h-[320px] bg-[#fff] flex">
      <div class="w-[400px] border-right flex flex-col items-center justify-center">
        <span class="text-[100px] text-[#979ba5]"><i class="bcs-icon bcs-icon-helm-app"></i></span>
        <span class="text-[#333c48] text-[14px] font-bold">{{ chartName }}</span>
        <div class="text-[12px] text-[#b2b5bd] mt-[8px] mb-[4px] px-[20px] text-ellipsis">
          {{ chartData.latestDescription }}
        </div>
        <bcs-popconfirm
          trigger="click"
          :confirm-text="$t('cluster.nodeList.button.copy.text')"
          v-if="isEdit"
          :disabled="!releaseDetail.notes"
          @confirm="handleCopyNotes">
          <bcs-button text size="small" :disabled="!releaseDetail.notes">{{ $t('deploy.helm.showNotes') }}</bcs-button>
          <template #content>
            <pre class="break-words whitespace-pre-wrap">{{ releaseDetail.notes }}</pre>
          </template>
        </bcs-popconfirm>
      </div>
      <div class="flex-1">
        <div class="px-[20px] h-[40px] flex items-center text-[14px] border-bottom">{{ $t('deploy.helm.args') }}</div>
        <!-- form组件实现强依赖 bk-form 名称 -->
        <bk-form
          form-type="vertical"
          class="px-[20px] py-[14px] grid grid-cols-2 gap-[20px] max-w-[800px]"
          :model="releaseData"
          :rules="rules"
          ref="formRef">
          <bk-form-item :label="$t('generic.label.name')" required property="name" error-display-type="normal">
            <bcs-input :maxlength="53" :disabled="isEdit" v-model="releaseData.name"></bcs-input>
          </bk-form-item>
          <bk-form-item :label="$t('generic.label.version')" required property="chartVersion" error-display-type="normal">
            <div class="flex items-center">
              <bcs-select class="flex-1" :loading="versionLoading" searchable :clearable="false" v-model="releaseData.chartVersion">
                <bcs-option v-for="item in versionList" :key="item.version" :id="item.version" :name="item.version">
                </bcs-option>
              </bcs-select>
              <span
                class="bcs-icon-btn flex items-center justify-center w-[32px] h-[32px] ml-[-1px]"
                style="border: 1px solid #c4c6cc"
                v-bk-tooltips="$t('generic.button.refresh')"
                @click="handleGetVersionList">
                <i class="bcs-icon bcs-icon-reset"></i>
              </span>
            </div>
          </bk-form-item>
          <bk-form-item :label="$t('generic.label.cluster1')" required property="clusterID" error-display-type="normal">
            <ClusterSelect :disabled="isEdit" searchable cluster-type="all" class="!w-[auto]" v-model="clusterID">
            </ClusterSelect>
          </bk-form-item>
          <bk-form-item :label="$t('k8s.namespace')" :desc="$t('deploy.helm.chartNSTips')" desc-type="icon" required property="namespace" error-display-type="normal">
            <NamespaceSelect :disabled="isEdit" :cluster-id="clusterID" :clearable="false" v-model="releaseData.namespace">
            </NamespaceSelect>
          </bk-form-item>
          <bk-form-item :label="$t('deploy.helm.upgradeDesc')" property="description" required error-display-type="normal" v-if="releaseName">
            <bcs-input type="textarea" v-model="args['--description']"></bcs-input>
          </bk-form-item>
        </bk-form>
      </div>
    </div>
    <!-- Values信息 -->
    <bcs-tab class="mt-[20px]" :active.sync="activeTab" v-bkloading="{ isLoading: versionDetailLoading }" v-if="showTab">
      <bcs-tab-panel name="values" :label="$t('deploy.helm.chartFlags')">
        <bcs-alert class="mb-[10px]" type="info" :title="$t('deploy.helm.defaultDesc')">
        </bcs-alert>
        <div class="flex items-center bcs-border h-[64px] bg-[#f9fbfd] px-[16px]">
          <template v-if="!isEdit || !lockValues">
            <span class="text-[14px]">{{ $t('deploy.helm.valuesFile') }}:</span>
            <bcs-select class="w-[360px] bg-[#fff] ml-[10px]" :clearable="false" searchable v-model="valuesData.valueFile">
              <bcs-option v-for="item in valuesFileList" :key="item" :id="item" :name="item">
              </bcs-option>
            </bcs-select>
            <span class="ml-[5px] mr-[10px]" v-bk-tooltips="$t('deploy.helm.values')">
              <i class="bk-icon icon-question-circle"></i>
            </span>
          </template>
          <bcs-checkbox v-model="lockValues" class="flex flex-1 release-chart-lock-checkbox" v-if="isEdit">
            <div class="flex" :title="$t('deploy.helm.defaultLockedValuesContent', { version: releaseDetail.chartVersion })">
              {{ lockValues ? $t('deploy.helm.locked') : $t('deploy.helm.unlocked') }}
              <i18n path="deploy.helm.defaultLockedValuesContent" class="flex-1 text-[#979ba5] text-[12px] bcs-ellipsis">
                <span place="version">{{ releaseDetail.chartVersion }}</span>
              </i18n>
            </div>
          </bcs-checkbox>
        </div>
        <AiEditor :value="valuesData.valueFileContent" :multi-document="false" :height="'h-[600px]'" ref="AiEditorRef" />
      </bcs-tab-panel>
      <bcs-tab-panel name="params" :label="$t('deploy.helm.helmFlags')">
        <div class="flex flex-col">
          <bcs-checkbox false-value="false" true-value="true" v-model="args['--skip-crds']">
            {{ $t('deploy.helm.skipCrds') }}
            <span class="bcs-icon-btn" v-bk-tooltips="'--skip-crds'">
              <i class="bcs-icon bcs-icon-info-circle"></i>
            </span>
          </bcs-checkbox>
          <bcs-checkbox false-value="false" true-value="true" class="mt-[10px]" v-model="args['--wait-for-jobs']">
            {{ $t('deploy.helm.waitforJobs') }}
            <span class="bcs-icon-btn" v-bk-tooltips="'--wait-for-jobs'">
              <i class="bcs-icon bcs-icon-info-circle"></i>
            </span>
          </bcs-checkbox>
          <bcs-checkbox false-value="false" true-value="true" class="mt-[10px]" v-model="args['--wait']">
            {{ $t('deploy.helm.wait') }}
            <span class="bcs-icon-btn" v-bk-tooltips="'--wait'">
              <i class="bcs-icon bcs-icon-info-circle"></i>
            </span>
          </bcs-checkbox>
          <div class="flex items-center text-[14px] mt-[10px]">
            {{ $t('deploy.helm.timeout') }}
            <bcs-input type="number" class="w-[200px] ml-[5px]" v-model="args['--timeout']">
              <template #append>
                <div class="group-text">{{ $t('units.suffix.seconds') }}</div>
              </template>
            </bcs-input>
            <span class="bcs-icon-btn ml-[5px]" v-bk-tooltips="'--timeout'">
              <i class="bcs-icon bcs-icon-info-circle"></i>
            </span>
          </div>
          <div v-if="isEdit" class="flex items-center text-[14px] mt-[10px]">
            {{ $t('deploy.helm.hisotryMax.label') }}
            <bcs-input type="number" :min="1" :max="100" class="w-[150px] ml-[5px]" v-model="args['--history-max']">
              <template #append>
                <div class="group-text">{{ $t('units.suffix.units') }}</div>
              </template>
            </bcs-input>
            <span class="bcs-icon-btn ml-[5px]" v-bk-tooltips="$t('deploy.helm.hisotryMax.desc')">
              <i class="bcs-icon bcs-icon-info-circle"></i>
            </span>
          </div>
          <!-- 自定义参数 -->
          <div class="mt-[10px]">
            <bcs-button text size="small" class="!pl-[0px]" @click="showCustomArgs = !showCustomArgs">
              {{ $t('deploy.helm.flags.label') }}
              <i :class="['bcs-icon', showCustomArgs ? 'bcs-icon-angle-double-up' : 'bcs-icon-angle-double-down']"></i>
            </bcs-button>
            <span class="bcs-icon-btn" v-bk-tooltips="$t('deploy.helm.flags.desc')">
              <i class="bcs-icon bcs-icon-info-circle"></i>
            </span>
            <KeyValue
              v-show="showCustomArgs"
              class="mt-[5px] max-w-[600px]"
              :show-header="false"
              :show-footer="false"
              :key-rules="[{
                message: $t('deploy.helm.keyTips'),
                validator: '^--'
              }]"
              :model-value="customArgs"
              :min-items="0"
              :unique-key="false"
              key-required
              ref="keyValueRef">
            </KeyValue>
          </div>
        </div>
      </bcs-tab-panel>
    </bcs-tab>
    <!-- 预览 -->
    <bcs-sideslider
      :is-show.sync="showPreview"
      quick-close :width="1000"
      :title="$t('generic.title.preview')"
      @hidden="isPreview = false">
      <template #content>
        <ChartFileTree
          :contents="previewData.newContents"
          v-bkloading="{ isLoading: previewLoading }"
          class="bcs-sideslider-content"
          style="height: calc(100vh - 100px)" />
      </template>
    </bcs-sideslider>
    <!-- diff -->
    <bcs-dialog v-model="showDiffDialog" :mask-close="false" :width="1200">
      <bcs-alert class="mb-[10px]" type="info" :title="$t('deploy.helm.flagsChange')">
      </bcs-alert>
      <div class="flex mb-[5px]">
        <span class="flex-1">{{ $t('generic.title.curVersion') }}: {{ releaseDetail.chartVersion }}</span>
        <span class="flex-1">{{ $t('deploy.helm.upgrade') }}: {{ releaseData.chartVersion }}</span>
      </div>
      <CodeEditor
        v-bkloading="{ isLoading: confirmLoading }"
        diff-editor full-screen
        :value="previewData.newContent"
        :original="previewData.oldContent"
        class="grid !min-h-[460px]"
        readonly>
      </CodeEditor>
      <template #footer>
        <bcs-button
          theme="primary"
          :disabled="!Object.keys(previewData).length"
          :loading="confirmLoading"
          @click="handleConfirmUpdate">
          {{ $t('generic.button.confirm') }}
        </bcs-button>
        <bcs-button
          :disabled="confirmLoading"
          @click="showDiffDialog = false">
          {{ $t('generic.button.cancel') }}
        </bcs-button>
      </template>
    </bcs-dialog>
    <!-- 操作 -->
    <div class="mt-[20px]">
      <bcs-button
        theme="primary"
        :loading="releaseLoading"
        :disabled="releaseLoading"
        v-authority="{
          clickable: !isEdit,
          actionId: 'namespace_scoped_update',
          resourceName: releaseData.namespace,
          disablePerms: !isEdit,
          permCtx: {
            resource_type: 'namespace',
            project_id: projectID,
            cluster_id: clusterID,
            name: releaseData.namespace
          }
        }"
        @click="handleReleaseOrUpdateChart">
        {{ isEdit ? $t('generic.button.update') : $t('deploy.helm.install') }}
      </bcs-button>
      <bcs-button :loading="releaseLoading" :disabled="releaseLoading" @click="handleShowPreview">
        {{ $t('generic.title.preview') }}
      </bcs-button>
      <bcs-button :loading="releaseLoading" :disabled="releaseLoading" @click="handleBack">
        {{ $t('generic.button.cancel') }}
      </bcs-button>
    </div>
  </BcsContent>
</template>
<script lang="ts">
import { computed, defineComponent, onMounted, ref, toRefs, watch } from 'vue';

import { filterPlainText } from '@blueking/xss-filter';

import ChartFileTree from './chart-file-tree.vue';
import useHelm from './use-helm';

import { copyText } from '@/common/util';
import $bkInfo from '@/components/bk-magic-2.0/bk-info';
import ClusterSelect from '@/components/cluster-selector/cluster-select.vue';
import KeyValue, { IData } from '@/components/key-value.vue';
import BcsContent from '@/components/layout/Content.vue';
import AiEditor from '@/components/monaco-editor/ai-editor.vue';
import CodeEditor from '@/components/monaco-editor/new-editor.vue';
import NamespaceSelect from '@/components/namespace-selector/namespace-select.vue';
import { useCluster, useProject } from '@/composables/use-app';
import $i18n from '@/i18n/i18n-setup';
import $router from '@/router';

export default defineComponent({
  name: 'ReleaseChart',
  components: {
    BcsContent,
    ClusterSelect,
    CodeEditor,
    ChartFileTree,
    NamespaceSelect,
    KeyValue,
    AiEditor,
  },
  props: {
    cluster: {
      type: String,
      default: '',
    },
    repoName: {
      type: String,
      default: '',
    },
    chartName: {
      type: String,
      default: '',
    },
    // 更新时 release 名称
    releaseName: {
      type: String,
      default: '',
    },
    // 更新时命名空间
    namespace: {
      type: String,
      default: '',
    },
  },
  async beforeRouteLeave(to, from, next) {
    const validateChange = () => new Promise((resolve, reject) => {
      if (!this.isFormChanged) {
        resolve(true);
        return;
      };
      $bkInfo({
        title: $i18n.t('generic.msg.info.exitTips.text'),
        subTitle: $i18n.t('generic.msg.info.exitTips.subTitle'),
        clsName: 'custom-info-confirm default-info',
        okText: $i18n.t('generic.button.exit'),
        cancelText: $i18n.t('generic.button.cancel'),
        confirmFn() {
          resolve(true);
        },
        cancelFn() {
          reject(false);
        },
      });
    });
    const result = await validateChange().catch(() => false);
    if (result) {
      next();
    } else {
      next(false);
    }
  },
  setup(props) {
    const { releaseName, repoName, chartName, namespace, cluster } = toRefs(props);

    const { projectID } = useProject();
    const { curClusterId } = useCluster();
    const {
      handleGetRepoChartVersions,
      handleGetRepoChartVersionDetail,
      handleReleaseChart,
      handleUpdateRelease,
      handlePreviewRelease,
      handleGetReleaseDetail,
      handleGetChartDetail,
    } = useHelm();

    const chartData = ref<Record<string, any>>({});
    const lockValues = ref(true);
    const clusterID = ref(cluster.value);
    const releaseDetail = ref<Record<string, any>>({});
    const releaseData = ref({
      name: releaseName.value,
      chartVersion: '',
      namespace: namespace.value,
    });
    const isPreview = ref(false);
    const activeTab = ref('values');
    // 表单化参数
    const args = ref({
      '--timeout': 300,
    });
    // 自定义参数
    const customArgs = ref<IData[]>([]);
    const valuesData = ref({
      valueFile: '',
      valueFileContent: '',
    });
    const rules = ref({
      name: [
        {
          required: true,
          message: $i18n.t('generic.validate.required'),
          trigger: 'blur',
        },
        {
          validator(v) {
            return /^[a-z0-9]([-a-z0-9]*[a-z0-9])?$/.test(v);
          },
          trigger: 'blur',
          message: $i18n.t('deploy.helm.regexReleaseName'),
        },
      ],
      chartVersion: [
        {
          required: true,
          message: $i18n.t('generic.validate.required'),
          trigger: 'blur',
        },
      ],
      clusterID: [
        {
          required: true,
          message: $i18n.t('generic.validate.required'),
          trigger: 'blur',
        },
      ],
      namespace: [
        {
          required: true,
          message: $i18n.t('generic.validate.required'),
          trigger: 'blur',
        },
      ],
      description: [
        {
          validator() {
            return isPreview.value || args.value['--description'];
          },
          trigger: 'blur',
          message: $i18n.t('generic.validate.required'),
        },
      ],
    });
    const showCustomArgs = ref(false);
    const formRef = ref<any>(null);
    const keyValueRef = ref<any>(null);
    const versionList = ref<any[]>([]);
    const versionDetail = ref<Record<string, any>>({});
    const valuesFileList = computed(() => versionDetail.value?.valuesFile || []);
    const showTab = computed(() => {
      const { name, chartVersion, namespace } = releaseData.value;
      return name && chartVersion && clusterID.value && namespace;
    });
    const isEdit = computed(() => !!releaseName.value && !!namespace.value);
    if (isEdit.value) {
      args.value['--history-max'] = 10;
    }

    // 获取表单数据
    const AiEditorRef = ref<InstanceType<typeof AiEditor>>();

    // 表单是否修改
    const isFormChanged = ref(false);
    watch(
      () => [
        releaseData.value,
        valuesData.value,
        AiEditorRef.value?.content,
      ],
      () => {
        isFormChanged.value = true;
      },
      { deep: true },
    );

    // 设置当前编辑器内容
    const setValuesContent = () => {
      let content = '';
      if (isEdit.value && lockValues.value) {
        // 锁定模式时始终读取当前values最后一个
        content = releaseDetail.value?.values?.[releaseDetail.value?.values.length - 1] || '';
      } else {
        // 非锁定模式时读取当前版本下values文件对应的内容
        content = versionDetail.value?.contents?.[valuesData.value.valueFile]?.content || '';
      }
      valuesData.value.valueFileContent = content;
      isFormChanged.value = false;
    };
    // 锁定当前values
    watch(lockValues, () => {
      setValuesContent();
    });

    // 获取版本列表
    const versionLoading = ref(false);
    const handleGetVersionList = async () => {
      versionLoading.value = true;
      const data = await handleGetRepoChartVersions(repoName.value, chartName.value);
      versionList.value = data.data || [];
      versionLoading.value = false;
    };

    const versionDetailLoading = ref(false);
    watch(() => releaseData.value.chartVersion, async () => {
      // 获取当前版本的values文件
      versionDetailLoading.value = true;
      versionDetail.value = await handleGetRepoChartVersionDetail(
        repoName.value,
        chartName.value,
        releaseData.value.chartVersion,
      );
      if (!valuesFileList.value?.some(name => name === valuesData.value.valueFile)) {
        valuesData.value.valueFile = valuesFileList.value?.[0];
      }
      setValuesContent();
      versionDetailLoading.value = false;
    });

    watch(() => valuesData.value.valueFile, async () => {
      setValuesContent();
    });

    // 校验数据
    const contentRef = ref<any>(null);
    const validateData = async () => {
      const validate = await formRef.value?.validate().catch(() => false);
      if (!validate) {
        // 聚集到顶部
        contentRef.value.handleScrollTop();
        return false;
      }
      const isArgsValidate = await keyValueRef.value?.validateAll();
      if (!isArgsValidate) {
        // 聚焦到helm部署参数
        activeTab.value = 'params';
        return false;
      };
      return true;
    };
    // 获取提交参数
    const handleGetReleaseParams = () => {
      if (AiEditorRef.value?.content) {
        valuesData.value.valueFileContent = AiEditorRef.value?.content;
      }
      const { namespace, name, chartVersion } = releaseData.value;
      const { valueFileContent, valueFile } = valuesData.value;
      // 处理command参数信息
      const data = Object.keys(args.value).map(key => ({
        key,
        value: args.value[key],
      }))
        .concat(keyValueRef.value?.keyValueData || []);
      const commands = data.map((item) => {
        if (item.key === '--timeout') {
          return `${item.key}=${item.value}s`;
        }
        if (item.key === '--description') {
          const result = filterPlainText(item.value);
          if (result !== item.value) {
            console.warn('Intercepted by XSS');
          }
          return `${item.key}=${result}`;
        }
        return `${item.key}=${item.value}`;
      });
      return {
        $clusterId: clusterID.value,
        $namespaceId: namespace,
        $releaseName: name,
        repository: repoName.value,
        version: chartVersion,
        chart: chartName.value,
        valueFile,
        values: [valueFileContent],
        args: commands,
      };
    };
    // 部署
    const releaseLoading = ref(false);
    const showDiffDialog = ref(false);
    const confirmLoading = ref(false);
    const handleReleaseOrUpdateChart = async () => {
      const validate = await validateData();
      if (!validate) return;

      if (isEdit.value) {
        showDiffDialog.value = true;
        // 获取diff数据
        confirmLoading.value = true;
        const params = handleGetReleaseParams();
        previewData.value = await handlePreviewRelease(params);
        confirmLoading.value = false;
      } else {
        handleConfirmRelease();
      }
    };
    const handleConfirmRelease = async () => {
      releaseLoading.value = true;
      const params = handleGetReleaseParams();
      const result = await handleReleaseChart(params);
      if (result) {
        isFormChanged.value = false;
        $router.push({
          name: 'releaseList',
        });
      }
      releaseLoading.value = false;
    };
    const handleConfirmUpdate = async () => {
      confirmLoading.value = true;
      const params = handleGetReleaseParams();
      const result = await handleUpdateRelease(params);
      if (result) {
        isFormChanged.value = false;
        $router.push({
          name: 'releaseList',
        });
      }
      confirmLoading.value = false;
    };
    // 预览
    const showPreview = ref(false);
    const previewData = ref<any>({});
    const previewLoading = ref(false);
    const handleShowPreview = async () => {
      isPreview.value = true;
      const result = await validateData();
      if (!result) return;

      showPreview.value = true;
      previewLoading.value = true;
      const params = handleGetReleaseParams();
      previewData.value = await handlePreviewRelease(params);
      previewLoading.value = false;
    };
    // 取消
    const handleBack = () => {
      $router.back();
    };

    // 表单化key
    const formArgsKey = ['--skip-crds', '--wait-for-jobs', '--wait', '--timeout', '--history-max', '--description'];
    // 详情数据
    const isLoading = ref(false);
    const handleGetDetailData = async () => {
      if (!isEdit.value) return;

      // 缓存详情数据
      isLoading.value = true;
      releaseDetail.value = await handleGetReleaseDetail(clusterID.value, namespace.value, releaseName.value);
      isLoading.value = false;
      const {
        name,
        chartVersion,
        namespace: ns,
        args: commands = [],
      } = releaseDetail.value;
      // 转换详情数据结构
      // 基本数据
      releaseData.value = {
        name,
        chartVersion,
        namespace: ns,
      };
      // 参数信息
      commands.forEach((item: string) => {
        const index = item.indexOf('=');
        const key = index > -1 ? item.slice(0, index) : item;
        const value = index > -1 ? item.slice(index + 1, item.length) : '';
        if (formArgsKey.includes(key)) {
          if (key === '--timeout') {
            // 去除单位
            args.value[key] = parseInt(value, 10);
          } else if (key !== '--description') {
            // 描述字段每次都需要更新
            args.value[key] = value;
          }
        } else {
          customArgs.value.push({
            key,
            value,
          });
        }
      });
      if (customArgs.value.length) {
        showCustomArgs.value = true;
      }
      // values文件名称
      valuesData.value.valueFile = releaseDetail.value?.valueFile;
      // values内容
      setValuesContent();
    };

    // 复制notes
    const handleCopyNotes = () => {
      copyText(releaseDetail.value.notes);
    };

    onMounted(async () => {
      handleGetVersionList();

      isLoading.value = true;
      chartData.value = await handleGetChartDetail(repoName.value, chartName.value);
      isLoading.value = false;

      isEdit.value && handleGetDetailData();
    });

    return {
      isFormChanged,
      contentRef,
      AiEditorRef,
      isPreview,
      activeTab,
      isLoading,
      chartData,
      curClusterId,
      clusterID,
      projectID,
      isEdit,
      lockValues,
      args,
      customArgs,
      showCustomArgs,
      formRef,
      keyValueRef,
      showTab,
      rules,
      releaseDetail,
      releaseData,
      valuesData,
      versionLoading,
      valuesFileList,
      versionList,
      previewLoading,
      showPreview,
      previewData,
      releaseLoading,
      showDiffDialog,
      confirmLoading,
      versionDetailLoading,
      handleShowPreview,
      handleBack,
      handleReleaseOrUpdateChart,
      handleGetVersionList,
      handleConfirmUpdate,
      handleCopyNotes,
    };
  },
});
</script>
<style lang="postcss" scoped>
.border-right {
  border-right: 1px solid #dfe0e5;
}

.border-bottom {
  border-bottom: 1px solid #dfe0e5;
}

.code-editor {
  min-height: 600px;
}

>>>.bk-form.bk-form-vertical .bk-form-item+.bk-form-item {
  margin-top: 0px;
}

>>>.release-chart-lock-checkbox .bk-checkbox-text {
  flex: 1;
}
</style>
