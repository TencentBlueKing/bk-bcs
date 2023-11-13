<template>
  <bk-loading :loading="loading" style="height: 100%">
    <table class="config-groups-table">
      <thead>
        <tr class="config-groups-table-tr">
          <th class="name">配置文件名称</th>
          <th class="version">配置模板版本</th>
          <th class="path">配置路径</th>
          <th class="user">创建人</th>
          <th class="user">修改人</th>
          <th class="datetime">修改时间</th>
          <th class="status">变更状态</th>
          <th class="operation">操作</th>
        </tr>
      </thead>
      <tbody>
        <template v-for="group in tableGroupsData" :key="group.id" v-if="allConfigCount !== 0">
          <tr class="config-groups-table-tr group-title-row" v-if="group.configs.length > 0">
            <td colspan="8" class="config-groups-table-td">
              <div class="configs-group">
                <div class="name-wrapper" @click="group.expand = !group.expand">
                  <DownShape :class="['fold-icon', { fold: !group.expand }]" />
                  {{ group.name }}
                </div>
                <div
                  v-if="isUnNamedVersion && group.id !== 0"
                  v-cursor="{ active: !hasEditServicePerm }"
                  :class="['delete-btn', { 'bk-text-with-no-perm': !hasEditServicePerm }]"
                  @click="handleDeletePkg(group.id, group.name)"
                >
                  <Close class="close-icon" />
                  移除套餐
                </div>
              </div>
            </td>
          </tr>
          <template v-if="group.expand">
            <tr class="config-groups-table-tr">
              <td colspan="8" class="config-groups-table-td">
                <div class="configs-list-wrapper">
                  <table class="config-list-table">
                    <tbody>
                      <tr v-for="config in group.configs" :key="config.id" class="config-row">
                        <td>
                          <template v-if="group.id === 0">
                            <bk-button
                              v-if="isUnNamedVersion"
                              v-cursor="{ active: !hasEditServicePerm }"
                              text
                              theme="primary"
                              :class="{ 'bk-text-with-no-perm': !hasEditServicePerm }"
                              :disabled="hasEditServicePerm && config.file_state === 'DELETE'"
                              @click="handleEditOpen(config)"
                            >
                              {{ config.name }}
                            </bk-button>
                            <bk-button
                              v-else
                              text
                              theme="primary"
                              :disabled="config.file_state === 'DELETE'"
                              @click="handleViewConfig(config.id, 'config')"
                            >
                              {{ config.name }}
                            </bk-button>
                          </template>
                          <bk-button
                            v-else
                            text
                            theme="primary"
                            :disabled="config.file_state === 'DELETE'"
                            @click="handleViewConfig(config.versionId, 'template')"
                          >
                            {{ config.name }}
                          </bk-button>
                        </td>
                        <td>{{ config.versionName }}</td>
                        <td>{{ config.path }}</td>
                        <td class="user">{{ config.creator }}</td>
                        <td class="user">{{ config.reviser }}</td>
                        <td class="datetime">{{ config.update_at }}</td>
                        <td class="status"><StatusTag :status="config.file_state" /></td>
                        <td class="operation">
                          <div class="config-actions">
                            <!-- 非套餐配置文件 -->
                            <template v-if="group.id === 0">
                              <template v-if="isUnNamedVersion">
                                <bk-button
                                  v-cursor="{ active: !hasEditServicePerm }"
                                  text
                                  theme="primary"
                                  :class="{ 'bk-text-with-no-perm': !hasEditServicePerm }"
                                  :disabled="hasEditServicePerm && config.file_state === 'DELETE'"
                                  @click="handleEditOpen(config)"
                                >
                                  编辑
                                </bk-button>
                                <bk-button
                                  v-cursor="{ active: !hasEditServicePerm }"
                                  text
                                  theme="primary"
                                  :class="{ 'bk-text-with-no-perm': !hasEditServicePerm }"
                                  :disabled="hasEditServicePerm && config.file_state === 'DELETE'"
                                  @click="handleDel(config)"
                                >
                                  删除
                                </bk-button>
                              </template>
                              <template v-else>
                                <bk-button text theme="primary" @click="handleViewConfig(config.id, 'config')"
                                  >查看</bk-button
                                >
                                <bk-button
                                  v-if="versionData.status.publish_status !== 'editing'"
                                  text
                                  theme="primary"
                                  @click="handleConfigDiff(group.id, config)"
                                  >对比</bk-button
                                >
                              </template>
                            </template>
                            <!-- 套餐模板 -->
                            <template v-else>
                              <bk-button
                                v-if="isUnNamedVersion"
                                v-cursor="{ active: !hasEditServicePerm }"
                                text
                                theme="primary"
                                :class="{ 'bk-text-with-no-perm': !hasEditServicePerm }"
                                @click="handleOpenReplaceVersionDialog(group.id, config)"
                              >
                                替换版本
                              </bk-button>
                              <template v-else>
                                <bk-button text theme="primary" @click="handleViewConfig(config.versionId, 'template')"
                                  >查看</bk-button
                                >
                                <bk-button
                                  v-if="versionData.status.publish_status !== 'editing'"
                                  text
                                  theme="primary"
                                  @click="handleConfigDiff(group.id, config)"
                                  >对比</bk-button
                                >
                              </template>
                            </template>
                          </div>
                        </td>
                      </tr>
                    </tbody>
                  </table>
                </div>
              </td>
            </tr>
          </template>
        </template>
        <tr v-else>
          <td colspan="8">
            <TableEmpty :is-search-empty="isSearchEmpty" @clear="emits('clearStr')" style="width: 100%" />
          </td>
        </tr>
      </tbody>
    </table>
  </bk-loading>
  <edit-config
    v-model:show="editPanelShow"
    :config-id="activeConfig"
    :bk-biz-id="props.bkBizId"
    :app-id="props.appId"
    @confirm="handleEditConfigConfirm"
  />
  <ViewConfig
    v-model:show="viewConfigSliderData.open"
    v-bind="viewConfigSliderData.data"
    :bk-biz-id="props.bkBizId"
    :app-id="props.appId"
    :version-id="versionData.id"
  />
  <VersionDiff v-model:show="isDiffPanelShow" :current-version="versionData" :selected-config="diffConfig" />
  <ReplaceTemplateVersion
    v-model:show="replaceDialogData.open"
    v-bind="replaceDialogData.data"
    :binding-id="bindingId"
    :bk-biz-id="props.bkBizId"
    :app-id="props.appId"
    @updated="getAllConfigList"
  />
  <DeleteConfirmDialog
    v-model:isShow="isDeleteConfigDialogShow"
    title="确认删除该配置文件？"
    @confirm="handleDeleteConfigConfirm"
  >
    <div style="margin-bottom: 8px;">配置文件：<span style="color: #313238;">{{ deleteConfig?.name }}</span></div>
    <div>一旦删除，该操作将无法撤销，请谨慎操作</div>
  </DeleteConfirmDialog>
  <DeleteConfirmDialog
    v-model:isShow="isDeletePkgDialogShow"
    title="确认移除该配置模板套餐？"
    @confirm="handleDeletePkgConfirm"
  >
    <div style="margin-bottom: 8px;">配置模板套餐: <span style="color: #313238;">{{ deleteTemplatePkgName }}</span></div>
    <div>移除后本服务配置将不再引用该配置模板套餐，以后需要时可以重新从配置模板导入</div>
  </DeleteConfirmDialog>
</template>
<script lang="ts" setup>
import { ref, computed, watch, onMounted } from 'vue';
import { storeToRefs } from 'pinia';
import { Message } from 'bkui-vue/lib';
import { DownShape, Close } from 'bkui-vue/lib/icon';
import useConfigStore from '../../../../../../../../store/config';
import useServiceStore from '../../../../../../../../store/service';
import { ICommonQuery } from '../../../../../../../../../types/index';
import { datetimeFormat } from '../../../../../../../../utils/index';
import { IConfigItem, IBoundTemplateGroup, IConfigDiffSelected } from '../../../../../../../../../types/config';
import {
  getConfigList,
  getReleasedConfigList,
  getBoundTemplates,
  getBoundTemplatesByAppVersion,
  deleteServiceConfigItem,
  deleteBoundPkg,
} from '../../../../../../../../api/config';
import { getAppPkgBindingRelations } from '../../../../../../../../api/template';
import StatusTag from './status-tag';
import EditConfig from '../edit-config.vue';
import ViewConfig from '../view-config.vue';
import VersionDiff from '../../../components/version-diff/index.vue';
import ReplaceTemplateVersion from '../replace-template-version.vue';
import TableEmpty from '../../../../../../../../components/table/table-empty.vue';
import DeleteConfirmDialog from '../../../../../../../../components/delete-confirm-dialog.vue';

interface IConfigsGroupData {
  id: number;
  name: string;
  expand: boolean;
  configs: IConfigTableItem[];
}

interface IPermissionType {
  privilege: string;
  user: string;
  user_group: string;
}

interface IConfigTableItem {
  id: number;
  name: string;
  versionId: number;
  versionName: string;
  path: string;
  creator: string;
  reviser: string;
  update_at: string;
  file_state: string;
  permission?: IPermissionType;
}

const configStore = useConfigStore();
const serviceStore = useServiceStore();
const { versionData, allConfigCount } = storeToRefs(configStore);
const { checkPermBeforeOperate } = serviceStore;
const { permCheckLoading, hasEditServicePerm } = storeToRefs(serviceStore);

const props = defineProps<{
  bkBizId: string;
  appId: number;
  searchStr: string;
}>();

const emits = defineEmits(['clearStr', 'deleteConfig']);

const loading = ref(false);
const commonConfigListLoading = ref(false);
const bindingId = ref(0);
const configList = ref<IConfigItem[]>([]); // 非模板配置文件
const configsCount = ref(0);
const boundTemplateListLoading = ref(false);
const templateGroupList = ref<IBoundTemplateGroup[]>([]); // 配置文件模板（按套餐分组）
const templatesCount = ref(0);
const tableGroupsData = ref<IConfigsGroupData[]>([]);
const editPanelShow = ref(false);
const activeConfig = ref(0);
const isDiffPanelShow = ref(false);
const isSearchEmpty = ref(false);
const isDeleteConfigDialogShow = ref(false);
const deleteConfig = ref<IConfigTableItem>();
const isDeletePkgDialogShow = ref(false);
const deleteTemplatePkgName = ref('');
const deleteTemplatePkgId = ref(0);
const viewConfigSliderData = ref({
  open: false,
  data: { id: 0, type: '' },
});
const replaceDialogData = ref({
  open: false,
  data: {
    pkgId: 0,
    templateId: 0,
    versionId: 0,
    versionName: '',
  },
});
const diffConfig = ref<IConfigDiffSelected>({
  pkgId: 0,
  id: 0,
  version: 0,
});

// 是否为未命名版本
const isUnNamedVersion = computed(() => versionData.value.id === 0);

watch(
  () => versionData.value.id,
  async () => {
    await getBindingId();
    getAllConfigList();
  },
);

watch(
  () => props.searchStr,
  () => {
    props.searchStr ? (isSearchEmpty.value = true) : (isSearchEmpty.value = false);
    getAllConfigList();
  },
);

watch([() => configsCount.value, () => templatesCount.value], () => {
  configStore.$patch((state) => {
    state.allConfigCount = configsCount.value + templatesCount.value;
  });
});

onMounted(async () => {
  await getBindingId();
  getAllConfigList();
});

const getBindingId = async () => {
  const res = await getAppPkgBindingRelations(props.bkBizId, props.appId);
  bindingId.value = res.details.length === 1 ? res.details[0].id : 0;
};

const getAllConfigList = async () => {
  await Promise.all([getCommonConfigList(), getBoundTemplateList()]);
  tableGroupsData.value = transListToTableData();
};

// 获取非模板配置文件列表
const getCommonConfigList = async () => {
  commonConfigListLoading.value = true;
  try {
    const params: ICommonQuery = {
      start: 0,
      all: true,
    };

    let res;
    if (isUnNamedVersion.value) {
      if (props.searchStr) {
        params.search_fields = 'name,path,memo,creator,reviser';
        params.search_value = props.searchStr;
      }
      res = await getConfigList(props.bkBizId, props.appId, params);
    } else {
      if (props.searchStr) {
        params.search_fields = 'name,path,memo,creator';
        params.search_value = props.searchStr;
      }
      res = await getReleasedConfigList(props.bkBizId, props.appId, versionData.value.id, params);
    }
    configList.value = res.details;
    configsCount.value = res.count;
  } catch (e) {
    console.error(e);
  } finally {
    commonConfigListLoading.value = false;
  }
};

// 获取模板配置文件列表
const getBoundTemplateList = async () => {
  boundTemplateListLoading.value = true;
  try {
    const params: ICommonQuery = {
      start: 0,
      all: true,
    };
    if (props.searchStr) {
      params.search_fields = 'revision_name,revision_memo,name,path,creator';
      params.search_value = props.searchStr;
    }

    let res;
    if (isUnNamedVersion.value) {
      res = await getBoundTemplates(props.bkBizId, props.appId, params);
    } else {
      res = await getBoundTemplatesByAppVersion(props.bkBizId, props.appId, versionData.value.id, params);
    }
    templateGroupList.value = res.details;
    templatesCount.value = res.details.reduce(
      (acc: number, crt: IBoundTemplateGroup) => acc + crt.template_revisions.length,
      0,
    );
  } catch (e) {
    console.error(e);
  } finally {
    boundTemplateListLoading.value = false;
  }
};

const transListToTableData = () => {
  const pkgsGroups = groupTplsByPkg(templateGroupList.value);
  return [
    { id: 0, name: '非模板配置', expand: true, configs: transConfigsToTableItemData(configList.value) },
    ...pkgsGroups,
  ];
};

// 将非模板配置文件数据转为表格数据
const transConfigsToTableItemData = (list: IConfigItem[]) => list.map((item: IConfigItem) => {
  const { id, spec, revision, file_state } = item;
  const { name, path, permission } = spec;
  const { creator, reviser, update_at } = revision;
  return {
    id,
    name,
    versionId: 0,
    versionName: '--',
    path,
    creator,
    reviser,
    update_at: datetimeFormat(update_at),
    file_state,
    permission,
  };
});

// 将模板按套餐分组，并将模板数据格式转为表格数据
const groupTplsByPkg = (list: IBoundTemplateGroup[]) => {
  const groups: IConfigsGroupData[] = list.map((groupItem) => {
    const { template_space_name, template_set_id, template_set_name, template_revisions } = groupItem;
    const group: IConfigsGroupData = {
      id: template_set_id,
      name: `${template_space_name} - ${template_set_name}`,
      expand: true,
      configs: [],
    };
    template_revisions.forEach((tpl) => {
      const {
        template_id: id,
        name,
        template_revision_id: versionId,
        template_revision_name: versionName,
        path,
        creator,
        create_at,
        file_state,
      } = tpl;
      group.configs.push({
        id,
        name,
        versionId,
        versionName,
        path,
        creator,
        reviser: creator,
        update_at: datetimeFormat(create_at),
        file_state,
      });
    });
    return group;
  });
  return groups;
};

const handleEditOpen = (config: IConfigTableItem) => {
  if (permCheckLoading.value || !checkPermBeforeOperate('update')) {
    return;
  }
  activeConfig.value = config.id;
  editPanelShow.value = true;
};

// 编辑模板
const handleEditConfigConfirm = async () => {
  await getCommonConfigList();
  tableGroupsData.value = transListToTableData();
  emits('deleteConfig');
};

// 查看配置文件或模板版本
const handleViewConfig = (id: number, type: string) => {
  viewConfigSliderData.value = {
    open: true,
    data: { id, type },
  };
};

const handleOpenReplaceVersionDialog = (pkgId: number, config: IConfigTableItem) => {
  if (permCheckLoading.value || !checkPermBeforeOperate('update')) {
    return;
  }
  const { id: templateId, versionId, versionName } = config;
  replaceDialogData.value = {
    open: true,
    data: { pkgId, templateId, versionId, versionName },
  };
};

// 删除模板套餐
const handleDeletePkg = async (pkgId: number, name: string) => {
  if (permCheckLoading.value || !checkPermBeforeOperate('update')) {
    return;
  }
  isDeletePkgDialogShow.value = true;
  deleteTemplatePkgName.value = name;
  deleteTemplatePkgId.value = pkgId;
};

const handleDeletePkgConfirm = async () => {
  await deleteBoundPkg(props.bkBizId, props.appId, bindingId.value, [deleteTemplatePkgId.value]);
  await getBoundTemplateList();
  tableGroupsData.value = transListToTableData();
  emits('deleteConfig');
  Message({
    theme: 'success',
    message: '删除模板套餐成功',
  });
  isDeletePkgDialogShow.value = false;
};

// 非模板配置文件diff
const handleConfigDiff = (groupId: number, config: IConfigTableItem) => {
  diffConfig.value = {
    pkgId: groupId,
    id: config.id,
    version: config.versionId,
    permission: config.permission,
  };
  isDiffPanelShow.value = true;
};

// 删除配置文件
const handleDel = (config: IConfigTableItem) => {
  if (permCheckLoading.value || !checkPermBeforeOperate('update')) {
    return;
  }
  isDeleteConfigDialogShow.value = true;
  deleteConfig.value = config;
};

const handleDeleteConfigConfirm = async () => {
  await deleteServiceConfigItem(deleteConfig.value!.id, props.bkBizId, props.appId);
  await getCommonConfigList();
  emits('deleteConfig');
  tableGroupsData.value = transListToTableData();
  isDeleteConfigDialogShow.value = false;
};

defineExpose({
  refresh: getAllConfigList,
});
</script>
<style lang="scss" scoped>
.config-groups-table {
  width: 100%;
  border: 1px solid #dedee5;
  border-collapse: collapse;
  table-layout: fixed;
  .config-groups-table-tr {
    th {
      padding: 11px 16px;
      color: #313238;
      font-weight: normal;
      font-size: 12px;
      line-height: 20px;
      text-align: left;
      background: #fafbfd;
      border-bottom: 1px solid #dedee5;
    }
  }
  .group-title-row:hover {
    background: #f5f7fa;
  }
  .config-groups-table-td {
    padding: 0;
    font-size: 12px;
    line-height: 20px;
    text-align: left;
    border-bottom: 1px solid #dedee5;
    color: #63656e;
  }
  .configs-group {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 0 16px;
    height: 42px;
    .name-wrapper {
      display: flex;
      align-items: center;
      height: 100%;
      cursor: pointer;
      .fold-icon {
        margin-right: 8px;
        font-size: 14px;
        color: #3a84ff;
        transition: transform 0.2s ease-in-out;
        &.fold {
          color: #c4c6cc;
          transform: rotate(-90deg);
        }
      }
    }
    .delete-btn {
      display: flex;
      align-items: center;
      margin-right: 20px;
      color: #3a84ff;
      cursor: pointer;
      .close-icon {
        margin-right: 4px;
        font-size: 14px;
      }
    }
  }
  .user {
    width: 120px;
  }
  .datetime {
    width: 158px;
  }
  .status {
    width: 100px;
  }
  .operation {
    width: 100px;
  }
  .exception-tips {
    margin: 20px 0;
  }
}
.configs-list-wrapper {
  max-height: 420px;
  overflow: auto;
}
.config-list-table {
  width: 100%;
  border-collapse: collapse;
  table-layout: fixed;
  .config-row {
    &:hover {
      background: #f5f7fa;
    }
    &:last-child td {
      border: none;
    }
  }
  td {
    padding: 11px 16px;
    height: 42px;
    border-bottom: 1px solid #dedee5;
    font-size: 12px;
    line-height: 20px;
    color: #63656e;
    white-space: nowrap;
    text-overflow: ellipsis;
    overflow: hidden;
  }
  .config-actions {
    .bk-button:not(:last-child) {
      margin-right: 8px;
    }
  }
}
</style>

<style lang="scss">
.delete-template-pkg {
  .bk-modal-body {
    padding: 0 !important;
    .bk-dialog-header {
      padding: 48px 24px 104px !important;
      .bk-dialog-title {
        width: 352px;
        height: 32px;
        font-size: 20px;
        color: #313238;
        letter-spacing: 0;
        text-align: center;
        line-height: 32px;
      }
    }
    .bk-modal-content {
      display: none;
    }
    .bk-modal-footer {
      bottom: 48px !important;
      border: none !important;
      background-color: #fff !important;
    }
  }
}
</style>
