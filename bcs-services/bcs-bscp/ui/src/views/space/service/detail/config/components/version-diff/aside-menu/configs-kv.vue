<template>
  <div :class="['configs-menu', { 'search-opened': isOpenSearch }]">
    <div class="title-area">
      <div class="title">{{ t('配置项') }}</div>
      <div class="title-extend">
        <bk-checkbox
          v-if="isBaseVersionExist && diffConfigList.length > 0"
          v-model="isOnlyShowDiff"
          class="view-diff-checkbox"
          @change="handleToggleShowDiff">
          {{ t('只查看差异项') }}({{ diffCount }})
        </bk-checkbox>
        <div :class="['search-trigger', { actived: isOpenSearch }]" @click="isOpenSearch = !isOpenSearch">
          <Search />
        </div>
      </div>
    </div>
    <div v-if="isOpenSearch" class="search-wrapper">
      <SearchInput v-model="searchStr" :placeholder="t('搜索配置项名称')" @search="handleSearch" />
    </div>
    <div class="groups-wrapper">
      <div v-for="group in groupedConfigListOnShow" class="config-group-item" :key="group.name">
        <div :class="['group-header', { expand: group.expand }]" @click="group.expand = !group.expand">
          <RightShape class="arrow-icon" />
          <span class="name">{{ group.name === 'singleLine' ? t('单行配置') : t('多行配置') }}</span>
        </div>
        <div v-if="group.expand" class="config-list">
          <div
            v-for="config in group.configs"
            v-overflow-title
            :key="config.id"
            :class="['config-item', { actived: props.actived && config.id === selected }]"
            @click="handleSelectItem(config.id)">
            <i v-if="config.diffType" :class="['status-icon', config.diffType]"></i>
            {{ config.key }}
          </div>
        </div>
      </div>
      <tableEmpty
        v-if="groupedConfigListOnShow.length === 0"
        class="empty-tips"
        :is-search-empty="isSearchEmpty"
        :empty-title="t('没有差异配置项')"
        @clear="clearStr">
      </tableEmpty>
    </div>
  </div>
</template>
<script lang="ts" setup>
  import { ref, computed, watch, onMounted } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { useRoute } from 'vue-router';
  import { Search, RightShape } from 'bkui-vue/lib/icon';
  import { IConfigKvType } from '../../../../../../../../../types/config';
  import { ISingleLineKVDIffItem } from '../../../../../../../../../types/service';
  import { getReleaseKvList } from '../../../../../../../../api/config';
  import SearchInput from '../../../../../../../../components/search-input.vue';
  import tableEmpty from '../../../../../../../../components/table/table-empty.vue';

  interface IConfigDiffItem {
    diffType: string;
    kvType: string;
    id: number;
    key: string;
    baseContent: string;
    currentContent: string;
  }

  interface IDiffGroupData {
    name: string;
    expand: boolean;
    configs: IConfigDiffItem[];
  }

  const props = withDefaults(
    defineProps<{
      currentVersionId: number;
      baseVersionId: number | undefined;
      selectedId: number;
      actived: boolean;
    }>(),
    {
      selectedId: 0,
    },
  );

  const { t } = useI18n();
  const route = useRoute();

  const emits = defineEmits(['selected']);

  const SINGLE_LINE_TYPE = ['string', 'number'];

  const bkBizId = ref(String(route.params.spaceId));
  const appId = ref(Number(route.params.appId));
  const diffCount = ref(0);
  const selected = ref();
  const currentList = ref<IConfigKvType[]>([]);
  const baseList = ref<IConfigKvType[]>([]);
  // 汇总的配置文件列表，包含未修改、增加、删除、修改的所有配置文件
  const aggregatedList = ref<IConfigDiffItem[]>([]);
  const groupedConfigListOnShow = ref<IDiffGroupData[]>([]);
  const isOnlyShowDiff = ref(true); // 只显示差异项
  const isOpenSearch = ref(false);
  const searchStr = ref('');
  const isSearchEmpty = ref(false);

  // 是否实际选择了对比的基准版本，为了区分的未命名版本id为0的情况
  const isBaseVersionExist = computed(() => typeof props.baseVersionId === 'number');
  // 差异项配置文件列表
  const diffConfigList = computed(() => {
    return aggregatedList.value.filter((item) => item.diffType !== '');
  });

  // 基准版本变化，更新选中对比项
  watch(
    () => props.baseVersionId,
    async () => {
      baseList.value = await getConfigsOfVersion(props.baseVersionId);
      aggregatedList.value = calcDiff();
      isOnlyShowDiff.value = diffConfigList.value.length > 0;
      groupedConfigListOnShow.value = getGroupedList();
      setDefaultSelected();
    },
  );

  // 当前版本默认选中的配置文件
  watch(
    () => props.selectedId,
    (val) => {
      if (val) {
        selected.value = val;
      }
    },
    {
      immediate: true,
    },
  );

  onMounted(async () => {
    currentList.value = await getConfigsOfVersion(props.currentVersionId);
    baseList.value = await getConfigsOfVersion(props.baseVersionId);
    aggregatedList.value = calcDiff();
    isOnlyShowDiff.value = diffConfigList.value.length > 0;
    groupedConfigListOnShow.value = getGroupedList();
    setDefaultSelected();
  });

  // 获取某一版本下配置文件
  const getConfigsOfVersion = async (releaseId: number | undefined) => {
    if (typeof releaseId !== 'number') {
      return [];
    }
    const res = await getReleaseKvList(bkBizId.value, appId.value, releaseId, { start: 0, all: true });
    return res.details;
  };

  // 计算配置被修改、被删除、新增的差异
  const calcDiff = () => {
    const list: IConfigDiffItem[] = [];
    diffCount.value = 0;
    currentList.value.forEach((currentItem) => {
      let diffType = '';
      let baseContent = '';
      const baseItem = baseList.value.find((item) => item.spec.key === currentItem.spec.key);
      if (baseItem) {
        baseContent = baseItem.spec.value;
        // 当前版本修改项
        if (baseItem.spec.value !== currentItem.spec.value || baseItem.spec.kv_type !== currentItem.spec.kv_type) {
          diffCount.value += 1;
          diffType = isBaseVersionExist.value ? 'modify' : '';
        }
      } else {
        // 当前版本新增项
        diffCount.value += 1;
        diffType = isBaseVersionExist.value ? 'add' : '';
      }
      list.push({
        diffType,
        kvType: currentItem.spec.kv_type,
        key: currentItem.spec.key,
        id: currentItem.id,
        baseContent,
        currentContent: currentItem.spec.value,
      });
    });
    // 计算当前版本删除项
    baseList.value.forEach((baseItem) => {
      const currentItem = currentList.value.find((item) => item.spec.key === baseItem.spec.key);
      if (!currentItem) {
        diffCount.value += 1;
        list.push({
          diffType: isBaseVersionExist.value ? 'delete' : '',
          kvType: baseItem.spec.kv_type,
          key: baseItem.spec.key,
          id: baseItem.id,
          baseContent: baseItem.spec.value,
          currentContent: '',
        });
      }
    });
    list.sort((a, b) => a.key.charCodeAt(0) - b.key.charCodeAt(0));
    return list;
  };

  const getGroupedList = () => {
    const groupedList: IDiffGroupData[] = [];
    let resultList = [];
    if (isOnlyShowDiff.value) {
      resultList = aggregatedList.value.filter(
        (item) => item.diffType !== '' && item.key.toLocaleLowerCase().includes(searchStr.value.toLocaleLowerCase()),
      );
    } else {
      resultList = aggregatedList.value.filter((item) =>
        item.key.toLocaleLowerCase().includes(searchStr.value.toLocaleLowerCase()),
      );
    }
    resultList.forEach((item) => {
      if (SINGLE_LINE_TYPE.includes(item.kvType)) {
        if (groupedList[0]?.name === 'singleLine') {
          groupedList[0].configs.push(item);
        } else {
          groupedList.unshift({
            name: 'singleLine',
            expand: true,
            configs: [item],
          });
        }
      } else {
        const multiLineGroup = groupedList.find((item) => item.name === 'multiLine');
        if (multiLineGroup) {
          multiLineGroup.configs.push(item);
        } else {
          groupedList.push({
            name: 'multiLine',
            expand: true,
            configs: [item],
          });
        }
      }
    });
    return groupedList;
  };

  // 设置默认选中的配置文件
  // 如果props有设置选中项，取props值
  // 如果选中项有值，保持上一次选中项
  // 否则取第一个非空分组的第一个配置文件
  const setDefaultSelected = () => {
    if (props.selectedId) {
      handleSelectItem(props.selectedId);
    } else if (aggregatedList.value.find((item) => item.id === selected.value)) {
      handleSelectItem(selected.value);
    } else {
      if (groupedConfigListOnShow.value[0]) {
        handleSelectItem(groupedConfigListOnShow.value[0].configs[0].id);
      }
    }
  };

  const handleToggleShowDiff = () => {
    groupedConfigListOnShow.value = getGroupedList();
  };

  const handleSearch = () => {
    groupedConfigListOnShow.value = getGroupedList();
    isSearchEmpty.value = searchStr.value !== '' && groupedConfigListOnShow.value.length === 0;
  };

  // 选择对比配置文件后，加载配置文件详情，组装对比数据
  const handleSelectItem = async (selectedId: number) => {
    const config = aggregatedList.value.find((item) => item.id === selectedId);
    if (config) {
      selected.value = selectedId;
      const data = getConfigDiffDetail(config);
      emits('selected', data);
    }
  };

  // 差异对比详情数据
  const getConfigDiffDetail = (config: IConfigDiffItem) => {
    // 单行配置
    if (SINGLE_LINE_TYPE.includes(config.kvType)) {
      const configs: ISingleLineKVDIffItem[] = groupedConfigListOnShow.value[0].configs.map((item) => {
        const { diffType, id, key, baseContent, currentContent } = item;
        return {
          id,
          name: key,
          diffType,
          base: {
            content: baseContent,
          },
          current: {
            content: currentContent,
          },
        };
      });
      return {
        contentType: 'singleLineKV',
        id: config.id,
        singleLineKVDiff: configs,
      };
    }
    // 多行配置
    return {
      contentType: 'text',
      id: config.id,
      base: {
        content: config.baseContent,
        variables: [],
      },
      current: {
        content: config.currentContent,
        variables: [],
      },
    };
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
  }
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

  .empty-tips {
    margin-top: 40px;
    font-size: 12px;
    color: #63656e;
  }
</style>
