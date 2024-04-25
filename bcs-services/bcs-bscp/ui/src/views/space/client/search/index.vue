<template>
  <section class="client-search-page">
    <div class="header">
      <ClientHeader :title="t('客户端查询')" @search="loadList" />
    </div>
    <div v-if="appId" class="content">
      <!-- @todo 重试功能待接口支持 -->
      <!-- <bk-button style="margin-bottom: 16px" :disabled="!selectedClient.length">批量重试</bk-button> -->
      <bk-loading style="min-height: 100px" :loading="listLoading">
        <bk-table
          :data="tableData"
          :border="['outer', 'row']"
          :remote-pagination="true"
          :pagination="pagination"
          :key="appId"
          :checked="selectedClient"
          :is-row-select-enable="isRowSelectEnable"
          show-overflow-tooltip
          @page-limit-change="handlePageLimitChange"
          @page-value-change="loadList"
          @column-filter="handleFilter">
          <!-- <bk-table-column type="selection" :min-width="40" :width="40"> </bk-table-column> -->
          <bk-table-column label="UID" :width="254" prop="attachment.uid"></bk-table-column>
          <bk-table-column label="IP" :width="120" prop="spec.ip"></bk-table-column>
          <bk-table-column :label="t('客户端标签')" :min-width="296">
            <template #default="{ row }">
              <div v-if="row.spec && row.labels.length" class="labels">
                <span v-for="(label, index) in row.labels" :key="index">
                  <Tag v-if="index < 3">
                    {{ label.key + '=' + label.value }}
                  </Tag>
                </span>
                <span v-if="row.labels.length > 3">
                  <Tag> +{{ row.labels.length - 3 }} </Tag>
                </span>
              </div>
              <span v-else>--</span>
            </template>
          </bk-table-column>
          <bk-table-column :label="t('当前配置版本')" :width="140">
            <template #default="{ row }">
              <div
                v-if="row.spec && row.spec.current_release_id"
                class="current-version"
                @click="linkToApp(row.spec.current_release_id)">
                <Share fill="#979BA5" />
                <span class="text">{{ row.spec.current_release_name }}</span>
              </div>
              <span v-else>--</span>
            </template>
          </bk-table-column>
          <bk-table-column
            :label="t('最近一次拉取配置状态')"
            :width="178"
            :filter="{
              filterFn: () => true,
              list: releaseChangeStatusFilterList,
              checked: releaseChangeStatusFilterChecked,
            }">
            <template #default="{ row }">
              <div v-if="row.spec" class="release_change_status">
                <Spinner v-if="row.spec.release_change_status === 'Sikp'" class="spinner-icon" fill="#3A84FF" />
                <div v-else :class="['dot', row.spec.release_change_status]"></div>
                <span>{{ CLIENT_STATUS_MAP[row.spec.release_change_status as keyof typeof CLIENT_STATUS_MAP] }}</span>
                <InfoLine
                  v-if="row.spec.release_change_status === 'Failed'"
                  class="info-icon"
                  fill="#979BA5"
                  v-bk-tooltips="{ content: row.spec.failed_detail_reason }" />
              </div>
            </template>
          </bk-table-column>
          <!-- <bk-table-column label="附加信息" :width="244"></bk-table-column> -->
          <bk-table-column
            :label="t('在线状态')"
            :width="94"
            :filter="{
              filterFn: () => true,
              list: onlineStatusFilterList,
              checked: onlineStatusFilterChecked,
            }">
            <template #default="{ row }">
              <div v-if="row.spec" class="online-status">
                <div :class="['dot', row.spec.online_status]"></div>
                <span>{{ row.spec.online_status === 'Online' ? t('在线') : t('离线') }}</span>
              </div>
            </template>
          </bk-table-column>
          <bk-table-column :label="t('首次连接时间')" :width="154">
            <template #default="{ row }">
              <span v-if="row.spec">
                {{ datetimeFormat(row.spec.first_connect_time) }}
              </span>
            </template>
          </bk-table-column>
          <bk-table-column :label="t('最后心跳时间')" :width="154">
            <template #default="{ row }">
              <span v-if="row.spec">
                {{ datetimeFormat(row.spec.last_heartbeat_time) }}
              </span>
            </template>
          </bk-table-column>
          <bk-table-column :label="t('CPU资源占用(当前/最大)')" :width="174">
            <template #default="{ row }">
              <span v-if="row.spec">
                {{ showResourse(row.spec.resource).cpuResourse }}
              </span>
            </template>
          </bk-table-column>
          <bk-table-column :label="t('内容资源占用(当前/最大)')" :width="170">
            <template #default="{ row }">
              <span v-if="row.spec">
                {{ showResourse(row.spec.resource).memoryResource }}
              </span>
            </template>
          </bk-table-column>
          <bk-table-column :label="t('客户端组件类型')" :width="128" prop="spec.client_type"></bk-table-column>
          <bk-table-column :label="t('客户端组件版本')" :width="128" prop="spec.client_version"></bk-table-column>
          <bk-table-column :label="t('操作')" :width="148" fixed="right">
            <template #default="{ row }">
              <div v-if="row.spec">
                <bk-button theme="primary" text @click="handleShowPullRecord(row.attachment.uid, row.id)">
                  {{ t('配置拉取记录') }}
                </bk-button>
                <!-- <bk-button
                  v-if="row.spec.release_change_status === 'Failed'"
                  style="margin-left: 18px"
                  theme="primary"
                  text
                  @click="console.log(row)">
                  重试
                </bk-button> -->
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
  import { ref, watch } from 'vue';
  import { useRoute, useRouter } from 'vue-router';
  import { Share, Spinner, InfoLine } from 'bkui-vue/lib/icon';
  import { storeToRefs } from 'pinia';
  import { Tag } from 'bkui-vue';
  import { getClientQueryList, createClientSearchRecord } from '../../../../api/client';
  import ClientHeader from '../components/client-header.vue';
  import PullRecord from './components/pull-record.vue';
  import { datetimeFormat } from '../../../../utils';
  import { CLIENT_STATUS_MAP } from '../../../../constants/client';
  import { IClinetCommonQuery } from '../../../../../types/client';
  import useClientStore from '../../../../store/client';
  import TableEmpty from '../../../../components/table/table-empty.vue';
  import Exception from '../components/exception.vue';
  import { useI18n } from 'vue-i18n';

  const { t } = useI18n();

  interface IResourseType {
    cpu_usage: number;
    cpu_max_usage: number;
    memory_usage: string;
    memory_max_usage: string;
  }

  const clientStore = useClientStore();
  const { searchQuery } = storeToRefs(clientStore);

  const route = useRoute();
  const router = useRouter();
  const bkBizId = ref(String(route.params.spaceId));
  const appId = ref(Number(route.params.appId));
  const viewPullRecordClientId = ref(0);
  const viewPullRecordClientUid = ref('');
  const listLoading = ref(false);
  const selectedClient = ref([]);
  const isSearchEmpty = ref(false);
  const isShowPullRecordSlider = ref(false);
  const tableData = ref();
  const pagination = ref({
    count: 0,
    current: 1,
    limit: 10,
  });
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
    {
      text: t('跳过'),
      value: 'Skip',
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
      loadList();
    },
    { deep: true },
  );

  const showResourse = (resourse: IResourseType) => {
    return {
      cpuResourse: `${resourse.cpu_usage} ${t('核')}/${resourse.cpu_max_usage} ${t('核')}`,
      memoryResource: `${resourse.memory_usage}MB/${resourse.memory_max_usage}MB`,
    };
  };

  const isRowSelectEnable = ({ row, isCheckAll }: any) =>
    isCheckAll || !(row.spec && row.spec.release_change_status !== 'Failed');

  // const tableTips = {
  //   clientTag: '客户端标签与服务分组配合使用实现服务配置灰度发布场景',
  //   information: '主要用于记录客户端非标识性元数据，例如客户端用途等附加信息（标识性元数据使用客户端标签）',
  //   status:
  //     '客户端每 15 秒会向服务端发送一次心跳数据，如果服务端连续3个周期没有接收到客户端心跳数据，视客户端为离线状态',
  // };

  const loadList = async () => {
    const params: IClinetCommonQuery = {
      start: pagination.value.limit * (pagination.value.current - 1),
      limit: pagination.value.limit,
      last_heartbeat_time: searchQuery.value.last_heartbeat_time,
      search: searchQuery.value.search,
    };
    try {
      listLoading.value = true;
      const res = await getClientQueryList(bkBizId.value, appId.value, params);
      tableData.value = res.data.details;
      tableData.value.forEach((item: any) => {
        item.labels = Object.entries(JSON.parse(item.spec.labels)).map(([key, value]) => ({ key, value }));
      });
      pagination.value.count = res.data.count;
      // 添加最近查询
      if (Object.keys(searchQuery.value.search!).length > 0) {
        await createClientSearchRecord(bkBizId.value, appId.value, {
          search_type: 'recent',
          search_condition: searchQuery.value.search!,
        });
      }
    } catch (error) {
      console.error(error);
    } finally {
      listLoading.value = false;
    }
  };

  const linkToApp = (versionId: number) => {
    router.push({ name: 'service-config', params: { spaceId: bkBizId.value, appId: appId.value, versionId } });
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
    pagination.value.limit = val;
    loadList();
  };

  const handleFilter = ({ checked, index }: any) => {
    if (index === 4) {
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
</script>

<style scoped lang="scss">
  .header {
    height: 120px;
    padding: 40px 120px 0 40px;
    background: #eff5ff;
  }
  .content {
    padding: 24px;
  }

  .labels {
    display: flex;
    flex-wrap: wrap;
  }

  .current-version {
    display: flex;
    align-items: center;
    cursor: pointer;
    .text {
      margin-left: 4px;
    }
  }
  .release_change_status {
    display: flex;
    align-items: center;
    .spinner-icon {
      font-size: 12px;
      margin-right: 10px;
    }
    .dot {
      margin-right: 10px;
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
