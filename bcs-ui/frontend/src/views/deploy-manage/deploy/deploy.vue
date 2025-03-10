<template>
  <BcsContent :title="$t('templateFile.title.deployFile')" :desc="fileMetadata?.name">
    <FormGroup class="!px-[24px]" :title="$t('templateFile.title.deploySetting')" :allow-toggle="false">
      <bk-form
        :model="formData"
        :rules="rules"
        :label-width="300"
        class="flex items-start mt-[-16px]"
        form-type="vertical"
        ref="formRef">
        <bk-form-item
          property="templateVersions"
          error-display-type="normal"
          class="flex-1 mr-[24px]"
          :label="$t('templateFile.label.fileVersion')"
          required>
          <version-selector
            v-model="formData.templateVersions"
            :id="id"
            on-draft
            @change="handleChange" />
        </bk-form-item>
        <bk-form-item
          property="namespace"
          error-display-type="normal"
          class="flex-1 !mt-0"
          :label="$t('k8s.namespace')"
          required>
          <Namespace @change="handleClusterNsChange" />
        </bk-form-item>
      </bk-form>
    </FormGroup>
    <FormGroup
      title="Values"
      class="mt-[16px] !px-[24px]"
      :allow-toggle="false"
      v-bkloading="{ isLoading: varLoading }"
      v-if="formData.templateVersions.length && formData.clusterID && formData.namespace">
      <!-- 入参 -->
      <LayoutGroup collapsible class="mt-[-16px]" v-if="varListByGroup.params.length">
        <template #title>
          {{ groupText.params }}
        </template>
        <bk-form class="grid grid-cols-2 gap-[24px] mx-[-24px]" form-type="vertical">
          <bk-form-item
            class="!mt-0 w-full"
            v-for="varData in varListByGroup.params"
            :key="varData.key"
            :label="varData.key">
            <bcs-input v-model="formData.variables[varData.key]" />
          </bk-form-item>
        </bk-form>
      </LayoutGroup>
      <!-- 其他变量 -->
      <LayoutGroup
        collapsible
        :class="varListByGroup.params.length ? 'mt-[16px]' : 'mt-[-16px]'"
        v-if="varListByGroup.readonly.length">
        <template #title>
          {{ groupText.readonly }}
        </template>
        <bk-form class="grid grid-cols-2 gap-[24px] mx-[-24px]" form-type="vertical">
          <bk-form-item
            class="!mt-0 w-full"
            v-for="varData in varListByGroup.readonly"
            :key="varData.key"
            :label="varData.key">
            <bcs-input v-model="formData.variables[varData.key]" readonly />
          </bk-form-item>
        </bk-form>
      </LayoutGroup>
      <bk-exception
        type="empty"
        scene="part"
        class="w-[300px]"
        v-if="!varListByGroup.params.length && !varListByGroup.readonly.length">
      </bk-exception>
    </FormGroup>
    <div class="flex items-center sticky bottom-0 py-[8px] mt-[8px]">
      <bcs-button
        theme="primary"
        class="min-w-[88px]"
        @click="handlePreviewData">
        {{ $t('templateFile.button.previewAndDeploy') }}
      </bcs-button>
      <bcs-button
        class="min-w-[88px]"
        @click="back">
        {{ $t('generic.button.cancel') }}
      </bcs-button>
    </div>
    <!-- 模板文件预览 -->
    <bcs-sideslider
      :is-show.sync="showPreviewSideslider"
      quick-close
      :width="1134"
      :title="$t('templateFile.title.deployReview')"
      class="sideslider-full-content">
      <template #content>
        <div class="flex h-[calc(100%-48px)]" v-bkloading="{ isLoading: previewLoading }">
          <div class="w-[200px] py-[12px] text-[12px]">
            <div
              v-for="item in previewData"
              :key="item.name"
              :class="[
                'flex items-center cursor-pointer px-[36px] h-[32px]',
                curPreviewName === item.name ? 'bg-[#E1ECFF] text-[#3A84FF]' : ''
              ]"
              @click="handleChangePreviewFile(item)">
              <span class="bcs-ellipsis" v-bk-overflow-tips>{{ item.name }}</span>
            </div>
          </div>
          <div
            :class="[
              'flex flex-col',
              'shadow-[0_2px_4px_0_rgba(0,0,0,0.16)]',
              'flex-1 w-0 bg-[#2E2E2E] h-full rounded-t-sm'
            ]"
            ref="contentRef">
            <!-- 工具栏 -->
            <div
              :class="[
                'flex items-center justify-between pl-[24px] pr-[16px] h-[40px]',
                'border-b-[1px] border-solid border-[#000]'
              ]">
              <span class="text-[#C4C6CC] text-[14px]"></span>
              <span class="flex items-center text-[12px] gap-[20px] text-[#979BA5]">
                <AiAssistant ref="assistantRef" preset="KubernetesProfessor" />
                <i
                  :class="[
                    'hover:text-[#699df4] cursor-pointer',
                    isFullscreen ? 'bcs-icon bcs-icon-zoom-out' : 'bcs-icon bcs-icon-enlarge'
                  ]"
                  @click="switchFullScreen">
                </i>
              </span>
            </div>
            <!-- 代码编辑器 -->
            <bcs-resize-layout
              placement="bottom"
              :border="false"
              :auto-minimize="true"
              :initial-divide="editorErrMsg ? 100 : 0"
              :max="300"
              :min="100"
              :disabled="!editorErrMsg"
              class="!h-0 flex-1 file-editor">
              <template #aside>
                <EditorStatus
                  :message="editorErrMsg"
                  v-show="!!editorErrMsg" />
              </template>
              <template #main>
                <CodeEditor
                  width="100%"
                  height="100%"
                  readonly
                  class="flex-1"
                  :options="{
                    roundedSelection: false,
                    scrollBeyondLastLine: false,
                    renderLineHighlight: 'none',
                  }"
                  :value="curPreviewItem?.content">
                </CodeEditor>
              </template>
            </bcs-resize-layout>
          </div>
        </div>
        <div class="bcs-border-top flex items-center h-[48px] px-[24px] bg-[#FAFBFD]">
          <bcs-button theme="primary" :disabled="!!editorErrMsg" :loading="deploying" @click="handleDeployTemplateFile">
            {{ $t('templateFile.button.confirmDeploy') }}
          </bcs-button>
          <bcs-button @click="showPreviewSideslider = false">{{ $t('generic.button.cancel') }}</bcs-button>
        </div>
      </template>
    </bcs-sideslider>
    <!-- 提示框 -->
    <bcs-dialog
      v-model="successDialog"
      theme="primary"
      width="420px"
      :close-icon="false"
      :draggable="false"
      :show-footer="false">
      <div class="flex justify-center h-[100px]">
        <svg class="icon svg-icon w-[80px]">
          <use xlink:href="#bcs-icon-color-dui"></use>
        </svg>
      </div>
      <p class="text-center text-[20px] font-bold mt-[10px]">{{ $t('templateFile.tips.success') }}</p>
      <div class="text-[16px] mt-[20px] text-center">
        {{ $t('templateFile.tips.subTitle', {
          templateFile: `${fileMetadata?.templateSpace}/${fileMetadata?.name}: ${curVersionData?.version}`,
          cluster: `${curCluster?.clusterName}(${formData?.clusterID}) ${formData?.namespace}`
        }) }}
      </div>
      <div class="flex justify-between mt-[20px] px-[30px]">
        <bcs-button
          theme="primary"
          class="min-w-[88px]"
          @click="handleGotoResourceView">
          {{ $t('templateFile.button.toResourceView') }}
        </bcs-button>
        <bcs-button
          class="min-w-[88px]"
          @click="back()">
          {{ $t('templateFile.button.toFileList') }}
        </bcs-button>
      </div>
    </bcs-dialog>
  </BcsContent>
</template>
<script setup lang="ts">
import { computed, onBeforeMount, ref, watch } from 'vue';

import versionSelector from '../components/version-selector.vue';

import Namespace from './namespace-v2.vue';

import { IListTemplateMetadataItem, IPreviewItem, ITemplateVersionItem, IVarItem } from '@/@types/cluster-resource-patch';
import { TemplateSetService  } from '@/api/modules/new-cluster-resource';
import AiAssistant from '@/components/ai-assistant.vue';
import FormGroup from '@/components/form-group.vue';
import BcsContent from '@/components/layout/Content.vue';
import CodeEditor from '@/components/monaco-editor/new-editor.vue';
import { ICluster, useCluster } from '@/composables/use-app';
import useFullScreen from '@/composables/use-fullscreen';
import $i18n from '@/i18n/i18n-setup';
import $router from '@/router';
import LayoutGroup from '@/views/cluster-manage/components/layout-group.vue';
import EditorStatus from '@/views/resource-view/resource-update/editor-status.vue';

type VarGroup = Record<keyof typeof groupText, IVarItem[]>;

interface Props {
  id: string
  version?: string
}

const props = defineProps<Props>();

const assistantRef = ref<InstanceType<typeof AiAssistant>>();
const editorErrMsg = ref('');
const deploying = ref(false);
const loading = ref(false);
const formRef = ref();
const formData = ref({
  templateVersions: '',
  variables: {},
  clusterID: '',
  namespace: '',
});
const rules = ref({
  templateVersions: [
    {
      message: $i18n.t('generic.validate.required'),
      validator() {
        return formData.value.templateVersions?.length;
      },
      trigger: 'blur',
    },
  ],
  namespace: [
    {
      message: $i18n.t('generic.validate.required'),
      validator() {
        return formData.value.clusterID?.length && formData.value.namespace?.length;
      },
      trigger: 'blur',
    },
  ],
});
const { clusterList } = useCluster();
const curCluster = computed<Partial<ICluster>>(() => clusterList.value
  ?.find(item => item.clusterID === formData.value.clusterID) || {});
const showPreviewSideslider = ref(false);
const groupText = {
  readonly: $i18n.t('templateFile.label.otherParams'), // 其他变量
  params: $i18n.t('templateFile.label.params'), // 入参
};
const previewLoading = ref(false);
const previewData = ref<IPreviewItem[]>([]);
const curPreviewName = ref('');
const curPreviewItem = computed(() => previewData.value.find(item => item.name === curPreviewName.value));

// 设置命名空间和集群
function handleClusterNsChange(clusterID: string, ns: string) {
  formData.value.clusterID = clusterID;
  formData.value.namespace = ns;
}

// 获取模板文件详情数据
const fileMetadata = ref<IListTemplateMetadataItem>();
async function getTemplateMetadata() {
  if (!props.id) return;

  loading.value = true;
  fileMetadata.value = await TemplateSetService.GetTemplateMetadata({
    $id: props.id,
  }).catch(() => ({}));
  loading.value = false;
}

// 获取模板文件的变量列表
const varLoading = ref(false);
const varListByGroup = ref<VarGroup>({
  readonly: [],
  params: [],
});
async function listTemplateFileVariables() {
  if (!formData.value.templateVersions?.length || !formData.value.clusterID || !formData.value.namespace) return;

  varLoading.value = true;
  const { vars = [] } = await TemplateSetService.ListTemplateFileVariables({
    templateVersions: Array.isArray(formData.value.templateVersions)
      ? formData.value.templateVersions
      : [formData.value.templateVersions],
    clusterID: formData.value.clusterID,
    namespace: formData.value.namespace,
  }).catch(() => ({ vars: [] }));
  varListByGroup.value = (vars as IVarItem[]).reduce<VarGroup>(
    (pre, item) => {
      if (item.readonly) {
      // 只读参数
        pre.readonly.push(item);
      } else {
      // 入参
        pre.params.push(item);
      }
      return pre;
    },
    {
      readonly: [],
      params: [],
    },
  );
  // 初始化表单数据
  formData.value.variables = vars.reduce((pre, item) => {
    pre[item.key] = item.value;
    return pre;
  }, {});
  varLoading.value = false;
}

// 部署
const curVersionData = ref<ITemplateVersionItem>();
function handleChange(versionData) {
  curVersionData.value = versionData;
}
async function handleDeployTemplateFile() {
  const validate = await formRef.value?.validate().catch(() => false);
  if (!validate) return;

  deploying.value = true;
  const result = await TemplateSetService.DeployTemplateFile({
    ...formData.value,
    templateVersions: Array.isArray(formData.value.templateVersions)
      ? formData.value.templateVersions
      : [formData.value.templateVersions],
  }).then(() => true)
    .catch(() => false);
  deploying.value = false;
  if (result) {
    successDialog.value = true;
  }
}

// 预览
async function handlePreviewData() {
  const validate = await formRef.value?.validate().catch(() => false);
  if (!validate) return;

  showPreviewSideslider.value = true;
  previewLoading.value = true;
  const { items = [], error = '' } = await TemplateSetService.PreviewTemplateFile({
    ...formData.value,
    templateVersions: Array.isArray(formData.value.templateVersions)
      ? formData.value.templateVersions
      : [formData.value.templateVersions],
  }).catch(() => ({ items: [] }));
  previewData.value = items;
  curPreviewName.value = previewData.value.at(0)?.name || '';
  if (error) {
    editorErrMsg.value = error;
    assistantRef.value?.handleSendMsg(editorErrMsg.value);
    assistantRef.value?.showAITips();// 弹出提示
  } else {
    editorErrMsg.value = '';
  }
  previewLoading.value = false;
}
async function handleChangePreviewFile(item: IPreviewItem) {
  curPreviewName.value = item.name;
}

// 全屏幕
const { contentRef, isFullscreen, switchFullScreen } = useFullScreen();

// 返回
function back() {
  $router.back();
}

const successDialog = ref(false);
function handleGotoResourceView() {
  successDialog.value = false;
  $router.push({
    name: 'dashboardWorkloadDeployments',
    params: {
      clusterId: formData.value?.clusterID,
    },
    query: {
      source: 'Template',
      namespace: formData.value.namespace,
      templateName: fileMetadata.value?.name,
      templateVersion: curVersionData.value?.version,
    },
  });
}

watch(
  () => [
    formData.value.templateVersions,
    formData.value.clusterID,
    formData.value.namespace,
  ],
  () => {
    listTemplateFileVariables();
  },
);

onBeforeMount(() => {
  getTemplateMetadata();
});
</script>
<style scoped lang="postcss">
>>> .sideslider-full-content .bk-sideslider-content {
  height: 100%;
}
/deep/ .file-editor .bk-resize-layout-aside {
  border-color: #292929;
}
</style>
