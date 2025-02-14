<template>
  <BcsContent :title="formData.name" :padding="0">
    <template #title>
      <i class="bcs-icon bcs-icon-arrows-left back mr-[4px]" @click="back"></i>
      <span>{{ formData.name }}</span>
      <span class="bcs-icon-btn ml-[10px]" @click="handleShowMetadataDialog">
        <i class="bk-icon icon-edit-line"></i>
      </span>
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
    </template>
    <div
      :class="[
        'overflow-auto px-[24px] pt-[24px] h-[calc(100%-48px)]'
      ]">
      <FormMode
        :value="formData.content"
        ref="formMode"
        is-add
        v-if="editMode === 'form'" />
      <YamlMode
        class="h-full"
        is-add
        v-else-if="editMode === 'yaml'"
        :value="formData.content"
        ref="yamlMode"
        render-mode="Simple"
        @getUpgradeStatus="getUpgradeStatus"
        @change="handleChange" />
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
          <span v-if="isHelm">{{ $t('templateFile.tips.disableYamlToFormReasonHelm', [reason]) }}</span>
          <span v-else>{{ $t('templateFile.tips.disableYamlToFormReason', [reason]) }}</span>
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
        'flex items-center z-10 sticky bottom-0 h-[48px] px-[24px] bg-[#FAFBFD]',
      ]">
      <div
        class="mr-[8px]"
        v-bk-tooltips="{
          content: $t('templateFile.tips.emptyContent'),
          disabled: hasContent || editMode === 'form'
        }">
        <bcs-button
          theme="primary"
          class="min-w-[88px]"
          :disabled="!hasContent && editMode === 'yaml'"
          @click="handleSaveData">
          {{ $t('generic.button.save') }}
        </bcs-button>
      </div>
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
      :data="formData"
      @cancel="showMetadataDialog = false"
      @confirm="setMetadata" />
    <!-- 设置版号 -->
    <VersionDialog
      :value="showVersionDialog"
      :title="$t('templateFile.title.createVersion')"
      :sub-title="formData.name"
      :version="formData.version"
      :loading="creating"
      @cancel="showVersionDialog = false"
      @confirm="handleCreateFile" />
  </BcsContent>
</template>
<script setup lang="ts">
import { onBeforeMount, reactive, ref } from 'vue';

import FileMetadataDialog from './file-metadata.vue';
import FormMode from './form-mode.vue';
import VersionDialog from './version.vue';
import YamlMode from './yaml-mode.vue';

import { ITemplateVersionItem } from '@/@types/cluster-resource-patch';
import { ResourceService, TemplateSetService } from '@/api/modules/new-cluster-resource';
import $bkInfo from '@/components/bk-magic-2.0/bk-info';
import BcsContent from '@/components/layout/Content.vue';
import $i18n from '@/i18n/i18n-setup';
import $router from '@/router';

interface Props {
  templateSpace: string
  versionID?: string
}

const props = defineProps<Props>();

const creating = ref(false);
const formData = reactive<ClusterResource.CreateTemplateMetadataReq>({
  $templateSpaceID: props.templateSpace,
  name: `template-${Math.floor(Date.now() / 1000)}`,
  description: '',
  content: '',
  version: '',
  tags: [],
  versionDescription: '',
  isDraft: false,
  draftContent: '',
  draftVersion: '',
  draftEditFormat: 'form',
});

// 切换模式
const showErrorTipsDialog = ref(false);
const reason = ref('');
const editMode = ref<'yaml'|'form'>('form');
async function changeMode(type: 'yaml'|'form') {
  if (editMode.value === type) return;

  if (type === 'form') {
    formData.content = await handleGetReqData();
    // 校验yaml能否转换表单
    const { canTransform: isCanTransform, message } = await canTransform(formData.content);
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

    formData.content = await handleGetReqData();
  }
  editMode.value = type;
  formData.draftEditFormat = type;
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

// 基本信息
const showMetadataDialog = ref(false);
const handleShowMetadataDialog = () => {
  showMetadataDialog.value = true;
};
const setMetadata = (data: Pick<ClusterResource.CreateTemplateMetadataReq, 'name'|'description'>) => {
  // 设置元信息
  formData.name = data.name;
  formData.description = data.description;
  showMetadataDialog.value = false;
};

// 新建模板文件
const showVersionDialog = ref(false);
async function handleSaveData() {
  let result;
  if (editMode.value === 'form') {
    result = await formMode.value?.validate();
  } else if (editMode.value === 'yaml') {
    result = await yamlMode.value?.validate();
  }
  if (!result) return;
  showVersionDialog.value = true;
}

// 是否升级为Helm 模板
const isHelm = ref(false);
function getUpgradeStatus(data) {
  isHelm.value = data.isHelm;
};

async function handleCreateFile(versionData: Pick<ClusterResource.CreateTemplateMetadataReq, 'version'|'versionDescription'>) {
  // 设置版本信息
  formData.version = versionData.version;
  formData.versionDescription = versionData.versionDescription;
  const content = await handleGetReqData();

  creating.value = true;
  const params = isHelm.value ? { renderMode: 'Helm' } : {};
  const result = await TemplateSetService.CreateTemplateMetadata({
    ...formData,
    $templateSpaceID: props.templateSpace,
    content,
    ...params,
  }).then(() => true)
    .catch(() => false);
  creating.value = false;
  if (result) {
    showVersionDialog.value = false;
    $router.push({
      name: 'templateFileList',
      params: {
        templateSpace: props.templateSpace,
        isChanged: 'false',
      },
    });
  }
}

// 保存草稿态
async function handleSaveDraft() {
  formData.isDraft = true;
  formData.draftContent = await handleGetReqData();

  creating.value = true;
  const params = isHelm.value ? { renderMode: 'Helm' } : {};
  const result = await TemplateSetService.CreateTemplateMetadata({
    ...formData,
    $templateSpaceID: props.templateSpace,
    ...params,
  }).then(() => true)
    .catch(() => false);
  creating.value = false;
  if (result) {
    showVersionDialog.value = false;
    $router.push({
      name: 'templateFileList',
      params: {
        templateSpace: props.templateSpace,
        isChanged: 'false',
      },
    });
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

// 获取模板文件版本数据
async function getTemplateContent() {
  if (!props.versionID) return;
  const data: ITemplateVersionItem = await TemplateSetService.GetTemplateVersion({
    $id: props.versionID,
  }).catch(() => ({}));
  formData.content = data.content;
}

const hasContent = ref(false);
function handleChange(content) {
  hasContent.value = !!content.trim();
}

onBeforeMount(() => {
  getTemplateContent();
});
</script>

<script lang="ts">
export default {
  async beforeRouteLeave(to, from, next) {
    if (to.params?.isChanged === 'false') return next();
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
  },
};
</script>
