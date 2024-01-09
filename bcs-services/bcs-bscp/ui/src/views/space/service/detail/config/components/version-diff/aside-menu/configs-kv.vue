<template>
  <div :class="['configs-menu', { 'search-opened': isOpenSearch }]">
    <div class="title-area">
      <div class="title">{{ t('配置项') }}</div>
      <div class="title-extend">
        <bk-checkbox
          v-if="isBaseVersionExist"
          v-model="isOnlyShowDiff"
          class="view-diff-checkbox"
          @change="handleSearch"
        >
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
      <div
        v-for="(config, index) in groupedConfigListOnShow"
        v-overflow-title
        :key="index"
        :class="['config-item', { actived: getItemSelectedStatus(config) }]"
        @click="handleSelectItem(config.id)"
      >
        <i v-if="config.diff_type" :class="['status-icon', config.diff_type]"></i>
        {{ config.key }}
      </div>
      <tableEmpty
        v-if="groupedConfigListOnShow.length === 0"
        class="empty-tips"
        :is-search-empty="isSearchEmpty"
        :empty-title="t('没有差异配置项')"
        @clear="clearStr"
      >
      </tableEmpty>
    </div>
  </div>
</template>
<script lang="ts" setup>
import { ref, computed, watch, onMounted } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRoute } from 'vue-router';
import { Search } from 'bkui-vue/lib/icon';
import { ICommonQuery } from '../../../../../../../../../types/index';
import { IConfigKvType } from '../../../../../../../../../types/config';
import { IVariableEditParams } from '../../../../../../../../../types/variable';
import { getReleaseKvList } from '../../../../../../../../api/config';
import SearchInput from '../../../../../../../../components/search-input.vue';
import tableEmpty from '../../../../../../../../components/table/table-empty.vue';

interface IConfigDiffItem {
  diff_type: string;
  key: string;
  id: number;
  baseContent: string;
  currentContent: string;
  baseType: string;
  currentType: string;
}

const props = withDefaults(
  defineProps<{
    currentVersionId: number;
    unNamedVersionVariables?: IVariableEditParams[]; // 未命名版本变量列表
    baseVersionId: number | undefined;
    selectedConfig: number;
    actived: boolean;
    isPublish: boolean;
  }>(),
  {
    unNamedVersionVariables: () => [],
    selectedConfig: 0,
  },
);

const { t } = useI18n();
const emits = defineEmits(['selected']);

const route = useRoute();
const bkBizId = ref(String(route.params.spaceId));
const appId = ref(Number(route.params.appId));

const diffCount = ref(0);
const selected = ref();
const currentList = ref<IConfigKvType[]>([]);
const currentVariables = ref<IVariableEditParams[]>([]);
const baseList = ref<IConfigKvType[]>([]);
const baseVariables = ref<IVariableEditParams[]>([]);
// 汇总的配置文件列表，包含未修改、增加、删除、修改的所有配置文件
const aggregatedList = ref<IConfigDiffItem[]>([]);
const groupedConfigListOnShow = ref<IConfigDiffItem[]>([]);
const isOnlyShowDiff = ref(false); // 只显示差异项
const isOpenSearch = ref(false);
const searchStr = ref('');
const isSearchEmpty = ref(false);

// 是否实际选择了对比的基准版本，为了区分的未命名版本id为0的情况
const isBaseVersionExist = computed(() => typeof props.baseVersionId === 'number');

// 基准版本变化，更新选中对比项
watch(
  () => props.baseVersionId,
  async () => {
    const base = await getConfigsOfVersion(props.baseVersionId);
    baseList.value = base.details;
    aggregatedList.value = calcDiff();
    aggregatedList.value.sort((a, b) => a.key.charCodeAt(0) - b.key.charCodeAt(0));
    groupedConfigListOnShow.value = aggregatedList.value.slice();
    setDefaultSelected();
    isOnlyShowDiff.value && handleSearch();
  },
);

// 当前版本默认选中的配置文件
watch(
  () => props.selectedConfig,
  (val) => {
    if (val) {
      selected.value = val;
    }
  },
  {
    immediate: true,
  },
);

watch(
  () => isOnlyShowDiff.value,
  () => {
    let hasSelectConfig = false;
    groupedConfigListOnShow.value.forEach((group) => {
      if (group.id === selected.value) {
        hasSelectConfig = true;
        handleSelectItem(group.id);
      }
    });
    if (!hasSelectConfig) {
      emits('selected', {
        contentType: 'text',
        base: { content: '', variables: '' },
        current: { content: '', variables: '' },
      });
    }
  },
);

watch(
  () => searchStr.value,
  (val) => {
    isSearchEmpty.value = !!val;
  },
);

onMounted(async () => {
  await getAllConfigList();
  aggregatedList.value = calcDiff();
  aggregatedList.value.sort((a, b) => a.key.charCodeAt(0) - b.key.charCodeAt(0));
  groupedConfigListOnShow.value = aggregatedList.value.slice();
  setDefaultSelected();
  // 如果是上线版本 默认选中只差看差异项
  if (props.isPublish) {
    isOnlyShowDiff.value = true;
    handleSearch();
  }
});

// 获取当前版本和基准版本的所有配置文件列表(非模板配置和套餐下模板)
const getAllConfigList = async () => {
  const [current, base] = await Promise.all([
    getConfigsOfVersion(props.currentVersionId),
    getConfigsOfVersion(props.baseVersionId),
  ]);
  currentList.value = current.details;
  baseList.value = base.details || base;
};

// 获取某一版本下配置文件
const getConfigsOfVersion = async (releaseId: number | undefined) => {
  if (typeof releaseId !== 'number') {
    return [];
  }
  const params: ICommonQuery = {
    start: 0,
    all: true,
  };
  return await getReleaseKvList(bkBizId.value, appId.value, releaseId, params);
};

// 计算配置被修改、被删除、新增的差异
const calcDiff = () => {
  const list: IConfigDiffItem[] = [];
  diffCount.value = 0;
  currentList.value.forEach((currentItem) => {
    let baseItem: IConfigKvType | undefined;
    baseList.value.forEach((item) => {
      if (item.spec.key === currentItem.spec.key) {
        baseItem = item;
      }
    });
    if (baseItem) {
      // 当前版本修改项
      if (baseItem.spec.value !== currentItem.spec.value || baseItem.spec.kv_type !== currentItem.spec.kv_type) {
        diffCount.value += 1;
        list.push({
          diff_type: isBaseVersionExist.value ? 'modify' : '',
          key: baseItem.spec.key,
          id: currentItem.id,
          baseContent: baseItem.spec.value,
          currentContent: currentItem.spec.value,
          baseType: baseItem.spec.kv_type,
          currentType: currentItem.spec.kv_type,
        });
      } else {
        list.push({
          diff_type: '',
          key: baseItem.spec.key,
          id: currentItem.id,
          baseContent: baseItem.spec.value,
          currentContent: currentItem.spec.value,
          baseType: baseItem.spec.kv_type,
          currentType: currentItem.spec.kv_type,
        });
      }
    } else {
      // 当前版本新增项
      diffCount.value += 1;
      list.push({
        diff_type: isBaseVersionExist.value ? 'add' : '',
        key: currentItem.spec.key,
        id: currentItem.id,
        baseContent: '',
        currentContent: currentItem.spec.value,
        baseType: '',
        currentType: currentItem.spec.kv_type,
      });
    }
  });
  // 计算当前版本删除项
  baseList.value.forEach((baseItem) => {
    let currentItem: IConfigKvType | undefined;
    currentList.value.some((item) => {
      if (baseItem.spec.key === item.spec.key) {
        currentItem = item;
        return true;
      }
      return false;
    });
    if (!currentItem) {
      diffCount.value += 1;
      list.push({
        diff_type: isBaseVersionExist.value ? 'delete' : '',
        key: baseItem.spec.key,
        id: baseItem.id,
        baseContent: baseItem.spec.value,
        currentContent: '',
        baseType: baseItem.spec.kv_type,
        currentType: '',
      });
    }
  });
  return list;
};

// 设置默认选中的配置文件
// 如果props有设置选中项，取props值
// 如果选中项有值，保持上一次选中项
// 否则取第一个非空分组的第一个配置文件
const setDefaultSelected = () => {
  if (props.selectedConfig) {
    handleSelectItem(props.selectedConfig);
  } else if (aggregatedList.value.find(item => item.id === selected.value)) {
    handleSelectItem(selected.value);
  } else {
    if (aggregatedList.value[0]) {
      handleSelectItem(aggregatedList.value[0].id);
    }
  }
};

const handleSearch = () => {
  if (!searchStr.value && !isOnlyShowDiff.value) {
    groupedConfigListOnShow.value = aggregatedList.value.slice();
  } else {
    // 点击只查看配置文件 默认展示第一个
    groupedConfigListOnShow.value = aggregatedList.value.filter((config) => {
      const isSearchHit = config.key.toLocaleLowerCase().includes(searchStr.value.toLocaleLowerCase());
      if (isOnlyShowDiff.value) {
        return config.diff_type !== '' && isSearchHit;
      }
      return isSearchHit;
    });
    if (groupedConfigListOnShow.value.length === 0) return;
    handleSelectItem(groupedConfigListOnShow.value[0].id);
  }
};

const getItemSelectedStatus = (config: IConfigDiffItem) => props.actived && config.id === selected.value;

// 选择对比配置文件后，加载配置文件详情，组装对比数据
const handleSelectItem = async (selectedConfig: number) => {
  const config = aggregatedList.value.find(item => item.id === selectedConfig);
  if (config) {
    selected.value = selectedConfig;
    const data = await getConfigDiffDetail(config);
    emits('selected', data);
  }
};

const getConfigDiffDetail = async (config: IConfigDiffItem) => {
  const currentConfigContent = config.currentContent;
  const baseConfigContent = config.baseContent;
  let configType: string;
  const kvType = ['string', 'number'];
  if (config.currentType) {
    configType = kvType.includes(config.currentType) ? 'kv' : 'text';
  } else {
    configType = kvType.includes(config.baseType) ? 'kv' : 'text';
  }
  return {
    contentType: configType,
    base: {
      content: baseConfigContent,
      variables: baseVariables.value,
    },
    current: {
      content: currentConfigContent,
      variables: currentVariables.value,
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
