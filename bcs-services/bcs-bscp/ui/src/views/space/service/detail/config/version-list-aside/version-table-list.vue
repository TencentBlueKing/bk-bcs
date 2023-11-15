<template>
  <section class="version-detail-table">
    <div class="head-operate-wrapper">
      <div class="type-tabs">
        <div :class="['tab-item', { active: currentTab === 'avaliable' }]" @click="handleTabChange('avaliable')">
          可用版本
        </div>
        <div class="split-line"></div>
        <div :class="['tab-item', { active: currentTab === 'deprecate' }]" @click="handleTabChange('deprecate')">
          废弃版本
        </div>
      </div>
      <bk-input
        v-model="searchStr"
        class="version-search-input"
        placeholder="版本名称/版本说明/修改人"
        :clearable="true"
        @change="handleSearchInputChange"
        @enter="handleSearch"
        @clear="handleClearSearchStr"
      >
        <template #suffix>
          <Search class="search-input-icon" />
        </template>
      </bk-input>
    </div>
    <bk-loading :loading="listLoading">
      <bk-table
        :border="['outer']"
        :data="versionList"
        :row-class="getRowCls"
        :remote-pagination="true"
        :pagination="pagination"
        @row-click="handleSelectVersion"
        @page-limit-change="handlePageLimitChange"
        @page-value-change="refreshVersionList($event)"
      >
        <bk-table-column label="版本" prop="spec.name" show-overflow-tooltip></bk-table-column>
        <bk-table-column label="版本描述" prop="spec.memo" show-overflow-tooltip>
          <template #default="{ row }">
            {{ row.spec?.memo || '--' }}
          </template>
        </bk-table-column>
        <bk-table-column label="已上线分组" show-overflow-tooltip>
          <template #default="{ row }">
            {{ getGroupNames(row) }}
          </template>
        </bk-table-column>
        <bk-table-column label="创建人">
          <template #default="{ row }">
            {{ row.revision?.creator || '--' }}
          </template>
        </bk-table-column>
        <bk-table-column label="生成时间" width="220">
          <template #default="{ row }">
            <span v-if="row.revision">{{
              row.revision.create_at ? datetimeFormat(row.revision.create_at) : '--'
            }}</span>
          </template>
        </bk-table-column>
        <bk-table-column label="状态">
          <template #default="{ row }">
            <template v-if="row.status">
              <template v-if="!VERSION_STATUS_MAP[row.status.publish_status as keyof typeof VERSION_STATUS_MAP]">
                --
              </template>
              <div v-else :class="['status-tag', row.status.publish_status]">
                {{ VERSION_STATUS_MAP[row.status.publish_status as keyof typeof VERSION_STATUS_MAP] }}
              </div>
            </template>
          </template>
        </bk-table-column>
        <bk-table-column label="操作">
          <template #default="{ row }">
            <bk-button
              v-if="row.status?.publish_status !== 'editing'"
              style="margin-right: 16px"
              text
              theme="primary"
              @click.stop="handleOpenDiff(row)"
            >
              版本对比
            </bk-button>
            <template v-else>--</template>
            <!-- <bk-button text theme="primary" @click.stop="handleDeprecate(row.id)">废弃</bk-button> -->
          </template>
        </bk-table-column>
        <template #empty>
          <tableEmpty :is-search-empty="isSearchEmpty" @clear="handleClearSearchStr"></tableEmpty>
        </template>
      </bk-table>
    </bk-loading>
    <VersionDiff v-model:show="showDiffPanel" :current-version="diffVersion" />
  </section>
</template>
<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { useRouter } from 'vue-router'
import { Search } from 'bkui-vue/lib/icon';
// import { InfoBox } from 'bkui-vue/lib';
import { storeToRefs } from 'pinia';
import useConfigStore from '../../../../../../store/config';
import { getConfigVersionList } from '../../../../../../api/config';
import { datetimeFormat } from '../../../../../../utils/index';
import { VERSION_STATUS_MAP, GET_UNNAMED_VERSION_DATA } from '../../../../../../constants/config';
import { IConfigVersion, IConfigVersionQueryParams } from '../../../../../../../types/config';
import VersionDiff from '../../config/components/version-diff/index.vue';
import tableEmpty from '../../../../../../components/table/table-empty.vue';

const configStore = useConfigStore();
const { versionData } = storeToRefs(configStore);

const router = useRouter()

const props = defineProps<{
  bkBizId: string;
  appId: number;
}>();

const UN_NAMED_VERSION = GET_UNNAMED_VERSION_DATA();

const listLoading = ref(true);
const versionList = ref<Array<IConfigVersion>>([]);
const currentTab = ref('avaliable');
const searchStr = ref('');
const showDiffPanel = ref(false);
const diffVersion = ref();
const pagination = ref({
  current: 1,
  count: 0,
  limit: 10,
});
const isSearchEmpty = ref(false);

// 可用版本非搜索查看视图
const isAvaliableView = computed(() => currentTab.value === 'avaliable' && searchStr.value === '');

onMounted(async () => {
  getVersionList();
});

const getVersionList = async () => {
  listLoading.value = true;
  const notFirstPageStart = isAvaliableView.value
    ? (pagination.value.current - 1) * pagination.value.limit - 1
    : (pagination.value.current - 1) * pagination.value.limit;
  const params: IConfigVersionQueryParams = {
    start: pagination.value.current === 1 ? 0 : notFirstPageStart,
    limit:
      pagination.value.current === 1 && isAvaliableView.value ? pagination.value.limit - 1 : pagination.value.limit,
    deprecated: currentTab.value !== 'avaliable',
  };
  if (searchStr.value) {
    params.searchKey = searchStr.value;
  }
  const res = await getConfigVersionList(props.bkBizId, props.appId, params);
  const count = isAvaliableView.value ? res.data.count + 1 : res.data.count;
  if (isAvaliableView.value && pagination.value.current === 1) {
    versionList.value = [UN_NAMED_VERSION, ...res.data.details];
  } else {
    versionList.value = res.data.details;
  }
  pagination.value.count = count;
  listLoading.value = false;
};

const getRowCls = (data: IConfigVersion) => {
  if (data.id === versionData.value.id) {
    return 'selected';
  }
  return '';
};

const getGroupNames = (data: IConfigVersion) => (data.status?.released_groups.length ? data.status.released_groups.map(item => item.name).join('; ') : '--');

const handleTabChange = (tab: string) => {
  currentTab.value = tab;
  refreshVersionList();
};

// 选择某个版本
const handleSelectVersion = (event: Event | undefined, data: IConfigVersion) => {
  configStore.$patch((state) => {
    state.versionData = data;
  });
  const params: { spaceId: string, appId: number, versionId?: number } = {
    spaceId: props.bkBizId,
    appId: props.appId
  }
  if (data.id !== 0) {
    params.versionId = data.id;
  }
  router.push({ name: 'service-config', params });
};

// 搜索框输入事件处理，内容为空时触发一次搜索
const handleSearchInputChange = (val: string) => {
  if (!val) {
    refreshVersionList();
  }
};

// 打开版本对比
const handleOpenDiff = (version: IConfigVersion) => {
  showDiffPanel.value = true;
  diffVersion.value = version;
};

// 废弃
// const handleDeprecate = (id: number) => {
//   InfoBox({
//     title: '确认废弃此版本？',
//     subTitle: '废弃操作无法撤回，请谨慎操作！',
//     headerAlign: 'center' as const,
//     footerAlign: 'center' as const,
//     onConfirm: () => {
//     },
//   } as any);
// };

const handlePageLimitChange = (limit: number) => {
  pagination.value.limit = limit;
  refreshVersionList();
};

const refreshVersionList = (current = 1) => {
  pagination.value.current = current;
  getVersionList();
};

const handleSearch = () => {
  isSearchEmpty.value = true;
  refreshVersionList();
};

const handleClearSearchStr = () => {
  searchStr.value = '';
  isSearchEmpty.value = false;
  refreshVersionList();
};
</script>
<style lang="scss" scoped>
.version-detail-table {
  padding: 24px;
  height: 100%;
  background: #ffffff;
}
.head-operate-wrapper {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
}
.type-tabs {
  display: flex;
  align-items: center;
  padding: 3px 4px;
  background: #f0f1f5;
  border-radius: 4px;
  .tab-item {
    padding: 6px 14px;
    font-size: 12px;
    line-height: 14px;
    color: #63656e;
    border-radius: 4px;
    cursor: pointer;
    &.active {
      color: #3a84ff;
      background: #ffffff;
    }
  }
  .split-line {
    margin: 0 4px;
    width: 1px;
    height: 14px;
    background: #dcdee5;
  }
}
.version-search-input {
  width: 320px;
}
.search-input-icon {
  padding-right: 10px;
  color: #979ba5;
}
.bk-table {
  :deep(.bk-table-body) {
    tr {
      cursor: pointer;
      &.selected td {
        background: #e1ecff !important;
      }
    }
  }
}
.status-tag {
  display: inline-block;
  padding: 0 10px;
  line-height: 20px;
  font-size: 12px;
  border: 1px solid #cccccc;
  border-radius: 11px;
  text-align: center;
  &.not_released {
    color: #fe9000;
    background: #ffe8c3;
    border-color: rgba(254, 156, 0, 0.3);
  }
  &.full_released,
  &.partial_released {
    color: #14a568;
    background: #e4faf0;
    border-color: rgba(20, 165, 104, 0.3);
  }
}
.header-wrapper {
  display: flex;
  align-items: center;
  padding: 0 24px;
  height: 100%;
  font-size: 12px;
  line-height: 1;
}
.header-name {
  display: flex;
  align-items: center;
  font-size: 12px;
  color: #3a84ff;
  cursor: pointer;
}
.arrow-left {
  font-size: 26px;
  color: #3884ff;
}
.arrow-right {
  font-size: 24px;
  color: #c4c6cc;
}
.diff-left-panel-head {
  padding: 0 24px;
  font-size: 12px;
}
</style>
