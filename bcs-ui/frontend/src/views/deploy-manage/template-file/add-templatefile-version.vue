<template>
  <BcsContent :title="fileMetadata?.name" :padding="0" v-bkloading="{ isLoading }">
    <template #title>
      <i class="bcs-icon bcs-icon-arrows-left back mr-[4px]" @click="back"></i>
      <span>{{ fileMetadata?.name }}</span>
      <span class="bcs-icon-btn ml-[10px]" @click="handleShowMetadataDialog">
        <i class="bk-icon icon-edit-line"></i>
      </span>
      <bk-tag v-if="fileMetadata?.isDraft" class="ml-4" type="stroke">
        {{ $t('templateFile.tag.draft')
          + ' ( ' + $t('templateFile.tag.baseOn')
          + (fileMetadata?.draftVersion || '--') + ' ) '}}</bk-tag>
    </template>
    <template #header-right>
      <div class="bk-button-group absolute left-[50%] -translate-x-1/2">
        <bk-button
          :class="editMode === 'form' ? 'is-selected' : ''"
          @click="changeMode('form')">{{ $t('templateFile.button.formMode') }}</bk-button>
        <bk-button
          :class="editMode === 'yaml' ? 'is-selected' : ''"
          @click="changeMode('yaml')">{{ $t('templateFile.button.yamlMode') }}</bk-button>
      </div>
      <div>{{ $t('templateFile.tag.latestUpdate') + ' : ' }}
        {{ fileMetadata?.updator }}
        {{ (fileMetadata?.updateAt && formatTime(fileMetadata.updateAt * 1000, 'yyyy-MM-dd hh:mm:ss')) || '--' }}</div>
    </template>
    <div
      :class="[
        'overflow-auto px-[24px] pt-[24px] h-[calc(100%-48px)]'
      ]">
      <FormMode
        :value="versionDetail.content"
        ref="formMode"
        v-if="editMode === 'form'" />
      <YamlMode
        class="h-full"
        v-else-if="editMode === 'yaml'"
        :value="versionDetail.content"
        ref="yamlMode" />
    </div>
    <!-- 转换异常提示 -->
    <bcs-dialog :show-footer="false" v-model="showErrorTipsDialog">
      <div class="flex flex-col items-center justify-center">
        <i
          :class="[
            'bk-icon icon-exclamation',
            'flex items-center justify-center w-[42px] h-[42px]',
            'text-[26px] text-[#ff9c01] bg-[#ffe8c3] rounded-full'
          ]">
        </i>
        <span class="mt-[20px] leading-[32px] text-[#313238] text-[20px]">
          {{ $t('templateFile.title.disableYamlToForm') }}
        </span>
        <div class="bg-[#F5F6FA] px-[16px] py-[12px] mt-[16px]">
          {{ $t('templateFile.tips.disableYamlToFormReason', [reason]) }}
        </div>
        <bcs-button theme="primary" class="mt-[16px]" @click="showErrorTipsDialog = false">
          {{ $t('generic.button.know') }}
        </bcs-button>
      </div>
    </bcs-dialog>
    <!-- 操作 -->
    <div
      :class="[
        'bcs-border-top',
        'flex items-center z-10 sticky bottom-0 h-[48px] px-[24px] bg-[#FAFBFD]'
      ]">
      <bcs-button
        theme="primary"
        class="min-w-[88px]"
        @click="handleShowDiffSlider">
        {{ $t('generic.button.save') }}
      </bcs-button>
      <bcs-button class="min-w-[88px]" @click="handleSaveDraft">{{$t('deploy.templateset.saveDraft')}}</bcs-button>
      <bcs-button
        class="min-w-[88px]"
        @click="back">
        {{ $t('generic.button.cancel') }}
      </bcs-button>
    </div>
    <!-- 元信息 -->
    <FileMetadataDialog
      :value="showMetadataDialog"
      :data="fileMetadata"
      @cancel="showMetadataDialog = false"
      @confirm="setMetadata" />
    <!-- diff -->
    <bcs-sideslider
      :is-show.sync="showDiffSlider"
      quick-close
      :title="$t('templateFile.label.versionDiff')"
      :width="1134"
      class="sideslider-full-content">
      <template #content>
        <DiffYaml
          class="h-[calc(100%-48px)]"
          :value="versionDetail.content"
          :original="originalContent"
          :original-version="versionDetail.version"
          :is-draft="fileMetadata?.isDraft" />
        <div class="bcs-border-top flex items-center h-[48px] px-[24px] bg-[#FAFBFD]">
          <bcs-button
            theme="primary"
            class="min-w-[88px]"
            :disabled="versionDetail.content === originalContent && !fileMetadata?.isDraft"
            @click="showVersionDialog = true">
            {{ $t('generic.button.confirmSave') }}
          </bcs-button>
          <bcs-button @click="showDiffSlider = false">{{ $t('generic.button.cancel') }}</bcs-button>
        </div>
      </template>
    </bcs-sideslider>
    <!-- 设置版号 -->
    <VersionDialog
      :value="showVersionDialog"
      :title="$t('templateFile.title.createVersion')"
      :sub-title="fileMetadata?.name"
      :version="versionDetail?.version"
      :loading="creating"
      auto-update
      @cancel="showVersionDialog = false"
      @confirm="createTemplateVersion" />
  </BcsContent>
</template>
<script setup lang="ts">
import { onBeforeMount, ref } from 'vue';

import DiffYaml from './diff-yaml.vue';
import FileMetadataDialog from './file-metadata.vue';
import FormMode from './form-mode.vue';
import VersionDialog from './version.vue';
import YamlMode from './yaml-mode.vue';

import { IListTemplateMetadataItem, ITemplateVersionItem } from '@/@types/cluster-resource-patch';
import { ResourceService, TemplateSetService } from '@/api/modules/new-cluster-resource';
import $bkMessage from '@/common/bkmagic';
import { formatTime } from '@/common/util';
import $bkInfo from '@/components/bk-magic-2.0/bk-info';
import BcsContent from '@/components/layout/Content.vue';
import $i18n from '@/i18n/i18n-setup';
import $router from '@/router';

interface Props {
  id: string // 模板文件ID
  versionID?: string // 基于当前模板文件版本ID创建新版本
  mode?: 'yaml'|'form' // 默认模式
}

const props = defineProps<Props>();

const isLoading = ref(false);
const creating = ref(false);
const versionDetail = ref<Partial<ITemplateVersionItem>>({
  content: '',
});
const originalContent = ref('');

// 切换模式
const showErrorTipsDialog = ref(false);
const reason = ref('');
const editMode = ref<'yaml'|'form'|undefined>(props.mode);
async function changeMode(type: 'yaml'|'form') {
  if (editMode.value === type) return;

  // 校验yaml能否转换表单
  if (type === 'form') {
    versionDetail.value.content = await handleGetReqData();
    const { canTransform: isCanTransform, message } = await canTransform(versionDetail.value.content);
    const result = !!isCanTransform;
    if (!result) {
      reason.value = message;
      showErrorTipsDialog.value = true;
      return;
    }
  } else if (type === 'yaml') {
    // 校验表单能否转yaml
    const result = await formMode.value?.validate();
    if (!result) return;

    versionDetail.value.content = await handleGetReqData();
  }
  editMode.value = type;
}

// 返回
function back() {
  $router.back();
}

// 获取表单数据
const formMode = ref<InstanceType<typeof FormMode>>();
const yamlMode = ref<InstanceType<typeof YamlMode>>();
const handleGetReqData = async () => {
  let content;
  switch (editMode.value) {
    case 'form':
      content = await formMode.value?.getData();
      break;
    case 'yaml':
      content = await yamlMode.value?.getData();
      break;
  }
  return content;
};

// 获取模板文件元数据
const loading = ref(false);
const fileMetadata = ref<IListTemplateMetadataItem>();
async function getTemplateMetadata() {
  if (!props.id) return;

  loading.value = true;
  fileMetadata.value = await TemplateSetService.GetTemplateMetadata({
    $id: props.id,
  }).catch(() => ({}));
  loading.value = false;
}

// 获取模板文件版本数据
async function getTemplateContent() {
  const versionID = props.versionID || fileMetadata.value?.versionID;
  if (!versionID) return;

  versionDetail.value = await TemplateSetService.GetTemplateVersion({
    $id: versionID,
  }).catch(() => ({}));
  // 保存原始数据
  originalContent.value = versionDetail.value.content || '';
  editMode.value = versionDetail.value?.editFormat || 'yaml';
}

// 获取草稿态数据
async function getDraftContent() {
  versionDetail.value = {
    content: fileMetadata.value?.draftContent,
    version: fileMetadata.value?.version,
  };
  // 保存原始数据
  originalContent.value = versionDetail.value.content || '';
  editMode.value = fileMetadata.value?.draftEditFormat || 'yaml';
}

// 修改元信息
const showMetadataDialog = ref(false);
const handleShowMetadataDialog = () => {
  showMetadataDialog.value = true;
};
const setMetadata = async (data: { name: string; description: string }) => {
  const result = await TemplateSetService.UpdateTemplateMetadata({
    $id: props.id, // 模板文件元数据 ID
    name: data.name, // 模板文件元数据名称
    description: data.description, // 模板文件元数据描述
    tags: fileMetadata.value?.tags || [], // 模板文件元数据标签
    version: fileMetadata.value?.version || '', // 模板文件版本
    versionMode: 0,
    // 保留之前草稿态
    isDraft: fileMetadata.value?.isDraft || false,
    draftVersion: fileMetadata.value?.draftVersion || '',
    draftContent: fileMetadata.value?.draftContent || '',
    draftEditFormat: editMode.value,
  }).then(() => true)
    .catch(() => false);
  if (result) {
    $bkMessage({
      theme: 'success',
      message: $i18n.t('generic.msg.success.save'),
    });
    showMetadataDialog.value = false;
    getTemplateMetadata();
  }
  showMetadataDialog.value = false;
};

const showDiffSlider = ref(false);
const showVersionDialog = ref(false);
async function handleShowDiffSlider() {
  let result;
  if (editMode.value === 'form') {
    result = await formMode.value?.validate();
  } else if (editMode.value === 'yaml') {
    result = await yamlMode.value?.validate();
  }
  if (!result) return;

  versionDetail.value.content = await handleGetReqData();
  showDiffSlider.value = true;
}

// 创建新版本
async function createTemplateVersion(versionData: { version: string; versionDescription: string }) {
  if (!fileMetadata.value?.id || !versionDetail.value.content) return;

  creating.value = true;
  const result = await TemplateSetService.CreateTemplateVersion({
    description: versionData?.versionDescription,
    $templateID: fileMetadata.value?.id,
    version: versionData?.version,
    content: versionDetail.value.content,
    force: false,
    editFormat: editMode.value as string,
  }).catch(() => false);
  creating.value = false;
  if (result) {
    showVersionDialog.value = false;
    $router.replace({
      name: 'templateFileDetail',
      params: {
        templateSpace: fileMetadata.value.templateSpaceID,
        id: props.id,
        isChanged: 'false',
      },
      query: {
        versionID: result?.id,
      },
    });
  }
}

// 保存草稿态
async function handleSaveDraft() {
  if (!fileMetadata.value?.id || !versionDetail.value.content) return;

  let isValid;
  if (editMode.value === 'form') {
    isValid = await formMode.value?.validate();
  } else if (editMode.value === 'yaml') {
    isValid = await yamlMode.value?.validate();
  }
  if (!isValid) return;

  versionDetail.value.content = await handleGetReqData();
  const result = await TemplateSetService.UpdateTemplateMetadata({
    $id: props.id, // 模板文件元数据 ID
    name: fileMetadata.value?.name, // 模板文件元数据名称
    description: fileMetadata.value?.description, // 模板文件元数据描述
    tags: fileMetadata.value?.tags || [], // 模板文件元数据标签
    version: fileMetadata.value?.version || '', // 模板文件版本
    versionMode: 0,
    // 保存草稿态
    isDraft: true,
    draftVersion: versionDetail.value.version || '',
    draftContent: versionDetail.value.content || '',
    draftEditFormat: editMode.value,
  }).then(() => true)
    .catch(() => false);
  if (result) {
    $bkMessage({
      theme: 'success',
      message: $i18n.t('generic.msg.success.save'),
    });
    await getTemplateMetadata();
    await getDraftContent();
  }
}

// 校验能否转化成表单
async function canTransform(content?: string): Promise<{ canTransform: boolean; message: string }> {
  if (!content) return { canTransform: true, message: '' };

  const data = await ResourceService.YAMLToForm({
    yaml: content,
  }).catch(() => ({ resources: [], canTransform: false }));
  return data;
}

const isContentChanged = ref(false);
const hasChanged = async () => {
  if (isContentChanged.value) {
    return true;
  }
  let resultContent;
  if (editMode.value === 'form') {
    resultContent = await formMode.value?.getData();
  } else if (editMode.value === 'yaml') {
    resultContent = await yamlMode.value?.getData();
  }
  isContentChanged.value = resultContent !== originalContent.value;
  return isContentChanged.value;
};
defineExpose({
  hasChanged,
});

onBeforeMount(async () => {
  isLoading.value = true;
  await getTemplateMetadata();
  if (fileMetadata.value?.isDraft) {
    // 草稿态
    await getDraftContent();
  } else {
    // 非草稿态
    await getTemplateContent();
  }
  isLoading.value = false;
});
</script>

<script lang="ts">
export default {
  async beforeRouteLeave(to, from, next) {
    if (to.params?.isChanged === 'false') return next();
    const result = await (this as any).hasChanged();
    if (result) {
      $bkInfo({
        title: $i18n.t('generic.msg.info.exitTips.text'),
        subTitle: $i18n.t('generic.msg.info.exitTips.subTitle'),
        clsName: 'custom-info-confirm default-info',
        okText: $i18n.t('generic.button.exit'),
        cancelText: $i18n.t('generic.button.cancel'),
        confirmFn() {
          next();
        },
        cancelFn() {
          next(false);
        },
      });
    } else {
      next();
    }
  },
};
</script>
<style scoped lang="postcss">
>>> .sideslider-full-content .bk-sideslider-content {
  height: 100%;
}
</style>
