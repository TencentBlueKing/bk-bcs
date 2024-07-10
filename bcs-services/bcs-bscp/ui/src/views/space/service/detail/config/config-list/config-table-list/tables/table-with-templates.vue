<template>
  <div ref="tableRef" class="table-container">
    <bk-loading :loading="loading" style="height: 100%">
      <table class="config-groups-table" :key="appId">
        <thead>
          <tr class="config-groups-table-tr">
            <th v-if="isUnNamedVersion" class="selection">
              <bk-checkbox :model-value="isIndeterminate" :indeterminate="isIndeterminate" @change="handleSelectAll" />
            </th>
            <th class="name">{{ t('配置文件绝对路径') }}</th>
            <th class="version">{{ t('配置模板版本') }}</th>
            <th class="user">{{ t('创建人') }}</th>
            <th class="user">{{ t('修改人') }}</th>
            <th class="datetime">{{ t('修改时间') }}</th>
            <th class="status" v-if="versionData.id === 0">
              {{ t('变更状态') }}
              <TableFilter :filter-list="statusFilterList" @selected="handleStatusFilterSelected" />
            </th>
            <th class="operation">{{ t('操作') }}</th>
          </tr>
        </thead>
        <tbody>
          <template v-for="(group, index) in tableGroupsData" :key="group.id" v-if="allConfigCount !== 0">
            <tr
              :class="[
                'config-groups-table-tr',
                'group-title-row',
                group.expand ? 'expand' : '',
                { sticky: stickyIndex === index },
              ]"
              v-if="group.configs.length > 0">
              <td :colspan="colsLen" class="config-groups-table-td">
                <div class="configs-group">
                  <div class="name-wrapper" @click="group.expand = !group.expand">
                    <DownShape :class="['fold-icon', { fold: !group.expand }]" />
                    {{ group.name }}
                  </div>
                  <div
                    v-if="isUnNamedVersion && group.id !== 0"
                    v-cursor="{ active: !hasEditServicePerm }"
                    :class="['delete-btn', { 'bk-text-with-no-perm': !hasEditServicePerm }]"
                    @click="handleDeletePkg(group.id, group.name)">
                    <Close class="close-icon" />
                    {{ t('移除套餐') }}
                  </div>
                </div>
              </td>
            </tr>
            <tr v-show="group.expand && group.configs.length > 0" ref="collapseHeader" class="config-groups-table-tr">
              <td :colspan="colsLen" class="config-groups-table-td">
                <div class="configs-list-wrapper">
                  <table class="config-list-table">
                    <tbody>
                      <tr v-for="config in group.configs" :key="config.id" :class="getRowCls(config)">
                        <td v-if="isUnNamedVersion" class="selection">
                          <bk-checkbox
                            :disabled="group.id > 0 || config.file_state === 'DELETE'"
                            :model-value="selectedConfigItems.some((item) => item.id === config.id)"
                            @change="handleRowSelectionChange($event, config.id)" />
                        </td>
                        <td class="name">
                          <ContentWidthOverflowTips>
                            <template v-if="group.id === 0">
                              <div
                                v-if="isUnNamedVersion"
                                v-cursor="{ active: !hasEditServicePerm }"
                                :class="[
                                  'file-name-btn',
                                  {
                                    'bk-text-with-no-perm': !hasEditServicePerm,
                                    disabled: !hasEditServicePerm || config.file_state === 'DELETE',
                                  },
                                ]"
                                @click="
                                  () => {
                                    hasEditServicePerm && config.file_state !== 'DELETE' && handleViewConfig(config.id);
                                  }
                                ">
                                {{ fileAP(config) }}
                              </div>
                              <div
                                v-else
                                :class="['file-name-btn', { disabled: config.file_state === 'DELETE' }]"
                                @click="
                                  () => {
                                    config.file_state !== 'DELETE' && handleViewConfig(config.id);
                                  }
                                ">
                                {{ fileAP(config) }}
                              </div>
                            </template>
                            <div
                              v-else
                              :class="['file-name-btn', { disabled: config.file_state === 'DELETE' }]"
                              @click="
                                () => {
                                  config.file_state !== 'DELETE' && handleViewTemplateConfig(config, group);
                                }
                              ">
                              {{ fileAP(config) }}
                            </div>
                          </ContentWidthOverflowTips>
                        </td>
                        <td class="version">{{ config.is_latest ? 'latest' : config.versionName }}</td>
                        <td class="user">{{ config.creator }}</td>
                        <td class="user">{{ config.reviser }}</td>
                        <td class="datetime">{{ config.update_at }}</td>
                        <td class="status" v-if="versionData.id === 0">
                          <StatusTag :status="config.file_state" />
                        </td>
                        <td class="operation">
                          <div class="config-actions">
                            <!-- 非套餐配置文件 -->
                            <template v-if="group.id === 0">
                              <template v-if="isUnNamedVersion">
                                <template v-if="config.file_state !== 'DELETE'">
                                  <bk-button
                                    v-cursor="{ active: !hasEditServicePerm }"
                                    text
                                    theme="primary"
                                    :class="{ 'bk-text-with-no-perm': !hasEditServicePerm }"
                                    :disabled="!hasEditServicePerm"
                                    @click="handleEditOpen(config)">
                                    {{ t('编辑') }}
                                  </bk-button>
                                  <bk-button
                                    v-if="config.file_state === 'REVISE'"
                                    v-cursor="{ active: !hasEditServicePerm }"
                                    text
                                    theme="primary"
                                    :class="{ 'bk-text-with-no-perm': !hasEditServicePerm }"
                                    :disabled="!hasEditServicePerm"
                                    @click="handleUnModify(config.id)">
                                    {{ t('撤销') }}
                                  </bk-button>
                                  <DownloadConfigBtn
                                    type="config"
                                    :bk-biz-id="props.bkBizId"
                                    :app-id="props.appId"
                                    :id="config.id"
                                    :disabled="config.file_state === 'DELETE'" />
                                  <bk-button
                                    v-cursor="{ active: !hasEditServicePerm }"
                                    text
                                    theme="primary"
                                    :class="{ 'bk-text-with-no-perm': !hasEditServicePerm }"
                                    :disabled="!hasEditServicePerm"
                                    @click="handleDel(config)">
                                    {{ t('删除') }}
                                  </bk-button>
                                </template>
                                <bk-button
                                  v-else
                                  v-cursor="{ active: !hasEditServicePerm }"
                                  text
                                  theme="primary"
                                  :class="{ 'bk-text-with-no-perm': !hasEditServicePerm }"
                                  :disabled="!hasEditServicePerm"
                                  @click="handleUnDelete(config)">
                                  {{ t('恢复') }}
                                </bk-button>
                              </template>
                              <template v-else>
                                <bk-button text theme="primary" @click="handleViewConfig(config.id)">
                                  {{ t('查看') }}
                                </bk-button>
                                <bk-button
                                  v-if="versionData.status.publish_status !== 'editing'"
                                  text
                                  theme="primary"
                                  @click="handleConfigDiff(group.id, config)">
                                  {{ t('对比') }}
                                </bk-button>
                                <DownloadConfigBtn
                                  type="config"
                                  :bk-biz-id="props.bkBizId"
                                  :app-id="props.appId"
                                  :id="config.id" />
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
                                @click="handleOpenReplaceVersionDialog(group.id, config)">
                                {{ t('替换版本') }}
                              </bk-button>
                              <template v-else>
                                <bk-button text theme="primary" @click="handleViewTemplateConfig(config, group)">
                                  {{ t('查看') }}
                                </bk-button>
                                <bk-button
                                  v-if="versionData.status.publish_status !== 'editing'"
                                  text
                                  theme="primary"
                                  @click="handleConfigDiff(group.id, config)">
                                  {{ t('对比') }}
                                </bk-button>
                              </template>
                              <DownloadConfigBtn
                                type="template"
                                :bk-biz-id="props.bkBizId"
                                :app-id="props.appId"
                                :id="config.versionId"
                                :disabled="config.file_state === 'DELETE'" />
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
          <tr v-else>
            <td :colspan="colsLen">
              <TableEmpty :is-search-empty="isSearchEmpty" @clear="emits('clearStr')" style="width: 100%" />
            </td>
          </tr>
        </tbody>
      </table>
    </bk-loading>
  </div>
  <edit-config
    v-model:show="editPanelShow"
    :config-id="activeConfig"
    :bk-biz-id="props.bkBizId"
    :app-id="props.appId"
    @confirm="handleEditConfigConfirm" />
  <ViewConfig
    v-model:show="viewConfigSliderData.open"
    v-bind="viewConfigSliderData.data"
    :bk-biz-id="props.bkBizId"
    :app-id="props.appId"
    :version-id="versionData.id"
    @open-edit="handleSwitchToEdit" />
  <VersionDiff v-model:show="isDiffPanelShow" :current-version="versionData" :selected-config="diffConfig" />
  <ReplaceTemplateVersion
    v-model:show="replaceDialogData.open"
    v-bind="replaceDialogData.data"
    :binding-id="bindingId"
    :bk-biz-id="props.bkBizId"
    :app-id="props.appId"
    @updated="getAllConfigList" />
  <DeleteConfirmDialog
    v-model:isShow="isDeleteConfigDialogShow"
    :title="t('确认删除该配置文件？')"
    @confirm="handleDeleteConfigConfirm">
    <div style="margin-bottom: 8px">
      {{ t('配置文件') }}：<span style="color: #313238">{{ fileAP(deleteConfig!) }}</span>
    </div>
    <div>{{ deleteConfigTips }}</div>
  </DeleteConfirmDialog>
  <DeleteConfirmDialog
    v-model:isShow="isDeletePkgDialogShow"
    :title="t('确认移除该配置模板套餐？')"
    :confirm-text="t('移除')"
    @confirm="handleDeletePkgConfirm">
    <div style="margin-bottom: 8px">
      {{ t('配置模板套餐') }}: <span style="color: #313238">{{ deleteTemplatePkgName }}</span>
    </div>
    <div>{{ t('移除后本服务配置将不再引用该配置模板套餐，以后需要时可以重新从配置模板导入') }}</div>
  </DeleteConfirmDialog>
  <DeleteConfirmDialog
    v-model:isShow="isRecoverConfigDialogShow"
    :title="t('确认恢复该配置文件?')"
    :confirm-text="t('恢复')"
    @confirm="handleRecoverConfigConfirm">
    <div style="margin-bottom: 8px">
      {{ t('配置文件') }}：<span style="color: #313238">{{ fileAP(recoverConfig!) }}</span>
    </div>
    <div>{{ t(`配置文件恢复后，将覆盖新添加的配置文件`) + fileAP(recoverConfig!) }}</div>
  </DeleteConfirmDialog>
</template>
<script lang="ts" setup>
  import { ref, computed, watch, onMounted, onBeforeUnmount } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { storeToRefs } from 'pinia';
  import Message from 'bkui-vue/lib/message';
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
    unModifyConfigItem,
    unDeleteConfigItem,
  } from '../../../../../../../../api/config';
  import { getAppPkgBindingRelations } from '../../../../../../../../api/template';
  import StatusTag from './status-tag';
  import EditConfig from '../edit-config.vue';
  import ViewConfig from '../view-config.vue';
  import VersionDiff from '../../../components/version-diff/index.vue';
  import ReplaceTemplateVersion from '../replace-template-version.vue';
  import TableEmpty from '../../../../../../../../components/table/table-empty.vue';
  import DeleteConfirmDialog from '../../../../../../../../components/delete-confirm-dialog.vue';
  import TableFilter from '../../../components/table-filter.vue';
  import DownloadConfigBtn from '../download-config-btn.vue';
  import ContentWidthOverflowTips from '../../../../../../../../components/content-width-overflow-tips/index.vue';

  interface IConfigsGroupData {
    id: number;
    name: string;
    expand: boolean;
    configs: IConfigTableItem[];
    template_space_id?: number;
    template_space_name?: string;
    template_set_id?: number;
    template_set_name?: string;
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
    is_conflict: boolean;
    is_latest?: boolean;
  }

  interface ITemplateConfigMeta {
    template_space_id: number;
    template_space_name: string;
    template_set_id: number;
    template_set_name: string;
  }

  const { t } = useI18n();
  const configStore = useConfigStore();
  const serviceStore = useServiceStore();
  const { versionData, allConfigCount } = storeToRefs(configStore);
  const { checkPermBeforeOperate } = serviceStore;
  const { permCheckLoading, hasEditServicePerm, topIds } = storeToRefs(serviceStore);

  const props = defineProps<{
    bkBizId: string;
    appId: number;
    searchStr: string;
  }>();

  const emits = defineEmits(['clearStr', 'deleteConfig', 'updateSelectedIds', 'updateSelectedItems']);

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
  const recoverConfig = ref<IConfigTableItem>();
  const isRecoverConfigDialogShow = ref(false);
  const oldConfigIndex = ref(-1);
  const isDeletePkgDialogShow = ref(false);
  const deleteTemplatePkgName = ref('');
  const deleteTemplatePkgId = ref(0);
  const statusFilterChecked = ref<string[]>([]);
  const viewConfigSliderData = ref<{
    open: boolean;
    data: {
      id: number;
      type: string;
      templateMeta?: ITemplateConfigMeta;
      versionName?: string;
      isLatest?: boolean;
    };
  }>({
    open: false,
    data: {
      id: 0,
      type: '',
    },
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
  const stickyIndex = ref(1);
  const tableRef = ref();
  const collapseHeader = ref();
  const selectedConfigItems = ref<IConfigItem[]>([]);

  // 是否为未命名版本
  const isUnNamedVersion = computed(() => versionData.value.id === 0);

  // 表格列长度
  const colsLen = computed(() => (isUnNamedVersion.value ? 9 : 8));

  // 全选checkbox选中状态
  const isIndeterminate = computed(() => {
    return selectedConfigItems.value.length > 0 && selectedConfigItems.value.length <= configsCount.value;
  });

  const deleteConfigTips = computed(() => {
    if (deleteConfig.value) {
      return deleteConfig.value.file_state === 'ADD'
        ? t('一旦删除，该操作将无法撤销，请谨慎操作')
        : t('配置文件删除后，可以通过恢复按钮撤销删除');
    }
    return '';
  });

  // 配置文件绝对路径
  const fileAP = (config: IConfigTableItem) => {
    const { path, name } = config;
    if (path.endsWith('/')) {
      return `${path}${name}`;
    }
    return `${path}/${name}`;
  };

  // 状态过滤列表
  const statusFilterList = computed(() => {
    return [
      {
        value: 'ADD',
        text: t('新增'),
      },
      {
        value: 'REVISE',
        text: t('修改'),
      },
      {
        value: 'DELETE',
        text: t('删除'),
      },
      {
        value: 'UNCHANGE',
        text: t('无修改'),
      },
    ];
  });

  watch(
    () => versionData.value.id,
    async () => {
      await getBindingId();
      getAllConfigList();
      selectedConfigItems.value = [];
      emits('updateSelectedIds', []);
    },
  );

  watch(
    () => props.searchStr,
    () => {
      props.searchStr ? (isSearchEmpty.value = true) : (isSearchEmpty.value = false);
      getAllConfigList();
    },
  );

  watch(
    [() => configsCount.value, () => templatesCount.value],
    () => {
      configStore.$patch((state) => {
        state.allConfigCount = configsCount.value + templatesCount.value;
      });
    },
    { immediate: true },
  );

  onMounted(async () => {
    tableRef.value.addEventListener('scroll', handleScroll);
    await getBindingId();
    getAllConfigList();
  });

  onBeforeUnmount(() => {
    tableRef.value.removeEventListener('scroll', handleScroll);
  });

  const getBindingId = async () => {
    const res = await getAppPkgBindingRelations(props.bkBizId, props.appId);
    bindingId.value = res.details.length === 1 ? res.details[0].id : 0;
  };

  const getAllConfigList = async (createConfig = false) => {
    loading.value = true;
    await Promise.all([getCommonConfigList(createConfig), getBoundTemplateList()]);
    loading.value = false;
    tableGroupsData.value = transListToTableData();
  };

  // 获取非模板配置文件列表
  const getCommonConfigList = async (createConfig = false) => {
    commonConfigListLoading.value = true;
    try {
      const params: ICommonQuery = {
        start: 0,
        all: true,
      };
      if (!createConfig) topIds.value = [];
      if (topIds.value.length > 0) params.ids = topIds.value.join(',');
      let res;
      if (isUnNamedVersion.value) {
        if (props.searchStr) {
          params.search_fields = 'name,path,memo,creator,reviser';
          params.search_value = props.searchStr;
        }
        if (statusFilterChecked.value.length > 0) params.status = statusFilterChecked.value;
        res = await getConfigList(props.bkBizId, props.appId, params);
      } else {
        if (props.searchStr) {
          params.search_fields = 'name,path,memo,creator';
          params.search_value = props.searchStr;
        }
        res = await getReleasedConfigList(props.bkBizId, props.appId, versionData.value.id, params);
      }
      configList.value = res.details.sort((a: IConfigItem, b: IConfigItem) => {
        if (a.file_state === 'DELETE' && b.file_state !== 'DELETE') {
          return 1;
        }
        if (a.file_state !== 'DELETE' && b.file_state === 'DELETE') {
          return -1;
        }
        return 0;
      });
      configsCount.value = res.count;
      configStore.$patch((state) => {
        state.conflictFileCount = res.conflict_number || 0;
      });
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
        if (statusFilterChecked.value.length > 0) params.status = statusFilterChecked.value;
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
      { id: 0, name: t('非模板配置'), expand: true, configs: transConfigsToTableItemData(configList.value) },
      ...pkgsGroups,
    ];
  };

  // 将非模板配置文件数据转为表格数据
  const transConfigsToTableItemData = (list: IConfigItem[]) =>
    list.map((item: IConfigItem) => {
      const { id, spec, revision, file_state, is_conflict } = item;
      const { name, path, permission } = spec;
      const { creator, reviser, update_at, create_at } = revision;
      return {
        id,
        name,
        versionId: 0,
        versionName: '--',
        path,
        creator,
        reviser,
        update_at: datetimeFormat(update_at || create_at),
        file_state,
        permission,
        is_conflict,
        is_latest: false,
      };
    });

  // 将模板按套餐分组，并将模板数据格式转为表格数据
  const groupTplsByPkg = (list: IBoundTemplateGroup[]) => {
    const groups: IConfigsGroupData[] = list.map((groupItem) => {
      const { template_space_name, template_set_id, template_set_name, template_revisions, template_space_id } =
        groupItem;
      const group: IConfigsGroupData = {
        id: template_set_id,
        template_set_id,
        template_set_name,
        template_space_name,
        template_space_id,
        name: `${template_space_name} - ${template_set_name}`,
        expand: true,
        configs: [],
      };
      template_revisions.forEach((tpl) => {
        const {
          template_id,
          template_revision_id,
          name,
          template_revision_id: versionId,
          template_revision_name: versionName,
          path,
          creator,
          create_at,
          file_state,
          is_conflict,
          is_latest,
        } = tpl;
        group.configs.push({
          id: isUnNamedVersion.value ? template_id : template_revision_id,
          name,
          versionId,
          versionName,
          path,
          creator,
          reviser: creator,
          update_at: datetimeFormat(create_at),
          file_state,
          is_conflict,
          is_latest,
        });
      });
      return group;
    });
    return groups;
  };

  // 全选
  const handleSelectAll = (val: boolean) => {
    if (val) {
      selectedConfigItems.value = configList.value.filter((item) => item.file_state !== 'DELETE');
    } else {
      selectedConfigItems.value = [];
    }
    emits(
      'updateSelectedIds',
      selectedConfigItems.value.map((item) => item.id),
    );
    emits('updateSelectedItems', selectedConfigItems.value);
  };

  // 非模板配置选择/取消选择
  const handleRowSelectionChange = (val: boolean, id: number) => {
    const index = selectedConfigItems.value.findIndex((item) => item.id === id);
    if (val) {
      index === -1 && selectedConfigItems.value.push(configList.value.find((item) => item.id === id)!);
    } else {
      index > -1 && selectedConfigItems.value.splice(index, 1);
    }
    emits(
      'updateSelectedIds',
      selectedConfigItems.value.map((item) => item.id),
    );
    emits('updateSelectedItems', selectedConfigItems.value);
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

  // 查看配置文件
  const handleViewConfig = (id: number) => {
    viewConfigSliderData.value = {
      open: true,
      data: { id, type: 'config' },
    };
  };

  // 查看配置模板版本
  const handleViewTemplateConfig = (config: IConfigTableItem, group: IConfigsGroupData) => {
    const { id, versionName, is_latest } = config;
    const { template_set_id, template_space_id, template_set_name, template_space_name } = group;
    const templateMeta = { template_space_id, template_space_name, template_set_id, template_set_name };
    viewConfigSliderData.value = {
      open: true,
      data: {
        id,
        versionName,
        templateMeta: templateMeta as ITemplateConfigMeta,
        type: 'template',
        isLatest: is_latest,
      },
    };
  };

  // 由查看态切换为编辑态
  const handleSwitchToEdit = () => {
    if (permCheckLoading.value || !checkPermBeforeOperate('update')) {
      return;
    }
    if (!permCheckLoading.value && checkPermBeforeOperate('update')) {
      viewConfigSliderData.value.open = false;
      activeConfig.value = viewConfigSliderData.value.data.id;
      editPanelShow.value = true;
    }
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
      message: t('移除模板套餐成功'),
    });
    isDeletePkgDialogShow.value = false;
    getAllConfigList();
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
    Message({
      theme: 'success',
      message: t('删除配置文件成功'),
    });
    await getAllConfigList();
    emits('deleteConfig');
    isDeleteConfigDialogShow.value = false;
  };

  // 设置新增行的标记class
  const getRowCls = (data: IConfigTableItem) => {
    if (data.file_state === 'DELETE') {
      return 'delete-row config-row';
    }
    if (data.is_conflict) {
      return 'conflict-row config-row';
    }
    if (topIds.value.includes(data.id)) {
      return 'new-row-marked config-row';
    }
    return 'config-row';
  };

  // 变更状态过滤
  const handleStatusFilterSelected = (filterStatus: string[]) => {
    statusFilterChecked.value = filterStatus;
    getAllConfigList();
  };

  // 配置文件撤销修改
  const handleUnModify = async (id: number) => {
    if (permCheckLoading.value || !checkPermBeforeOperate('update')) {
      return;
    }
    await unModifyConfigItem(props.bkBizId, props.appId, id);
    Message({ theme: 'success', message: t('撤销修改配置文件成功') });
    getAllConfigList();
  };

  // 配置文件恢复删除
  const handleUnDelete = async (config: IConfigTableItem) => {
    if (permCheckLoading.value || !checkPermBeforeOperate('update')) {
      return;
    }
    recoverConfig.value = config;
    const configs = tableGroupsData.value.find((group) => group.id === 0)?.configs;
    oldConfigIndex.value = configs!.findIndex(
      (item) => fileAP(item) === fileAP(config) && item.file_state !== 'DELETE',
    );
    if (oldConfigIndex.value === -1) {
      handleRecoverConfigConfirm();
    } else {
      isRecoverConfigDialogShow.value = true;
    }
  };

  const handleRecoverConfigConfirm = async () => {
    await unDeleteConfigItem(props.bkBizId, props.appId, recoverConfig.value!.id);
    isRecoverConfigDialogShow.value = false;
    Message({ theme: 'success', message: t('恢复配置文件成功') });

    // 获取冲突的模板套裁数据 直接覆盖
    await getBoundTemplateList();
    tableGroupsData.value = transListToTableData();
    tableGroupsData.value
      .find((group) => group.id === 0)!
      .configs.find((config) => config.id === recoverConfig.value!.id)!.file_state = 'UNCHANGE';

    // 获取冲突的非模板配置数据
    await getCommonConfigList();
    const conflictFileIds = configList.value.filter((config) => config.is_conflict).map((config) => config.id);
    tableGroupsData.value
      .find((group) => group.id === 0)
      ?.configs.forEach((config) => {
        if (conflictFileIds.includes(config.id)) {
          config.is_conflict = true;
        }
      });
    if (oldConfigIndex.value !== -1) {
      const configs = tableGroupsData.value.find((group) => group.id === 0)?.configs;
      configs!.splice(oldConfigIndex.value, 1);
    }
  };

  // 批量操作配置项后刷新配置项列表
  const refreshAfterBatchSet = () => {
    selectedConfigItems.value = [];
    emits('updateSelectedIds', []);
    emits('updateSelectedItems', []);
    getAllConfigList();
  };

  // 监听表格滚动事件
  const handleScroll = () => {
    const tableRect = tableRef.value.getBoundingClientRect();
    collapseHeader.value.forEach((header: Element, index: number) => {
      const headerRect = header.getBoundingClientRect();
      if (headerRect.top - 82 <= tableRect.top && headerRect.bottom >= tableRect.top) {
        stickyIndex.value = index;
      }
    });
  };

  defineExpose({
    refreshAfterBatchSet,
    refresh: getAllConfigList,
  });
</script>
<style lang="scss" scoped>
  .table-container {
    max-height: 100%;
    border: 1px solid #dcdee5;
    overflow: auto;
  }
  .config-groups-table {
    width: 100%;
    border-collapse: collapse;
    // table-layout: fixed;
    > tbody tr:last-child > td:before {
      background: none;
    }
    .config-groups-table-tr {
      background: #ffffff;
      th,
      td {
        position: relative;
        // 底部border
        &:before {
          content: '';
          display: block;
          position: absolute;
          bottom: 0;
          right: 0;
          height: 1px;
          width: 100%;
          background: linear-gradient(0deg, transparent 50%, #dcdee5 50%);
        }
      }
      th {
        position: sticky;
        top: 0;
        padding: 11px 16px;
        color: #313238;
        font-weight: normal;
        font-size: 12px;
        line-height: 20px;
        text-align: left;
        background: #fafbfd;
        z-index: 3;
      }
      &.group-title-row {
        &.expand {
          box-shadow: 0 5px 10px -5px rgba(0, 0, 0, 0.12);
        }
        &:hover {
          background: #f5f7fa;
        }
      }
      &.sticky {
        position: sticky;
        top: 43px;
        z-index: 100;
      }
    }
    .config-groups-table-td {
      padding: 0;
      font-size: 12px;
      line-height: 20px;
      text-align: left;
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
    .selection {
      width: 50px;
    }
    .name {
      white-space: nowrap;
    }
    .version {
      width: 200px;
    }
    .path {
      width: 331px;
    }
    .user {
      width: 120px;
    }
    .datetime {
      width: 158px;
    }
    .status {
      width: 140px;
    }
    .operation {
      width: 150px;
    }
    .exception-tips {
      margin: 20px 0;
    }
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
      font-size: 12px;
      line-height: 20px;
      color: #63656e;
      white-space: nowrap;
      text-overflow: ellipsis;
      overflow: hidden;
    }
    .file-name-btn {
      color: #3a84ff;
      cursor: pointer;
      overflow: hidden;
      text-overflow: ellipsis;
      &:hover {
        color: #5594fa;
      }
      &.disabled {
        color: #c4c6cc;
        cursor: not-allowed;
      }
    }
    .config-actions {
      .bk-button:not(:last-child) {
        margin-right: 8px;
      }
    }
  }
  .new-row-marked td {
    background: #f2fff4 !important;
  }
  .conflict-row td {
    background-color: #fff3e1 !important;
  }
  .delete-row td {
    background: #fafbfd !important;
    color: #c4c6cc !important;
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
