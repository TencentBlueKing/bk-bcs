<template>
  <section class="client-search-page">
    <div class="header">
      <ClientHeader :title="t('客户端查询')" @search="loadList" />
    </div>
    <div v-if="appId" class="content">
      <BatchRetryBtn
        :bk-biz-id="bkBizId"
        :app-id="appId"
        :selections="selectedClient"
        :is-across-checked="isAcrossChecked"
        @retried="handleRetryConfirm" />
      <bk-loading style="min-height: 100px" :loading="listLoading">
        <bk-table
          ref="tableRef"
          :data="tableData"
          :border="['outer', 'row']"
          class="client-search-table"
          :remote-pagination="true"
          :pagination="pagination"
          :key="appId"
          :settings="settings"
          show-overflow-tooltip
          @page-limit-change="handlePageLimitChange"
          @page-value-change="loadList(true)"
          @column-filter="handleFilter"
          @column-sort="handleSort"
          @setting-change="handleSettingsChange">
          <template #prepend>
            <render-table-tip />
          </template>
          <bk-table-column
            :min-width="80"
            fixed="left"
            :width="80"
            :label="renderSelection"
            :show-overflow-tooltip="false">
            <template #default="{ row }">
              <across-check-box
                :checked="isChecked(row)"
                :disabled="
                  row.client?.spec &&
                  (row.client.spec.release_change_status !== 'Failed' || row.client.spec.online_status !== 'Online')
                "
                :handle-change="() => handleSelectionChange(row)" />
            </template>
          </bk-table-column>
          <bk-table-column label="UID" fixed="left" :width="254" prop="client.attachment.uid"></bk-table-column>
          <bk-table-column
            v-if="selectedShowColumn.includes('ip')"
            label="IP"
            :width="120"
            prop="client.spec.ip"></bk-table-column>
          <bk-table-column v-if="selectedShowColumn.includes('label')" :label="t('客户端标签')" :min-width="296">
            <template #default="{ row }">
              <div v-if="row.client && row.client.labels.length" class="labels">
                <span v-for="(label, index) in row.client.labels" :key="index">
                  <Tag v-if="index < 3">
                    {{ label.key + '=' + label.value }}
                  </Tag>
                </span>
                <div v-if="row.client.labels.length > 3">
                  <bk-popover theme="light" placement="top-center">
                    <Tag> +{{ row.client.labels.length - 3 }} </Tag>
                    <template #content>
                      <Tag v-for="(label, index) in row.client.labels.slice(3)" :key="index">
                        {{ label.key + '=' + label.value }}
                      </Tag>
                    </template>
                  </bk-popover>
                </div>
              </div>
              <span v-else>--</span>
            </template>
          </bk-table-column>
          <bk-table-column
            v-if="selectedShowColumn.includes('current-version')"
            :label="t('目标配置版本')"
            :width="140">
            <template #default="{ row }">
              <div
                v-if="row.client && row.client.spec.target_release_id"
                class="current-version"
                @click="linkToApp(row.client.spec.target_release_id)">
                <Share class="icon" />
                <span class="text">{{ row.client.spec.target_release_name }}</span>
              </div>
              <span v-else>--</span>
            </template>
          </bk-table-column>
          <bk-table-column
            v-if="selectedShowColumn.includes('pull-status')"
            :label="t('最近一次拉取配置状态')"
            :width="178"
            :filter="{
              filterFn: () => true,
              list: releaseChangeStatusFilterList,
              checked: releaseChangeStatusFilterChecked,
            }">
            <template #default="{ row }">
              <div v-if="row.client" class="release_change_status">
                <Spinner v-if="row.client.spec.release_change_status === 'Processing'" class="spinner-icon" />
                <div v-else :class="['dot', row.client.spec.release_change_status]"></div>
                <span>
                  {{ CLIENT_STATUS_MAP[row.client.spec.release_change_status as keyof typeof CLIENT_STATUS_MAP] }}
                </span>
                <InfoLine
                  v-if="row.client.spec.release_change_status === 'Failed'"
                  class="info-icon"
                  fill="#979BA5"
                  v-bk-tooltips="{ content: getErrorDetails(row.client.spec) }" />
              </div>
            </template>
          </bk-table-column>
          <bk-table-column
            v-if="selectedShowColumn.includes('pull-time')"
            :label="t('最后一次拉取配置耗时')"
            :width="200"
            :sort="true">
            <template #default="{ row }">
              <span v-if="row.client">
                {{
                  row.client.spec.total_seconds > 1
                    ? `${Math.round(row.client.spec.total_seconds)}s`
                    : `${Math.round(row.client.spec.total_seconds * 1000)}ms`
                }}
              </span>
            </template>
          </bk-table-column>
          <!-- <bk-table-column label="附加信息" :width="244"></bk-table-column> -->
          <bk-table-column
            v-if="selectedShowColumn.includes('online-status')"
            :label="t('在线状态')"
            :width="94"
            :filter="{
              filterFn: () => true,
              list: onlineStatusFilterList,
              checked: onlineStatusFilterChecked,
            }">
            <template #default="{ row }">
              <div v-if="row.client" class="online-status">
                <div :class="['dot', row.client.spec.online_status]"></div>
                <span>{{ row.client.spec.online_status === 'Online' ? t('在线') : t('离线') }}</span>
              </div>
            </template>
          </bk-table-column>
          <bk-table-column
            v-if="selectedShowColumn.includes('first-connect-time')"
            :label="t('首次连接时间')"
            :width="154">
            <template #default="{ row }">
              <span v-if="row.client">
                {{ datetimeFormat(row.client.spec.first_connect_time) }}
              </span>
            </template>
          </bk-table-column>
          <bk-table-column
            v-if="selectedShowColumn.includes('last-heartbeat-time')"
            :label="t('最后心跳时间')"
            :width="154">
            <template #default="{ row }">
              <span v-if="row.client">
                {{ datetimeFormat(row.client.spec.last_heartbeat_time) }}
              </span>
            </template>
          </bk-table-column>
          <!-- <bk-table-column
            :label="
              () =>
                h('span', [
                  h('span', t('CPU资源占用')),
                  h('span', { style: 'color: #979BA5; margin-left: 4px' }, t('(当前/最大)')),
                ])
            "
            :width="174">
            <template #default="{ row }">
              <span v-if="row.client.spec">
                {{ showResourse(row.client.spec.resource).cpuResourse }}
              </span>
            </template>
          </bk-table-column> -->
          <bk-table-column
            v-if="selectedShowColumn.includes('cpu-resource')"
            :label="t('CPU资源占用(当前/最大)')"
            :width="174">
            <template #default="{ row }">
              <span v-if="row.client">
                {{ `${row.cpu_usage_str} ${t('核')}/${row.cpu_max_usage_str} ${t('核')}` }}
              </span>
            </template>
          </bk-table-column>
          <!-- <bk-table-column
            :label="
              () =>
                h('div', [
                  h('span', t('内存资源占用')),
                  h('span', { style: 'color: #979BA5; margin-left: 4px' }, t('(当前/最大)')),
                ])
            "
            :width="170">
            <template #default="{ row }">
              <span v-if="row.client.spec">
                {{ showResourse(row.client.spec.resource).memoryResource }}
              </span>
            </template>
          </bk-table-column> -->
          <bk-table-column
            v-if="selectedShowColumn.includes('memory-resource')"
            :label="t('内存资源占用(当前/最大)')"
            :width="170">
            <template #default="{ row }">
              <span v-if="row.client">
                {{ `${row.memory_usage_str}MB/${row.memory_max_usage_str}MB` }}
              </span>
            </template>
          </bk-table-column>
          <bk-table-column
            v-if="selectedShowColumn.includes('client-type')"
            :label="t('客户端组件类型')"
            :width="128"
            prop="client.spec.client_type"></bk-table-column>
          <bk-table-column
            v-if="selectedShowColumn.includes('client-version')"
            :label="t('客户端组件版本')"
            :width="128"
            prop="client.spec.client_version"></bk-table-column>
          <bk-table-column :label="t('操作')" :width="locale === 'zh-cn' ? 160 : 230" fixed="right">
            <template #default="{ row }">
              <div v-if="row.client">
                <bk-button theme="primary" text @click="handleShowPullRecord(row.client.attachment.uid, row.client.id)">
                  {{ t('配置拉取记录') }}
                </bk-button>
                <RetryBtn
                  v-if="
                    row.client.spec.release_change_status === 'Failed' && row.client.spec.online_status === 'Online'
                  "
                  :bk-biz-id="bkBizId"
                  :app-id="appId"
                  :client="row.client"
                  @retried="handleRetryConfirm" />
              </div>
            </template>
          </bk-table-column>
          <template #empty>
            <TableEmpty :is-search-empty="isSearchEmpty" @clear="handleClearConditionalList" />
          </template>
        </bk-table>
      </bk-loading>
    </div>
    <Exception v-else />
  </section>
  <PullRecord
    :bk-biz-id="bkBizId"
    :app-id="appId"
    :id="viewPullRecordClientId"
    :uid="viewPullRecordClientUid"
    :show="isShowPullRecordSlider"
    @close="handleCloseSlider" />
</template>

<script lang="ts" setup>
  import { ref, watch, onBeforeMount, onBeforeUnmount, computed } from 'vue';
  import { useRoute, useRouter } from 'vue-router';
  import { Share, InfoLine, Spinner } from 'bkui-vue/lib/icon';
  import { storeToRefs } from 'pinia';
  import { Tag } from 'bkui-vue';
  import { getClientQueryList } from '../../../../api/client';
  import ClientHeader from '../components/client-header.vue';
  import PullRecord from './components/pull-record.vue';
  import { datetimeFormat } from '../../../../utils';
  import {
    CLIENT_STATUS_MAP,
    CLIENT_ERROR_CATEGORY_MAP,
    CLIENT_ERROR_SUBCLASSES_MAP,
  } from '../../../../constants/client';
  import { IClinetCommonQuery } from '../../../../../types/client';
  import useClientStore from '../../../../store/client';
  import useTablePagination from '../../../../utils/hooks/use-table-pagination';
  import TableEmpty from '../../../../components/table/table-empty.vue';
  import Exception from '../components/exception.vue';
  import BatchRetryBtn from './components/batch-retry-btn.vue';
  import RetryBtn from './components/retry-btn.vue';
  import useTableAcrossCheck from '../../../../utils/hooks/use-table-acrosscheck';
  import acrossCheckBox from '../../../../components/across-checkbox.vue';
  import CheckType from '../../../../../types/across-checked';
  import { useI18n } from 'vue-i18n';

  const { t, locale } = useI18n();

  const clientStore = useClientStore();
  const { searchQuery } = storeToRefs(clientStore);

  const route = useRoute();
  const router = useRouter();

  const { pagination, updatePagination } = useTablePagination('clientSearch');

  const bkBizId = ref(String(route.params.spaceId));
  const appId = ref(Number(route.params.appId));
  const viewPullRecordClientId = ref(0);
  const viewPullRecordClientUid = ref('');
  const listLoading = ref(false);
  const selectedClient = ref<
    {
      id: number;
      uid: string;
      current_release_name: string;
      target_release_name: string;
    }[]
  >([]);
  const isSearchEmpty = ref(false);
  const isShowPullRecordSlider = ref(false);
  const tableData = ref();
  const tableRef = ref();
  const isAcrossChecked = ref(false);
  const selecTableDataCount = ref(0);

  const releaseChangeStatusFilterList = [
    {
      text: t('成功'),
      value: 'Success',
    },
    {
      text: t('失败'),
      value: 'Failed',
    },
    {
      text: t('处理中'),
      value: 'Processing',
    },
  ];
  const releaseChangeStatusFilterChecked = ref<string[]>([]);
  const onlineStatusFilterList = [
    {
      text: t('在线'),
      value: 'Online',
    },
    {
      text: t('离线'),
      value: 'Offline',
    },
  ];
  const onlineStatusFilterChecked = ref<string[]>([]);
  const pollTimer = ref(0);
  const updateSortType = ref('null');

  // 当前页数据，不含禁用
  const selecTableData = computed(() => {
    return tableData.value
      .filter(
        (item: any) =>
          item.client?.spec &&
          !(item.client.spec.release_change_status !== 'Failed' || item.client.spec.online_status !== 'Online'),
      )
      .map((item: any) => ({
        ...item,
        id: item.client.id,
        uid: item.client.attachment.uid,
      }));
  });
  const crossPageSelect = computed(
    () => pagination.value.limit < pagination.value.count && selecTableDataCount.value !== 0,
  );
  const { selectType, selections, renderSelection, renderTableTip, handleRowCheckChange, handleClearSelection } =
    useTableAcrossCheck({
      dataCount: selecTableDataCount,
      curPageData: selecTableData, // 当前页数据，不含禁用
      rowKey: ['id'],
      crossPageSelect,
    });

  watch(
    () => route.params.appId,
    (val) => {
      if (val) {
        appId.value = Number(val);
        bkBizId.value = String(route.params.spaceId);
      }
    },
  );

  watch(
    () => searchQuery.value.search,
    (val) => {
      isSearchEmpty.value = Object.keys(val!).length !== 0;
      pagination.value.current = 1;
      loadList();
    },
    { deep: true },
  );

  watch(
    () => tableData.value,
    () => {
      if (pollTimer.value) {
        clearTimeout(pollTimer.value);
      }
      const pollClientIds = tableData.value
        .filter(
          (item: any) =>
            item.client.spec.release_change_status === 'Processing' && item.client.spec.online_status === 'Online',
        )
        .map((item: any) => item.client.id);
      if (pollClientIds.length > 0) {
        pollTimer.value = setInterval(() => {
          pollClientStatus(pollClientIds);
        }, 3000);
      }
    },
    { deep: true },
  );

  watch(
    selections,
    () => {
      isAcrossChecked.value = [CheckType.HalfAcrossChecked, CheckType.AcrossChecked].includes(selectType.value);
      selectedClient.value = selections.value.map((item) => {
        return {
          id: item.client.id,
          uid: item.client.attachment.uid,
          current_release_name: item.client.spec.current_release_name,
          target_release_name: item.client.spec.target_release_name,
        };
      });
    },
    {
      deep: true,
    },
  );

  onBeforeMount(() => {
    const tableSet = localStorage.getItem('client-show-column');
    settings.value.size = 'medium';
    if (tableSet) {
      const { checked, size } = JSON.parse(tableSet);
      const requiredChecked = settings.value.fields.filter((item) => item.disabled).map((item) => item.id);
      selectedShowColumn.value = [...requiredChecked, ...checked];
      settings.value.checked = checked;
      settings.value.size = size;
    }
  });

  onBeforeUnmount(() => {
    if (pollTimer.value) {
      clearTimeout(pollTimer.value);
    }
  });

  // 行的启用/禁用
  const isChecked = (row: any) => {
    if (![CheckType.AcrossChecked, CheckType.HalfAcrossChecked].includes(selectType.value)) {
      // 当前页状态传递
      return selections.value.some((item) => item.client?.id === row.client?.id);
    }
    // 跨页状态传递
    return !selections.value.some((item) => item.client?.id === row.client?.id);
  };

  const settings = ref({
    trigger: 'click',
    extCls: 'client-settings-custom',
    fields: [
      {
        name: 'UID',
        id: 'uid',
        disabled: true,
      },
      {
        name: 'IP',
        id: 'ip',
        disabled: true,
      },
      {
        name: t('客户端标签'),
        id: 'label',
        disabled: true,
      },
      {
        name: t('源版本'),
        id: 'current-version',
        disabled: true,
      },
      {
        name: t('最近一次拉取配置状态'),
        id: 'pull-status',
        disabled: true,
      },
      {
        name: t('最后一次拉取配置耗时'),
        id: 'pull-time',
        disabled: true,
      },
      {
        name: t('在线状态'),
        id: 'online-status',
        disabled: true,
      },
      {
        name: t('首次连接时间'),
        id: 'first-connect-time',
      },
      {
        name: t('最后心跳时间'),
        id: 'last-heartbeat-time',
      },
      {
        name: t('CPU资源占用'),
        id: 'cpu-resource',
      },
      {
        name: t('内存资源占用'),
        id: 'memory-resource',
      },
      {
        name: t('客户端组件类型'),
        id: 'client-type',
      },
      {
        name: t('客户端组件版本'),
        id: 'client-version',
      },
    ],
    checked: [
      'uid',
      'ip',
      'label',
      'current-version',
      'pull-status',
      'pull-time',
      'online-status',
      'first-connect-time',
      'last-heartbeat-time',
      'cpu-resource',
      'memory-resource',
      'client-type',
      'client-version',
    ],
    size: 'small',
  });

  const selectedShowColumn = ref([
    'uid',
    'ip',
    'label',
    'current-version',
    'pull-status',
    'pull-time',
    'online-status',
    'first-connect-time',
    'last-heartbeat-time',
    'cpu-resource',
    'memory-resource',
    'client-type',
    'client-version',
  ]);

  // const tableTips = {
  //   clientTag: '客户端标签与服务分组配合使用实现服务配置灰度发布场景',
  //   information: '主要用于记录客户端非标识性元数据，例如客户端用途等附加信息（标识性元数据使用客户端标签）',
  //   status:
  //     '客户端每 15 秒会向服务端发送一次心跳数据，如果服务端连续3个周期没有接收到客户端心跳数据，视客户端为离线状态',
  // };

  const loadList = async (pageChange = false) => {
    // 非跨页全选/半选 需要重置全选状态
    if (![CheckType.HalfAcrossChecked, CheckType.AcrossChecked].includes(selectType.value) || !pageChange) {
      handleClearSelection();
    }
    const params: IClinetCommonQuery = {
      start: pagination.value.limit * (pagination.value.current - 1),
      limit: pagination.value.limit,
      last_heartbeat_time: searchQuery.value.last_heartbeat_time,
      search: searchQuery.value.search,
      order: {
        desc: 'online_status',
      },
    };
    if (updateSortType.value === 'desc') {
      params.order!.desc = 'online_status,total_seconds';
    } else if (updateSortType.value === 'asc') {
      params.order!.asc = 'total_seconds';
    }
    try {
      listLoading.value = true;
      const res = await getClientQueryList(bkBizId.value, appId.value, params);
      tableData.value = res.data.details;
      tableData.value.forEach((item: any) => {
        const { client } = item;
        client.labels = Object.entries(JSON.parse(client.spec.labels)).map(([key, value]) => ({ key, value }));
      });
      pagination.value.count = res.data.count;
      selecTableDataCount.value = Number(res.data.exclusion_count);
    } catch (error) {
      console.error(error);
    } finally {
      listLoading.value = false;
    }
  };

  const linkToApp = (versionId: number) => {
    const routeData = router.resolve({
      name: 'service-config',
      params: { spaceId: bkBizId.value, appId: appId.value, versionId },
    });
    window.open(routeData.href, '_blank');
  };

  const handleShowPullRecord = (uid: string, id: number) => {
    viewPullRecordClientId.value = id;
    viewPullRecordClientUid.value = uid;
    isShowPullRecordSlider.value = true;
  };

  const handleCloseSlider = () => {
    isShowPullRecordSlider.value = false;
  };

  const handleClearConditionalList = () => {
    clientStore.$patch((state) => {
      state.searchQuery.search = {};
    });
    releaseChangeStatusFilterChecked.value = [];
    onlineStatusFilterChecked.value = [];
  };

  // 更改每页条数
  const handlePageLimitChange = (val: number) => {
    updatePagination('limit', val);
    loadList();
  };

  const handleFilter = ({ checked, index }: any) => {
    if (index === 5) {
      // 调整最近一次拉取配置筛选条件
      clientStore.$patch((state) => {
        state.searchQuery.search.release_change_status = [...checked];
      });
    } else {
      // 调整在线状态筛选条件
      clientStore.$patch((state) => {
        state.searchQuery.search.online_status = [...checked];
      });
    }
  };

  const handleSort = ({ type }: any) => {
    updateSortType.value = type;
    loadList();
  };

  const handleSettingsChange = ({ checked, size }: any) => {
    selectedShowColumn.value = [...checked];
    localStorage.setItem('client-show-column', JSON.stringify({ checked, size }));
  };

  // 选择单行
  const handleSelectionChange = (row: any) => {
    const isSelected = selections.value.some((item) => item.client.id === row.client.id);
    // 根据选择类型决定传递的状态
    const shouldBeChecked = isAcrossChecked.value ? isSelected : !isSelected;
    handleRowCheckChange(shouldBeChecked, { ...row, id: row.client.id });
  };

  const getErrorDetails = (item: any) => {
    const {
      release_change_failed_reason,
      specific_failed_reason,
      failed_detail_reason,
      current_release_name,
      target_release_name,
    } = item;
    const category = CLIENT_ERROR_CATEGORY_MAP.find((item) => item.value === release_change_failed_reason)?.name;
    const subclasses = CLIENT_ERROR_SUBCLASSES_MAP.find((item) => item.value === specific_failed_reason)?.name || '--';
    return `【 ${current_release_name} -> ${target_release_name} 】
    ${t('错误类别')}: ${category}
    ${t('错误子类别')}: ${subclasses}
    ${t('错误详情')}: ${failed_detail_reason}`;
  };

  // 重试成功
  const handleRetryConfirm = (ids: number[]) => {
    handleClearSelection(); // 清空选择框
    ids.forEach((id) => {
      const index = selectedClient.value.findIndex((item) => item.id === id);
      if (index > -1) {
        selectedClient.value.splice(index, 1);
        tableRef.value.toggleRowSelection(
          tableData.value.find((item: any) => item.client.id === id),
          false,
        );
      }
    });
    pollClientStatus(ids, true);
  };

  // 当有客户端状态处于处理中时 开启轮询
  const pollClientStatus = async (ids: number[], isRetry = false) => {
    const params: IClinetCommonQuery = {
      limit: ids.length,
      search: {
        client_ids: ids,
      },
    };
    try {
      const res = await getClientQueryList(bkBizId.value, appId.value, params);
      res.data.details.forEach((item: any) => {
        if (isRetry || item.client.spec.release_change_status !== 'Processing') {
          const pollClient = tableData.value.find((tableItem: any) => tableItem.client.id === item.client.id);
          pollClient.client.spec.release_change_status = item.client.spec.release_change_status;
          pollClient.client.spec.resource = item.client.spec.resource;
          pollClient.client.spec.last_heartbeat_time = item.client.spec.last_heartbeat_time;
        }
      });
    } catch (error) {
      console.error(error);
    }
  };
</script>

<style scoped lang="scss">
  .client-search-page {
    height: 100%;
    overflow: auto;
  }
  .header {
    position: relative;
    min-height: 120px;
    padding: 40px 120px 0 40px;
    background-image: linear-gradient(-82deg, #e8f0ff 10%, #f0f5ff 93%);
    box-shadow: 0 2px 4px 0 #1919290d;
    :deep(.head) {
      z-index: 10;
    }
    &::after {
      position: absolute;
      right: 0;
      top: 10px;
      content: '';
      width: 80px;
      height: 120px;
      background-image: url('../../../../assets/client-head-right.png');
      z-index: 0;
    }
    &::before {
      position: absolute;
      left: 0;
      top: 0px;
      content: '';
      width: 200px;
      height: 120px;
      background-image: url('../../../../assets/client-head-left.png');
      z-index: 0;
    }
  }
  .row-selection-display {
    line-height: 32px;
    background: #dcdee5;
    text-align: center;
    font-size: 12px;
    color: #63656e;
  }
  .content {
    padding: 24px;
  }

  .labels {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    min-height: 100%;
    span {
      margin-right: 4px;
      line-height: 28px;
    }
  }

  .current-version {
    display: flex;
    align-items: center;
    cursor: pointer;
    .text {
      margin-left: 4px;
    }
    .icon {
      color: #979ba5;
    }
    &:hover {
      color: #3a84ff;
      .icon {
        color: #3a84ff;
      }
    }
  }
  .release_change_status {
    display: flex;
    align-items: center;
    .spinner-icon {
      font-size: 14px;
      margin: 0 7px 0 1px;
      color: #3a84ff;
    }
    .dot {
      margin: 0 10px 0 4px;
      width: 8px;
      height: 8px;
      background: #f0f1f5;
      border: 1px solid #c4c6cc;
      border-radius: 50%;
      &.Success {
        background: #e5f6ea;
        border: 1px solid #3fc06d;
      }
      &.Failed {
        background: #ffe6e6;
        border: 1px solid #ea3636;
      }
    }
    .info-icon {
      font-size: 14px;
      margin-left: 9px;
    }
  }

  .online-status {
    display: flex;
    align-items: center;
    .dot {
      margin-right: 10px;
      width: 13px;
      height: 13px;
      border-radius: 50%;
      &.Online {
        background: #3fc06d;
        border: 3px solid #e0f5e7;
      }
      &.Offline {
        background: #979ba5;
        border: 3px solid #eeeef0;
      }
    }
  }
</style>

<style lang="scss">
  .client-settings-custom {
    padding: 0px !important;
    .setting-body {
      .setting-body-fields {
        max-height: inherit !important;
      }
    }
    .field-item {
      width: 200px !important;
    }
  }
</style>
