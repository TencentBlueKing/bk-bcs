<template>
  <div class="h-full overflow-auto">
    <div class="flex items-center px-[24px] mt-[24px]">
      <!-- 基本信息 -->
      <span
        :class="[
          'inline-flex items-center justify-center h-[32px] bg-[#EAEBF0] px-[12px]',
          'text-[#313238] text-[14px] font-bold rounded-full mr-[16px]'
        ]">
        <div class="flex items-center">
          <span>{{ spaceDetail?.name }}</span>
          <i class="bcs-icon bcs-icon-angle-right font-black mx-[6px] text-[12px] text-[#c4c6cc]"></i>
          <span>{{ fileMetadata?.name }}</span>
        </div>
        <i class="bcs-icon-btn ml-[10px] bk-icon icon-edit-line text-[#979ba5]" @click="handleShowMetaDataDialog"></i>
      </span>
      <!-- 版本切换选择器 -->
      <VersionSelector
        ref="versionRef"
        class="w-[160px]"
        v-model="curVersionID"
        show-extension
        :id="id"
        @change="handleVersionChange"
        @link="showVersionList = true" />
    </div>
    <!-- 表单模式 & yaml模式 -->
    <div class="px-[24px] mt-[16px] h-[calc(100%-72px)] overflow-auto" v-if="fileStore.editMode === 'form'">
      <FormMode
        is-edit
        :value="curVersionData?.content"
        :template-space="templateSpace"
        ref="formMode" />
    </div>
    <YamlMode
      class="px-[24px] mt-[16px] h-[calc(100%-96px)]"
      is-edit
      v-else-if="fileStore.editMode === 'yaml'"
      :value="curVersionData?.content"
      :version="curVersionID"
      :render-mode="curVersionData?.renderMode"
      ref="yamlMode" />
    <!-- 元信息 -->
    <FileMetadataDialog
      :value="showMetadataDialog"
      :data="fileMetadata"
      @cancel="showMetadataDialog = false"
      @confirm="setMetadata" />
    <!-- 版本管理 -->
    <bcs-sideslider
      :is-show.sync="showVersionList"
      quick-close
      :title="`${fileMetadata?.name} ${$t('templateFile.title.versionManage')}`"
      :width="960">
      <template #content>
        <VersionList
          :template-i-d="fileMetadata?.id"
          :template-space="templateSpace"
          class="px-[24px] py-[20px]"
          @delete="refreshVersionList"
          @deleteFile="refresh" />
      </template>
    </bcs-sideslider>
  </div>
</template>
<script setup lang="ts">
import { onBeforeMount, ref, watch } from 'vue';

import VersionSelector from '../components/version-selector.vue';

import FileMetadataDialog from './file-metadata.vue';
import FormMode from './form-mode.vue';
import { store as fileStore, updateShowDeployBtn, updateTemplateMetadataList } from './use-store';
import VersionList from './version-list.vue';
import YamlMode from './yaml-mode.vue';

import { IListTemplateMetadataItem, ITemplateSpaceData, ITemplateVersionItem } from '@/@types/cluster-resource-patch';
import { ResourceService, TemplateSetService  } from '@/api/modules/new-cluster-resource';
import $bkMessage from '@/common/bkmagic';
import $i18n from '@/i18n/i18n-setup';
import $router from '@/router';

interface Props {
  templateSpace: string  // 空间
  id: string // 模板文件ID
  versionID?: string  // 模板文件版本ID
}
const props = defineProps<Props>();

const loading = ref(false);
const showVersionList = ref(false);

// 获取文件夹详情
const spaceDetail = ref<ITemplateSpaceData>();
async function getTemplateSpace() {
  if (!props.templateSpace) return;

  loading.value = true;
  spaceDetail.value = await TemplateSetService.GetTemplateSpace({
    $id: props.templateSpace,
  }).catch(() => ({}));
  loading.value = false;
}

// 获取模板文件详情数据
const fileMetadata = ref<IListTemplateMetadataItem>();
async function getTemplateMetadata() {
  if (!props.id) return;

  loading.value = true;
  fileMetadata.value = await TemplateSetService.GetTemplateMetadata({
    $id: props.id,
  }, { cancelWhenRouteChange: false }).catch(() => ({}));
  loading.value = false;
}

// 元信息修改
const showMetadataDialog = ref(false);
function handleShowMetaDataDialog() {
  showMetadataDialog.value = true;
}
async function setMetadata(data: Pick<ClusterResource.UpdateTemplateMetadataReq, 'name'|'description'>) {
  const result = await TemplateSetService.UpdateTemplateMetadata({
    $id: props.id, // 模板文件元数据 ID
    name: data.name, // 模板文件元数据名称
    description: data.description, // 模板文件元数据描述
    tags: fileMetadata.value?.tags || [], // 模板文件元数据标签(暂时不支持修改)
    version: fileMetadata.value?.version || '', // 模板文件版本(暂时不支持修改)
    versionMode: fileMetadata.value?.versionMode, // 版本规则(暂时不支持修改)
    // 保留之前草稿态
    isDraft: fileMetadata.value?.isDraft || false,
    draftVersion: fileMetadata.value?.draftVersion || '',
    draftContent: fileMetadata.value?.draftContent || '',
  }).then(() => true)
    .catch(() => false);
  if (result) {
    $bkMessage({
      theme: 'success',
      message: $i18n.t('generic.msg.success.save'),
    });
    showMetadataDialog.value = false;
    getTemplateMetadata();
    updateTemplateMetadataList(props.templateSpace);// 更新右侧列表
  }
};

// 版本信息
const curVersionID = ref(props.versionID);
const versionRef = ref<InstanceType<typeof VersionSelector>>();
const curVersionData = ref<ITemplateVersionItem>();
function refreshVersionList() {
  versionRef.value?.listTemplateVersion();
}
// 切换版本详情
async function handleVersionChange(versionData) {
  curVersionData.value = versionData;
  // if (!curVersionData.value?.content) return;

  // 更新部署按钮显示状态
  updateShowDeployBtn(!curVersionData.value?.draft);

  isCanTransform.value = await canTransform(curVersionData.value?.content || '');
  fileStore.editMode = curVersionData.value?.editFormat || 'yaml';
  fileStore.isFormModeDisabled = !isCanTransform.value;// 设置表单禁用（详情的顶部模式切换在首页，这里用全局共享数据处理）
}

// 更新路径上的版本ID参数
async function handleUpdateRouteVersionID() {
  if (curVersionID.value === props.versionID) return;
  await $router.replace({
    query: {
      versionID: curVersionID.value,
    },
  });
}

// 校验能否转化成表单
const isCanTransform = ref(false);
async function canTransform(content: string) {
  if (!content) return false;

  const data = await ResourceService.YAMLToForm({
    yaml: content,
  }, { cancelWhenRouteChange: false }).catch(() => ({ resources: [], canTransform: false }));
  return data?.canTransform;
}

// 刷新列表
function refresh() {
  showVersionList.value = false;
  $router.replace({
    name: 'templateFileList',
    params: {
      templateSpace: props.templateSpace as string,
    },
  });
  updateTemplateMetadataList(props.templateSpace);
}

watch(() => props.id, () => {
  fileStore.editMode = undefined;// 重置mode
  getTemplateMetadata();
});

watch(() => props.templateSpace, () => {
  getTemplateSpace();
});

watch(() => props.versionID, (val) => {
  if (!val) return;
  curVersionID.value = props.versionID;
});

watch(curVersionID, handleUpdateRouteVersionID);

onBeforeMount(() => {
  getTemplateSpace();
  getTemplateMetadata();
});
</script>
