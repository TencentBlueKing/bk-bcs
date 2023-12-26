<template>
  <section class="version-detail-table">
    <div class="service-selector-wrapper">
      <ServiceSelector :value="props.appId" />
    </div>
    <div class="content-container">
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
        <SearchInput
          v-model="searchStr"
          class="version-search-input"
          placeholder="版本名称/版本说明/修改人"
          :width="320"
          @search="handleSearch"/>
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
              <template v-if="row.status">
                <template v-if="row.status.publish_status !== 'partial_released'">{{ getGroupNames(row) }}</template>
                <ReleasedGroupViewer
                  v-else
                  placement="bottom-start"
                  :bk-biz-id="props.bkBizId"
                  :app-id="props.appId"
                  :groups="row.status.released_groups"
                  :is-default-group="isVersionInDefaultGroup(row.status.released_groups)">
                  <div>{{ getGroupNames(row) }}</div>
                </ReleasedGroupViewer>
              </template>
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
              <div v-if="row.spec && row.spec.deprecated" class="status-tag deprecated">已废弃</div>
              <template v-else-if="row.status">
                <template v-if="!VERSION_STATUS_MAP[row.status.publish_status as keyof typeof VERSION_STATUS_MAP]">
                  --
                </template>
                <div v-else :class="['status-tag', row.status.publish_status]">
                  {{ row.status.publish_status === 'not_released' ? '未上线' : '已上线' }}
                </div>
              </template>
            </template>
          </bk-table-column>
          <bk-table-column label="操作">
            <template #default="{ row }">
              <template v-if="row.status">
                <template v-if="currentTab === 'avaliable'">
                  <template v-if="row.status.publish_status === 'editing'">--</template>
                  <div v-else class="actions-wrapper">
                    <bk-button
                      text
                      theme="primary"
                      @click.stop="handleOpenDiff(row)">
                      版本对比
                    </bk-button>
                    <bk-button
                      v-bk-tooltips="{
                        disabled: row.status.publish_status === 'not_released',
                        placement: 'bottom',
                        content: '只支持未上线版本'
                      }"
                      text
                      theme="primary"
                      :disabled="row.status.publish_status !== 'not_released'"
                      @click.stop="handleDeprecate(row)">
                      版本废弃
                    </bk-button>
                  </div>
                </template>
                <div v-else class="actions-wrapper">
                  <bk-button text theme="primary" @click.stop="handleUndeprecate(row)">恢复</bk-button>
                  <bk-button text theme="primary" @click.stop="handleDelete(row)">删除</bk-button>
                </div>
              </template>
            </template>
          </bk-table-column>
          <template #empty>
            <tableEmpty :is-search-empty="isSearchEmpty" @clear="handleClearSearchStr"></tableEmpty>
          </template>
        </bk-table>
      </bk-loading>
    </div>
    <VersionDiff v-model:show="showDiffPanel" :current-version="diffVersion" />
    <VersionOperateConfirmDialog
      v-model:show="operateConfirmDialog.open"
      :title="operateConfirmDialog.title"
      :tips="operateConfirmDialog.tips"
      :confirm-fn="operateConfirmDialog.confirmFn"
      :version="operateConfirmDialog.version" />
  </section>
</template>
<script setup lang="ts">
import { ref, computed, watch, onMounted, version } from 'vue';
import { useRouter } from 'vue-router'
import { storeToRefs } from 'pinia';
import useConfigStore from '../../../../../../store/config';
import { getConfigVersionList, deprecateVersion, undeprecateVersion, deleteVersion } from '../../../../../../api/config';
import { datetimeFormat } from '../../../../../../utils/index';
import { VERSION_STATUS_MAP, GET_UNNAMED_VERSION_DATA } from '../../../../../../constants/config';
import { IConfigVersion, IConfigVersionQueryParams, IReleasedGroup } from '../../../../../../../types/config';
import ServiceSelector from '../../components/service-selector.vue';
import SearchInput from '../../../../../../components/search-input.vue';
import VersionDiff from '../../config/components/version-diff/index.vue';
import tableEmpty from '../../../../../../components/table/table-empty.vue';
import ReleasedGroupViewer from '../components/released-group-viewer.vue';
import VersionOperateConfirmDialog from './version-operate-confirm-dialog.vue';

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
const operateConfirmDialog = ref({
  open: false,
  version: UN_NAMED_VERSION,
  title: '',
  tips: '',
  confirmFn: () => {},
});

// 可用版本非搜索查看视图
const isAvaliableView = computed(() => currentTab.value === 'avaliable' && searchStr.value === '');

watch(() => props.appId, () => {
  getVersionList();
});

onMounted(() => {
  getVersionList();
});

const getVersionList = async () => {
  listLoading.value = true;
  const notFirstPageStart = isAvaliableView.value
    ? (pagination.value.current - 1) * pagination.value.limit - 1
    : (pagination.value.current - 1) * pagination.value.limit;
  const params: IConfigVersionQueryParams = {
    start: pagination.value.current === 1 ? 0 : notFirstPageStart,
    limit: pagination.value.current === 1 && isAvaliableView.value ? pagination.value.limit - 1 : pagination.value.limit,
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

const getGroupNames = (data: IConfigVersion) => {
  const status = data.status?.publish_status
  if (status === 'partial_released') {
    return data.status.released_groups.map(item => item.name).join('; ');
  } else if (status === 'full_released') {
    return '全部实例';
  }
  return '--'
}

const isVersionInDefaultGroup = (groups: IReleasedGroup[]) => {
  return groups.findIndex(item => item.id === 0) > -1;
};

const handleTabChange = (tab: string) => {
  currentTab.value = tab;
  pagination.value.current = 1;
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

// 打开版本对比
const handleOpenDiff = (version: IConfigVersion) => {
  showDiffPanel.value = true;
  diffVersion.value = version;
};

// 废弃
const handleDeprecate = (version: IConfigVersion) => {
  operateConfirmDialog.value.open = true;
  operateConfirmDialog.value.title = '确认废弃该版本';
  operateConfirmDialog.value.tips = '此操作不会删除版本，如需找回或彻底删除请去版本详情的废弃版本列表操作';
  operateConfirmDialog.value.version = version;
  operateConfirmDialog.value.confirmFn = () => {
    return new Promise(() => {
      deprecateVersion(props.bkBizId, props.appId, version.id)
        .then(() => {
          operateConfirmDialog.value.open = false;
          updateListAndSetVersionAfterOperate(version.id);
        });
    })
  };
};

// 恢复
const handleUndeprecate = (version: IConfigVersion) => {
  operateConfirmDialog.value.open = true;
  operateConfirmDialog.value.title = '确认恢复该版本';
  operateConfirmDialog.value.tips = '此操作会把改版本恢复至可用版本列表';
  operateConfirmDialog.value.version = version;
  operateConfirmDialog.value.confirmFn = () => {
    return new Promise(() => {
      undeprecateVersion(props.bkBizId, props.appId, version.id)
        .then(() => {
          operateConfirmDialog.value.open = false;
          updateListAndSetVersionAfterOperate(version.id);
        });
    })
  };
};

// 删除
const handleDelete = (version: IConfigVersion) => {
  operateConfirmDialog.value.open = true;
  operateConfirmDialog.value.title = '确认删除该版本';
  operateConfirmDialog.value.tips = '一旦删除，该操作将无法撤销，请谨慎操作';
  operateConfirmDialog.value.version = version;
  operateConfirmDialog.value.confirmFn = () => {
    return new Promise(() => {
      deleteVersion(props.bkBizId, props.appId, version.id)
        .then(() => {
          operateConfirmDialog.value.open = false;
          updateListAndSetVersionAfterOperate(version.id);
        });
    })
  };
};

// 更新列表数据以及设置选中版本
const updateListAndSetVersionAfterOperate = async(id: number) => {
  const index = versionList.value.findIndex(item => item.id === id);
  const currentPage = pagination.value.current;
  pagination.value.current = (versionList.value.length === 1 && currentPage > 1) ? currentPage - 1 : currentPage;
  await getVersionList();
  if (id === versionData.value.id) {
    const len = versionList.value.length;
    if (len > 0) {
      const version = len - 1 >= index ? versionList.value[index] : versionList.value[len - 1];
      handleSelectVersion(undefined, version);
    } else {
      handleSelectVersion(undefined, UN_NAMED_VERSION)
    }
  }
};

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
  height: 100%;
  background: #ffffff;
}
.service-selector-wrapper {
  padding: 10px 24px;
  border-bottom: 1px solid #eaebf0;
  :deep(.service-selector) {
    width: 264px;
  }
}
.content-container {
  padding: 12px 24px;
  height: calc(100% - 53px);
  overflow: auto;
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
  &.deprecated {
    color: #ea3536;
    background-color: #feebea;
    border-color: #ea35364d;
  }
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
.actions-wrapper {
  .bk-button:not(:first-child) {
    margin-left: 8px;
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
