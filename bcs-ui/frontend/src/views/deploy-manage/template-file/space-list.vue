<template>
  <div class="flex flex-col bg-[#FAFBFD] h-full" v-bkloading="{ isLoading: fileStore.spaceLoading }">
    <div class="flex items-center h-[42px] pr-[8px] pl-[14px] justify-between">
      <span class="text-[13px] font-bold">{{ $t('templateFile.title.fileList') }}</span>
      <div class="flex items-center">
        <i
          v-bk-tooltips="$t('templateFile.tips.createSpace')"
          class="text-[16px] mr-[5px] bcs-icon bcs-icon-xinjianwenjianjia
            cursor-pointer hover:text-[#3a84ff] transition"
          @click="showCreateFileDialog"></i>
        <!-- 更多操作 -->
        <span @click.stop>
          <PopoverSelector placement="bottom-end">
            <span class="bcs-icon-more-btn w-[16px] h-[16px]">
              <i class="bcs-icon bcs-icon-more"></i>
            </span>
            <template #content>
              <ul class="bg-[#fff]">
                <li class="bcs-dropdown-item" @click="showImportDialog = true">
                  {{ $t('templateFile.button.import') }}
                </li>
                <li class="bcs-dropdown-item" @click="exportTemplateSpaceAll">
                  {{ $t('templateFile.button.export') }}
                </li>
              </ul>
            </template>
          </PopoverSelector>
        </span>
      </div>
    </div>
    <div class="flex-[0_0_auto] flex items-center justify-center p-[8px] pt-0">
      <bk-input
        left-icon="bk-icon icon-search"
        clearable
        :placeholder="$t('templateFile.placeholder.searchSpace')"
        v-model.trim="searchKey">
      </bk-input>
    </div>
    <div class="flex-1 overflow-auto">
      <!-- 空间列表 -->
      <CollapseItem
        v-for="space in fileStore.spaceList"
        :key="space.id"
        :title="space.name"
        :active="curSpaceID === space.id && !curFileID"
        :collapse="!collapseSpaceIDs.includes(space.id)"
        :id="space.id"
        :loading="fileStore.loadingSpaceIDs.includes(space.id)"
        class="bg-[#fff]"
        @collapse-change="() => handleCollapseChange(space)">
        <template #title>
          <div
            class="flex-1 flex items-center justify-between h-[32px]"
            @mouseenter="hoverItemID = space.id"
            @mouseleave="hoverItemID = ''"
            @click="handleChangeSpace(space.id)">
            <span class="flex items-center">
              <i
                class="bcs-icon bcs-icon-star-shape text-[#ff9C01] mr-[3px] leading-[18px]"
                v-if="space?.fav"></i>
              <span class="bcs-ellipsis" v-bk-overflow-tips="{ interactive: false }">{{ space.name }}</span>
            </span>
            <!-- 空间操作 -->
            <span v-if="hoverItemID === space.id || curPopover === space.id" @click.stop>
              <PopoverSelector offset="0, 6" :on-hide="hidePopover" :on-show="() => showPopover(space.id)">
                <span class="bcs-icon-more-btn w-[16px] h-[16px]">
                  <i class="bcs-icon bcs-icon-more"></i>
                </span>
                <template #content>
                  <ul class="bg-[#fff]">
                    <li class="bcs-dropdown-item" @click="showRenameSpaceDialog(space)">
                      {{ $t('templateFile.button.rename') }}
                    </li>
                    <li class="bcs-dropdown-item" @click="showCloneSpaceDialog(space)">
                      {{ $t('generic.button.clone') }}
                    </li>
                    <li class="bcs-dropdown-item" @click="addTemplateFile(space)">
                      {{ $t('templateFile.button.createFile') }}
                    </li>
                    <li class="bcs-dropdown-item" @click="handleFavorite(space)">
                      {{ space?.fav ?
                        $t('templateFile.button.RemoveFromFavorites') :
                        $t('templateFile.button.AddToFavorites')
                      }}
                    </li>
                    <li class="bcs-dropdown-item" @click="exportTemplateSpace(space)">
                      {{ $t('templateFile.button.exportFolder') }}
                    </li>
                    <li class="bcs-dropdown-item" @click="deleteSpace(space)">
                      {{ $t('templateFile.button.delete') }}
                    </li>
                  </ul>
                </template>
              </PopoverSelector>
            </span>
          </div>
        </template>
        <!-- 空间下文件列表 -->
        <div
          v-for="file in fileStore.fileListMap[space.id]"
          :key="file.id"
          offset="0, 6"
          :class="[
            'flex items-center justify-between h-[32px] pl-[36px] pr-[8px] text-[12px] hover:bg-[#F5F7FA]',
            'cursor-pointer',
            {
              '!text-[#3A84FF] !bg-[#E1ECFF]': curFileID === file.id
            }
          ]"
          @mouseenter="hoverItemID = file.id"
          @mouseleave="hoverItemID = ''"
          @click="handleChangeFile(space.id, file)">
          <span class="bcs-ellipsis" v-bk-overflow-tips="{ interactive: false }">{{ file.name }}</span>
          <!-- 文件操作 -->
          <span @click.stop>
            <PopoverSelector
              v-if="hoverItemID === file.id || curPopover === file.id"
              :on-hide="hidePopover"
              :on-show="() => showPopover(file.id)">
              <span class="bcs-icon-more-btn w-[16px] h-[16px]"><i class="bcs-icon bcs-icon-more"></i></span>
              <template #content>
                <ul>
                  <li class="bcs-dropdown-item" @click="editFile(file)">{{ $t('generic.button.edit') }}</li>
                  <li class="bcs-dropdown-item" @click="deployFile(file)">
                    {{ $t('templateSet.button.deploy') }}
                  </li>
                  <li class="bcs-dropdown-item" @click="cloneVersion(file)">
                    {{ $t('generic.button.clone') }}
                  </li>
                  <li class="bcs-dropdown-item" @click="mangeFileVersion(space, file)">
                    {{ $t('templateFile.button.versionManage') }}
                  </li>
                  <li class="bcs-dropdown-item" @click="deleteFile(space, file)">{{ $t('generic.button.delete') }}</li>
                </ul>
              </template>
            </PopoverSelector>
          </span>
        </div>
      </CollapseItem>
      <bcs-exception type="empty" scene="part" v-if="!fileStore.spaceList.length" />
    </div>
    <!-- 新增 & 修改文件夹 -->
    <bcs-dialog
      v-model="showNameDialog"
      :title="typeText.title[curType]"
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
              v-model.trim="curEditSpace.name"
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
            :disabled="!curEditSpace.name"
            @click="fileNameConfirm">
            {{ typeText.button[curType] }}
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
          :template-space="curEditSpace.id"
          class="px-[24px] py-[20px]"
          @deleteFile="refresh" />
      </template>
    </bcs-sideslider>
    <!-- 导入文件夹 -->
    <bcs-dialog
      v-model="showImportDialog"
      :title="$t('templateFile.title.import')"
      width="580"
      header-position="left">
      <bcs-upload
        :tip="$t('templateFile.tips.uploadTips', [50])"
        with-credentials
        accept=".tgz,.tar.gz"
        :custom-request="customRequest"
        :limit="1"
        :size="50"
        ref="uploadRef"
        @on-delete="handleDelete"
      />
      <template #footer>
        <bcs-button
          theme="primary"
          :loading="importing"
          :disabled="disabledImport"
          @click="importConfirm">
          {{ $t('generic.button.import') }}
        </bcs-button>
        <bcs-button @click="showImportDialog = false">{{ $t('generic.button.cancel') }}</bcs-button>
      </template>
    </bcs-dialog>
  </div>
</template>
<script setup lang="ts">
import { cloneDeep } from 'lodash';
import { onBeforeMount, ref, set, watch } from 'vue';

import { searchKey, store as fileStore, updateListTemplateSpaceList, updateTemplateMetadataList } from './use-store';
import VersionList from './version-list.vue';

import { IListTemplateMetadataItem, ITemplateSpaceData } from '@/@types/cluster-resource-patch';
import { exportTemplate, importTemplate } from '@/api/modules/cluster-resource';
import { TemplateSetService } from '@/api/modules/new-cluster-resource';
import $bkMessage from '@/common/bkmagic';
import { download } from '@/common/util';
import $bkInfo from '@/components/bk-magic-2.0/bk-info';
import CollapseItem from '@/components/collapse-item.vue';
import PopoverSelector from '@/components/popover-selector.vue';
import Validate from '@/components/validate.vue';
import $i18n from '@/i18n/i18n-setup';
import $router from '@/router';

interface Props {
  templateSpace?: string// 空间ID
  id?: string // 文件ID
}
type Emits = (e: 'reload') => void;
const props = defineProps<Props>();
const emits = defineEmits<Emits>();

// 当前空间列表
const curSpaceID = ref(props.templateSpace);
// 当前空间文件
const curFileID = ref(props.id);
const hoverItemID = ref('');
// 当前操作Popover
const curPopover = ref('');
const hidePopover = () => {
  curPopover.value = '';
};
const showPopover = (id: string) => {
  curPopover.value = id;
};
// 文件夹是否重复
const repeatMsg = ref('');

// 展开 & 收起 空间
const collapseSpaceIDs = ref<string[]>([]);
function handleCollapseChange(space: ITemplateSpaceData) {
  const index = collapseSpaceIDs.value.findIndex(id => space.id === id);
  if (index > -1) {
    collapseSpaceIDs.value.splice(index, 1);
  } else {
    collapseSpaceIDs.value.push(space.id);
  }
}

// 切换空间
async function handleChangeSpace(spaceID: string) {
  if (curSpaceID.value === spaceID && !curFileID.value) return;

  curSpaceID.value = spaceID;
  curFileID.value = '';
  await $router.replace({
    name: 'templateFileList',
    params: {
      templateSpace: spaceID,
    },
  });
}

// 切换详情
function handleChangeFile(spaceID: string, file: IListTemplateMetadataItem) {
  if (curSpaceID.value === spaceID && curFileID.value === file.id) return;

  curSpaceID.value = spaceID;
  curFileID.value = file.id;
  $router.replace({
    name: 'templateFileDetail',
    params: {
      templateSpace: spaceID,
      id: file.id,
    },
    query: {
      versionID: file.versionID,
    },
  });
};

// 获取空间列表
async function listTemplateSpace() {
  await updateListTemplateSpaceList();
  // 校验当前active的空间是否存在列表中
  const existSpace = fileStore.spaceList.find(item => item.id === curSpaceID.value);
  if (!existSpace) {
    await handleChangeSpace(fileStore.spaceList[0]?.id);
  }
  initCollapseSpaceID();
}

// 收藏操作
async function handleFavorite(space) {
  if (space.fav) {
    await TemplateSetService.UnCollectFolder({ $templateSpaceID: space.id });
  } else {
    await TemplateSetService.CollectFolder({ $templateSpaceID: space.id });
  }
  updateListTemplateSpaceList();
};

// 获取空间下文件列表
async function getTemplateMetadata(spaceID: string) {
  await updateTemplateMetadataList(spaceID);
}

// 创建 & 修改工作空间
const showNameDialog = ref(false);
const showImportDialog = ref(false);
const fileNameRef = ref();
const curEditSpace = ref<Pick<ITemplateSpaceData, 'name'|'id'>>({
  name: '',
  id: '',
});
const curType = ref<'create' | 'rename' | 'clone'>('create');
const typeText = ref({
  title: {
    create: $i18n.t('templateFile.title.createSpace'),
    rename: $i18n.t('templateFile.title.rename'),
    clone: $i18n.t('templateFile.title.clone'),
  },
  button: {
    create: $i18n.t('generic.button.create'),
    rename: $i18n.t('generic.button.save'),
    clone: $i18n.t('generic.button.clone'),
  },
  message: {
    create: $i18n.t('generic.msg.success.create'),
    rename: $i18n.t('generic.msg.success.save'),
    clone: $i18n.t('generic.msg.success.ok'),
  },
});
const saving = ref(false);
function showCreateFileDialog() {
  curEditSpace.value = { name: '', id: '' };
  curType.value = 'create';
  showNameDialog.value = true;
};
function showRenameSpaceDialog(space: ITemplateSpaceData) {
  curEditSpace.value = cloneDeep(space);
  curType.value = 'rename';
  showNameDialog.value = true;
}
function showCloneSpaceDialog(space: ITemplateSpaceData) {
  curEditSpace.value = cloneDeep(space);
  curType.value = 'clone';
  showNameDialog.value = true;
}
// 创建 & 编辑工作空间的对话框确认事件
async function fileNameConfirm() {
  if (!curEditSpace.value?.name || (curType.value !== 'create' && !curEditSpace.value?.id)) return;

  saving.value = true;
  let res;
  if (curType.value === 'rename') {
    res = await TemplateSetService.UpdateTemplateSpace({
      name: curEditSpace.value.name,
      description: '',
      $id: curEditSpace.value.id,
    }, { globalError: false, needRes: true }).catch(() => false);
  } else if (curType.value === 'create') {
    res = await TemplateSetService.CreateTemplateSpace({
      name: curEditSpace.value.name,
      description: '',
    }, { globalError: false, needRes: true }).catch(() => false);
  } else {
    res = await TemplateSetService.CloneTemplateSpace({
      name: curEditSpace.value.name,
      description: '',
      $id: curEditSpace.value.id,
    }, { globalError: false, needRes: true }).catch(() => false);
  }
  if (res && res?.code === 0) {
    $bkMessage({
      theme: 'success',
      message: typeText.value.message[curType.value],
    });
    await listTemplateSpace();
    if (curType.value !== 'rename') {
      // 跳转到新增的空间下
      handleChangeSpace(res?.data?.id);
    } else {
      emits('reload');
    }
    showNameDialog.value = false;
  } else if (res?.code === 7) {
    // 文件夹重名
    repeatMsg.value = res?.message || $i18n.t('generic.validate.fieldRepeat', [$i18n.t('templateFile.label.spaceName')]);
  }
  saving.value = false;
}

// 导入文件夹
let options: any;// todo types
const importing = ref(false);
const uploadRef = ref();
const disabledImport = ref(true);
async function customRequest(importOptions) {
  options = importOptions;
  disabledImport.value = false;
};
// 导出文件夹
function exportTemplateSpaceAll() {
  if (fileStore.spaceList.length === 0) return;
  const templateSpaceNames = fileStore.spaceList.map(space => space.name);
  downloadTemplateFile(templateSpaceNames);
};
// 导出单个文件夹
function exportTemplateSpace(space: ITemplateSpaceData) {
  const templateSpaceNames = [space.name];
  downloadTemplateFile(templateSpaceNames);
};
// 下载模板文件
async function downloadTemplateFile(templateSpaceNames: string[]) {
  const res = await exportTemplate({ templateSpaceNames }, { responseType: 'blob', needRes: true });
  if (!res) return;

  const dispositionList: string[] = res.headers?.['content-disposition']?.split(';') || [];
  const fileName = dispositionList.at(-1)?.split('=')
    .at(-1) || `template-${new Date().getTime()}.tgz`;
  download(res.data, fileName);
};
// 点击导入按钮
async function importConfirm() {
  if (!options) return;
  importing.value = true;
  const formData = new FormData();
  formData.append('templateFile', options.fileObj?.origin, options.fileObj?.origin?.name);
  const result = await importTemplate(formData).then(() => true)
    .catch(() => false);
  if (result) {
    $bkMessage({
      theme: 'success',
      message: $i18n.t('generic.msg.success.import'),
    });
    showImportDialog.value = false;
    listTemplateSpace();
  } else {
    $bkMessage({
      theme: 'error',
      message: $i18n.t('generic.msg.error.import'),
    });
  }
  importing.value = false;
}
// 点击导入取消按钮
function clearFiles() {
  if (options?.fileObj?.origin) {
    uploadRef.value.deleteFile(0, options?.fileObj?.origin);
    handleDelete();
  }
}
// 删除操作
function handleDelete() {
  options = null;
  disabledImport.value = true;
};

// 添加模板文件
function addTemplateFile(space: ITemplateSpaceData) {
  if (!space.id) return;
  $router.push({
    name: 'addTemplateFile',
    params: {
      templateSpace: space.id,
    },
  });
}

// 删除工作空间
async function deleteSpace(space: ITemplateSpaceData) {
  if (!space.id) return;
  $bkInfo({
    type: 'warning',
    clsName: 'custom-info-confirm',
    title: $i18n.t('generic.title.confirmDelete1', { name: space.name }),
    subTitle: $i18n.t('templateFile.tips.spaceSubTitle'),
    defaultInfo: true,
    okText: $i18n.t('generic.button.delete'),
    confirmFn: async () => {
      const result = await TemplateSetService.DeleteTemplateSpace({
        $id: space.id,
      }).then(() => true)
        .catch(() => false);
      if (result) {
        listTemplateSpace();
      }
    },
  });
}

// 删除文件
function deleteFile(space: ITemplateSpaceData, file: IListTemplateMetadataItem) {
  if (!space.id || !file.id) return;

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
        await getTemplateMetadata(space.id);
        $router.replace({
          name: 'templateFileList',
          params: {
            templateSpace: props.templateSpace as string,
          },
        });
      }
    },
  });
}

// 编辑文件版本
function editFile(file: IListTemplateMetadataItem) {
  hidePopover();
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

// 文件版本管理
const curEditFile = ref<IListTemplateMetadataItem>();
const showVersionList = ref(false);
function mangeFileVersion(space: ITemplateSpaceData, file: IListTemplateMetadataItem) {
  curEditSpace.value = cloneDeep(space);
  curEditFile.value = cloneDeep(file);
  showVersionList.value = true;
}

// 部署文件
function deployFile(file: IListTemplateMetadataItem) {
  hidePopover();
  $router.push({
    name: 'templateFileDeploy',
    params: {
      id: file.id,
    },
  });
}

// 克隆版本
function cloneVersion(file: IListTemplateMetadataItem) {
  hidePopover();
  $router.push({
    name: 'addTemplateFile',
    params: {
      templateSpace: props.templateSpace as string,
      versionID: file.versionID,
    },
  });
}

// 初始化默认折叠空间
function initCollapseSpaceID() {
  if (!props.templateSpace || !fileStore.spaceList.length) return;

  // 删除已经不存在的ID
  collapseSpaceIDs.value.forEach((id, index) => {
    const exist = fileStore.spaceList.find(item => item.id === id);
    if (!exist) {
      collapseSpaceIDs.value.splice(index, 1);
    }
  });
  // 添加默认展开的空间
  const index = collapseSpaceIDs.value.findIndex(spaceID => spaceID === props.templateSpace);
  if (index === -1) {
    collapseSpaceIDs.value.push(props.templateSpace);
  }
}

// 刷新列表
async function refresh() {
  showVersionList.value = false;
  updateTemplateMetadataList(curSpaceID.value || '');
  emits('reload');
}

watch(() => props.templateSpace, () => {
  curSpaceID.value = props.templateSpace;
  initCollapseSpaceID();
});

watch(() => props.id, () => {
  curFileID.value = props.id;
  initCollapseSpaceID();
});

// 新建 & 修改空间名称时聚焦输入框
watch(showNameDialog, () => {
  if (showNameDialog.value && fileNameRef.value) {
    setTimeout(() => {
      fileNameRef.value.focus();
    });
  }
});

// 获取空间下文件
watch(collapseSpaceIDs, async () => {
  const ids = collapseSpaceIDs.value.filter(spaceID => !fileStore.fileListMap[spaceID]);
  const list = ids.map(spaceID => TemplateSetService.ListTemplateMetadata({
    $templateSpaceID: spaceID,
  }).then((data) => {
    set(fileStore.fileListMap, spaceID, data);
  }));
  fileStore.loadingSpaceIDs = [...ids];
  await Promise.all(list).catch(() => []);
  fileStore.loadingSpaceIDs = [];
}, { deep: true });

const watchOnce = watch(() => fileStore.fileListMap[props.templateSpace || ''], () => {
  if (!fileStore.fileListMap[props.templateSpace || '']?.length) return;
  setTimeout(() => {
    // 自动滚动到当前空间
    const spaceDoms = document.getElementById(`${props.templateSpace}`);
    spaceDoms?.scrollIntoView();
  });
  watchOnce();
});

watch(() => props.templateSpace, () => {
  if (!props.templateSpace) {
    handleChangeSpace(fileStore.spaceList[0]?.id);
  }
});

watch(searchKey, () => {
  updateListTemplateSpaceList();
});

watch(showImportDialog, () => {
  if (!showImportDialog.value) clearFiles();
});

onBeforeMount(() => {
  // 重置数据
  fileStore.spaceList = [];
  fileStore.fileListMap = {};
  fileStore.loadingSpaceIDs = [];
  listTemplateSpace();
});
</script>
