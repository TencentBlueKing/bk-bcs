<template>
  <section>
    <div class="record-table-wrapper">
      <bk-loading style="min-height: 300px" :loading="loading">
        <bk-table
          class="record-table"
          show-overflow-tooltip
          :row-height="0"
          :border="['outer']"
          :data="tableData"
          @column-sort="handleSort"
          @column-filter="handleFilter">
          <bk-table-column :label="t('操作时间')" width="155" :sort="true">
            <template #default="{ row }">
              {{ row.audit?.revision.created_at }}
            </template>
          </bk-table-column>
          <bk-table-column :label="t('所属服务')" width="190">
            <template #default="{ row }"> {{ row.app?.name || '--' }} </template>
          </bk-table-column>
          <bk-table-column
            :label="t('资源类型')"
            :width="locale === 'zh-cn' ? '96' : '160'"
            :filter="{
              filterFn: () => true,
              list: resTypeFilterList,
              checked: resTypeFilterChecked,
            }">
            <template #default="{ row }">
              {{ RECORD_RES_TYPE[row.audit?.spec.res_type as keyof typeof RECORD_RES_TYPE] || '--' }}
            </template>
          </bk-table-column>
          <bk-table-column
            :label="t('操作行为')"
            :width="locale === 'zh-cn' ? '114' : '240'"
            :filter="{
              filterFn: () => true,
              list: actionFilterList,
              checked: actionFilterChecked,
            }">
            <template #default="{ row }">
              {{ ACTION[row.audit?.spec.action as keyof typeof ACTION] || '--' }}
            </template>
          </bk-table-column>
          <bk-table-column :label="t('资源实例')" min-width="363">
            <template #default="{ row }">
              <div
                v-if="row.audit && row.audit.spec.res_instance"
                v-html="convertInstance(row.audit.spec.res_instance)"
                class="multi-line-styles"></div>
              <div v-else class="multi-line-styles">--</div>
              <!-- <div>{{ row.audit?.spec.res_instance || '--' }}</div> -->
            </template>
          </bk-table-column>
          <bk-table-column :label="t('操作人')" width="140">
            <template #default="{ row }">
              {{ row.audit?.spec.operator || '--' }}
            </template>
          </bk-table-column>
          <bk-table-column :label="t('操作途径')" :width="locale === 'zh-cn' ? '90' : '150'">
            <template #default="{ row }"> {{ row.audit?.spec.operate_way || '--' }} </template>
          </bk-table-column>
          <bk-table-column
            :label="t('状态')"
            :show-overflow-tooltip="false"
            :width="locale === 'zh-cn' ? '130' : '190'"
            :filter="{
              filterFn: () => true,
              list: approveStatusFilterList,
              checked: approveStatusFilterChecked,
            }">
            <template #default="{ row }">
              <template v-if="row.audit?.spec.status">
                <div :class="['dot', ...setApprovalClass(row.audit.spec.status)]"></div>
                {{ STATUS[row.audit.spec.status as keyof typeof STATUS] || '--' }}
                <!-- 上线时间icon -->
                <div
                  v-if="
                    row.strategy?.publish_time &&
                    [APPROVE_STATUS.PendApproval, APPROVE_STATUS.PendPublish].includes(row.audit.spec.status)
                  "
                  v-bk-tooltips="{
                    content: `${t('定时上线')}: ${row.strategy.publish_time}${
                      isTimeout(row.strategy.publish_time) ? `(${t('已过时')})` : ''
                    }`,
                    placement: 'top',
                  }"
                  class="time-icon"></div>
                <!-- 信息提示icon -->
                <info-line
                  v-if="![APPROVE_STATUS.PendApproval, APPROVE_STATUS.PendPublish].includes(row.audit.spec.status)"
                  v-bk-tooltips="{
                    content: statusTip(row),
                    placement: 'top',
                  }"
                  class="info-line" />
              </template>
              <template v-else>--</template>
            </template>
          </bk-table-column>
          <bk-table-column
            :label="t('操作')"
            :show-overflow-tooltip="false"
            :width="locale === 'zh-cn' ? '110' : '150'">
            <template #default="{ row }">
              <!-- 仅上线配置版本存在待审批或待上线等状态和相关操作 -->
              <div v-if="row.audit && row.audit.spec.action === 'PublishVersionConfig'" class="action-btns">
                <!-- 创建者且版本待上线 才展示上线按钮;审批通过的时间在定时上线的时间以前，上线按钮置灰 -->
                <bk-button
                  v-if="row.audit.spec.status === APPROVE_STATUS.PendPublish && row.app.creator === userInfo.username"
                  class="action-btn"
                  text
                  theme="primary"
                  :disabled="!isTimeout(row.strategy.publish_time) && !!row.strategy?.publish_time"
                  @click="handleConfirm(row, 'publish')">
                  {{ t('上线') }}
                </bk-button>
                <!-- 1.待审批状态 且 对应审批人才可显示 -->
                <!-- 2.版本首次在分组上线的情况，显示审批，点击审批直接通过 -->

                <template
                  v-else-if="
                    row.audit.spec.status === APPROVE_STATUS.PendApproval &&
                    row.strategy.approver_progress.includes(userInfo.username)
                  ">
                  <!-- 当前的记录在目标分组首次上线，直接审批通过 -->
                  <bk-button
                    v-if="row.audit.spec.is_compare"
                    class="action-btn"
                    text
                    theme="primary"
                    @click="handleApproval(row)">
                    {{ t('去审批') }}
                  </bk-button>
                  <!-- 非首次上线，需要打开对比抽屉 -->
                  <bk-button v-else class="action-btn" text theme="primary" @click="handleApproved(row)">
                    {{ t('审批') }}
                  </bk-button>
                </template>
                <!-- 审批驳回/已撤销才可显示 -->
                <bk-button
                  v-else-if="
                    [APPROVE_STATUS.RejectedApproval, APPROVE_STATUS.RevokedPublish].includes(row.audit.spec.status) &&
                    row.app.creator === userInfo.username
                  "
                  class="action-btn"
                  text
                  theme="primary"
                  @click="retrySubmission(row)">
                  {{ t('再次提交') }}
                </bk-button>
                <span v-else class="empty-action">--</span>
                <!-- 待上线/去审批状态 才显示更多操作；目前仅创建者有撤销权限 -->
                <MoreActions
                  v-if="
                    [APPROVE_STATUS.PendApproval, APPROVE_STATUS.PendPublish].includes(row.audit.spec.status) &&
                    row.strategy.creator === userInfo.username
                  "
                  @handle-undo="handleConfirm(row, $event)" />
              </div>
              <template v-else>--</template>
            </template>
          </bk-table-column>
          <template #empty>
            <TableEmpty :is-search-empty="isSearchEmpty" />
          </template>
        </bk-table>
        <bk-pagination
          v-model="pagination.current"
          class="table-list-pagination"
          location="left"
          :limit="pagination.limit"
          :layout="['total', 'limit', 'list']"
          :count="pagination.count"
          @change="handlePageChange"
          @limit-change="handlePageLimitChange" />
      </bk-loading>
    </div>
    <!-- 上线/撤销弹窗 -->
    <DialogConfirm
      v-model:show="confirmShow"
      :space-id="spaceId"
      :app-id="rowAppId"
      :release-id="rowReleaseId"
      :dialog-type="confirmType"
      :data="confirmData"
      @refresh-list="loadRecordList" />
    <!-- 审批对比弹窗 -->
    <VersionDiff
      :show="approvalShow"
      :space-id="spaceId"
      :app-id="rowAppId"
      :release-id="rowReleaseId"
      :released-groups="rowReleaseGroups"
      @close="closeApprovalDialog" />
  </section>
</template>

<script setup lang="ts">
  import { ref, watch } from 'vue';
  import { useRouter, useRoute } from 'vue-router';
  import { debounce } from 'lodash';
  import { useI18n } from 'vue-i18n';
  import { IRecordQuery, IDialogData, IRowData } from '../../../../../types/record';
  import { RECORD_RES_TYPE, ACTION, STATUS, INSTANCE, APPROVE_STATUS } from '../../../../constants/record';
  import { storeToRefs } from 'pinia';
  import useUserStore from '../../../../store/user';
  import { getRecordList, approve } from '../../../../api/record';
  import useTablePagination from '../../../../utils/hooks/use-table-pagination';
  import TableEmpty from '../../../../components/table/table-empty.vue';
  import MoreActions from './more-actions.vue';
  import DialogConfirm from './dialog-confirm.vue';
  import { InfoLine } from 'bkui-vue/lib/icon';
  import VersionDiff from './version-diff.vue';
  import BkMessage from 'bkui-vue/lib/message';
  import dayjs from 'dayjs';

  const props = withDefaults(
    defineProps<{
      spaceId: string;
      searchParams: IRecordQuery;
    }>(),
    {
      spaceId: '',
    },
  );

  const router = useRouter();
  const route = useRoute();
  const { t, locale } = useI18n();
  const { userInfo } = storeToRefs(useUserStore());
  const { pagination, updatePagination } = useTablePagination('recordList');

  const loading = ref(true);
  const isSearchEmpty = ref(false);
  const searchParams = ref<IRecordQuery>({});
  const actionTimeSrotMode = ref('');
  const tableData = ref<IRowData[]>([]);
  const approvalShow = ref(false);
  const rowAppId = ref(-1);
  const rowReleaseId = ref(-1);
  const rowReleaseGroups = ref<number[]>([]);
  const confirmShow = ref(false);
  const confirmType = ref('');
  const confirmData = ref<IDialogData>({
    service: '',
    version: '',
    group: '',
  });

  // 数据过滤 S
  // 1. 资源类型
  const resTypeFilterChecked = ref<string[]>([]);
  const resTypeFilterList = Object.entries(RECORD_RES_TYPE).map(([key, value]) => ({
    text: value,
    value: key,
  }));
  // 2. 操作行为
  const actionFilterChecked = ref<string[]>([]);
  const actionFilterList = Object.entries(ACTION).map(([key, value]) => ({
    text: value,
    value: key,
  }));
  // 3. 状态
  const approveStatusFilterChecked = ref<string[]>([]);
  const approveStatusFilterList = Object.entries(STATUS).map(([key, value]) => ({
    text: value,
    value: key,
  }));
  // 数据过滤 E

  watch(
    () => props.searchParams,
    (newV) => {
      searchParams.value = {
        ...newV,
      };
      searchParams.value.all = !(route.params.appId && Number(route.params.appId) > -1);
      if (searchParams.value.all) {
        delete searchParams.value.app_id;
      } else {
        searchParams.value.app_id = Number(route.params.appId);
      }
      loadRecordList();
    },
    { deep: true },
  );

  watch(
    () => route.params.appId,
    (newV) => {
      searchParams.value.all = !(newV && Number(newV) > -1);
      if (searchParams.value.all) {
        delete searchParams.value.app_id;
      } else {
        searchParams.value.app_id = Number(route.params.appId);
      }
      delete searchParams.value.id;
      loadRecordList();
    },
  );

  // 加载操作记录列表数据
  const loadRecordList = async () => {
    try {
      loading.value = true;
      const params: IRecordQuery = {
        start: pagination.value.limit * (pagination.value.current - 1),
        limit: pagination.value.limit,
        ...searchParams.value,
      };
      const res = await getRecordList(props.spaceId, params);
      // actionTimeSrotMode.value ? tableDataSort(res.details) : (tableData.value = res.details);
      tableDataSort(res.details);
      pagination.value.count = res.count;
      // 是否打开审批抽屉
      if (route.query.id) {
        openApprovalSideBar();
      }
    } catch (e) {
      console.error(e);
    } finally {
      loading.value = false;
    }
  };

  // 关闭审批对比弹窗
  const closeApprovalDialog = (refresh: string) => {
    approvalShow.value = false;
    // 去除url操作记录id
    if (route.query.id) {
      const newQuery = { ...route.query };
      delete newQuery.id;
      router.replace({
        query: {
          ...newQuery,
        },
      });
    }
    // 审批通过/驳回：刷新
    if (refresh) {
      loadRecordList();
    }
  };

  // 资源示例映射
  const convertInstance = (data: string) => {
    if (data.length) {
      let result = data.replace(/\n/g, '<br />');
      Object.keys(INSTANCE).forEach((key) => {
        result = result.replace(`${key}:`, `${INSTANCE[key as keyof typeof INSTANCE]}：`);
      });
      return result;
    }
  };

  // 状态提示信息
  const statusTip = (row: IRowData) => {
    if (!row) {
      return '--';
    }
    const { status } = row.audit.spec;
    const { updated_at: time, reject_reason: reason, reviser } = row.strategy;
    switch (status) {
      case APPROVE_STATUS.AlreadyPublish:
        return t('提示-已上线文案', { reviser, time });
      case APPROVE_STATUS.RejectedApproval:
        return t('提示-审批驳回', {
          reviser,
          time,
          reason,
        });
      case APPROVE_STATUS.RevokedPublish:
        return t('提示-已撤销', { reviser, time });
      case APPROVE_STATUS.Failure:
        return t('提示-失败');
      default:
        return '--';
    }
  };

  // 上线时间是否超时
  const isTimeout = (time: string) => {
    const currentTime = dayjs();
    const publishTime = dayjs(time);
    // 定时的上线时间是否在当前时间之前
    return publishTime.isBefore(currentTime);
  };

  // 上线/撤回提示框
  const handleConfirm = (row: IRowData, type: string) => {
    confirmType.value = type;
    confirmShow.value = true;
    const matchVersion = row.audit.spec.res_instance.match(/releases_name:([^\n]*)/);
    const matchGroup = row.audit.spec.res_instance.match(/group:([^\n]*)/);
    rowAppId.value = row.audit.attachment.app_id;
    rowReleaseId.value = row.strategy.release_id;
    confirmData.value = {
      service: row.app.name,
      version: matchVersion ? matchVersion[1] : '--',
      group: matchGroup ? matchGroup[1] : '--',
    };
  };

  // 再次提交
  const retrySubmission = (row: IRowData) => {
    const url = router.resolve({
      name: 'service-config',
      params: {
        appId: row.audit.attachment.app_id,
        versionId: row.strategy.release_id,
      },
    }).href;
    window.open(url, '_blank');
  };

  // 审批通过
  const handleApproved = debounce(async (row: IRowData) => {
    try {
      const { biz_id, app_id } = row.audit.attachment;
      const { release_id } = row.strategy;
      await approve(String(biz_id), app_id, release_id, {
        publish_status: APPROVE_STATUS.PendPublish,
      });
      BkMessage({
        theme: 'success',
        message: t('操作成功'),
      });
      loadRecordList();
    } catch (e) {
      console.log(e);
    }
  }, 300);

  // 去审批
  const handleApproval = debounce(
    (row: IRowData) => {
      rowAppId.value = row.audit?.attachment.app_id;
      rowReleaseId.value = row.strategy?.release_id;
      // 当前row已上线版本的分组id,为空表示全部分组上线
      rowReleaseGroups.value = row.strategy.scope.groups.map((group) => group.id);
      approvalShow.value = true;
      router.replace({
        query: {
          ...route.query,
          id: row.audit.id,
        },
      });
    },
    300,
    { leading: true, trailing: false },
  );

  // 是否打开审批抽屉
  const openApprovalSideBar = () => {
    // 如果url的操作记录id为待审批状态，且为可对比状态并且当前登录用户有权限审批时，允许打开审批抽屉
    const isCompare = tableData.value[0]?.audit.spec.is_compare; // 是否可以对比版本不同
    const pendApproval = tableData.value[0]?.strategy.publish_status === APPROVE_STATUS.PendApproval; // 是否待审批状态
    const isAuthorized = tableData.value[0]?.strategy.approver_progress.includes(userInfo.value.username); // 当前用户是否有权限审批
    if (isCompare && pendApproval && isAuthorized) {
      handleApproval(tableData.value[0]);
    }
  };

  // 数据过滤
  const handleFilter = ({ checked, index }: any) => {
    // index: 2.资源类型 3.操作行为 7.状态
    switch (index) {
      case 2:
        searchParams.value.resource_type = checked.join(',');
        break;
      case 3:
        searchParams.value.action = checked.join(',');
        break;
      case 7:
        searchParams.value.status = checked.join(',');
        break;

      default:
        break;
    }
    loadRecordList();
  };

  // 触发的排序模式
  const handleSort = ({ type }: any) => {
    actionTimeSrotMode.value = type === 'null' ? '' : type;
    tableDataSort(tableData.value);
  };

  // 列表排序
  const tableDataSort = (data: IRowData[]) => {
    if (actionTimeSrotMode.value === 'desc') {
      tableData.value = data.sort(
        (a, b) => dayjs(b.audit.revision.created_at).valueOf() - dayjs(a.audit.revision.created_at).valueOf(),
      );
    } else if (actionTimeSrotMode.value === 'asc') {
      tableData.value = data.sort(
        (a, b) => dayjs(a.audit.revision.created_at).valueOf() - dayjs(b.audit.revision.created_at).valueOf(),
      );
    } else {
      tableData.value = data;
    }
  };

  // 审批状态颜色
  const setApprovalClass = (status: APPROVE_STATUS) => {
    return [
      [APPROVE_STATUS.AlreadyPublish, APPROVE_STATUS.Success].includes(status) ? 'green' : '',
      status === APPROVE_STATUS.PendPublish ? 'gray' : '',
      [APPROVE_STATUS.RevokedPublish, APPROVE_STATUS.RejectedApproval, APPROVE_STATUS.Failure].includes(status)
        ? 'red'
        : '',
      status === APPROVE_STATUS.PendApproval ? 'orange' : '',
    ];
  };

  //  翻页
  const handlePageChange = (val: number) => {
    pagination.value.current = val;
    loadRecordList();
  };

  const handlePageLimitChange = (val: number) => {
    updatePagination('limit', val);
    if (pagination.value.current === 1) {
      loadRecordList();
    }
  };
</script>

<style lang="scss" scoped>
  .record-table-wrapper {
    :deep(.bk-table-body) {
      max-height: calc(100vh - 280px);
      overflow: auto;
    }
    .dot {
      margin-right: 8px;
      display: inline-block;
      width: 8px;
      height: 8px;
      border-radius: 50%;
      &.green {
        border: 1px solid #3fc06d;
        background-color: #e5f6ea;
      }
      &.gray {
        border: 1px solid #c4c6cc;
        background-color: #f0f1f5;
      }
      &.red {
        border: 1px solid #ea3636;
        background-color: #ffe6e6;
      }
      &.orange {
        border: 1px solid #ff9c01;
        background-color: #ffe8c3;
      }
    }
    // .status-text {
    //   display: inline-block;
    // }
    .time-icon {
      position: relative;
      margin-left: 8px;
      display: inline-block;
      width: 16px;
      height: 16px;
      vertical-align: bottom;
      border: 1px solid #3a84ff;
      border-radius: 50%;
      box-shadow: inset 0 0 0 0.4px #3a84ff;
      &::after {
        content: '';
        position: absolute;
        bottom: calc(50% - 1px);
        left: calc(50% - 1px);
        width: 35%;
        height: 35%;
        border-left: 1px solid #3a84ff;
        border-bottom: 1px solid #3a84ff;
        box-shadow:
          0 0.4px 0 0 #3a84ff,
          -0.4px 0 0 0 #3a84ff;
      }
    }
    .info-line {
      margin-left: 8px;
      font-size: 16px;
      vertical-align: bottom;
      transform: scale(1.1);
    }
  }
  .action-btns {
    position: relative;
  }
  .action-btn {
    vertical-align: sub;
    // & + .more-actions {
    //   margin-left: 8px;
    // }
  }
  .table-list-pagination {
    padding: 12px;
    border: 1px solid #dcdee5;
    border-top: none;
    border-radius: 0 0 2px 2px;
    background: #ffffff;
    :deep(.bk-pagination-list.is-last) {
      margin-left: auto;
    }
  }
  .record-table {
    :deep(.bk-table-body table tbody tr td) {
      .cell {
        display: inline-block;
        height: auto;
        line-height: normal;
        vertical-align: middle;
      }
      &:last-child .cell {
        // 更多操作显示
        overflow: unset;
      }
    }
  }
  // .ellipsis {
  //   overflow: hidden;
  //   text-overflow: ellipsis;
  //   white-space: nowrap;
  // }
  .multi-line-styles {
    padding: 7px 0;
    display: flex;
    justify-content: flex-start;
    align-items: center;
    width: 100%;
    height: 100%;
    min-height: 42px;
    overflow: hidden;
    white-space: normal;
    word-wrap: break-word;
    word-break: break-all;
    line-height: 21px;
  }
  .empty-action {
    margin-right: 50px;
    vertical-align: sub;
  }
</style>
