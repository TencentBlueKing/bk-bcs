<template>
  <div>
    <bcs-input
      right-icon="bk-icon icon-search"
      :placeholder="$t('templateFile.placeholder.searchVersion')"
      clearable
      v-model.trim="searchValue">
    </bcs-input>
    <bcs-table
      class="mt-[16px]"
      v-bkloading="{ isLoading: loading }"
      :data="curPageData"
      :pagination="pagination"
      @page-change="pageChange"
      @page-limit-change="pageSizeChange">
      <bcs-table-column :label="$t('templateFile.label.version')" prop="version">
        <template #default="{ row }">
          <div class="flex items-center">
            <span class="truncate">{{ getVersionText(row) }}</span>
            <bcs-tag class="flex-shrink-0" theme="warning" v-if="row.latest">latest</bcs-tag>
          </div>
        </template>
      </bcs-table-column>
      <bcs-table-column
        :label="$t('templateFile.label.versionDesc')"
        prop="description"
        show-overflow-tooltip>
        <template #default="{ row }">
          {{ row.description || '--' }}
        </template>
      </bcs-table-column>
      <bcs-table-column :label="$t('generic.label.updator')" prop="creator">
        <template #default="{ row }">
          {{ row.creator || '--' }}
        </template>
      </bcs-table-column>
      <bcs-table-column :label="$t('generic.label.updatedAt')" prop="createAt">
        <template #default="{ row }">
          {{ formatTime(row.createAt * 1000, 'yyyy-MM-dd hh:mm:ss') }}
        </template>
      </bcs-table-column>
      <bcs-table-column :label="$t('generic.label.action')" width="120">
        <template #default="{ row }">
          <bcs-button
            text
            @click="handleAction(row)">
            {{ row.draft ? $t('generic.button.edit') : $t('templateFile.button.deploy') }}
          </bcs-button>
          <bcs-button
            :disabled="row.latest"
            class="ml-[5px]"
            text
            @click="handleDelete(row)">{{ $t('generic.button.delete') }}</bcs-button>
        </template>
      </bcs-table-column>
      <template #empty>
        <BcsEmptyTableStatus :type="searchValue ? 'search-empty' : 'empty'" @clear="searchValue = ''" />
      </template>
    </bcs-table>
  </div>
</template>
<script setup lang="ts">
import { onBeforeMount, ref, watch } from 'vue';

import { store as fileStore } from './use-store';

import { IListTemplateMetadataItem, ITemplateVersionItem } from '@/@types/cluster-resource-patch';
import { TemplateSetService } from '@/api/modules/new-cluster-resource';
import $bkMessage from '@/common/bkmagic';
import { formatTime } from '@/common/util';
import $bkInfo from '@/components/bk-magic-2.0/bk-info';
import usePage from '@/composables/use-page';
import useSearch from '@/composables/use-search';
import $i18n from '@/i18n/i18n-setup';
import $router from '@/router';

const props = defineProps({
  templateID: {
    type: String,
    default: '',
    required: true,
  },
  templateSpace: {
    type: String,
    default: '',
    required: true,
  },
});
const emits = defineEmits(['delete', 'deleteFile']);

// 版本列表
const loading = ref(false);
const versionList = ref<ITemplateVersionItem[]>([]);
const keys = ref(['version', 'description', 'creator']);
const { searchValue, tableDataMatchSearch } = useSearch(versionList, keys);
const {
  pagination,
  pageConf,
  curPageData,
  pageChange,
  pageSizeChange,
} = usePage<ITemplateVersionItem>(tableDataMatchSearch);
async function listTemplateVersion() {
  if (!props.templateID) return;

  loading.value = true;
  versionList.value = await TemplateSetService.ListTemplateVersion({
    $templateID: props.templateID,
  }).catch(() => []);
  pageConf.current = 1;
  loading.value = false;
}

// 部署文件
const deployFile = (row: ITemplateVersionItem) => {
  $router.push({
    name: 'templateFileDeploy',
    params: {
      id: props.templateID,
    },
    query: {
      version: row.version,
    },
  });
};

// 编辑文件
const editFile = (row: ITemplateVersionItem) => {
  $router.push({
    name: 'addTemplateFileVersion',
    params: {
      id: props.templateID,
    },
    query: {
      mode: fileStore.editMode,
      versionID: row.version,
    },
  });
};

// 操作
const handleAction = (row: ITemplateVersionItem) => {
  if (!row.draft) {
    deployFile(row);
  } else {
    editFile(row);
  }
};

// 获取版本文案
const getVersionText = (row: ITemplateVersionItem) => {
  let result = '';
  if (row.draft) {
    result = row.version
      ? `${$i18n.t('templateFile.tag.draft')}( ${$i18n.t('templateFile.tag.baseOn')} ${row.version} )`
      : $i18n.t('templateFile.tag.draft');
  } else {
    result = row.version;
  }
  return result;
};

// 获取模板文件详情数据
const fileMetadata = ref<IListTemplateMetadataItem>();
async function getTemplateMetadata() {
  if (!props.templateID) return;

  loading.value = true;
  fileMetadata.value = await TemplateSetService.GetTemplateMetadata({
    $id: props.templateID,
  }).catch(() => ({}));
  loading.value = false;
}

// 删除模板草稿版本
const deleteDraft = async () => {
  const result = await TemplateSetService.UpdateTemplateMetadata({
    $id: fileMetadata.value?.id || '', // 模板文件元数据 ID
    name: fileMetadata.value?.name || '', // 模板文件元数据名称
    description: fileMetadata.value?.description || '', // 模板文件元数据描述
    tags: fileMetadata.value?.tags || [], // 模板文件元数据标签
    version: fileMetadata.value?.version || '', // 模板文件版本
    versionMode: 0,
    // 删除之前草稿态
    isDraft: false,
    draftVersion: '',
    draftContent: '',
  }).then(() => true)
    .catch(() => false);
  if (result) {
    $bkMessage({
      theme: 'success',
      message: $i18n.t('generic.msg.success.delete'),
    });
    getTemplateMetadata();
    listTemplateVersion();
    emits('delete');
  }
};

// 删除模板文件
async function deleteFile() {
  $bkInfo({
    type: 'warning',
    clsName: 'custom-info-confirm',
    title: $i18n.t('generic.title.confirmDelete1', { name: fileMetadata.value?.name }),
    defaultInfo: true,
    okText: $i18n.t('generic.button.delete'),
    confirmFn: async () => {
      const result = await TemplateSetService.DeleteTemplateMetadata({
        $id: props.templateID as string,
      }).then(() => true)
        .catch(() => false);
      if (result) {
        emits('deleteFile', fileMetadata.value?.id);
      }
    },
  });
}

// 删除操作
const handleDelete = (row: ITemplateVersionItem) => {
  if (!row.draft) {
    deleteVersion(row);
  } else {
    // 只有一个版本且是草稿态时，直接删除模板文件
    if (versionList.value.length === 1 && versionList.value[0].draft === true) {
      deleteFile();
    } else {
      // 否则，只删除草稿版本
      deleteDraft();
    }
  }
};

// 删除版本
function deleteVersion(row: ITemplateVersionItem) {
  $bkInfo({
    type: 'warning',
    clsName: 'custom-info-confirm',
    title: $i18n.t('generic.title.confirmDelete1', { name: row.version }),
    defaultInfo: true,
    okText: $i18n.t('generic.button.delete'),
    confirmFn: async () => {
      const result = await TemplateSetService.DeleteTemplateVersion({
        $id: row.id,
      }).then(() => true)
        .catch(() => false);
      if (result) {
        $bkMessage({
          theme: 'success',
          message: $i18n.t('generic.msg.success.delete'),
        });
        listTemplateVersion();
        emits('delete');
      }
    },
  });
}
watch(
  () => [
    props.templateID,
  ],
  () => {
    listTemplateVersion();
  },
  { immediate: true },
);

onBeforeMount(() => {
  getTemplateMetadata();
});
</script>
