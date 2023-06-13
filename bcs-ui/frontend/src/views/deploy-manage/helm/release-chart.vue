<!-- eslint-disable max-len -->
<template>
  <BcsContent :title="isEdit ? releaseName : $t('Chart部署')" v-bkloading="{ isLoading }" ref="contentRef">
    <!-- 基本信息 -->
    <div class="bcs-border min-h-[320px] bg-[#fff] flex">
      <div class="w-[400px] border-right flex flex-col items-center justify-center">
        <span class="text-[100px] text-[#979ba5]"><i class="bcs-icon bcs-icon-helm-app"></i></span>
        <span class="text-[#333c48] text-[14px] font-bold">{{chartName}}</span>
        <div
          class="text-[12px] text-[#b2b5bd] mt-[8px] mb-[4px] px-[20px] text-ellipsis">
          {{chartData.latestDescription}}
        </div>
        <bcs-popconfirm
          trigger="click"
          :confirm-text="$t('复制')"
          v-if="isEdit"
          :disabled="!releaseDetail.notes"
          @confirm="handleCopyNotes">
          <bcs-button text size="small" :disabled="!releaseDetail.notes">{{$t('查看Notes')}}</bcs-button>
          <template #content>
            <div class="break-all">{{releaseDetail.notes}}</div>
          </template>
        </bcs-popconfirm>
      </div>
      <div class="flex-1">
        <div class="px-[20px] h-[40px] flex items-center text-[14px] border-bottom">{{$t('配置选项')}}</div>
        <!-- form组件实现强依赖 bk-form 名称 -->
        <bk-form
          form-type="vertical"
          class="px-[20px] py-[14px] grid grid-cols-2 gap-[20px] max-w-[800px]"
          :model="releaseData"
          :rules="rules"
          ref="formRef">
          <bk-form-item :label="$t('名称')" required property="name" error-display-type="normal">
            <bcs-input :maxlength="53" :disabled="isEdit" v-model="releaseData.name"></bcs-input>
          </bk-form-item>
          <bk-form-item :label="$t('版本')" required property="chartVersion" error-display-type="normal">
            <div class="flex items-center">
              <bcs-select
                class="flex-1"
                :loading="versionLoading"
                searchable
                :clearable="false"
                v-model="releaseData.chartVersion">
                <bcs-option
                  v-for="item in versionList"
                  :key="item.version"
                  :id="item.version"
                  :name="item.version">
                </bcs-option>
              </bcs-select>
              <span
                class="bcs-icon-btn flex items-center justify-center w-[32px] h-[32px] ml-[-1px]"
                style="border: 1px solid #c4c6cc"
                v-bk-tooltips="$t('刷新列表')"
                @click="handleGetVersionList">
                <i class="bcs-icon bcs-icon-reset"></i>
              </span>
            </div>
          </bk-form-item>
          <bk-form-item :label="$t('所属集群')" required property="clusterID" error-display-type="normal">
            <ClusterSelect
              :disabled="isEdit"
              searchable
              cluster-type="all"
              class="!w-[auto]"
              v-model="clusterID">
            </ClusterSelect>
          </bk-form-item>
          <bk-form-item
            :label="$t('命名空间')"
            :desc="$t('如果Chart中已经配置命名空间，则会使用Chart中的命名空间，会导致不匹配等问题;建议Chart中不要配置命名空间')"
            desc-type="icon"
            required
            property="namespace"
            error-display-type="normal">
            <NamespaceSelect
              :disabled="isEdit"
              :cluster-id="clusterID"
              :clearable="false"
              v-model="releaseData.namespace">
            </NamespaceSelect>
          </bk-form-item>
          <bk-form-item
            :label="$t('更新说明')"
            property="description"
            required
            error-display-type="normal"
            v-if="releaseName">
            <bcs-input type="textarea" v-model="args['--description']"></bcs-input>
          </bk-form-item>
        </bk-form>
      </div>
    </div>
    <!-- Values信息 -->
    <bcs-tab
      class="mt-[20px]"
      :active.sync="activeTab"
      v-bkloading="{ isLoading: versionDetailLoading }"
      v-if="showTab">
      <bcs-tab-panel name="values" :label="$t('Chart部署选项')">
        <bcs-alert
          class="mb-[10px]"
          type="info"
          :title="$t('YAML初始值为创建时Chart中values.yaml内容，后续更新部署以该YAML内容为准，YAML内容最终通过`--values`选项传递给`helm template`命令')">
        </bcs-alert>
        <div class="flex items-center bcs-border h-[64px] bg-[#f9fbfd] px-[16px]">
          <template v-if="!isEdit || !lockValues">
            <span class="text-[14px]">{{$t('Values文件')}}:</span>
            <bcs-select
              class="w-[360px] bg-[#fff] ml-[10px]"
              :clearable="false"
              searchable
              v-model="valuesData.valueFile">
              <bcs-option
                v-for="item in valuesFileList"
                :key="item"
                :id="item"
                :name="item">
              </bcs-option>
            </bcs-select>
            <span
              class="ml-[5px] mr-[10px]"
              v-bk-tooltips="$t('Values文件包含两类: <br/>- 以values.yaml或以values.yml结尾, 例如xxx-values.yaml文件 <br/>- bcs-values目录下的以.yml或.yaml结尾的文件')">
              <i class="bk-icon icon-question-circle"></i>
            </span>
          </template>
          <bcs-checkbox v-model="lockValues" class="flex flex-1 release-chart-lock-checkbox" v-if="isEdit">
            <div
              class="flex"
              :title="$t('(默认锁定values内容为当前release(版本: {version} )的内容, 解除锁定后, 加载为对应Chart中的values内容)', { version: releaseDetail.chartVersion })">
              {{lockValues ? $t('已锁定') : $t('已解锁')}}
              <i18n
                path="(默认锁定values内容为当前release(版本: {version} )的内容, 解除锁定后, 加载为对应Chart中的values内容)"
                class="flex-1 text-[#979ba5] text-[12px] bcs-ellipsis">
                <span place="version">{{releaseDetail.chartVersion}}</span>
              </i18n>
            </div>
          </bcs-checkbox>
        </div>
        <CodeEditor
          ref="codeEditorRef"
          class="code-editor"
          full-screen
          v-model="valuesData.valueFileContent">
        </CodeEditor>
      </bcs-tab-panel>
      <bcs-tab-panel name="params" :label="$t('Helm部署参数')">
        <div class="flex flex-col">
          <bcs-checkbox false-value="false" true-value="true" v-model="args['--skip-crds']">
            {{$t('忽略CRD')}}
            <span class="bcs-icon-btn" v-bk-tooltips="'--skip-crds'">
              <i class="bcs-icon bcs-icon-info-circle"></i>
            </span>
          </bcs-checkbox>
          <bcs-checkbox false-value="false" true-value="true" class="mt-[10px]" v-model="args['--wait-for-jobs']">
            {{$t('等待所有Jobs完成')}}
            <span class="bcs-icon-btn" v-bk-tooltips="'--wait-for-jobs'">
              <i class="bcs-icon bcs-icon-info-circle"></i>
            </span>
          </bcs-checkbox>
          <bcs-checkbox false-value="false" true-value="true" class="mt-[10px]" v-model="args['--wait']">
            {{$t('等待所有Pod，PVC处于ready状态')}}
            <span class="bcs-icon-btn" v-bk-tooltips="'--wait'">
              <i class="bcs-icon bcs-icon-info-circle"></i>
            </span>
          </bcs-checkbox>
          <div class="flex items-center text-[14px] mt-[10px]">
            {{$t('超时时间')}}
            <bcs-input type="number" class="w-[200px] ml-[5px]" v-model="args['--timeout']">
              <template #append>
                <div class="group-text">{{$t('秒')}}</div>
              </template>
            </bcs-input>
            <span class="bcs-icon-btn ml-[5px]" v-bk-tooltips="'--timeout'">
              <i class="bcs-icon bcs-icon-info-circle"></i>
            </span>
          </div>
          <div v-if="isEdit" class="flex items-center text-[14px] mt-[10px]">
            {{$t('保留最大历史版本')}}
            <bcs-input
              type="number"
              :min="1"
              :max="100"
              class="w-[150px] ml-[5px]"
              v-model="args['--history-max']">
              <template #append>
                <div class="group-text">{{$t('个')}}</div>
              </template>
            </bcs-input>
            <span class="bcs-icon-btn ml-[5px]" v-bk-tooltips="$t('--history-max，可用于回滚的历史版本个数')">
              <i class="bcs-icon bcs-icon-info-circle"></i>
            </span>
          </div>
          <!-- 自定义参数 -->
          <div class="mt-[10px]">
            <bcs-button text size="small" class="!pl-[0px]" @click="showCustomArgs = !showCustomArgs">
              {{$t('自定义参数')}}
              <i :class="['bcs-icon', showCustomArgs ? 'bcs-icon-angle-double-up' : 'bcs-icon-angle-double-down']"></i>
            </bcs-button>
            <span class="bcs-icon-btn" v-bk-tooltips="$t('设置Flags，如设置wait，输入格式为 --wait = true')">
              <i class="bcs-icon bcs-icon-info-circle"></i>
            </span>
            <KeyValue
              v-show="showCustomArgs"
              class="mt-[5px] max-w-[600px]"
              :show-header="false"
              :show-footer="false"
              :key-rules="[{
                message: $t('参数Key必须由 -- 字符开头'),
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
      quick-close
      :width="1000"
      :title="$t('预览')"
      @hidden="isPreview = false">
      <template #content>
        <ChartFileTree
          :contents="previewData.newContents"
          v-bkloading="{ isLoading: previewLoading }"
          class="bcs-sideslider-content"
          style="height: calc(100vh - 100px)"
        />
      </template>
    </bcs-sideslider>
    <!-- diff -->
    <bcs-dialog
      v-model="showDiffDialog"
      :mask-close="false"
      :width="1200">
      <bcs-alert
        class="mb-[10px]"
        type="info"
        :title="$t('Helm Release参数发生如下变化，请确认后再点击“确定”更新')">
      </bcs-alert>
      <div class="flex mb-[5px]">
        <span class="flex-1">{{$t('当前版本')}}: {{releaseDetail.chartVersion}}</span>
        <span class="flex-1">{{$t('更新版本')}}: {{releaseData.chartVersion}}</span>
      </div>
      <CodeEditor
        v-bkloading="{ isLoading: confirmLoading }"
        diff-editor
        full-screen
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
          @click="handleConfirmUpdate">{{$t('确定')}}</bcs-button>
        <bcs-button :disabled="confirmLoading" @click="showDiffDialog = false">{{$t('取消')}}</bcs-button>
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
        {{isEdit ? $t('更新') : $t('部署')}}
      </bcs-button>
      <bcs-button
        :loading="releaseLoading"
        :disabled="releaseLoading"
        @click="handleShowPreview">
        {{$t('预览')}}
      </bcs-button>
      <bcs-button
        :loading="releaseLoading"
        :disabled="releaseLoading"
        @click="handleBack">
        {{$t('取消')}}
      </bcs-button>
    </div>
  </BcsContent>
</template>
<script lang="ts">
import { defineComponent, onMounted, toRefs, ref, watch, computed } from 'vue';
import BcsContent from '@/components/layout/Content.vue';
import ClusterSelect from '@/components/cluster-selector/cluster-select.vue';
import CodeEditor from '@/components/monaco-editor/new-editor.vue';
import ChartFileTree from './chart-file-tree.vue';
import NamespaceSelect from '@/components/namespace-selector/namespace-select.vue';
import KeyValue, { IData } from '@/components/key-value.vue';
import useHelm from './use-helm';
import { useCluster, useProject } from '@/composables/use-app';
import $router from '@/router';
import $i18n from '@/i18n/i18n-setup';
import { copyText } from '@/common/util';

export default defineComponent({
  name: 'ReleaseChart',
  components: {
    BcsContent,
    ClusterSelect,
    CodeEditor,
    ChartFileTree,
    NamespaceSelect,
    KeyValue,
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
      '--timeout': 600,
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
          message: $i18n.t('必填项'),
          trigger: 'blur',
        },
        {
          validator(v) {
            return /^[a-z0-9]([-a-z0-9]*[a-z0-9])?$/.test(v);
          },
          trigger: 'blur',
          message: $i18n.t('Release名称只能由小写字母数字或者-组成'),
        },
      ],
      chartVersion: [
        {
          required: true,
          message: $i18n.t('必填项'),
          trigger: 'blur',
        },
      ],
      clusterID: [
        {
          required: true,
          message: $i18n.t('必填项'),
          trigger: 'blur',
        },
      ],
      namespace: [
        {
          required: true,
          message: $i18n.t('必填项'),
          trigger: 'blur',
        },
      ],
      description: [
        {
          validator() {
            return isPreview.value || args.value['--description'];
          },
          trigger: 'blur',
          message: $i18n.t('必填项'),
        },
      ],
    });
    const showCustomArgs = ref(false);
    const formRef = ref<any>(null);
    const keyValueRef = ref<any>(null);
    const codeEditorRef = ref<any>(null);
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
      codeEditorRef.value?.setValue(valuesData.value.valueFileContent);
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
        contentRef.value.handleScollTop();
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
      result && $router.push({
        name: 'releaseList',
      });
      releaseLoading.value = false;
    };
    const handleConfirmUpdate = async () => {
      confirmLoading.value = true;
      const params = handleGetReleaseParams();
      const result = await handleUpdateRelease(params);
      result && $router.push({
        name: 'releaseList',
      });
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
      } =  releaseDetail.value;
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
      contentRef,
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
      codeEditorRef,
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
>>> .bk-form.bk-form-vertical .bk-form-item+.bk-form-item {
  margin-top: 0px;
}
>>> .release-chart-lock-checkbox .bk-checkbox-text {
  flex: 1;
}
</style>
