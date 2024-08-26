<template>
  <BcsContent :title="$t('templateFile.title.file')" :padding="0" hide-back>
    <template #header-right v-if="id">
      <div class="bk-button-group absolute left-[50%] -translate-x-1/2">
        <bk-button
          :class="fileStore.editMode === 'form' ? 'is-selected' : ''"
          :disabled="fileStore.isFormModeDisabled"
          @click="changeMode('form')">{{ $t('templateFile.button.formMode') }}</bk-button>
        <bk-button
          :class="fileStore.editMode === 'yaml' ? 'is-selected' : ''"
          @click="changeMode('yaml')">{{ $t('templateFile.button.yamlMode') }}</bk-button>
      </div>
      <div class="absolute right-[24px] top-[10px]">
        <bk-button theme="primary" @click="deployFile" v-show="fileStore.showDeployBtn">
          {{ $t('templateSet.button.deploy' ) }}</bk-button>
        <bk-button @click="editFile">{{ $t('generic.button.edit') }}</bk-button>
        <bk-button theme="danger" :loading="deleting" @click="deleteFile">
          {{ $t('generic.button.delete') }}</bk-button>
      </div>
    </template>
    <bcs-resize-layout
      collapsible
      class="h-full border-x-0"
      :initial-divide="200"
      :min="200"
      :max="420">
      <template #aside>
        <SpaceList
          :template-space="templateSpace"
          :id="id"
          @reload="reload" />
      </template>
      <template #main>
        <RouterView v-if="reloadView" />
      </template>
    </bcs-resize-layout>
  </BcsContent>
</template>
<script setup lang="ts">
import { onBeforeMount, ref } from 'vue';

import SpaceList from './space-list.vue';
import { store as fileStore, updateTemplateMetadataList } from './use-store';

import { IListTemplateMetadataItem } from '@/@types/cluster-resource-patch';
import { TemplateSetService  } from '@/api/modules/new-cluster-resource';
import $bkInfo from '@/components/bk-magic-2.0/bk-info';
import BcsContent from '@/components/layout/Content.vue';
import $i18n from '@/i18n/i18n-setup';
import $router from '@/router';

interface Props {
  templateSpace?: string// 空间ID
  id?: string // 文件ID
  versionID?: string
  mode?: 'yaml'|'form'// 编辑态默认模式
}

const props = defineProps<Props>();

const reloadView = ref(true);
async function changeMode(type: 'yaml'|'form') {
  if (fileStore.editMode === type) return;

  fileStore.editMode = type;
  await $router.replace({
    query: {
      mode: type,
      version: $router.currentRoute?.query?.version,
      versionID: $router.currentRoute?.query?.versionID,
    },
  });
}

function reload() {
  reloadView.value = false;
  setTimeout(() => {
    reloadView.value = true;
  });
}

// 部署文件
function deployFile() {
  $router.push({
    name: 'templateFileDeploy',
    params: {
      id: props.id as string,
    },
  });
}

// 编辑文件
const editFile = () => {
  $router.push({
    name: 'addTemplateFileVersion',
    params: {
      id: props.id as string,
    },
    query: {
      mode: fileStore.editMode,
      versionID: props.versionID as string,
    },
  });
};

// 删除
const deleting = ref(false);
async function getTemplateMetadata(id: string) {
  deleting.value = true;
  const data = await TemplateSetService.GetTemplateMetadata({
    $id: id,
  }).catch(() => ({}));
  deleting.value = false;
  return data as IListTemplateMetadataItem;
}
async function deleteFile() {
  const data = await getTemplateMetadata(props.id as string);
  $bkInfo({
    type: 'warning',
    clsName: 'custom-info-confirm',
    title: $i18n.t('generic.title.confirmDelete1', { name: data.name }),
    defaultInfo: true,
    okText: $i18n.t('generic.button.delete'),
    confirmFn: async () => {
      const result = await TemplateSetService.DeleteTemplateMetadata({
        $id: props.id as string,
      }).then(() => true)
        .catch(() => false);
      if (result) {
        $router.replace({
          name: 'templateFileList',
          params: {
            templateSpace: props.templateSpace as string,
          },
        });
        updateTemplateMetadataList(props.templateSpace || ''); // 更新当前空间下文件列表
      }
    },
  });
}

onBeforeMount(() => {
  fileStore.editMode = props.mode;
});
</script>
