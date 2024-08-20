<template>
  <div class="p-[24px] h-full overflow-y-auto">
    <Row class="mb-[16px]">
      <template #left>
        <span
          :class="[
            'inline-flex items-center justify-center h-[32px] bg-[#EAEBF0] px-[12px]',
            'text-[#313238] text-[14px] font-bold rounded-full mr-[16px]'
          ]"
          v-if="templateSpace">
          {{ spaceDetail?.name }}
          <i
            class="hover:text-[#3a84ff] ml-[6px] cursor-pointer bk-icon icon-edit-line"
            @click="showRenameSpaceDialog">
          </i>
        </span>
        <bcs-button theme="primary" @click="addTemplateFile">
          {{ $t('templateFile.button.createFile') }}
        </bcs-button>
      </template>
      <template #right>
        <bcs-input
          :placeholder="$t('templateFile.placeholder.searchFile')"
          class="w-[420px]"
          clearable
          right-icon="bk-icon icon-search"
          v-model.trim="searchValue">
        </bcs-input>
      </template>
    </Row>
    <bcs-table
      v-bkloading="{ isLoading: fileLoading }"
      :data="curPageData"
      :pagination="pagination"
      @page-change="pageChange"
      @page-limit-change="pageSizeChange">
      <bcs-table-column :label="$t('templateFile.label.name')" prop="name" show-overflow-tooltip>
        <template #default="{ row }">
          <div class="flex items-center">
            <bcs-button
              class="bcs-ellipsis"
              text
              @click="handleGotoFileDetail(row)">{{ row.name }}</bcs-button>
            <bk-tag class="flex-shrink-0 px-1.5" theme="success" radius="3px" v-if="row.isDraft">
              {{ $t('templateFile.tag.draft') }}</bk-tag>
          </div>
        </template>
      </bcs-table-column>
      <bcs-table-column :label="$t('templateFile.label.resourceType')" prop="resourceType">
        <template #default="{ row }">
          {{ row.resourceType?.join(',') || '--' }}
        </template>
      </bcs-table-column>
      <bcs-table-column :label="$t('templateFile.label.latestVersion')" prop="version">
        <template #default="{ row }">
          {{ row.version || '--' }}
        </template>
      </bcs-table-column>
      <bcs-table-column :label="$t('generic.label.updator')" prop="updator"></bcs-table-column>
      <bcs-table-column :label="$t('generic.label.updatedAt')" prop="updateAt">
        <template #default="{ row }">
          {{ formatTime(row.updateAt * 1000, 'yyyy-MM-dd hh:mm:ss') }}
        </template>
      </bcs-table-column>
      <bcs-table-column :label="$t('generic.label.action')" width="140">
        <template #default="{ row }">
          <bcs-button text class="mr-[5px]" @click="editFile(row)">{{ $t('generic.button.edit') }}</bcs-button>
          <span v-bk-tooltips="{ content: $t('templateSet.tips.cantDeploy'), disabled: row.version, delay: 300 }">
            <bcs-button text :disabled="!row.version" @click="deployFile(row)">
              {{ $t('templateSet.button.deploy') }}</bcs-button>
          </span>
          <PopoverSelector>
            <span class="bcs-icon-more-btn"><i class="bcs-icon bcs-icon-more"></i></span>
            <template #content>
              <ul>
                <li class="bcs-dropdown-item" @click="cloneVersion(row)">
                  {{ $t('generic.button.clone') }}
                </li>
                <li class="bcs-dropdown-item" @click="mangeFileVersion(row)">
                  {{ $t('templateFile.button.versionManage') }}
                </li>
                <li class="bcs-dropdown-item" @click="deleteFile(row)">
                  {{ $t('generic.button.delete') }}
                </li>
              </ul>
            </template>
          </PopoverSelector>
        </template>
      </bcs-table-column>
      <template #empty>
        <BcsEmptyTableStatus :type="searchValue ? 'search-empty' : 'empty'" @clear="searchValue = ''" />
      </template>
    </bcs-table>
    <!-- 修改文件夹名称 -->
    <bcs-dialog
      v-model="showNameDialog"
      :title="$t('templateFile.title.rename')"
      width="480"
      header-position="left">
      <bcs-form :label-width="110">
        <bcs-form-item required :label="$t('templateFile.label.spaceName')">
          <Validate
            :message="repeatMsg"
            error-display-type="normal">
            <bcs-input
              ref="fileNameRef"
              :maxlength="64"
              v-model.trim="curSpaceName"
              @change="repeatMsg = ''">
            </bcs-input>
          </Validate>
        </bcs-form-item>
      </bcs-form>
      <template #footer>
        <div>
          <bcs-button
            theme="primary"
            :loading="saving"
            :disabled="!curSpaceName"
            @click="fileNameConfirm">
            {{ $t('generic.button.save') }}
          </bcs-button>
          <bcs-button @click="showNameDialog = false">{{ $t('generic.button.cancel') }}</bcs-button>
        </div>
      </template>
    </bcs-dialog>
    <!-- 版本管理 -->
    <bcs-sideslider
      :is-show.sync="showVersionList"
      quick-close
      :title="`${curEditFile?.name} ${$t('templateFile.title.versionManage')}`"
      :width="960">
      <template #content>
        <VersionList
          :template-i-d="curEditFile?.id"
          :template-space="templateSpace"
          class="px-[24px] py-[20px]"
          @delete="getTemplateMetadata"
          @deleteFile="refresh" />
      </template>
    </bcs-sideslider>
  </div>
</template>
<script setup lang="ts">
import { cloneDeep } from 'lodash';
import { onActivated, onBeforeMount, ref, watch } from 'vue';

import { updateListTemplateSpaceList, updateTemplateMetadataList } from './use-store';
import VersionList from './version-list.vue';

import { IListTemplateMetadataItem, ITemplateSpaceData } from '@/@types/cluster-resource-patch';
import { TemplateSetService } from '@/api/modules/new-cluster-resource';
import $bkMessage from '@/common/bkmagic';
import { formatTime } from '@/common/util';
import $bkInfo from '@/components/bk-magic-2.0/bk-info';
import Row from '@/components/layout/Row.vue';
import PopoverSelector from '@/components/popover-selector.vue';
import Validate from '@/components/validate.vue';
import usePage from '@/composables/use-page';
import useSearch from '@/composables/use-search';
import $i18n from '@/i18n/i18n-setup';
import $router from '@/router';

interface Props {
  templateSpace: string// 空间ID
}

const props = defineProps<Props>();

// 获取空间详情
const spaceDetail = ref<ITemplateSpaceData>();
async function getTemplateSpace() {
  if (!props.templateSpace) return;

  spaceDetail.value = await TemplateSetService.GetTemplateSpace({
    $id: props.templateSpace,
  });
}

// 获取空间下文件列表
const fileLoading = ref(false);
const fileList = ref<IListTemplateMetadataItem[]>([]);
const keys = ref(['name', 'resourceType', 'version', 'updator', 'updateAt']);
const { searchValue, tableDataMatchSearch } = useSearch(fileList, keys);
const {
  pagination,
  curPageData,
  pageChange,
  pageSizeChange,
} = usePage<IListTemplateMetadataItem>(tableDataMatchSearch);
async function getTemplateMetadata() {
  if (!props.templateSpace) return;

  fileLoading.value = true;
  fileList.value = await TemplateSetService.ListTemplateMetadata({
    $templateSpaceID: props.templateSpace,
  }).catch(() => []);
  fileLoading.value = false;
}

// 添加模板文件
function addTemplateFile() {
  $router.push({
    name: 'addTemplateFile',
    params: {
      templateSpace: props.templateSpace,
    },
  });
}

// 重命名空间
const repeatMsg = ref('');
const showNameDialog = ref(false);
const curSpaceName = ref('');
const saving = ref(false);
function showRenameSpaceDialog() {
  curSpaceName.value = spaceDetail.value?.name || '';
  showNameDialog.value = true;
}
async function fileNameConfirm() {
  saving.value = true;
  const res = await TemplateSetService.UpdateTemplateSpace({
    name: curSpaceName.value,
    description: '',
    $id: props.templateSpace,
  }, { globalError: false, needRes: true }).catch(() => false);

  if (res && res?.code === 0) {
    $bkMessage({
      theme: 'success',
      message: $i18n.t('generic.msg.success.save'),
    });
    getTemplateSpace();
    updateListTemplateSpaceList();// 更新空间列表
    showNameDialog.value = false;
  } else if (res?.code === 7) {
    // 文件夹重名
    repeatMsg.value = res?.message || $i18n.t('generic.validate.fieldRepeat', [$i18n.t('templateFile.label.spaceName')]);
  }
  saving.value = false;
}

// 文件夹详情
function handleGotoFileDetail(file: IListTemplateMetadataItem) {
  $router.replace({
    name: 'templateFileDetail',
    params: {
      templateSpace: props.templateSpace,
      id: file.id,
    },
    query: {
      versionID: file.versionID,
    },
  });
}

// 编辑文件版本
function editFile(file: IListTemplateMetadataItem) {
  $router.push({
    name: 'addTemplateFileVersion',
    params: {
      id: file.id,
    },
    query: {
      versionID: file.versionID,
    },
  });
}

// 部署文件
function deployFile(file: IListTemplateMetadataItem) {
  $router.push({
    name: 'templateFileDeploy',
    params: {
      id: file.id,
    },
    query: {
      version: file.version,
    },
  });
}

// 文件版本管理
const curEditFile = ref<IListTemplateMetadataItem>();
const showVersionList = ref(false);
function mangeFileVersion(file: IListTemplateMetadataItem) {
  curEditFile.value = cloneDeep(file);
  showVersionList.value = true;
}

// 克隆版本
function cloneVersion(file: IListTemplateMetadataItem) {
  $router.push({
    name: 'addTemplateFile',
    params: {
      templateSpace: props.templateSpace,
      versionID: file.versionID,
    },
  });
}

// 删除文件
function deleteFile(file: IListTemplateMetadataItem) {
  $bkInfo({
    type: 'warning',
    clsName: 'custom-info-confirm',
    title: $i18n.t('generic.title.confirmDelete1', { name: file.name }),
    defaultInfo: true,
    okText: $i18n.t('generic.button.delete'),
    confirmFn: async () => {
      const result = await TemplateSetService.DeleteTemplateMetadata({
        $id: file.id as string,
      }).then(() => true)
        .catch(() => false);
      if (result) {
        getTemplateMetadata();
        updateTemplateMetadataList(props.templateSpace);// 更新空间下的文件
      }
    },
  });
}

// 刷新列表
function refresh() {
  showVersionList.value = false;
  getTemplateMetadata();
  updateTemplateMetadataList(props.templateSpace);
}

watch(() => props.templateSpace, () => {
  getTemplateSpace();
  getTemplateMetadata();
});

onActivated(() => {
  getTemplateMetadata();
  updateTemplateMetadataList(props.templateSpace);
});

onBeforeMount(() => {
  getTemplateSpace();
  getTemplateMetadata();
});
</script>
