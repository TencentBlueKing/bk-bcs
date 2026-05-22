<template>
  <BcsContent :title="$t('templateFile.title.batchDeployFile')" :desc="spaceName" :padding="16">
    <div class="flex flex-col h-full">
      <!-- 主体：左右结构 -->
      <div class="flex flex-1 min-h-0">
        <!-- 左侧：文件树选择区 -->
        <div class="flex flex-col w-[380px] mr-4 flex-shrink-0 bg-[#fff]">
          <div class="flex items-center justify-between pt-[20px] px-[16px]">
            <div>
              <span class="text-[14px] font-bold">{{ $t('templateFile.title.selectTemplateFile') }}</span>
              <span class="text-[12px] text-[#3A84FF]">
                {{ $t('templateFile.label.selectedCount', { count: selectedFileList.length }) }}
              </span>
            </div>
            <i
              v-bk-tooltips="$t('generic.button.refresh')"
              class="bcs-icon bcs-icon-refresh text-[16px] cursor-pointer hover:text-[#3a84ff] transition"
              @click="refreshFileTree"></i>
          </div>
          <div class="px-[12px] pt-[8px]">
            <bk-input
              left-icon="bk-icon icon-search"
              clearable
              :placeholder="$t('templateFile.placeholder.searchFileName')"
              v-model.trim="treeSearchKey">
            </bk-input>
          </div>
          <div class="flex-1 overflow-auto p-[12px]" v-bkloading="{ isLoading: treeLoading }">
            <FileTreeNode
              v-for="node in fileTree"
              :key="node.path"
              :node="node"
              :depth="0"
              mode="batch"
              :checked-paths="checkedPaths"
              :expanded-paths="expandedPaths"
              :search-key="treeSearchKey"
              @toggle-check="handleToggleCheck"
              @toggle-expand="handleToggleExpand"></FileTreeNode>
            <EmptyTableStatus
              v-if="!treeLoading && isTreeEmpty"
              :type="treeSearchKey && rawFileList.length ? 'search-empty' : 'empty'"
              @clear="treeSearchKey = ''">
            </EmptyTableStatus>
          </div>
        </div>
        <!-- 右侧：部署配置 + 已选文件 -->
        <div class="flex-1 overflow-auto bg-[#fff]">
          <!-- 部署设置 -->
          <FormGroup :title="$t('templateFile.title.deploySetting')" :allow-toggle="false">
            <bk-form
              :model="formData"
              :rules="rules"
              :label-width="200"
              class="flex items-start mt-[-16px]"
              form-type="vertical"
              ref="formRef">
              <bk-form-item
                property="namespace"
                error-display-type="normal"
                class="flex-1"
                :label="$t('k8s.namespace')"
                required>
                <Namespace @change="handleClusterNsChange"></Namespace>
              </bk-form-item>
            </bk-form>
          </FormGroup>
          <!-- Values 变量 -->
          <FormGroup
            title="Values"
            class="mt-[16px]"
            :allow-toggle="false"
            v-bkloading="{ isLoading: varLoading }"
            v-if="allVersionIDs.length && formData.clusterID && formData.namespace">
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
                  <bcs-input v-model="formData.variables[varData.key]"></bcs-input>
                </bk-form-item>
              </bk-form>
            </LayoutGroup>
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
                  <bcs-input v-model="formData.variables[varData.key]" readonly></bcs-input>
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
          <!-- 已选择的模板文件 -->
          <div class="mx-[16px] rounded-[2px]">
            <div class="flex items-center px-[16px] py-[12px] border border-[#DCDEE5] border-b-0">
              <span class="font-bold text-[14px]">
                {{ $t('templateFile.title.selectedFiles') }}（{{ selectedFileList.length }}）
              </span>
            </div>
            <bcs-table
              row-key="id"
              :data="selectedFileList"
              :empty-text="$t('templateFile.tips.noFilesSelected')"
              max-height="400">
              <bcs-table-column :label="$t('templateFile.title.file')" prop="name" show-overflow-tooltip>
                <template #default="{ row }">
                  {{ getFileName(row.name) }}
                </template>
              </bcs-table-column>
              <bcs-table-column :label="$t('templateFile.label.path')" prop="name" show-overflow-tooltip>
                <template #default="{ row }">
                  {{ getFilePath(row.name) }}
                </template>
              </bcs-table-column>
              <bcs-table-column :label="$t('generic.label.version')" :resizable="false" width="260">
                <template #default="{ row }">
                  <bk-select
                    :value="getSelectedVersionID(row.id)"
                    :loading="getVersionLoading(row.id)"
                    :clearable="false"
                    searchable
                    size="small"
                    @change="(val) => handleVersionChange(row.id, val)">
                    <bk-option
                      v-for="ver in getVersionList(row.id)"
                      :key="ver.versionID"
                      :id="ver.versionID"
                      :name="getVersionName(ver)">
                    </bk-option>
                  </bk-select>
                </template>
              </bcs-table-column>
              <bcs-table-column fixed="right" :resizable="false" :label="$t('generic.label.action')" width="80">
                <template #default="{ row }">
                  <span
                    class="cursor-pointer opacity-50 bcs-icon bcs-icon-minus-circle-shape"
                    @click="handleRemoveFile(row)"></span>
                </template>
              </bcs-table-column>
            </bcs-table>
          </div>
        </div>
      </div>
      <!-- Footer -->
      <div class="flex items-center pt-[8px]">
        <bcs-button
          theme="primary"
          class="min-w-[88px]"
          :disabled="!selectedFileList.length || !allVersionIDs.length"
          @click="handlePreviewData">
          {{ $t('templateFile.button.previewAndDeploy') }}
        </bcs-button>
        <bcs-button
          class="min-w-[88px]"
          @click="back">
          {{ $t('generic.button.cancel') }}
        </bcs-button>
      </div>
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
              v-for="(item, index) in previewData"
              :key="`${item.name}${index}`"
              :class="[
                'flex items-center cursor-pointer px-[36px] h-[32px]',
                curPreviewIndex === index ? 'bg-[#E1ECFF] text-[#3A84FF]' : ''
              ]"
              @click="curPreviewIndex = index">
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
            <div
              :class="[
                'flex items-center justify-between pl-[24px] pr-[16px] h-[40px]',
                'border-b-[1px] border-solid border-[#000]'
              ]">
              <span class="text-[#C4C6CC] text-[14px]"></span>
              <span class="flex items-center text-[12px] gap-[20px] text-[#979BA5]">
                <i
                  :class="[
                    'hover:text-[#699df4] cursor-pointer',
                    isFullscreen ? 'bcs-icon bcs-icon-zoom-out' : 'bcs-icon bcs-icon-enlarge'
                  ]"
                  @click="switchFullScreen">
                </i>
              </span>
            </div>
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
    <!-- 部署成功提示框 -->
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
          templateFile: `${spaceName}: ${allVersionIDs.length} ${$t('templateFile.title.files')}`,
          cluster: `${curCluster?.clusterName || ''}(${formData?.clusterID}) ${formData?.namespace}`
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

import FileTreeNode from '../template-file/file-tree-node.vue';
import { buildFileTree, filterTreeBySearch, IFileTreeNode } from '../template-file/use-file-tree';

import Namespace from './namespace-v2.vue';

import { IListTemplateMetadataItem, IMetadataVersionGroup, IMetadataVersionItem, IPreviewItem, ITemplateSpaceData, IVarItem } from '@/@types/cluster-resource-patch';
import { TemplateSetService } from '@/api/modules/new-cluster-resource';
import EmptyTableStatus from '@/components/empty-table-status.vue';
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

interface VersionData {
  versionList: IMetadataVersionItem[];
  versionLoading: boolean;
  selectedVersionID: string;
}

interface SelectedFileItem {
  id: string;
  name: string;
  templateSpace: string;
  templateSpaceID: string;
}

interface Props {
  templateSpaceID: string;
}

const props = defineProps<Props>();

const groupText = {
  readonly: $i18n.t('templateFile.label.otherParams'),
  params: $i18n.t('templateFile.label.params'),
};

// ====== 空间信息 ======
const spaceName = ref('');
const { clusterList } = useCluster();

async function getTemplateSpace() {
  if (!props.templateSpaceID) return;
  const data: Partial<ITemplateSpaceData> = await TemplateSetService.GetTemplateSpace({
    $id: props.templateSpaceID,
  }).catch(() => ({}));
  spaceName.value = data?.name || '';
}

// ====== 文件树 ======
const treeLoading = ref(false);
const fileTree = ref<IFileTreeNode[]>([]);
const rawFileList = ref<IListTemplateMetadataItem[]>([]);
const treeSearchKey = ref('');
const isTreeEmpty = computed(() => {
  if (!rawFileList.value.length) return true;
  if (!treeSearchKey.value) return !fileTree.value.length;
  return !filterTreeBySearch(fileTree.value, treeSearchKey.value).length;
});

async function loadFileTree() {
  if (!props.templateSpaceID) return;
  treeLoading.value = true;
  rawFileList.value = await TemplateSetService.ListTemplateMetadata({
    $templateSpaceID: props.templateSpaceID,
  }).catch(() => []);
  fileTree.value = buildFileTree(rawFileList.value, []);
  treeLoading.value = false;
  loadAllVersions();
}

function refreshFileTree() {
  loadFileTree();
}

// ====== Checkbox 选中逻辑 ======
const checkedPaths = ref<Set<string>>(new Set());
const expandedPaths = ref<Set<string>>(new Set());

function handleToggleExpand(path: string) {
  const newSet = new Set(expandedPaths.value);
  if (newSet.has(path)) {
    newSet.delete(path);
  } else {
    newSet.add(path);
  }
  expandedPaths.value = newSet;
}

// 从树中收集所有叶子文件节点路径
function collectLeafPaths(node: IFileTreeNode): string[] {
  if (!node.isFolder && node.file) {
    return [node.path];
  }
  const paths: string[] = [];
  for (const child of node.children) {
    paths.push(...collectLeafPaths(child));
  }
  return paths;
}

function handleToggleCheck(node: IFileTreeNode, isChecked: boolean) {
  const newSet = new Set(checkedPaths.value);
  const leafPaths = collectLeafPaths(node);
  if (isChecked) {
    leafPaths.forEach(p => newSet.add(p));
    newSet.add(node.path);
  } else {
    leafPaths.forEach(p => newSet.delete(p));
    newSet.delete(node.path);
  }
  checkedPaths.value = newSet;
}

// ====== 已选文件列表 ======
const selectedFileList = computed<SelectedFileItem[]>(() => {
  const result: SelectedFileItem[] = [];
  const addedIDs = new Set<string>();
  for (const file of rawFileList.value) {
    if (checkedPaths.value.has(file.name) && !addedIDs.has(file.id)) {
      result.push({
        id: file.id,
        name: file.name,
        templateSpace: file.templateSpace,
        templateSpaceID: file.templateSpaceID,
      });
      addedIDs.add(file.id);
    }
  }
  return result;
});

// ====== 版本数据管理 ======
const versionMap = ref<Map<string, VersionData>>(new Map());

function getVersionData(fileID: string): VersionData {
  return versionMap.value.get(fileID) || {
    versionList: [],
    versionLoading: false,
    selectedVersionID: '',
  };
}

function getVersionLoading(fileID: string): boolean {
  return getVersionData(fileID).versionLoading;
}

function getVersionList(fileID: string): IMetadataVersionItem[] {
  return getVersionData(fileID).versionList;
}

function getSelectedVersionID(fileID: string): string {
  return getVersionData(fileID).selectedVersionID;
}

function handleVersionChange(fileID: string, versionID: string) {
  const newMap = new Map(versionMap.value);
  const entry = newMap.get(fileID) || {
    versionList: [],
    versionLoading: false,
    selectedVersionID: '',
  };
  newMap.set(fileID, { ...entry, selectedVersionID: versionID });
  versionMap.value = newMap;
}

async function loadAllVersions() {
  if (!props.templateSpaceID) return;
  const list: IMetadataVersionGroup[] = await TemplateSetService.ListTemplateMetadataVersion({
    $templateID: props.templateSpaceID,
  }).catch(() => []);

  const newMap = new Map<string, VersionData>();
  for (const group of list) {
    // 通过 templateName 匹配 rawFileList 中的文件
    const file = rawFileList.value.find(f => f.name === group.templateName);
    if (!file) continue;
    const { versionList, latestVersion } = group;
    // 默认选择 latestVersion，如果找不到则选择第一个版本
    let defaultVersionID = '';
    if (versionList.length) {
      const latestVersionItem = versionList.find(v => v.version === latestVersion);
      defaultVersionID = latestVersionItem?.versionID || versionList[0].versionID;
    }
    const existingEntry = versionMap.value.get(file.id);
    newMap.set(file.id, {
      versionList,
      versionLoading: false,
      selectedVersionID: existingEntry?.selectedVersionID || defaultVersionID,
    });
  }
  versionMap.value = newMap;
}

function getVersionName(ver: IMetadataVersionItem): string {
  return ver.version;
}

// 收集所有版本 ID
const allVersionIDs = computed(() => selectedFileList.value
  .map(item => getSelectedVersionID(item.id))
  .filter(Boolean));

function getFileName(fullName: string): string {
  const segments = fullName.split('/');
  return segments[segments.length - 1] || fullName;
}

function getFilePath(fullName: string): string {
  const segments = fullName.split('/');
  if (segments.length <= 1) return '/';
  return segments.slice(0, -1).join('/');
}

function handleRemoveFile(row: SelectedFileItem) {
  const newSet = new Set(checkedPaths.value);
  const file = rawFileList.value.find(f => f.id === row.id);
  if (file) {
    newSet.delete(file.name);
  }
  checkedPaths.value = newSet;
}

// ====== 部署配置 ======
const formRef = ref();
const formData = ref<{
  variables: Record<string, string>;
  clusterID: string;
  namespace: string;
}>({
  variables: {},
  clusterID: '',
  namespace: '',
});
const rules = ref({
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

const curCluster = computed<Partial<ICluster>>(() => clusterList.value
  ?.find(item => item.clusterID === formData.value.clusterID) || {});

function handleClusterNsChange(clusterID: string, ns: string) {
  formData.value.clusterID = clusterID;
  formData.value.namespace = ns;
}

// ====== Values 变量 ======
const varLoading = ref(false);
const varListByGroup = ref<VarGroup>({
  readonly: [],
  params: [],
});

async function listTemplateFileVariables() {
  if (!allVersionIDs.value.length || !formData.value.clusterID || !formData.value.namespace) return;

  varLoading.value = true;
  const { vars = [] } = await TemplateSetService.ListTemplateFileVariables({
    templateVersions: allVersionIDs.value,
    clusterID: formData.value.clusterID,
    namespace: formData.value.namespace,
  }).catch(() => ({ vars: [] }));
  const varItems: IVarItem[] = vars;
  varListByGroup.value = varItems.reduce<VarGroup>(
    (pre, item) => {
      if (item.readonly) {
        pre.readonly.push(item);
      } else {
        pre.params.push(item);
      }
      return pre;
    },
    { readonly: [], params: [] },
  );
  formData.value.variables = varItems.reduce<Record<string, string>>((pre, item) => ({
    ...pre,
    [item.key]: item.value,
  }), {});
  varLoading.value = false;
}

watch(
  () => [allVersionIDs.value, formData.value.clusterID, formData.value.namespace],
  () => {
    listTemplateFileVariables();
  },
  { deep: true },
);

// ====== 预览 ======
const showPreviewSideslider = ref(false);
const previewLoading = ref(false);
const previewData = ref<IPreviewItem[]>([]);
const curPreviewIndex = ref(0);
const curPreviewItem = computed(() => previewData.value[curPreviewIndex.value]);
const editorErrMsg = ref('');

async function handlePreviewData() {
  const validate = await formRef.value?.validate().catch(() => false);
  if (!validate) return;

  const missingVersion = selectedFileList.value
    .some(item => !getSelectedVersionID(item.id));
  if (missingVersion) {
    editorErrMsg.value = $i18n.t('templateFile.tips.selectVersion');
    return;
  }

  showPreviewSideslider.value = true;
  previewLoading.value = true;
  const { items = [], error = '' } = await TemplateSetService.PreviewTemplateFile({
    templateVersions: allVersionIDs.value,
    variables: formData.value.variables,
    clusterID: formData.value.clusterID,
    namespace: formData.value.namespace,
  }).catch(() => ({ items: [] }));
  previewData.value = items;
  curPreviewIndex.value = 0;
  editorErrMsg.value = error || '';
  previewLoading.value = false;
}

// ====== 部署 ======
const deploying = ref(false);
async function handleDeployTemplateFile() {
  deploying.value = true;
  const result = await TemplateSetService.DeployTemplateFile({
    templateVersions: allVersionIDs.value,
    variables: formData.value.variables,
    clusterID: formData.value.clusterID,
    namespace: formData.value.namespace,
  }).then(() => true)
    .catch(() => false);
  deploying.value = false;
  if (result) {
    successDialog.value = true;
  }
}

// ====== 全屏 ======
const { contentRef, isFullscreen, switchFullScreen } = useFullScreen();

// ====== 返回 ======
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
    },
  });
}

onBeforeMount(() => {
  getTemplateSpace();
  loadFileTree();
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
