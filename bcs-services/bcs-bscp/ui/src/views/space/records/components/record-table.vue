<template>
  <section>
    <div class="record-table-wrapper">
      <bk-loading style="min-height: 300px" :loading="loading">
        <bk-table class="record-table" show-overflow-tooltip :row-height="0" :border="['outer']" :data="tableData">
          <bk-table-column :label="t('操作时间')" width="105">
            <template #default="{ row }">
              {{ row.audit?.revision.created_at }}
            </template>
          </bk-table-column>
          <bk-table-column :label="t('所属服务')" min-width="180">
            <template #default="{ row }"> {{ row.app?.name || '--' }} </template>
          </bk-table-column>
          <bk-table-column :label="t('资源类型')" width="96">
            <template #default="{ row }">
              {{ RECORD_RES_TYPE[row.audit?.spec.res_type as keyof typeof RECORD_RES_TYPE] || '--' }}
            </template>
          </bk-table-column>
          <bk-table-column :label="t('操作行为')" width="114">
            <template #default="{ row }">
              {{ ACTION[row.audit?.spec.action as keyof typeof ACTION] || '--' }}
            </template>
          </bk-table-column>
          <bk-table-column :label="t('资源实例')" min-width="363">
            <template #default="{ row }">
              <div
                v-if="row.audit && row.audit.spec.res_instance"
                class="multi-line-styles"
                v-html="convertInstance(row.audit.spec.res_instance)"></div>
              <div v-else class="multi-line-styles">--</div>
              <!-- <div>{{ row.audit?.spec.res_instance || '--' }}</div> -->
            </template>
          </bk-table-column>
          <bk-table-column :label="t('操作人')" min-width="110">
            <template #default="{ row }">
              {{ row.audit?.spec.operator || '--' }}
            </template>
          </bk-table-column>
          <bk-table-column :label="t('操作途径')" width="90">
            <template #default="{ row }"> {{ row.audit?.spec.operate_way || '--' }} </template>
          </bk-table-column>
          <bk-table-column
            :label="t('状态')"
            :show-overflow-tooltip="false"
            :width="locale === 'zh-cn' ? '130' : '160'">
            <template #default="{ row }">
              <template v-if="row.audit?.spec.status">
                <div
                  :class="[
                    'dot',
                    {
                      green: [APPROVE_STATUS.AlreadyPublish, APPROVE_STATUS.Success].includes(row.audit.spec.status),
                      gray: row.audit.spec.status === APPROVE_STATUS.PendPublish,
                      red: [
                        APPROVE_STATUS.RevokedPublish,
                        APPROVE_STATUS.RejectedApproval,
                        APPROVE_STATUS.Failure,
                      ].includes(row.audit.spec.status),
                      orange: row.audit.spec.status === APPROVE_STATUS.PendApproval,
                    },
                  ]"></div>
                {{ STATUS[row.audit.spec.status as keyof typeof STATUS] || '--' }}
                <!-- 上线时间icon -->
                <div
                  v-if="
                    row.strategy?.publish_time &&
                    [APPROVE_STATUS.PendApproval, APPROVE_STATUS.PendPublish].includes(row.audit.spec.status)
                  "
                  v-bk-tooltips="{
                    content: `${t('定时上线')}: ${row.strategy.publish_time}${
                      isTimeout(row.strategy.publish_time) ? '（已过时）' : ''
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
            :width="locale === 'zh-cn' ? '110' : '120'">
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
                  v-if="
                    row.audit.spec.status === APPROVE_STATUS.PendApproval &&
                    row.strategy.approver_progress.includes(userInfo.username)
                  ">
                  <!-- 当前记录在目标分组首次上线，直接审批通过 -->
                  <bk-button v-if="groupStatus(row)" class="action-btn" text theme="primary" @click="handlePass(row)">
                    {{ t('审批') }}
                  </bk-button>
                  <!-- 非首次上线，需要打开审批抽屉 -->
                  <bk-button v-else class="action-btn" text theme="primary" @click="handleApproval(row)">
                    {{ t('去审批') }}
                  </bk-button>
                </template>

                <!-- 审批驳回/已撤销才可显示 -->
                <bk-button
                  v-if="
                    [APPROVE_STATUS.RejectedApproval, APPROVE_STATUS.RevokedPublish].includes(row.audit.spec.status)
                  "
                  class="action-btn"
                  text
                  theme="primary"
                  @click="retrySubmission(row)">
                  {{ t('再次提交') }}
                </bk-button>
                <!-- 待上线/去审批状态 才显示更多操作；目前仅创建者有撤销权限 -->
                <MoreActions
                  v-if="
                    [APPROVE_STATUS.PendApproval, APPROVE_STATUS.PendPublish].includes(row.audit.spec.status) &&
                    row.app.creator === userInfo.username
                  "
                  @handle-undo="handleConfirm(row, $event)" />
                <!-- 当前登录用户在审批人和创建者名单都没有时，表示无权操作此条item -->
                <template
                  v-if="
                    row.audit.spec.status === APPROVE_STATUS.AlreadyPublish ||
                    !`${row.strategy.approver_progress},${row.app.creator}`.includes(userInfo.username)
                  ">
                  --
                </template>
              </div>
              <template v-else>--</template>
            </template>
          </bk-table-column>
          <template #empty>
            <TableEmpty :is-search-empty="isSearchEmpty" @clear="clearSearchInfo" />
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
      :data="confirmData" />
    <!-- 审批对比弹窗 -->
    <VersionDiff
      v-model:show="approvalShow"
      :space-id="spaceId"
      :app-id="rowAppId"
      :releases-id="rowReleaseId"
      :released-groups="rowReleaseGroups" />
  </section>
</template>

<script setup lang="ts">
  import { ref, onMounted } from 'vue';
  import { useRouter } from 'vue-router';
  import { debounce } from 'lodash';
  import { useI18n } from 'vue-i18n';
  import { IRecordQuery, IDialogData, IRowData } from '../../../../../types/record';
  import { RECORD_RES_TYPE, ACTION, STATUS, INSTANCE, APPROVE_STATUS } from '../../../../constants/record';
  import { storeToRefs } from 'pinia';
  import useUserStore from '../../../../store/user';
  import { getRecordList } from '../../../../api/record';
  import useTablePagination from '../../../../utils/hooks/use-table-pagination';
  import TableEmpty from '../../../../components/table/table-empty.vue';
  import MoreActions from './more-actions.vue';
  import DialogConfirm from './dialog-confirm.vue';
  import { InfoLine } from 'bkui-vue/lib/icon';
  import VersionDiff from './version-diff.vue';

  const props = withDefaults(
    defineProps<{
      spaceId: string;
      searchParams?: IRecordQuery;
    }>(),
    {
      spaceId: '',
      searchParams: () => ({
        all: true,
      }),
    },
  );

  const router = useRouter();
  const { t, locale } = useI18n();
  const { userInfo } = storeToRefs(useUserStore());
  const { pagination, updatePagination } = useTablePagination('recordList');

  const loading = ref(false);
  const isSearchEmpty = ref(false);
  const tableData = ref([]);
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

  // watch(
  //   () => spaceId.value,
  //   async () => {
  //     pagination.value.current = 1;
  //     await loadRecordList();
  //   },
  // );

  // watch(confirmShow, (newV) => {
  //   if (!newV) {
  //     confirmData.value = {
  //       biz_id: '',
  //       app_id: -1,
  //       release_id: -1,
  //       service: '',
  //       version: '',
  //       group: '',
  //     };
  //   }
  // });

  // watch(confirmShow, (newV) => {
  //   if (!newV) {
  //     rowAppId.value = -1;
  //     rowReleaseId.value = -1;
  //   }
  // });

  onMounted(async () => {
    await loadRecordList();
  });

  // 加载操作记录列表数据
  const loadRecordList = async () => {
    try {
      loading.value = true;
      const params: IRecordQuery = {
        start: pagination.value.limit * (pagination.value.current - 1),
        limit: pagination.value.limit,
        ...props.searchParams,
      };
      const res = await getRecordList(props.spaceId, params);
      console.log(res, 'res');
      tableData.value = res.details;
      pagination.value.count = res.count;
    } catch (e) {
      console.error(e);
    } finally {
      loading.value = false;
    }
  };

  // 清空搜索框
  const clearSearchInfo = () => {
    // 清空搜索框
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
    const currentTime = Date.now();
    const publishTime = new Date(time).getTime();
    return publishTime < currentTime;
  };

  // 上线/撤回提示框
  const handleConfirm = (row: IRowData, type: string) => {
    confirmType.value = type;
    confirmShow.value = true;
    const matchVersion = row.audit.spec.res_instance.match(/releases_name:([^\n]*)/);
    const matchGroup = row.audit.spec.res_instance.match(/group:([^\n]*)/);
    // biz_id: String(row.audit.attachment.biz_id),
    // release_id: row.strategy.release_id,
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

  // 是否首次在目标分组上线
  const groupStatus = (row: IRowData) => {
    console.log(row.audit.id, '---------');
    return true;
  };

  // 审批通过
  const handlePass = (row: IRowData) => {
    console.log(row);
  };
  // 去审批
  const handleApproval = debounce(
    (row: IRowData) => {
      console.log('哇哈哈哈');
      rowAppId.value = row.audit?.attachment.app_id;
      rowReleaseId.value = row.strategy?.release_id;
      // 当前row已上线版本的分组id,为空表示全部分组上线
      rowReleaseGroups.value = row.strategy.scope.groups.map((group) => group.id);
      approvalShow.value = true;
    },
    300,
    { leading: true, trailing: false },
  );

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
  // .action-btns {
  //   display: flex;
  //   align-items: center;
  // }
  .action-btn {
    vertical-align: sub;
    & + .more-actions {
      margin-left: 8px;
    }
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
</style>
