<template>
  <div :class="['configs-menu', { 'search-opened': isOpenSearch }]">
    <div class="title-area">
      <div class="title">{{ t('配置文件') }}</div>
      <div class="title-extend">
        <bk-checkbox
          v-if="isBaseVersionExist"
          v-model="isOnlyShowDiff"
          class="view-diff-checkbox"
          @change="handleToggleShowDiff">
          {{ t('只查看差异文件') }}({{ diffCount }})
        </bk-checkbox>
        <div :class="['search-trigger', { actived: isOpenSearch }]" @click="isOpenSearch = !isOpenSearch">
          <Search />
        </div>
      </div>
    </div>
    <div v-if="isOpenSearch" class="search-wrapper">
      <SearchInput v-model="searchStr" :placeholder="t('搜索配置文件名称')" @search="handleSearch" />
    </div>
    <div class="groups-wrapper">
      <div v-for="group in groupedConfigListOnShow" class="config-group-item" :key="group.id">
        <div :class="['group-header', { expand: group.expand }]" @click="group.expand = !group.expand">
          <RightShape class="arrow-icon" />
          <span v-overflow-title class="name">{{ group.name }}</span>
        </div>
        <div v-if="group.expand" class="config-list">
          <div
            v-for="config in group.configs"
            v-overflow-title
            :key="config.id"
            :class="['config-item', { actived: getItemSelectedStatus(group.id, config) }]"
            @click="
              handleSelectItem({
                pkgId: group.id,
                id: config.id,
                version: config.template_revision_id,
                permission: config.permission,
              })
            ">
            <i v-if="config.diffType" :class="['status-icon', config.diffType]"></i>
            {{ config.name }}
          </div>
        </div>
      </div>
      <tableEmpty
        v-if="groupedConfigListOnShow.length === 0"
        class="empty-tips"
        :is-search-empty="isSearchEmpty"
        :empty-title="t('没有差异配置文件')"
        @clear="clearStr">
      </tableEmpty>
    </div>
  </div>
</template>
<script lang="ts" setup>
  import { ref, computed, watch, onMounted } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { useRoute } from 'vue-router';
  import { storeToRefs } from 'pinia';
  import { Search, RightShape } from 'bkui-vue/lib/icon';
  import useServiceStore from '../../../../../../../../store/service';
  import { datetimeFormat, byteUnitConverse } from '../../../../../../../../utils';
  import { ICommonQuery } from '../../../../../../../../../types/index';
  import {
    IConfigItem,
    IBoundTemplateGroup,
    IConfigDiffSelected,
    IFileConfigContentSummary,
  } from '../../../../../../../../../types/config';

  import { IVariableEditParams } from '../../../../../../../../../types/variable';
  import {
    getConfigList,
    getReleasedConfigList,
    downloadConfigContent,
    getBoundTemplates,
    getBoundTemplatesByAppVersion,
  } from '../../../../../../../../api/config';
  import { getReleasedAppVariables } from '../../../../../../../../api/variable';

  import SearchInput from '../../../../../../../../components/search-input.vue';
  import tableEmpty from '../../../../../../../../components/table/table-empty.vue';

  interface IConfigMenuItem {
    type: string;
    id: number;
    name: string;
    file_type: string;
    file_state: string;
    update_at: string;
    byte_size: string;
    signature: string;
    template_revision_id: number;
    permission?: IPermissionType;
    diffSignature?: string;
  }

  interface IPermissionType {
    privilege: string;
    user: string;
    user_group: string;
  }

  interface IConfigDiffItem extends IConfigMenuItem {
    diffType: string;
    current: string;
    base: string;
    currentPermission?: IPermissionType;
    basePermission?: IPermissionType;
  }

  interface IConfigsGroupData {
    template_space_id: number;
    id: number;
    name: string;
    expand: boolean;
    configs: IConfigMenuItem[];
  }

  interface IDiffGroupData extends IConfigsGroupData {
    configs: IConfigDiffItem[];
  }

  const { t } = useI18n();
  const props = withDefaults(
    defineProps<{
      currentVersionId: number;
      unNamedVersionVariables?: IVariableEditParams[]; // 未命名版本变量列表
      baseVersionId: number | undefined;
      selectedConfig: IConfigDiffSelected;
      actived: boolean;
      isPublish: boolean;
    }>(),
    {
      unNamedVersionVariables: () => [],
      selectedConfig: () => ({ pkgId: 0, id: 0, version: 0 }),
    },
  );

  const emits = defineEmits(['selected']);

  const route = useRoute();
  const bkBizId = ref(String(route.params.spaceId));
  const { appData } = storeToRefs(useServiceStore());

  const diffCount = ref(0);
  const selected = ref<IConfigDiffSelected>({ pkgId: 0, id: 0, version: 0 });
  const currentGroupList = ref<IConfigsGroupData[]>([]);
  const currentVariables = ref<IVariableEditParams[]>([]);
  const baseGroupList = ref<IConfigsGroupData[]>([]);
  const baseVariables = ref<IVariableEditParams[]>([]);
  // 汇总的配置文件列表，包含未修改、增加、删除、修改的所有配置文件
  const aggregatedList = ref<IDiffGroupData[]>([]);
  // 分组后需要展示的配置文件列表
  const groupedConfigListOnShow = ref<IDiffGroupData[]>([]);
  const isOnlyShowDiff = ref(true); // 只显示差异项
  const isOpenSearch = ref(false);
  const searchStr = ref('');
  const isSearchEmpty = ref(false);

  // 是否实际选择了对比的基准版本，为了区分的未命名版本id为0的情况
  const isBaseVersionExist = computed(() => typeof props.baseVersionId === 'number');

  // 只包含差异文件的配置文件列表
  const aggregatedListOfDiff = computed(() => {
    const list: IDiffGroupData[] = [];
    aggregatedList.value.forEach((group) => {
      const configs = group.configs.filter((item) => item.diffType !== '');
      if (configs.length > 0) {
        list.push({
          ...group,
          configs,
        });
      }
    });
    return list;
  });

  // 基准版本变化，更新选中对比项
  watch(
    () => props.baseVersionId,
    async () => {
      baseGroupList.value = await getConfigsOfVersion(props.baseVersionId);
      baseVariables.value = await getVariableList(props.baseVersionId);
      aggregatedList.value = calcDiff();
      groupedConfigListOnShow.value = getMenuList();
      setDefaultSelected();
    },
  );

  // 当前版本默认选中的配置文件
  watch(
    () => props.selectedConfig,
    (val) => {
      if (val) {
        selected.value = { ...val };
      }
    },
    {
      immediate: true,
    },
  );

  onMounted(async () => {
    await getAllConfigList();
    // 未命名版本变量取正在编辑中的变量列表
    if (isUnNamedVersion(props.currentVersionId)) {
      currentVariables.value = props.unNamedVersionVariables;
    } else {
      currentVariables.value = await getVariableList(props.currentVersionId);
    }
    baseVariables.value = await getVariableList(props.baseVersionId);
    aggregatedList.value = calcDiff();
    isOnlyShowDiff.value = aggregatedListOfDiff.value.length > 0;
    groupedConfigListOnShow.value = getMenuList();
    setDefaultSelected();
  });

  // 判断版本是否为未命名版本
  const isUnNamedVersion = (id: number) => id === 0;

  // 获取当前版本和基准版本的所有配置文件列表(非模板配置和套餐下模板)
  const getAllConfigList = async () => {
    currentGroupList.value = await getConfigsOfVersion(props.currentVersionId);
    baseGroupList.value = await getConfigsOfVersion(props.baseVersionId);
  };

  // 获取某一版本下配置文件和模板列表
  const getConfigsOfVersion = async (releaseId: number | undefined) => {
    if (typeof releaseId !== 'number') {
      return [];
    }

    const [commonConfigList, templateList] = await Promise.all([
      getCommonConfigList(releaseId),
      getBoundTemplateList(releaseId),
    ]);

    return commonConfigList.concat(templateList);
  };

  // 获取非模板配置文件列表
  const getCommonConfigList = async (id: number): Promise<IConfigsGroupData[]> => {
    const unNamedVersion = isUnNamedVersion(id);
    const params: ICommonQuery = {
      start: 0,
      all: true,
    };
    let configsRes;

    if (unNamedVersion) {
      configsRes = await getConfigList(bkBizId.value, appData.value.id as number, params);
    } else {
      configsRes = await getReleasedConfigList(bkBizId.value, appData.value.id as number, id, params);
    }

    // 未命名版本中包含被删除的配置文件，需要过滤掉
    const configs: IConfigItem[] = configsRes.details.filter((item: IConfigItem) => item.file_state !== 'DELETE');

    return [
      {
        template_space_id: 0,
        id: 0,
        name: '非模板配置',
        expand: true,
        configs: configs.map((config) => {
          const { id, spec, commit_spec, revision, file_state } = config;
          const { name, file_type, permission } = spec;
          const { origin_byte_size, byte_size, signature, origin_signature } = commit_spec.content;
          return {
            type: 'config',
            id,
            name,
            file_type,
            file_state,
            update_at: datetimeFormat(revision.update_at || revision.create_at),
            byte_size: unNamedVersion ? byte_size : origin_byte_size,
            signature: unNamedVersion ? signature : origin_signature,
            template_revision_id: 0,
            permission,
            diffSignature: signature,
          };
        }),
      },
    ];
  };

  // 获取模板配置文件列表
  const getBoundTemplateList = async (id: number) => {
    const unNamedVersion = isUnNamedVersion(id);
    const params: ICommonQuery = {
      start: 0,
      all: true,
    };
    let res;
    if (unNamedVersion) {
      res = await getBoundTemplates(bkBizId.value, appData.value.id as number, params);
    } else {
      res = await getBoundTemplatesByAppVersion(bkBizId.value, appData.value.id as number, id, params);
    }
    return res.details.map((groupItem: IBoundTemplateGroup) => {
      const { template_space_id, template_space_name, template_set_id, template_set_name } = groupItem;
      const group: IConfigsGroupData = {
        template_space_id,
        id: template_set_id,
        name: `${template_space_name} - ${template_set_name}`,
        expand: false,
        configs: [],
      };
      groupItem.template_revisions.forEach((tpl) => {
        const {
          template_id,
          name,
          file_type,
          file_state,
          origin_byte_size,
          byte_size,
          origin_signature,
          signature,
          template_revision_id,
          create_at,
        } = tpl;
        if (file_state !== 'DELETE') {
          group.configs.push({
            type: 'template',
            id: template_id,
            name,
            file_type,
            file_state,
            update_at: datetimeFormat(create_at),
            byte_size: unNamedVersion ? byte_size : origin_byte_size,
            signature: unNamedVersion ? signature : origin_signature,
            template_revision_id,
          });
        }
      });
      return group;
    });
  };

  // 获取版本下变量列表
  const getVariableList = async (id: number | undefined) => {
    if (id === undefined || isUnNamedVersion(id)) {
      return [];
    }
    const res = await getReleasedAppVariables(bkBizId.value, appData.value.id as number, id);
    return res.details;
  };

  // 计算配置被修改、被删除、新增的差异
  const calcDiff = () => {
    diffCount.value = 0;
    const list: IDiffGroupData[] = [];
    currentGroupList.value.forEach((currentGroupItem) => {
      const { template_space_id, id, name, expand, configs } = currentGroupItem;
      const diffGroup: IDiffGroupData = { template_space_id, id, name, expand, configs: [] };
      configs.forEach((crtItem) => {
        let baseItem: IConfigMenuItem | undefined;
        baseGroupList.value.some((baseGroupItem) => {
          if (baseGroupItem.template_space_id === currentGroupItem.template_space_id) {
            return baseGroupItem.configs.some((config) => {
              if (config.id === crtItem.id) {
                baseItem = config;
                return true;
              }
              return false;
            });
          }
          return false;
        });
        if (baseItem) {
          // 修改项
          const diffConfig = {
            ...crtItem,
            diffType: '',
            base: baseItem.signature,
            basePermission: baseItem.permission,
            baseContent: baseItem.diffSignature,
            current: crtItem.signature,
            currentPermission: crtItem.permission,
            currentContent: crtItem.diffSignature,
          };
          if (
            crtItem.template_revision_id !== baseItem.template_revision_id ||
            diffConfig.current !== diffConfig.base ||
            diffConfig.currentPermission?.privilege !== diffConfig.basePermission?.privilege ||
            diffConfig.currentPermission?.user !== diffConfig.basePermission?.user ||
            diffConfig.currentPermission?.user_group !== diffConfig.basePermission?.user_group ||
            diffConfig.currentContent !== diffConfig.baseContent
          ) {
            diffCount.value += 1;
            diffConfig.diffType = isBaseVersionExist.value ? 'modify' : '';
          }
          diffGroup.configs.push(diffConfig);
        } else {
          // 当前版本新增项
          diffCount.value += 1;
          diffGroup.configs.push({
            ...crtItem,
            diffType: isBaseVersionExist.value ? 'add' : '',
            current: crtItem.signature,
            currentPermission: crtItem.permission,
            base: '',
            basePermission: { privilege: '', user: '', user_group: '' },
          });
        }
      });
      if (diffGroup.configs.length > 0) {
        list.push(diffGroup);
      }
    });
    // 计算当前版本删除项
    baseGroupList.value.forEach((baseGroupItem) => {
      const { template_space_id, id, name, expand, configs } = baseGroupItem;
      const groupIndex = list.findIndex((item) => item.id === baseGroupItem.id);
      const diffGroup: IDiffGroupData =
        groupIndex > -1 ? list[groupIndex] : { template_space_id, id, name, expand, configs: [] };

      configs.forEach((baseItem) => {
        let currentItem: IConfigMenuItem | undefined;
        currentGroupList.value.some((baseGroupItem) => {
          if (baseGroupItem.template_space_id === baseGroupItem.template_space_id) {
            return baseGroupItem.configs.some((config) => {
              if (config.id === baseItem.id) {
                currentItem = config;
                return true;
              }
              return undefined;
            });
          }
          return false;
        });
        if (!currentItem) {
          diffCount.value += 1;
          diffGroup.configs.push({
            ...baseItem,
            diffType: isBaseVersionExist.value ? 'delete' : '',
            current: '',
            currentPermission: { privilege: '', user: '', user_group: '' },
            base: baseItem.signature,
            basePermission: baseItem.permission,
          });
        }
      });

      if (groupIndex === -1 && diffGroup.configs.length > 0) {
        list.push(diffGroup);
      }
    });
    return list;
  };

  const getMenuList = () => {
    const fullList = isOnlyShowDiff.value ? aggregatedListOfDiff.value : aggregatedList.value;
    if (searchStr.value !== '') {
      const searchedList: IDiffGroupData[] = [];
      fullList.forEach((group) => {
        const configs = group.configs.filter((item) => {
          const isSearchHit = item.name.toLocaleLowerCase().includes(searchStr.value.toLocaleLowerCase());
          if (isOnlyShowDiff.value) {
            return item.diffType !== '' && isSearchHit;
          }
          return isSearchHit;
        });
        if (configs.length > 0) {
          searchedList.push({
            ...group,
            configs,
          });
        }
      });
      return searchedList;
    }
    return fullList.slice();
  };

  // 设置默认选中的配置文件
  // 如果props有设置选中项，取props值
  // 如果选中项有值，保持上一次选中项
  // 否则取第一个非空分组的第一个配置文件
  const setDefaultSelected = () => {
    if (props.selectedConfig.id) {
      const pkg = aggregatedList.value.find((group) => group.id === props.selectedConfig.pkgId);
      if (pkg) {
        pkg.expand = true;
      }
      handleSelectItem(props.selectedConfig);
    } else {
      const selectedGroup = aggregatedList.value.find((group) => group.id === selected.value.pkgId);
      if (selectedGroup) {
        const selectedConfig = selectedGroup.configs.find((config) => config.id === selected.value.id);
        if (selectedConfig) {
          handleSelectItem(selected.value);
          return;
        }
      }
      const group = groupedConfigListOnShow.value.find((group) => group.configs.length > 0);
      if (group) {
        handleSelectItem({
          pkgId: group.id,
          id: group.configs[0].id,
          version: group.configs[0].template_revision_id,
          permission: group.configs[0].permission,
        });
      }
    }
  };

  const handleSearch = () => {
    groupedConfigListOnShow.value = getMenuList();
    isSearchEmpty.value = searchStr.value !== '' && groupedConfigListOnShow.value.length === 0;
  };

  const getItemSelectedStatus = (pkgId: number, config: IConfigDiffItem) => {
    const { id, template_revision_id } = config;
    return (
      props.actived &&
      pkgId === selected.value.pkgId &&
      id === selected.value.id &&
      template_revision_id === selected.value.version
    );
  };

  // 选择对比配置文件后，加载配置文件详情，组装对比数据
  const handleSelectItem = async (selectedConfig: IConfigDiffSelected) => {
    const pkg = aggregatedList.value.find((item) => item.id === selectedConfig.pkgId);
    if (pkg) {
      const config = pkg.configs.find((item) => {
        const res = item.id === selectedConfig.id && item.template_revision_id === selectedConfig.version;
        return res;
      });
      if (config) {
        selected.value = selectedConfig;
        const data = await getConfigDiffDetail(config);
        emits('selected', data);
      }
    }
  };

  // 切换“只查看差异文件”
  const handleToggleShowDiff = () => {
    groupedConfigListOnShow.value = getMenuList();
  };

  const getConfigDiffDetail = async (config: IConfigDiffItem) => {
    let currentConfigContent: string | IFileConfigContentSummary = '';
    let baseConfigContent: string | IFileConfigContentSummary = '';
    const {
      id,
      name,
      file_type,
      update_at,
      byte_size,
      current: currentSignature,
      base: baseSignature,
      currentPermission,
      basePermission,
    } = config;

    if (config.current) {
      currentConfigContent = await loadConfigContent({
        id,
        name,
        file_type,
        update_at,
        byte_size,
        signature: currentSignature,
      });
    }

    if (config.base) {
      baseConfigContent = await loadConfigContent({
        id,
        name,
        file_type,
        update_at,
        byte_size,
        signature: baseSignature,
      });
    }

    return {
      contentType: config.file_type === 'binary' ? 'file' : 'text',
      base: {
        content: baseConfigContent,
        variables: baseVariables.value,
        permission: basePermission,
      },
      current: {
        content: currentConfigContent,
        variables: currentVariables.value,
        permission: currentPermission,
      },
    };
  };
  // 加载配置内容详情
  const loadConfigContent = async ({
    id,
    name,
    file_type,
    update_at,
    signature,
    byte_size,
  }: {
    id: number;
    name: string;
    file_type: string;
    update_at: string;
    signature: string;
    byte_size: string;
  }) => {
    if (!signature) {
      return '';
    }
    if (file_type === 'binary') {
      return {
        id,
        name,
        signature,
        update_at,
        size: byteUnitConverse(Number(byte_size)),
      };
    }
    const configContent = await downloadConfigContent(bkBizId.value, appData.value.id as number, signature);
    return String(configContent);
  };

  // 清空筛选条件
  const clearStr = () => {
    searchStr.value = '';
    handleSearch();
  };
</script>
<style lang="scss" scoped>
  .configs-menu {
    background: #fafbfd;
    height: 100%;
    &.search-opened {
      .groups-wrapper {
        height: calc(100% - 80px);
      }
    }
  }
  .title-area {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 12px 12px 8px 24px;
    .title {
      font-size: 14px;
      color: #313238;
      font-weight: 700;
    }
    .title-extend {
      display: flex;
      align-items: center;
      .view-diff-checkbox {
        padding-right: 8px;
        border-right: 1px solid #dcdee5;
        :deep(.bk-checkbox-label) {
          font-size: 12px;
        }
      }
    }
    .search-trigger {
      display: flex;
      align-items: center;
      justify-content: center;
      margin-left: 8px;
      width: 20px;
      height: 20px;
      font-size: 12px;
      color: #63656e;
      background: #edeff1;
      border-radius: 2px;
      cursor: pointer;
      &.actived,
      &:hover {
        background: #e1ecff;
        color: #3a84ff;
      }
    }
  }
  .search-wrapper {
    padding: 0 12px 8px;
  }
  .groups-wrapper {
    height: calc(100% - 40px);
    overflow: auto;
  }
  .config-group-item {
    .group-header {
      display: flex;
      align-items: center;
      padding: 8px 12px;
      line-height: 20px;
      font-size: 12px;
      color: #313238;
      cursor: pointer;
      &.expand {
        .arrow-icon {
          transform: rotate(90deg);
          color: #3a84ff;
        }
      }
    }
    .arrow-icon {
      margin-right: 8px;
      font-size: 14px;
      color: #c4c6cc;
      transition: transform 0.2s ease-in-out;
    }
    .config-list {
      margin-bottom: 8px;
      .config-item {
        position: relative;
        padding: 0 12px 0 32px;
        height: 40px;
        line-height: 40px;
        font-size: 12px;
        color: #63656e;
        border-bottom: 1px solid #dcdee5;
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
        cursor: pointer;
        &:hover {
          background: #e1ecff;
        }
        &.actived {
          background: #e1ecff;
          color: #3a84ff;
        }
        .status-icon {
          position: absolute;
          top: 18px;
          left: 16px;
          width: 4px;
          height: 4px;
          border-radius: 50%;
          &.add {
            background: #3a84ff;
          }
          &.delete {
            background: #ea3536;
          }
          &.modify {
            background: #fe9c00;
          }
        }
      }
    }
  }
  .empty-tips {
    margin-top: 40px;
    font-size: 12px;
    color: #63656e;
  }
</style>
